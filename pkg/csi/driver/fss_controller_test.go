// Copyright 2023 Oracle and/or its affiliates. All rights reserved.
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
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/filestorage"
	fss "github.com/oracle/oci-go-sdk/v65/filestorage"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

type MockFileStorageClient struct {
	filestorage util.MockOCIFileStorageClient
}

var (
	mountTargets = map[string]*fss.MountTarget{
		"mount-target-stuck-creating": {
			DisplayName:        common.String("mount-target-stuck-creating"),
			LifecycleState:     fss.MountTargetLifecycleStateCreating,
			AvailabilityDomain: common.String("NWuj:PHX-AD-2"),
			Id:                 common.String("mount-target-stuck-creating"),
		},
	}

	fileSystems = map[string]*fss.FileSystem{
		"file-system-stuck-creating": {
			DisplayName:        common.String("file-system-stuck-creating"),
			LifecycleState:     fss.FileSystemLifecycleStateCreating,
			AvailabilityDomain: common.String("NWuj:PHX-AD-2"),
			Id:                 common.String("file-system-stuck-creating"),
		},
	}

	exports = map[string]*fss.Export{
		"export-stuck-creating": {
			LifecycleState: fss.ExportLifecycleStateCreating,
			Id:             common.String("export-stuck-creating"),
		},
	}
)

func (c *MockFileStorageClient) GetMountTarget(ctx context.Context, id string) (*filestorage.MountTarget, error) {
	if mountTargets[id] != nil {
		return mountTargets[id], nil
	}
	idMt := "oc1.mounttarget.xxxx"
	ad := "zkJl:US-ASHBURN-AD-1"
	privateIpIds := []string{"10.0.20.1"}
	displayName := "mountTarget"
	idEx := "oc1.export.xxxx"
	return &filestorage.MountTarget{
		Id:                 &idMt,
		AvailabilityDomain: &ad,
		DisplayName:        &displayName,
		PrivateIpIds:       privateIpIds,
		ExportSetId:        &idEx,
		LifecycleState:     fss.MountTargetLifecycleStateActive,
	}, nil
}

func (c *MockFileStorageClient) GetMountTargetSummaryByDisplayName(ctx context.Context, compartmentID, ad, mountTargetName string) (bool, []filestorage.MountTargetSummary, error) {
	if mountTargetName == "mount-target-idempotency-check-timeout-volume" {
		var page *string
		var requestMetadata common.RequestMetadata
		mountTargetSummaries := make([]fss.MountTargetSummary, 0)
		conflictingMountTargetSummaries := make([]fss.MountTargetSummary, 0)
		foundConflicting := false
		for {

			resp, err := c.filestorage.ListMountTargets(ctx, fss.ListMountTargetsRequest{
				CompartmentId:      &compartmentID,
				AvailabilityDomain: &ad,
				DisplayName:        &mountTargetName,
				RequestMetadata:    requestMetadata,
			})
			if err != nil {
				return foundConflicting, nil, errors.WithStack(err)
			}

			for _, mountTargetSummary := range resp.Items {
				lifecycleState := mountTargetSummary.LifecycleState
				if lifecycleState == fss.MountTargetSummaryLifecycleStateActive ||
					lifecycleState == fss.MountTargetSummaryLifecycleStateCreating {
					mountTargetSummaries = append(mountTargetSummaries, mountTargetSummary)
				} else {
					conflictingMountTargetSummaries = append(conflictingMountTargetSummaries, mountTargetSummary)
					foundConflicting = true
				}
			}

			if page = resp.OpcNextPage; page == nil {
				break
			}
		}
	}
	return false, []filestorage.MountTargetSummary{}, nil
}

// CreateFileSystem mocks the FileStorage CreateFileSystem implementation.
func (c *MockFileStorageClient) CreateFileSystem(ctx context.Context, details filestorage.CreateFileSystemDetails) (*filestorage.FileSystem, error) {
	idFs := *details.DisplayName
	ad := "zkJl:US-ASHBURN-AD-1"
	return &filestorage.FileSystem{
		Id:                 &idFs,
		AvailabilityDomain: &ad,
	}, nil
}

