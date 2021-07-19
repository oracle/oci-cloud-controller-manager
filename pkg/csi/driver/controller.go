package driver

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/oracle/oci-go-sdk/v31/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kubeAPI "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/disk"
)

const (
	// minimumVolumeSizeInBytes is used to validate that the user is not trying
	// to create a volume that is smaller than what we support
	minimumVolumeSizeInBytes int64 = 50 * client.GiB

	// maximumVolumeSizeInBytes is used to validate that the user is not trying
	// to create a volume that is larger than what we support
	maximumVolumeSizeInBytes int64 = 32 * client.TiB

	// defaultVolumeSizeInBytes is used when the user did not provide a size or
	// the size they provided did not satisfy our requirements
	defaultVolumeSizeInBytes int64 = minimumVolumeSizeInBytes

	// Prefix to apply to the name of a created volume. This should be the same as the option '--volume-name-prefix' of csi-provisioner.
	pvcPrefix = "csi"
	csiDriver = "csi"

	timeout                       = time.Minute * 3
	kmsKey                        = "kms-key-id"
	attachmentType                = "attachment-type"
	attachmentTypeISCSI           = "iscsi"
	attachmentTypeParavirtualized = "paravirtualized"
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
}

// VolumeAttachmentOption holds config for attachments
type VolumeAttachmentOption struct {
	//whether the attachment type is paravirtualized
	useParavirtualizedAttachment bool
	//whether to encrypt the compute to BV attachment as in-transit encryption.
	enableInTransitEncryption bool
}

func extractVolumeParameters(parameters map[string]string) (VolumeParameters, error) {
	p := VolumeParameters{
		diskEncryptionKey:   "",
		attachmentParameter: make(map[string]string),
	}
	for k, v := range parameters {
		switch k {
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
		}
	}
	return p, nil
}

