package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kubeAPI "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/disk"
	"github.com/oracle/oci-go-sdk/v50/core"
)

const (
	// Prefix to apply to the name of a created volume. This should be the same as the option '--volume-name-prefix' of csi-provisioner.
	pvcPrefix                     = "csi"
	csiDriver                     = "csi"
	fsTypeKey                     = "csi.storage.k8s.io/fstype"
	fsTypeKeyDeprecated           = "fstype"
	timeout                       = time.Minute * 3
	kmsKey                        = "kms-key-id"
	attachmentType                = "attachment-type"
	attachmentTypeISCSI           = "iscsi"
	attachmentTypeParavirtualized = "paravirtualized"
	initialFreeformTagsOverride   = "oci.oraclecloud.com/initial-freeform-tags-override"
	initialDefinedTagsOverride    = "oci.oraclecloud.com/initial-defined-tags-override"
	//device is the consistent device path that would be used for paravirtualized attachment
	device = "device"
)

var (
	// OCI currently only support a single node to be attached to a single node
	// in read/write mode. This corresponds to `accessModes.ReadWriteOnce` in a
	// PVC resource on Kubernetes
	supportedAccessMode = &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
)

// VolumeParameters holds configuration
type VolumeParameters struct {
	//kmsKey is the KMS key that would be used as CMEK key for BV attachment
	diskEncryptionKey   string
	attachmentParameter map[string]string
	// freeform tags to add for BVs
	freeformTags map[string]string
	// defined tags to add for BVs
	definedTags map[string]map[string]interface{}
	//volume performance units per gb describes the block volume performance level
	vpusPerGB int64
}

// VolumeAttachmentOption holds config for attachments
type VolumeAttachmentOption struct {
	//whether the attachment type is paravirtualized
	useParavirtualizedAttachment bool
	//whether to encrypt the compute to BV attachment as in-transit encryption.
	enableInTransitEncryption bool
}

func extractVolumeParameters(log *zap.SugaredLogger, parameters map[string]string) (VolumeParameters, error) {
	p := VolumeParameters{
		diskEncryptionKey:   "",
		attachmentParameter: make(map[string]string),
		vpusPerGB:           10, // default value is 10 -> Balanced
	}
	for k, v := range parameters {
		switch k {
		case fsTypeKeyDeprecated:
			log.Warnf("%s is deprecated, please use %s instead", fsTypeKeyDeprecated, fsTypeKey)
		case kmsKey:
			if v != "" {
				p.diskEncryptionKey = v
			}
		case attachmentType:
			attachmentTypeLower := strings.ToLower(v)
			if attachmentTypeLower != attachmentTypeISCSI && attachmentTypeLower != attachmentTypeParavirtualized {
				return p, status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid attachment-type: %s provided "+
					"for storageclass. supported attachment-types are %s and %s", v, attachmentTypeISCSI, attachmentTypeParavirtualized))
			}
			p.attachmentParameter[attachmentType] = attachmentTypeLower

		case initialFreeformTagsOverride:
			if v == "" {
				continue
			}
			freeformTags := make(map[string]string)
			err := json.Unmarshal([]byte(v), &freeformTags)
			if err != nil {
				return p, status.Errorf(codes.InvalidArgument, "failed to parse freeform tags provided "+
					"for storageclass. please check the parameters block on the storage class")
			}

			p.freeformTags = freeformTags
		case initialDefinedTagsOverride:
			if v == "" {
				continue
			}
			definedTags := make(map[string]map[string]interface{})
			err := json.Unmarshal([]byte(v), &definedTags)
			if err != nil {
				return p, status.Errorf(codes.InvalidArgument, "failed to parse defined tags provided "+
					"for storageclass. please check the parameters block on the storage class")
			}
			p.definedTags = definedTags

		case csi_util.VpusPerGB:
			vpusPerGB, err := csi_util.ExtractBlockVolumePerformanceLevel(v)
			if err != nil {
				return p, status.Error(codes.InvalidArgument, err.Error())
			}
			p.vpusPerGB = vpusPerGB
		}

	}
	return p, nil
}

