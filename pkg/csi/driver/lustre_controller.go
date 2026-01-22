package driver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	lustre "github.com/oracle/oci-go-sdk/v65/lustrefilestorage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/pointer"
)

var (
	lustreSupportedVolumeCapabilities = []csi.VolumeCapability_AccessMode{
		{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER},
		{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY},
		{Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY},
		{Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER},
		{Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER},
	}
)

const (
	// KB is 1000 bytes
	KB = 1000
	// MB is 1000 KB
	MB = 1000 * KB
	// GB is 1000 MB
	GB = 1000 * MB
)

// ControllerGetCapabilities advertises the controller RPCs supported by Lustre.
func (d *LustreControllerDriver) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	newCap := func(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}
	var caps []*csi.ControllerServiceCapability
	caps = append(caps, newCap(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME))
	return &csi.ControllerGetCapabilitiesResponse{Capabilities: caps}, nil
}

// CreateVolume implements CSI CreateVolume for Lustre.
func (d *LustreControllerDriver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (resp *csi.CreateVolumeResponse, err error) {
	defer MakeCSIPanicRecoveryWithError(d.logger, d.metricPusher, "LustreControllerDriver.CreateVolume", map[string]string{metrics.ResourceOCIDDimension: req.GetName()}, &err, codes.Internal)()
	startTime := time.Now()
	log := d.logger.With("csiOperation", "create", "volumeName", req.GetName())
	log.Debugf("CreateVolume request (lustre): %v", req)

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "Name must be provided in CreateVolumeRequest")
	}
	if caps := req.GetVolumeCapabilities(); caps == nil || len(caps) == 0 {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapabilities must be provided in CreateVolumeRequest")
	}
	if err := d.checkLustreSupportedVolumeCapabilities(req.GetVolumeCapabilities()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Requested Volume Capability not supported: %v", err)
	}

	// Prepare metrics dimensions
	metricDimensions := map[string]string{
		metrics.ResourceOCIDDimension: req.GetName(),
	}

	identityClient := d.client.Identity(nil)
	if identityClient == nil {
		return nil, status.Error(codes.Internal, "Unable to create identity client")
	}
	lustreClient := d.client.Lustre()
	if lustreClient == nil {
		return nil, status.Error(codes.Internal, "Unable to create lustre client")
	}

	// Parse StorageClass parameters
	log, _, sc, err := extractLustreStorageClassParameters(ctx, d, log, req.GetName(), req.GetParameters(), identityClient)
	if err != nil {
		metricDimensions[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.LustreProvision, time.Since(startTime).Seconds(), metricDimensions)
		return nil, err
	}

	// Merge cluster default tags with SC overrides
	if enableOkeSystemTags && util.IsCommonTagPresent(d.config.Tags) {
		sc.SCTags = util.MergeTagConfig(sc.SCTags, d.config.Tags.Common)
	}

	displayName := req.GetName()
	existingLustreFileSystems, err := lustreClient.ListLustreFileSystems(ctx, sc.CompartmentId, sc.AvailabilityDomain, displayName)
	if err != nil && !client.IsNotFound(err) {
		log.With("service", "lustre", "verb", "list", "resource", "lustreFilesystem", "statusCode", util.GetHttpStatusCode(err)).
			With(zap.Error(err)).Error("Failed to list Lustre file systems for idempotency check")
		metricDimensions[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.LustreProvision, time.Since(startTime).Seconds(), metricDimensions)
		return nil, status.Errorf(codes.Internal, "Failed to check existing file systems: %v", err)
	}

	if len(existingLustreFileSystems) > 1 {
		log.Errorf("Duplicate Lustre file systems with displayName %v exist", displayName)
		metricDimensions[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.LustreProvision, time.Since(startTime).Seconds(), metricDimensions)
		return nil, status.Errorf(codes.AlreadyExists, "Duplicate Lustre file systems with displayName %v exist", displayName)
	}

	if len(existingLustreFileSystems) == 1 {
		existingId := existingLustreFileSystems[0].Id
		existingLustreFs, err := lustreClient.GetLustreFileSystem(ctx, *existingId)
		if err != nil {
			metricDimensions[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.LustreProvision, time.Since(startTime).Seconds(), metricDimensions)
			log.Errorf("Failed to fetch existing filesystem: %v", err)
			return nil, status.Errorf(codes.Internal, "Failed to fetch existing filesystem: %v", err)
		}
		switch existingLustreFs.LifecycleState {
		case lustre.LustreFileSystemLifecycleStateActive:
			return d.sendCreateVolumeSuccessResponse(log, existingLustreFs, metricDimensions, startTime, sc)
		case lustre.LustreFileSystemLifecycleStateCreating:
			existingLustreFs, err = lustreClient.AwaitLustreFileSystemActive(ctx, log, *existingId)
			if err != nil || existingLustreFs == nil {
				metricDimensions[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
				metrics.SendMetricData(d.metricPusher, metrics.LustreProvision, time.Since(startTime).Seconds(), metricDimensions)
				return nil, status.Errorf(codes.DeadlineExceeded, "deadline reached while waiting for LustreFilesystem to become active, error : %v", err)
			}
			return d.sendCreateVolumeSuccessResponse(log, existingLustreFs, metricDimensions, startTime, sc)

		default:
			// Best-effort: if FAILED, try to fetch work request errors to append a reason
			errMsg := fmt.Errorf("LustreFileSystem provisioning failed. Filesystem is in %v state and not usable.", existingLustreFs.LifecycleState)
			if existingLustreFs.LifecycleState == lustre.LustreFileSystemLifecycleStateFailed {
				// Attempt to find the most recent CREATE_LUSTRE_FILE_SYSTEM work request for this filesystem to find out why it went into FAILED state
				var errorFromWorkRequest string
				wrs, lerr := d.client.Lustre().ListWorkRequests(ctx, sc.CompartmentId, *existingLustreFs.Id)
				if lerr == nil {
					for _, wr := range wrs {
						if wr.OperationType == lustre.OperationTypeCreateLustreFileSystem {
							// Fetch errors for this work request id
							if wr.Id != nil {
								errs, eerr := d.client.Lustre().ListWorkRequestErrors(ctx, *wr.Id, *existingLustreFs.Id)
								if eerr == nil && len(errs) > 0 {
									// pick the latest error message
									if errs[0].Message != nil {
										errorFromWorkRequest = *errs[0].Message
									}
								}
							}
							break
						}
					}
				}
				if errorFromWorkRequest != "" {
					log.With("workRequestError", errorFromWorkRequest).Error("Filesystem failed with work request error")
					errMsg = fmt.Errorf("%v Error from workrequest : %s", errMsg, errorFromWorkRequest)
				}
			}
			log.Error(errMsg)
			metricDimensions[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.LustreProvision, time.Since(startTime).Seconds(), metricDimensions)
			return nil, status.Error(codes.Aborted, errMsg.Error())
		}
	}

	capacityRange := req.GetCapacityRange()
	capacityInBytes := capacityRange.GetRequiredBytes()
	capacityInGbs := 31200 //Setting default capacity of 31200 GB
	if capacityInBytes > 0 {
		capacityInGbs = int(csi_util.RoundUpSize(req.CapacityRange.RequiredBytes, 1*GB))
	}
	log = log.With("Capacity", capacityInGbs)

	// Create new filesystem
	createRequestDetails := createLustreFilesystemRequest(displayName, sc, capacityInGbs)
	lustreFs, err := lustreClient.CreateLustreFileSystem(ctx, createRequestDetails)
	if err != nil {
		log.With("service", "lustre", "verb", "create", "resource", "lustreFilesystem", "statusCode", util.GetHttpStatusCode(err)).
			With(zap.Error(err)).Error("Lustre filesystem creation failed")
		metricDimensions[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.LustreProvision, time.Since(startTime).Seconds(), metricDimensions)
		return nil, status.Errorf(codes.Internal, "Lustre filesystem creation failed, error :  %v", err)
	}

	lustreFs, err = lustreClient.AwaitLustreFileSystemActive(ctx, log, *lustreFs.Id)
	if err != nil {
		log.With(zap.Error(err)).Error("Error occurred while waiting for LustreFilesystem to become active.")
		metricDimensions[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.LustreProvision, time.Since(startTime).Seconds(), metricDimensions)
		return nil, status.Errorf(codes.DeadlineExceeded, "deadline reached while waiting for LustreFilesystem to become active, error : %v", err)
	}

	return d.sendCreateVolumeSuccessResponse(log, lustreFs, metricDimensions, startTime, sc)
}

