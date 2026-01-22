package driver

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	ociidentity "github.com/oracle/oci-go-sdk/v65/identity"
	lustre "github.com/oracle/oci-go-sdk/v65/lustrefilestorage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/pointer"
)

// mockNotFoundError implements ServiceError for testing oci error cases
type mockNotFoundError struct {
	statusCode   int
	message      string
	code         string
	opcRequestID string
}

func (e mockNotFoundError) GetHTTPStatusCode() int  { return e.statusCode }
func (e mockNotFoundError) GetMessage() string      { return e.message }
func (e mockNotFoundError) GetCode() string         { return e.code }
func (e mockNotFoundError) GetOpcRequestID() string { return e.opcRequestID }
func (e mockNotFoundError) Error() string           { return e.message }

// ########################## MockOCILustreFileStorageClient client  ##########################

// MockOCILustreFileStorageClient implements client.LustreInterface and allows configuring responses.
type MockOCILustreFileStorageClient struct {
	// create
	CreateResp   *lustre.LustreFileSystem
	CreateErr    error
	CreateCalled bool
	CreateArgs   *lustre.CreateLustreFileSystemDetails

	// list
	ListResp []lustre.LustreFileSystemSummary
	ListErr  error

	// get
	GetResp *lustre.LustreFileSystem
	GetErr  error

	// await active
	AwaitActiveResp *lustre.LustreFileSystem
	AwaitActiveErr  error

	// delete
	DeleteErr error

	// await deleted
	AwaitDeletedErr error

	// work requests
	WRListResp []lustre.WorkRequestSummary
	WRListErr  error
	WRErrsResp []lustre.WorkRequestError
	WRErrsErr  error
}

func (f *MockOCILustreFileStorageClient) CreateLustreFileSystem(ctx context.Context, details lustre.CreateLustreFileSystemDetails) (*lustre.LustreFileSystem, error) {
	f.CreateCalled = true
	d := details // capture
	f.CreateArgs = &d
	return f.CreateResp, f.CreateErr
}
func (f *MockOCILustreFileStorageClient) GetLustreFileSystem(ctx context.Context, id string) (*lustre.LustreFileSystem, error) {
	return f.GetResp, f.GetErr
}
func (f *MockOCILustreFileStorageClient) ListLustreFileSystems(ctx context.Context, compartmentID, ad, displayName string) ([]lustre.LustreFileSystemSummary, error) {
	return f.ListResp, f.ListErr
}
func (f *MockOCILustreFileStorageClient) DeleteLustreFileSystem(ctx context.Context, id string) error {
	return f.DeleteErr
}
func (f *MockOCILustreFileStorageClient) AwaitLustreFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*lustre.LustreFileSystem, error) {
	return f.AwaitActiveResp, f.AwaitActiveErr
}
func (f *MockOCILustreFileStorageClient) AwaitLustreFileSystemDeleted(ctx context.Context, logger *zap.SugaredLogger, id string) error {
	return f.AwaitDeletedErr
}
func (f *MockOCILustreFileStorageClient) ListWorkRequests(ctx context.Context, compartmentID, resourceID string) ([]lustre.WorkRequestSummary, error) {
	return f.WRListResp, f.WRListErr
}
func (f *MockOCILustreFileStorageClient) ListWorkRequestErrors(ctx context.Context, workRequestID string, volumeID string) ([]lustre.WorkRequestError, error) {
	return f.WRErrsResp, f.WRErrsErr
}

// ########################## MockOCIIdentityClient    ##########################

type MockOCIIdentityClient struct {
	ads     []string
	getErr  error
	listErr error
}

func (i *MockOCIIdentityClient) GetAvailabilityDomainByName(ctx context.Context, compartmentID, name string) (*ociidentity.AvailabilityDomain, error) { // interface{} to avoid ctx import
	if i.getErr != nil {
		return nil, i.getErr
	}
	if len(i.ads) > 0 {
		return &ociidentity.AvailabilityDomain{Name: &i.ads[0]}, nil
	}
	return &ociidentity.AvailabilityDomain{Name: &name}, nil
}

// ########################## CreateVolume Tests   ##########################

func TestCreateVolume_Missing_VolumeName(t *testing.T) {
	d := newControllerWith(nil, nil)
	req := &csi.CreateVolumeRequest{
		Name:               "",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125"},
	}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
	if !containsErr(err, "Name must be provided in CreateVolumeRequest") {
		t.Fatalf("expected error to mention 'Name must be provided in CreateVolumeRequest', got %v", err)
	}
}

func TestCreateVolume_Missing_Capabilities(t *testing.T) {
	d := newControllerWith(nil, nil)
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol",
		VolumeCapabilities: []*csi.VolumeCapability{},
		Parameters:         map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125"},
	}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
	if !containsErr(err, "VolumeCapabilities must be provided in CreateVolumeRequest") {
		t.Fatalf("expected error to mention 'VolumeCapabilities must be provided in CreateVolumeRequest', got %v", err)
	}
}

