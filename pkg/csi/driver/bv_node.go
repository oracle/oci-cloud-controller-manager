// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/disk"
	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kubeAPI "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util/hostutil"
)

const (
	maxVolumesPerNode               = 32
	volumeOperationAlreadyExistsFmt = "An operation for the volume: %s already exists."
	FSTypeXfs                       = "xfs"
)

// NodeStageVolume mounts the volume to a staging path on the node.
func (d BlockVolumeNodeDriver) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
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

	logger := d.logger.With("volumeID", req.VolumeId, "stagingPath", req.StagingTargetPath)

	stagingTargetFilePath := csi_util.GetPathForBlock(req.StagingTargetPath)

	isRawBlockVolume := false

	if _, ok := req.VolumeCapability.GetAccessType().(*csi.VolumeCapability_Block); ok {
		isRawBlockVolume = true
	}

	logger.Infof("Is Volume Mode set to Raw Block Volume %s", isRawBlockVolume)

	attachment, ok := req.PublishContext[attachmentType]

	if !ok {
		logger.Error("Unable to get the attachmentType from the attribute list, assuming iscsi")
		attachment = attachmentTypeISCSI
	}

	var devicePath string
	var mountHandler disk.Interface
	var scsiInfo *disk.Disk
	var err error
	var multipathDevices []core.MultipathDevice
	multipathEnabledVolume := false

	if req.PublishContext[multipathEnabled] != "" {
		multipathEnabledVolume, err = strconv.ParseBool(req.PublishContext[multipathEnabled])
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to determine if volume is multipath enabled")
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	switch attachment {
	case attachmentTypeISCSI:
		if multipathEnabledVolume {
			logger.Info("Volume attachment is multipath enabled")
			multipathDevices, err = getMultipathDevicesFromReq(req)
			devicePath, err = disk.GetMultipathIscsiDevicePath(ctx, req.PublishContext[device], logger)
			if err != nil {
				logger.With(zap.Error(err)).Error("Failed to get device path for multipath enabled volume")
				return nil, status.Error(codes.Internal, "Failed to get device path for multipath enabled volume")
			}
			mountHandler = disk.NewISCSIUHPMounter(d.logger)
			logger.Info("starting to stage UHP iSCSI Mounting.")
		} else {
			logger.Info("Volume attachment is multipath disabled")
			scsiInfo, err = csi_util.ExtractISCSIInformation(req.PublishContext)
			if err != nil {
				logger.With(zap.Error(err)).Error("Failed to get SCSI info from publish context.")
				return nil, status.Error(codes.InvalidArgument, "PublishContext is invalid.")
			}

			if strings.EqualFold(d.nodeMetadata.PreferredNodeIpFamily, csi_util.Ipv6Stack) {
				scsiInfo.IscsiIp, err = csi_util.ConvertIscsiIpFromIpv4ToIpv6(scsiInfo.IscsiIp)
				if err != nil {
					logger.With(zap.Error(err)).Error("Failed get ipv6 address for Iscsi Target.")
					return nil, status.Errorf(codes.Internal, "Failed get ipv6 address for Iscsi Target.")
				}
			}

			mountHandler = disk.NewFromISCSIDisk(d.logger, scsiInfo)
			logger.Info("starting to stage iSCSI Mounting.")
		}
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

	if !isRawBlockVolume {
		isMounted, oErr := mountHandler.IsMounted(devicePath, req.StagingTargetPath)
		if oErr != nil {
			logger.With(zap.Error(oErr)).Error("getting error to get the details about volume is already mounted or not.")
			return nil, status.Error(codes.Internal, oErr.Error())
		} else if isMounted {
			logger.Info("volume is already mounted on the staging path.")
			return &csi.NodeStageVolumeResponse{}, nil
		}
	}

	err = mountHandler.AddToDB()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to add the iSCSI node record.")
		return nil, status.Error(codes.Internal, err.Error())
	}

	v, ok := req.PublishContext[csi_util.VpusPerGB]
	if !ok {
		logger.Infof("vpusPerGB not found in PublishContext %v, applying default 10 vpusPerGB", req.PublishContext)
		v = "10"
	}
	vpusPerGB, err := csi_util.ExtractBlockVolumePerformanceLevel(v)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if vpusPerGB == csi_util.HigherPerformanceOption {
		err := mountHandler.UpdateQueueDepth()
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to update queue depth in the iSCSI node record.")
			return nil, status.Error(codes.Internal, err.Error())
		}
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
	if attachment == attachmentTypeISCSI && !multipathEnabledVolume {
		// Wait and get device path using the publish context
		devicePath, err = disk.WaitForDevicePathToExist(ctx, scsiInfo, logger)
		if err != nil {
			logger.With(zap.Error(err)).Error("Failed to get /dev/disk/by-path device path for iscsi volume.")
			err = mountHandler.ISCSILogoutOnFailure()
			if err != nil {
				return nil, status.Error(codes.Internal, "Failed to iscsi logout after timeout on waiting for device path to exist")
			}
			return nil, status.Error(codes.InvalidArgument, "Failed to get device path for iscsi volume")
		}
	}

	err = mountHandler.WaitForVolumeLoginOrTimeout(ctx, multipathDevices)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !mountHandler.WaitForPathToExist(devicePath, 20) {
		logger.Error("failed to wait for device to exist.")
		return nil, status.Error(codes.DeadlineExceeded, "Failed to wait for device to exist.")
	}

	if isRawBlockVolume {
		err := csi_util.CreateFilePath(logger, stagingTargetFilePath)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to create the stagingTargetFile.")
			err = mountHandler.ISCSILogoutOnFailure()
			if err != nil {
				return nil, status.Error(codes.Internal, "Failed to iscsi logout after mount failure")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		options := []string{"bind"} // Append the "bind" option if it is a raw block volume
		err = mountHandler.Mount(devicePath, stagingTargetFilePath, "", options)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to bind mount raw block volume to stagingTargetFile")
			err = mountHandler.ISCSILogoutOnFailure()
			if err != nil {
				return nil, status.Error(codes.Internal, "Failed to iscsi logout after mount failure")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &csi.NodeStageVolumeResponse{}, nil
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
			err = mountHandler.ISCSILogoutOnFailure()
			if err != nil {
				return nil, status.Error(codes.Internal, "Failed to iscsi logout after failure to check if staging path exists")
			}
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
			err = mountHandler.ISCSILogoutOnFailure()
			if err != nil {
				return nil, status.Error(codes.Internal, "Failed to iscsi logout after failure to create StagingTargetPath directory")
			}
			return nil, status.Error(codes.Internal, "Failed to create StagingTargetPath directory")
		}
	}

	//XFS does not allow mounting two volumes with same UUID,
	//this block is needed for mounting a volume and a volume
	//restored from it's snapshot on the same node
	if fsType == FSTypeXfs {
		if !hasMountOption(options, "nouuid") {
			options = append(options, "nouuid")
		}
	}

	existingFs, err := mountHandler.GetDiskFormat(devicePath)
	if err != nil {
		logger.With("devicePath", devicePath, zap.Error(err)).Error("GetDiskFormatFailed")
	}

	if existingFs != "" && existingFs != fsType {
		returnError := fmt.Sprintf("FS Type mismatch detected. The existing fs type on the volume: %q doesn't match the requested fs type: %q. Please change fs type in PV to match the existing fs type.", existingFs, fsType)
		logger.Error(returnError)
		err = mountHandler.ISCSILogoutOnFailure()
		if err != nil {
			return nil, status.Error(codes.Internal, "Failed to iscsi logout after failure due to FS Type mismatch")
		}
		return nil, status.Error(codes.Internal, returnError)
	}

	logger.With("devicePath", devicePath,
		"fsType", fsType).Info("mounting the volume to staging path.")
	err = mountHandler.FormatAndMount(devicePath, req.StagingTargetPath, fsType, options)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to format and mount volume to staging path.")
		err = mountHandler.ISCSILogoutOnFailure()
		if err != nil {
			return nil, status.Error(codes.Internal, "Failed to iscsi logout after mount failure")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	logger.With("devicePath", devicePath, "fsType", fsType, "attachmentType", attachment).
		Info("Mounting the volume to staging path is completed.")

	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume unstage the volume from the staging path
func (d BlockVolumeNodeDriver) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.StagingTargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "Staging target path must be provided")
	}

	logger := d.logger.With("volumeID", req.VolumeId, "stagingPath", req.StagingTargetPath)

	stagingTargetFilePath := csi_util.GetPathForBlock(req.StagingTargetPath)

	if acquired := d.volumeLocks.TryAcquire(req.VolumeId); !acquired {
		logger.Error("Could not acquire lock for NodeUnstageVolume.")
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, req.VolumeId)
	}

	defer d.volumeLocks.Release(req.VolumeId)

	hostUtil := hostutil.NewHostUtil()
	isRawBlockVolume, rbvCheckErr := hostUtil.PathIsDevice(stagingTargetFilePath)

	if rbvCheckErr != nil {
		logger.With(zap.Error(rbvCheckErr)).Warn("failed to check if it is a device file")
		isRawBlockVolume = false
	}

	var diskPath []string
	var err error

	if isRawBlockVolume {
		diskPath, err = disk.GetDiskPathFromBindDeviceFilePath(logger, stagingTargetFilePath)

		if err != nil {
			logger.With(zap.Error(err)).With("mountPath", stagingTargetFilePath).Error("unable to get diskPath from mount path")
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		diskPath, err = disk.GetDiskPathFromMountPath(logger, req.GetStagingTargetPath())

		if err != nil {
			// do a clean exit in case of mount point not found
			if err == disk.ErrMountPointNotFound {
				logger.With(zap.Error(err)).With("mountPath", req.GetStagingTargetPath()).Warn("unable to fetch mount point")
				return &csi.NodeUnstageVolumeResponse{}, nil
			}
			logger.With(zap.Error(err)).With("mountPath", req.GetStagingTargetPath()).Error("unable to get diskPath from mount path")
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	attachmentType, devicePath, err := getDevicePathAndAttachmentType(diskPath)
	if err != nil {
		logger.With(zap.Error(err)).With("diskPath", diskPath).Error("unable to determine the attachment type")
		return nil, status.Error(codes.Internal, err.Error())
	}

	// for multipath enabled volumes the device path will be eg: /dev/mapper/mpathd
	isMultipathEnabled := strings.HasPrefix(devicePath, "/dev/mapper")

	var scsiInfo *disk.Disk
	var mountHandler disk.Interface
	switch attachmentType {
	case attachmentTypeISCSI:
		if !isMultipathEnabled {
			scsiInfo, err = csi_util.ExtractISCSIInformationFromMountPath(d.logger, diskPath)
			if err != nil {
				logger.With(zap.Error(err)).Error("failed to ISCSI info.")
				return nil, status.Error(codes.Internal, err.Error())
			}
			if scsiInfo == nil {
				logger.Warn("unable to get the ISCSI info")
				return &csi.NodeUnstageVolumeResponse{}, nil
			}
			mountHandler = disk.NewFromISCSIDisk(d.logger, scsiInfo)
		} else {
			mountHandler = disk.NewISCSIUHPMounter(d.logger)
			logger.With("diskPath", diskPath).Info("Volume is multipath enabled")
		}
		logger.Info("starting to unstage iscsi Mounting.")
	case attachmentTypeParavirtualized:
		mountHandler = disk.NewFromPVDisk(d.logger)
		logger.Info("starting to unstage paravirtualized Mounting.")
	default:
		logger.Error("unknown attachment type. supported attachment types are iscsi and paravirtualized")
		return nil, status.Error(codes.InvalidArgument, "unknown attachment type. supported attachment types are iscsi and paravirtualized")
	}

	if !isRawBlockVolume {
		isMounted, oErr := mountHandler.DeviceOpened(devicePath)
		if oErr != nil {
			logger.With(zap.Error(oErr)).Error("getting error to get the details about volume is already unmounted or not.")
			return nil, status.Error(codes.Internal, oErr.Error())
		} else if !isMounted {
			logger.Info("volume is already unmounted from the staging path.")
			return &csi.NodeUnstageVolumeResponse{}, nil
		}
	}

	var unMountErr error

	if isRawBlockVolume {
		unMountErr = mountHandler.UnmountPath(stagingTargetFilePath)
	} else {
		unMountErr = mountHandler.UnmountPath(req.StagingTargetPath)
	}

	if unMountErr != nil {
		logger.With(zap.Error(unMountErr)).Error("failed to unmount the staging path")
		return nil, status.Error(codes.Internal, unMountErr.Error())
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
func (d BlockVolumeNodeDriver) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
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

	logger := d.logger.With("volumeID", req.VolumeId, "targetPath", req.TargetPath)

	stagingTargetFilePath := csi_util.GetPathForBlock(req.StagingTargetPath)

	isRawBlockVolume := false

	if _, ok := req.VolumeCapability.GetAccessType().(*csi.VolumeCapability_Block); ok {
		isRawBlockVolume = true
	}

	logger.With("isRawBlockVolume", isRawBlockVolume)

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
	if !isRawBlockVolume {
		if err := os.MkdirAll(req.TargetPath, 0750); err != nil {
			logger.With(zap.Error(err)).Error("Failed to create TargetPath directory")
			return nil, status.Error(codes.Internal, "Failed to create TargetPath directory")
		}
	}

	multipathEnabledVolume := false

	if req.PublishContext[multipathEnabled] != "" {
		var err error
		multipathEnabledVolume, err = strconv.ParseBool(req.PublishContext[multipathEnabled])
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to determine if volume is multipath enabled")
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	var mountHandler disk.Interface

	switch attachment {
	case attachmentTypeISCSI:
		if multipathEnabledVolume {
			mountHandler = disk.NewISCSIUHPMounter(d.logger)
		} else {
			scsiInfo, err := csi_util.ExtractISCSIInformation(req.PublishContext)
			if err != nil {
				logger.With(zap.Error(err)).Error("Failed to get iSCSI info from publish context")
				return nil, status.Error(codes.InvalidArgument, "PublishContext is invalid")
			}
			mountHandler = disk.NewFromISCSIDisk(logger, scsiInfo)
		}
		logger.Info("starting to publish iSCSI Mounting.")

	case attachmentTypeParavirtualized:
		mountHandler = disk.NewFromPVDisk(d.logger)
		logger.Info("starting to publish paravirtualized Mounting.")
	default:
		logger.Error("unknown attachment type. supported attachment types are iscsi and paravirtualized")
		return nil, status.Error(codes.InvalidArgument, "unknown attachment type. supported attachment types are iscsi and paravirtualized")
	}

	if isRawBlockVolume {
		options := []string{"bind"}
		if req.Readonly {
			options = append(options, "ro")
		}

		err := csi_util.CreateFilePath(logger, req.TargetPath)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to create the target file.")
			return nil, status.Error(codes.Internal, err.Error())
		}

		err = mountHandler.Mount(stagingTargetFilePath, req.TargetPath, "", options)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to bind mount raw block volume to target file.")
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		mnt := req.VolumeCapability.GetMount()
		options := mnt.MountFlags

		options = append(options, "bind")
		if req.Readonly {
			options = append(options, "ro")
		}

		fsType := csi_util.ValidateFsType(logger, mnt.FsType)

		//XFS does not allow mounting two volumes with same UUID,
		//this block is needed for mounting a volume and a volume
		//restored from it's snapshot on the same node
		if fsType == FSTypeXfs {
			if !hasMountOption(options, "nouuid") {
				options = append(options, "nouuid")
			}
		}

		err := mountHandler.Mount(req.StagingTargetPath, req.TargetPath, fsType, options)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to format and mount.")
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	logger.With("attachmentType", attachment).Info("Publish volume to the Node is Completed.")

	if req.PublishContext[needResize] != "" {
		needsResize, err := strconv.ParseBool(req.PublishContext[needResize])
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to determine if resize is required")
			return nil, status.Error(codes.Internal, err.Error())
		}

		if needsResize {
			logger.Info("Starting to expand volume to requested size")

			requestedSize, err := strconv.ParseInt(req.PublishContext[newSize], 10, 64)
			if err != nil {
				logger.With(zap.Error(err)).Error("failed to get new requested size of volume")
				return nil, status.Errorf(codes.OutOfRange, "failed to get new requested size of volume: %v", err)
			}
			requestedSizeGB := csi_util.RoundUpSize(requestedSize, 1*client.GiB)

			var diskPath []string
			var diskErr error

			if isRawBlockVolume {
				diskPath, diskErr = disk.GetDiskPathFromBindDeviceFilePath(logger, req.TargetPath)
				if diskErr != nil {
					logger.With(zap.Error(diskErr)).With("mountPath", req.TargetPath).Error("unable to get diskPath from mount path")
					return nil, status.Error(codes.Internal, diskErr.Error())
				}
			} else {
				diskPath, diskErr = disk.GetDiskPathFromMountPath(d.logger, req.StagingTargetPath)
				if diskErr != nil {
					// do a clean exit in case of mount point not found
					if diskErr == disk.ErrMountPointNotFound {
						logger.With(zap.Error(diskErr)).With("volumePath", req.StagingTargetPath).Warn("unable to fetch mount point")
						return &csi.NodePublishVolumeResponse{}, diskErr
					}
					logger.With(zap.Error(diskErr)).With("volumePath", req.StagingTargetPath).Error("unable to get diskPath from mount path")
					return nil, status.Error(codes.Internal, diskErr.Error())
				}
			}

			attachmentType, devicePath, err := getDevicePathAndAttachmentType(diskPath)
			if err != nil {
				logger.With(zap.Error(err)).With("diskPath", diskPath).Error("unable to determine the attachment type")
				return nil, status.Error(codes.Internal, err.Error())
			}
			logger.With("diskPath", diskPath, "attachmentType", attachmentType, "devicePath", devicePath).Infof("Extracted attachment type and device path")

			var mountHandler disk.Interface
			switch attachmentType {
			case attachmentTypeISCSI:
				if multipathEnabledVolume {
					mountHandler = disk.NewISCSIUHPMounter(d.logger)
				} else {
					mountHandler = disk.NewFromISCSIDisk(d.logger, nil)
				}
			case attachmentTypeParavirtualized:
				mountHandler = disk.NewFromPVDisk(d.logger)
				logger.Info("starting to expand paravirtualized Mounting.")
			default:
				logger.Error("unknown attachment type. supported attachment types are iscsi and paravirtualized")
				return nil, status.Error(codes.InvalidArgument, "unknown attachment type. supported attachment types are iscsi and paravirtualized")
			}

			if err := mountHandler.Rescan(devicePath); err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to rescan volume %q (%q):  %v", req.VolumeId, devicePath, err)
			}
			logger.With("devicePath", devicePath).Debug("Rescan completed")

			if !isRawBlockVolume {
				if _, err := mountHandler.Resize(devicePath, req.TargetPath); err != nil {
					return nil, status.Errorf(codes.Internal, "Failed to resize volume %q (%q):  %v", req.VolumeId, devicePath, err)
				}
			}

			allocatedSizeBytes, err := csi_util.GetBlockSizeBytes(logger, devicePath)
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to get size of block volume at path %s: %v", devicePath, err))
			}

			allocatedSizeGB := csi_util.RoundUpSize(allocatedSizeBytes, 1*client.GiB)

			if allocatedSizeGB < requestedSizeGB {
				return nil, status.Error(codes.Internal, fmt.Sprintf("Expand volume after restore from snapshot failed, requested size in GB %d but resize allocated only %d", requestedSizeGB, allocatedSizeGB))
			}

			logger.Info("Volume successfully expanded after restore from snapshot")
		}
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume unmounts the volume from the target path
func (d BlockVolumeNodeDriver) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Volume ID must be provided")
	}

	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Target Path must be provided")
	}

	logger := d.logger.With("volumeID", req.VolumeId, "targetPath", req.TargetPath)

	hostUtil := hostutil.NewHostUtil()
	isRawBlockVolume, rbvCheckErr := hostUtil.PathIsDevice(req.TargetPath)

	if rbvCheckErr != nil {
		if alreadyDeletedPathCheck(rbvCheckErr) {
			logger.With(zap.Error(rbvCheckErr)).With("mountPath", req.TargetPath).Warn("mount point not found, marking unpublish success")
			return &csi.NodeUnpublishVolumeResponse{}, nil
		}
		logger.With(zap.Error(rbvCheckErr)).Error("failed to check if it is a device file")
		return nil, status.Errorf(codes.Internal, rbvCheckErr.Error())
	}

	if acquired := d.volumeLocks.TryAcquire(req.VolumeId); !acquired {
		logger.Error("Could not acquire lock for NodeUnpublishVolume.")
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, req.VolumeId)
	}

	defer d.volumeLocks.Release(req.VolumeId)

	var diskPath []string
	var err error

	if isRawBlockVolume {
		diskPath, err = disk.GetDiskPathFromBindDeviceFilePath(logger, req.TargetPath)
		if err != nil {
			logger.With(zap.Error(err)).With("mountPath", req.TargetPath).Error("unable to get diskPath from mount path")
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		diskPath, err = disk.GetDiskPathFromMountPath(d.logger, req.TargetPath)
		if err != nil {
			// do a clean exit in case of mount point not found
			if err == disk.ErrMountPointNotFound {
				logger.With(zap.Error(err)).With("mountPath", req.TargetPath).Warn("unable to fetch mount point")
				return &csi.NodeUnpublishVolumeResponse{}, nil
			}
			logger.With(zap.Error(err)).With("mountPath", req.TargetPath).Error("unable to get diskPath from mount path")
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	attachmentType, devicePath, err := getDevicePathAndAttachmentType(diskPath)
	if err != nil {
		logger.With(zap.Error(err)).With("diskPath", diskPath).Error("unable to determine the attachment type")
		return nil, status.Error(codes.Internal, err.Error())
	}

	// for multipath enabled volumes the device path will be eg: /dev/mapper/mpathd
	isMultipathEnabled := strings.HasPrefix(devicePath, "/dev/mapper")

	var mountHandler disk.Interface
	switch attachmentType {
	case attachmentTypeISCSI:
		if isMultipathEnabled {
			mountHandler = disk.NewISCSIUHPMounter(d.logger)
		} else {
			mountHandler = disk.NewFromISCSIDisk(d.logger, nil)
		}
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

func getDevicePathAndAttachmentType(path []string) (string, string, error) {
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
		if strings.HasPrefix(diskByPath, "/dev/mapper") {
			return attachmentTypeISCSI, diskByPath, nil
		}
	}

	return "", "", errors.New("unable to determine the attachment type")
}

func alreadyDeletedPathCheck(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "does not exist")
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (d BlockVolumeNodeDriver) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	var nscaps []*csi.NodeServiceCapability
	nodeCaps := []csi.NodeServiceCapability_RPC_Type{csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME, csi.NodeServiceCapability_RPC_GET_VOLUME_STATS, csi.NodeServiceCapability_RPC_EXPAND_VOLUME}
	for _, nodeCap := range nodeCaps {
		c := &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: nodeCap,
				},
			},
		}
		nscaps = append(nscaps, c)
	}

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: nscaps,
	}, nil
}

