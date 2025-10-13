package e2e

import (
	"context"
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	v1 "k8s.io/api/core/v1"
)

func validateDefinedTags(instanceTags map[string]map[string]interface{}, expected map[string]map[string]interface{}) error {
	for ns, expectedTags := range expected {
		actualTags, ok := instanceTags[ns]
		if !ok {
			return fmt.Errorf("missing defined tag namespace %s", ns)
		}
		for key, value := range expectedTags {
			actualValue, ok := actualTags[key]
			if !ok {
				return fmt.Errorf("missing defined tag %s/%s", ns, key)
			}
			if actualValue != value {
				return fmt.Errorf("defined tag %s/%s mismatch: expected %v got %v", ns, key, value, actualValue)
			}
		}
	}
	return nil
}

const (
	e2eOpenShiftTagNamespace = "openshift-tags"
	e2eOpenShiftTagKey       = "openshift-resource"
	e2eOpenShiftTagValue     = "openshift-resource-infra"
	openshiftNodeLabelEnv    = "OPENSHIFT_NODE_LABEL_ID"
	e2eOCIProviderPrefix     = "oci://"
)

// isOSOCluster returns true when OPENSHIFT_NODE_LABEL_ID matches at least one node label in the cluster.
func isOSOCluster(nodes []v1.Node) bool {
	labelIdentifier := strings.TrimSpace(os.Getenv(openshiftNodeLabelEnv))
	if labelIdentifier == "" {
		return false
	}

	var (
		labelKey   string
		labelValue string
	)
	if strings.Contains(labelIdentifier, "=") {
		parts := strings.SplitN(labelIdentifier, "=", 2)
		labelKey = strings.TrimSpace(parts[0])
		labelValue = strings.TrimSpace(parts[1])
	} else {
		labelKey = labelIdentifier
	}

	if labelKey == "" {
		return false
	}

	for _, node := range nodes {
		value, exists := node.Labels[labelKey]
		if !exists {
			continue
		}
		if labelValue == "" || value == labelValue {
			return true
		}
	}

	return false
}

var _ = Describe("Tagging Controller", func() {
	f := sharedfw.NewDefaultFramework("tagging-controller")

	var (
		nodes []v1.Node
	)

	BeforeEach(func() {
		nodeList := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet)
		Expect(len(nodeList.Items)).NotTo(BeZero())
		nodes = nodeList.Items
	})

	Context("[cloudprovider][tagging-controller]", func() {
		It("ensures common defined tags are present on instances", func() {

			if !isOSOCluster(nodes) {
				Skip(fmt.Sprintf("%s not detected on any node", openshiftNodeLabelEnv))
			}

			if f.CloudProviderConfig == nil || f.CloudProviderConfig.Tags == nil || f.CloudProviderConfig.Tags.Common == nil || len(f.CloudProviderConfig.Tags.Common.DefinedTags) == 0 {
				Skip("no common defined tags configured")
			}

			computeClient := f.Client.Compute()
			ctx := context.Background()
			for _, node := range nodes {
				Expect(node.Spec.ProviderID).To(HavePrefix(e2eOCIProviderPrefix), fmt.Sprintf("node %s providerID must have %s prefix in OSO clusters", node.Name, e2eOCIProviderPrefix))
				instanceID, err := oci.MapProviderIDToResourceID(node.Spec.ProviderID)
				Expect(err).NotTo(HaveOccurred())

				instance, err := computeClient.GetInstance(ctx, instanceID)
				Expect(err).NotTo(HaveOccurred())

				Expect(validateDefinedTags(instance.DefinedTags, f.CloudProviderConfig.Tags.Common.DefinedTags)).NotTo(HaveOccurred(),
					"defined tags missing for instance %s", instanceID)
			}
		})

		It("ensures OpenShift defined tag is applied", func() {
			if !isOSOCluster(nodes) {
				Skip(fmt.Sprintf("%s not detected on any node", openshiftNodeLabelEnv))
			}

			computeClient := f.Client.Compute()
			ctx := context.Background()
			for _, node := range nodes {
				Expect(node.Spec.ProviderID).To(HavePrefix(e2eOCIProviderPrefix), fmt.Sprintf("node %s providerID must have %s prefix in OSO clusters", node.Name, e2eOCIProviderPrefix))
				instanceID, err := oci.MapProviderIDToResourceID(node.Spec.ProviderID)
				Expect(err).NotTo(HaveOccurred())

				instance, err := computeClient.GetInstance(ctx, instanceID)
				Expect(err).NotTo(HaveOccurred())

				ns := e2eOpenShiftTagNamespace
				Expect(instance.DefinedTags).To(HaveKey(ns), "instance %s missing OpenShift namespace %s", instanceID, ns)
				Expect(instance.DefinedTags[ns]).To(HaveKeyWithValue(e2eOpenShiftTagKey, e2eOpenShiftTagValue),
					"instance %s missing OpenShift defined tag", instanceID)
			}
		})
	})
})
