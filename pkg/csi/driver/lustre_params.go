package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	lustre "github.com/oracle/oci-go-sdk/v65/lustrefilestorage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LustreStorageClassParameters holds Lustre-specific StorageClass attributes.
type LustreStorageClassParameters struct {
	SubnetId           string
	CompartmentId      string
	AvailabilityDomain string
	NSGIds             []string

	// filesystem settings
	FileSystemName  string
	PerformanceTier string // MBPS_PER_TB_125|250|500|1000
	KmsKeyId        string

	// root squash settings
	RootSquashEnabled          bool
	RootSquashUid              int
	RootSquashGid              int
	RootSquashUidSpecified     bool
	RootSquashGidSpecified     bool
	RootSquashClientExceptions []string

	// pass-through to VolumeContext
	SetupLnet                 string
	LustreSubnetCidr          string
	LustrePostMountParameters string

	// initial tags
	SCTags *config.TagConfig
}

// extractLustreStorageClassParameters parses and validates Lustre SC parameters.
// Returns:
//   - updated logger (with useful fields)
//   - early CreateVolumeResponse (if validation decides to short-circuit)
//   - parsed parameters
//   - error (grpc status)
//   - done (true if CreateVolume should return early)
func extractLustreStorageClassParameters(
	ctx context.Context,
	d *LustreControllerDriver,
	log *zap.SugaredLogger,
	volumeName string,
	parameters map[string]string,
	identityClient client.IdentityInterface,
) (*zap.SugaredLogger, *csi.CreateVolumeResponse, *LustreStorageClassParameters, error) {

	params := &LustreStorageClassParameters{
		CompartmentId:     d.config.CompartmentID,
		RootSquashEnabled: false,
		SCTags:            &config.TagConfig{},
	}

	if d.config.Tags != nil && d.config.Tags.Lustre != nil {
		if d.config.Tags.Lustre.FreeformTags != nil {
			params.SCTags.FreeformTags = d.config.Tags.Lustre.FreeformTags
		}
		if d.config.Tags.Lustre.DefinedTags != nil {
			params.SCTags.DefinedTags = d.config.Tags.Lustre.DefinedTags
		}
	}

	// subnetId (required)
	subnetId, ok := parameters["subnetId"]
	if !ok || strings.TrimSpace(subnetId) == "" {
		log.Errorf("Missing required parameter: subnetId")
		return log, nil, nil, status.Errorf(codes.InvalidArgument, "Missing required parameter: subnetId")
	}
	params.SubnetId = subnetId
	log = log.With("subnetId", subnetId)

	// compartmentId (optional; default to cluster compartment)
	if compartmentId, ok := parameters["compartmentId"]; ok && strings.TrimSpace(compartmentId) != "" {
		params.CompartmentId = compartmentId
	}
	log = log.With("compartmentId", params.CompartmentId)

	if nsgJSON, ok := parameters["nsgIds"]; ok && strings.TrimSpace(nsgJSON) != "" {
		var nsgs []string
		if err := json.Unmarshal([]byte(nsgJSON), &nsgs); err != nil {
			log.With(zap.Error(err)).Error("Failed to parse nsgIds (expect JSON array of strings)")
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "Failed to parse nsgIds (expect JSON array of strings) : %s", err.Error())
		}
		params.NSGIds = nsgs
		log = log.With("nsgIds", nsgs)
	}

	// performanceTier (required)
	if tier, ok := parameters["performanceTier"]; ok && strings.TrimSpace(tier) != "" {
		t := strings.TrimSpace(tier)
		if tierEnum, valid := lustre.GetMappingCreateLustreFileSystemDetailsPerformanceTierEnum(t); !valid {
			return log, nil, nil, status.Errorf(
				codes.InvalidArgument,
				"Invalid performanceTier: %s. Supported values: %s",
				t,
				strings.Join(lustre.GetCreateLustreFileSystemDetailsPerformanceTierEnumStringValues(), ","),
			)
		} else {
			// Normalize to canonical enum string (e.g., MBPS_PER_TB_125)
			params.PerformanceTier = string(tierEnum)
		}
	} else {
		return log, nil, nil, status.Errorf(codes.InvalidArgument, "Missing required parameter: performanceTier")
	}
	log = log.With("performanceTier", params.PerformanceTier)

	// kmsKeyId (optional)
	if kms, ok := parameters["kmsKeyId"]; ok && strings.TrimSpace(kms) != "" {
		params.KmsKeyId = strings.TrimSpace(kms)
	}

	// root squash (optional)
	if rsEnabled, ok := parameters["rootSquashEnabled"]; ok && strings.TrimSpace(rsEnabled) != "" {
		rsEnabledTrim := strings.TrimSpace(rsEnabled)
		rsEnabledLower := strings.ToLower(rsEnabledTrim)

		if rsEnabledLower != "true" && rsEnabledLower != "false" {
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "Invalid rootSquashEnabled: %s. Allowed values: true or false", rsEnabled)
		}
		params.RootSquashEnabled = strings.EqualFold(rsEnabledLower, "true")
		log = log.With("rootSquashEnabled", params.RootSquashEnabled)
	}
	if rsUid, ok := parameters["rootSquashUid"]; ok && strings.TrimSpace(rsUid) != "" {
		if uid, err := strconv.Atoi(strings.TrimSpace(rsUid)); err == nil && uid >= 0 {
			params.RootSquashUid = uid
			params.RootSquashUidSpecified = true
		} else {
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "Invalid rootSquashUid: %s", rsUid)
		}
	}
	if rsGid, ok := parameters["rootSquashGid"]; ok && strings.TrimSpace(rsGid) != "" {
		if gid, err := strconv.Atoi(strings.TrimSpace(rsGid)); err == nil && gid >= 0 {
			params.RootSquashGid = gid
			params.RootSquashGidSpecified = true
		} else {
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "Invalid rootSquashGid: %s", rsGid)
		}
	}
	if rsEx, ok := parameters["rootSquashClientExceptions"]; ok && strings.TrimSpace(rsEx) != "" {
		var exceptions []string
		if err := json.Unmarshal([]byte(rsEx), &exceptions); err != nil {
			log.With(zap.Error(err)).Error("Failed to parse rootSquashClientExceptions (expect JSON array of strings)")
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "Failed to parse rootSquashClientExceptions (expect JSON array of strings) : %s", err.Error())
		}
		if len(exceptions) > 10 {
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "rootSquashClientExceptions supports max 10 entries")
		}
		params.RootSquashClientExceptions = exceptions
	}

	// pass-through
	if v, ok := parameters["setupLnet"]; ok && strings.TrimSpace(v) != "" {
		setupLnet := strings.TrimSpace(v)
		setupLnetLower := strings.ToLower(setupLnet)
		if setupLnetLower != "true" && setupLnetLower != "false" {
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "Invalid setupLnet: %s. Allowed values: true or false", v)
		}
		params.SetupLnet = setupLnetLower
	}
	if v, ok := parameters["lustreSubnetCidr"]; ok {
		cidr := strings.TrimSpace(v)
		if cidr != "" {
			if _, _, err := net.ParseCIDR(cidr); err != nil {
				return log, nil, nil, status.Errorf(codes.InvalidArgument, "Invalid lustreSubnetCidr: %s", cidr)
			}
			params.LustreSubnetCidr = cidr
		}
	}
	if v, ok := parameters["lustrePostMountParameters"]; ok {
		lustrePostMountParameters := strings.TrimSpace(v)
		if lustrePostMountParameters != "" {
			if err := csi_util.ValidateLustreParameters(log, lustrePostMountParameters); err != nil {
				log.With(zap.Error(err)).Error("Invalid lustrePostMountParameters")
				return log, nil, nil, status.Errorf(codes.InvalidArgument, "Invalid lustrePostMountParameters: %v", err)
			}
			params.LustrePostMountParameters = lustrePostMountParameters
		}
	}

	// filesystemName (optional, default derive from volumeName last 8 allowed chars)
	if fsn, ok := parameters["fileSystemName"]; ok && strings.TrimSpace(fsn) != "" {
		fileSystemName := strings.TrimSpace(fsn)
		if err := validateLustreFileSystemName(fileSystemName); err != nil {
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "Invalid fileSystemName: %v", err)
		}
		params.FileSystemName = fileSystemName
	} else {
		params.FileSystemName = deriveDefaultFSName(volumeName)
	}

	// initialFreeformTagsOverride, initialDefinedTagsOverride come
	if freeformStr, ok := parameters[initialFreeformTagsOverride]; ok && strings.TrimSpace(freeformStr) != "" {
		freeform := make(map[string]string)
		if err := json.Unmarshal([]byte(freeformStr), &freeform); err != nil {
			log.With(zap.Error(err)).Errorf("failed to parse freeform tags provided for storageclass, freeformStr : %v", freeformStr)
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "failed to parse freeform tags provided for storageclass : %s", err.Error())
		}
		params.SCTags.FreeformTags = freeform
	}
	if definedStr, ok := parameters[initialDefinedTagsOverride]; ok && strings.TrimSpace(definedStr) != "" {
		defined := make(map[string]map[string]interface{})
		if err := json.Unmarshal([]byte(definedStr), &defined); err != nil {
			log.With(zap.Error(err)).Errorf("failed to parse defined tags provided for storageclass, definedStr : %v", definedStr)
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "failed to parse defined tags provided for storageclass : %s", err.Error())
		}
		params.SCTags.DefinedTags = defined
	}

	// availabilityDomain (mandatory)
	if ad, ok := parameters["availabilityDomain"]; ok && strings.TrimSpace(ad) != "" {
		// Validate with Identity if provided, and normalize to full AD name (tenancy-prefixed)
		adTrim := strings.TrimSpace(ad)
		if full, err := identityClient.GetAvailabilityDomainByName(ctx, params.CompartmentId, adTrim); err != nil {
			log.With(zap.Error(err)).Errorf("Invalid availabilityDomain: %s (from storage class) for compartment: %s", adTrim, params.CompartmentId)
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "Invalid availabilityDomain: %s for compartment %s, error: %v", adTrim, params.CompartmentId, err)
		} else {
			params.AvailabilityDomain = *full.Name
		}
		log = log.With("availabilityDomain", params.AvailabilityDomain)
		log.Info("AD is provided in storage class.")
	} else {
		log.Errorf("Missing required parameter: availabilityDomain")
		return log, nil, nil, status.Errorf(codes.InvalidArgument, "Missing required parameter: availabilityDomain")
	}

	log.Info("Successfully parsed Lustre storage class parameters")
	return log, nil, params, nil
}

