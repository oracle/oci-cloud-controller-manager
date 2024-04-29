package util

import (
	"context"
	"fmt"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/discovery"
	v11 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	v1alpha11 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1alpha1"
	v1beta11 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
	v1alpha12 "k8s.io/client-go/kubernetes/typed/apiserverinternal/v1alpha1"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	v1beta12 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	v1beta21 "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
	v14 "k8s.io/client-go/kubernetes/typed/authentication/v1"
	v1alpha13 "k8s.io/client-go/kubernetes/typed/authentication/v1alpha1"
	v1beta13 "k8s.io/client-go/kubernetes/typed/authentication/v1beta1"
	v15 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	v1beta14 "k8s.io/client-go/kubernetes/typed/authorization/v1beta1"
	v16 "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	v21 "k8s.io/client-go/kubernetes/typed/autoscaling/v2"
	v2beta11 "k8s.io/client-go/kubernetes/typed/autoscaling/v2beta1"
	v2beta21 "k8s.io/client-go/kubernetes/typed/autoscaling/v2beta2"
	v17 "k8s.io/client-go/kubernetes/typed/batch/v1"
	v1beta15 "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	v18 "k8s.io/client-go/kubernetes/typed/certificates/v1"
	v1alpha18 "k8s.io/client-go/kubernetes/typed/certificates/v1alpha1"
	v1beta16 "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	v19 "k8s.io/client-go/kubernetes/typed/coordination/v1"
	v1beta17 "k8s.io/client-go/kubernetes/typed/coordination/v1beta1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"
	v110 "k8s.io/client-go/kubernetes/typed/discovery/v1"
	v1beta18 "k8s.io/client-go/kubernetes/typed/discovery/v1beta1"
	v111 "k8s.io/client-go/kubernetes/typed/events/v1"
	v1beta19 "k8s.io/client-go/kubernetes/typed/events/v1beta1"
	v1beta110 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	v13 "k8s.io/client-go/kubernetes/typed/flowcontrol/v1"
	v1beta111 "k8s.io/client-go/kubernetes/typed/flowcontrol/v1beta1"
	v1beta22 "k8s.io/client-go/kubernetes/typed/flowcontrol/v1beta2"
	v1beta31 "k8s.io/client-go/kubernetes/typed/flowcontrol/v1beta3"
	v112 "k8s.io/client-go/kubernetes/typed/networking/v1"
	v1alpha15 "k8s.io/client-go/kubernetes/typed/networking/v1alpha1"
	v1beta112 "k8s.io/client-go/kubernetes/typed/networking/v1beta1"
	v113 "k8s.io/client-go/kubernetes/typed/node/v1"
	v1alpha16 "k8s.io/client-go/kubernetes/typed/node/v1alpha1"
	v1beta113 "k8s.io/client-go/kubernetes/typed/node/v1beta1"
	v114 "k8s.io/client-go/kubernetes/typed/policy/v1"
	v1beta114 "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
	v115 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	v1alpha17 "k8s.io/client-go/kubernetes/typed/rbac/v1alpha1"
	v1beta115 "k8s.io/client-go/kubernetes/typed/rbac/v1beta1"
	"k8s.io/client-go/kubernetes/typed/resource/v1alpha2"
	v116 "k8s.io/client-go/kubernetes/typed/scheduling/v1"
	v1alpha19 "k8s.io/client-go/kubernetes/typed/scheduling/v1alpha1"
	v1beta116 "k8s.io/client-go/kubernetes/typed/scheduling/v1beta1"
	v117 "k8s.io/client-go/kubernetes/typed/storage/v1"
	"k8s.io/client-go/kubernetes/typed/storage/v1alpha1"
	v1beta117 "k8s.io/client-go/kubernetes/typed/storage/v1beta1"
	alpha1 `k8s.io/client-go/kubernetes/typed/storagemigration/v1alpha1`
	"k8s.io/client-go/rest"
)

type MockKubeClient struct {
	CoreClient *MockCoreClient
}

func (m MockKubeClient) StoragemigrationV1alpha1() alpha1.StoragemigrationV1alpha1Interface {
	return nil
}

type MockCoreClient v12.CoreV1Client

type MockNodes struct {
	client rest.Interface
}

func (m MockNodes) Create(ctx context.Context, node *api.Node, opts metav1.CreateOptions) (*api.Node, error) {
	return nil, nil
}

func (m MockNodes) Update(ctx context.Context, node *api.Node, opts metav1.UpdateOptions) (*api.Node, error) {
	return nil, nil
}

func (m MockNodes) UpdateStatus(ctx context.Context, node *api.Node, opts metav1.UpdateOptions) (*api.Node, error) {
	return nil, nil
}

func (m MockNodes) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return nil
}

func (m MockNodes) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return nil
}

func (m MockNodes) List(ctx context.Context, opts metav1.ListOptions) (*api.NodeList, error) {
	return nil, nil
}

func (m MockNodes) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return nil, nil
}

func (m MockNodes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *api.Node, err error) {
	return nil, nil
}