// CreateVolume creates a new volume from the given request. The function is
// idempotent.
func (d *ControllerDriver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	startTime := time.Now()
	log := d.logger.With("volumeName", req.Name)

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume Name must be provided")
	}

	if req.VolumeCapabilities == nil || len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapabilities must be provided in CreateVolumeRequest")
	}

	if !d.validateCapabilities(req.VolumeCapabilities) {
		return nil, status.Error(codes.InvalidArgument, "invalid volume capabilities requested. Only SINGLE_NODE_WRITER is supported ('accessModes.ReadWriteOnce' on Kubernetes)")
	}

	size, err := extractStorage(req.CapacityRange)
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

	if availableDomainShortName == "" {
		metrics.SendMetricData(d.metricPusher, metrics.PVProvisionFailure, time.Since(startTime).Seconds(), csiDriver, "")
		log.Error("Available domain short name is not found")
		return nil, status.Errorf(codes.InvalidArgument, "%s is required in PreferredTopologies or allowedTopologies", kubeAPI.LabelZoneFailureDomain)
	}

	volumeName := req.Name

	//make sure this method is idempotent by checking existence of volume with same name.
	volumes, err := d.client.BlockStorage().GetVolumesByName(context.Background(), volumeName, d.config.CompartmentID)
	if err != nil {
		log.Error("Failed to find existence of volume %s", err)
		metrics.SendMetricData(d.metricPusher, metrics.PVProvisionFailure, time.Since(startTime).Seconds(), csiDriver, "")
		return nil, status.Errorf(codes.Internal, "failed to check existence of volume %v", err)
	}

	if len(volumes) > 1 {
		log.Error("Duplicate volume exists")
		metrics.SendMetricData(d.metricPusher, metrics.PVProvisionFailure, time.Since(startTime).Seconds(), csiDriver, "")
		return nil, fmt.Errorf("duplicate volume %q exists", volumeName)
	}

	volumeParams, err := extractVolumeParameters(req.GetParameters())
	if err != nil {
		log.Error("Failed to parse storageclass parameters %s", err)
		metrics.SendMetricData(d.metricPusher, metrics.PVProvisionFailure, time.Since(startTime).Seconds(), csiDriver, "")
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse storageclass parameters %v", err)
	}

	var sourceOcid string
	sourceVolume := req.GetVolumeContentSource()
	log.Info("Volume Content Source: ", sourceVolume)

	if sourceVolume != nil {
		sourceOcid = sourceVolume.GetVolume().VolumeId
		if sourceOcid != "" {
			log.Info("Source Volume Ocid obtained: ", sourceOcid)
		}

		source, err := d.client.BlockStorage().GetVolume(ctx, sourceOcid)
		if err != nil {
			log.Errorf("Error finding source details", err)
		}

		// Size Validation
		if *source.SizeInMBs*client.MiB < req.CapacityRange.RequiredBytes {
			log.Error("Volume Expansion is not supported : Source Volume Size must be equal to the requested volume size")
			return nil, fmt.Errorf("Requested volume size is greater than the source volume")
		}
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
			metrics.SendMetricData(d.metricPusher, metrics.PVProvisionFailure, time.Since(startTime).Seconds(), csiDriver, "")
			return nil, status.Errorf(codes.InvalidArgument, "invalid available domain: %s or compartment ID: %s", availableDomainShortName, d.config.CompartmentID)
		}

		provisionedVolume, err = provision(log, d.client, volumeName, size, sourceOcid, *ad.Name, d.config.CompartmentID, "", volumeParams.diskEncryptionKey, timeout)
		if err != nil {
			log.With("Ad name", *ad.Name, "Compartment Id", d.config.CompartmentID).Error("New volume creation failed %s", err)
			metrics.SendMetricData(d.metricPusher, metrics.PVProvisionFailure, time.Since(startTime).Seconds(), csiDriver, "")
			return nil, status.Errorf(codes.Internal, "New volume creation failed %v", err.Error())
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	_, err = d.client.BlockStorage().AwaitVolumeAvailableORTimeout(ctx, *provisionedVolume.Id)
	if err != nil {
		log.With("volumeName", volumeName).Error("Create volume failed with time out")
		metrics.SendMetricData(d.metricPusher, metrics.PVProvisionFailure, time.Since(startTime).Seconds(), csiDriver, "")
		status.Errorf(codes.DeadlineExceeded, "Create volume failed with time out")
		return nil, err
	}

	volumeOCID := ""
	if provisionedVolume.Id != nil {
		volumeOCID = *provisionedVolume.Id
	}
	log.With("volumeID", volumeOCID).Info("Volume is created.")
	metrics.SendMetricData(d.metricPusher, metrics.PVProvisionSuccess, time.Since(startTime).Seconds(), csiDriver, volumeOCID)

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      *provisionedVolume.Id,
			CapacityBytes: *provisionedVolume.SizeInMBs * client.MiB,
			AccessibleTopology: []*csi.Topology{
				{
					Segments: map[string]string{
						kubeAPI.LabelZoneFailureDomain: d.util.getAvailableDomainInNodeLabel(*provisionedVolume.AvailabilityDomain),
					},
				},
			},
			VolumeContext: volumeParams.attachmentParameter,
			ContentSource: sourceVolume,
		},
	}, nil
}

