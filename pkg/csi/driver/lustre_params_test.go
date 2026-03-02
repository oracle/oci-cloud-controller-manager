package driver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// helper to call extractLustreStorageClassParameters
func parseParams(t *testing.T, p map[string]string, identity client.IdentityInterface) (*LustreStorageClassParameters, error) {
	d := &LustreControllerDriver{ControllerDriver: ControllerDriver{config: &providercfg.Config{CompartmentID: "ocid1.compartment.oc1..unit-test"}}}
	_, _, parsed, err := extractLustreStorageClassParameters(context.Background(), d, zap.S(), "vol-1", p, identity)
	if err != nil {
		return nil, err
	}
	return parsed, err
}

func TestLustreParams_MissingSubnetId(t *testing.T) {
	p := map[string]string{
		"performanceTier": "MBPS_PER_TB_125", "availabilityDomain": "PHX-AD-2",
	}
	_, err := parseParams(t, p, nil)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestLustreParams_PerformanceTierValidation(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "mbps_per_tb_250",
		"availabilityDomain": "PHX-AD-2",
	}
	parsed, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if parsed.PerformanceTier != "MBPS_PER_TB_250" {
		t.Fatalf("expected canonical tier MBPS_PER_TB_250 got %s", parsed.PerformanceTier)
	}

	p["performanceTier"] = "FASTEST"
	_, err = parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for invalid tier, got %v", err)
	}
}

func TestLustreParams_NSGIdsParsing(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
	}
	ids := []string{"ocid1.nsg.oc1..a", "ocid1.nsg.oc1..b"}
	b, _ := json.Marshal(ids)
	p["nsgIds"] = string(b)
	parsed, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(parsed.NSGIds) != 2 {
		t.Fatalf("expected 2 nsg ids, got %v", parsed.NSGIds)
	}

	p["nsgIds"] = "not-json"
	_, err = parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for bad json, got %v", err)
	}
}

func TestLustreParams_SetupLnetAndCidrAndPostMountParams(t *testing.T) {
	p := map[string]string{
		"subnetId":                  "ocid1.subnet.oc1..example",
		"performanceTier":           "MBPS_PER_TB_500",
		"availabilityDomain":        "PHX-AD-2",
		"setupLnet":                 "TrUE",
		"lustreSubnetCidr":          "10.0.0.0/24",
		"lustrePostMountParameters": "[{}]",
	}
	parsed, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if parsed.SetupLnet != "true" {
		t.Fatalf("expected setupLnet normalized true, got %s", parsed.SetupLnet)
	}

	p["lustreSubnetCidr"] = "bad-cidr"
	_, err = parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for bad cidr, got %v", err)
	}

	p["lustreSubnetCidr"] = "10.0.1.0/24"
	p["lustrePostMountParameters"] = "not-json"
	_, err = parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for bad post mount params, got %v", err)
	}
}

func TestLustreParams_RootSquashValidation(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
		"rootSquashEnabled":  "truely",
		"rootSquashUid":      "-1",
	}
	_, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for negative uid, got %v", err)
	}

	p["rootSquashUid"] = "1000"
	// >10 exceptions
	ex := make([]string, 11)
	for i := 0; i < 11; i++ {
		ex[i] = "10.0.0.1"
	}
	b, _ := json.Marshal(ex)
	p["rootSquashClientExceptions"] = string(b)
	_, err = parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for >10 exceptions, got %v", err)
	}
}

func TestLustreParams_FileSystemNameValidation(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
		"fileSystemName":     "INVALID-NAME",
	}
	_, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for invalid fs name, got %v", err)
	}

	p["fileSystemName"] = "mylustre"
	parsed, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if parsed.FileSystemName != "mylustre" {
		t.Fatalf("expected mylustre, got %s", parsed.FileSystemName)
	}
}

