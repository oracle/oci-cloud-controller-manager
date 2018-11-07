// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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

package fss

import (
	"context"
	"testing"
	"time"

	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/filestorage"
	"github.com/oracle/oci-go-sdk/identity"
	"github.com/oracle/oci-go-sdk/loadbalancer"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	// VolumeBackupID of backup volume
	VolumeBackupID = "dummyVolumeBackupId"
	fileSystemID   = "dummyFileSystemId"
	exportID       = "dummyExportID"
	exportSetID    = "dummyExportSetID"
	// NilListMountTargetsADID lists no mount targets for the given AD
	NilListMountTargetsADID = "dummyNilListMountTargetsForADID"
	mountTargetID           = "dummyMountTargetID"
	// CreatedMountTargetID for dynamically created mount target
	CreatedMountTargetID = "dummyCreatedMountTargetID"
	// ServerIPs address for mount target
	ServerIPs = []string{"dummyServerIP"}
	privateIP = "127.0.0.1"
)

// MockBlockStorageClient mocks BlockStorage client implementation
type MockBlockStorageClient struct {
	VolumeState core.VolumeLifecycleStateEnum
}

// CreateVolume mocks the BlockStorage CreateVolume implementation
func (c *MockBlockStorageClient) CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error) {
	return &core.Volume{Id: &VolumeBackupID}, nil
}

// DeleteVolume mocks the BlockStorage DeleteVolume implementation
func (c *MockBlockStorageClient) DeleteVolume(ctx context.Context, id string) error {
	return nil
}

func (c *MockBlockStorageClient) AwaitVolumeAvailable(ctx context.Context, id string) (*core.Volume, error) {
	return &core.Volume{
		Id:             &id,
		LifecycleState: c.VolumeState,
	}, nil
}

// MockFileStorageClient mocks FileStorage client implementation.
type MockFileStorageClient struct{}

// CreateFileSystem mocks the FileStorage CreateFileSystem implementation.
func (c *MockFileStorageClient) CreateFileSystem(ctx context.Context, details filestorage.CreateFileSystemDetails) (*filestorage.FileSystem, error) {
	return &filestorage.FileSystem{Id: &fileSystemID}, nil
}

// GetFileSystem mocks the FileStorage GetFileSystem implementation.
func (c *MockFileStorageClient) GetFileSystem(ctx context.Context, id string) (*filestorage.FileSystem, error) {
	return &filestorage.FileSystem{
		Id:             &id,
		LifecycleState: filestorage.FileSystemLifecycleStateActive,
	}, nil
}

func (c *MockFileStorageClient) AwaitFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.FileSystem, error) {
	return &filestorage.FileSystem{
		Id:             &id,
		LifecycleState: filestorage.FileSystemLifecycleStateActive,
	}, nil
}

func (c *MockFileStorageClient) GetFileSystemSummaryByDisplayName(ctx context.Context, compartmentID, ad, displayName string) (*filestorage.FileSystemSummary, error) {
	return &filestorage.FileSystemSummary{
		Id:             &fileSystemID,
		DisplayName:    &displayName,
		LifecycleState: filestorage.FileSystemSummaryLifecycleStateActive,
	}, nil
}

// DeleteFileSystem mocks the FileStorage DeleteFileSystem implementation
func (c *MockFileStorageClient) DeleteFileSystem(ctx context.Context, id string) error {
	return nil
}

// CreateExport mocks the FileStorage CreateExport implementation
func (c *MockFileStorageClient) CreateExport(ctx context.Context, details filestorage.CreateExportDetails) (*filestorage.Export, error) {
	return &filestorage.Export{Id: &exportID}, nil
}

// GetExport mocks the FileStorage CreateExport implementation.
func (c *MockFileStorageClient) GetExport(ctx context.Context, request filestorage.GetExportRequest) (response filestorage.GetExportResponse, err error) {
	return filestorage.GetExportResponse{
		Export: filestorage.Export{
			Id:             common.String(exportID),
			FileSystemId:   &fileSystemID,
			ExportSetId:    &exportSetID,
			LifecycleState: filestorage.ExportLifecycleStateActive,
			Path:           common.String("/" + fileSystemID),
		},
	}, nil
}
func (c *MockFileStorageClient) AwaitExportActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.Export, error) {
	return &filestorage.Export{
		Id:             common.String(exportID),
		FileSystemId:   &fileSystemID,
		ExportSetId:    &exportSetID,
		LifecycleState: filestorage.ExportLifecycleStateActive,
		Path:           common.String("/" + fileSystemID),
	}, nil
}

func (c *MockFileStorageClient) FindExport(ctx context.Context, compartmentID, fsID, exportSetID string) (*filestorage.ExportSummary, error) {
	return &filestorage.ExportSummary{
		Id:             &exportID,
		ExportSetId:    &exportSetID,
		FileSystemId:   &fsID,
		LifecycleState: filestorage.ExportSummaryLifecycleStateActive,
	}, nil
}

// DeleteExport mocks the FileStorage DeleteExport implementation
func (c *MockFileStorageClient) DeleteExport(ctx context.Context, id string) error {
	return nil
}

// GetMountTarget mocks the FileStorage GetMountTarget implementation
func (c *MockFileStorageClient) AwaitMountTargetActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.MountTarget, error) {
	return &filestorage.MountTarget{
		PrivateIpIds:   ServerIPs,
		Id:             &CreatedMountTargetID,
		LifecycleState: filestorage.MountTargetLifecycleStateActive,
		ExportSetId:    &exportSetID,
	}, nil
}