// NodeGetInfo returns the supported capabilities of the node server.
// The result of this function will be used by the CO in ControllerPublishVolume.
func (d BlockVolumeNodeDriver) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {

	if !d.nodeMetadata.IsNodeMetadataLoaded {
		err := d.util.LoadNodeMetadataFromApiServer(ctx, d.KubeClient, d.nodeID, d.nodeMetadata)
		if err != nil || d.nodeMetadata.AvailabilityDomain == "" {
			d.logger.With(zap.Error(err)).With("nodeId", d.nodeID).Error("Failed to get availability domain of node from kube api server.")
			return nil, status.Error(codes.Internal, "Failed to get availability domain of node from kube api server.")
		}
	}
	segments := map[string]string{
		kubeAPI.LabelZoneFailureDomain: d.nodeMetadata.AvailabilityDomain,
		kubeAPI.LabelTopologyZone:      d.nodeMetadata.AvailabilityDomain,
	}

	//set full ad name in segments only for IPv6 single stack
	if csi_util.IsIpv6SingleStackNode(d.nodeMetadata) {
		if d.nodeMetadata.FullAvailabilityDomain == "" {
			d.logger.With("nodeId", d.nodeID).Error("Failed to get full availability domain name of IPv6 single stack node from node labels.")
			return nil, status.Error(codes.Internal, "Failed to get full availability domain name of IPv6 single stack node from node labels.")
		}

		segments[csi_util.AvailabilityDomainLabel] = d.nodeMetadata.FullAvailabilityDomain
	}

	d.logger.With("nodeId", d.nodeID, "availabilityDomain", d.nodeMetadata.AvailabilityDomain).Info("Availability domain of node identified.")

	return &csi.NodeGetInfoResponse{
		NodeId:            d.nodeID,
		MaxVolumesPerNode: maxVolumesPerNode,

		// make sure that the driver works on this particular AD only
		AccessibleTopology: &csi.Topology{
			Segments: segments,
		},
	}, nil
}