func (m MockNodes) Apply(ctx context.Context, node *applyconfigurationscorev1.NodeApplyConfiguration, opts metav1.ApplyOptions) (result *api.Node, err error) {
	return nil, nil
}

func (m MockNodes) ApplyStatus(ctx context.Context, node *applyconfigurationscorev1.NodeApplyConfiguration, opts metav1.ApplyOptions) (result *api.Node, err error) {
	return nil, nil
}

func (m MockNodes) PatchStatus(ctx context.Context, nodeName string, data []byte) (*api.Node, error) {
	return nil, nil
}

func (m MockCoreClient) RESTClient() rest.Interface {
	return nil
}

func (m MockCoreClient) ComponentStatuses() v12.ComponentStatusInterface {
	return nil
}

func (m MockCoreClient) ConfigMaps(namespace string) v12.ConfigMapInterface {
	return nil
}

func (m MockCoreClient) Endpoints(namespace string) v12.EndpointsInterface {
	return nil
}

func (m MockCoreClient) Events(namespace string) v12.EventInterface {
	return nil
}

func (m MockCoreClient) LimitRanges(namespace string) v12.LimitRangeInterface {
	return nil
}

func (m MockCoreClient) Namespaces() v12.NamespaceInterface {
	return nil
}

func (m MockCoreClient) PersistentVolumes() v12.PersistentVolumeInterface {
	return nil
}

func (m MockCoreClient) PersistentVolumeClaims(namespace string) v12.PersistentVolumeClaimInterface {
	return nil
}

func (m MockCoreClient) Pods(namespace string) v12.PodInterface {
	return nil
}

func (m MockCoreClient) PodTemplates(namespace string) v12.PodTemplateInterface {
	return nil
}

func (m MockCoreClient) ReplicationControllers(namespace string) v12.ReplicationControllerInterface {
	return nil
}

func (m MockCoreClient) ResourceQuotas(namespace string) v12.ResourceQuotaInterface {
	return nil
}

func (m MockCoreClient) Secrets(namespace string) v12.SecretInterface {
	return nil
}

func (m MockCoreClient) Services(namespace string) v12.ServiceInterface {
	return nil
}

func (m MockCoreClient) ServiceAccounts(namespace string) v12.ServiceAccountInterface {
	return nil
}

func (m MockKubeClient) ResourceV1alpha2() v1alpha2.ResourceV1alpha2Interface {
	return nil
}

func (m MockKubeClient) CertificatesV1alpha1() v1alpha18.CertificatesV1alpha1Interface {
	return nil
}

func (m MockKubeClient) NetworkingV1() v112.NetworkingV1Interface {
	return nil
}

func (m MockKubeClient) NodeV1() v113.NodeV1Interface {
	return nil
}

func (m MockKubeClient) PolicyV1() v114.PolicyV1Interface {
	return nil
}

func (m MockKubeClient) RbacV1() v115.RbacV1Interface {
	return nil
}

func (m MockKubeClient) SchedulingV1() v116.SchedulingV1Interface {
	return nil
}

func (m MockKubeClient) StorageV1() v117.StorageV1Interface {
	return nil
}

func (m MockKubeClient) AuthenticationV1() v14.AuthenticationV1Interface {
	return nil
}

func (m MockKubeClient) AuthorizationV1() v15.AuthorizationV1Interface {
	return nil
}

func (m MockKubeClient) AutoscalingV1() v16.AutoscalingV1Interface {
	return nil
}

func (m MockKubeClient) BatchV1() v17.BatchV1Interface {
	return nil
}

func (m MockKubeClient) CertificatesV1() v18.CertificatesV1Interface {
	return nil
}

func (m MockKubeClient) CoordinationV1() v19.CoordinationV1Interface {
	return nil
}

func (m MockKubeClient) DiscoveryV1() v110.DiscoveryV1Interface {
	return nil
}

func (m MockKubeClient) EventsV1() v111.EventsV1Interface {
	return nil
}

func (m MockKubeClient) PolicyV1beta1() v1beta114.PolicyV1beta1Interface {
	return nil
}

func (m MockKubeClient) RbacV1beta1() v1beta115.RbacV1beta1Interface {
	return nil
}

func (m MockKubeClient) SchedulingV1beta1() v1beta116.SchedulingV1beta1Interface {
	return nil
}

func (m MockKubeClient) StorageV1beta1() v1beta117.StorageV1beta1Interface {
	return nil
}

func (m MockKubeClient) NodeV1beta1() v1beta113.NodeV1beta1Interface {
	return nil
}

func (m MockKubeClient) NetworkingV1beta1() v1beta112.NetworkingV1beta1Interface {
	return nil
}

func (m MockKubeClient) FlowcontrolV1beta1() v1beta111.FlowcontrolV1beta1Interface {
	return nil
}

func (m MockKubeClient) ExtensionsV1beta1() v1beta110.ExtensionsV1beta1Interface {
	return nil
}

func (m MockKubeClient) StorageV1alpha1() v1alpha1.StorageV1alpha1Interface {
	return nil
}

func (m MockKubeClient) EventsV1beta1() v1beta19.EventsV1beta1Interface {
	return nil
}

