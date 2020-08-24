package driver

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/iscsi"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kubeAPI "k8s.io/api/core/v1"
)

const (
	maxVolumesPerNode = 32
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

	scsiInfo, err := extractISCSIInformation(req.PublishContext)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to get SCSI info from publish context.")
		return nil, status.Error(codes.InvalidArgument, "PublishContext is invalid.")
	}

	// Get the device path using the publish context
	devicePath := getDevicePath(scsiInfo)

	iSCSIMounter := iscsi.NewFromISCSIDisk(d.logger, scsiInfo)

	isMounted, oErr := iSCSIMounter.DeviceOpened(devicePath)
	if oErr != nil {
		logger.With(zap.Error(oErr)).Error("getting error to get the details about volume is already mounted or not.")
		return nil, status.Error(codes.Internal, oErr.Error())
	} else if isMounted {
		logger.Info("volume is already mounted on the staging path.")
		return &csi.NodeStageVolumeResponse{}, nil
	}

	err = iSCSIMounter.AddToDB()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to add the iSCSI node record.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = iSCSIMounter.SetAutomaticLogin()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to set the iSCSI node to automatically login.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = iSCSIMounter.Login()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to log into the iSCSI target.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !d.util.waitForPathToExist(devicePath, 20) {
		logger.Error("failed to wait for device to exist.")
		return nil, status.Error(codes.DeadlineExceeded, "Failed to wait for device to exist.")
	}

	mnt := req.VolumeCapability.GetMount()
	options := mnt.MountFlags

	fsType := validateFsType(logger, mnt.FsType)

	logger.With("device path", devicePath, "staging target path", req.StagingTargetPath,
		"fs type", fsType).Info("mounting the volume to staging path.")
	err = iSCSIMounter.FormatAndMount(devicePath, req.StagingTargetPath, fsType, options)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to format and mount volume to staging path.")
		return nil, status.Error(codes.Internal, err.Error())
	}
	logger.With("device path", devicePath, "staging target path", req.StagingTargetPath,
		"fs type", fsType).Info("Mounting the volume to staging path is completed.")
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

	scsiInfo, err := extractISCSIInformationFromMountPath(d.logger, req.GetStagingTargetPath())
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to ISCSI info.")
		return nil, status.Error(codes.Internal, err.Error())
	}
	if scsiInfo == nil {
		logger.Warn("unable to get the ISCSI info")
		return &csi.NodeUnstageVolumeResponse{}, nil
	}

	iSCSIMounter := iscsi.NewFromISCSIDisk(d.logger, scsiInfo)

	devicePath := getDevicePath(scsiInfo)

	isMounted, oErr := iSCSIMounter.DeviceOpened(devicePath)
	if oErr != nil {
		logger.With(zap.Error(oErr)).Error("getting error to get the details about volume is already mounted or not.")
		return nil, status.Error(codes.Internal, oErr.Error())
	} else if !isMounted {
		logger.Info("volume is already mounted on the staging path.")
		return &csi.NodeUnstageVolumeResponse{}, nil
	}

	err = iSCSIMounter.UnmountPath(req.StagingTargetPath)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to unmount the staging path")
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = iSCSIMounter.Logout()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to logout from the iSCSI target")
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = iSCSIMounter.RemoveFromDB()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to remove the iSCSI node record")
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.With("device path", devicePath, "staging target path",
		req.StagingTargetPath).Info("Un-mounting the volume from staging path is completed.")
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

	scsiInfo, err := extractISCSIInformation(req.PublishContext)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to get iSCSI info from publish context")
		return nil, status.Error(codes.InvalidArgument, "PublishContext is invalid")
	}

	//iSCSIMounter := iscsi.New(d.logger, "", "", 0)
	iSCSIMounter := iscsi.NewFromISCSIDisk(logger, scsiInfo)

	mnt := req.VolumeCapability.GetMount()
	options := mnt.MountFlags

	options = append(options, "bind")
	if req.Readonly {
		options = append(options, "ro")
	}

	fsType := validateFsType(logger, mnt.FsType)

	err = iSCSIMounter.Mount(req.StagingTargetPath, req.TargetPath, fsType, options)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to format and mount.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Info("Publish volume to the Node is Completed.")
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

	scsiInfo, _ := extractISCSIInformationFromMountPath(d.logger, req.TargetPath)
	if scsiInfo == nil {
		logger.Warn("unable to get the ISCSI info")
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}
	d.logger.With("ISCSIInfo", scsiInfo, "Mount Path", req.GetTargetPath()).Info("Found ISCSIInfo for NodeUnpublishVolume.")

	iSCSIMounter := iscsi.NewFromISCSIDisk(d.logger, scsiInfo)

	if err := iSCSIMounter.UnmountPath(req.TargetPath); err != nil {
		logger.With(zap.Error(err)).Error("failed to unmount the target path, error")
		return nil, status.Error(codes.Internal, err.Error())
	}

	logger.Info("Un-publish volume from the Node is Completed.")
	return &csi.NodeUnpublishVolumeResponse{}, nil
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
	ad, err := d.util.lookupNodeAvailableDomain(d.KubeClient, d.nodeID)

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