// GetFileSystem mocks the FileStorage GetFileSystem implementation.
func (c *MockFileStorageClient) GetFileSystem(ctx context.Context, id string) (*filestorage.FileSystem, error) {
	if fileSystems[id] != nil {
		return fileSystems[id], nil
	}
	idFs := id
	ad := "zkJl:US-ASHBURN-AD-1"
	displayName := id
	compartmentOcid := "oc1.comp.xxxx"
	return &filestorage.FileSystem{
		Id:                 &idFs,
		AvailabilityDomain: &ad,
		DisplayName:        &displayName,
		CompartmentId:      &compartmentOcid,
		LifecycleState:     fss.FileSystemLifecycleStateActive,
	}, nil
}

func (c *MockFileStorageClient) AwaitFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.FileSystem, error) {
	var fs *fss.FileSystem
	err := wait.PollImmediateUntil(testPollInterval, func() (bool, error) {
		var err error
		fs, err = c.GetFileSystem(ctx, id)
		if err != nil {
			return false, err
		}
		switch state := fs.LifecycleState; state {
		case fss.FileSystemLifecycleStateActive:
			return true, nil
		case fss.FileSystemLifecycleStateDeleting, fss.FileSystemLifecycleStateDeleted, fss.FileSystemLifecycleStateFailed:
			return false, errors.Errorf("file system %q is in lifecycle state %q", *fs.Id, state)
		default:
			return false, nil
		}
	}, ctx.Done())
	if err != nil {
		return nil, err
	}
	idFs := id
	ad := "zkJl:US-ASHBURN-AD-1"
	displayName := "filesystem"
	return &filestorage.FileSystem{
		Id:                 &idFs,
		AvailabilityDomain: &ad,
		DisplayName:        &displayName,
	}, nil
}

func (c *MockFileStorageClient) GetFileSystemSummaryByDisplayName(ctx context.Context, compartmentID, ad, displayName string) (bool, []filestorage.FileSystemSummary, error) {
	if displayName == "file-system-idempotency-check-timeout-volume" {
		var page *string
		fileSystemSummaries := make([]fss.FileSystemSummary, 0)
		conflictingFileSystemSummaries := make([]fss.FileSystemSummary, 0)
		foundConflicting := false
		var requestMetadata common.RequestMetadata
		for {
			resp, err := c.filestorage.ListFileSystems(ctx, fss.ListFileSystemsRequest{
				CompartmentId:      &compartmentID,
				AvailabilityDomain: &ad,
				DisplayName:        &displayName,
				RequestMetadata:    requestMetadata,
			})
			if err != nil {
				return foundConflicting, nil, errors.WithStack(err)
			}

			for _, fileSystemSummary := range resp.Items {
				lifecycleState := fileSystemSummary.LifecycleState
				if lifecycleState == fss.FileSystemSummaryLifecycleStateActive ||
					lifecycleState == fss.FileSystemSummaryLifecycleStateCreating {
					fileSystemSummaries = append(fileSystemSummaries, fileSystemSummary)
				} else {
					conflictingFileSystemSummaries = append(fileSystemSummaries, fileSystemSummary)
					foundConflicting = true
				}
			}

			if page = resp.OpcNextPage; page == nil {
				break
			}
		}

		if foundConflicting {
			return foundConflicting, conflictingFileSystemSummaries, errors.Errorf("Found file system summary neither active nor creating state")
		}
		return foundConflicting, fileSystemSummaries, nil
	}
	idFs := displayName
	fileSystemSummary := filestorage.FileSystemSummary{
		Id: &idFs,
	}
	fileSystemSummaries := []filestorage.FileSystemSummary{fileSystemSummary}
	return false, fileSystemSummaries, nil
}

// DeleteFileSystem mocks the FileStorage DeleteFileSystem implementation
func (c *MockFileStorageClient) DeleteFileSystem(ctx context.Context, id string) error {
	return nil
}

// CreateExport mocks the FileStorage CreateExport implementation
func (c *MockFileStorageClient) CreateExport(ctx context.Context, details filestorage.CreateExportDetails) (*filestorage.Export, error) {
	idEx := "oc1.export.xxxx"
	idFs := *details.FileSystemId
	return &filestorage.Export{
		Id:           &idEx,
		FileSystemId: &idFs,
	}, nil
}

