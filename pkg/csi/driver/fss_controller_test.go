package driver

import (
	"context"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/filestorage"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"strings"
	"testing"
	"time"
)

type MockFileStorageClient struct{}

func (c *MockFileStorageClient) GetMountTarget(ctx context.Context, id string) (*filestorage.MountTarget, error) {
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

func (c *MockFileStorageClient) GetMountTargetSummaryByDisplayName(ctx context.Context, compartmentID, ad, mountTargetName string) (bool, []filestorage.MountTargetSummary, error) {
	return true, nil, nil
}

// CreateFileSystem mocks the FileStorage CreateFileSystem implementation.
func (c *MockFileStorageClient) CreateFileSystem(ctx context.Context, details filestorage.CreateFileSystemDetails) (*filestorage.FileSystem, error) {
	idFs := "oc1.filesystem.xxxx"
	ad := "zkJl:US-ASHBURN-AD-1"
	return &filestorage.FileSystem{
		Id:                 &idFs,
		AvailabilityDomain: &ad,
	}, nil
}

// GetFileSystem mocks the FileStorage GetFileSystem implementation.
func (c *MockFileStorageClient) GetFileSystem(ctx context.Context, id string) (*filestorage.FileSystem, error) {
	idFs := "oc1.filesystem.xxxx"
	ad := "zkJl:US-ASHBURN-AD-1"
	displayName := "filesystem"
	compartmentOcid := "oc1.comp.xxxx"
	return &filestorage.FileSystem{
		Id:                 &idFs,
		AvailabilityDomain: &ad,
		DisplayName:        &displayName,
		CompartmentId:      &compartmentOcid,
	}, nil
}

func (c *MockFileStorageClient) AwaitFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.FileSystem, error) {
	idFs := "oc1.filesystem.xxxx"
	ad := "zkJl:US-ASHBURN-AD-1"
	displayName := "filesystem"
	return &filestorage.FileSystem{
		Id:                 &idFs,
		AvailabilityDomain: &ad,
		DisplayName:        &displayName,
	}, nil
}

func (c *MockFileStorageClient) GetFileSystemSummaryByDisplayName(ctx context.Context, compartmentID, ad, displayName string) (bool, []filestorage.FileSystemSummary, error) {
	idFs := "oc1.filesystem.xxxx"
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
	idFs := "oc1.filesystem.xxxx"
	return &filestorage.Export{
		Id:           &idEx,
		FileSystemId: &idFs,
	}, nil
}

// GetExport mocks the FileStorage CreateExport implementation.
func (c *MockFileStorageClient) GetExport(ctx context.Context, request filestorage.GetExportRequest) (response filestorage.GetExportResponse, err error) {
	return filestorage.GetExportResponse{}, nil
}

func (c *MockFileStorageClient) AwaitExportActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.Export, error) {
	idEx := "oc1.export.xxxx"
	idFs := "oc1.filesystem.xxxx"
	return &filestorage.Export{
		Id:           &idEx,
		FileSystemId: &idFs,
	}, nil
}

func (c *MockFileStorageClient) FindExport(ctx context.Context, fsID, path, exportSetID string) (*filestorage.ExportSummary, error) {
	idEx := "oc1.export.xxxx"
	idFs := "oc1.filesystem.xxxx"
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

func (m MockFSSProvisionerClient) ContainerEngine() client.ContainerEngineInterface {
	return &MockContainerEngineClient{}
}

func (m MockFSSProvisionerClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

func (m MockFSSProvisionerClient) LoadBalancer(s string) client.GenericLoadBalancerInterface {
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

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			d := &FSSControllerDriver{ControllerDriver{
				KubeClient: nil,
				logger:     zap.S(),
				config:     &providercfg.Config{CompartmentID: "oc1.compartment.xxxx"},
				client:     NewClientProvisioner(nil, nil, &MockFileStorageClient{}),
				util:       &csi_util.Util{},
			}}
			_, _, gotStorageClassParameters, err, _ := extractStorageClassParameters(d, d.logger, map[string]string{}, "ut-volume", tt.parameters, time.Now())
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
