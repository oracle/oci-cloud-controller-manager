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

package disk

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	"k8s.io/mount-utils"
	"k8s.io/utils/exec"
	"path/filepath"
)

// iSCSIMounter implements Interface.
type pvMounter struct {
	runner  exec.Interface
	mounter mount.Interface

	// iscsiadmPath is the cached absolute path to iscsiadm.
	iscsiadmPath string
	logger       *zap.SugaredLogger
}

// NewFromPVDisk creates a new PV handler from PVDisk.
func NewFromPVDisk(logger *zap.SugaredLogger) Interface {
	return &pvMounter{
		runner:  exec.New(),
		mounter: mount.New(mountCommand),
		logger:  logger,
	}
}

func (c *pvMounter) WaitForVolumeLoginOrTimeout(ctx context.Context, multipathDevice []core.MultipathDevice) error {
	c.logger.Info("Attachment type paravirtualized. WaitForVolumeLoginOrTimeout() not needed for paravirtualized attachment")
	return nil
}

func (c *pvMounter) AddToDB() error {
	c.logger.Info("Attachment type paravirtualized. AddToDB() not needed for paravirtualized attachment")
	return nil
}

func (c *pvMounter) SetAutomaticLogin() error {
	c.logger.Info("Attachment type paravirtualized. SetAutomaticLogin() not needed for paravirtualized attachment")
	return nil
}

func (c *pvMounter) Login() error {
	c.logger.Info("Attachment type paravirtualized. Login() not needed for paravirtualized attachment")
	return nil
}

func (c *pvMounter) Logout() error {
	c.logger.Info("Attachment type paravirtualized. Logout() not needed for paravirtualized attachment")
	return nil
}

func (c *pvMounter) UpdateQueueDepth() error {
	c.logger.Info("Attachment type paravirtualized. UpdateQueueDepth() not needed for paravirtualized attachment")
	return nil
}

func (c *pvMounter) RemoveFromDB() error {
	c.logger.Info("Attachment type paravirtualized. RemoveFromDB() not needed for paravirtualized attachment")
	return nil
}

func (c *pvMounter) FormatAndMount(source string, target string, fstype string, options []string) error {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Exec:      c.runner,
	}
	return formatAndMount(source, target, fstype, options, safeMounter)
}

func (c *pvMounter) Mount(source string, target string, fstype string, options []string) error {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Exec:      c.runner,
	}
	return mnt(source, target, fstype, options, safeMounter)
}

func (c *pvMounter) DeviceOpened(pathname string) (bool, error) {
	return deviceOpened(pathname, c.logger)
}

func (c *pvMounter) IsMounted(devicePath string, targetPath string) (bool, error) {
	notMnt, err := c.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err){
			return false, nil
		}
		return false, fmt.Errorf("failed to check if %s is a mount point: %v", targetPath, err)
	}
	if notMnt {
		return false, nil
	}

	resolvedDevicePath, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return false, fmt.Errorf("failed to resolve symlink for consistent device path %s: %v", devicePath, err)
	}

	mounts, err := c.mounter.List()
	if err != nil {
		return false, fmt.Errorf("could not list mount points: %v", err)
	}

	for _, m := range mounts {
		if m.Path == targetPath {
			if m.Device == resolvedDevicePath {
				return true, nil
			}
			return false, fmt.Errorf("expected device %s but found %s mounted at %s", resolvedDevicePath, m.Device, targetPath)
		}
	}
	return false, nil
}

func (c *pvMounter) ISCSILogoutOnFailure() error {
	return nil
}

func (c *pvMounter) UnmountPath(path string) error {
	return UnmountPath(c.logger, path, c.mounter)
}

func (c *pvMounter) Rescan(devicePath string) error {
	return Rescan(c.logger, devicePath)
}

func (c *pvMounter) Resize(devicePath string, volumePath string) (bool, error) {
	resizefs := mount.NewResizeFs(c.runner)
	return resizefs.Resize(devicePath, volumePath)
}

func (c *pvMounter) GetDiskFormat(disk string) (string, error) {
	return getDiskFormat(c.runner, disk, c.logger)
}

func waitForPathToExist(path string, maxRetries int) bool {
	for i := 0; i < maxRetries; i++ {
		var err error
		_, err = os.Stat(path)
		if err == nil {
			return true
		}
		if err != nil && !os.IsNotExist(err) {
			return false
		}
		if i == maxRetries-1 {
			break
		}
		time.Sleep(waitForPathDelay)
	}
	return false
}

func (c *pvMounter) WaitForPathToExist(path string, maxRetries int) bool {
	return waitForPathToExist(path, maxRetries)
}
