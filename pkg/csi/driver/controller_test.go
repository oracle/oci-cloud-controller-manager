package driver

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v31/common"
	"github.com/oracle/oci-go-sdk/v31/core"
	"github.com/oracle/oci-go-sdk/v31/filestorage"
	"github.com/oracle/oci-go-sdk/v31/identity"
	"github.com/oracle/oci-go-sdk/v31/loadbalancer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

const (
	testMinimumVolumeSizeInBytes int64 = 50 * client.GiB
)

var (
	inTransitEncryptionEnabled  = true
	inTransitEncryptionDisabled = false
	instances                   = map[string]*core.Instance{
		"inTransitEnabled": {
			LaunchOptions: &core.LaunchOptions{
				IsPvEncryptionInTransitEnabled: &inTransitEncryptionEnabled,
			},
		},
		"inTransitDisabled": {
			LaunchOptions: &core.LaunchOptions{
				IsPvEncryptionInTransitEnabled: &inTransitEncryptionDisabled,
			},
		},
	}
)

type MockOCIClient struct{}

func (MockOCIClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

func (MockOCIClient) LoadBalancer() client.LoadBalancerInterface {
	return &MockLoadBalancerClient{}
}

func (MockOCIClient) Networking() client.NetworkingInterface {
	return &MockVirtualNetworkClient{}
}

func (MockOCIClient) BlockStorage() client.BlockStorageInterface {
	return &MockBlockStorageClient{}
}

func (MockOCIClient) FSS() client.FileStorageInterface {
	return &MockFileStorageClient{}
}

func (MockOCIClient) Identity() client.IdentityInterface {
	return &MockIdentityClient{}
}

type MockBlockStorageClient struct {
}

type MockProvisionerClient struct {
	Storage *MockBlockStorageClient
}

func (c *MockBlockStorageClient) AwaitVolumeAvailableORTimeout(ctx context.Context, id string) (*core.Volume, error) {
	return &core.Volume{}, nil
}

func (c *MockBlockStorageClient) GetVolume(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

func (c *MockBlockStorageClient) GetVolumesByName(ctx context.Context, volumeName, compartmentID string) ([]core.Volume, error) {
	return []core.Volume{}, nil
}

// CreateVolume mocks the BlockStorage CreateVolume implementation
func (c *MockBlockStorageClient) CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error) {
	id := "oc1.volume1.xxxx"
	ad := "zkJl:US-ASHBURN-AD-1"
	return &core.Volume{
		Id:                 &id,
		AvailabilityDomain: &ad,
	}, nil
}

// DeleteVolume mocks the BlockStorage DeleteVolume implementation
func (c *MockBlockStorageClient) DeleteVolume(ctx context.Context, id string) error {
	return nil
}

func (c *MockBlockStorageClient) AwaitVolumeAvailable(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

// BlockStorage mocks client BlockStorage implementation
func (p *MockProvisionerClient) BlockStorage() client.BlockStorageInterface {
	return p.Storage
}

// MockVirtualNetworkClient mocks VirtualNetwork client implementation
type MockVirtualNetworkClient struct {
}

// GetPrivateIP mocks the VirtualNetwork GetPrivateIP implementation
func (c *MockVirtualNetworkClient) GetPrivateIP(ctx context.Context, id string) (*core.PrivateIp, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSubnetFromCacheByIP(ip string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetVcn(ctx context.Context, id string) (*core.Vcn, error) {
	return &core.Vcn{}, nil
}

func (c *MockVirtualNetworkClient) GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error) {
	return core.GetSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) UpdateSecurityList(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) (core.UpdateSecurityListResponse, error) {
	return core.UpdateSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) IsRegionalSubnet(ctx context.Context, id string) (bool, error) {
	return true, nil
}

// Networking mocks client VirtualNetwork implementation.
func (p *MockProvisionerClient) Networking() client.NetworkingInterface {
	return &MockVirtualNetworkClient{}
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

func (c *MockLoadBalancerClient) CreateBackend(ctx context.Context, lbID, bsName string, details loadbalancer.BackendDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteBackend(ctx context.Context, lbID, bsName, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateLoadBalancerShape(ctx context.Context, lbID string, details loadbalancer.UpdateLoadBalancerShapeDetails) (string, error) {
	return "", nil
}

// Networking mocks client VirtualNetwork implementation.
func (p *MockProvisionerClient) LoadBalancer() client.LoadBalancerInterface {
	return &MockLoadBalancerClient{}
}

type MockComputeClient struct{}

// GetInstance gets information about the specified instance.
func (c *MockComputeClient) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	if instance, ok := instances[id]; ok {
		return instance, nil
	}
	return nil, errors.New("instance not found")
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

func (c *MockComputeClient) FindActiveVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
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

func (p *MockProvisionerClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
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
	ad1 := "AD1"
	return &identity.AvailabilityDomain{Name: &ad1}, nil
}

// Identity mocks client Identity implementation
func (p *MockProvisionerClient) Identity() client.IdentityInterface {
	return &MockIdentityClient{}
}

type MockFileStorageClient struct{}

// CreateFileSystem mocks the FileStorage CreateFileSystem implementation.
func (c *MockFileStorageClient) CreateFileSystem(ctx context.Context, details filestorage.CreateFileSystemDetails) (*filestorage.FileSystem, error) {
	return nil, nil
}

// GetFileSystem mocks the FileStorage GetFileSystem implementation.
func (c *MockFileStorageClient) GetFileSystem(ctx context.Context, id string) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (c *MockFileStorageClient) AwaitFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (c *MockFileStorageClient) GetFileSystemSummaryByDisplayName(ctx context.Context, compartmentID, ad, displayName string) (*filestorage.FileSystemSummary, error) {
	return nil, nil
}

// DeleteFileSystem mocks the FileStorage DeleteFileSystem implementation
func (c *MockFileStorageClient) DeleteFileSystem(ctx context.Context, id string) error {
	return nil
}

// CreateExport mocks the FileStorage CreateExport implementation
func (c *MockFileStorageClient) CreateExport(ctx context.Context, details filestorage.CreateExportDetails) (*filestorage.Export, error) {
	return nil, nil
}

// GetExport mocks the FileStorage CreateExport implementation.
func (c *MockFileStorageClient) GetExport(ctx context.Context, request filestorage.GetExportRequest) (response filestorage.GetExportResponse, err error) {
	return filestorage.GetExportResponse{}, nil
}
func (c *MockFileStorageClient) AwaitExportActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.Export, error) {
	return nil, nil
}

func (c *MockFileStorageClient) FindExport(ctx context.Context, compartmentID, fsID, exportSetID string) (*filestorage.ExportSummary, error) {
	return nil, nil
}

// DeleteExport mocks the FileStorage DeleteExport implementation
func (c *MockFileStorageClient) DeleteExport(ctx context.Context, id string) error {
	return nil
}

// GetMountTarget mocks the FileStorage GetMountTarget implementation
func (c *MockFileStorageClient) AwaitMountTargetActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.MountTarget, error) {
	return nil, nil
}

// FSS mocks client FileStorage implementation
func (p *MockProvisionerClient) FSS() client.FileStorageInterface {
	return &MockFileStorageClient{}
}

func NewClientProvisioner(pcData client.Interface, storage *MockBlockStorageClient) client.Interface {
	return &MockProvisionerClient{Storage: storage}
}

func TestControllerDriver_CreateVolume(t *testing.T) {
	type fields struct {
		KubeClient kubernetes.Interface
		logger     *zap.SugaredLogger
		config     *providercfg.Config
		client     client.Interface
		util       *Util
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
				ctx: nil,
				req: &csi.CreateVolumeRequest{Name: ""},
			},
			want:    nil,
			wantErr: errors.New("CreateVolume Name must be provided"),
		},
		{
			name:   "Error for unsupported VolumeCapabilities: MULTI_NODE_MULTI_WRITER provided in CreateVolumeRequest",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("invalid volume capabilities requested"),
		},
		{
			name:   "Error for no VolumeCapabilities provided in CreateVolumeRequest",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name:               "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{},
				},
			},
			want:    nil,
			wantErr: errors.New("VolumeCapabilities must be provided in CreateVolumeRequest"),
		},
		{
			name:   "Error for unsupported VolumeCapabilities: MULTI_NODE_READER_ONLY provided in CreateVolumeRequest",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("invalid volume capabilities requested"),
		},
		{
			name:   "Error for unsupported VolumeCapabilities: MULTI_NODE_SINGLE_WRITER provided in CreateVolumeRequest",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
						},
					}},
				},
			},
			want:    nil,
			wantErr: errors.New("invalid volume capabilities requested"),
		},
		{
			name:   "Error for exceeding capacity range",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{&csi.VolumeCapability{
						AccessMode: &csi.VolumeCapability_AccessMode{
							Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
						},
					}},
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: int64(maximumVolumeSizeInBytes) + int64(1024),
						LimitBytes:    minimumVolumeSizeInBytes,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("invalid capacity range"),
		},
		{
			name:   "Error in finding topology requirement",
			fields: fields{},
			args: args{
				ctx: nil,
				req: &csi.CreateVolumeRequest{
					Name: "ut-volume",
					VolumeCapabilities: []*csi.VolumeCapability{
						&csi.VolumeCapability{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
							},
						}},
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: int64(maximumVolumeSizeInBytes),
					},
					AccessibilityRequirements: &csi.TopologyRequirement{Requisite: []*csi.Topology{
						{
							Segments: map[string]string{"x": "ad1"},
						},
					},
					},
				},
			},
			want:    nil,
			wantErr: errors.New("required in PreferredTopologies or allowedTopologies"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: ""},
				client:     NewClientProvisioner(nil, &MockBlockStorageClient{}),
				util:       &Util{},
			}
			got, err := d.CreateVolume(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.CreateVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestControllerDriver_DeleteVolume(t *testing.T) {
	type fields struct {
		KubeClient kubernetes.Interface
		logger     *zap.SugaredLogger
		config     *providercfg.Config
		client     client.Interface
		util       *Util
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
			wantErr: errors.New("DeleteVolume Volume ID must be provided"),
		},
		{
			name:   "Delete volume and get empty response",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				req: &csi.DeleteVolumeRequest{VolumeId: "oc1.volume1.xxxx"},
			},
			want:    &csi.DeleteVolumeResponse{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: ""},
				client:     NewClientProvisioner(nil, &MockBlockStorageClient{}),
				util:       &Util{},
			}
			got, err := d.DeleteVolume(tt.args.ctx, tt.args.req)
			if tt.wantErr == nil && err != nil {
				t.Errorf("got error %q, want none", err)
			}
			if tt.wantErr != nil && !strings.Contains(err.Error(), tt.wantErr.Error()) {
				t.Errorf("want error %q to include %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ControllerDriver.CreateVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractStorage(t *testing.T) {
	type args struct {
		capRange *csi.CapacityRange
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "Nil CapacityRange",
			args:    args{capRange: nil},
			want:    testMinimumVolumeSizeInBytes,
			wantErr: false,
		},
		{
			name:    "Empty CapacityRange",
			args:    args{capRange: &csi.CapacityRange{}},
			want:    testMinimumVolumeSizeInBytes,
			wantErr: false,
		},
		{
			name: "Limit bytes is less than required",
			args: args{capRange: &csi.CapacityRange{
				RequiredBytes: 100 * client.GiB,
				LimitBytes:    50 * client.GiB,
			},
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Required set and limit not set",
			args: args{capRange: &csi.CapacityRange{
				RequiredBytes: 100 * client.GiB,
			},
			},
			want:    100 * client.GiB,
			wantErr: false,
		},
		{
			name: "Required set and limit set",
			args: args{capRange: &csi.CapacityRange{
				RequiredBytes: 70 * client.GiB,
				LimitBytes:    100 * client.GiB,
			},
			},
			want:    100 * client.GiB,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractStorage(tt.args.capRange)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractVolumeParameters(t *testing.T) {
	tests := map[string]struct {
		storageParameters map[string]string
		volumeParameters  VolumeParameters
		wantErr           bool
	}{
		"Wrong Attachment Type": {
			storageParameters: map[string]string{
				attachmentType: "foo",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
			},
			wantErr: true,
		},
		"StorageClass Parameters are empty": {
			storageParameters: map[string]string{},
			volumeParameters: VolumeParameters{
				diskEncryptionKey:   "",
				attachmentParameter: make(map[string]string),
			},
			wantErr: false,
		},
		"StorageClass with CMEK and attachment type paravirtualized": {
			storageParameters: map[string]string{
				attachmentType: attachmentTypeParavirtualized,
				kmsKey:         "foo",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey: "foo",
				attachmentParameter: map[string]string{
					attachmentType: attachmentTypeParavirtualized,
				},
			},
			wantErr: false,
		},
		"StorageClass with CMEK and attachment type iscsi": {
			storageParameters: map[string]string{
				attachmentType: attachmentTypeISCSI,
				kmsKey:         "bar",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey: "bar",
				attachmentParameter: map[string]string{
					attachmentType: attachmentTypeISCSI,
				},
			},
			wantErr: false,
		},
		"StorageClass with CMEK and attachment type IScsi(string casing is different)": {
			storageParameters: map[string]string{
				attachmentType: "IScsi",
				kmsKey:         "bar",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey: "bar",
				attachmentParameter: map[string]string{
					attachmentType: attachmentTypeISCSI,
				},
			},
			wantErr: false,
		},
		"StorageClass with CMEK and attachment type ParaVirtualized(string casing is different)": {
			storageParameters: map[string]string{
				attachmentType: "ParaVirtualized",
				kmsKey:         "foo",
			},
			volumeParameters: VolumeParameters{
				diskEncryptionKey: "foo",
				attachmentParameter: map[string]string{
					attachmentType: attachmentTypeParavirtualized,
				},
			},
			wantErr: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			volumeParameters, err := extractVolumeParameters(tt.storageParameters)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractVolumeParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(volumeParameters, tt.volumeParameters) {
				t.Errorf("extractStorage() = %v, want %v", volumeParameters, tt.volumeParameters)
			}
		})
	}
}

func TestGetAttachmentOptions(t *testing.T) {
	tests := map[string]struct {
		attachmentType         string
		instanceID             string
		volumeAttachmentOption VolumeAttachmentOption
		wantErr                bool
	}{
		"PV attachment with instance in-transit encryption enabled": {
			attachmentType: attachmentTypeParavirtualized,
			instanceID:     "inTransitEnabled",
			volumeAttachmentOption: VolumeAttachmentOption{
				enableInTransitEncryption:    true,
				useParavirtualizedAttachment: true,
			},
			wantErr: false,
		},

		"PV attachment with instance in-transit encryption disabled": {
			attachmentType: attachmentTypeParavirtualized,
			instanceID:     "inTransitDisabled",
			volumeAttachmentOption: VolumeAttachmentOption{
				enableInTransitEncryption:    false,
				useParavirtualizedAttachment: true,
			},
			wantErr: false,
		},
		"ISCSI attachment with instance in-transit encryption enabled": {
			attachmentType: attachmentTypeISCSI,
			instanceID:     "inTransitEnabled",
			volumeAttachmentOption: VolumeAttachmentOption{
				enableInTransitEncryption:    true,
				useParavirtualizedAttachment: true,
			},
			wantErr: false,
		},
		"ISCSI attachment with instance in-transit encryption disabled": {
			attachmentType: attachmentTypeISCSI,
			instanceID:     "inTransitDisabled",
			volumeAttachmentOption: VolumeAttachmentOption{
				enableInTransitEncryption:    false,
				useParavirtualizedAttachment: false,
			},
			wantErr: false,
		},
		"API error": {
			attachmentType:         attachmentTypeISCSI,
			instanceID:             "foo",
			volumeAttachmentOption: VolumeAttachmentOption{},
			wantErr:                true,
		},
	}

	computeClient := MockOCIClient{}.Compute()

	for name, tt := range tests {

		t.Run(name, func(t *testing.T) {
			volumeAttachmentOption, err := getAttachmentOptions(context.Background(), computeClient, tt.attachmentType, tt.instanceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAttachmentOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(volumeAttachmentOption, tt.volumeAttachmentOption) {
				t.Errorf("getAttachmentOptions() = %v, want %v", volumeAttachmentOption, tt.volumeAttachmentOption)
			}
		})
	}
}
