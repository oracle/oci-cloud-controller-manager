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

type mockFlexvolumeDriver struct{}

func (driver mockFlexvolumeDriver) Init() DriverStatus {
	return Succeed()
}

func (driver mockFlexvolumeDriver) Attach(opts Options, nodeName string) DriverStatus {
	return NotSupported()
}

func (driver mockFlexvolumeDriver) Detach(mountDevice, nodeName string) DriverStatus {
	return Succeed()
}

func (driver mockFlexvolumeDriver) WaitForAttach(mountDevice string, opts Options) DriverStatus {
	return Succeed()
}

func (driver mockFlexvolumeDriver) IsAttached(opts Options, nodeName string) DriverStatus {
	return Succeed()
}

func (driver mockFlexvolumeDriver) MountDevice(mountDir, mountDevice string, opts Options) DriverStatus {
	return Succeed()
}

func (driver mockFlexvolumeDriver) UnmountDevice(mountDevice string) DriverStatus {
	return Succeed()
}

func (driver mockFlexvolumeDriver) Mount(mountDir string, opts Options) DriverStatus {
	return Succeed()
}

func (driver mockFlexvolumeDriver) Unmount(mountDir string) DriverStatus {
	return Succeed()
}