// validateLustreFileSystemName enforces max 8 chars and allowed charset [A-Za-z0-9_]
func validateLustreFileSystemName(name string) error {
	if name == "" {
		return fmt.Errorf("empty")
	}
	if len(name) > 8 {
		return fmt.Errorf("length must be <= 8")
	}
	ok, _ := regexp.MatchString(`^[A-Za-z0-9_]+$`, name)
	if !ok {
		return fmt.Errorf("allowed characters are [A-Za-z0-9_]")
	}
	return nil
}

// deriveDefaultFSName picks the last 8 allowed characters from volumeName.
// If insufficient allowed chars, fall back to truncation with sanitization.
func deriveDefaultFSName(volumeName string) string {
	if volumeName == "" {
		return "pvlustre" //Just having default here, but volume name will not be empty as we check that in CreateVolume
	}
	// collect allowed chars from the end
	allowed := []rune{}
	for i := len(volumeName) - 1; i >= 0 && len(allowed) <= 6; i-- {
		ch := rune(volumeName[i])
		if (ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') {
			allowed = append([]rune{ch}, allowed...)
		}
	}

	fs := string(allowed)
	if len(fs) > 6 {
		fs = fs[len(fs)-6:]
	}
	fs = "pv" + fs //autogenerated names will start with pv
	// final sanity check
	if err := validateLustreFileSystemName(fs); err != nil {
		return "pvlustre"
	}
	return fs
}