// CreateVolume creates a new volume from the given request. The function is
// idempotent.
func (d *ControllerDriver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	startTime := time.Now()
	log := d.logger.With("volumeName", req.Name)
	var errorType string
	var csiMetricDimension string

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume Name must be provided")
	}

	if req.VolumeCapabilities == nil || len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapabilities must be provided in CreateVolumeRequest")
	}

	if !d.validateCapabilities(req.VolumeCapabilities) {
		return nil, status.Error(codes.InvalidArgument, "invalid volume capabilities requested. Only SINGLE_NODE_WRITER is supported ('accessModes.ReadWriteOnce' on Kubernetes)")
	}

	size, err := csi_util.ExtractStorage(req.CapacityRange)
	if err != nil {
		return nil, status.Errorf(codes.OutOfRange, "invalid capacity range: %v", err)
	}

	availableDomainShortName := ""
	if req.AccessibilityRequirements != nil && req.AccessibilityRequirements.Preferred != nil {
		for _, t := range req.AccessibilityRequirements.Preferred {
			availableDomainShortName, _ = t.Segments[kubeAPI.LabelZoneFailureDomain]
			log.With("AD", availableDomainShortName).Info("Using preferred topology for AD.")
			if len(availableDomainShortName) > 0 {
				break
			}
		}
	}

	if availableDomainShortName == "" {
		if req.AccessibilityRequirements != nil && req.AccessibilityRequirements.Requisite != nil {
			for _, t := range req.AccessibilityRequirements.Requisite {
				availableDomainShortName, _ = t.Segments[kubeAPI.LabelZoneFailureDomain]
				log.With("AD", availableDomainShortName).Info("Using requisite topology for AD.")
				if len(availableDomainShortName) > 0 {
					break
				}
			}
		}
	}

	volumeName := req.Name

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = volumeName

	if availableDomainShortName == "" {
		csiMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
		log.Error("Available domain short name is not found")
		return nil, status.Errorf(codes.InvalidArgument, "%s is required in PreferredTopologies or allowedTopologies", kubeAPI.LabelZoneFailureDomain)
	}

	//make sure this method is idempotent by checking existence of volume with same name.
	volumes, err := d.client.BlockStorage().GetVolumesByName(context.Background(), volumeName, d.config.CompartmentID)
	if err != nil {
		log.Error("Failed to find existence of volume %s", err)
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Internal, "failed to check existence of volume %v", err)
	}

	if len(volumes) > 1 {
		log.Error("Duplicate volume exists")
		csiMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, fmt.Errorf("duplicate volume %q exists", volumeName)
	}

	volumeParams, err := extractVolumeParameters(log, req.GetParameters())
	if err != nil {
		log.Error("Failed to parse storageclass parameters %s", err)
		csiMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse storageclass parameters %v", err)
	}

	provisionedVolume := core.Volume{}

	if len(volumes) > 0 {
		//Volume already exists so checking state of the volume and returning the same.
		log.Info("Volume already created!")
		//Assigning existing volume
		provisionedVolume = volumes[0]

	} else {
		// Creating new volume
		ad, err := d.client.Identity().GetAvailabilityDomainByName(context.Background(), d.config.CompartmentID, availableDomainShortName)
		if err != nil {
			log.With("Compartment Id", d.config.CompartmentID).Error("Failed to get available domain %s", err)
			errorType = util.GetError(err)
			csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.InvalidArgument, "invalid available domain: %s or compartment ID: %s", availableDomainShortName, d.config.CompartmentID)
		}

		// use initial tags for all BVs
		bvTags := &config.TagConfig{}
		if d.config.Tags != nil && d.config.Tags.BlockVolume != nil {
			bvTags = d.config.Tags.BlockVolume
		}

		// use storage class level tags if provided
		scTags := &config.TagConfig{
			FreeformTags: volumeParams.freeformTags,
			DefinedTags:  volumeParams.definedTags,
		}

		// storage class tags overwrite initial BV Tags
		if scTags.FreeformTags != nil || scTags.DefinedTags != nil {
			bvTags = scTags
		}

		provisionedVolume, err = provision(log, d.client, volumeName, size, *ad.Name, d.config.CompartmentID, "",
			volumeParams.diskEncryptionKey, volumeParams.vpusPerGB, timeout, bvTags)
		if err != nil {
			log.With("Ad name", *ad.Name, "Compartment Id", d.config.CompartmentID).Error("New volume creation failed %s", err)
			errorType = util.GetError(err)
			csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.Internal, "New volume creation failed %v", err.Error())
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	log.Info("Waiting for volume to become available.")
	_, err = d.client.BlockStorage().AwaitVolumeAvailableORTimeout(ctx, *provisionedVolume.Id)
	if err != nil {
		log.With("volumeName", volumeName).Error("Create volume failed with time out")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.DeadlineExceeded, "Create volume failed with time out %v", err.Error())
	}

	volumeOCID := volumeName
	if provisionedVolume.Id != nil {
		volumeOCID = *provisionedVolume.Id
	}
	log.With("volumeID", volumeOCID).Info("Volume is created.")
	csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	dimensionsMap[metrics.ResourceOCIDDimension] = volumeOCID
	metrics.SendMetricData(d.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      *provisionedVolume.Id,
			CapacityBytes: *provisionedVolume.SizeInMBs * client.MiB,
			AccessibleTopology: []*csi.Topology{
				{
					Segments: map[string]string{
						kubeAPI.LabelZoneFailureDomain: d.util.GetAvailableDomainInNodeLabel(*provisionedVolume.AvailabilityDomain),
					},
				},
			},

			VolumeContext: map[string]string{
				attachmentType:     volumeParams.attachmentParameter[attachmentType],
				csi_util.VpusPerGB: strconv.FormatInt(volumeParams.vpusPerGB, 10),
			},
		},
	}, nil
}

