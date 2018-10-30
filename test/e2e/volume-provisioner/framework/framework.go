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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	ocicore "github.com/oracle/oci-go-sdk/core"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	OCIConfigVar       = "OCICONFIG_VAR"
	KubeConfigVar      = "KUBECONFIG_VAR"
	SubnetOCID         = "SUBNET_OCID"
	MntTargetOCID      = "MNT_TARGET_OCID"
	AD                 = "AD"
	KubeSystemNS       = "kube-system"
	ClassOCI           = "oci"
	ClassOCIExt3       = "oci-ext3"
	ClassOCINoParamFss = "oci-fss-noparam"
	ClassOCIMntFss     = "oci-fss-mnt"
	ClassOCISubnetFss  = "oci-fss-subnet"
	MinVolumeBlock     = "50Gi"
	VolumeFss          = "1Gi"
	FSSProv            = "oci-volume-provisioner-fss"
	OCIProv            = "oci-volume-provisioner"
	SecretNameDefault  = "oci-volume-provisioner"
)

// Framework is used in the execution of e2e tests.
type Framework struct {
	BaseName string

	ClientSet clientset.Interface

	BlockStorageClient ocicore.BlockstorageClient
	IsBackup           bool
	BackupIDs          []string
	StorageClasses     []string

	Namespace          *v1.Namespace   // Every test has at least one namespace unless creation is skipped
	namespacesToDelete []*v1.Namespace // Some tests have more than one.

	// To make sure that this framework cleans up after itself, no matter what,
	// we install a Cleanup action before each test and clear it after.  If we
	// should abort, the AfterSuite hook should run all Cleanup actions.
	cleanupHandle CleanupActionHandle
}

// NewDefaultFramework constructs a new e2e test Framework with default options.
func NewDefaultFramework(baseName string) *Framework {
	f := NewFramework(baseName, nil, false)
	return f
}

// NewFramework constructs a new e2e test Framework.
func NewFramework(baseName string, client clientset.Interface, backup bool) *Framework {
	f := &Framework{
		BaseName:  baseName,
		ClientSet: client,
		IsBackup:  backup,
	}

	BeforeEach(f.BeforeEach)
	AfterEach(f.AfterEach)

	return f
}

// NewBackupFramework constructs a new e2e test Framework initialising a storage client used to create a backup
func NewBackupFramework(baseName string) *Framework {
	f := NewFramework(baseName, nil, true)
	return f
}

// CreateNamespace creates a e2e test namespace.
func (f *Framework) CreateNamespace(baseName string, labels map[string]string) (*v1.Namespace, error) {
	if labels == nil {
		labels = map[string]string{}
	}

	namespaceObj := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("volume-provisioner-e2e-tests-%v-", baseName),
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

	Logf("Namespace %v deletion completed in %s", namespace, time.Now().Sub(startTime))
	return nil
}

