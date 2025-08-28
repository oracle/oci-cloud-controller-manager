// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package driver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kubeAPI "k8s.io/api/core/v1"
	"k8s.io/mount-utils"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/disk"
)

const (
	FipsEnabled                = "1"
	fssMountSemaphoreTimeout   = time.Second * 30
	fssUnmountSemaphoreTimeout = time.Second * 30
)

var fssMountSemaphore = semaphore.NewWeighted(int64(2))
var fssUnmountSemaphore = semaphore.NewWeighted(int64(4))

// NodeStageVolume mounts the volume to a staging path on the node.
func (d FSSNodeDriver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Staging path must be provided")
	}

	logger := d.logger.With("volumeID", req.VolumeId)

	volumeHandler := csi_util.ValidateFssId(req.VolumeId)
	_, mountTargetIP, exportPath := volumeHandler.FilesystemOcid, volumeHandler.MountTargetIPAddress, volumeHandler.FsExportPath

	logger.Debugf("volumeHandler :  %v", volumeHandler)

	if mountTargetIP == "" || exportPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Volume ID provided")
	}

	if !d.nodeMetadata.IsNodeMetadataLoaded {
		d.util.LoadNodeMetadataFromApiServer(ctx, d.KubeClient, d.nodeID, d.nodeMetadata)
	}

	if csi_util.IsIpv4(mountTargetIP) && !d.nodeMetadata.Ipv4Enabled {
		return nil, status.Error(codes.InvalidArgument, "Ipv4 mount target identified in volume id, but worker node does not support ipv4 ip family.")
	} else if csi_util.IsIpv6(mountTargetIP) && !d.nodeMetadata.Ipv6Enabled {
		return nil, status.Error(codes.InvalidArgument, "Ipv6 mount target identified in volume id, but worker node does not support ipv6 ip family.")
	}

	logger.Debugf("volume context: %v", req.VolumeContext)

	var fsType = ""

	accessType := req.VolumeCapability.GetMount()
	var options []string
	if accessType != nil {
		if accessType.MountFlags != nil {
			options = accessType.MountFlags
		}
		if accessType.FsType != "" {
			fsType = accessType.FsType
		}
	}
	encryptInTransit, err := isInTransitEncryptionEnabled(req.VolumeContext)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "EncryptInTransit must be a boolean value")
	}

	mounter := mount.New("")

	if encryptInTransit {
		isPackageInstalled, err := csi_util.IsInTransitEncryptionPackageInstalled()
		if err != nil {
			logger.With(zap.Error(err)).Error("FSS in-transit encryption Package installation check failed")
			return nil, status.Error(codes.Internal, "FSS in-transit encryption Package installation check failed")
		}
		if !isPackageInstalled {
			logger.Errorf("Package %s not installed for in-transit encryption", csi_util.InTransitEncryptionPackageName)
			return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("Package %s not installed for in-transit encryption", csi_util.InTransitEncryptionPackageName))
		}
		logger.Debug("In-transit encryption enabled")
		fsType = "oci-fss"
		content, err := csi_util.IsFipsEnabled()
		if err != nil {
			logger.With(zap.Error(err)).Error("Could not verify if FIPS enabled")
			return nil, status.Error(codes.Internal, "Could not verify if FIPS enabled")
		}

		if len(content) > 0 && strings.Contains(content, FipsEnabled) {
			options = append(options, "fips")
			logger.Debug("Fips mode enabled")
		}
	}

	if acquired := d.volumeLocks.TryAcquire(req.VolumeId); !acquired {
		logger.Error("Could not acquire lock for NodeStageVolume.")
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, req.VolumeId)
	}

	defer d.volumeLocks.Release(req.VolumeId)

	logger.Debug("Trying to stage.")
	startTime := time.Now()

	fssMountSemaphoreCtx, cancel := context.WithTimeout(ctx, fssMountSemaphoreTimeout)
	defer cancel()

	err = fssMountSemaphore.Acquire(fssMountSemaphoreCtx, 1)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("Semaphore acquire timed out.")
		} else if errors.Is(err, context.Canceled) {
			logger.Error("Stage semaphore acquire context was canceled.")
		} else {
			logger.With(zap.Error(err)).Error("Error acquiring lock during stage.")
		}
		return nil, status.Error(codes.Aborted, "Too many mount requests.")
	}
	defer fssMountSemaphore.Release(1)

	logger.Info("Stage started.")

	targetPath := req.StagingTargetPath
	mountPoint, err := isMountPoint(mounter, targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.With("StagingTargetPath", targetPath).Infof("Mount point does not pre-exist, creating now.")
			// k8s v1.20+ will not create the TargetPath directory
			// https://github.com/kubernetes/kubernetes/pull/88759
			// if the path exists already (<v1.20) this is a no op
			// https://golang.org/pkg/os/#MkdirAll
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				logger.With(zap.Error(err)).Error("Failed to create StagingTargetPath directory")
				return nil, status.Error(codes.Internal, "Failed to create StagingTargetPath directory")
			}
			mountPoint = false
		} else {
			logger.With(zap.Error(err)).Error("Invalid Mount Point ", targetPath)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if mountPoint {
		logger.Infof("Volume is already mounted to: %v", targetPath)
		return &csi.NodeStageVolumeResponse{}, nil
	}

	source := fmt.Sprintf("%s:%s", csi_util.FormatValidIp(mountTargetIP), exportPath)

	if encryptInTransit {
		err = disk.MountWithEncrypt(logger, source, targetPath, fsType, options)
	} else {
		err = mounter.Mount(source, targetPath, fsType, options)
	}
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to mount volume to staging target path.")
		return nil, status.Error(codes.Internal, err.Error())
	}
	logger.With("mountTarget", mountTargetIP, "exportPath", exportPath, "StagingTargetPath", targetPath, "StageTime", time.Since(startTime).Milliseconds()).
		Info("Mounting the volume to staging target path is completed.")

	return &csi.NodeStageVolumeResponse{}, nil
}