func TestLustreParams_AvailabilityDomain_ExplicitValidNormalization(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
	}
	d := &LustreControllerDriver{ControllerDriver: ControllerDriver{config: &providercfg.Config{CompartmentID: "ocid1.compartment.oc1..unit-test"}}}
	_, _, parsed, err := extractLustreStorageClassParameters(context.Background(), d, zap.S(), "vol-xyz", p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.AvailabilityDomain == "PHX-AD-2" || parsed.AvailabilityDomain == "" {
		t.Fatalf("expected normalized AD name, got %q", parsed.AvailabilityDomain)
	}
}

func TestLustreParams_TagsOverrides_ValidAndInvalid(t *testing.T) {
	free := map[string]string{"env": "dev"}
	def := map[string]map[string]interface{}{"ns": {"k": "v"}}
	freeJSON, _ := json.Marshal(free)
	defJSON, _ := json.Marshal(def)
	p := map[string]string{
		"subnetId":                  "ocid1.subnet.oc1..example",
		"performanceTier":           "MBPS_PER_TB_125",
		"availabilityDomain":        "PHX-AD-2",
		initialFreeformTagsOverride: string(freeJSON),
		initialDefinedTagsOverride:  string(defJSON),
	}
	parsed, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if parsed.SCTags.FreeformTags["env"] != "dev" {
		t.Fatalf("expected freeform tag env=dev, got %v", parsed.SCTags.FreeformTags)
	}
	if _, ok := parsed.SCTags.DefinedTags["ns"]; !ok {
		t.Fatalf("expected defined tag ns")
	}

	// invalid JSON
	p[initialFreeformTagsOverride] = "not-json"
	_, err = parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for invalid freeform tags json, got %v", err)
	}
	delete(p, initialFreeformTagsOverride)
	p[initialDefinedTagsOverride] = "not-json"
	_, err = parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for invalid defined tags json, got %v", err)
	}
}

func TestLustreParams_NSGIds_Omitted(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
	}
	parsed, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(parsed.NSGIds) != 0 {
		t.Fatalf("expected empty NSGIds when omitted, got %v", parsed.NSGIds)
	}
}

func TestLustreParams_DeriveDefaultFSName(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
	}
	d := &LustreControllerDriver{ControllerDriver: ControllerDriver{config: &providercfg.Config{CompartmentID: "ocid1.compartment.oc1..unit-test"}}}
	_, _, parsed, err := extractLustreStorageClassParameters(context.Background(), d, zap.S(), "vol-abc_DEF123", p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.FileSystemName == "" || len(parsed.FileSystemName) > 8 {
		t.Fatalf("expected derived fs name <=8 chars, got %q", parsed.FileSystemName)
	}
}

func TestDeriveDefaultFSName_Empty(t *testing.T) {
	got := deriveDefaultFSName("")
	if got != "pvlustre" {
		t.Fatalf("expected 'pvlustre' for empty volumeName, got %q", got)
	}
}

func TestDeriveDefaultFSName_WithSixLetters(t *testing.T) {
	got := deriveDefaultFSName("abcdef")
	if got != "pvabcdef" {
		t.Fatalf("expected 'pvabcdef' for 'abcdef', got %q", got)
	}
}

func TestDeriveDefaultFSName_WithMoreThanSixLetters(t *testing.T) {
	got := deriveDefaultFSName("abcdefghijk")
	if got != "pvfghijk" {
		t.Fatalf("expected 'pvfghijk' for 'abcdefghijk', got %q", got)
	}
}

func TestDeriveDefaultFSName_WithFewerThanSixLetters(t *testing.T) {
	got := deriveDefaultFSName("abc")
	if got != "pvabc" {
		t.Fatalf("expected 'pvabc' for 'abc', got %q", got)
	}
}

func TestDeriveDefaultFSName_NoLetters(t *testing.T) {
	got := deriveDefaultFSName("123456")
	if got != "pv123456" {
		t.Fatalf("expected 'pv123456' for '123456', got %q", got)
	}
}

func TestDeriveDefaultFSName_WithNumbersAndSymbols(t *testing.T) {
	got := deriveDefaultFSName("a1b@c2d#e3f")
	if got != "pvc2de3f" {
		t.Fatalf("expected 'pvc2de3f' for 'a1b@c2d#e3f', got %q", got)
	}
}