// DeleteVolume deletes the given volume. The function is idempotent.
func (d *ControllerDriver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	startTime := time.Now()
	log := d.logger.With("volumeID", req.VolumeId)
	var errorType string
	var csiMetricDimension string
	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.VolumeId

	if req.VolumeId == "" {
		log.Info("Unable to get Volume Id")
		csiMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDelete, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Error(codes.InvalidArgument, "DeleteVolume Volume ID must be provided")
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := d.client.BlockStorage().DeleteVolume(ctx, req.VolumeId)
	if err != nil && !apierrors.IsNotFound(err) {
		log.With(zap.Error(err)).Error("Failed to delete volume.")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDelete, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, fmt.Errorf("failed to delete volume, volumeId: %s, error: %v", req.VolumeId, err)
	}

	log.Info("Volume is deleted.")
	csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.PVDelete, time.Since(startTime).Seconds(), dimensionsMap)
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume attaches the given volume to the node
func (d *ControllerDriver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	startTime := time.Now()
	var errorType string
	var csiMetricDimension string

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Node ID must be provided")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability must be provided")
	}

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.VolumeId

	log := d.logger.With("volumeID", req.VolumeId, "nodeId", req.NodeId)

	id, err := d.util.LookupNodeID(d.KubeClient, req.NodeId)
	if err != nil {
		log.With(zap.Error(err)).Error("Failed to lookup node")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.InvalidArgument, "failed to get ProviderID by nodeName. error : %s", err)
	}
	id = client.MapProviderIDToInstanceID(id)

	//if the attachmentType is missing, default is iscsi
	attachType, ok := req.VolumeContext[attachmentType]
	if !ok {
		attachType = attachmentTypeISCSI
	}
	volumeAttachmentOptions, err := getAttachmentOptions(context.Background(), d.client.Compute(), attachType, id)
	if err != nil {
		log.With(zap.Error(err)).With("attachmentType", attachType, "instanceID", id).Error("failed to get the attachment options")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Unknown, "failed to get the attachment options. error : %s", err)
	}
	//in transit encryption is not supported for other attachment type than paravirtualized
	if volumeAttachmentOptions.enableInTransitEncryption && !volumeAttachmentOptions.useParavirtualizedAttachment {
		log.Error("node %s has in transit encryption enabled, but attachment type is not paravirtualized. invalid input", id)
		csiMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.InvalidArgument, "node %s has in transit encryption enabled, but attachment type is not paravirtualized. invalid input", id)
	}

	compartmentID, err := util.LookupNodeCompartment(d.KubeClient, req.NodeId)
	if err != nil {
		log.With(zap.Error(err)).With("instanceID", id).Errorf("failed to get compartmentID from node annotation: %s", util.CompartmentIDAnnotation)
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Unknown, "failed to get compartmentID from node annotation:. error : %s", err)
	}

	volumeAttached, err := d.client.Compute().FindActiveVolumeAttachment(context.Background(), compartmentID, req.VolumeId)

	if err != nil && !client.IsNotFound(err) {
		log.With(zap.Error(err)).Error("Got error in finding volume attachment: %s", err)
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, err
	}

	vpusPerGB, ok := req.VolumeContext[csi_util.VpusPerGB]
	if !ok || vpusPerGB == "" {
		log.Warnf("No vpusPerGB found in Volume Context falling back to balanced performance")
		vpusPerGB = "10"
	}

	// volume already attached to an instance
	if err == nil {
		if volumeAttached.GetLifecycleState() == core.VolumeAttachmentLifecycleStateDetaching {
			log.Info("Waiting for volume to get detached before attaching.")
			err = d.client.Compute().WaitForVolumeDetached(ctx, *volumeAttached.GetId())
			if err != nil {
				log.With(zap.Error(err)).Error("Error while waiting for volume to get detached before attaching: %s", err)
				errorType = util.GetError(err)
				csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
				dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
				metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
				return nil, status.Errorf(codes.Internal, "Error while waiting for volume to get detached before attaching: %s", err)
			}
		} else {
			if id != *volumeAttached.GetInstanceId() {
				log.Error("Volume is already attached to another node: %s", *volumeAttached.GetInstanceId())
				csiMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
				dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
				metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
				return nil, status.Errorf(codes.Internal, "Failed to attach volume to node. "+
					"The volume is already attached to another node.")
			}
			if volumeAttached.GetLifecycleState() == core.VolumeAttachmentLifecycleStateAttaching ||
				volumeAttached.GetLifecycleState() == core.VolumeAttachmentLifecycleStateAttached {
				log.Info("Volume is ATTACHING to node.")
				volumeAttached, err = d.client.Compute().WaitForVolumeAttached(ctx, *volumeAttached.GetId())
				if err != nil {
					log.With(zap.Error(err)).Error("Error while waiting: failed to attach volume to the node: %s.", err)
					errorType = util.GetError(err)
					csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
					dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
					metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
					return nil, status.Errorf(codes.Internal, "Failed to attach volume to the node: %s", err)
				}
				log.Info("Volume is already ATTACHED to node.")
				return generatePublishContext(volumeAttachmentOptions, log, volumeAttached, vpusPerGB), nil
			}
		}
	}

	log.Info("Attaching volume to instance")

	if volumeAttachmentOptions.useParavirtualizedAttachment {
		volumeAttached, err = d.client.Compute().AttachParavirtualizedVolume(context.Background(), id, req.VolumeId, volumeAttachmentOptions.enableInTransitEncryption)
		if err != nil {
			log.With(zap.Error(err)).Info("failed paravirtualized attachment instance to volume.")
			errorType = util.GetError(err)
			csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.Internal, "failed paravirtualized attachment instance to volume. error : %s", err)
		}
	} else {
		volumeAttached, err = d.client.Compute().AttachVolume(context.Background(), id, req.VolumeId)
		if err != nil {
			log.With(zap.Error(err)).Info("failed iscsi attachment instance to volume.")
			errorType = util.GetError(err)
			csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.Internal, "failed iscsi attachment instance to volume : %s", err)
		}
	}

	volumeAttached, err = d.client.Compute().WaitForVolumeAttached(ctx, *volumeAttached.GetId())
	if err != nil {
		log.With(zap.Error(err)).Error("Failed to attach volume to the node.")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Internal, "Failed to attach volume to the node %s", err)
	}

	csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
	return generatePublishContext(volumeAttachmentOptions, log, volumeAttached, vpusPerGB), nil

}