func isInTransitEncryptionEnabled(volumeContext map[string]string) (bool, error) {
	if volumeContext != nil {
		if encryptInTransit, ok := volumeContext["encryptInTransit"]; ok {
			return strconv.ParseBool(encryptInTransit)
		}
	}
	return false, nil
}

func isMountPoint(mounter mount.Interface, path string) (bool, error) {
	ok, err := mounter.IsLikelyNotMountPoint(path)
	if err != nil {
		return false, err
	}
	return !ok, nil
}

// NodePublishVolume mounts the volume to the target path
func (d FSSNodeDriver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Staging path must be provided")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Target Path must be provided")
	}

	logger := d.logger.With("volumeID", req.VolumeId)
	logger.Debugf("volume context: %v", req.VolumeContext)

	var fsType = ""

	mounter := mount.New("")

	targetPath := req.GetTargetPath()
	readOnly := req.GetReadonly()
	// Use mount.IsNotMountPoint because mounter.IsLikelyNotMountPoint can't detect bind mounts
	isNotMountPoint, err := mount.IsNotMountPoint(mounter, targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.With("TargetPath", targetPath).Infof("mount point does not exist")
			// k8s v1.20+ will not create the TargetPath directory
			// https://github.com/kubernetes/kubernetes/pull/88759
			// if the path exists already (<v1.20) this is a no op
			// https://golang.org/pkg/os/#MkdirAll
			if err := os.MkdirAll(targetPath, 0750); err != nil {
				logger.With(zap.Error(err)).Error("Failed to create TargetPath directory")
				return nil, status.Error(codes.Internal, "Failed to create TargetPath directory")
			}
			isNotMountPoint = true
		} else {
			logger.With(zap.Error(err)).Error("Invalid Mount Point ", targetPath)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if !isNotMountPoint {
		logger.Infof("Volume is already mounted to: %v", targetPath)
		return &csi.NodePublishVolumeResponse{}, nil
	}

	options := []string{"bind"}
	if readOnly {
		options = append(options, "ro")
	}
	source := req.GetStagingTargetPath()

	logger.Debug("Trying to publish.")
	startTime := time.Now()

	fssMountSemaphoreCtx, cancel := context.WithTimeout(ctx, fssMountSemaphoreTimeout)
	defer cancel()

	err = fssMountSemaphore.Acquire(fssMountSemaphoreCtx, 1)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Error("Semaphore acquire timed out.")
		} else if errors.Is(err, context.Canceled) {
			logger.Error("Publish semaphore acquire context was canceled.")
		} else {
			logger.With(zap.Error(err)).Error("Error acquiring lock during stage.")
		}
		return nil, status.Error(codes.Aborted, "Too many mount requests.")
	}
	defer fssMountSemaphore.Release(1)

	logger.Info("Publish started.")

	err = mounter.Mount(source, targetPath, fsType, options)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to bind mount volume to target path.")
		return nil, status.Error(codes.Internal, err.Error())
	}
	logger.With("staging target path", source, "TargetPath", targetPath, "PublishTime", time.Since(startTime).Milliseconds()).
		Info("Bind mounting the volume to target path is completed.")

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the volume from the target path
func (d FSSNodeDriver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Volume ID must be provided")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Target Path must be provided")
	}

	logger := d.logger.With("volumeID", req.VolumeId, "targetPath", req.TargetPath)

	mounter := mount.New("")
	targetPath := req.GetTargetPath()

	// Use mount.IsNotMountPoint because mounter.IsLikelyNotMountPoint can't detect bind mounts
	isNotMountPoint, err := mount.IsNotMountPoint(mounter, targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.With("TargetPath", targetPath).Infof("mount point does not exist")
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if isNotMountPoint {
		err = os.RemoveAll(targetPath)
		if err != nil {
			logger.With(zap.Error(err)).Error("Remove target path failed with error")
			return nil, status.Error(codes.Internal, "Failed to remove target path")
		}
		logger.With("TargetPath", targetPath).Infof("Not a mount point, removing path")
		return &csi.NodeUnpublishVolumeResponse{}, nil
	}
	logger.Debug("Trying to unmount.")
	startTime := time.Now()

	fssUnmountSemaphoreCtx, cancel := context.WithTimeout(ctx, fssUnmountSemaphoreTimeout)
	defer cancel()

	err = fssUnmountSemaphore.Acquire(fssUnmountSemaphoreCtx, 1)
	if err != nil {
		logger.With(zap.Error(err)).Errorf("Error aquiring lock during unstage.")
		return nil, status.Error(codes.Aborted, "To many unmount requests.")
	}
	defer fssUnmountSemaphore.Release(1)

	logger.Info("Unmount started")
	if err := mounter.Unmount(targetPath); err != nil {
		logger.With(zap.Error(err)).Error("failed to unmount target path.")
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}
	logger.With("UnmountTime", time.Since(startTime).Milliseconds()).With("TargetPath", targetPath).
		Info("Unmounting volume completed")
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeUnstageVolume unstage the volume from the staging path
func (d FSSNodeDriver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Staging path must be provided")
	}

	volumeHandler := csi_util.ValidateFssId(req.VolumeId)

	logger := d.logger.With("volumeID", req.VolumeId, "stagingPath", req.StagingTargetPath)

	logger.Debugf("volumeHandler :  %v", volumeHandler)

	_, mountTargetIP, exportPath := volumeHandler.FilesystemOcid, volumeHandler.MountTargetIPAddress, volumeHandler.FsExportPath

	if mountTargetIP == "" || exportPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid Volume ID provided")
	}

	if acquired := d.volumeLocks.TryAcquire(req.VolumeId); !acquired {
		logger.Error("Could not acquire lock for NodeUnstageVolume.")
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, req.VolumeId)
	}

	defer d.volumeLocks.Release(req.VolumeId)

	targetPath := req.GetStagingTargetPath()
	logger.Debug("Trying to unstage.")
	startTime := time.Now()

	fssUnmountSemaphoreCtx, cancel := context.WithTimeout(ctx, fssUnmountSemaphoreTimeout)
	defer cancel()

	err := fssUnmountSemaphore.Acquire(fssUnmountSemaphoreCtx, 1)
	if err != nil {
		logger.With(zap.Error(err)).Errorf("Error aquiring lock during unstage.")
		return nil, status.Error(codes.Aborted, "To many unmount requests.")
	}
	defer fssUnmountSemaphore.Release(1)

	logger.Info("Unstage started.")
	err = d.unmountAndCleanup(logger, targetPath, exportPath, mountTargetIP)
	if err != nil {
		return nil, err
	}

	logger.With("UnstageTime", time.Since(startTime).Milliseconds()).
		With("StagingTargetPath", targetPath).Info("Unmounting volume completed")
	return &csi.NodeUnstageVolumeResponse{}, nil
}

