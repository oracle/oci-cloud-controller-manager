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
	"go.uber.org/zap"
	"k8s.io/utils/exec"

	"github.com/oracle/oci-cloud-controller-manager/pkg/util/mount"
)

// iSCSIMounter implements Interface.
type pvMounter struct {
	runner  exec.Interface
	mounter mount.Interface

	// iscsiadmPath is the cached absolute path to iscsiadm.
	iscsiadmPath string
	logger       *zap.SugaredLogger
}

//NewFromPVDisk creates a new PV handler from PVDisk.
func NewFromPVDisk(logger *zap.SugaredLogger) Interface {
	return &pvMounter{
		runner:  exec.New(),
		mounter: mount.New(logger, mountCommand),
		logger:  logger,
	}
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

func (c *pvMounter) RemoveFromDB() error {
	c.logger.Info("Attachment type paravirtualized. RemoveFromDB() not needed for paravirtualized attachment")
	return nil
}

func (c *pvMounter) DeviceOpened(path string) (bool, error) {
	return c.mounter.DeviceOpened(path)
}

func (c *pvMounter) FormatAndMount(source string, target string, fstype string, options []string) error {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Runner:    c.runner,
		Logger:    c.logger,
	}
	return formatAndMount(source, target, fstype, options, safeMounter)
}

func (c *pvMounter) Mount(source string, target string, fstype string, options []string) error {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Runner:    c.runner,
		Logger:    c.logger,
	}
	return mnt(source, target, fstype, options, safeMounter)
}

func (c *pvMounter) UnmountPath(path string) error {
	return mount.UnmountPath(c.logger, path, c.mounter)
}

func (c *pvMounter) Resize(devicePath string, volumePath string) (bool, error) {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Runner:    c.runner,
		Logger:    c.logger,
	}
	return resize(devicePath, volumePath, safeMounter)
}


func (c *pvMounter) Rescan(devicePath string) error {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Runner:    c.runner,
		Logger:    c.logger,
	}
	return rescan(devicePath, safeMounter)
}

func (c *pvMounter) GetBlockSizeBytes(devicePath string) (int64, error) {
	safeMounter := &mount.SafeFormatAndMount{
		Interface: c.mounter,
		Runner:    c.runner,
		Logger:    c.logger,
	}
	return getBlockSizeBytes(devicePath, safeMounter)
}