func TestCreateVolume_Invalid_Capabilities(t *testing.T) {
	d := newControllerWith(nil, nil)
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Block{Block: &csi.VolumeCapability_BlockVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125"},
	}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
	if !containsErr(err, "Requested Volume Capability not supported") {
		t.Fatalf("expected error to mention 'Requested Volume Capability not supported', got %v", err)
	}
}

func TestCreateVolume_Missing_AvailabilityDomain(t *testing.T) {
	// Preferred topology present; Identity GetAvailabilityDomainByName returns ctx deadline
	fid := &MockOCIIdentityClient{getErr: context.DeadlineExceeded}
	fl := &MockOCILustreFileStorageClient{ListResp: []lustre.LustreFileSystemSummary{}} // so it goes to create path
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125"},
	}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for  Missing required parameter: availabilityDomain, got %v", err)
	}
	if !containsErr(err, "Missing required parameter: availabilityDomain") {
		t.Fatalf("expected error to mention 'Missing required parameter: availabilityDomain', got %v", err)
	}
}

func TestCreateVolume_ADResolution_PreferredTimeout(t *testing.T) {
	// Preferred topology present; Identity GetAvailabilityDomainByName returns ctx deadline
	fid := &MockOCIIdentityClient{getErr: context.DeadlineExceeded}
	fl := &MockOCILustreFileStorageClient{ListResp: []lustre.LustreFileSystemSummary{}} // so it goes to create path
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"},
	}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for preferred AD resolution timeout, got %v", err)
	}
	if !containsErr(err, "Invalid availabilityDomain") {
		t.Fatalf("expected error to mention Invalid availabilityDomain, got %v", err)
	}
}

func TestCreateVolume_CreateDetails_ContainsKmsAndNsgs(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	fl := &MockOCILustreFileStorageClient{ListResp: []lustre.LustreFileSystemSummary{}, CreateResp: &lustre.LustreFileSystem{Id: ptrString("ocid1.lustrefilesystem.oc1.phx.new")}, AwaitActiveResp: func() *lustre.LustreFileSystem {
		cap := 50
		id := "ocid1.lustrefilesystem.oc1.phx.new"
		return &lustre.LustreFileSystem{Id: &id, CapacityInGBs: &cap}
	}()}
	d := newControllerWith(fl, fid)
	nsgs := []string{"ocid1.nsg.oc1..a", "ocid1.nsg.oc1..b"}
	params := map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2", "kmsKeyId": "ocid1.key.oc1..kms", "nsgIds": "[\"ocid1.nsg.oc1..a\",\"ocid1.nsg.oc1..b\"]"}
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol",
		CapacityRange:      &csi.CapacityRange{RequiredBytes: 31200 * GB},
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         params,
	}
	_, _ = d.CreateVolume(context.Background(), req)
	if fl.CreateArgs == nil {
		t.Fatalf("expected create to be called and args captured")
	}
	if fl.CreateArgs.KmsKeyId == nil || *fl.CreateArgs.KmsKeyId != "ocid1.key.oc1..kms" {
		t.Fatalf("expected kmsKeyId to be set in create details")
	}
	if len(fl.CreateArgs.NsgIds) != len(nsgs) {
		t.Fatalf("expected nsgIds length %d got %d", len(nsgs), len(fl.CreateArgs.NsgIds))
	}
}

func TestCreateVolume_IdentityClientNil(t *testing.T) {
	// newControllerWith always injects a client; to simulate identity nil, pass nil identity and expect Internal from CreateVolume
	fl := &MockOCILustreFileStorageClient{}
	// Build driver with nil identity via a small inlined client wrapper
	d := func() *LustreControllerDriver {
		inner := &testOCIClient{lustre: fl, id: nil}
		return newControllerWith(inner.Lustre(), inner.Identity(nil))
	}()
	req := &csi.CreateVolumeRequest{Name: "test-vol", VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}}, Parameters: map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"}}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal for nil identity client, got %v", err)
	}
}

func TestCreateVolume_LustreClientNil(t *testing.T) {
	// Simulate lustre client nil
	inner := &testOCIClient{lustre: nil, id: &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}}
	d := newControllerWith(inner.Lustre(), inner.Identity(nil))
	req := &csi.CreateVolumeRequest{Name: "test-vol", VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}}, Parameters: map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"}}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal for nil lustre client, got %v", err)
	}
}

func TestCreateVolume_ListError_Internal(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	fl := &MockOCILustreFileStorageClient{ListErr: status.Error(codes.Internal, "internal error")}
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{Name: "test-vol", VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}}, Parameters: map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"}}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal for list error, got %v", err)
	}
}