// DeleteVolume deletes the given volume. The function is idempotent.
func (d *ControllerDriver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	startTime := time.Now()
	log := d.logger.With("volumeID", req.VolumeId)

	if req.VolumeId == "" {
		metrics.SendMetricData(d.metricPusher, metrics.PVDeleteFailure, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
		return nil, status.Error(codes.InvalidArgument, "DeleteVolume Volume ID must be provided")
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := d.client.BlockStorage().DeleteVolume(ctx, req.VolumeId)
	if err != nil && !errors.IsNotFound(err) {
		log.With(zap.Error(err)).Error("Failed to delete volume.")
		metrics.SendMetricData(d.metricPusher, metrics.PVDeleteFailure, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
		return nil, fmt.Errorf("failed to delete volume, volumeId: %s, error: %v", req.VolumeId, err)
	}

	log.Info("Volume is deleted.")
	metrics.SendMetricData(d.metricPusher, metrics.PVDeleteSuccess, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume attaches the given volume to the node
func (d *ControllerDriver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	// startTime := time.Now()

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Node ID must be provided")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability must be provided")
	}

	log := d.logger.With("volumeID", req.VolumeId, "nodeId", req.NodeId)

	id, err := d.util.lookupNodeID(d.KubeClient, req.NodeId)
	if err != nil {
		log.With(zap.Error(err)).Error("Failed to lookup node")
		// metrics.SendMetricData(d.metricPusher, pvAttachFailureMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
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
		return nil, status.Errorf(codes.Unknown, "failed to get the attachment options. error : %s", err)
	}

	compartmentID, err := util.LookupNodeCompartment(d.KubeClient, req.NodeId)
	if err != nil {
		log.With(zap.Error(err)).With("instanceID", id).Errorf("failed to get compartmentID from node annotation: %s", util.CompartmentIDAnnotation)
		return nil, status.Errorf(codes.Unknown, "failed to get compartmentID from node annotation:. error : %s", err)
	}

	volumeAttached, err := d.client.Compute().FindActiveVolumeAttachment(context.Background(), compartmentID, req.VolumeId)

	if err != nil && !client.IsNotFound(err) {
		log.With(zap.Error(err)).Error("Got error in finding volume attachment: %s", err)
		// metrics.SendMetricData(d.metricPusher, pvAttachFailureMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
		return nil, err
	}

	// volume already attached to an instance
	if err == nil {
		if volumeAttached.GetLifecycleState() == core.VolumeAttachmentLifecycleStateDetaching {
			log.Info("Waiting for volume to get detached before attaching.")
			err = d.client.Compute().WaitForVolumeDetached(ctx, *volumeAttached.GetId())
			if err != nil {
				log.With(zap.Error(err)).Error("Error while waiting for volume to get detached before attaching: %s", err)
				// metrics.SendMetricData(d.metricPusher, pvAttachFailureMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
				return nil, status.Errorf(codes.Internal, "Error while waiting for volume to get detached before attaching: %s", err)
			}
		} else {
			if id != *volumeAttached.GetInstanceId() {
				log.Error("Volume is already attached to another node: %s", *volumeAttached.GetInstanceId())
				// metrics.SendMetricData(d.metricPusher, pvAttachFailureMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
				return nil, status.Errorf(codes.Internal, "Failed to attach volume to node. "+
					"The volume is already attached to another node.")
			}
			if volumeAttached.GetLifecycleState() == core.VolumeAttachmentLifecycleStateAttaching {
				log.Info("Volume is ATTACHING to node.")
				volumeAttached, err = d.client.Compute().WaitForVolumeAttached(ctx, *volumeAttached.GetId())
				if err != nil {
					log.With(zap.Error(err)).Error("Error while waiting: failed to attach volume to the node: %s.", err)
					// metrics.SendMetricData(d.metricPusher, pvAttachFailureMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
					return nil, status.Errorf(codes.Internal, "Failed to attach volume to the node: %s", err)
				}
				log.Info("Volume is already ATTACHED to node.")
				return generatePublishContext(volumeAttachmentOptions, log, volumeAttached), nil
			}
		}
	}

	log.Info("Attaching volume to instance")

	if volumeAttachmentOptions.useParavirtualizedAttachment {
		volumeAttached, err = d.client.Compute().AttachParavirtualizedVolume(context.Background(), id, req.VolumeId, volumeAttachmentOptions.enableInTransitEncryption)
		if err != nil {
			log.With(zap.Error(err)).Info("failed paravirtualized attachment instance to volume.")
			// metrics.SendMetricData(d.metricPusher, pvAttachFailureMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
			return nil, status.Errorf(codes.Internal, "failed paravirtualized attachment instance to volume. error : %s", err)
		}
	} else {
		volumeAttached, err = d.client.Compute().AttachVolume(context.Background(), id, req.VolumeId)
		if err != nil {
			log.With(zap.Error(err)).Info("failed iscsi attachment instance to volume.")
			// metrics.SendMetricData(d.metricPusher, pvAttachFailureMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
			return nil, status.Errorf(codes.Internal, "failed iscsi attachment instance to volume : %s", err)
		}
	}

	volumeAttached, err = d.client.Compute().WaitForVolumeAttached(ctx, *volumeAttached.GetId())
	if err != nil {
		log.With(zap.Error(err)).Error("Failed to attach volume to the node.")
		// metrics.SendMetricData(d.metricPusher, pvAttachFailureMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
		return nil, status.Errorf(codes.Internal, "Failed to attach volume to the node %s", err)
	}

	// metrics.SendMetricData(d.metricPusher, pvAttachSuccessMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
	return generatePublishContext(volumeAttachmentOptions, log, volumeAttached), nil

}

func generatePublishContext(volumeAttachmentOptions VolumeAttachmentOption, log *zap.SugaredLogger, volumeAttached core.VolumeAttachment) *csi.ControllerPublishVolumeResponse {
	if volumeAttachmentOptions.useParavirtualizedAttachment {
		log.With("volumeAttachedId", *volumeAttached.GetId()).Info("Publishing paravirtualized Volume Completed.")
		return &csi.ControllerPublishVolumeResponse{
			PublishContext: map[string]string{
				attachmentType: attachmentTypeParavirtualized,
				device:         *volumeAttached.GetDevice(),
			},
		}
	}
	iSCSIVolumeAttached := volumeAttached.(core.IScsiVolumeAttachment)

	log.With("volumeAttachedId", *volumeAttached.GetId()).Info("Publishing iSCSI Volume Completed.")

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			attachmentType: attachmentTypeISCSI,
			disk.ISCSIIQN:  *iSCSIVolumeAttached.Iqn,
			disk.ISCSIIP:   *iSCSIVolumeAttached.Ipv4,
			disk.ISCSIPORT: strconv.Itoa(*iSCSIVolumeAttached.Port),
		},
	}
}

// ControllerUnpublishVolume deattaches the given volume from the node
func (d *ControllerDriver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	// startTime := time.Now()
	log := d.logger.With("volumeID", req.VolumeId)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	compartmentID, err := util.LookupNodeCompartment(d.KubeClient, req.NodeId)
	if err != nil {
		log.With(zap.Error(err)).Errorf("failed to get compartmentID from node annotation: %s", util.CompartmentIDAnnotation)
		return nil, status.Errorf(codes.Unknown, "failed to get compartmentID from node annotation:. error : %s", err)
	}

	attachedVolume, err := d.client.Compute().FindVolumeAttachment(context.Background(), compartmentID, req.VolumeId)
	if err != nil {
		if client.IsNotFound(err) {
			log.With(zap.Error(err)).With("compartmentID", compartmentID).With("nodeId", req.NodeId).Error("Unable to find volume " +
				"attachment for volume to detach. Volume is possibly already detached. Nothing to do in Un-publish Volume.")
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}
		log.With(zap.Error(err)).With("nodeId", req.NodeId).Error("Volume is not detached from the node.")
		// metrics.SendDataMetric(d.metricPusher, pvDetachFailureMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
		return nil, err
	}

	log.With("volumeAttachedId", attachedVolume.GetId()).Info("Detaching Volume.")
	err = d.client.Compute().DetachVolume(context.Background(), *attachedVolume.GetId())
	if err != nil {
		log.With(zap.Error(err)).With("nodeId", req.NodeId).Error("Volume can not be detached.")
		// metrics.SendDataMetric(d.metricPusher, pvDetachFailureMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
		return nil, status.Errorf(codes.Unknown, "volume can not be detached %s", err)
	}

	err = d.client.Compute().WaitForVolumeDetached(context.Background(), *attachedVolume.GetId())
	if err != nil {
		log.With(zap.Error(err)).With("nodeId", req.NodeId).Error("timed out waiting for volume to be detached.")
		return nil, status.Errorf(codes.Unknown, "timed out waiting for volume to be detached %s", err)
	}

	log.With("volumeAttachedId", attachedVolume.GetId()).Info("Un-publishing Volume Completed.")

	// metrics.SendMetricData(d.metricPusher, pvDetachSuccessMetric, time.Since(startTime).Seconds(), csiDriver, req.VolumeId)
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
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
	} {
		caps = append(caps, newCap(cap))
	}

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}

	return resp, nil
}

