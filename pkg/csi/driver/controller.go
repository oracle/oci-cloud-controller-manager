package driver

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/oracle/oci-go-sdk/core"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kubeAPI "k8s.io/api/core/v1"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/iscsi"
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

	timeout = time.Minute * 3
)

var (
	// DO currently only support a single node to be attached to a single node
	// in read/write mode. This corresponds to `accessModes.ReadWriteOnce` in a
	// PVC resource on Kubernetes
	supportedAccessMode = &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
)

// CreateVolume creates a new volume from the given request. The function is
// idempotent.
func (d *ControllerDriver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

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
		log.Error("Available domain short name is not found")
		return nil, status.Errorf(codes.InvalidArgument, "%s is required in PreferredTopologies or allowedTopologies", kubeAPI.LabelZoneFailureDomain)
	}

	volumeName := req.Name

	//make sure this method is idempotent by checking existence of volume with same name.
	volumes, err := d.client.BlockStorage().GetVolumesByName(context.Background(), volumeName, d.config.CompartmentID)
	if err != nil {
		log.Error("Failed to find existence of volume %s", err)
		return nil, status.Errorf(codes.Internal, "failed to check existence of volume %v", err)
	}

	if len(volumes) > 1 {
		log.Error("Duplicate volume exists")
		return nil, fmt.Errorf("duplicate volume %q exists", volumeName)
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
			return nil, status.Errorf(codes.InvalidArgument, "invalid available domain: %s or compartment ID: %s", availableDomainShortName, d.config.CompartmentID)
		}

		provisionedVolume, err = provision(log, d.client, volumeName, size, *ad.Name, d.config.CompartmentID, "", timeout)
		if err != nil {
			log.With("Ad name", *ad.Name, "Compartment Id", d.config.CompartmentID).Error("New volume creation failed %s", err)
			return nil, status.Errorf(codes.Internal, "New volume creation failed %v", err.Error())
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	_, err = d.client.BlockStorage().AwaitVolumeAvailableORTimeout(ctx, *provisionedVolume.Id)
	if err != nil {
		log.With("volumeName", volumeName).Error("Create volume failed with time out")
		status.Errorf(codes.DeadlineExceeded, "Create volume failed with time out")
		return nil, err
	}

	d.logger.With("volumeID", *provisionedVolume.Id).Info("Volume is created.")
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
		},
	}, nil
}

// DeleteVolume deletes the given volume. The function is idempotent.
func (d *ControllerDriver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	log := d.logger.With("volumeID", req.VolumeId)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "DeleteVolume Volume ID must be provided")
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := d.client.BlockStorage().DeleteVolume(ctx, req.VolumeId)
	if err != nil && !errors.IsNotFound(err) {
		log.With(zap.Error(err)).Error("Failed to delete volume.")
		return nil, fmt.Errorf("failed to delete volume, volumeId: %s, error: %v", req.VolumeId, err)
	}

	d.logger.With("volumeID", req.VolumeId).Info("Volume is deleted.")
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume attaches the given volume to the node
func (d *ControllerDriver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {

	log := d.logger.With("volumeID", req.VolumeId)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Node ID must be provided")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability must be provided")
	}

	id, err := d.util.lookupNodeID(d.KubeClient, req.NodeId)
	if err != nil {
		log.With(zap.Error(err)).With("nodeId", req.NodeId).Error("Failed to lookup node")
		return nil, status.Errorf(codes.InvalidArgument, "failed to get ProviderID by nodeName. error : %s", err)
	}
	id = client.MapProviderIDToInstanceID(id)

	volumeAttached, err := d.client.Compute().FindVolumeAttachment(context.Background(), d.config.CompartmentID, req.VolumeId)
	if err != nil {
		if client.IsNotFound(err) {
			volumeAttached, err = d.client.Compute().AttachVolume(context.Background(), id, req.VolumeId)
			if err != nil {
				d.logger.With(zap.Error(err)).With("nodeId", req.NodeId).Info("Failed to attach instance to node.")
				return nil, status.Errorf(codes.Unknown, "Failed to attach instance to node. error : %s", err)
			}
		} else {
			d.logger.With(zap.Error(err)).With("nodeId", req.NodeId).Error("Volume is not already attached to node.")
			return nil, err
		}
	}

	// Check if volumeAttached to another instance or not.
	if id != *volumeAttached.GetInstanceId() {
		d.logger.With("nodeId", req.NodeId).Error("Volume is already attached to another instance")
		return nil, status.Errorf(codes.Unknown, "Failed to attach instance to node. "+
			"The volume is already attached to another instance.")
	}

	volumeAttached, err = d.client.Compute().WaitForVolumeAttached(ctx, *volumeAttached.GetId())
	if err != nil {
		d.logger.With(zap.Error(err)).With("nodeId", req.NodeId).Error("Failed to attache volume to the node.")
		return nil, status.Errorf(codes.Unknown, "Failed to attach volume to the node %s", err)
	}

	iSCSIVolumeAttached := volumeAttached.(core.IScsiVolumeAttachment)
	log.With("volumeAttachedId", *volumeAttached.GetId()).Info("Publishing Volume Completed.")
	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			iscsi.ISCSIIQN:  *iSCSIVolumeAttached.Iqn,
			iscsi.ISCSIIP:   *iSCSIVolumeAttached.Ipv4,
			iscsi.ISCSIPORT: strconv.Itoa(*iSCSIVolumeAttached.Port),
		},
	}, nil

}

// ControllerUnpublishVolume deattaches the given volume from the node
func (d *ControllerDriver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {

	log := d.logger.With("volumeID", req.VolumeId)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	attachedVolume, err := d.client.Compute().FindVolumeAttachment(context.Background(), d.config.CompartmentID, req.VolumeId)
	if err != nil {
		if client.IsNotFound(err) {
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}
		log.With(zap.Error(err)).With("nodeId", req.NodeId).Error("Volume is not detached from the node.")
		return nil, err
	}

	log.With("volumeAttachedId", attachedVolume.GetId()).Info("Detaching Volume.")
	err = d.client.Compute().DetachVolume(context.Background(), *attachedVolume.GetId())
	if err != nil {
		log.With(zap.Error(err)).With("nodeId", req.NodeId).Error("Volume can not be detached.")
		return nil, status.Errorf(codes.Unknown, "volume can not be detached %s", err)
	}

	log.With("volumeAttachedId", attachedVolume.GetId()).Info("Un-publishing Volume Completed.")
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

func provision(log *zap.SugaredLogger, c client.Interface, volName string, volSize int64, availDomainName, compartmentID, backupID string, timeout time.Duration) (core.Volume, error) {

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

	if backupID != "" {
		volumeDetails.SourceDetails = &core.VolumeSourceFromVolumeBackupDetails{Id: &backupID}
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