func generatePublishContext(volumeAttachmentOptions VolumeAttachmentOption, log *zap.SugaredLogger, volumeAttached core.VolumeAttachment, vpusPerGB string) *csi.ControllerPublishVolumeResponse {
	if volumeAttachmentOptions.useParavirtualizedAttachment {
		log.With("volumeAttachedId", *volumeAttached.GetId()).Info("Publishing paravirtualized Volume Completed.")
		return &csi.ControllerPublishVolumeResponse{
			PublishContext: map[string]string{
				attachmentType:     attachmentTypeParavirtualized,
				device:             *volumeAttached.GetDevice(),
				csi_util.VpusPerGB: vpusPerGB,
			},
		}
	}
	iSCSIVolumeAttached := volumeAttached.(core.IScsiVolumeAttachment)

	log.With("volumeAttachedId", *volumeAttached.GetId()).Info("Publishing iSCSI Volume Completed.")

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			attachmentType:     attachmentTypeISCSI,
			disk.ISCSIIQN:      *iSCSIVolumeAttached.Iqn,
			disk.ISCSIIP:       *iSCSIVolumeAttached.Ipv4,
			disk.ISCSIPORT:     strconv.Itoa(*iSCSIVolumeAttached.Port),
			csi_util.VpusPerGB: vpusPerGB,
		},
	}
}

