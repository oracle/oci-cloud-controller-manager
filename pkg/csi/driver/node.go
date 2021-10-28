package driver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kubeAPI "k8s.io/api/core/v1"

	"github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/disk"
)

const (
	maxVolumesPerNode = 32
	volumeOperationAlreadyExistsFmt = "An operation for the volume: %s already exists."
)

// NodeStageVolume mounts the volume to a staging path on the node.
func (d *NodeDriver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.PublishContext == nil || len(req.PublishContext) == 0 {
		return nil, status.Error(codes.InvalidArgument, "PublishContext must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Staging Target Path must be provided")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capability must be provided")
	}

	logger := d.logger.With("volumeId", req.VolumeId, "stagingPath", req.StagingTargetPath)

	attachment, ok := req.PublishContext[attachmentType]

	if !ok {
		logger.Error("Unable to get the attachmentType from the attribute list, assuming iscsi")
		attachment = attachmentTypeISCSI
	}
	var devicePath string
	var mountHandler disk.Interface

	switch attachment {
	case attachmentTypeISCSI:
		scsiInfo, err := csi_util.ExtractISCSIInformation(req.PublishContext)
		if err != nil {
			logger.With(zap.Error(err)).Error("Failed to get SCSI info from publish context.")
			return nil, status.Error(codes.InvalidArgument, "PublishContext is invalid.")
		}

		// Get the device path using the publish context
		devicePath = csi_util.GetDevicePath(scsiInfo)

		mountHandler = disk.NewFromISCSIDisk(d.logger, scsiInfo)
		logger.With("devicePath", devicePath).Info("starting to stage iSCSI Mounting.")

	case attachmentTypeParavirtualized:
		devicePath, ok = req.PublishContext[device]
		if !ok {
			logger.Error("Unable to get the device from the attribute list")
			return nil, status.Error(codes.InvalidArgument, "Unable to get the device from the attribute list")
		}
		mountHandler = disk.NewFromPVDisk(d.logger)
		logger.With("devicePath", devicePath).Info("starting to stage paravirtualized Mounting.")
	default:
		logger.Error("unknown attachment type. supported attachment types are iscsi and paravirtualized")
		return nil, status.Error(codes.InvalidArgument, "unknown attachment type. supported attachment types are iscsi and paravirtualized")
	}

	if acquired := d.volumeLocks.TryAcquire(req.VolumeId); !acquired {
		logger.Error("Could not acquire lock for NodeStageVolume.")
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, req.VolumeId)
	}

	defer d.volumeLocks.Release(req.VolumeId)

	isMounted, oErr := mountHandler.DeviceOpened(devicePath)
	if oErr != nil {
		logger.With(zap.Error(oErr)).Error("getting error to get the details about volume is already mounted or not.")
		return nil, status.Error(codes.Internal, oErr.Error())
	} else if isMounted {
		logger.Info("volume is already mounted on the staging path.")
		return &csi.NodeStageVolumeResponse{}, nil
	}

	err := mountHandler.AddToDB()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to add the iSCSI node record.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = mountHandler.SetAutomaticLogin()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to set the iSCSI node to automatically login.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = mountHandler.Login()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to log into the iSCSI target.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !d.util.WaitForPathToExist(devicePath, 20) {
		logger.Error("failed to wait for device to exist.")
		return nil, status.Error(codes.DeadlineExceeded, "Failed to wait for device to exist.")
	}

	mnt := req.VolumeCapability.GetMount()
	options := mnt.MountFlags

	fsType := csi_util.ValidateFsType(logger, mnt.FsType)

	exists := true
	_, err = os.Stat(req.StagingTargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			exists = false
		} else {
			logger.With(zap.Error(err)).Errorf("failed to check if stagingTargetPath %q exists", req.StagingTargetPath)
			message := fmt.Sprintf("failed to check if stagingTargetPath %q exists", req.StagingTargetPath)
			return nil, status.Error(codes.Internal, message)
		}
	}

	// When exists is true it means target path was created but device isn't mounted.
	// We don't want to do anything in that case and let the operation proceed.
	// Otherwise we need to create the target directory.
	if !exists {
		if err := os.MkdirAll(req.StagingTargetPath, 0750); err != nil {
			logger.With(zap.Error(err)).Error("Failed to create StagingTargetPath directory")
			return nil, status.Error(codes.Internal, "Failed to create StagingTargetPath directory")
		}
	}

	logger.With("devicePath", devicePath,
		"fsType", fsType).Info("mounting the volume to staging path.")
	err = mountHandler.FormatAndMount(devicePath, req.StagingTargetPath, fsType, options)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to format and mount volume to staging path.")
		return nil, status.Error(codes.Internal, err.Error())
	}
	logger.With("devicePath", devicePath, "fsType", fsType, "attachmentType", attachment).
		Info("Mounting the volume to staging path is completed.")

	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume unstage the volume from the staging path