func TestCreateVolume_Duplicate_AlreadyExists(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	id1 := "fs1"
	id2 := "fs2"
	fl := &MockOCILustreFileStorageClient{ListResp: []lustre.LustreFileSystemSummary{{Id: &id1}, {Id: &id2}}}
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{Name: "dup-vol", VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}}, Parameters: map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"}}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("expected AlreadyExists for duplicates, got %v", err)
	}
}

func TestCreateVolume_ExistingCreating_AwaitSuccess_Idempotent(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	fsID := ptrString("ocid1.lustrefilesystem.oc1.phx.id")
	ms := ptrString("10.0.0.10")
	lnet := ptrString("tcp")
	fsname := ptrString("fs1")
	cap := 50
	fl := &MockOCILustreFileStorageClient{ListResp: []lustre.LustreFileSystemSummary{{Id: fsID}}, GetResp: &lustre.LustreFileSystem{Id: fsID, LifecycleState: lustre.LustreFileSystemLifecycleStateCreating}, AwaitActiveResp: &lustre.LustreFileSystem{Id: fsID, ManagementServiceAddress: ms, Lnet: lnet, FileSystemName: fsname, CapacityInGBs: &cap}}
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{Name: "test-vol", VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}}, Parameters: map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"}}
	_, err := d.CreateVolume(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error idempotent CREATING→success: %v", err)
	}
}

func TestCreateVolume_CreateCallError_Internal(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	fl := &MockOCILustreFileStorageClient{ListResp: []lustre.LustreFileSystemSummary{}, CreateErr: status.Error(codes.Internal, "boom")}
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{Name: "test-vol", CapacityRange: &csi.CapacityRange{RequiredBytes: 1 << 30}, VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}}, Parameters: map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"}}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal for create error, got %v", err)
	}
}

func TestCreateVolume_VolumeContext_Present(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	cap := 50
	id := "ocid1.lustrefilesystem.oc1.phx.new"
	fl := &MockOCILustreFileStorageClient{
		ListResp:        []lustre.LustreFileSystemSummary{},
		CreateResp:      &lustre.LustreFileSystem{Id: &id},
		AwaitActiveResp: &lustre.LustreFileSystem{Id: &id, ManagementServiceAddress: ptrString("10.0.0.10"), Lnet: ptrString("tcp"), FileSystemName: ptrString("fs1"), CapacityInGBs: &cap},
	}
	d := newControllerWith(fl, fid)
	params := map[string]string{
		"subnetId":                  "ocid1.subnet.oc1..x",
		"performanceTier":           "MBPS_PER_TB_125",
		"setupLnet":                 "true",
		"lustreSubnetCidr":          "10.0.0.0/24",
		"lustrePostMountParameters": "[{}]",
		"availabilityDomain":        "PHX-AD-2",
	}
	req := &csi.CreateVolumeRequest{
		Name:               "vol-vc-present",
		CapacityRange:      &csi.CapacityRange{RequiredBytes: int64(cap) << 30},
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         params,
	}
	resp, err := d.CreateVolume(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vc := resp.GetVolume().GetVolumeContext()
	if vc["setupLnet"] != "true" || vc["lustreSubnetCidr"] != "10.0.0.0/24" || vc["lustrePostMountParameters"] == "" {
		t.Fatalf("expected volumeContext keys set, got %v", vc)
	}
}

func TestCreateVolume_VolumeContext_Absent(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	capacityInGbs := 31200
	id := "ocid1.lustrefilesystem.oc1.phx.new2"
	fl := &MockOCILustreFileStorageClient{
		ListResp:        []lustre.LustreFileSystemSummary{},
		CreateResp:      &lustre.LustreFileSystem{Id: &id},
		AwaitActiveResp: &lustre.LustreFileSystem{Id: &id, ManagementServiceAddress: ptrString("10.0.0.11"), Lnet: ptrString("tcp"), FileSystemName: ptrString("fs2"), CapacityInGBs: &capacityInGbs, LifecycleState: lustre.LustreFileSystemLifecycleStateActive},
	}
	d := newControllerWith(fl, fid)
	params := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..x",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
	}
	req := &csi.CreateVolumeRequest{
		Name:               "vol-vc-absent",
		CapacityRange:      &csi.CapacityRange{RequiredBytes: int64(capacityInGbs * 1024)},
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         params,
	}
	resp, err := d.CreateVolume(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	vc := resp.GetVolume().GetVolumeContext()
	if _, ok := vc["setupLnet"]; ok {
		t.Fatalf("did not expect setupLnet in volumeContext: %v", vc)
	}
	if _, ok := vc["lustreSubnetCidr"]; ok {
		t.Fatalf("did not expect lustreSubnetCidr in volumeContext: %v", vc)
	}
	if _, ok := vc["lustrePostMountParameters"]; ok {
		t.Fatalf("did not expect lustrePostMountParameters in volumeContext: %v", vc)
	}
}

func TestCreateVolume_NewCreate_Success(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	cap := 100
	id := "ocid1.lustrefilesystem.oc1.phx.new3"
	fl := &MockOCILustreFileStorageClient{
		ListResp:        []lustre.LustreFileSystemSummary{},
		CreateResp:      &lustre.LustreFileSystem{Id: &id},
		AwaitActiveResp: &lustre.LustreFileSystem{Id: &id, ManagementServiceAddress: ptrString("10.0.0.12"), Lnet: ptrString("tcp"), FileSystemName: ptrString("fs3"), CapacityInGBs: &cap},
	}
	d := newControllerWith(fl, fid)
	params := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..x",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
	}
	req := &csi.CreateVolumeRequest{
		Name:               "vol-new-create-success",
		CapacityRange:      &csi.CapacityRange{RequiredBytes: int64(cap) << 30},
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         params,
	}
	if _, err := d.CreateVolume(context.Background(), req); err != nil {
		t.Fatalf("unexpected error for full new-create success: %v", err)
	}
}