// ControllerUnpublishVolume detaches the given volume from the node
func (d *ControllerDriver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	startTime := time.Now()
	log := d.logger.With("volumeID", req.VolumeId)
	var errorType string
	var csiMetricDimension string

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.VolumeId

	compartmentID, err := util.LookupNodeCompartment(d.KubeClient, req.NodeId)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Infof("Node with nodeID %s is not found, volume is likely already detached", req.NodeId)
			csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}
		log.With(zap.Error(err)).Errorf("failed to get compartmentID from node annotation: %s", util.CompartmentIDAnnotation)
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Unknown, "failed to get compartmentID from node annotation:: error : %s", err)
	}

	attachedVolume, err := d.client.Compute().FindVolumeAttachment(context.Background(), compartmentID, req.VolumeId)
	if err != nil {
		if client.IsNotFound(err) {
			log.With(zap.Error(err)).With("compartmentID", compartmentID).With("nodeId", req.NodeId).Error("Unable to find volume " +
				"attachment for volume to detach. Volume is possibly already detached. Nothing to do in Un-publish Volume.")
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}
		log.With(zap.Error(err)).With("nodeId", req.NodeId).Error("Volume is not detached from the node.")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, err
	}

	log.With("volumeAttachedId", attachedVolume.GetId()).Info("Detaching Volume.")
	err = d.client.Compute().DetachVolume(context.Background(), *attachedVolume.GetId())
	if err != nil {
		log.With(zap.Error(err)).With("nodeId", req.NodeId).Error("Volume can not be detached.")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Unknown, "volume can not be detached %s", err)
	}

	err = d.client.Compute().WaitForVolumeDetached(context.Background(), *attachedVolume.GetId())
	if err != nil {
		log.With(zap.Error(err)).With("nodeId", req.NodeId).Error("timed out waiting for volume to be detached.")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Unknown, "timed out waiting for volume to be detached %s", err)
	}

	log.With("volumeAttachedId", attachedVolume.GetId()).Info("Un-publishing Volume Completed.")
	csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities checks whether the volume capabilities requested
// are supported.
func (d *ControllerDriver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	log := d.logger.With("volumeID", req.VolumeId)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.VolumeCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities must be provided")
	}

	// check if volume exist before trying to validate it
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	volume, err := d.client.BlockStorage().GetVolume(ctx, req.VolumeId)
	if err != nil {
		log.With(zap.Error(err)).Error("Volume ID not found.")
		return nil, status.Errorf(codes.NotFound, "Volume ID not found.")
	}

	if *volume.Id == req.VolumeId {
		return &csi.ValidateVolumeCapabilitiesResponse{
			Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessMode: supportedAccessMode,
					},
				},
			},
		}, nil
	}
	return nil, status.Errorf(codes.NotFound, "VolumeId mis-match.")
}

// ListVolumes returns a list of all requested volumes
func (d *ControllerDriver) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity returns the capacity of the storage pool
func (d *ControllerDriver) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerGetCapabilities returns the capabilities of the controller service.
func (d *ControllerDriver) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
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
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
	} {
		caps = append(caps, newCap(cap))
	}

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}

	return resp, nil
}

// validateCapabilities validates the requested capabilities. It returns false
// if it doesn't satisfy the currently supported modes of OCI Block Volume
func (d *ControllerDriver) validateCapabilities(caps []*csi.VolumeCapability) bool {
	vcaps := []*csi.VolumeCapability_AccessMode{supportedAccessMode}

	hasSupport := func(mode csi.VolumeCapability_AccessMode_Mode) bool {
		for _, m := range vcaps {
			if mode == m.Mode {
				return true
			}
		}
		return false
	}

	supported := false
	for _, cap := range caps {
		if hasSupport(cap.AccessMode.Mode) {
			supported = true
		} else {
			// we need to make sure all capabilities are supported. Revert back
			// in case we have a cap that is supported, but is invalidated now
			d.logger.Errorf("The VolumeCapability isn't supported: %s", cap.AccessMode)
			supported = false
			break
		}
	}

	return supported
}

// CreateSnapshot will be called by the CO to create a new snapshot from a
// source volume on behalf of a user.
func (d *ControllerDriver) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "CreateSnapshot is not supported yet")
}

// DeleteSnapshot will be called by the CO to delete a snapshot.
func (d *ControllerDriver) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteSnapshot is not supported yet")
}

// ListSnapshots returns all the matched snapshots
func (d *ControllerDriver) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListSnapshots is not supported yet")
}

