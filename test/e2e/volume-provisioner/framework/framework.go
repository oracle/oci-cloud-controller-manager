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
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	ocicore "github.com/oracle/oci-go-sdk/core"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	utilpointer "k8s.io/kubernetes/pkg/util/pointer"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
)

const (
	OCIConfigVar       = "OCICONFIG_VAR"
	KubeConfigVar      = "KUBECONFIG_VAR"
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

	config, err := providercfg.FromFile(TestContext.OCIConfig)
	if err != nil {
		Failf("Unable to load configuration file: %v", TestContext.OCIConfig)
	}

	provider, err := providercfg.NewConfigurationProvider(config)
	if err != nil {
		Failf("Unable to create configuration provider %v", err)
	}

	blockStorageClient, err := ocicore.NewBlockstorageClientWithConfigurationProvider(provider)
	if err != nil {
		Failf("Unable to load volume provisioner client %v", err)
	}

	return blockStorageClient
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
	blockDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oci-block-volume-provisioner",
			Namespace: "kube-system",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: utilpointer.Int32Ptr(int32(1)),
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
			Replicas: utilpointer.Int32Ptr(int32(1)),
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

	err := sharedfw.CreateAndAwaitDeployment(client, blockDeployment)
	if err != nil {
		return errors.Wrap(err, "deploying block volume provisioner")
	}

	err = sharedfw.CreateAndAwaitDeployment(client, fssDeployment)
	if err != nil {
		return errors.Wrap(err, "deploying fss volume provisioner")
	}

	return nil
}

// DeleteVolumeProvisioner deletes both the block and fss volume provisioners.
func DeleteVolumeProvisioner(client clientset.Interface) {
	Logf("Deleting oci-block-volume-provisioner Deployment")

	err := client.AppsV1().Deployments("kube-system").Delete("oci-block-volume-provisioner", nil)
	if err != nil && !apierrors.IsNotFound(err) {
		Logf("Error deleting oci-block-volume-provisioner: %+v", err)
	}

	Logf("Deleting oci-file-system-volume-provisioner Deployment")

	err = client.AppsV1().Deployments("kube-system").Delete("oci-file-system-volume-provisioner", nil)
	if err != nil && !apierrors.IsNotFound(err) {
		Logf("Error deleting oci-file-system-volume-provisioner: %+v", err)
	}
}

// InstallFlexvolumeDriver installs the flexvolume driver and waits for it to be
// ready.
func InstallFlexvolumeDriver(client clientset.Interface) error {
	hostPathDirectoryOrCreate := v1.HostPathDirectoryOrCreate
	masterDS := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oci-flexvolume-driver-master",
			Namespace: "kube-system",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "oci-flexvolume-driver-master"},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "oci-flexvolume-driver-master",
					Labels: map[string]string{"app": "oci-flexvolume-driver-master"},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "oci-flexvolume-driver",
					NodeSelector:       map[string]string{"node-role.kubernetes.io/master": ""},
					Tolerations: []v1.Toleration{{
						Key:    "node.cloudprovider.kubernetes.io/uninitialized",
						Value:  "true",
						Effect: v1.TaintEffectNoSchedule,
					}, {
						Key:      "node-role.kubernetes.io/master",
						Operator: v1.TolerationOpExists,
						Effect:   v1.TaintEffectNoSchedule,
					}},
					Containers: []v1.Container{{
						Name:    "oci-flexvolume-driver",
						Image:   TestContext.Image,
						Command: []string{"/usr/local/bin/install.py", "-c", "/tmp/config.yaml"},
						SecurityContext: &v1.SecurityContext{
							Privileged: utilpointer.BoolPtr(true),
						},
						VolumeMounts: []v1.VolumeMount{{
							Name:      "flexvolume-mount",
							MountPath: "/flexmnt",
						}, {
							Name:      "config",
							MountPath: "/tmp",
							ReadOnly:  true,
						}},
					}},
					Volumes: []v1.Volume{{
						Name: "config",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: "oci-flexvolume-driver",
							},
						},
					}, {
						Name: "kubeconfig",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: "oci-flexvolume-driver-kubeconfig",
							},
						},
					}, {
						Name: "flexvolume-mount",
						VolumeSource: v1.VolumeSource{
							HostPath: &v1.HostPathVolumeSource{
								Path: "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/",
								Type: &hostPathDirectoryOrCreate,
							},
						},
					}},
				},
			},
		},
	}

	workerDS := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "oci-flexvolume-driver-worker",
			Namespace: "kube-system",
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "oci-flexvolume-driver-worker"},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "oci-flexvolume-driver-worker",
					Labels: map[string]string{"app": "oci-flexvolume-driver-worker"},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "oci-flexvolume-driver",
					Containers: []v1.Container{{
						Name:    "oci-flexvolume-driver",
						Image:   TestContext.Image,
						Command: []string{"/usr/local/bin/install.py"},
						SecurityContext: &v1.SecurityContext{
							Privileged: utilpointer.BoolPtr(true),
						},
						VolumeMounts: []v1.VolumeMount{{
							Name:      "flexvolume-mount",
							MountPath: "/flexmnt",
						}},
					}},
					Volumes: []v1.Volume{{
						Name: "flexvolume-mount",
						VolumeSource: v1.VolumeSource{
							HostPath: &v1.HostPathVolumeSource{
								Path: "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/",
								Type: &hostPathDirectoryOrCreate,
							},
						},
					}},
				},
			},
		},
	}

	if err := sharedfw.CreateAndAwaitDaemonSet(client, masterDS); err != nil {
		return errors.Wrap(err, "installing flexvolume driver master DaemonSet")
	}

	if err := sharedfw.CreateAndAwaitDaemonSet(client, workerDS); err != nil {
		return errors.Wrap(err, "installing flexvolume driver worker DaemonSet")
	}

	return nil
}

// DeleteFlexvolumeDriver deletes both the master and worker flexvolume driver
// DaemonSets.
func DeleteFlexvolumeDriver(client clientset.Interface) {
	Logf("Deleting oci-flexvolume-driver-master DaemonSet")

	err := client.AppsV1().DaemonSets("kube-system").Delete("oci-flexvolume-driver-master", nil)
	if err != nil && !apierrors.IsNotFound(err) {
		Logf("Error deleting oci-flexvolume-driver-master: %+v", err)
	}

	Logf("Deleting oci-flexvolume-driver-worker DaemonSet")

	err = client.AppsV1().DaemonSets("kube-system").Delete("oci-flexvolume-driver-worker", nil)
	if err != nil && !apierrors.IsNotFound(err) {
		Logf("Error deleting oci-flexvolume-driver-worker: %+v", err)
	}
}