// BeforeEach gets a client and makes a namespace.
func (f *Framework) BeforeEach() {
	// The fact that we need this feels like a bug in ginkgo.
	// https://github.com/onsi/ginkgo/issues/222
	f.cleanupHandle = AddCleanupAction(f.AfterEach)

	if f.ClientSet == nil {
		By("Creating a kubernetes client")
		config, err := clientcmd.BuildConfigFromFlags("", TestContext.KubeConfig)
		Expect(err).NotTo(HaveOccurred())
		f.ClientSet, err = clientset.NewForConfig(config)
		Expect(err).NotTo(HaveOccurred())
	}

	if TestContext.Namespace == "" {
		By("Building a namespace api object")
		namespace, err := f.CreateNamespace(f.BaseName, map[string]string{
			"e2e-framework": f.BaseName,
		})
		Expect(err).NotTo(HaveOccurred())
		f.Namespace = namespace
	} else {
		By(fmt.Sprintf("Getting existing namespace %q", TestContext.Namespace))
		namespace, err := f.ClientSet.CoreV1().Namespaces().Get(TestContext.Namespace, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		f.Namespace = namespace
	}

	if f.IsBackup {
		f.BlockStorageClient = f.createStorageClient()
	}
}

// AfterEach deletes the namespace(s).
func (f *Framework) AfterEach() {
	RemoveCleanupAction(f.cleanupHandle)

	nsDeletionErrors := map[string]error{}

	// Whether to delete namespace is determined by 3 factors: delete-namespace flag, delete-namespace-on-failure flag and the test result
	// if delete-namespace set to false, namespace will always be preserved.
	// if delete-namespace is true and delete-namespace-on-failure is false, namespace will be preserved if test failed.
	if TestContext.DeleteNamespace && (TestContext.DeleteNamespaceOnFailure || !CurrentGinkgoTestDescription().Failed) {
		for _, ns := range f.namespacesToDelete {
			By(fmt.Sprintf("Destroying namespace %q for this suite.", ns.Name))
			if err := f.DeleteNamespace(ns.Name, 5*time.Minute); err != nil {
				nsDeletionErrors[ns.Name] = err
			}
		}
	}

	for _, storageClass := range f.StorageClasses {
		By(fmt.Sprintf("Deleting storage class %q", storageClass))
		err := f.ClientSet.StorageV1beta1().StorageClasses().Delete(storageClass, nil)
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

	// if we had errors deleting, report them now.
	if len(nsDeletionErrors) != 0 {
		messages := []string{}
		for namespaceKey, namespaceErr := range nsDeletionErrors {
			messages = append(messages, fmt.Sprintf("Couldn't delete ns: %q: %s (%#v)", namespaceKey, namespaceErr, namespaceErr))
		}
		Failf(strings.Join(messages, ","))
	}
}

func (f *Framework) createStorageClient() ocicore.BlockstorageClient {
	By("Creating an OCI block storage client")
	configPath := TestContext.OCIConfig

	file, err := os.Open(configPath)
	if err != nil {
		Failf("Unable to load volume provisioner configuration file: %v", configPath)
	}
	defer file.Close()
	cfg, err := core.LoadConfig(file)
	if err != nil {
		Failf("Unable to load volume provisioner configuration file %q: %v", file, err)
	}
	config, err := f.newConfigurationProvider(cfg)
	if err != nil {
		Failf("Unable to load volume provisioner configuration file %q: %v", cfg, err)
	}
	blockStorageClient, err := ocicore.NewBlockstorageClientWithConfigurationProvider(config)
	if err != nil {
		Logf("Unable to load volume provisioner client %q: %v", config, err)
	}
	return blockStorageClient
}

func (f *Framework) newConfigurationProvider(cfg *core.Config) (common.ConfigurationProvider, error) {
	var conf common.ConfigurationProvider
	if cfg != nil {
		err := cfg.Validate()
		if err != nil {
			return nil, errors.Wrap(err, "invalid client config")
		}
		if cfg.UseInstancePrincipals {
			Logf("Using instance principals configuration provider")
			cp, err := auth.InstancePrincipalConfigurationProvider()
			if err != nil {
				return nil, errors.Wrap(err, "InstancePrincipalConfigurationProvider")
			}
			return cp, nil
		}
		Logf("Using raw configuration provider")
		conf = common.NewRawConfigurationProvider(
			cfg.Auth.TenancyID,
			cfg.Auth.UserID,
			cfg.Auth.Region,
			cfg.Auth.Fingerprint,
			cfg.Auth.PrivateKey,
			common.String(cfg.Auth.PrivateKeyPassphrase))
	} else {
		conf = common.DefaultConfigProvider()
	}
	return conf, nil
}

// NewClientSetFromFlags builds a kubernetes client from flags.
func NewClientSetFromFlags() (clientset.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", TestContext.KubeConfig)
	if err != nil {
		return nil, err
	}
	cs, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// InstallVolumeProvisioner installs both block and fss volume provisioners and
// waits for them to be ready.
func InstallVolumeProvisioner(client clientset.Interface) error {
	replica := int32(1)

	blockDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oci-block-volume-provisioner",
			Namespace: "kube-system",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "oci-block-volume-provisioner"},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "oci-block-volume-provisioner"},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "oci-volume-provisioner",
					Containers: []v1.Container{{
						Name:            "oci-block-volume-provisioner",
						Image:           TestContext.Image,
						Command:         []string{"/usr/local/bin/oci-volume-provisioner"},
						ImagePullPolicy: v1.PullAlways,
						Env: []v1.EnvVar{{
							Name: "NODE_NAME",
							ValueFrom: &v1.EnvVarSource{
								FieldRef: &v1.ObjectFieldSelector{
									FieldPath: "spec.nodeName",
								},
							},
						}, {
							Name:  "PROVISIONER_TYPE",
							Value: core.ProvisionerNameDefault,
						}},
						VolumeMounts: []v1.VolumeMount{{
							Name:      "config",
							MountPath: "/etc/oci/",
							ReadOnly:  true,
						}},
					},
					},
					Volumes: []v1.Volume{{
						Name: "config",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: "oci-volume-provisioner",
							},
						},
					}},
				},
			},
		},
	}

	// TODO(apryde): Decide whether we're adding --controllers="block,filesystem"
	// to run both provisioners from the same binary and if not dedup this
	// code. Otherwise it won't be needed :D
	fssDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oci-file-system-volume-provisioner",
			Namespace: "kube-system",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "oci-file-system-volume-provisioner"},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "oci-file-system-volume-provisioner"},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "oci-volume-provisioner",
					Containers: []v1.Container{{
						Name:            "oci-file-system-volume-provisioner",
						Image:           TestContext.Image,
						Command:         []string{"/usr/local/bin/oci-volume-provisioner"},
						ImagePullPolicy: v1.PullAlways,
						Env: []v1.EnvVar{{
							Name: "NODE_NAME",
							ValueFrom: &v1.EnvVarSource{
								FieldRef: &v1.ObjectFieldSelector{
									FieldPath: "spec.nodeName",
								},
							},
						}, {
							Name:  "PROVISIONER_TYPE",
							Value: core.ProvisionerNameFss,
						}},
						VolumeMounts: []v1.VolumeMount{{
							Name:      "config",
							MountPath: "/etc/oci/",
							ReadOnly:  true,
						}},
					},
					},
					Volumes: []v1.Volume{{
						Name: "config",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: "oci-volume-provisioner",
							},
						},
					}},
				},
			},
		},
	}

	err := createAndAwaitDeployment(client, blockDeployment)
	if err != nil {
		return errors.Wrap(err, "deploying block volume provisioner")
	}

	err = createAndAwaitDeployment(client, fssDeployment)
	if err != nil {
		return errors.Wrap(err, "deploying fss volume provisioner")
	}

	return nil
}