// GetExport mocks the FileStorage CreateExport implementation.
func (c *MockFileStorageClient) GetExport(ctx context.Context, id string) (*fss.Export, error) {
	if exports[id] != nil {
		return exports[id], nil
	}
	return &fss.Export{}, nil
}

func (c *MockFileStorageClient) AwaitExportActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.Export, error) {
	logger.Info("Waiting for Export to be in lifecycle state ACTIVE")

	var export *fss.Export
	if err := wait.PollImmediateUntil(testPollInterval, func() (bool, error) {
		logger.Debug("Polling export lifecycle state")

		var err error
		export, err = c.GetExport(ctx, id)
		if err != nil {
			return false, err
		}

		switch state := export.LifecycleState; state {
		case fss.ExportLifecycleStateActive:
			logger.Infof("Export is in lifecycle state %q", state)
			return true, nil
		case fss.ExportLifecycleStateDeleting, fss.ExportLifecycleStateDeleted:
			logger.Errorf("Export is in lifecycle state %q", state)
			return false, fmt.Errorf("export %q is in lifecycle state %q", *export.Id, state)
		default:
			logger.Debugf("Export is in lifecycle state %q", state)
			return false, nil
		}
	}, ctx.Done()); err != nil {
		return nil, err
	}
	idEx := "oc1.export.xxxx"
	idFs := "oc1.filesystem.xxxx"
	return &filestorage.Export{
		Id:           &idEx,
		FileSystemId: &idFs,
	}, nil
}