func createLustreFilesystemRequest(displayName string, sc *LustreStorageClassParameters, capacityInGbs int) lustre.CreateLustreFileSystemDetails {
	createDetails := lustre.CreateLustreFileSystemDetails{
		DisplayName:        &displayName,
		CompartmentId:      &sc.CompartmentId,
		AvailabilityDomain: &sc.AvailabilityDomain,
		FileSystemName:     &sc.FileSystemName,
		CapacityInGBs:      &capacityInGbs,
		SubnetId:           &sc.SubnetId,
		PerformanceTier:    lustre.CreateLustreFileSystemDetailsPerformanceTierEnum(sc.PerformanceTier),
		RootSquashConfiguration: &lustre.RootSquashConfiguration{
			IdentitySquash: func() lustre.RootSquashConfigurationIdentitySquashEnum {
				if sc.RootSquashEnabled {
					return lustre.RootSquashConfigurationIdentitySquashRoot
				}
				return lustre.RootSquashConfigurationIdentitySquashNone
			}(),
		},
	}
	if sc.RootSquashEnabled {
		if len(sc.RootSquashClientExceptions) > 0 {
			createDetails.RootSquashConfiguration.ClientExceptions = sc.RootSquashClientExceptions
		}
		if sc.RootSquashUidSpecified {
			createDetails.RootSquashConfiguration.SquashUid = func() *int64 { v := int64(sc.RootSquashUid); return &v }()
		}
		if sc.RootSquashGidSpecified {
			createDetails.RootSquashConfiguration.SquashGid = func() *int64 { v := int64(sc.RootSquashGid); return &v }()
		}
	}
	if sc.SCTags.FreeformTags != nil {
		createDetails.FreeformTags = sc.SCTags.FreeformTags
	}
	if sc.SCTags.DefinedTags != nil {
		createDetails.DefinedTags = sc.SCTags.DefinedTags
	}
	if sc.KmsKeyId != "" {
		createDetails.KmsKeyId = &sc.KmsKeyId
	}
	if len(sc.NSGIds) > 0 {
		createDetails.NsgIds = sc.NSGIds
	}
	return createDetails
}