func TestCreateVolume_GetExistingError(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	fsID := ptrString("ocid1.lustrefilesystem.oc1.phx.id")
	fl := &MockOCILustreFileStorageClient{
		ListResp: []lustre.LustreFileSystemSummary{{Id: fsID}},
		GetErr:   status.Error(codes.Internal, "error while fetching existingLustreFs"),
	}
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol-get-error",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"},
	}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", err)
	}
	if !containsErr(err, "error while fetching existingLustreFs") {
		t.Fatalf("expected error to mention fetch error, got %v", err)
	}
}

func TestCreateVolume_ExistingActive_Success(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	fsID := ptrString("ocid1.lustrefilesystem.oc1.phx.id")
	cap := 50
	fl := &MockOCILustreFileStorageClient{
		ListResp: []lustre.LustreFileSystemSummary{{Id: fsID}},
		GetResp:  &lustre.LustreFileSystem{Id: fsID, LifecycleState: lustre.LustreFileSystemLifecycleStateActive, ManagementServiceAddress: ptrString("10.0.0.10"), Lnet: ptrString("tcp"), FileSystemName: ptrString("fs1"), CapacityInGBs: &cap},
	}
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol-active",
		CapacityRange:      &csi.CapacityRange{RequiredBytes: 31200 * GB},
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"},
	}
	resp, err := d.CreateVolume(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetVolume().GetVolumeId() == "" {
		t.Fatalf("expected volume id in response")
	}
}

func TestCreateVolume_ExistingCreating_AwaitDeadline(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	fsID := ptrString("ocid1.lustrefilesystem.oc1.phx.id")
	fl := &MockOCILustreFileStorageClient{
		ListResp:       []lustre.LustreFileSystemSummary{{Id: fsID}},
		GetResp:        &lustre.LustreFileSystem{Id: fsID, LifecycleState: lustre.LustreFileSystemLifecycleStateCreating},
		AwaitActiveErr: context.DeadlineExceeded,
	}
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol-creating-deadline",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"},
	}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestCreateVolume_ExistingFailed_WorkRequestError(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	fsID := ptrString("ocid1.lustrefilesystem.oc1.phx.id")
	wrID := ptrString("wr-123")
	fl := &MockOCILustreFileStorageClient{
		ListResp:   []lustre.LustreFileSystemSummary{{Id: fsID}},
		GetResp:    &lustre.LustreFileSystem{Id: fsID, LifecycleState: lustre.LustreFileSystemLifecycleStateFailed},
		WRListResp: []lustre.WorkRequestSummary{{Id: wrID, OperationType: lustre.OperationTypeCreateLustreFileSystem}},
		WRErrsResp: []lustre.WorkRequestError{{Code: pointer.String("CREATE_FAILED"), Message: pointer.String("work request error")}},
	}
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol-failed-wr",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"},
	}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.Aborted {
		t.Fatalf("expected Aborted, got %v", err)
	}
	if !containsErr(err, "work request error") {
		t.Fatalf("expected error to mention work request error, got %v", err)
	}
}

func TestCreateVolume_NewCreate_AwaitError(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	cap := 50
	id := "ocid1.lustrefilesystem.oc1.phx.new-await-err"
	fl := &MockOCILustreFileStorageClient{
		ListResp:       []lustre.LustreFileSystemSummary{},
		CreateResp:     &lustre.LustreFileSystem{Id: &id},
		AwaitActiveErr: status.Error(codes.Internal, "await error"),
	}
	d := newControllerWith(fl, fid)
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol-new-await-err",
		CapacityRange:      &csi.CapacityRange{RequiredBytes: int64(cap) << 30},
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         map[string]string{"subnetId": "ocid1.subnet.oc1..x", "performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2"},
	}
	_, err := d.CreateVolume(context.Background(), req)
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if !containsErr(err, "await error") {
		t.Fatalf("expected error to mention await error, got %v", err)
	}
}