func (d FSSNodeDriver) unmountAndCleanup(logger *zap.SugaredLogger, targetPath string, exportPath string, mountTargetIP string) error {
	mounter := mount.New("")
	// Use mount.IsNotMountPoint because mounter.IsLikelyNotMountPoint can't detect bind mounts
	isNotMountPoint, err := mount.IsNotMountPoint(mounter, targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.With("StagingTargetPath", targetPath).Infof("mount point does not exist")
			return nil
		}
		return status.Error(codes.Internal, err.Error())
	}

	if isNotMountPoint {
		err = os.RemoveAll(targetPath)
		if err != nil {
			logger.With(zap.Error(err)).Error("Remove target path failed with error")
			return status.Error(codes.Internal, "Failed to remove target path")
		}
		logger.With("StagingTargetPath", targetPath).Infof("Not a mount point, removing path")
		return nil
	}

	sources, err := disk.FindMount(targetPath)
	if err != nil {
		logger.With(zap.Error(err)).Error("Find Mount failed for target path")
		return status.Error(codes.Internal, "Find Mount failed for target path")
	}

	inTransitEncryption := false
	for _, device := range sources {

		logger.With("device", device).With("exportPath", exportPath).
			With("mountTargetIP", mountTargetIP).Debugf("Identifying intransit encryption.")
		if strings.HasSuffix(device, exportPath) && !strings.HasPrefix(device, mountTargetIP) {
			logger.Debugf("Intransit encryption identified.")
			inTransitEncryption = true
			break
		}
	}

	logger.Debugf("Sources mounted at staging target path %v", sources)

	if inTransitEncryption {
		err = disk.UnmountWithEncrypt(logger, targetPath)
	} else {
		err = mounter.Unmount(targetPath)
	}
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to unmount StagingTargetPath")
		return status.Errorf(codes.Internal, "%s", err.Error())
	}
	return nil
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (d FSSNodeDriver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
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
func (d FSSNodeDriver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {

	if !d.nodeMetadata.IsNodeMetadataLoaded {
		err := d.util.LoadNodeMetadataFromApiServer(ctx, d.KubeClient, d.nodeID, d.nodeMetadata)
		if err != nil || d.nodeMetadata.AvailabilityDomain == "" {
			d.logger.With(zap.Error(err)).With("nodeId", d.nodeID).Error("Failed to get availability domain of node from kube api server.")
			return nil, status.Error(codes.Internal, "Failed to get availability domain of node from kube api server.")
		}
	}

	d.logger.With("nodeId", d.nodeID, "availableDomain", d.nodeMetadata.AvailabilityDomain).Info("Available domain of node identified.")

	return &csi.NodeGetInfoResponse{
		NodeId: d.nodeID,
		// make sure that the driver works on this particular AD only
		AccessibleTopology: &csi.Topology{
			Segments: map[string]string{
				kubeAPI.LabelZoneFailureDomain: d.nodeMetadata.AvailabilityDomain,
				kubeAPI.LabelTopologyZone:      d.nodeMetadata.AvailabilityDomain,
			},
		},
	}, nil
}

// NodeGetVolumeStats return the stats of the volume
func (d FSSNodeDriver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetVolumeStats is not supported yet")
}

// NodeExpandVolume returns the expand of the volume
func (d FSSNodeDriver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeExpandVolume is not supported yet")
}