func (d *LustreControllerDriver) sendCreateVolumeSuccessResponse(log *zap.SugaredLogger, lustreFs *lustre.LustreFileSystem, metricDimensions map[string]string, startTime time.Time, sc *LustreStorageClassParameters) (*csi.CreateVolumeResponse, error) {
	volumeHandle := buildLustreVolumeHandle(lustreFs)
	metricDimensions[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	metricDimensions[metrics.ResourceOCIDDimension] = *lustreFs.Id
	metrics.SendMetricData(d.metricPusher, metrics.LustreProvision, time.Since(startTime).Seconds(), metricDimensions)
	log.With("volumeID", *lustreFs.Id).With("volumeHandle", volumeHandle).Info("Lustre filesystem successfully created")
	volumeContext := map[string]string{}
	if sc.SetupLnet != "" {
		volumeContext["setupLnet"] = sc.SetupLnet
	}
	if sc.LustreSubnetCidr != "" {
		volumeContext["lustreSubnetCidr"] = sc.LustreSubnetCidr
	}
	if sc.LustrePostMountParameters != "" {
		volumeContext["lustrePostMountParameters"] = sc.LustrePostMountParameters
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeHandle,
			CapacityBytes: int64(*lustreFs.CapacityInGBs * GB),
			VolumeContext: volumeContext,
		},
	}, nil
}

