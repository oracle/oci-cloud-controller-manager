package driver

import (
	"context"
	"encoding/json"
	"fmt"

	apiauthenticationv1 "k8s.io/api/authentication/v1"
	apicorev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	applyconfigurationscorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	fakediscovery "k8s.io/client-go/discovery/fake"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/testing"
)

// NewSimpleClientset returns a clientset that will respond with the provided objects.
// It's backed by a very simple object tracker that processes creates, updates and deletions as-is,
// without applying any validations and/or defaults. It shouldn't be considered a replacement
// for a real clientset and is mostly useful in simple unit tests.
func NewSimpleClientset(objects ...runtime.Object) *Clientset {
	return &Clientset{Clientset: fake.NewSimpleClientset(objects...)}
}

// Clientset implements clientset.Interface. Meant to be embedded into a
// struct to get a default implementation. This makes faking out just the method
// you want to test easier.
type Clientset struct {
	*fake.Clientset
	discovery *fakediscovery.FakeDiscovery
	tracker   testing.ObjectTracker
}

var (
	_ clientset.Interface = &fake.Clientset{}
	_ testing.FakeClient  = &fake.Clientset{}
)

// CoreV1 retrieves the CoreV1Client
func (c *Clientset) CoreV1() corev1.CoreV1Interface {
	return &FakeCoreV1{Fake: &c.Fake}
}

type FakeCoreV1 struct {
	*fakecorev1.FakeCoreV1
	*testing.Fake
}

func (c *FakeCoreV1) ServiceAccounts(namespace string) corev1.ServiceAccountInterface {
	return &FakeServiceAccounts{Fake: c, ns: namespace}
}

// FakeServiceAccounts implements ServiceAccountInterface
type FakeServiceAccounts struct {
	Fake *FakeCoreV1
	ns   string
}

var serviceaccountsResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccounts"}

var serviceaccountsKind = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"}

// Get takes name of the serviceAccount, and returns the corresponding serviceAccount object, and an error if there is any.
func (c *FakeServiceAccounts) Get(ctx context.Context, name string, options v1.GetOptions) (result *apicorev1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(serviceaccountsResource, c.ns, name), &apicorev1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*apicorev1.ServiceAccount), err
}

// List takes label and field selectors, and returns the list of ServiceAccounts that match those selectors.
func (c *FakeServiceAccounts) List(ctx context.Context, opts v1.ListOptions) (result *apicorev1.ServiceAccountList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(serviceaccountsResource, serviceaccountsKind, c.ns, opts), &apicorev1.ServiceAccountList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &apicorev1.ServiceAccountList{ListMeta: obj.(*apicorev1.ServiceAccountList).ListMeta}
	for _, item := range obj.(*apicorev1.ServiceAccountList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested serviceAccounts.
func (c *FakeServiceAccounts) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(serviceaccountsResource, c.ns, opts))

}

// Create takes the representation of a serviceAccount and creates it.  Returns the server's representation of the serviceAccount, and an error, if there is any.
func (c *FakeServiceAccounts) Create(ctx context.Context, serviceAccount *apicorev1.ServiceAccount, opts v1.CreateOptions) (result *apicorev1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(serviceaccountsResource, c.ns, serviceAccount), &apicorev1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*apicorev1.ServiceAccount), err
}

// Update takes the representation of a serviceAccount and updates it. Returns the server's representation of the serviceAccount, and an error, if there is any.
func (c *FakeServiceAccounts) Update(ctx context.Context, serviceAccount *apicorev1.ServiceAccount, opts v1.UpdateOptions) (result *apicorev1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(serviceaccountsResource, c.ns, serviceAccount), &apicorev1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*apicorev1.ServiceAccount), err
}

// Delete takes name of the serviceAccount and deletes it. Returns an error if one occurs.
func (c *FakeServiceAccounts) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(serviceaccountsResource, c.ns, name, opts), &apicorev1.ServiceAccount{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeServiceAccounts) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(serviceaccountsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &apicorev1.ServiceAccountList{})
	return err
}

// Patch applies the patch and returns the patched serviceAccount.
func (c *FakeServiceAccounts) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *apicorev1.ServiceAccount, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(serviceaccountsResource, c.ns, name, pt, data, subresources...), &apicorev1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*apicorev1.ServiceAccount), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied serviceAccount.
func (c *FakeServiceAccounts) Apply(ctx context.Context, serviceAccount *applyconfigurationscorev1.ServiceAccountApplyConfiguration, opts v1.ApplyOptions) (result *apicorev1.ServiceAccount, err error) {
	if serviceAccount == nil {
		return nil, fmt.Errorf("serviceAccount provided to Apply must not be nil")
	}
	data, err := json.Marshal(serviceAccount)
	if err != nil {
		return nil, err
	}
	name := serviceAccount.Name
	if name == nil {
		return nil, fmt.Errorf("serviceAccount.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(serviceaccountsResource, c.ns, *name, types.ApplyPatchType, data), &apicorev1.ServiceAccount{})

	if obj == nil {
		return nil, err
	}
	return obj.(*apicorev1.ServiceAccount), err
}

// CreateToken takes the representation of a tokenRequest and creates it.  Returns the server's representation of the tokenRequest, and an error, if there is any.
func (c *FakeServiceAccounts) CreateToken(ctx context.Context, serviceAccountName string, tokenRequest *apiauthenticationv1.TokenRequest, opts v1.CreateOptions) (result *apiauthenticationv1.TokenRequest, err error) {
	tokenRequest.Status.Token = "abc"
	return tokenRequest, nil
}