// NodeGetVolumeStats return the stats of the volume
func (d BlockVolumeNodeDriver) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	logger := d.logger.With("volumeID", req.VolumeId, "volumePath", req.VolumePath)

	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		logger.Errorf("Volume ID not provided")
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	volumePath := req.GetVolumePath()
	if len(volumePath) == 0 {
		logger.Errorf("Volume path not provided")
		return nil, status.Error(codes.InvalidArgument, "volume path must be provided")
	}

	hostUtil := hostutil.NewHostUtil()
	isRawBlockVolume, rbvCheckErr := hostUtil.PathIsDevice(volumePath)

	if rbvCheckErr != nil {
		logger.With(zap.Error(rbvCheckErr)).Errorf("failed to check if the volumePath is a Device %s", volumePath)
		return nil, status.Error(codes.Internal, rbvCheckErr.Error())
	}

	if isRawBlockVolume {
		metricsProvider := volume.NewMetricsBlock(volumePath)
		metrics, err := metricsProvider.GetMetrics()
		if err != nil {
			logger.With(zap.Error(err)).Errorf("failed to get metrics for device at %s", volumePath)
			return nil, status.Error(codes.Internal, err.Error())
		}

		return &csi.NodeGetVolumeStatsResponse{
			Usage: []*csi.VolumeUsage{
				{
					Unit:  csi.VolumeUsage_BYTES,
					Total: metrics.Capacity.AsDec().UnscaledBig().Int64(),
				},
			},
		}, nil
	}

	exists, err := hostUtil.PathExists(volumePath)
	if err != nil {
		logger.With(zap.Error(err)).Errorf("Failed to find if path exists %s", volumePath)
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !exists {
		logger.Infof("Path does not exist %s", volumePath)
		return nil, status.Errorf(codes.NotFound, "path %s does not exist", volumePath)
	}

	metricsProvider := volume.NewMetricsStatFS(volumePath)
	metrics, err := metricsProvider.GetMetrics()
	if err != nil {
		logger.With(zap.Error(err)).Errorf("failed to get block volume info on path %s: %v", volumePath, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Unit:      csi.VolumeUsage_BYTES,
				Available: metrics.Available.AsDec().UnscaledBig().Int64(),
				Total:     metrics.Capacity.AsDec().UnscaledBig().Int64(),
				Used:      metrics.Used.AsDec().UnscaledBig().Int64(),
			},
			{
				Unit:      csi.VolumeUsage_INODES,
				Available: metrics.InodesFree.AsDec().UnscaledBig().Int64(),
				Total:     metrics.Inodes.AsDec().UnscaledBig().Int64(),
				Used:      metrics.InodesUsed.AsDec().UnscaledBig().Int64(),
			},
		},
	}, nil
}

