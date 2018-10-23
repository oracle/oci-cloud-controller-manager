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
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci" // register oci cloud provider
	client "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	common "github.com/oracle/oci-go-sdk/common"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	wait "k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

const (
	// Poll defines how regularly to poll kubernetes resources.
	Poll = 2 * time.Second
)

var (
	kubeconfig      string // path to kubeconfig file
	deleteNamespace bool   // whether or not to delete test namespaces
	cloudConfigFile string // path to cloud provider config file
	nodePortTest    bool   // whether or not to test the connectivity of node ports.
	ccmSeclistID    string // The ocid of the loadbalancer subnet seclist. Optional.
	k8sSeclistID    string // The ocid of the k8s worker subnet seclist. Optional.
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to Kubeconfig file with authorization and master location information.")
	flag.BoolVar(&deleteNamespace, "delete-namespace", true, "If true tests will delete namespace after completion. It is only designed to make debugging easier, DO NOT turn it off by default.")
	flag.StringVar(&cloudConfigFile, "cloud-config", "", "The path to the cloud provider configuration file. Empty string for no configuration file.")
	flag.BoolVar(&nodePortTest, "nodeport-test", false, "If true test will include 'nodePort' connectectivity tests.")
	flag.StringVar(&ccmSeclistID, "ccm-seclist-id", "", "The ocid of the loadbalancer subnet seclist. Enables additional seclist rule tests. If specified the 'k8s-seclist-id parameter' is also required.")
	flag.StringVar(&k8sSeclistID, "k8s-seclist-id", "", "The ocid of the k8s worker subnet seclist. Enables additional seclist rule tests. If specified the 'ccm-seclist-id parameter' is also required.")
}

// Framework is used in the execution of e2e tests.
type Framework struct {
	BaseName string

	InitCloudProvider bool                    // Whether to initialise a cloud provider interface for testing
	CloudProvider     cloudprovider.Interface // Every test has a cloud provider unless initialisation is skipped

	ClientSet         clientset.Interface
	InternalClientset internalclientset.Interface

	CloudProviderConfig *oci.Config      // If specified, the CloudProviderConfig. This provides information on the configuration of the test cluster.
	Client              client.Interface // An OCI client for checking the state of any provisioned OCI infrastructure during testing.
	NodePortTest        bool             // An optional configuration for E2E testing. If set to true, then will run additional E2E nodePort connectivity checks during testing.
	CCMSecListID        string           // An optional configuration for E2E testing. If present can be used to run additional checks against seclist during testing.
	K8SSecListID        string           // An optional configuration for E2E testing. If present can be used to run additional checks against seclist during testing.

	SkipNamespaceCreation bool            // Whether to skip creating a namespace
	Namespace             *v1.Namespace   // Every test has at least one namespace unless creation is skipped
	Secret                *v1.Secret      // Every test has at least one namespace unless creation is skipped
	namespacesToDelete    []*v1.Namespace // Some tests have more than one.

	// To make sure that this framework cleans up after itself, no matter what,
	// we install a Cleanup action before each test and clear it after.  If we
	// should abort, the AfterSuite hook should run all Cleanup actions.
	//
	// NB: This can fail from the CI when external temrination (e.g. timeouts) occur.
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
	f.InitCloudProvider = true
	return f
}

// NewFramework constructs a new e2e test Framework.
func NewFramework(baseName string, client clientset.Interface) *Framework {
	f := &Framework{
		BaseName:  baseName,
		ClientSet: client,
	}
	// Dev/CI only configuration. Enable NodePort tests.
	npt, err := strconv.ParseBool(os.Getenv("NODEPORT_TEST"))
	if err != nil {
		f.NodePortTest = false
	} else {
		f.NodePortTest = npt
	}
	// Dev/CI only configuration. The seclist for CCM load-balancer routes.
	f.CCMSecListID = os.Getenv("CCM_SECLIST_ID")
	if ccmSeclistID != "" {
		f.CCMSecListID = ccmSeclistID // Commandline override.
	}
	// Dev/CI only configuration. The seclist for K8S worker node routes.
	f.K8SSecListID = os.Getenv("K8S_SECLIST_ID")
	if k8sSeclistID != "" {
		f.K8SSecListID = k8sSeclistID // Commandline override.
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

	// Create an OCI client if the cloudConfig has been specified.
	if cloudConfigFile != "" && f.Client == nil {
		By("Creating OCI client")
		cloudProviderConfig, err := createCloudProviderConfig(cloudConfigFile)
		Expect(err).NotTo(HaveOccurred())
		f.CloudProviderConfig = cloudProviderConfig
		ociClient, err := createOCIClient(cloudProviderConfig)
		Expect(err).NotTo(HaveOccurred())
		f.Client = ociClient
	}

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

// createCloudProviderConfig unmarshalls the CCM's cloud provider config from
// the specified location so it can be used for testing.
func createCloudProviderConfig(cloudConfigFile string) (*oci.Config, error) {
	file, err := os.Open(cloudConfigFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't open cloud provider configuration: %s.", cloudConfigFile)
	}
	defer file.Close()
	cloudProviderConfig, err := oci.ReadConfig(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't create cloud provider configuration: %s.", cloudConfigFile)
	}
	return cloudProviderConfig, nil
}

// createOCIClient creates an OCI client derived from the CCM's cloud provider config file.
func createOCIClient(cloudProviderConfig *oci.Config) (client.Interface, error) {
	cpc := cloudProviderConfig.Auth
	ociClientConfig := common.NewRawConfigurationProvider(cpc.TenancyID, cpc.UserID, cpc.Region, cpc.Fingerprint, cpc.PrivateKey, &cpc.PrivateKeyPassphrase)
	logger := zap.L()
	rateLimiter := oci.NewRateLimiter(logger.Sugar(), cloudProviderConfig.RateLimiter)
	ociClient, err := client.New(logger.Sugar(), ociClientConfig, &rateLimiter)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't create oci client from configuration: %s.", cloudConfigFile)
	}
	return ociClient, nil
}
