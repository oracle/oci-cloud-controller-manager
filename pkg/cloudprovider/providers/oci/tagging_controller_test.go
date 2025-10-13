package oci

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	ociClient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestLogger(t *testing.T) *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	return logger.Sugar()
}

func TestHasRequiredDefinedTags(t *testing.T) {
	tc := &TaggingController{}
	tests := []struct {
		name     string
		instance *core.Instance
		required *config.TagConfig
		want     bool
	}{
		{
			name:     "no requirements",
			instance: &core.Instance{},
			required: nil,
			want:     true,
		},
		{
			name: "all defined tags present",
			instance: &core.Instance{
				DefinedTags: map[string]map[string]interface{}{
					"openshift-tags": {openshiftTagKey: openshiftTagValue},
				},
			},
			required: &config.TagConfig{
				DefinedTags: map[string]map[string]interface{}{
					"openshift-tags": {openshiftTagKey: openshiftTagValue},
				},
			},
			want: true,
		},
		{
			name: "missing defined namespace",
			instance: &core.Instance{
				DefinedTags: map[string]map[string]interface{}{},
			},
			required: &config.TagConfig{
				DefinedTags: map[string]map[string]interface{}{
					"openshift-tags": {openshiftTagKey: openshiftTagValue},
				},
			},
			want: false,
		},
		{
			name: "missing defined tag value",
			instance: &core.Instance{
				DefinedTags: map[string]map[string]interface{}{
					"openshift-tags": {"other": "value"},
				},
			},
			required: &config.TagConfig{
				DefinedTags: map[string]map[string]interface{}{
					"openshift-tags": {openshiftTagKey: openshiftTagValue},
				},
			},
			want: false,
		},
		{
			name: "no required defined tags",
			instance: &core.Instance{
				DefinedTags: map[string]map[string]interface{}{
					"foo": {"bar": "baz"},
				},
			},
			required: &config.TagConfig{DefinedTags: map[string]map[string]interface{}{}},
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requiredDefinedTags map[string]map[string]interface{}
			if tt.required != nil {
				requiredDefinedTags = tt.required.DefinedTags
			}
			if got := tc.hasRequiredDefinedTags(tt.instance.DefinedTags, requiredDefinedTags); got != tt.want {
				t.Fatalf("hasRequiredDefinedTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapProviderIDToResourceID_Basic(t *testing.T) {
	tests := []struct {
		name       string
		providerID string
		wantID     string
		wantError  string
	}{
		{
			name:       "valid oci provider prefix",
			providerID: providerPrefix + "ocid1.instance.oc1..aaaaaaaaxxx",
			wantID:     "ocid1.instance.oc1..aaaaaaaaxxx",
			wantError:  "",
		},
		{
			name:       "empty provider id",
			providerID: "",
			wantID:     "",
			wantError:  "empty",
		},
		{
			name:       "wrong provider prefix",
			providerID: "aws://i-xxxxx",
			wantID:     "aws://i-xxxxx",
			wantError:  "not valid for oci",
		},
		{
			name:       "raw ocid without provider prefix",
			providerID: "ocid1.instance.oc1..rawid",
			wantID:     "ocid1.instance.oc1..rawid",
			wantError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := MapProviderIDToResourceID(tt.providerID)
			if gotID != tt.wantID {
				t.Errorf("got ID %q, want %q", gotID, tt.wantID)
			}
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("expected error containing %q, got %v", tt.wantError, err)
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

// --- ReconcileNodeTags tests ---

// recordingComputeClient wraps MockComputeClient to record UpdateInstance invocations and return a custom instance.
type recordingComputeClient struct {
	MockComputeClient
	instance     *core.Instance
	updateCalled bool
}

func (r *recordingComputeClient) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	return r.instance, nil
}

func (r *recordingComputeClient) UpdateInstance(ctx context.Context, request core.UpdateInstanceRequest) (*core.Instance, error) {
	r.updateCalled = true
	return r.instance, nil
}

// recordingOCIClient implements client.Interface returning the provided compute client and stubs others.
type recordingOCIClient struct{ compute ociClient.ComputeInterface }

func (m recordingOCIClient) Compute() ociClient.ComputeInterface { return m.compute }
func (m recordingOCIClient) LoadBalancer(*zap.SugaredLogger, string, *ociClient.OCIClientConfig) ociClient.GenericLoadBalancerInterface {
	return nil
}
func (m recordingOCIClient) Networking(*ociClient.OCIClientConfig) ociClient.NetworkingInterface {
	return nil
}
func (m recordingOCIClient) BlockStorage() ociClient.BlockStorageInterface { return nil }
func (m recordingOCIClient) FSS(*ociClient.OCIClientConfig) ociClient.FileStorageInterface {
	return nil
}
func (m recordingOCIClient) Identity(*ociClient.OCIClientConfig) ociClient.IdentityInterface {
	return nil
}
func (m recordingOCIClient) ContainerEngine() ociClient.ContainerEngineInterface { return nil }
func (m recordingOCIClient) NewWorkloadIdentityClient(*zap.SugaredLogger, string, *ociClient.OCIClientConfig) ociClient.Interface {
	return m
}

func TestReconcileNodeTags_SkipWhenAtLimit(t *testing.T) {
	coreLogger, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(coreLogger).Sugar()

	// Create an instance with 64 defined tags and missing the required OpenShift tag
	dt := map[string]map[string]interface{}{"ns": {}}
	for i := 0; i < 64; i++ {
		key := fmt.Sprintf("k%d", i)
		dt["ns"][key] = "v"
	}
	inst := &core.Instance{DefinedTags: dt}

	rcc := &recordingComputeClient{instance: inst}
	roc := recordingOCIClient{compute: rcc}

	cp := &CloudProvider{config: &providercfg.Config{Tags: &providercfg.InitialTags{Common: &config.TagConfig{DefinedTags: map[string]map[string]interface{}{openshiftTagNamespace: {openshiftTagKey: openshiftTagValue}}}}}, logger: logger}

	tc := &TaggingController{logger: logger, cloud: cp, ociClient: roc}

	node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "n1"}, Spec: v1.NodeSpec{ProviderID: providerPrefix + "ocid1.limit"}}

	tc.ReconcileNodeTags(context.Background(), node)

	if rcc.updateCalled {
		t.Fatalf("expected UpdateInstance to be skipped when instance has 64 defined tags, but it was called")
	}

	// Optional: ensure a warning was logged about skipping due to limit
	foundWarn := false
	for _, e := range recorded.All() {
		if strings.Contains(e.Message, "Skipping update") {
			foundWarn = true
			break
		}
	}
	if !foundWarn {
		t.Fatalf("expected a warning log about skipping update due to tag limit")
	}
}