func (c *MockFileStorageClient) FindExport(ctx context.Context, fsID, path, exportSetID string) (*filestorage.ExportSummary, error) {
	var page *string
	var requestMetadata common.RequestMetadata
	for {
		resp, err := c.filestorage.ListExports(ctx, fss.ListExportsRequest{
			FileSystemId:    &fsID,
			ExportSetId:     &exportSetID,
			Page:            page,
			RequestMetadata: requestMetadata,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		for _, export := range resp.Items {
			if *export.Path == path {
				if export.LifecycleState == fss.ExportSummaryLifecycleStateCreating ||
					export.LifecycleState == fss.ExportSummaryLifecycleStateActive {
					return &export, nil
				}
				return &export, errors.Errorf("Found export in conflicting state %s: %s", *export.Id, export.LifecycleState)
			}
		}
		if page = resp.OpcNextPage; resp.OpcNextPage == nil {
			break
		}
	}
	idEx := "oc1.export.xxxx"
	idFs := fsID
	lifeCycleStatus := filestorage.ExportSummaryLifecycleStateActive
	return &filestorage.ExportSummary{
		ExportSetId:    &idEx,
		FileSystemId:   &idFs,
		LifecycleState: lifeCycleStatus,
	}, nil
}

// DeleteExport mocks the FileStorage DeleteExport implementation
func (c *MockFileStorageClient) DeleteExport(ctx context.Context, id string) error {
	return nil
}

// GetMountTarget mocks the FileStorage GetMountTarget implementation
func (c *MockFileStorageClient) AwaitMountTargetActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.MountTarget, error) {
	var mt *fss.MountTarget
	if err := wait.PollImmediateUntil(testPollInterval, func() (bool, error) {
		var err error
		mt, err = c.GetMountTarget(ctx, id)
		if err != nil {
			return false, err
		}

		switch state := mt.LifecycleState; state {
		case fss.MountTargetLifecycleStateActive:
			logger.Infof("Mount target is in lifecycle state %q", state)
			return true, nil
		case fss.MountTargetLifecycleStateFailed,
			fss.MountTargetLifecycleStateDeleting,
			fss.MountTargetLifecycleStateDeleted:
			return false, fmt.Errorf("mount target %q is in lifecycle state %q and will not become ACTIVE", *mt.Id, state)
		default:
			logger.Debugf("Mount target is in lifecycle state %q", state)
			return false, nil
		}
	}, ctx.Done()); err != nil {
		return nil, err
	}
	idMt := "oc1.mounttarget.xxxx"
	ad := "zkJl:US-ASHBURN-AD-1"
	privateIpIds := []string{"10.0.20.1"}
	displayName := "mountTarget"
	idEx := "oc1.export.xxxx"
	return &filestorage.MountTarget{
		Id:                 &idMt,
		AvailabilityDomain: &ad,
		DisplayName:        &displayName,
		PrivateIpIds:       privateIpIds,
		ExportSetId:        &idEx,
	}, nil
}

// CreateMountTarget mocks the FileStorage CreateMountTarget implementation.
func (c *MockFileStorageClient) CreateMountTarget(ctx context.Context, details filestorage.CreateMountTargetDetails) (*filestorage.MountTarget, error) {
	if mountTargets[*details.DisplayName] != nil {
		return mountTargets[*details.DisplayName], nil
	}
	idMt := "oc1.mounttarget.xxxx"
	ad := "zkJl:US-ASHBURN-AD-1"
	privateIpIds := []string{"10.0.20.1"}
	displayName := "mountTarget"
	idEx := "oc1.export.xxxx"
	return &filestorage.MountTarget{
		Id:                 &idMt,
		AvailabilityDomain: &ad,
		DisplayName:        &displayName,
		PrivateIpIds:       privateIpIds,
		ExportSetId:        &idEx,
	}, nil
}

// DeleteMountTarget mocks the FileStorage DeleteMountTarget implementation
func (c *MockFileStorageClient) DeleteMountTarget(ctx context.Context, id string) error {
	return nil
}

// FSS mocks client FileStorage implementation
func (p *MockProvisionerClient) FSS() client.FileStorageInterface {
	return &MockFileStorageClient{}
}

type MockFSSProvisionerClient struct {
	Storage *MockFileStorageClient
}

func (m MockFSSProvisionerClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

func (m MockFSSProvisionerClient) LoadBalancer(*zap.SugaredLogger, string, string, *authv1.TokenRequest) client.GenericLoadBalancerInterface {
	return &MockLoadBalancerClient{}
}

func (m MockFSSProvisionerClient) Networking() client.NetworkingInterface {
	return &MockVirtualNetworkClient{}
}

func (m MockFSSProvisionerClient) BlockStorage() client.BlockStorageInterface {
	return &MockBlockStorageClient{}
}

func (m MockFSSProvisionerClient) FSS() client.FileStorageInterface {
	return &MockFileStorageClient{}
}

func (m MockFSSProvisionerClient) Identity() client.IdentityInterface {
	return &MockIdentityClient{}
}

func NewFSSClientProvisioner(pcData client.Interface, storage *MockFileStorageClient) client.Interface {
	return &MockFSSProvisionerClient{Storage: storage}
}

func TestFSSControllerDriver_CreateVolume(t *testing.T) {
	type fields struct {
		KubeClient kubernetes.Interface
		logger     *zap.SugaredLogger
		config     *providercfg.Config
		client     client.Interface
		util       *csi_util.Util
	}
	type args struct {
		ctx context.Context
		req *csi.CreateVolumeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *csi.CreateVolumeResponse
		wantErr error
	}{
		{
			name:   "Error for name not provided for creating volume",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name:       "",
					Parameters: map[string]string{"availabilityDomain": "US-ASHBURN-AD-1", "mountTargetOcid": "oc1.mounttarget.xxxx"},
				},
			},
			want:    nil,
			wantErr: errors.New("CreateVolume Name must be provided"),
		},
		{
			name:   "Error for no VolumeCapabilities provided in CreateVolumeRequest",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name:               "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{},
					Parameters:         map[string]string{"availabilityDomain": "US-ASHBURN-AD-1", "mountTargetOcid": "oc1.mounttarget.xxxx"},
				},
			},
			want:    nil,
			wantErr: errors.New("VolumeCapabilities must be provided in CreateVolumeRequest"),
		},
		{
			name:   "Error for unsupported VolumeCapabilities: SINGLE_NODE_MULTI_WRITER provided in CreateVolumeRequest",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_MULTI_WRITER,
						},
					}},
					Parameters: map[string]string{"availabilityDomain": "US-ASHBURN-AD-1", "mountTargetOcid": "oc1.mounttarget.xxxx"},
				},
			},
			want:    nil,
			wantErr: errors.New("Requested Volume Capability not supported"),
		},
		{
			name:   "Error when Availability Domain is not specified",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name:       "ut-volume",
					Parameters: map[string]string{"mountTargetOcid": "oc1.mounttarget.xxxx"},
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("AvailabilityDomain not provided in storage class"),
		},
		{
			name:   "Error when both mount target OCID and mount target subnet OCID are not specified",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name:       "ut-volume",
					Parameters: map[string]string{"availabilityDomain": "US-ASHBURN-AD-1"},
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("Neither Mount Target Ocid nor Mount Target Subnet Ocid provided in storage class"),
		},
		{
			name:   "Time out during file system idempotency check",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name: "file-system-idempotency-check-timeout-volume",
					Parameters: map[string]string{"availabilityDomain": "US-ASHBURN-AD-1",
						"mountTargetOcid": "oc1.mounttarget.xxxx"},
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("context deadline exceeded"),
		},
		{
			name:   "Time out during mount target idempotency check",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name:       "mount-target-idempotency-check-timeout-volume",
					Parameters: map[string]string{"availabilityDomain": "US-ASHBURN-AD-1", "mountTargetSubnetOcid": "oc1.subnet.xxxx"},
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("context deadline exceeded"),
		},
		{
			name:   "Time out due to mount target stuck in creating state",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name:       "mount-target-stuck-creating",
					Parameters: map[string]string{"availabilityDomain": "US-ASHBURN-AD-1", "mountTargetSubnetOcid": "oc1.subnet.xxxx"},
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("await mount target to be available failed with time out"),
		},
		{
			name:   "Time out due to file system stuck in creating state",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name:       "file-system-stuck-creating",
					Parameters: map[string]string{"availabilityDomain": "US-ASHBURN-AD-1", "mountTargetSubnetOcid": "oc1.subnet.xxxx"},
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("Await File System failed with time out"),
		},
		{
			name:   "Timed out finding export",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name:       "export-idempotency-check-timeout",
					Parameters: map[string]string{"availabilityDomain": "US-ASHBURN-AD-1", "mountTargetSubnetOcid": "oc1.subnet.xxxx"},
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("failed to check existence of export"),
		},
		{
			name:   "Time out due to export stuck in creating state",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.CreateVolumeRequest{
					Name:       "export-stuck-creating",
					Parameters: map[string]string{"availabilityDomain": "US-ASHBURN-AD-1", "mountTargetSubnetOcid": "oc1.subnet.xxxx"},
					VolumeCapabilities: []*csi.VolumeCapability{{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("await export failed with time out"),
		},
	}
	for _, tt := range tests {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()
		t.Run(tt.name, func(t *testing.T) {
			d := &FSSControllerDriver{ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: ""},
				client:     NewClientProvisioner(nil, nil, &MockFileStorageClient{}),
				util:       &csi_util.Util{},
			}}
			got, err := d.CreateVolume(ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && err == nil {
				t.Errorf("want error %q, got none", tt.wantErr)
			} else if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.CreateVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFSSControllerDriver_DeleteVolume(t *testing.T) {
	type fields struct {
		KubeClient kubernetes.Interface
		logger     *zap.SugaredLogger
		config     *providercfg.Config
		client     client.Interface
		util       *csi_util.Util
	}
	type args struct {
		ctx context.Context
		req *csi.DeleteVolumeRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *csi.DeleteVolumeResponse
		wantErr error
	}{
		{
			name:   "Error for volume OCID missing in delete block volume",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.DeleteVolumeRequest{},
			},
			want:    nil,
			wantErr: errors.New("Invalid Volume ID provided "),
		},
		{
			name:   "Error for filesystem OCID missing in volume ID",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.DeleteVolumeRequest{VolumeId: ":10.0.10.207:/export-path"},
			},
			want:    nil,
			wantErr: errors.New("Invalid Volume ID provided :10.0.10.207:/export-path"),
		},
		{
			name:   "Error for mount target IP missing in volume ID",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.DeleteVolumeRequest{VolumeId: "oc1.filesystem.xxxx::/export-path"},
			},
			want:    nil,
			wantErr: errors.New("Invalid Volume ID provided oc1.filesystem.xxxx::/export-path"),
		},
		{
			name:   "Error for export path missing in volume ID",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.DeleteVolumeRequest{VolumeId: "oc1.filesystem.xxxx:10.0.10.207:"},
			},
			want:    nil,
			wantErr: errors.New("Invalid Volume ID provided oc1.filesystem.xxxx:10.0.10.207:"),
		},
		{
			name:   "Delete volume and get empty response",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.DeleteVolumeRequest{VolumeId: "oc1.filesystem.xxxx:10.0.10.207:/export-path"},
			},
			want:    &csi.DeleteVolumeResponse{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &FSSControllerDriver{ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: ""},
				client:     NewClientProvisioner(nil, nil, &MockFileStorageClient{}),
				util:       &csi_util.Util{},
			}}
			got, err := d.DeleteVolume(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.DeleteVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractStorageClassParameters(t *testing.T) {
	tests := map[string]struct {
		parameters                     map[string]string
		expectedStorageClassParameters *StorageClassParameters
		wantErr                        bool
		wantErrMessage                 string
	}{
		"Extract storage class parameters with mountTargetOcid": {
			parameters: map[string]string{
				"availabilityDomain": "US-ASHBURN-AD-1",
				"mountTargetOcid":    "oc1.mounttarget.xxxx",
			},
			expectedStorageClassParameters: &StorageClassParameters{
				availabilityDomain:    "AD1",
				compartmentOcid:       "oc1.compartment.xxxx",
				kmsKey:                "",
				exportPath:            "/ut-volume",
				exportOptions:         []filestorage.ClientOptions{},
				mountTargetOcid:       "oc1.mounttarget.xxxx",
				mountTargetSubnetOcid: "",
				encryptInTransit:      "false",
				scTags:                &config.TagConfig{},
			},
			wantErr:        false,
			wantErrMessage: "",
		},
		"Extract storage class parameters with mountTargetSubnetOcid": {
			parameters: map[string]string{
				"availabilityDomain":    "AD1",
				"mountTargetSubnetOcid": "oc1.subnet.xxxx",
			},
			expectedStorageClassParameters: &StorageClassParameters{
				availabilityDomain:    "AD1",
				compartmentOcid:       "oc1.compartment.xxxx",
				kmsKey:                "",
				exportPath:            "/ut-volume",
				exportOptions:         []filestorage.ClientOptions{},
				mountTargetOcid:       "",
				mountTargetSubnetOcid: "oc1.subnet.xxxx",
				encryptInTransit:      "false",
				scTags:                &config.TagConfig{},
			},
			wantErr:        false,
			wantErrMessage: "",
		},
		"Extract storage class parameters with export-path": {
			parameters: map[string]string{
				"availabilityDomain": "AD1",
				"mountTargetOcid":    "oc1.mounttarget.xxxx",
				"exportPath":         "/abc",
			},
			expectedStorageClassParameters: &StorageClassParameters{
				availabilityDomain:    "AD1",
				compartmentOcid:       "oc1.compartment.xxxx",
				kmsKey:                "",
				exportPath:            "/abc",
				exportOptions:         []filestorage.ClientOptions{},
				mountTargetOcid:       "oc1.mounttarget.xxxx",
				mountTargetSubnetOcid: "",
				encryptInTransit:      "false",
				scTags:                &config.TagConfig{},
			},
			wantErr:        false,
			wantErrMessage: "",
		},
		"Extract storage class parameters with kmskey": {
			parameters: map[string]string{
				"availabilityDomain": "AD1",
				"mountTargetOcid":    "oc1.mounttarget.xxxx",
				"kmsKeyOcid":         "oc1.key.xxxx",
			},
			expectedStorageClassParameters: &StorageClassParameters{
				availabilityDomain:    "AD1",
				compartmentOcid:       "oc1.compartment.xxxx",
				kmsKey:                "oc1.key.xxxx",
				exportPath:            "/ut-volume",
				exportOptions:         []filestorage.ClientOptions{},
				mountTargetOcid:       "oc1.mounttarget.xxxx",
				mountTargetSubnetOcid: "",
				encryptInTransit:      "false",
				scTags:                &config.TagConfig{},
			},
			wantErr:        false,
			wantErrMessage: "",
		},
		"Extract storage class parameters with in-transit encryption": {
			parameters: map[string]string{
				"availabilityDomain": "AD1",
				"mountTargetOcid":    "oc1.mounttarget.xxxx",
				"encryptInTransit":   "true",
			},
			expectedStorageClassParameters: &StorageClassParameters{
				availabilityDomain:    "AD1",
				compartmentOcid:       "oc1.compartment.xxxx",
				kmsKey:                "",
				exportPath:            "/ut-volume",
				exportOptions:         []filestorage.ClientOptions{},
				mountTargetOcid:       "oc1.mounttarget.xxxx",
				mountTargetSubnetOcid: "",
				encryptInTransit:      "true",
				scTags:                &config.TagConfig{},
			},
			wantErr:        false,
			wantErrMessage: "",
		},
		"Extract storage class parameters with different compartment": {
			parameters: map[string]string{
				"availabilityDomain": "AD1",
				"mountTargetOcid":    "oc1.mounttarget.xxxx",
				"compartmentOcid":    "oc1.compartment.yyyy",
			},
			expectedStorageClassParameters: &StorageClassParameters{
				availabilityDomain:    "AD1",
				compartmentOcid:       "oc1.compartment.yyyy",
				kmsKey:                "",
				exportPath:            "/ut-volume",
				exportOptions:         []filestorage.ClientOptions{},
				mountTargetOcid:       "oc1.mounttarget.xxxx",
				mountTargetSubnetOcid: "",
				encryptInTransit:      "false",
				scTags:                &config.TagConfig{},
			},
			wantErr:        false,
			wantErrMessage: "",
		},
		"Error when availabilityDomain is not passed": {
			parameters: map[string]string{
				"mountTargetOcid": "oc1.mounttarget.xxxx",
			},
			expectedStorageClassParameters: &StorageClassParameters{},
			wantErr:                        true,
			wantErrMessage:                 "AvailabilityDomain not provided in storage class",
		},
		"Error when mountTargetOcid and mountTargetSubnetOcid is not passed": {
			parameters: map[string]string{
				"availabilityDomain": "AD1",
			},
			expectedStorageClassParameters: &StorageClassParameters{},
			wantErr:                        true,
			wantErrMessage:                 "Neither Mount Target Ocid nor Mount Target Subnet Ocid provided in storage class",
		},
	}
	ctx := context.Background()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			d := &FSSControllerDriver{ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: "oc1.compartment.xxxx"},
				client:     NewClientProvisioner(nil, nil, &MockFileStorageClient{}),
				util:       &csi_util.Util{},
			}}
			_, _, gotStorageClassParameters, err, _ := extractStorageClassParameters(ctx, d, d.logger, map[string]string{}, "ut-volume", tt.parameters, time.Now())
			if tt.wantErr == false && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != false && !strings.Contains(err.Error(), tt.wantErrMessage) {
				t.Errorf("want error %q to include %q", err, tt.wantErrMessage)
			}
			if tt.wantErr != true && !isStorageClassParametersEqual(gotStorageClassParameters, tt.expectedStorageClassParameters) {
				t.Errorf("extractStorageClassParameters() = %v, want %v", gotStorageClassParameters, tt.expectedStorageClassParameters)
			}
		})
	}
}

func isStorageClassParametersEqual(gotStorageClassParameters, expectedStorageClassParameters *StorageClassParameters) bool {
	return (gotStorageClassParameters.availabilityDomain == expectedStorageClassParameters.availabilityDomain) &&
		(gotStorageClassParameters.mountTargetSubnetOcid == expectedStorageClassParameters.mountTargetSubnetOcid) &&
		(gotStorageClassParameters.mountTargetOcid == expectedStorageClassParameters.mountTargetOcid) &&
		(gotStorageClassParameters.compartmentOcid == expectedStorageClassParameters.compartmentOcid) &&
		(gotStorageClassParameters.exportPath == expectedStorageClassParameters.exportPath) &&
		(gotStorageClassParameters.kmsKey == expectedStorageClassParameters.kmsKey)
}