func TestCreateVolume_RootSquashEnabled(t *testing.T) {
	fid := &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-1"}}
	cap := 50
	rsID := "ocid1.lustrefilesystem.oc1.phx.rs"
	fl := &MockOCILustreFileStorageClient{
		ListResp:        []lustre.LustreFileSystemSummary{},
		CreateResp:      &lustre.LustreFileSystem{Id: &rsID},
		AwaitActiveResp: &lustre.LustreFileSystem{Id: &rsID, CapacityInGBs: &cap},
	}
	d := newControllerWith(fl, fid)
	params := map[string]string{
		"subnetId":                   "ocid1.subnet.oc1..x",
		"performanceTier":            "MBPS_PER_TB_125",
		"availabilityDomain":         "PHX-AD-2",
		"rootSquashEnabled":          "true",
		"rootSquashClientExceptions": "[\"client1\",\"client2\"]",
		"rootSquashUid":              "1000",
		"rootSquashGid":              "1000",
	}
	req := &csi.CreateVolumeRequest{
		Name:               "test-vol-root-squash",
		CapacityRange:      &csi.CapacityRange{RequiredBytes: 31200 * GB},
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
		Parameters:         params,
	}
	_, _ = d.CreateVolume(context.Background(), req)
	if fl.CreateArgs == nil || fl.CreateArgs.RootSquashConfiguration == nil {
		t.Fatalf("expected RootSquashConfiguration to be set")
	}
	if fl.CreateArgs.RootSquashConfiguration.IdentitySquash != lustre.RootSquashConfigurationIdentitySquashRoot {
		t.Fatalf("expected IdentitySquash to be Root")
	}
	if len(fl.CreateArgs.RootSquashConfiguration.ClientExceptions) != 2 {
		t.Fatalf("expected 2 client exceptions, got %d", len(fl.CreateArgs.RootSquashConfiguration.ClientExceptions))
	}
	if fl.CreateArgs.RootSquashConfiguration.SquashUid == nil || *fl.CreateArgs.RootSquashConfiguration.SquashUid != 1000 {
		t.Fatalf("expected SquashUid to be 1000")
	}
	if fl.CreateArgs.RootSquashConfiguration.SquashGid == nil || *fl.CreateArgs.RootSquashConfiguration.SquashGid != 1000 {
		t.Fatalf("expected SquashGid to be 1000")
	}
}

// ########################## Delete Volume Tests   ########################++
func TestDeleteVolume_AwaitTimeout(t *testing.T) {
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	fl := &MockOCILustreFileStorageClient{GetResp: &lustre.LustreFileSystem{Id: &fsID, LifecycleState: lustre.LustreFileSystemLifecycleStateActive}, AwaitDeletedErr: context.DeadlineExceeded}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	req := &csi.DeleteVolumeRequest{VolumeId: fsID + ":10.0.0.10@tcp:/fs1"}
	_, err := d.DeleteVolume(context.Background(), req)
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if !containsErr(err, "Error while waiting for Lustre Filesystem to be Deleted") {
		t.Fatalf("expected error message to contain waiting for delete substring, got %v", err)
	}
}