// NodeExpandVolume returns the expand of the volume
func (d BlockVolumeNodeDriver) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID not provided")
	}
	volumePath := req.GetVolumePath()
	if len(volumePath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volume path must be provided")
	}

	logger := d.logger.With("volumeID", req.VolumeId, "volumePath", req.VolumePath)

	if acquired := d.volumeLocks.TryAcquire(req.VolumeId); !acquired {
		logger.Error("Could not acquire lock for NodeExpandVolume.")
		return nil, status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, req.VolumeId)
	}

	defer d.volumeLocks.Release(req.VolumeId)

	requestedSize, err := csi_util.ExtractStorage(req.CapacityRange)
	requestedSizeGB := csi_util.RoundUpSize(requestedSize, 1*client.GiB)

	if err != nil {
		logger.With(zap.Error(err)).Error("invalid capacity range")
		return nil, status.Errorf(codes.OutOfRange, "invalid capacity range: %v", err)
	}

	hostUtil := hostutil.NewHostUtil()
	isRawBlockVolume, rbvCheckErr := hostUtil.PathIsDevice(req.VolumePath)

	if rbvCheckErr != nil {
		logger.With(zap.Error(rbvCheckErr)).Error("failed to check if it is a device file.")
		return nil, status.Error(codes.Internal, rbvCheckErr.Error())
	}

	var diskPath []string

	if !isRawBlockVolume {
		diskPath, err = disk.GetDiskPathFromMountPath(logger, volumePath)
		if err != nil {
			if err == disk.ErrMountPointNotFound {
				logger.With(zap.Error(err)).With("volumePath", volumePath).Warn("unable to fetch mount point")
				return &csi.NodeExpandVolumeResponse{}, nil
			}
			logger.With(zap.Error(err)).Errorf("unable to get diskPath from mount path %s", volumePath)
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		diskPath, err = disk.GetDiskPathFromBindDeviceFilePath(logger, volumePath)
		if err != nil {
			logger.With(zap.Error(err)).Errorf("unable to get disk paths from volumePath %s", volumePath)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	attachmentType, devicePath, err := getDevicePathAndAttachmentType(diskPath)
	if err != nil {
		logger.With(zap.Error(err)).With("diskPath", diskPath).Error("unable to determine the attachment type")
		return nil, status.Error(codes.Internal, err.Error())
	}
	logger.With("diskPath", diskPath, "attachmentType", attachmentType, "devicePath", devicePath).Infof("Extracted attachment type and device path")

	// for multipath enabled volumes the device path will be eg: /dev/mapper/mpathd
	isMultipathEnabled := strings.HasPrefix(devicePath, "/dev/mapper")

	var mountHandler disk.Interface
	switch attachmentType {
	case attachmentTypeISCSI:
		if !isMultipathEnabled {
			mountHandler = disk.NewFromISCSIDisk(d.logger, nil)
		} else {
			mountHandler = disk.NewISCSIUHPMounter(d.logger)
		}
	case attachmentTypeParavirtualized:
		mountHandler = disk.NewFromPVDisk(d.logger)
		logger.Info("starting to expand paravirtualized Mounting.")
	default:
		logger.Error("unknown attachment type. supported attachment types are iscsi and paravirtualized")
		return nil, status.Error(codes.InvalidArgument, "unknown attachment type. supported attachment types are iscsi and paravirtualized")
	}

	if err := mountHandler.Rescan(devicePath); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to rescan volume %q (%q):  %v", volumeID, devicePath, err)
	}
	logger.With("devicePath", devicePath).Debug("Rescan completed")

	if !isRawBlockVolume {
		if _, err := mountHandler.Resize(devicePath, volumePath); err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to resize volume %q (%q):  %v", volumeID, devicePath, err)
		}
	}

	allocatedSizeBytes, err := csi_util.GetBlockSizeBytes(logger, devicePath)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to get size of block volume at path %s: %v", devicePath, err))
	}

	allocatedSizeGB := csi_util.RoundUpSize(allocatedSizeBytes, 1*client.GiB)

	if allocatedSizeGB < requestedSizeGB {
		return nil, status.Error(codes.Internal, fmt.Sprintf("Expand Volume Failed, requested size in GB %d but resize allocated only %d", requestedSizeGB, allocatedSizeGB))
	}

	return &csi.NodeExpandVolumeResponse{
		CapacityBytes: allocatedSizeBytes,
	}, nil
}

// hasMountOption returns a boolean indicating whether the given
// slice already contains a mount option. This is used to prevent
// passing duplicate option to the mount command.
func hasMountOption(options []string, opt string) bool {
	for _, o := range options {
		if o == opt {
			return true
		}
	}
	return false
}

func getMultipathDevicesFromReq(req *csi.NodeStageVolumeRequest) ([]core.MultipathDevice, error) {
	var multipathDevicesList []core.MultipathDevice

	err := json.Unmarshal([]byte(req.PublishContext[multipathDevices]), &multipathDevicesList)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get multipath devices from publish context.")
	}

	port, err := strconv.Atoi(req.PublishContext[disk.ISCSIPORT])
	if err != nil {
		return nil, status.Error(codes.Internal, "Invalid port number received for iscsi session")
	}

	iscsi_iqn := req.PublishContext[disk.ISCSIIQN]
	iscsi_ip := req.PublishContext[disk.ISCSIIP]

	multipathDevicesList = append(multipathDevicesList, core.MultipathDevice{
		Iqn:  &iscsi_iqn,
		Ipv4: &iscsi_ip,
		Port: &port,
	})

	return multipathDevicesList, nil
}
