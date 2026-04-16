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
	"fmt"
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/disk"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/mount-utils"
	"k8s.io/utils/exec"
)

func Test_getDevicePathAndAttachmentType(t *testing.T) {
	type args struct {
		path []string
	}
	tests := []struct {
		name           string
		args           args
		attachmentType string
		diskByPath     string
		wantErr        bool
	}{
		{
			"Testing PV path with digits only",
			args{path: []string{"/dev/disk/by-path/pci-0000:18:00.0-scsi-0:0:0:5"}},
			"paravirtualized",
			"/dev/disk/by-path/pci-0000:18:00.0-scsi-0:0:0:5",
			false,
		},
		{
			"Testing PV path with hexadecimal controller",
			args{path: []string{"/dev/disk/by-path/pci-0000:1a:00.0-scsi-0:0:4:1"}},
			"paravirtualized",
			"/dev/disk/by-path/pci-0000:1a:00.0-scsi-0:0:4:1",
			false,
		},
		{
			"Testing PV path with hexadecimal Bus",
			args{path: []string{"/dev/disk/by-path/pci-0000:00:ff.0-scsi-0:0:0:1"}},
			"paravirtualized",
			"/dev/disk/by-path/pci-0000:00:ff.0-scsi-0:0:0:1",
			false,
		},
		{
			"Testing ISCSI path",
			args{path: []string{"/dev/disk/by-path/ip-169.254.2.19:3260-iscsi-iqn.2015-12.com.oracleiaas:d0ee92cb-5220-423e-b029-07ae2b2ff08f-lun-1"}},
			"iscsi",
			"/dev/disk/by-path/ip-169.254.2.19:3260-iscsi-iqn.2015-12.com.oracleiaas:d0ee92cb-5220-423e-b029-07ae2b2ff08f-lun-1",
			false,
		},
		{
			"Testing UHP volume multidevice path",
			args{path: []string{"/dev/mapper/mpathd"}},
			"iscsi",
			"/dev/mapper/mpathd",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getDevicePathAndAttachmentType(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDevicePathAndAttachmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.attachmentType {
				t.Errorf("getDevicePathAndAttachmentType() got = %v, want attachmentType %v", got, tt.attachmentType)
			}
			if got1 != tt.diskByPath {
				t.Errorf("getDevicePathAndAttachmentType() got1 = %v, want diskByPath %v", got1, tt.diskByPath)
			}
		})
	}
}