func TestDeleteVolume_InvalidVolumeID(t *testing.T) {
	fl := &MockOCILustreFileStorageClient{}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	_, err := d.DeleteVolume(context.Background(), &csi.DeleteVolumeRequest{VolumeId: "bad-id"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for bad id, got %v", err)
	}
}

func TestDeleteVolume_DeletingAwaitSuccess(t *testing.T) {
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	fl := &MockOCILustreFileStorageClient{GetResp: &lustre.LustreFileSystem{Id: &fsID, LifecycleState: lustre.LustreFileSystemLifecycleStateDeleting}, AwaitDeletedErr: nil}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	_, err := d.DeleteVolume(context.Background(), &csi.DeleteVolumeRequest{VolumeId: fsID + ":10.0.0.10@tcp:/fs1"})
	if err != nil {
		t.Fatalf("unexpected error awaiting delete: %v", err)
	}
}

func TestDeleteVolume_MissingVolumeId(t *testing.T) {
	fl := &MockOCILustreFileStorageClient{}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	_, err := d.DeleteVolume(context.Background(), &csi.DeleteVolumeRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestDeleteVolume_LustreClientNil(t *testing.T) {
	inner := &testOCIClient{lustre: nil, id: &MockOCIIdentityClient{}}
	d := newControllerWith(inner.Lustre(), inner.Identity(nil))
	req := &csi.DeleteVolumeRequest{VolumeId: "ocid1.lustrefilesystem.oc1.phx.id:10.0.0.10@tcp:/fs1"}
	_, err := d.DeleteVolume(context.Background(), req)
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal for nil lustre client, got %v", err)
	}
}

func TestDeleteVolume_GetNotFound(t *testing.T) {
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	notFoundErr := mockNotFoundError{
		statusCode:   404,
		message:      "not found",
		code:         "NotFound",
		opcRequestID: "",
	}
	fl := &MockOCILustreFileStorageClient{GetErr: notFoundErr}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	req := &csi.DeleteVolumeRequest{VolumeId: fsID + ":10.0.0.10@tcp:/fs1"}
	_, err := d.DeleteVolume(context.Background(), req)
	if err != nil {
		t.Fatalf("expected success for not found, got %v", err)
	}
}

func TestDeleteVolume_GetInternalError(t *testing.T) {
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	fl := &MockOCILustreFileStorageClient{GetErr: status.Error(codes.Internal, "internal server error")}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	req := &csi.DeleteVolumeRequest{VolumeId: fsID + ":10.0.0.10@tcp:/fs1"}
	_, err := d.DeleteVolume(context.Background(), req)
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", err)
	}
	if !containsErr(err, "internal server error") {
		t.Fatalf("expected error to mention internal server error, got %v", err)
	}
}

func TestDeleteVolume_StateDeleted_Success(t *testing.T) {
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	fl := &MockOCILustreFileStorageClient{GetResp: &lustre.LustreFileSystem{Id: &fsID, LifecycleState: lustre.LustreFileSystemLifecycleStateDeleted}}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	req := &csi.DeleteVolumeRequest{VolumeId: fsID + ":10.0.0.10@tcp:/fs1"}
	_, err := d.DeleteVolume(context.Background(), req)
	if err != nil {
		t.Fatalf("expected success for deleted state, got %v", err)
	}
}

func TestDeleteVolume_Active_DeleteAwaitSuccess(t *testing.T) {
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	fl := &MockOCILustreFileStorageClient{
		GetResp:         &lustre.LustreFileSystem{Id: &fsID, LifecycleState: lustre.LustreFileSystemLifecycleStateActive},
		DeleteErr:       nil,
		AwaitDeletedErr: nil,
	}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	req := &csi.DeleteVolumeRequest{VolumeId: fsID + ":10.0.0.10@tcp:/fs1"}
	_, err := d.DeleteVolume(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error for delete success: %v", err)
	}
}

func TestDeleteVolume_Active_DeleteError(t *testing.T) {
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	fl := &MockOCILustreFileStorageClient{
		GetResp:   &lustre.LustreFileSystem{Id: &fsID, LifecycleState: lustre.LustreFileSystemLifecycleStateActive},
		DeleteErr: status.Error(codes.Internal, "delete error"),
	}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	req := &csi.DeleteVolumeRequest{VolumeId: fsID + ":10.0.0.10@tcp:/fs1"}
	_, err := d.DeleteVolume(context.Background(), req)
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal for delete error, got %v", err)
	}
	if !containsErr(err, "delete error") {
		t.Fatalf("expected error to mention delete error, got %v", err)
	}
}

// ########################## ValidateVolumeCapabilities Tests ##########################

func TestValidateVolumeCapabilities_InvalidVolumeID(t *testing.T) {
	fl := &MockOCILustreFileStorageClient{}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           "bad-volume-id",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
	}
	_, err := d.ValidateVolumeCapabilities(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestValidateVolumeCapabilities_BlockModeUnsupported(t *testing.T) {
	fl := &MockOCILustreFileStorageClient{}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	// valid-looking VolumeId to get through initial format checks
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           fsID + ":10.0.0.10@tcp:/fs1",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Block{Block: &csi.VolumeCapability_BlockVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
	}
	_, err := d.ValidateVolumeCapabilities(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for block mode, got %v", err)
	}
}

func TestValidateVolumeCapabilities_IdentityMismatch(t *testing.T) {
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	fl := &MockOCILustreFileStorageClient{GetResp: &lustre.LustreFileSystem{Id: &fsID, ManagementServiceAddress: ptrString("10.0.0.10"), Lnet: ptrString("tcp"), FileSystemName: ptrString("fs1")}}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	// Provide mismatching handle (different ms address)
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           fsID + ":10.0.0.11@tcp:/fs1",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
	}
	_, err := d.ValidateVolumeCapabilities(context.Background(), req)
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound for identity mismatch, got %v", err)
	}
}