// ControllerExpandVolume returns ControllerExpandVolume request
func (d *ControllerDriver) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	startTime := time.Now()
	volumeId := req.GetVolumeId()
	if volumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "UpdateVolume volumeId must be provided")
	}
	log := d.logger.With("volumeID", volumeId)
	var errorType string
	var csiMetricDimension string

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.VolumeId

	newSize, err := csi_util.ExtractStorage(req.CapacityRange)
	if err != nil {
		log.Error("invalid capacity range: %v", err)
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVExpand, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.OutOfRange, "invalid capacity range: %v", err)
	}

	//make sure this method is idempotent by checking existence of volume with same name.
	volume, err := d.client.BlockStorage().GetVolume(context.Background(), volumeId)
	if err != nil {
		log.Error("Failed to find existence of volume %s", err)
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVExpand, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Internal, "failed to check existence of volume %v", err)
	}

	log = log.With("volumeName", volume.DisplayName)
	newSizeInGB := csi_util.RoundUpSize(newSize, 1*client.GiB)
	oldSize := *volume.SizeInGBs

	if newSizeInGB <= oldSize {
		log.Infof("Existing volume size: %v Requested volume size: %v No action needed.", *volume.SizeInGBs, newSizeInGB)
		return &csi.ControllerExpandVolumeResponse{
			CapacityBytes:         oldSize,
			NodeExpansionRequired: true,
		}, nil
	}

	updateVolumeDetails := core.UpdateVolumeDetails{
		DisplayName: volume.DisplayName,
		SizeInGBs:   &newSizeInGB,
	}

	volume, err = d.client.BlockStorage().UpdateVolume(ctx, volumeId, updateVolumeDetails)

	if err != nil {
		message := fmt.Sprintf("Update volume failed %v", err)
		log.Error(message)
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVExpand, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Error(codes.Internal, message)
	}

	log.Info("Volume is expanded.")
	csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.PVExpand, time.Since(startTime).Seconds(), dimensionsMap)

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         newSize,
		NodeExpansionRequired: true,
	}, nil
}

// ControllerGetVolume returns ControllerGetVolumeResponse response
func (d *ControllerDriver) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerGetVolume is not supported yet")
}

func provision(log *zap.SugaredLogger, c client.Interface, volName string, volSize int64, availDomainName, compartmentID,
	backupID, kmsKeyID string, vpusPerGB int64, timeout time.Duration, bvTags *config.TagConfig) (core.Volume, error) {

	ctx := context.Background()

	volSizeGB, minSizeGB := csi_util.RoundUpSize(volSize, 1*client.GiB), csi_util.RoundUpMinSize()

	if minSizeGB > volSizeGB {
		volSizeGB = minSizeGB
	}

	volumeDetails := core.CreateVolumeDetails{
		AvailabilityDomain: &availDomainName,
		CompartmentId:      &compartmentID,
		DisplayName:        &volName,
		SizeInGBs:          &volSizeGB,
		VpusPerGB:          &vpusPerGB,
	}

	if backupID != "" {
		volumeDetails.SourceDetails = &core.VolumeSourceFromVolumeBackupDetails{Id: &backupID}
	}

	if kmsKeyID != "" {
		volumeDetails.KmsKeyId = &kmsKeyID
	}
	if bvTags != nil && bvTags.FreeformTags != nil {
		volumeDetails.FreeformTags = bvTags.FreeformTags
	}
	if bvTags != nil && bvTags.DefinedTags != nil {
		volumeDetails.DefinedTags = bvTags.DefinedTags
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	newVolume, err := c.BlockStorage().CreateVolume(ctx, volumeDetails)

	if err != nil {
		log.With(zap.Error(err)).With("volumeName", volName).Error("Create volume failed.")
		status.Errorf(codes.Unknown, "Create volume failed")
		return core.Volume{}, err
	}
	log.With("volumeName", volName).Info("Volume is provisioned.")
	return *newVolume, nil
}

//We would derive whether the customer wants in-transit encryption or not based on if the node is launched using
//in-transit encryption enabled or not.
func getAttachmentOptions(ctx context.Context, client client.ComputeInterface, attachmentType, instanceID string) (VolumeAttachmentOption, error) {
	volumeAttachmentOption := VolumeAttachmentOption{}
	if attachmentType == attachmentTypeParavirtualized {
		volumeAttachmentOption.useParavirtualizedAttachment = true
	}
	instance, err := client.GetInstance(ctx, instanceID)
	if err != nil {
		return volumeAttachmentOption, err
	}
	if *instance.LaunchOptions.IsPvEncryptionInTransitEnabled {
		volumeAttachmentOption.enableInTransitEncryption = true
	}
	return volumeAttachmentOption, nil
}