func TestDeriveDefaultFSName_CaseSensitive(t *testing.T) {
	got := deriveDefaultFSName("AbCdEf")
	if got != "pvAbCdEf" {
		t.Fatalf("expected 'pvAbCdEf' for 'AbCdEf', got %q", got)
	}
}

func TestDeriveDefaultFSName_MixedCaseAndNumbers(t *testing.T) {
	got := deriveDefaultFSName("Vol123Abc")
	if got != "pv123Abc" {
		t.Fatalf("expected 'pv123Abc' for 'Vol123Abc', got %q", got)
	}
}

func TestDeriveDefaultFSName_SingleLetter(t *testing.T) {
	got := deriveDefaultFSName("a")
	if got != "pva" {
		t.Fatalf("expected 'pva' for 'a', got %q", got)
	}
}

func TestDeriveDefaultFSName_AllLetters(t *testing.T) {
	got := deriveDefaultFSName("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if got != "pvUVWXYZ" {
		t.Fatalf("expected 'pvUVWXYZ' for 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', got %q", got)
	}
}

func TestValidateLustreFileSystemName_Empty(t *testing.T) {
	err := validateLustreFileSystemName("")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected error containing 'empty', got %v", err)
	}
}

func TestValidateLustreFileSystemName_InvalidPattern(t *testing.T) {
	err := validateLustreFileSystemName("invalid@")
	if err == nil || !strings.Contains(err.Error(), "allowed characters") {
		t.Fatalf("expected error containing 'allowed characters', got %v", err)
	}
}

func TestLustreParams_InvalidSetupLnet(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
		"setupLnet":          "invalid",
	}
	_, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument || !strings.Contains(err.Error(), "Invalid setupLnet") {
		t.Fatalf("expected InvalidArgument with 'Invalid setupLnet', got %v", err)
	}
}

func TestLustreParams_InvalidRootSquashClientExceptions(t *testing.T) {
	p := map[string]string{
		"subnetId":                   "ocid1.subnet.oc1..example",
		"performanceTier":            "MBPS_PER_TB_125",
		"availabilityDomain":         "PHX-AD-2",
		"rootSquashClientExceptions": "not-json",
	}
	_, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument || !strings.Contains(err.Error(), "Failed to parse rootSquashClientExceptions") {
		t.Fatalf("expected InvalidArgument with parse error, got %v", err)
	}
}

func TestLustreParams_MissingPerformanceTier(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"availabilityDomain": "PHX-AD-2",
	}
	_, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument || !strings.Contains(err.Error(), "Missing required parameter: performanceTier") {
		t.Fatalf("expected InvalidArgument for missing performanceTier, got %v", err)
	}
}

func TestLustreParams_CompartmentIdPassed(t *testing.T) {
	testComp := "ocid1.compartment.oc1..test"
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
		"compartmentId":      testComp,
	}
	parsed, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if parsed.CompartmentId != testComp {
		t.Fatalf("expected compartmentId %q, got %q", testComp, parsed.CompartmentId)
	}
}

func TestLustreParams_NonIntRootSquashUid(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
		"rootSquashUid":      "abc",
	}
	_, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument || !strings.Contains(err.Error(), "Invalid rootSquashUid") {
		t.Fatalf("expected InvalidArgument for non-int uid, got %v", err)
	}
}

func TestLustreParams_NonIntRootSquashGid(t *testing.T) {
	p := map[string]string{
		"subnetId":           "ocid1.subnet.oc1..example",
		"performanceTier":    "MBPS_PER_TB_125",
		"availabilityDomain": "PHX-AD-2",
		"rootSquashGid":      "abc",
	}
	_, err := parseParams(t, p, &MockOCIIdentityClient{ads: []string{"phx:PHX-AD-2"}})
	if status.Code(err) != codes.InvalidArgument || !strings.Contains(err.Error(), "Invalid rootSquashGid") {
		t.Fatalf("expected InvalidArgument for non-int gid, got %v", err)
	}
}