func (d *NodeDriver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Staging target path must be provided")
	}

	logger := d.logger.With("volumeId", req.VolumeId, "stagingPath", req.StagingTargetPath)

	if acquired := d.volumeLocks.TryAcquire(req.VolumeId); !acquired {
		logger.Error("Could not acquire lock for NodeUnstageVolume.")
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, req.VolumeId)
	}

	defer d.volumeLocks.Release(req.VolumeId)

	diskPath, err := disk.GetDiskPathFromMountPath(d.logger, req.GetStagingTargetPath())

	if err != nil {
		// do a clean exit in case of mount point not found
		if err == disk.ErrMountPointNotFound {
			logger.With(zap.Error(err)).With("mountPath", req.GetStagingTargetPath()).Warn("unable to fetch mount point")
			return &csi.NodeUnstageVolumeResponse{}, nil
		}
		logger.With(zap.Error(err)).With("mountPath", req.GetStagingTargetPath()).Error("unable to get diskPath from mount path")
		return nil, status.Error(codes.Internal, err.Error())
	}

	attachmentType, devicePath, err := getDevicePathAndAttachmentType(d.logger, diskPath)
	if err != nil {
		logger.With(zap.Error(err)).With("diskPath", diskPath).Error("unable to determine the attachment type")
		return nil, status.Error(codes.Internal, err.Error())
	}

	var mountHandler disk.Interface
	switch attachmentType {
	case attachmentTypeISCSI:
		scsiInfo, err := csi_util.ExtractISCSIInformationFromMountPath(d.logger, diskPath)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to ISCSI info.")
			return nil, status.Error(codes.Internal, err.Error())
		}
		if scsiInfo == nil {
			logger.Warn("unable to get the ISCSI info")
			return &csi.NodeUnstageVolumeResponse{}, nil
		}
		mountHandler = disk.NewFromISCSIDisk(d.logger, scsiInfo)
		logger.Info("starting to unstage iscsi Mounting.")
	case attachmentTypeParavirtualized:
		mountHandler = disk.NewFromPVDisk(d.logger)
		logger.Info("starting to unstage paravirtualized Mounting.")
	default:
		logger.Error("unknown attachment type. supported attachment types are iscsi and paravirtualized")
		return nil, status.Error(codes.InvalidArgument, "unknown attachment type. supported attachment types are iscsi and paravirtualized")
	}
	isMounted, oErr := mountHandler.DeviceOpened(devicePath)
	if oErr != nil {
		logger.With(zap.Error(oErr)).Error("getting error to get the details about volume is already mounted or not.")
		return nil, status.Error(codes.Internal, oErr.Error())
	} else if !isMounted {
		logger.Info("volume is already mounted on the staging path.")
		return &csi.NodeUnstageVolumeResponse{}, nil
	}

	err = mountHandler.UnmountPath(req.StagingTargetPath)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to unmount the staging path")
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = mountHandler.Logout()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to logout from the iSCSI target")
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = mountHandler.RemoveFromDB()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to remove the iSCSI node record")
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.With("devicePath", devicePath, "stagingPath",
		req.StagingTargetPath, "attachmentType", attachmentType).Info("Un-mounting the volume from staging path is completed.")
	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodePublishVolume mounts the volume to the target path