// extractStorage extracts the storage size in bytes from the given capacity
// range. If the capacity range is not satisfied it returns the default volume
// size. If the capacity range is below or above supported sizes, it returns an
// error.
func extractStorage(capRange *csi.CapacityRange) (int64, error) {
	if capRange == nil {
		return defaultVolumeSizeInBytes, nil
	}

	requiredBytes := capRange.GetRequiredBytes()
	requiredSet := 0 < requiredBytes
	limitBytes := capRange.GetLimitBytes()
	limitSet := 0 < limitBytes

	if !requiredSet && !limitSet {
		return defaultVolumeSizeInBytes, nil
	}

	if requiredSet && limitSet && limitBytes < requiredBytes {
		return 0, fmt.Errorf("limit (%v) can not be less than required (%v) size", formatBytes(limitBytes), formatBytes(requiredBytes))
	}

	if requiredSet && !limitSet {
		return maxOfInt(requiredBytes, minimumVolumeSizeInBytes), nil
	}

	if limitSet {
		return maxOfInt(limitBytes, minimumVolumeSizeInBytes), nil
	}

	if requiredSet && requiredBytes > maximumVolumeSizeInBytes {
		return 0, fmt.Errorf("required (%v) can not exceed maximum supported volume size (%v)", formatBytes(requiredBytes), formatBytes(maximumVolumeSizeInBytes))
	}

	if !requiredSet && limitSet && limitBytes > maximumVolumeSizeInBytes {
		return 0, fmt.Errorf("limit (%v) can not exceed maximum supported volume size (%v)", formatBytes(limitBytes), formatBytes(maximumVolumeSizeInBytes))
	}

	if requiredSet && limitSet {
		return maxOfInt(requiredBytes, limitBytes), nil
	}

	if requiredSet {
		return requiredBytes, nil
	}

	if limitSet {
		return limitBytes, nil
	}

	return defaultVolumeSizeInBytes, nil
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
	return nil, status.Error(codes.Unimplemented, "ControllerExpandVolume is not supported yet")
}