func TestValidateVolumeCapabilities_ValidPath(t *testing.T) {
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	ms := "10.0.0.10"
	lnet := "tcp"
	fsname := "fs1"
	fl := &MockOCILustreFileStorageClient{GetResp: &lustre.LustreFileSystem{Id: &fsID, ManagementServiceAddress: &ms, Lnet: &lnet, FileSystemName: &fsname}}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	reqCaps := []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}}
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           fsID + ":" + ms + "@" + lnet + ":/" + fsname,
		VolumeCapabilities: reqCaps,
	}
	resp, err := d.ValidateVolumeCapabilities(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetConfirmed() == nil || len(resp.GetConfirmed().GetVolumeCapabilities()) != 1 {
		t.Fatalf("expected confirmed capabilities echoed back")
	}
}
func TestValidateVolumeCapabilities_MissingCaps(t *testing.T) {
	d := newControllerWith(&MockOCILustreFileStorageClient{}, &MockOCIIdentityClient{})
	_, err := d.ValidateVolumeCapabilities(context.Background(), &csi.ValidateVolumeCapabilitiesRequest{VolumeId: "ocid1.lustrefilesystem.oc1.phx.id:10.0.0.10@tcp:/fs1"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for missing caps, got %v", err)
	}
}

func TestValidateVolumeCapabilities_UnsupportedAccessMode(t *testing.T) {
	// Simulate an unsupported mode by constructing a capability that checkLustreSupportedVolumeCapabilities rejects
	d := newControllerWith(&MockOCILustreFileStorageClient{GetResp: &lustre.LustreFileSystem{Id: ptrString("ocid1.lustrefilesystem.oc1.phx.id"), ManagementServiceAddress: ptrString("10.0.0.10"), Lnet: ptrString("tcp"), FileSystemName: ptrString("fs1")}}, &MockOCIIdentityClient{})
	cap := &csi.VolumeCapability{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: 999 /* invalid */}}
	_, err := d.ValidateVolumeCapabilities(context.Background(), &csi.ValidateVolumeCapabilitiesRequest{VolumeId: "ocid1.lustrefilesystem.oc1.phx.id:10.0.0.10@tcp:/fs1", VolumeCapabilities: []*csi.VolumeCapability{cap}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for unsupported access mode, got %v", err)
	}
}

func TestValidateVolumeCapabilities_MissingVolumeId(t *testing.T) {
	fl := &MockOCILustreFileStorageClient{}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
	}
	_, err := d.ValidateVolumeCapabilities(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for missing VolumeId, got %v", err)
	}
}

func TestValidateVolumeCapabilities_LustreClientNil(t *testing.T) {
	inner := &testOCIClient{lustre: nil, id: &MockOCIIdentityClient{}}
	d := newControllerWith(inner.Lustre(), inner.Identity(nil))
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           "ocid1.lustrefilesystem.oc1.phx.id:10.0.0.10@tcp:/fs1",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
	}
	_, err := d.ValidateVolumeCapabilities(context.Background(), req)
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal for nil lustre client, got %v", err)
	}
}

func TestValidateVolumeCapabilities_GetError(t *testing.T) {
	fsID := "ocid1.lustrefilesystem.oc1.phx.id"
	fl := &MockOCILustreFileStorageClient{GetErr: status.Error(codes.Internal, "get error")}
	d := newControllerWith(fl, &MockOCIIdentityClient{})
	req := &csi.ValidateVolumeCapabilitiesRequest{
		VolumeId:           fsID + ":10.0.0.10@tcp:/fs1",
		VolumeCapabilities: []*csi.VolumeCapability{{AccessType: &csi.VolumeCapability_Mount{Mount: &csi.VolumeCapability_MountVolume{}}, AccessMode: &csi.VolumeCapability_AccessMode{Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER}}},
	}
	_, err := d.ValidateVolumeCapabilities(context.Background(), req)
	if status.Code(err) != codes.NotFound {
		t.Fatalf("expected NotFound for get error, got %v", err)
	}
	if !containsErr(err, "get error") {
		t.Fatalf("expected error to mention get error, got %v", err)
	}
}

// ########################## ControllerGetCapabilities Tests ##########################
func TestLustreController_ControllerGetCapabilities_SingleCreateDeleteOnly(t *testing.T) {
	d := &LustreControllerDriver{}

	resp, err := d.ControllerGetCapabilities(context.Background(), &csi.ControllerGetCapabilitiesRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected non-nil response")
	}
	if len(resp.Capabilities) != 1 {
		t.Fatalf("expected 1 capability, got %d", len(resp.Capabilities))
	}
	cap := resp.Capabilities[0]
	if cap.GetRpc() == nil {
		t.Fatalf("expected RPC capability, got nil")
	}
	if got := cap.GetRpc().GetType(); got != csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME {
		t.Fatalf("expected capability CREATE_DELETE_VOLUME, got %v", got)
	}
}

