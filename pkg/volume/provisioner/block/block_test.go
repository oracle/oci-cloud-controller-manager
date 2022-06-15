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

package block

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/core"
	"github.com/oracle/oci-go-sdk/v50/filestorage"
	"github.com/oracle/oci-go-sdk/v50/identity"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v8/controller"
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

func (c *MockBlockStorageClient) AwaitVolumeAvailableORTimeout(ctx context.Context, id string) (*core.Volume, error) {
	return &core.Volume{
		Id:             &id,
		LifecycleState: c.VolumeState,
	}, nil
}

func (c *MockBlockStorageClient) GetVolume(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

func (c *MockBlockStorageClient) GetVolumesByName(ctx context.Context, volumeName, compartmentID string) ([]core.Volume, error) {
	return nil, nil
}

// CreateVolume mocks the BlockStorage CreateVolume implementation
func (c *MockBlockStorageClient) CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error) {
	return &core.Volume{Id: &VolumeBackupID}, nil
}

func (c *MockBlockStorageClient) UpdateVolume(ctx context.Context, volumeId string, details core.UpdateVolumeDetails) (*core.Volume, error) {
	return &core.Volume{Id: &volumeId}, nil
}

// DeleteVolume mocks the BlockStorage DeleteVolume implementation
func (c *MockBlockStorageClient) DeleteVolume(ctx context.Context, id string) error {
	return nil
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

func (MockComputeClient) AttachParavirtualizedVolume(ctx context.Context, instanceID, volumeID string, isPvEncryptionInTransitEnabled bool) (core.VolumeAttachment, error) {
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

func (c *MockComputeClient) FindActiveVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

// MockVirtualNetworkClient mocks VirtualNetwork client implementation
type MockVirtualNetworkClient struct {
}

func (c *MockVirtualNetworkClient) IsRegionalSubnet(ctx context.Context, id string) (bool, error) {
	return false, nil
}

// GetPrivateIP mocks the VirtualNetwork GetPrivateIp implementation
func (c *MockVirtualNetworkClient) GetPrivateIP(ctx context.Context, id string) (*core.PrivateIp, error) {
	return &core.PrivateIp{IpAddress: &privateIP}, nil
}

func (c *MockVirtualNetworkClient) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetVcn(ctx context.Context, id string) (*core.Vcn, error) {
	return &core.Vcn{}, nil
}

func (c *MockVirtualNetworkClient) GetSubnetFromCacheByIP(ip string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error) {
	return core.GetSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) UpdateSecurityList(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) (core.UpdateSecurityListResponse, error) {
	return core.UpdateSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) GetPublicIpByIpAddress(ctx context.Context, id string) (*core.PublicIp, error) {
	return nil, nil
}

// MockIdentityClient mocks identity client structure
type MockIdentityClient struct {
	common.BaseClient
}

func (client MockIdentityClient) ListAvailabilityDomains(ctx context.Context, compartmentID string) ([]identity.AvailabilityDomain, error) {
	return nil, nil
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
func (p *MockProvisionerClient) LoadBalancer(string) client.GenericLoadBalancerInterface {
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
	return "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}

type MockLoadBalancerClient struct{}

func (c *MockLoadBalancerClient) CreateLoadBalancer(ctx context.Context, details *client.GenericCreateLoadBalancerDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetLoadBalancer(ctx context.Context, id string) (*client.GenericLoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) GetLoadBalancerByName(ctx context.Context, compartmentID string, name string) (*client.GenericLoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetCertificateByName(ctx context.Context, lbID string, name string) (*client.GenericCertificate, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) CreateCertificate(ctx context.Context, lbID string, cert *client.GenericCertificate) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateBackendSet(ctx context.Context, lbID string, name string, details *client.GenericBackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateBackendSet(ctx context.Context, lbID string, name string, details *client.GenericBackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteBackendSet(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateListener(ctx context.Context, lbID string, name string, details *client.GenericListener) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateListener(ctx context.Context, lbID string, name string, details *client.GenericListener) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteListener(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateLoadBalancerShape(context.Context, string, *client.GenericUpdateLoadBalancerShapeDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) AwaitWorkRequest(ctx context.Context, id string) (*client.GenericWorkRequest, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) UpdateNetworkSecurityGroups(context.Context, string, []string) (string, error) {
	return "", nil
}

// NewClientProvisioner creates an OCI client from the given configuration.
func NewClientProvisioner(pcData client.Interface, storage *MockBlockStorageClient) client.Interface {
	return &MockProvisionerClient{Storage: storage}
}

var (
	volumeBackupID = "dummyVolumeBackupId"
	defaultAD      = identity.AvailabilityDomain{Name: common.String("PHX-AD-1"), CompartmentId: common.String("ocid1.compartment.oc1")}
	serverIPs      = []string{"dummyServerIP"}
)

func TestResolveFSTypeWhenNotConfigured(t *testing.T) {
	storageClass := v12.StorageClass{
		Parameters: map[string]string{FSType: ""},
	}
	provisionerOptions := controller.ProvisionOptions{
		StorageClass: &storageClass,
	}
	// test default fsType of 'ext4' is always returned.
	fst := resolveFSType(provisionerOptions)
	if fst != "ext4" {
		t.Fatalf("Unexpected filesystem type: '%s'.", fst)
	}
}

func TestResolveFSTypeWhenConfigured(t *testing.T) {
	// test default fsType of 'ext3' is always returned when configured.
	storageClass := v12.StorageClass{
		Parameters: map[string]string{FSType: "ext3"},
	}
	provisionerOptions := controller.ProvisionOptions{
		StorageClass: &storageClass,
	}
	fst := resolveFSType(provisionerOptions)
	if fst != "ext3" {
		t.Fatalf("Unexpected filesystem type: '%s'.", fst)
	}
}

func TestCreateVolumeFromBackup(t *testing.T) {
	// test creating a volume from an existing backup

	persistentVolumeReclaimPolicy := v1.PersistentVolumeReclaimPolicy("Test")

	storageClass := v12.StorageClass{
		Parameters:    map[string]string{"volumeRoundingUpEnabled": "true"},
		ReclaimPolicy: &persistentVolumeReclaimPolicy,
	}
	options := controller.ProvisionOptions{
		StorageClass: &storageClass,
		PVName:       "dummyVolumeOptions",
		PVC: &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					OCIVolumeBackupID: volumeBackupID,
				},
			},
			Spec: v1.PersistentVolumeClaimSpec{
				StorageClassName: common.String("oci"),
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceName(v1.ResourceStorage): resource.MustParse("50Gi"),
					},
				},
			},
		}}

	block := NewBlockProvisioner(zap.S(),
		NewClientProvisioner(nil, &MockBlockStorageClient{VolumeState: core.VolumeLifecycleStateAvailable}),
		"ocid1.",
		"phx",
		true,
		resource.MustParse("50Gi"))

	provisionedVolume, err := block.Provision(options, &defaultAD)
	if err != nil {
		t.Fatalf("Failed to provision volume from block storage: %v", err)
	}
	if provisionedVolume.Annotations[OCIVolumeID] != volumeBackupID {
		t.Fatalf("Failed to assign the id of the blockID: %s, assigned %s instead", volumeBackupID,
			provisionedVolume.Annotations[OCIVolumeID])
	}
}

func TestVolumeRoundingLogic(t *testing.T) {
	var volumeRoundingTests = []struct {
		requestedStorage string
		enabled          bool
		minVolumeSize    resource.Quantity
		expected         string
	}{
		{"20Gi", true, resource.MustParse("50Gi"), "50Gi"},
		{"30Gi", true, resource.MustParse("25Gi"), "30Gi"},
		{"50Gi", true, resource.MustParse("50Gi"), "50Gi"},
		{"20Gi", false, resource.MustParse("50Gi"), "20Gi"},
	}
	for i, tt := range volumeRoundingTests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			persistentVolumeReclaimPolicy := v1.PersistentVolumeReclaimPolicy("Test")

			storageClass := v12.StorageClass{
				Parameters:    map[string]string{"volumeRoundingUpEnabled": "true"},
				ReclaimPolicy: &persistentVolumeReclaimPolicy,
			}

			volumeOptions := controller.ProvisionOptions{
				StorageClass: &storageClass,
				PVC:          createPVC(tt.requestedStorage),
			}
			block := NewBlockProvisioner(zap.S(), NewClientProvisioner(nil, &MockBlockStorageClient{VolumeState: core.VolumeLifecycleStateAvailable}),
				"ocid1.",
				"phx",
				tt.enabled,
				tt.minVolumeSize)
			provisionedVolume, err := block.Provision(volumeOptions, &defaultAD)
			if err != nil {
				t.Fatalf("Expected no error but got %s", err)
			}

			expectedCapacity := resource.MustParse(tt.expected)
			actualCapacity := provisionedVolume.Spec.Capacity[v1.ResourceName(v1.ResourceStorage)]

			actual := actualCapacity.String()
			expected := expectedCapacity.String()
			if actual != expected {
				t.Fatalf("Expected volume to be %s but got %s", expected, actual)
			}
		})
	}
}

func TestVolumeBackupOCID(t *testing.T) {
	var volumeBackupOcidTests = []struct {
		in  string
		out bool
	}{
		{"ocid1.volumebackup.", true},
		{"ocidv1:volumebackup.", true},
		{"ocid2.volumebackup.", true},
		{"ocidv2.volumebackup.", true},
		{"ocid.volumebackup.", true},
		{"ocid1.volumebackupsdfljf", false},
		{"ocidv1:volume.", false},
		{"ocid2.volume.", false},
		{"ocidv2.volume.", false},
	}
	for i, tt := range volumeBackupOcidTests {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			expected := tt.out
			actual := isVolumeBackupOcid(tt.in)
			if expected != actual {
				t.Fatalf("Expected value to be %v but got %v", expected, actual)
			}
		})
	}
}

func createPVC(size string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1.PersistentVolumeClaimSpec{
			StorageClassName: common.String("oci"),
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse(size),
				},
			},
		},
	}
}