func provision(log *zap.SugaredLogger, c client.Interface, volName string, volSize int64, sourceOcid string, availDomainName, compartmentID, backupID, kmsKeyID string, timeout time.Duration) (core.Volume, error) {

	ctx := context.Background()

	volSizeGB := roundUpSize(volSize, 1*client.GiB)
	minSizeGB := roundUpSize(minimumVolumeSizeInBytes, 1*client.GiB)

	if minSizeGB > volSizeGB {
		volSizeGB = minSizeGB
	}

	volumeDetails := core.CreateVolumeDetails{
		AvailabilityDomain: &availDomainName,
		CompartmentId:      &compartmentID,
		DisplayName:        &volName,
		SizeInGBs:          &volSizeGB,
	}

	if sourceOcid != "" {
		volumeDetails.SourceDetails = &core.VolumeSourceFromVolumeDetails{Id: &sourceOcid}
		if volumeDetails.SourceDetails != nil {
			log.Info("Source Ocid added to volumeDetails: ", sourceOcid)
		}
	}

	if backupID != "" {
		volumeDetails.SourceDetails = &core.VolumeSourceFromVolumeBackupDetails{Id: &backupID}
	}

	if kmsKeyID != "" {
		volumeDetails.KmsKeyId = &kmsKeyID
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

func roundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	return (volumeSizeBytes + allocationUnitBytes - 1) / allocationUnitBytes
}

//We would derive whether the customer wants in-transit encryption or not based on if the node is launched using
//in-transit encryption enabled or not. we won't give our customers any explicit option to enable in-transit encryption
//via storage class. In the storage class if the customer opts for iscsi attachment and if the node is launched using
//in-transit encryption, we would still make the attachment type as paravirtualized and enable in-transit encryption. Refer
//https://confluence.oci.oraclecorp.com/pages/viewpage.action?spaceKey=OKE&title=Customer+Managed+Encryption+key+and+In-Transit+encryption+support+for+block+volume+for+CSI#CustomerManagedEncryptionkeyandIn-TransitencryptionsupportforblockvolumeforCSI-SupportforIn-TransitEncryptioninPhase1
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
		volumeAttachmentOption.useParavirtualizedAttachment = true
	}
	return volumeAttachmentOption, nil
}