type MockComputeClient struct{}

// GetInstance gets information about the specified instance.
func (c *MockComputeClient) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	return nil, nil
}

// GetInstanceByNodeName gets the OCI instance corresponding to the given
// Kubernetes node name.
func (c *MockComputeClient) GetInstanceByNodeName(ctx context.Context, compartmentID, vcnID, nodeName string) (*core.Instance, error) {
	return nil, nil
}

func (c *MockComputeClient) GetPrimaryVNICForInstance(ctx context.Context, compartmentID, instanceID string) (*core.Vnic, error) {
	return nil, nil
}

func (c *MockComputeClient) FindVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) AttachVolume(ctx context.Context, instanceID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) WaitForVolumeAttached(ctx context.Context, attachmentID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (c *MockComputeClient) DetachVolume(ctx context.Context, id string) error { return nil }

func (c *MockComputeClient) WaitForVolumeDetached(ctx context.Context, attachmentID string) error {
	return nil
}

// MockVirtualNetworkClient mocks VirtualNetwork client implementation
type MockVirtualNetworkClient struct {
}

// GetPrivateIP mocks the VirtualNetwork GetPrivateIP implementation
func (c *MockVirtualNetworkClient) GetPrivateIP(ctx context.Context, id string) (*core.PrivateIp, error) {
	return &core.PrivateIp{IpAddress: &privateIP}, nil
}

func (c *MockVirtualNetworkClient) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSubnetFromCacheByIP(ip string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error) {
	return core.GetSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) UpdateSecurityList(ctx context.Context, request core.UpdateSecurityListRequest) (core.UpdateSecurityListResponse, error) {
	return core.UpdateSecurityListResponse{}, nil
}

// MockIdentityClient mocks identity client structure
type MockIdentityClient struct {
	common.BaseClient
}

// ListAvailabilityDomains mocks the client ListAvailabilityDomains implementation
func (client MockIdentityClient) GetAvailabilityDomainByName(ctx context.Context, compartmentID, name string) (*identity.AvailabilityDomain, error) {
	return nil, nil
}

// MockProvisionerClient mocks client structure
type MockProvisionerClient struct {
	Storage *MockBlockStorageClient
}

// BlockStorage mocks client BlockStorage implementation
func (p *MockProvisionerClient) BlockStorage() client.BlockStorageInterface {
	return p.Storage
}

// Networking mocks client VirtualNetwork implementation.
func (p *MockProvisionerClient) Networking() client.NetworkingInterface {
	return &MockVirtualNetworkClient{}
}

// Networking mocks client VirtualNetwork implementation.
func (p *MockProvisionerClient) LoadBalancer() client.LoadBalancerInterface {
	return &MockLoadBalancerClient{}
}

func (p *MockProvisionerClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

// Identity mocks client Identity implementation
func (p *MockProvisionerClient) Identity() client.IdentityInterface {
	return &MockIdentityClient{}
}

// FSS mocks client FileStorage implementation
func (p *MockProvisionerClient) FSS() client.FileStorageInterface {
	return &MockFileStorageClient{}
}

// Context mocks client Context implementation
func (p *MockProvisionerClient) Context() context.Context {
	return context.Background()
}

// Timeout mocks client Timeout implementation
func (p *MockProvisionerClient) Timeout() time.Duration {
	return 30 * time.Second
}

// TenancyOCID mocks client TenancyOCID implementation
func (p *MockProvisionerClient) TenancyOCID() string {
	return "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq"
}

type MockLoadBalancerClient struct{}

func (c *MockLoadBalancerClient) CreateLoadBalancer(ctx context.Context, details loadbalancer.CreateLoadBalancerDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetLoadBalancer(ctx context.Context, id string) (*loadbalancer.LoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) GetLoadBalancerByName(ctx context.Context, compartmentID, name string) (*loadbalancer.LoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetCertificateByName(ctx context.Context, lbID, name string) (*loadbalancer.Certificate, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) CreateCertificate(ctx context.Context, lbID string, cert loadbalancer.CertificateDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateBackendSet(ctx context.Context, lbID, name string, details loadbalancer.BackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateBackendSet(ctx context.Context, lbID, name string, details loadbalancer.BackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteBackendSet(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateListener(ctx context.Context, lbID, name string, details loadbalancer.ListenerDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateListener(ctx context.Context, lbID, name string, details loadbalancer.ListenerDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteListener(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) AwaitWorkRequest(ctx context.Context, id string) (*loadbalancer.WorkRequest, error) {
	return nil, nil
}

// NewClientProvisioner creates an OCI client from the given configuration.
func NewClientProvisioner(pcData client.Interface, storage *MockBlockStorageClient) client.Interface {
	return &MockProvisionerClient{Storage: storage}
}

func TestCreateVolumeWithFSS(t *testing.T) {
	fsp := filesystemProvisioner{
		client: NewClientProvisioner(nil, nil),
		logger: zaptest.NewLogger(t).Sugar(),
		region: "phx",
	}
	_, err := fsp.Provision(
		controller.VolumeOptions{
			PVName: "dummyVolumeOptions",
			PVC: &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					UID: "my-uid",
				},
			},
			Parameters: map[string]string{MntTargetID: "dummyMountTargetID"},
		},
		&identity.AvailabilityDomain{
			Name:          common.String("dummyAdName"),
			CompartmentId: common.String("dummyCompartmentId"),
		},
	)
	if err != nil {
		t.Fatalf("Failed to provision volume from fss storage: %v", err)
	}
}