func (d *NodeDriver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.PublishContext == nil || len(req.PublishContext) == 0 {
		return nil, status.Error(codes.InvalidArgument, "PublishContext must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Staging path must be provided")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target Path must be provided")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capability must be provided")
	}

	logger := d.logger.With("volumeId", req.VolumeId, "targetPath", req.TargetPath)

	attachment, ok := req.PublishContext[attachmentType]
	if !ok {
		logger.Error("Unable to get the attachmentType from the attribute list, assuming iscsi")
		attachment = attachmentTypeISCSI
	}

	if acquired := d.volumeLocks.TryAcquire(req.VolumeId); !acquired {
		logger.Error("Could not acquire lock for NodePublishVolume.")
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, req.VolumeId)
	}

	defer d.volumeLocks.Release(req.VolumeId)

	// k8s v1.20+ will not create the TargetPath directory
	// https://github.com/kubernetes/kubernetes/pull/88759
	// if the path exists already (<v1.20) this is a no op
	// https://golang.org/pkg/os/#MkdirAll
	if err := os.MkdirAll(req.TargetPath, 0750); err != nil {
		logger.With(zap.Error(err)).Error("Failed to create TargetPath directory")
		return nil, status.Error(codes.Internal, "Failed to create TargetPath directory")
	}

	var mountHandler disk.Interface

	switch attachment {
	case attachmentTypeISCSI:
		scsiInfo, err := csi_util.ExtractISCSIInformation(req.PublishContext)
		if err != nil {
			logger.With(zap.Error(err)).Error("Failed to get iSCSI info from publish context")
			return nil, status.Error(codes.InvalidArgument, "PublishContext is invalid")
		}
		mountHandler = disk.NewFromISCSIDisk(logger, scsiInfo)
		logger.Info("starting to publish iSCSI Mounting.")

	case attachmentTypeParavirtualized:
		mountHandler = disk.NewFromPVDisk(d.logger)
		logger.Info("starting to publish paravirtualized Mounting.")
	default:
		logger.Error("unknown attachment type. supported attachment types are iscsi and paravirtualized")
		return nil, status.Error(codes.InvalidArgument, "unknown attachment type. supported attachment types are iscsi and paravirtualized")
	}

	mnt := req.VolumeCapability.GetMount()
	options := mnt.MountFlags

	options = append(options, "bind")
	if req.Readonly {
		options = append(options, "ro")
	}

	fsType := csi_util.ValidateFsType(logger, mnt.FsType)

	err := mountHandler.Mount(req.StagingTargetPath, req.TargetPath, fsType, options)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to format and mount.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.With("attachmentType", attachment).Info("Publish volume to the Node is Completed.")

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the volume from the target path
func (d *NodeDriver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Volume ID must be provided")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Target Path must be provided")
	}

	logger := d.logger.With("volumeId", req.VolumeId, "targetPath", req.TargetPath)

	if acquired := d.volumeLocks.TryAcquire(req.VolumeId); !acquired {
		logger.Error("Could not acquire lock for NodeUnpublishVolume.")
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, req.VolumeId)
	}

	defer d.volumeLocks.Release(req.VolumeId)

	diskPath, err := disk.GetDiskPathFromMountPath(d.logger, req.TargetPath)
	if err != nil {
		// do a clean exit in case of mount point not found
		if err == disk.ErrMountPointNotFound {
			logger.With(zap.Error(err)).With("mountPath", req.TargetPath).Warn("unable to fetch mount point")
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		logger.With(zap.Error(err)).With("mountPath", req.TargetPath).Error("unable to get diskPath from mount path")
		return nil, status.Error(codes.Internal, err.Error())
	}

	attachmentType, _, err := getDevicePathAndAttachmentType(d.logger, diskPath)
	if err != nil {
		logger.With(zap.Error(err)).With("diskPath", diskPath).Error("unable to determine the attachment type")
		return nil, status.Error(codes.Internal, err.Error())
	}

	var mountHandler disk.Interface
	switch attachmentType {
	case attachmentTypeISCSI:
		scsiInfo, _ := csi_util.ExtractISCSIInformationFromMountPath(d.logger, diskPath)
		if scsiInfo == nil {
			logger.Warn("unable to get the ISCSI info")
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		mountHandler = disk.NewFromISCSIDisk(d.logger, scsiInfo)
		d.logger.With("ISCSIInfo", scsiInfo, "mountPath", req.GetTargetPath()).Info("Found ISCSIInfo for NodeUnpublishVolume.")
	case attachmentTypeParavirtualized:
		mountHandler = disk.NewFromPVDisk(d.logger)
		logger.Info("starting to unpublish paravirtualized Mounting.")
	default:
		logger.Error("unknown attachment type. supported attachment types are iscsi and paravirtualized")
		return nil, status.Error(codes.InvalidArgument, "unknown attachment type. supported attachment types are iscsi and paravirtualized")
	}

	if err := mountHandler.UnmountPath(req.TargetPath); err != nil {
		logger.With(zap.Error(err)).Error("failed to unmount the target path, error")
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Info("Un-publish volume from the Node is Completed.")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func getDevicePathAndAttachmentType(logger *zap.SugaredLogger, path []string) (string, string, error) {
	for _, diskByPath := range path {
		matched, _ := regexp.MatchString(csi_util.DiskByPathPatternPV, diskByPath)
		if matched {
			return attachmentTypeParavirtualized, diskByPath, nil
		}
	}
	for _, diskByPath := range path {
		matched, _ := regexp.MatchString(csi_util.DiskByPathPatternISCSI, diskByPath)
		if matched {
			return attachmentTypeISCSI, diskByPath, nil
		}
	}

	return "", "", errors.New("unable to determine the attachment type")
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (d *NodeDriver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	nscap := &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
			},
		},
	}

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			nscap,
		},
	}, nil
}

// NodeGetInfo returns the supported capabilities of the node server.
// The result of this function will be used by the CO in ControllerPublishVolume.
func (d *NodeDriver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	ad, err := d.util.LookupNodeAvailableDomain(d.KubeClient, d.nodeID)

	if err != nil {
		d.logger.With(zap.Error(err)).With("nodeId", d.nodeID, "availableDomain", ad).Error("Available domain of node missing.")
	}

	d.logger.With("nodeId", d.nodeID, "availableDomain", ad).Info("Available domain of node identified.")
	return &csi.NodeGetInfoResponse{
		NodeId:            d.nodeID,
		MaxVolumesPerNode: maxVolumesPerNode,

		// make sure that the driver works on this particular AD only
		AccessibleTopology: &csi.Topology{
			Segments: map[string]string{
				kubeAPI.LabelZoneFailureDomain: ad,
			},
		},
	}, nil
}

// NodeGetVolumeStats return the stats of the volume
func (d *NodeDriver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetVolumeStats is not supported yet")
}

//NodeExpandVolume returns the expand of the volume
func (d *NodeDriver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeExpandVolume is not supported yet")
}