// DeleteVolume implements CSI DeleteVolume RPC for Lustre Driver.
func (d *LustreControllerDriver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (resp *csi.DeleteVolumeResponse, err error) {
	defer MakeCSIPanicRecoveryWithError(d.logger, d.metricPusher, "LustreControllerDriver.DeleteVolume", map[string]string{metrics.ResourceOCIDDimension: req.GetVolumeId()}, &err, codes.Internal)()
	startTime := time.Now()
	log := d.logger.With("csiOperation", "delete", "volumeID", req.GetVolumeId())
	log.Debug("Request being passed in DeleteVolume gRPC ", req)
	if req.GetVolumeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	dim := make(map[string]string)

	lustreFilesystemId := extractLustreFilesystemId(req.GetVolumeId())
	if lustreFilesystemId == "" {
		dim[metrics.ResourceOCIDDimension] = req.GetVolumeId()
		dim[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.LustreDelete, time.Since(startTime).Seconds(), dim)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Volume ID provided %s", req.GetVolumeId())
	}
	dim[metrics.ResourceOCIDDimension] = lustreFilesystemId

	log = log.With("lustreFilesystemId", lustreFilesystemId)

	lustreClient := d.client.Lustre()
	if lustreClient == nil {
		return nil, status.Error(codes.Internal, "Unable to create lustre client")
	}

	log.Info("Getting lustre file system to be deleted")

	fs, err := lustreClient.GetLustreFileSystem(ctx, lustreFilesystemId)
	if err != nil {
		if client.IsNotFound(err) {
			log.Info("Lustre File system does not exist, returning deletion success.")
			return &csi.DeleteVolumeResponse{}, nil
		}
		log.With("service", "lustre", "verb", "get", "resource", "lustreFilesystem", "statusCode", util.GetHttpStatusCode(err)).
			With(zap.Error(err)).Error("Failed to get Lustre filesystem for deletion.")
		dim[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.LustreDelete, time.Since(startTime).Seconds(), dim)
		return nil, status.Errorf(codes.Internal, "Failed to get Lustre filesystem for deletion, error: %v", err.Error())
	}

	switch fs.LifecycleState {
	case lustre.LustreFileSystemLifecycleStateDeleted:
		log.Info("Lustre File system is in Deleted state, returning deletion success.")
		return &csi.DeleteVolumeResponse{}, nil
	case lustre.LustreFileSystemLifecycleStateDeleting:
		log.Info("Lustre File system is in Deleting state, waiting for deletion to complete.")
	default:
		if err := lustreClient.DeleteLustreFileSystem(ctx, lustreFilesystemId); err != nil {
			log.With("service", "lustre", "verb", "delete", "resource", "lustreFilesystem", "statusCode", util.GetHttpStatusCode(err)).
				With(zap.Error(err)).Error("Failed to delete Lustre filesystem")
			dim[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.LustreDelete, time.Since(startTime).Seconds(), dim)
			return nil, status.Errorf(codes.Internal, "Failed to delete Lustre filesystem, error :  %v", err)
		}
	}

	if err := lustreClient.AwaitLustreFileSystemDeleted(ctx, log, lustreFilesystemId); err != nil {
		dim[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.LustreDelete, time.Since(startTime).Seconds(), dim)
		return nil, status.Errorf(codes.DeadlineExceeded, "Error while waiting for Lustre Filesystem to be Deleted, error : %v", err.Error())
	}
	dim[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	metrics.SendMetricData(d.metricPusher, metrics.LustreDelete, time.Since(startTime).Seconds(), dim)
	return &csi.DeleteVolumeResponse{}, nil
}

// ValidateVolumeCapabilities implements CSI ValidateVolumeCapabilities for Lustre.
func (d *LustreControllerDriver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (resp *csi.ValidateVolumeCapabilitiesResponse, err error) {
	defer MakeCSIPanicRecoveryWithError(d.logger, d.metricPusher, "LustreControllerDriver.ValidateVolumeCapabilities", map[string]string{metrics.ResourceOCIDDimension: req.GetVolumeId()}, &err, codes.Internal)()
	if req.GetVolumeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapabilities must be provided")
	}
	if err := d.checkLustreSupportedVolumeCapabilities(req.GetVolumeCapabilities()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Requested Volume Capability not supported: %v", err)
	}

	log := d.logger.With("csiOperation", "validate", "volumeID", req.GetVolumeId())

	lustreFilesystemId := extractLustreFilesystemId(req.GetVolumeId())
	if lustreFilesystemId == "" {
		log.Errorf("Invalid Volume ID provided")
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Volume ID provided")
	}
	log = d.logger.With("lustreFilesystemId", lustreFilesystemId)

	lustreClient := d.client.Lustre()
	if lustreClient == nil {
		log.Error("Unable to create lustre client")
		return nil, status.Error(codes.Internal, "Unable to create lustre client")
	}
	lustreFilesystem, err := lustreClient.GetLustreFileSystem(ctx, lustreFilesystemId)
	if err != nil {
		log.With("service", "lustre", "verb", "get", "resource", "lustreFilesystem", "statusCode", util.GetHttpStatusCode(err)).
			With(zap.Error(err)).Error("Lustre filesystem not found")
		return nil, status.Errorf(codes.NotFound, "Lustre filesystem not found: %v", err)
	}

	// Verify identity consistency
	expected := buildLustreVolumeHandle(lustreFilesystem)
	if req.GetVolumeId() != expected {
		log = d.logger.With("Volume indentity mismatch. VolumeId from request :  %v, Volume Id from actual filesystem %v.", req.GetVolumeId(), expected)
		return nil, status.Errorf(codes.NotFound, "Volume identity mismatch")
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: req.GetVolumeCapabilities(),
		},
	}, nil
}

func (d *LustreControllerDriver) checkLustreSupportedVolumeCapabilities(volumeCaps []*csi.VolumeCapability) error {
	hasSupport := func(cap *csi.VolumeCapability) error {
		if blk := cap.GetBlock(); blk != nil {
			return status.Error(codes.InvalidArgument, "Lustre contoller driver does not support volume mode block")
		}
		for _, c := range lustreSupportedVolumeCapabilities {
			if c.GetMode() == cap.AccessMode.GetMode() {
				return nil
			}
		}
		return status.Errorf(codes.InvalidArgument, "Lustre contoller driver does not support access mode %v", cap.AccessMode.GetMode())
	}

	for _, c := range volumeCaps {
		if err := hasSupport(c); err != nil {
			return err
		}
	}
	return nil
}

func (d *LustreControllerDriver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (d *LustreControllerDriver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (d *LustreControllerDriver) ControllerModifyVolume(ctx context.Context, req *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (d *LustreControllerDriver) ListVolumes(ctx context.Context, request *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (d *LustreControllerDriver) GetCapacity(ctx context.Context, request *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (d *LustreControllerDriver) CreateSnapshot(ctx context.Context, request *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (d *LustreControllerDriver) DeleteSnapshot(ctx context.Context, request *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (d *LustreControllerDriver) ListSnapshots(ctx context.Context, request *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (d *LustreControllerDriver) ControllerExpandVolume(ctx context.Context, request *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
func (d *LustreControllerDriver) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// helpers
// buildLustreVolumeHandle composes the CSI volume id from  Lustre Filesystem object
// <lustreFileSystemOCID>:<managementServiceAddress>@<lnet>:/<fileSystemName>
func buildLustreVolumeHandle(lustreFilesystem *lustre.LustreFileSystem) string {
	id := pointer.StringDeref(lustreFilesystem.Id, "")
	managementServiceAddress := pointer.StringDeref(lustreFilesystem.ManagementServiceAddress, "")
	lnet := pointer.StringDeref(lustreFilesystem.Lnet, "")
	fsName := pointer.StringDeref(lustreFilesystem.FileSystemName, "")
	return fmt.Sprintf("%s:%s@%s:/%s", id, managementServiceAddress, lnet, fsName)
}

// extractLustreFilesystemId returns the OCID prefix from a Lustre volume id or empty if malformed.
// ex: volumeID => ocid1.lustrefilesystem.oc1.phx.aaaaaaa:10.0.1.10@tcp:/lustrefs
func extractLustreFilesystemId(volumeID string) string {
	if volumeID == "" || !strings.HasPrefix(volumeID, "ocid") {
		return ""
	}
	idx := strings.Index(volumeID, ":")
	if idx == -1 {
		return ""
	}
	return volumeID[:idx]
}
