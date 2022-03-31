// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	ocicore "github.com/oracle/oci-go-sdk/v50/core"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci" // register oci cloud provider
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/core"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	cloudprovider "k8s.io/cloud-provider"
)

// CloudProviderFramework is used in the execution of e2e tests.
type CloudProviderFramework struct {
	BaseName string

	InitCloudProvider bool                    // Whether to initialise a cloud provider interface for testing
	CloudProvider     cloudprovider.Interface // Every test has a cloud provider unless initialisation is skipped

	ClientSet clientset.Interface

	CloudProviderConfig *providercfg.Config // If specified, the CloudProviderConfig. This provides information on the configuration of the test cluster.
	Client              client.Interface    // An OCI client for checking the state of any provisioned OCI infrastructure during testing.
	NodePortTest        bool                // An optional configuration for E2E testing. If set to true, then will run additional E2E nodePort connectivity checks during testing.
	CCMSecListID        string              // An optional configuration for E2E testing. If present can be used to run additional checks against seclist during testing.
	K8SSecListID        string              // An optional configuration for E2E testing. If present can be used to run additional checks against seclist during testing.

	SkipNamespaceCreation bool            // Whether to skip creating a namespace
	Namespace             *v1.Namespace   // Every test has at least one namespace unless creation is skipped
	Secret                *v1.Secret      // Every test has at least one namespace unless creation is skipped
	namespacesToDelete    []*v1.Namespace // Some tests have more than one.

	BlockStorageClient ocicore.BlockstorageClient
	IsBackup           bool
	BackupIDs          []string
	StorageClasses     []string
	VolumeIds          []string

	// To make sure that this framework cleans up after itself, no matter what,
	// we install a Cleanup action before each test and clear it after.  If we
	// should abort, the AfterSuite hook should run all Cleanup actions.
	//
	// NB: This can fail from the CI when external temrination (e.g. timeouts) occur.
	cleanupHandle CleanupActionHandle
}

// NewDefaultFramework constructs a new e2e test CloudProviderFramework with default options.
func NewDefaultFramework(baseName string) *CloudProviderFramework {
	f := NewCcmFramework(baseName, nil, false)
	return f
}

// NewFrameworkWithCloudProvider constructs a new e2e test CloudProviderFramework for testing
// cloudprovider.Interface directly.
func NewFrameworkWithCloudProvider(baseName string) *CloudProviderFramework {
	f := NewCcmFramework(baseName, nil, false)
	f.SkipNamespaceCreation = true
	f.InitCloudProvider = true
	return f
}

// NewBackupFramework constructs a new e2e test Framework initialising a storage client used to create a backup
func NewBackupFramework(baseName string) *CloudProviderFramework {
	f := NewCcmFramework(baseName, nil, true)
	return f
}

// NewCcmFramework constructs a new e2e test CloudProviderFramework.
func NewCcmFramework(baseName string, client clientset.Interface, backup bool) *CloudProviderFramework {
	f := &CloudProviderFramework{
		BaseName:  baseName,
		ClientSet: client,
		IsBackup:  backup,
	}
	f.NodePortTest = nodePortTest
	if ccmSeclistID != "" {
		f.CCMSecListID = ccmSeclistID
	}
	if k8sSeclistID != "" {
		f.K8SSecListID = k8sSeclistID
	}
	BeforeEach(f.BeforeEach)
	AfterEach(f.AfterEach)

	return f
}