func (m MockKubeClient) SchedulingV1alpha1() v1alpha19.SchedulingV1alpha1Interface {
	return nil
}

func (m MockKubeClient) DiscoveryV1beta1() v1beta18.DiscoveryV1beta1Interface {
	return nil
}

func (m MockKubeClient) CoordinationV1beta1() v1beta17.CoordinationV1beta1Interface {
	return nil
}

func (m MockKubeClient) RbacV1alpha1() v1alpha17.RbacV1alpha1Interface {
	return nil
}

func (m MockKubeClient) CertificatesV1beta1() v1beta16.CertificatesV1beta1Interface {
	return nil
}

func (m MockKubeClient) NodeV1alpha1() v1alpha16.NodeV1alpha1Interface {
	return nil
}

func (m MockKubeClient) BatchV1beta1() v1beta15.BatchV1beta1Interface {
	return nil
}

func (m MockKubeClient) NetworkingV1alpha1() v1alpha15.NetworkingV1alpha1Interface {
	return nil
}

func (m MockKubeClient) AuthorizationV1beta1() v1beta14.AuthorizationV1beta1Interface {
	return nil
}

func (m MockKubeClient) AuthenticationV1alpha1() v1alpha13.AuthenticationV1alpha1Interface {
	return nil
}

func (m MockKubeClient) AuthenticationV1beta1() v1beta13.AuthenticationV1beta1Interface {
	return nil
}

func (m MockKubeClient) InternalV1alpha1() v1alpha12.InternalV1alpha1Interface {
	return nil
}

func (m MockKubeClient) AppsV1() v1.AppsV1Interface {
	return nil
}

func (m MockKubeClient) AppsV1beta1() v1beta12.AppsV1beta1Interface {
	return nil
}

func (m MockKubeClient) FlowcontrolV1beta2() v1beta22.FlowcontrolV1beta2Interface {
	return nil
}

func (m MockKubeClient) Discovery() discovery.DiscoveryInterface {
	return nil
}

func (m MockKubeClient) AdmissionregistrationV1() v11.AdmissionregistrationV1Interface {
	return nil
}

func (m MockKubeClient) AdmissionregistrationV1alpha1() v1alpha11.AdmissionregistrationV1alpha1Interface {
	return nil
}

func (m MockKubeClient) AdmissionregistrationV1beta1() v1beta11.AdmissionregistrationV1beta1Interface {
	return nil
}

func (m MockKubeClient) AppsV1beta2() v1beta21.AppsV1beta2Interface {
	return nil
}

func (m MockKubeClient) AutoscalingV2() v21.AutoscalingV2Interface {
	return nil
}

func (m MockKubeClient) AutoscalingV2beta1() v2beta11.AutoscalingV2beta1Interface {
	return nil
}

func (m MockKubeClient) AutoscalingV2beta2() v2beta21.AutoscalingV2beta2Interface {
	return nil
}

func (m MockKubeClient) FlowcontrolV1beta3() v1beta31.FlowcontrolV1beta3Interface {
	return nil
}

func (m MockKubeClient) FlowcontrolV1() v13.FlowcontrolV1Interface {
	return nil
}

// edited functions

func (m MockKubeClient) CoreV1() v12.CoreV1Interface {
	return m.CoreClient
}

func (m MockCoreClient) Nodes() v12.NodeInterface {
	return &MockNodes{
		client: m.RESTClient(),
	}
}

var (
	LabelIpFamilyPreferred = "oci.oraclecloud.com/ip-family-preferred"
	LabelIpFamilyIpv4      = "oci.oraclecloud.com/ip-family-ipv4"
	LabelIpFamilyIpv6      = "oci.oraclecloud.com/ip-family-ipv6"
	nodes                  = map[string]*api.Node{
		"ipv6Preferred": {
			Spec: api.NodeSpec{
				ProviderID: "sample-provider-id",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					LabelIpFamilyPreferred: "IPv6",
					LabelIpFamilyIpv4:      "true",
					LabelIpFamilyIpv6:      "true",
				},
			},
		},
		"ipv4Preferred": {
			Spec: api.NodeSpec{
				ProviderID: "sample-provider-id",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					LabelIpFamilyPreferred: "IPv4",
					LabelIpFamilyIpv4:      "true",
					LabelIpFamilyIpv6:      "true",
				},
			},
		},
		"noIpPreference": {
			Spec: api.NodeSpec{
				ProviderID: "sample-provider-id",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{},
			},
		},
		"sample-provider-id": {
			Spec: api.NodeSpec{
				ProviderID: "sample-provider-id",
			},
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "sample-compartment-id",
				},
			},
		},
		"sample-node-id": {
			Spec: api.NodeSpec{
				ProviderID: "sample-provider-id",
			},
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "sample-compartment-id",
				},
			},
		},
	}
)

func (m MockNodes) Get(ctx context.Context, name string, opts metav1.GetOptions) (*api.Node, error) {
	if node, ok := nodes[name]; ok {
		return node, nil
	}
	return nil, fmt.Errorf("Node Not Present")

}