// Unit test: method should be safe to call with nil request (it doesn't use the request)
func TestLustreController_ControllerGetCapabilities_NilRequest(t *testing.T) {
	d := &LustreControllerDriver{}
	resp, err := d.ControllerGetCapabilities(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || len(resp.Capabilities) == 0 {
		t.Fatalf("expected capabilities in response")
	}
}

// Unit test: method should not dereference the receiver; nil receiver should not panic
func TestLustreController_ControllerGetCapabilities_NilReceiver(t *testing.T) {
	var d *LustreControllerDriver // nil receiver
	// Call should not panic because implementation doesn't use the receiver fields
	resp, err := d.ControllerGetCapabilities(context.Background(), &csi.ControllerGetCapabilitiesRequest{})
	if err != nil {
		t.Fatalf("unexpected error with nil receiver: %v", err)
	}
	if resp == nil || len(resp.Capabilities) != 1 {
		t.Fatalf("expected exactly one capability with nil receiver")
	}
	if resp.Capabilities[0].GetRpc().GetType() != csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME {
		t.Fatalf("expected CREATE_DELETE_VOLUME with nil receiver")
	}
}

// ########################## Helper Tests ##########################

func TestHelper_extractLustreFilesystemId(t *testing.T) {
	got := extractLustreFilesystemId("ocid1.lustrefilesystem.oc1.phx.id:10.0.0.10@tcp:/fs")
	if got == "" {
		t.Fatalf("expected non-empty ID extract")
	}
	if got != "ocid1.lustrefilesystem.oc1.phx.id" {
		t.Fatalf("expected ocid1.lustrefilesystem.oc1.phx.id, got %v", got)
	}
	if got := extractLustreFilesystemId("bad-id"); got != "" {
		t.Fatalf("expected empty for malformed, got %q", got)
	}
	if got := extractLustreFilesystemId("ocid1.lustrefilesystem.oc1.phx.id10.0.0.10@tcp/fs"); got != "" {
		t.Fatalf("expected empty for malformed, got %q", got)
	}
}

func TestHelper_buildLustreVolumeHandle(t *testing.T) {
	id := "ocid1.lustrefilesystem.oc1.phx.zabc"
	msa := "10.0.0.10"
	lnet := "tcp"
	fs := "fs1"
	got := buildLustreVolumeHandle(&lustre.LustreFileSystem{Id: &id, ManagementServiceAddress: &msa, Lnet: &lnet, FileSystemName: &fs})
	expected := id + ":" + msa + "@" + lnet + ":/" + fs
	if got != expected {
		t.Fatalf("handle mismatch: got %q, expected %q", got, expected)
	}
}

//########################### Controller Setup ###############################3

func newControllerWith(flustre client.LustreInterface, fid client.IdentityInterface) *LustreControllerDriver {
	logger, _ := zap.NewDevelopment()
	cd := &ControllerDriver{config: &providercfg.Config{CompartmentID: "ocid1.compartment.oc1..unit", Tags: &providercfg.InitialTags{
		LoadBalancer: nil,
		BlockVolume: &providercfg.TagConfig{
			FreeformTags: map[string]string{"Project": "Lustre"},
			DefinedTags:  map[string]map[string]interface{}{"orclcontainerengine": {"cluster": "ocid1.cluster..."}},
		},
	}}, client: &testOCIClient{lustre: flustre, id: fid}, logger: logger.Sugar()}
	os.Setenv("CPO_ENABLE_RESOURCE_ATTRIBUTION", "true")
	return &LustreControllerDriver{ControllerDriver: cd}
}

type testOCIClient struct {
	lustre client.LustreInterface
	id     client.IdentityInterface
}

func (t *testOCIClient) Compute() client.ComputeInterface { return nil }
func (t *testOCIClient) LoadBalancer(*zap.SugaredLogger, string, *client.OCIClientConfig) client.GenericLoadBalancerInterface {
	return nil
}
func (t *testOCIClient) Networking(*client.OCIClientConfig) client.NetworkingInterface { return nil }
func (t *testOCIClient) BlockStorage() client.BlockStorageInterface                    { return nil }
func (t *testOCIClient) FSS(*client.OCIClientConfig) client.FileStorageInterface       { return nil }
func (t *testOCIClient) Lustre() client.LustreInterface                                { return t.lustre }
func (t *testOCIClient) Identity(*client.OCIClientConfig) client.IdentityInterface     { return t.id }
func (t *testOCIClient) ContainerEngine() client.ContainerEngineInterface              { return nil }
func (t *testOCIClient) NewWorkloadIdentityClient(*zap.SugaredLogger, string, *client.OCIClientConfig) client.Interface {
	return t
}
func (t *testOCIClient) CertManager() client.CertificateManagerInterface { return nil }

// Keep for compatibility if other tests add additional helpers here.

func ptrString(s string) *string { return &s }

func containsErr(err error, substr string) bool {
	if err == nil {
		return false
	}
	msg := status.Convert(err).Message()
	return len(msg) > 0 && strings.Contains(msg, substr)
}
