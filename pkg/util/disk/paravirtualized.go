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
	"k8s.io/mount-utils"
	"k8s.io/utils/exec"
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
		mounter: mount.New(mountCommand),
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

func (c *pvMounter) UnmountPath(path string) error {
	return UnmountPath(c.logger, path, c.mounter)
}

func (c *pvMounter) Resize(devicePath string, volumePath string) (bool, error) {
	resizefs := mount.NewResizeFs(c.runner)
	return resizefs.Resize(devicePath, volumePath)
}
