/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"flag"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	wait "k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/cloudprovider"

	oci "github.com/oracle/oci-cloud-controller-manager/pkg/oci" // register oci cloud provider
)

const (
	// Poll defines how regularly to poll kubernetes resources.
	Poll = 2 * time.Second
)

var (
	kubeconfig      string // path to kubeconfig file
	deleteNamespace bool   // whether or not to delete test namespaces
	cloudConfigFile string // path to cloud provider config file
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to Kubeconfig file with authorization and master location information.")
	flag.BoolVar(&deleteNamespace, "delete-namespace", true, "If true tests will delete namespace after completion. It is only designed to make debugging easier, DO NOT turn it off by default.")
	flag.StringVar(&cloudConfigFile, "cloud-config", "", "The path to the cloud provider configuration file. Empty string for no configuration file.")
}

// Framework is used in the execution of e2e tests.
type Framework struct {
	BaseName string

	InitCloudProvider bool                    // Whether to initialise a cloud provider interface for testing
	CloudProvider     cloudprovider.Interface // Every test has a cloud provider unless initialisation is skipped

	ClientSet         clientset.Interface
	InternalClientset internalclientset.Interface

	SkipNamespaceCreation bool            // Whether to skip creating a namespace
	Namespace             *v1.Namespace   // Every test has at least one namespace unless creation is skipped
	namespacesToDelete    []*v1.Namespace // Some tests have more than one.

	// To make sure that this framework cleans up after itself, no matter what,
	// we install a Cleanup action before each test and clear it after.  If we
	// should abort, the AfterSuite hook should run all Cleanup actions.
	cleanupHandle CleanupActionHandle
}

// NewDefaultFramework constructs a new e2e test Framework with default options.
func NewDefaultFramework(baseName string) *Framework {
	f := NewFramework(baseName, nil)
	return f
}

// NewFrameworkWithCloudProvider constructs a new e2e test Framework for testing
// cloudprovider.Interface directly.
func NewFrameworkWithCloudProvider(baseName string) *Framework {
	f := NewFramework(baseName, nil)
	f.SkipNamespaceCreation = true
	return f
}

// NewFramework constructs a new e2e test Framework.
func NewFramework(baseName string, client clientset.Interface) *Framework {
	f := &Framework{
		BaseName:  baseName,
		ClientSet: client,
	}

	BeforeEach(f.BeforeEach)
	AfterEach(f.AfterEach)

	return f
}

// CreateNamespace creates a e2e test namespace.
func (f *Framework) CreateNamespace(baseName string, labels map[string]string) (*v1.Namespace, error) {
	if labels == nil {
		labels = map[string]string{}
	}

	namespaceObj := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("ccm-e2e-tests-%v-", baseName),
			Namespace:    "",
			Labels:       labels,
		},
		Status: v1.NamespaceStatus{},
	}

	// Be robust about making the namespace creation call.
	var got *v1.Namespace
	if err := wait.PollImmediate(Poll, 30*time.Second, func() (bool, error) {
		var err error
		got, err = f.ClientSet.CoreV1().Namespaces().Create(namespaceObj)
		if err != nil {
			Logf("Unexpected error while creating namespace: %v", err)
			return false, nil
		}
		return true, nil
	}); err != nil {
		return nil, err
	}

	if got != nil {
		f.namespacesToDelete = append(f.namespacesToDelete, got)
	}

	return got, nil
}

// DeleteNamespace deletes a given namespace and waits until its contents are
// deleted.
func (f *Framework) DeleteNamespace(namespace string, timeout time.Duration) error {
	startTime := time.Now()
	if err := f.ClientSet.CoreV1().Namespaces().Delete(namespace, nil); err != nil {
		if apierrors.IsNotFound(err) {
			Logf("Namespace %v was already deleted", namespace)
			return nil
		}
		return err
	}

	// wait for namespace to delete or timeout.
	err := wait.PollImmediate(Poll, timeout, func() (bool, error) {
		if _, err := f.ClientSet.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			Logf("Error while waiting for namespace to be terminated: %v", err)
			return false, nil
		}
		return false, nil
	})

	// Namespace deletion timed out.
	if err != nil {
		return fmt.Errorf("namespace %v was not deleted with limit: %v", namespace, err)
	}

	Logf("namespace %v deletion completed in %s", namespace, time.Now().Sub(startTime))
	return nil
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEach() {
	// The fact that we need this feels like a bug in ginkgo.
	// https://github.com/onsi/ginkgo/issues/222
	f.cleanupHandle = AddCleanupAction(f.AfterEach)

	if f.ClientSet == nil {
		By("Creating a kubernetes client")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		Expect(err).NotTo(HaveOccurred())
		f.ClientSet, err = clientset.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())
		f.InternalClientset, err = internalclientset.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())
	}

	if f.InitCloudProvider {
		cloud, err := cloudprovider.InitCloudProvider(oci.ProviderName(), cloudConfigFile)
		Expect(err).NotTo(HaveOccurred())
		f.CloudProvider = cloud
	}

	if !f.SkipNamespaceCreation {
		By("Building a namespace api object")
		namespace, err := f.CreateNamespace(f.BaseName, map[string]string{
			"e2e-framework": f.BaseName,
		})
		Expect(err).NotTo(HaveOccurred())
		f.Namespace = namespace
	}
}

// AfterEach deletes the namespace(s).
func (f *Framework) AfterEach() {
	RemoveCleanupAction(f.cleanupHandle)

	nsDeletionErrors := map[string]error{}
	if deleteNamespace {
		for _, ns := range f.namespacesToDelete {
			By(fmt.Sprintf("Destroying namespace %q for this suite.", ns.Name))
			if err := f.DeleteNamespace(ns.Name, 5*time.Minute); err != nil {
				nsDeletionErrors[ns.Name] = err
			}
		}
	}

	// if we had errors deleting, report them now.
	if len(nsDeletionErrors) != 0 {
		messages := []string{}
		for namespaceKey, namespaceErr := range nsDeletionErrors {
			messages = append(messages, fmt.Sprintf("Couldn't delete ns: %q: %s (%#v)", namespaceKey, namespaceErr, namespaceErr))
		}
		Failf(strings.Join(messages, ","))
	}
}