func createAndAwaitDeployment(client clientset.Interface, desired *appsv1.Deployment) error {
	actual, err := client.AppsV1().Deployments("kube-system").Create(desired)
	if err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "failed to create %q Deployment", desired.Name)
		}
		Logf("Provisioner already exists. Updating.")
		actual, err = client.AppsV1().Deployments("kube-system").Update(desired)
		if err != nil {
			return errors.Wrapf(err, "updating volume provisioner Deployment %q", desired.Name)
		}
	} else {
		Logf("Created Deployment %q in namespace %q", actual.Name, actual.Namespace)
	}

	return wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		actual, err := client.AppsV1().Deployments(actual.Namespace).Get(actual.Name, metav1.GetOptions{})
		if err != nil {
			return false, errors.Wrap(err, "waiting for Deployment to be ready")
		}
		if actual.Status.Replicas != 0 && actual.Status.Replicas == actual.Status.ReadyReplicas {
			return true, nil
		}
		Logf("%s Deployment not yet ready (replicas=%d, readyReplicas=%d). Waiting...",
			actual.Name, actual.Status.Replicas, actual.Status.ReadyReplicas)
		return false, nil
	})
}

// DeleteVolumeProvisioner deletes both the block and fss volume provisioners.
func DeleteVolumeProvisioner(client clientset.Interface) error {
	Logf("Deleteing oci-block-volume-provisioner Deployment")

	// TODO(apryde): We can probably use a label selector to delete both at
	// once as this currently leaves dangling resources if the first delete
	// fails.
	err := client.AppsV1().Deployments("kube-system").Delete("oci-block-volume-provisioner", nil)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "deleting oci-block-volume-provisioner")
	}

	Logf("Deleteing oci-file-system-volume-provisioner Deployment")

	err = client.AppsV1().Deployments("kube-system").Delete("oci-file-system-volume-provisioner", nil)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "deleting oci-file-system-volume-provisioner")
	}
	return nil
}