// CreateNamespace creates a e2e test namespace.
func (f *CloudProviderFramework) CreateNamespace(baseName string, labels map[string]string) (*v1.Namespace, error) {
	if labels == nil {
		labels = map[string]string{}
	}

	namespaceObj := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("cloud-provider-e2e-tests-%v-", baseName),
			Namespace:    "",
			Labels:       labels,
		},
		Status: v1.NamespaceStatus{},
	}

	// Be robust about making the namespace creation call.
	var got *v1.Namespace
	if err := wait.PollImmediate(K8sResourcePoll, 30*time.Second, func() (bool, error) {
		var err error
		got, err = f.ClientSet.CoreV1().Namespaces().Create(context.Background(), namespaceObj, metav1.CreateOptions{})
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
func (f *CloudProviderFramework) DeleteNamespace(namespace string, timeout time.Duration) error {
	startTime := time.Now()
	if err := f.ClientSet.CoreV1().Namespaces().Delete(context.Background(), namespace, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			Logf("Namespace %v was already deleted", namespace)
			return nil
		}
		return err
	}

	// wait for namespace to delete or timeout.
	err := wait.PollImmediate(K8sResourcePoll, timeout, func() (bool, error) {
		if _, err := f.ClientSet.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{}); err != nil {
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
func (f *CloudProviderFramework) BeforeEach() {
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
		config, err := clientcmd.BuildConfigFromFlags("", clusterkubeconfig)
		Expect(err).NotTo(HaveOccurred())
		f.ClientSet, err = clientset.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())
	}

	if f.InitCloudProvider {
		cloud, err := cloudprovider.InitCloudProvider(oci.ProviderName(), cloudConfigFile)
		Expect(err).NotTo(HaveOccurred())
		ccmProvider := cloud.(*oci.CloudProvider)
		factory := informers.NewSharedInformerFactory(f.ClientSet, 5*time.Minute)

		nodeInfoController := oci.NewNodeInfoController(
			factory.Core().V1().Nodes(),
			f.ClientSet,
			ccmProvider,
			zap.L().Sugar(),
			cache.NewTTLStore(instanceCacheKeyFn, time.Duration(24)*time.Hour),
			f.Client)
		nodeInformer := factory.Core().V1().Nodes()
		go nodeInformer.Informer().Run(wait.NeverStop)
		go nodeInfoController.Run(wait.NeverStop)
		if !cache.WaitForCacheSync(wait.NeverStop, nodeInformer.Informer().HasSynced) {
			utilruntime.HandleError(fmt.Errorf("Timed out waiting for informers to sync"))
		}
		ccmProvider.NodeLister = nodeInformer.Lister()
		f.CloudProvider = ccmProvider
	}

	if !f.SkipNamespaceCreation {
		By("Building a namespace api object")
		namespace, err := f.CreateNamespace(f.BaseName, map[string]string{
			"e2e-framework": f.BaseName,
		})
		Expect(err).NotTo(HaveOccurred())
		f.Namespace = namespace
	}

	if f.IsBackup {
		f.BlockStorageClient = f.createStorageClient()
	}
}

// AfterEach deletes the namespace(s).
func (f *CloudProviderFramework) AfterEach() {
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

	for _, storageClass := range f.StorageClasses {
		By(fmt.Sprintf("Deleting storage class %q", storageClass))
		err := f.ClientSet.StorageV1().StorageClasses().Delete(context.Background(), storageClass, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			Logf("Storage Class Delete API error: %v", err)
		}
	}

	for _, backupID := range f.BackupIDs {
		By(fmt.Sprintf("Deleting backups %q", backupID))
		ctx := context.TODO()
		_, err := f.BlockStorageClient.DeleteVolumeBackup(ctx, ocicore.DeleteVolumeBackupRequest{VolumeBackupId: &backupID})
		if err != nil && !apierrors.IsNotFound(err) {
			Logf("Failed to delete backup id %q: %v", backupID, err)
		}
	}

	for _, volId := range f.VolumeIds {
		By(fmt.Sprintf("Deleting volumes %q", volId))
		err := f.ClientSet.CoreV1().PersistentVolumes().Delete(context.Background(), volId, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			Logf("Failed to delete persistent volume %q: %v", volId, err)
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
func createCloudProviderConfig(cloudConfigFile string) (*providercfg.Config, error) {
	file, err := os.Open(cloudConfigFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't open cloud provider configuration: %s.", cloudConfigFile)
	}
	defer file.Close()
	cloudProviderConfig, err := providercfg.ReadConfig(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't create cloud provider configuration: %s.", cloudConfigFile)
	}
	return cloudProviderConfig, nil
}

// createOCIClient creates an OCI client derived from the CCM's cloud provider config file.
func createOCIClient(cloudProviderConfig *providercfg.Config) (client.Interface, error) {
	cpc := cloudProviderConfig.Auth
	ociClientConfig := common.NewRawConfigurationProvider(cpc.TenancyID, cpc.UserID, cpc.Region, cpc.Fingerprint, cpc.PrivateKey, &cpc.PrivateKeyPassphrase)
	logger := zap.L()
	rateLimiter := client.NewRateLimiter(logger.Sugar(), cloudProviderConfig.RateLimiter)
	ociClient, err := client.New(logger.Sugar(), ociClientConfig, &rateLimiter)
	if err != nil {
		return nil, errors.Wrapf(err, "Couldn't create oci client from configuration: %s.", cloudConfigFile)
	}
	return ociClient, nil
}

func (f *CloudProviderFramework) createStorageClient() ocicore.BlockstorageClient {
	By("Creating an OCI block storage client")

	provider, err := providercfg.NewConfigurationProvider(f.CloudProviderConfig)
	if err != nil {
		Failf("Unable to create configuration provider %v", err)
	}

	blockStorageClient, err := ocicore.NewBlockstorageClientWithConfigurationProvider(provider)
	if err != nil {
		Failf("Unable to load volume provisioner client %v", err)
	}

	return blockStorageClient
}

func instanceCacheKeyFn(obj interface{}) (string, error) {
	return *obj.(*core.Instance).Id, nil
}