func Test_alreadyDeletedPathCheck(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Error contains 'does not exist'",
			args{err: fmt.Errorf("path /some/path does not exist")},
			true,
		},
		{
			"Nil error",
			args{err: nil},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := alreadyDeletedPathCheck(tt.args.err)
			if got != tt.want {
				t.Errorf("alreadyDeletedPathCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockMounter implements Interface.
type mockMounter struct {
	runner                    exec.Interface
	mounter                   mount.Interface
	logger                    *zap.SugaredLogger
	devicePathExistWaitError  error
	ISCSILogoutOnFailureError error
}

func (m mockMounter) WaitForDevicePathToExist(ctx context.Context, disk *disk.Disk, logger *zap.SugaredLogger) (string, error) {
	if m.devicePathExistWaitError != nil {
		return "", m.devicePathExistWaitError
	}
	return "", nil
}

func (m mockMounter) GetMultipathIscsiDevicePath(ctx context.Context, consistentDevicePath string, logger *zap.SugaredLogger) (string, error) {
	if consistentDevicePath == "incorrectDevice" {
		return "", status.Error(codes.DeadlineExceeded, "context deadline exceeded")
	}
	return "", nil
}

func (m mockMounter) AddToDB() error {
	return nil
}

func (m mockMounter) FormatAndMount(source string, target string, fstype string, options []string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) Mount(source string, target string, fstype string, options []string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) Login() error {
	return nil
}

func (m mockMounter) Logout() error {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) DeviceOpened(pathname string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) IsMounted(devicePath string, targetPath string) (bool, error) {
	fmt.Println("IsMounted", devicePath, targetPath)
	if targetPath == "idempotency-check-failure" {
		return false, fmt.Errorf("failure during idempotency check")
	} else if targetPath == "idempotency-check-success" {
		return true, nil
	}
	return false, nil
}

func (m mockMounter) UpdateQueueDepth() error {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) RemoveFromDB() error {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) SetAutomaticLogin() error {
	return nil
}

func (m mockMounter) UnmountPath(path string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) Rescan(devicePath string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) Resize(devicePath string, volumePath string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) WaitForVolumeLoginOrTimeout(ctx context.Context, multipathDevices []core.MultipathDevice) error {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) GetDiskFormat(devicePath string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) WaitForPathToExist(path string, maxRetries int) bool {
	//TODO implement me
	panic("implement me")
}

func (m mockMounter) ISCSILogoutOnFailure() error {
	if m.ISCSILogoutOnFailureError != nil {
		return m.ISCSILogoutOnFailureError
	}
	return nil
}

func TestNodeStageVolume(t *testing.T) {
	multiPathDevicesJson := []byte{}
	var err error
	multiPathDevicesJson, err = json.Marshal([]core.MultipathDevice{
		{
			Ipv4: common.String("1.2.3.4"),
			Iqn:  common.String("iqn.2016-09.com.oraclecloud"),
			Port: common.Int(3034),
		},
	})
	if err != nil {
		t.Fatalf("Error constructing multipath devices Json: %v", err)
	}
	multipathDevicesString := string(multiPathDevicesJson)

	testCases := []struct {
		name        string
		req         *csi.NodeStageVolumeRequest
		setup       func(m *mockMounter)
		expectedErr error
	}{
		{
			name:        "Volume ID not present",
			req:         &csi.NodeStageVolumeRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "Volume ID must be provided"),
		},
		{
			name: "Publish context not present",
			req: &csi.NodeStageVolumeRequest{
				VolumeId: "ocid.abcd",
			},
			expectedErr: status.Error(codes.InvalidArgument, "PublishContext must be provided"),
		},
		{
			name: "Staging path not present",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:       "ocid.abcd",
				PublishContext: map[string]string{"attach-type": "iscsi"},
			},
			expectedErr: status.Error(codes.InvalidArgument, "Staging Target Path must be provided"),
		},
		{
			name: "Volume Capability not present",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "ocid.abcd",
				PublishContext:    map[string]string{"attach-type": "iscsi"},
				StagingTargetPath: "/staging-path",
			},
			expectedErr: status.Error(codes.InvalidArgument, "Volume Capability must be provided"),
		},
		{
			name: "Wrong value for multipath enabled",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "ocid.abcd",
				PublishContext:    map[string]string{multipathEnabled: "yes"},
				StagingTargetPath: "/staging-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "failed to determine if volume is multipath enabled: strconv.ParseBool: parsing \"yes\": invalid syntax"),
		},
		{
			name: "UHP - Invalid multipath device list",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "ocid.abcd",
				PublishContext:    map[string]string{multipathEnabled: "true", multipathDevices: "Not a valid device list"},
				StagingTargetPath: "/staging-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "Failed to get multipath device list for multipath enabled volume: rpc error: code = Internal desc = Failed to get multipath devices from publish context."),
		},
		{
			name: "UHP - Error finding friendly name for multipath device",
			req: &csi.NodeStageVolumeRequest{
				VolumeId: "ocid.abcd",
				PublishContext: map[string]string{multipathEnabled: "true", multipathDevices: multipathDevicesString, disk.ISCSIPORT: "3043",
					disk.ISCSIIQN: "iqn.2016-09.com.oraclecloud", disk.ISCSIIP: "1.2.3.4", device: "incorrectDevice"},
				StagingTargetPath: "/staging-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "Failed to get device path for multipath enabled volume: rpc error: code = DeadlineExceeded desc = context deadline exceeded"),
		},
		{
			name: "ISCSI - Information missing in publish context",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "ocid.abcd",
				PublishContext:    map[string]string{disk.ISCSIPORT: "3043", disk.ISCSIIP: "1.2.3.4"},
				StagingTargetPath: "/staging-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "PublishContext is invalid: unable to get the IQN from the attribute list"),
		},
		{
			name: "ISCSI - Error getting target IP in IPv6 single stack cluster",
			req: &csi.NodeStageVolumeRequest{
				VolumeId:          "ocid.abcd",
				PublishContext:    map[string]string{disk.ISCSIPORT: "3043", disk.ISCSIIP: "fd00:c1::a9fe:a9fe", disk.ISCSIIQN: "iqn.2016-09.com.oraclecloud"},
				StagingTargetPath: "/staging-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, "Failed get ipv6 address for Iscsi Target: invalid iSCSIIp identified fd00:c1::a9fe:a9fe"),
		},
		{
			name: "Paravirtualized - Unable to get device from publish context",
			req: &csi.NodeStageVolumeRequest{
				VolumeId: "ocid.abcd",
				PublishContext: map[string]string{disk.ISCSIPORT: "3043", disk.ISCSIIP: "fd00:c1::a9fe:a9fe",
					disk.ISCSIIQN: "iqn.2016-09.com.oraclecloud", attachmentType: attachmentTypeParavirtualized},
				StagingTargetPath: "/staging-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "Unable to get the device from the attribute list"),
		},
		{
			name: "Unknown attachment type",
			req: &csi.NodeStageVolumeRequest{
				VolumeId: "ocid.abcd",
				PublishContext: map[string]string{disk.ISCSIPORT: "3043", disk.ISCSIIP: "fd00:c1::a9fe:a9fe",
					disk.ISCSIIQN: "iqn.2016-09.com.oraclecloud", attachmentType: "unknown-attachment"},
				StagingTargetPath: "/staging-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "unknown attachment type. supported attachment types are iscsi and paravirtualized"),
		},
		{
			name: "Error acquiring lock",
			req: &csi.NodeStageVolumeRequest{
				VolumeId: "Test-volume",
				PublishContext: map[string]string{disk.ISCSIPORT: "3043", disk.ISCSIIP: "1.2.3.4",
					disk.ISCSIIQN: "iqn.2016-09.com.oraclecloud"},
				StagingTargetPath: "/staging-path",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			expectedErr: status.Errorf(codes.Aborted, volumeOperationAlreadyExistsFmt, "Test-volume"),
		},
		{
			name: "Error from IsMounted during idempotency check",
			req: &csi.NodeStageVolumeRequest{
				VolumeId: "Idempotency-check-failure",
				PublishContext: map[string]string{disk.ISCSIPORT: "3043", disk.ISCSIIP: "1.2.3.4",
					disk.ISCSIIQN: "iqn.2016-09.com.oraclecloud"},
				StagingTargetPath: "idempotency-check-failure",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			expectedErr: status.Error(codes.Internal, "failure during idempotency check"),
		},
		{
			name: "Idempotency check success",
			req: &csi.NodeStageVolumeRequest{
				VolumeId: "Test-volume",
				PublishContext: map[string]string{disk.ISCSIPORT: "3043", disk.ISCSIIP: "1.2.3.4",
					disk.ISCSIIQN: "iqn.2016-09.com.oraclecloud"},
				StagingTargetPath: "idempotency-check-success",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Error while logging out after WaitForDevicePathToExist failure",
			req: &csi.NodeStageVolumeRequest{
				VolumeId: "Test-volume",
				PublishContext: map[string]string{disk.ISCSIPORT: "3043", disk.ISCSIIP: "1.2.3.4",
					disk.ISCSIIQN: "iqn.2016-09.com.oraclecloud"},
				StagingTargetPath: "test",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			setup: func(m *mockMounter) {
				m.devicePathExistWaitError = fmt.Errorf("device exist error")
				m.ISCSILogoutOnFailureError = fmt.Errorf("logout error")
			},
			expectedErr: status.Errorf(codes.Internal, "Failed to iscsi logout after timeout on waiting for device path to exist: %v", "logout error"),
		},
		{
			name: "Error while logging out after WaitForDevicePathToExist failure",
			req: &csi.NodeStageVolumeRequest{
				VolumeId: "Test-volume",
				PublishContext: map[string]string{disk.ISCSIPORT: "3043", disk.ISCSIIP: "1.2.3.4",
					disk.ISCSIIQN: "iqn.2016-09.com.oraclecloud"},
				StagingTargetPath: "test",
				VolumeCapability: &csi.VolumeCapability{
					AccessMode: &csi.VolumeCapability_AccessMode{
						Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
					},
				},
			},
			setup: func(m *mockMounter) {
				m.devicePathExistWaitError = fmt.Errorf("device exist error")
				m.ISCSILogoutOnFailureError = fmt.Errorf("logout error")
			},
			expectedErr: status.Errorf(codes.Internal, "Failed to iscsi logout after timeout on waiting for device path to exist: %v", "logout error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMounter := &mockMounter{}
			if tc.setup != nil {
				tc.setup(testMounter)
			}
			testMounterFactory := func(attachment string, scsi *disk.Disk, multipath bool, log *zap.SugaredLogger) (disk.Interface, error) {
				return testMounter, nil
			}
			var driver *BlockVolumeNodeDriver
			if tc.name == "ISCSI - Error getting target IP in IPv6 single stack cluster" {
				driver = &BlockVolumeNodeDriver{
					NodeDriver: NodeDriver{
						logger:         logging.Logger().Sugar(),
						mounterFactory: testMounterFactory,
						nodeMetadata: &csi_util.NodeMetadata{
							PreferredNodeIpFamily: "IPv6",
						},
					},
				}
			} else if tc.name == "Error acquiring lock" {
				driver = &BlockVolumeNodeDriver{
					NodeDriver: NodeDriver{
						logger:         logging.Logger().Sugar(),
						mounterFactory: testMounterFactory,
						volumeLocks:    csi_util.NewVolumeLocks(),
						nodeMetadata: &csi_util.NodeMetadata{
							PreferredNodeIpFamily: "IPv4",
						},
					},
				}
				if acquired := driver.volumeLocks.TryAcquire("Test-volume"); !acquired {
					t.Fatalf("Error acquiring volume lock for unit test")
				}
			} else {
				driver = &BlockVolumeNodeDriver{
					NodeDriver: NodeDriver{
						logger:         logging.Logger().Sugar(),
						mounterFactory: testMounterFactory,
						volumeLocks:    csi_util.NewVolumeLocks(),
						nodeMetadata: &csi_util.NodeMetadata{
							PreferredNodeIpFamily: "IPv4",
						},
					},
				}
			}
			_, err := driver.NodeStageVolume(t.Context(), tc.req)
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Fatalf("Expected error '%v' but got '%v'", tc.expectedErr, err)
			}
		})
	}
}
