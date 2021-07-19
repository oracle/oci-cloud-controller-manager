// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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

package flexvolume

import (
	"go.uber.org/zap"
)

type mockFlexvolumeDriver struct{}

func (driver mockFlexvolumeDriver) Init(logger *zap.SugaredLogger) DriverStatus {
	return Succeed(logger)
}

func (driver mockFlexvolumeDriver) Attach(logger *zap.SugaredLogger, opts Options, nodeName string) DriverStatus {
	return NotSupported(logger)
}

func (driver mockFlexvolumeDriver) Detach(logger *zap.SugaredLogger, mountDevice, nodeName string) DriverStatus {
	return Succeed(logger)
}

func (driver mockFlexvolumeDriver) WaitForAttach(mountDevice string, opts Options) DriverStatus {
	return DriverStatus{
		Status: StatusSuccess,
		Device: mountDevice,
	}
}

func (driver mockFlexvolumeDriver) IsAttached(logger *zap.SugaredLogger, opts Options, nodeName string) DriverStatus {
	return Succeed(logger)
}

func (driver mockFlexvolumeDriver) MountDevice(logger *zap.SugaredLogger, mountDir, mountDevice string, opts Options) DriverStatus {
	return Succeed(logger)
}

func (driver mockFlexvolumeDriver) UnmountDevice(logger *zap.SugaredLogger, mountDevice string) DriverStatus {
	return Succeed(logger)
}

func (driver mockFlexvolumeDriver) Mount(logger *zap.SugaredLogger, mountDir string, opts Options) DriverStatus {
	return Succeed(logger)
}

func (driver mockFlexvolumeDriver) Unmount(logger *zap.SugaredLogger, mountDir string) DriverStatus {
	return Succeed(logger)
}
