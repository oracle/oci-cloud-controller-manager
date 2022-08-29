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
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/csi/driver"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/plugin"
	ocicore "github.com/oracle/oci-go-sdk/v50/core"
)

const (
	KmsKey                        = "kms-key-id"
	AttachmentTypeISCSI           = "iscsi"
	AttachmentTypeParavirtualized = "paravirtualized"
	AttachmentType                = "attachment-type"
	FstypeKey                     = "csi.storage.k8s.io/fstype"
)

// PVCTestJig is a jig to help create PVC tests.
type PVCTestJig struct {
	ID                 string
	Name               string
	Labels             map[string]string
	BlockStorageClient *ocicore.BlockstorageClient
	KubeClient         clientset.Interface
	config             *restclient.Config
}

// NewPVCTestJig allocates and inits a new PVCTestJig.
func NewPVCTestJig(kubeClient clientset.Interface, name string) *PVCTestJig {
	id := string(uuid.NewUUID())
	return &PVCTestJig{
		ID:   id,
		Name: name,
		Labels: map[string]string{
			"testID":   id,
			"testName": name,
		},
		KubeClient: kubeClient,
	}
}

func (j *PVCTestJig) CreatePVCTemplate(namespace, volumeSize string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: j.Name,
			Labels:       j.Labels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(volumeSize),
				},
			},
		},
	}
}

func (j *PVCTestJig) pvcAddLabelSelector(pvc *v1.PersistentVolumeClaim, adLabel string) *v1.PersistentVolumeClaim {
	if pvc != nil {
		pvc.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				plugin.LabelZoneFailureDomain: adLabel,
			},
		}
	}
	return pvc
}

func (j *PVCTestJig) pvcAddAccessMode(pvc *v1.PersistentVolumeClaim,
	accessMode v1.PersistentVolumeAccessMode) *v1.PersistentVolumeClaim {
	if pvc != nil {
		pvc.Spec.AccessModes = []v1.PersistentVolumeAccessMode{
			accessMode,
		}
	}
	return pvc
}

func (j *PVCTestJig) pvcAddStorageClassName(pvc *v1.PersistentVolumeClaim,
	storageClassName string) *v1.PersistentVolumeClaim {
	if pvc != nil {
		pvc.Spec.StorageClassName = &storageClassName
	}
	return pvc
}

func (j *PVCTestJig) pvcAddVolumeName(pvc *v1.PersistentVolumeClaim, volumeName string) *v1.PersistentVolumeClaim {
	if pvc != nil {
		pvc.Spec.VolumeName = volumeName
	}
	return pvc
}

func (j *PVCTestJig) pvcExpandVolume(claim *v1.PersistentVolumeClaim, size string) *v1.PersistentVolumeClaim {
	oldPVC, err := j.KubeClient.CoreV1().
		PersistentVolumeClaims(claim.Namespace).
		Get(context.Background(),
			claim.Name,
			metav1.GetOptions{})
	if err != nil || oldPVC == nil {
		Failf("Error expanding the volume : %q", err)
		return nil
	}
	pvc := oldPVC.DeepCopy()
	if pvc != nil {
		pvc.Spec.Resources.Requests = v1.ResourceList{
			v1.ResourceStorage: resource.MustParse(size),
		}
	}
	return pvc
}

// newPVCTemplate returns the default template for this jig, but
// does not actually create the PVC. The default PVC has the same name
// as the jig
func (j *PVCTestJig) newPVCTemplate(namespace, volumeSize, scName, adLabel string) *v1.PersistentVolumeClaim {
	pvc := j.CreatePVCTemplate(namespace, volumeSize)
	pvc = j.pvcAddAccessMode(pvc, v1.ReadWriteOnce)
	pvc = j.pvcAddLabelSelector(pvc, adLabel)
	pvc = j.pvcAddStorageClassName(pvc, scName)
	return pvc
}

// newPVCTemplateCSI returns the default template for this jig, but
// does not actually create the PVC.  The default PVC has the same name
// as the jig
func (j *PVCTestJig) newPVCTemplateCSI(namespace string, volumeSize string, scName string) *v1.PersistentVolumeClaim {
	pvc := j.CreatePVCTemplate(namespace, volumeSize)
	pvc = j.pvcAddAccessMode(pvc, v1.ReadWriteOnce)
	pvc = j.pvcAddStorageClassName(pvc, scName)
	return pvc
}

// newPVCTemplateFSS returns the default template for this jig, but
// does not actually create the PVC.  The default PVC has the same name
// as the jig
func (j *PVCTestJig) newPVCTemplateFSS(namespace, volumeSize, volumeName string) *v1.PersistentVolumeClaim {
	pvc := j.CreatePVCTemplate(namespace, volumeSize)
	pvc = j.pvcAddAccessMode(pvc, v1.ReadWriteMany)
	pvc = j.pvcAddStorageClassName(pvc, "")
	pvc = j.pvcAddVolumeName(pvc, volumeName)
	return pvc
}

func (j *PVCTestJig) CheckPVCorFail(pvc *v1.PersistentVolumeClaim, tweak func(pvc *v1.PersistentVolumeClaim),
	namespace, volumeSize string) *v1.PersistentVolumeClaim {
	if tweak != nil {
		tweak(pvc)
	}

	name := types.NamespacedName{Namespace: namespace, Name: j.Name}
	By(fmt.Sprintf("Creating a PVC %q of volume size %q", name, volumeSize))

	result, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Create(context.Background(), pvc, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create persistent volume claim %q: %v", name, err)
	}
	return result
}

func (j *PVCTestJig) UpdatePVCorFail(pvc *v1.PersistentVolumeClaim, tweak func(pvc *v1.PersistentVolumeClaim),
	namespace, volumeSize string) *v1.PersistentVolumeClaim {
	if tweak != nil {
		tweak(pvc)
	}

	By(fmt.Sprintf("Updating a PVC %q of volume size %q", pvc.Name, volumeSize))
	newPvc := j.pvcExpandVolume(pvc, volumeSize)

	result, err := j.KubeClient.CoreV1().PersistentVolumeClaims(newPvc.Namespace).Update(context.Background(), newPvc, metav1.UpdateOptions{})
	if err != nil {
		if !apierrors.IsConflict(err) && !apierrors.IsServerTimeout(err) {
			Failf("Failed to update persistent volume claim %q: %v", newPvc.Name, err)
		}
		Failf("Error updating a PVC %q of volume size %q : %q", newPvc.Name, volumeSize, err)
	}
	return result
}

// CreatePVCorFail creates a new claim based on the jig's
// defaults. Callers can provide a function to tweak the claim object
// before it is created.
func (j *PVCTestJig) CreatePVCorFail(namespace string, volumeSize string, scName string,
	adLabel string, tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	pvc := j.newPVCTemplate(namespace, volumeSize, scName, adLabel)
	return j.CheckPVCorFail(pvc, tweak, namespace, volumeSize)
}

// CreatePVCorFailCSI creates a new claim based on the jig's
// defaults. Callers can provide a function to tweak the claim object
// before it is created.
func (j *PVCTestJig) CreatePVCorFailCSI(namespace string, volumeSize string, scName string,
	tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	pvc := j.newPVCTemplateCSI(namespace, volumeSize, scName)
	return j.CheckPVCorFail(pvc, tweak, namespace, volumeSize)
}

// CreatePVCorFailFSS creates a new claim based on the jig's
// defaults. Callers can provide a function to tweak the claim object
// before it is created.
func (j *PVCTestJig) CreatePVCorFailFSS(namespace, volumeName, volumeSize string, tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	pvc := j.newPVCTemplateFSS(namespace, volumeSize, volumeName)
	return j.CheckPVCorFail(pvc, tweak, namespace, volumeSize)
}

// UpdatePVCorFailCSI updates a new claim based on the jig's
// defaults. Callers can provide a function to tweak the claim object
// before it is updated.
func (j *PVCTestJig) UpdatePVCorFailCSI(pvc *v1.PersistentVolumeClaim, volumeSize string,
	tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	return j.UpdatePVCorFail(pvc, tweak, pvc.Namespace, volumeSize)
}
func (j *PVCTestJig) CheckAndAwaitPVCOrFail(pvc *v1.PersistentVolumeClaim, namespace string,
	pvcPhase v1.PersistentVolumeClaimPhase) *v1.PersistentVolumeClaim {
	pvc = j.waitForConditionOrFail(namespace, pvc.Name, DefaultTimeout, "to be provisioned",
		func(pvc *v1.PersistentVolumeClaim) bool {
			err := j.WaitForPVCPhase(pvcPhase, namespace, pvc.Name)
			if err != nil {
				Failf("PVC %q did not reach %v state : %v", pvc.Name, pvcPhase, err)
				return false
			}
			return true
		})
	if pvcPhase == v1.ClaimBound {
		j.SanityCheckPV(pvc)
	} else if pvcPhase == v1.ClaimPending {
		zap.S().With(pvc.Namespace).With(pvc.Name).Info("PVC is created/updated successfully.")
	} else {
		zap.S().With(pvc.Namespace).With(pvc.Name).With(pvcPhase).Info("Unexpected value for pvcPhase")
	}
	return pvc
}

// CreateAndAwaitPVCOrFail creates a new PVC based on the
// jig's defaults, waits for it to become ready, and then sanity checks it and
// its dependant resources. Callers can provide a function to tweak the
// PVC object before it is created.
func (j *PVCTestJig) CreateAndAwaitPVCOrFail(namespace, volumeSize, scName, adLabel string,
	tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	pvc := j.CreatePVCorFail(namespace, volumeSize, scName, adLabel, tweak)
	return j.CheckAndAwaitPVCOrFail(pvc, namespace, v1.ClaimBound)
}

// CreateAndAwaitPVCOrFailFSS creates a new PVC based on the
// jig's defaults, waits for it to become ready, and then sanity checks it and
// its dependant resources. Callers can provide a function to tweak the
// PVC object before it is created.
func (j *PVCTestJig) CreateAndAwaitPVCOrFailFSS(namespace, volumeName, volumeSize string, tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	pvc := j.CreatePVCorFailFSS(namespace, volumeName, volumeSize, tweak)
	return j.CheckAndAwaitPVCOrFail(pvc, namespace, v1.ClaimBound)
}

// CreateAndAwaitPVCOrFailCSI creates a new PVC based on the
// jig's defaults, waits for it to become ready, and then sanity checks it and
// its dependant resources. Callers can provide a function to tweak the
// PVC object before it is created.
func (j *PVCTestJig) CreateAndAwaitPVCOrFailCSI(namespace, volumeSize, scName string,
	tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	pvc := j.CreatePVCorFailCSI(namespace, volumeSize, scName, tweak)
	return j.CheckAndAwaitPVCOrFail(pvc, namespace, v1.ClaimPending)
}

// UpdatedAndAwaitPVCOrFailCSI updates a  PVC based on the
// jig's defaults, waits for it to become ready, and then sanity checks it and
// its dependant resources. Callers can provide a function to tweak the
// PVC object before it is created.
func (j *PVCTestJig) UpdateAndAwaitPVCOrFailCSI(pvc *v1.PersistentVolumeClaim, namespace, volumeSize string,
	tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	updatedPvc := j.UpdatePVCorFailCSI(pvc, volumeSize, tweak)
	return j.CheckAndAwaitPVCOrFail(updatedPvc, namespace, v1.ClaimBound)
}

// CreateAndAwaitStaticPVCOrFailCSI creates a new PV and PVC based on the
// jig's defaults, waits for it to become ready, and then sanity checks it and
// its dependant resources. Callers can provide a function to tweak the
// PVC object before it is created.
func (j *PVCTestJig) CreateAndAwaitStaticPVCOrFailCSI(bs ocicore.BlockstorageClient, namespace string, volumeSize string, vpusPerGB int64, scName string, adLabel string, compartmentId string, tweak func(pvc *v1.PersistentVolumeClaim)) (*v1.PersistentVolumeClaim, string) {

	volumeOcid := j.CreateVolume(bs, adLabel, compartmentId, "test-volume", vpusPerGB)

	var pv *v1.PersistentVolume
	if vpusPerGB == 20 {
		pv = j.CreatePVorFailCSIHighPerf(namespace, scName, *volumeOcid)
	} else {
		pv = j.CreatePVorFailCSI(namespace, scName, *volumeOcid)
	}
	pv = j.waitForConditionOrFailForPV(pv.Name, DefaultTimeout, "to be dynamically provisioned", func(pvc *v1.PersistentVolume) bool {
		err := j.WaitForPVPhase(v1.VolumeAvailable, pv.Name)
		if err != nil {
			Failf("PV %q did not created: %v", pv.Name, err)
			return false
		}
		return true
	})

	return j.CreateAndAwaitPVCOrFailCSI(namespace, volumeSize, scName, tweak), *volumeOcid
}

func (j *PVCTestJig) CreatePVTemplate(namespace, annotation, storageClassName string,
	pvReclaimPolicy v1.PersistentVolumeReclaimPolicy) *v1.PersistentVolume {
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: j.Name,
			Labels:       j.Labels,
			Annotations: map[string]string{
				"pv.kubernetes.io/provisioned-by": annotation,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("50Gi"),
			},
			PersistentVolumeReclaimPolicy: pvReclaimPolicy,
			StorageClassName:              storageClassName,
		},
	}
}

func (j *PVCTestJig) pvAddAccessMode(pv *v1.PersistentVolume,
	accessMode v1.PersistentVolumeAccessMode) *v1.PersistentVolume {
	if pv != nil {
		pv.Spec.AccessModes = append(pv.Spec.AccessModes, accessMode)
	}
	return pv
}

func (j *PVCTestJig) pvAddVolumeMode(pv *v1.PersistentVolume,
	volumeMode v1.PersistentVolumeMode) *v1.PersistentVolume {
	if pv != nil {
		pv.Spec.VolumeMode = &volumeMode
	}
	return pv
}

func (j *PVCTestJig) pvAddPersistentVolumeSource(pv *v1.PersistentVolume,
	pvs v1.PersistentVolumeSource) *v1.PersistentVolume {
	if pv != nil {
		pv.Spec.PersistentVolumeSource = pvs
	}
	return pv
}

// newPVTemplateFSS returns the default template for this jig, but
// does not actually create the PV.  The default PV has the same name
// as the jig
func (j *PVCTestJig) newPVTemplateFSS(namespace, volumeHandle, enableIntransitEncrypt string) *v1.PersistentVolume {
	pv := j.CreatePVTemplate(namespace, "fss.csi.oraclecloud.com", "", "Retain")
	pv = j.pvAddVolumeMode(pv, v1.PersistentVolumeFilesystem)
	pv = j.pvAddAccessMode(pv, "ReadWriteMany")
	pv = j.pvAddPersistentVolumeSource(pv, v1.PersistentVolumeSource{
		CSI: &v1.CSIPersistentVolumeSource{
			Driver:       driver.FSSDriverName,
			VolumeHandle: volumeHandle,
			VolumeAttributes: map[string]string{
				"encryptInTransit": enableIntransitEncrypt,
			},
		},
	})

	return pv
}

// newPVTemplateCSI returns the default template for this jig, but
// does not actually create the PV.  The default PV has the same name
// as the jig
func (j *PVCTestJig) newPVTemplateCSI(namespace string, scName string, ocid string) *v1.PersistentVolume {
	pv := j.CreatePVTemplate(namespace, "blockvolume.csi.oraclecloud.com", scName, "Delete")
	pv = j.pvAddAccessMode(pv, "ReadWriteOnce")
	pv = j.pvAddPersistentVolumeSource(pv, v1.PersistentVolumeSource{
		CSI: &v1.CSIPersistentVolumeSource{
			Driver:       "blockvolume.csi.oraclecloud.com",
			FSType:       "ext4",
			VolumeHandle: ocid,
		},
	})
	return pv
}

// newPVTemplateCSI returns the default template for this jig, but
// does not actually create the PV.  The default PV has the same name
// as the jig
func (j *PVCTestJig) newPVTemplateCSIHighPerf(namespace string, scName string, ocid string) *v1.PersistentVolume {
	pv := j.CreatePVTemplate(namespace, "blockvolume.csi.oraclecloud.com", scName, "Delete")
	pv = j.pvAddAccessMode(pv, "ReadWriteOnce")
	pv = j.pvAddPersistentVolumeSource(pv, v1.PersistentVolumeSource{
		CSI: &v1.CSIPersistentVolumeSource{
			Driver:       "blockvolume.csi.oraclecloud.com",
			FSType:       "ext4",
			VolumeHandle: ocid,
			VolumeAttributes: map[string]string{
				csi_util.VpusPerGB: "20",
			},
		},
	})
	return pv
}

// CreatePVForFSSorFail creates a new claim based on the jig's
// defaults. Callers can provide a function to tweak the claim object
// before it is created.
func (j *PVCTestJig) CreatePVorFailFSS(namespace, volumeHandle, encryptInTransit string) *v1.PersistentVolume {
	pv := j.newPVTemplateFSS(namespace, volumeHandle, encryptInTransit)

	result, err := j.KubeClient.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create persistent volume claim %q: %v", pv.Name, err)
	}
	return result
}

// CreatePVorFail creates a new claim based on the jig's
// defaults. Callers can provide a function to tweak the claim object
// before it is created.
func (j *PVCTestJig) CreatePVorFailCSI(namespace string, scName string, ocid string) *v1.PersistentVolume {
	pv := j.newPVTemplateCSI(namespace, scName, ocid)

	result, err := j.KubeClient.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create persistent volume claim %q: %v", pv.Name, err)
	}
	return result
}

// CreatePVorFail creates a new claim based on the jig's
// defaults. Callers can provide a function to tweak the claim object
// before it is created.
func (j *PVCTestJig) CreatePVorFailCSIHighPerf(namespace string, scName string, ocid string) *v1.PersistentVolume {
	pv := j.newPVTemplateCSIHighPerf(namespace, scName, ocid)

	result, err := j.KubeClient.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create persistent volume claim %q: %v", pv.Name, err)
	}
	return result
}

// CreateVolume is a function to create the block volume
func (j *PVCTestJig) CreateVolume(bs ocicore.BlockstorageClient, adLabel string, compartmentId string, volName string, vpusPerGB int64) *string {
	var size int64 = 50
	request := ocicore.CreateVolumeRequest{
		CreateVolumeDetails: ocicore.CreateVolumeDetails{
			AvailabilityDomain: &adLabel,
			DisplayName:        &volName,
			SizeInGBs:          &size,
			CompartmentId:      &compartmentId,
			VpusPerGB:          &vpusPerGB,
		},
	}

	newVolume, err := bs.CreateVolume(context.Background(), request)
	if err != nil {
		Failf("Volume %q creation API error: %v", volName, err)
	}
	return newVolume.Id
}

// newPODTemplate returns the default template for this jig,
// creates the Pod. Attaches PVC to the Pod which is created by CSI
func (j *PVCTestJig) NewPodForCSI(name string, namespace string, claimName string, adLabel string) string {
	By("Creating a pod with the claiming PVC created by CSI")

	pod, err := j.KubeClient.CoreV1().Pods(namespace).Create(context.Background(), &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: j.Name,
			Namespace:    namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    name,
					Image:   centos,
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", "echo 'Hello World' > /data/testdata.txt; while true; do echo $(date -u) >> /data/out.txt; sleep 5; done"},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "persistent-storage",
							MountPath: "/data",
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "persistent-storage",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: claimName,
						},
					},
				},
			},
			NodeSelector: map[string]string{
				plugin.LabelZoneFailureDomain: adLabel,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		Failf("Pod %q Create API error: %v", pod.Name, err)
	}

	// Waiting for pod to be running
	err = j.waitTimeoutForPodRunningInNamespace(pod.Name, namespace, slowPodStartTimeout)
	if err != nil {
		Failf("Pod %q is not Running: %v", pod.Name, err)
	}
	zap.S().With(pod.Namespace).With(pod.Name).Info("CSI POD is created.")
	return pod.Name
}

// NewPodForCSIFSSWrite returns the CSI Fss template for this jig,
// creates the Pod. Attaches PVC to the Pod which is created by CSI Fss. It does not have a node selector unlike the default pod template.
func (j *PVCTestJig) NewPodForCSIFSSWrite(name string, namespace string, claimName string, fileName string, encryptionEnabled bool) string {
	By("Creating a pod with the claiming PVC created by CSI")

	nodeSelectorMap := make(map[string]string)
	if encryptionEnabled {
		nodeSelectorMap["oke.oraclecloud.com/e2e.oci-fss-util"] = "installed"
	}
	command := fmt.Sprintf("while true; do echo %s >> /data/%s; sleep 5; done", name, fileName)
	pod, err := j.KubeClient.CoreV1().Pods(namespace).Create(context.Background(), &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: j.Name,
			Namespace:    namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    name,
					Image:   centos,
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", command},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "persistent-storage",
							MountPath: "/data",
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "persistent-storage",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: claimName,
						},
					},
				},
			},
			NodeSelector: nodeSelectorMap,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		Failf("Pod %q Create API error: %v", pod.Name, err)
	}

	// Waiting for pod to be running
	err = j.waitTimeoutForPodRunningInNamespace(pod.Name, namespace, slowPodStartTimeout)
	if err != nil {
		Failf("Pod %q is not Running: %v", pod.Name, err)
	}
	zap.S().With(pod.Namespace).With(pod.Name).Info("CSI POD is created.")
	return pod.Name
}

// NewPodForCSIFSSRead returns the CSI Fss read pod template for this jig,
// creates the Pod. Attaches PVC to the Pod which is created by CSI Fss. It does not have a node selector unlike the default pod template.
// It does a grep on the file with string matchString and goes to completion with an exit code either 0 or 1.
func (j *PVCTestJig) NewPodForCSIFSSRead(matchString string, namespace string, claimName string, fileName string, encryptionEnabled bool) {
	By("Creating a pod with the claiming PVC created by CSI")

	nodeSelectorMap := make(map[string]string)
	if encryptionEnabled {
		nodeSelectorMap["oke.oraclecloud.com/e2e.oci-fss-util"] = "installed"
	}
	command := fmt.Sprintf("grep -q -i %s /data/%s; exit $?", matchString, fileName)
	pod, err := j.KubeClient.CoreV1().Pods(namespace).Create(context.Background(), &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: j.Name,
			Namespace:    namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "readapp",
					Image:   centos,
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", command},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "persistent-storage",
							MountPath: "/data",
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
			Volumes: []v1.Volume{
				{
					Name: "persistent-storage",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: claimName,
						},
					},
				},
			},
			NodeSelector: nodeSelectorMap,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		Failf("CSI Fss read POD Create API error: %v", err)
	}

	// Waiting for pod to be running
	err = j.waitTimeoutForPodCompletedSuccessfullyInNamespace(pod.Name, namespace, slowPodStartTimeout)
	if err != nil {
		Failf("Pod %q failed: %v", pod.Name, err)
	}
	zap.S().With(pod.Namespace).With(pod.Name).Info("CSI Fss read POD is created.")
}

// WaitForPVCPhase waits for a PersistentVolumeClaim to be in a specific phase or until timeout occurs, whichever comes first.
func (j *PVCTestJig) WaitForPVCPhase(phase v1.PersistentVolumeClaimPhase, ns string, pvcName string) error {
	Logf("Waiting up to %v for PersistentVolumeClaim %s to have phase %s", DefaultTimeout, pvcName, phase)
	for start := time.Now(); time.Since(start) < DefaultTimeout; time.Sleep(Poll) {
		pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(ns).Get(context.Background(), pvcName, metav1.GetOptions{})
		if err != nil {
			Logf("Failed to get claim %q, retrying in %v. Error: %v", pvcName, Poll, err)
			continue
		} else {
			if pvc.Status.Phase == phase {
				Logf("PersistentVolumeClaim %s found and phase=%s (%v)", pvcName, phase, time.Since(start))
				return nil
			}
		}
		Logf("PersistentVolumeClaim %s found but phase is %s instead of %s.", pvcName, pvc.Status.Phase, phase)
	}
	return fmt.Errorf("PersistentVolumeClaim %s not in phase %s within %v", pvcName, phase, DefaultTimeout)
}

// WaitForPVPhase waits for a PersistentVolume to be in a specific phase or until timeout occurs, whichever comes first.
func (j *PVCTestJig) WaitForPVPhase(phase v1.PersistentVolumePhase, pvName string) error {
	Logf("Waiting up to %v for PersistentVolumeClaim %s to have phase %s", DefaultTimeout, pvName, phase)
	for start := time.Now(); time.Since(start) < DefaultTimeout; time.Sleep(Poll) {
		pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), pvName, metav1.GetOptions{})
		if err != nil {
			Logf("Failed to get pv %q, retrying in %v. Error: %v", pvName, Poll, err)
			continue
		} else {
			if pv.Status.Phase == phase {
				Logf("PersistentVolumeClaim %s found and phase=%s (%v)", pvName, phase, time.Since(start))
				return nil
			}
		}
		Logf("PersistentVolume %s found but phase is %s instead of %s.", pvName, pv.Status.Phase, phase)
	}
	return fmt.Errorf("PersistentVolume %s not in phase %s within %v", pvName, phase, DefaultTimeout)
}

// SanityCheckPV checks basic properties of a given volume match
// our expectations.
func (j *PVCTestJig) SanityCheckPV(pvc *v1.PersistentVolumeClaim) {
	By("Checking the claim and volume are bound.")
	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(context.Background(), pvc.Name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	// Get the bound PV
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get persistent volume %q: %v", pvc.Spec.VolumeName, err)
	}
	if strings.Contains(pvc.Name, "expand") {
		Logf("Waiting upto 3 minutes for the block volume to resize.")
		iterator := 6
		for iterator > 0 && pv.Spec.Capacity[v1.ResourceStorage] != pvc.Spec.Resources.Requests[v1.ResourceStorage] {
			pvc, err = j.KubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(context.Background(), pvc.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			pv, _ = j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), pvc.Spec.VolumeName, metav1.GetOptions{})
			if len(pvc.Status.Conditions) >= 1 {
				Logf("Checking for PVC to resize. Type : %q, Status : %q", pvc.Status.Conditions[0].Type,
					pvc.Status.Conditions[0].Status)
			}
			Logf("pvCapacity : %q, pvStatus : %q, claimCapacity : %q", pv.Spec.Capacity[v1.ResourceStorage],
				pv.Status.Phase, pvc.Spec.Resources.Requests[v1.ResourceStorage])

			if len(pvc.Status.Conditions) > 0 {
				Logf("Resizer :: Type : %q, Status : %q, pvCapacity : %q, claimCapacity : %q",
					pvc.Status.Conditions[0].Type, pvc.Status.Conditions[0].Status,
					pv.Spec.Capacity[v1.ResourceStorage],
					pvc.Spec.Resources.Requests[v1.ResourceStorage])
			}

			time.Sleep(10 * time.Second)
			iterator -= 1
		}
	}
	// Check sizes
	pvCapacity := pv.Spec.Capacity[v1.ResourceStorage]
	claimCapacity := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	Expect(pvCapacity.Value()).To(Equal(claimCapacity.Value()), "pvCapacity is not equal to expectedCapacity")

	if strings.Contains(pvc.Name, "fss") {
		expectedAccessModes := []v1.PersistentVolumeAccessMode{v1.ReadWriteMany}
		Expect(pv.Spec.AccessModes).To(Equal(expectedAccessModes))
	} else {
		expectedAccessModes := []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
		Expect(pv.Spec.AccessModes).To(Equal(expectedAccessModes))
	}
	// Check PV properties
	Expect(pv.Spec.ClaimRef.Name).To(Equal(pvc.ObjectMeta.Name))
	Expect(pv.Spec.ClaimRef.Namespace).To(Equal(pvc.ObjectMeta.Namespace))

	// The pv and pvc are both bound, but to each other?
	// Check that the PersistentVolume.ClaimRef matches the PVC
	if pv.Spec.ClaimRef == nil {
		Failf("PV %q ClaimRef is nil", pv.Name)
	}
	if pv.Spec.ClaimRef.Name != pvc.Name {
		Failf("PV %q ClaimRef's name (%q) should be %q", pv.Name, pv.Spec.ClaimRef.Name, pvc.Name)
	}
	if pvc.Spec.VolumeName != pv.Name {
		Failf("PVC %q VolumeName (%q) should be %q", pvc.Name, pvc.Spec.VolumeName, pv.Name)
	}
	if pv.Spec.ClaimRef.UID != pvc.UID {
		Failf("PV %q ClaimRef's UID (%q) should be %q", pv.Name, pv.Spec.ClaimRef.UID, pvc.UID)
	}
}

func (j *PVCTestJig) waitForConditionOrFail(namespace, name string, timeout time.Duration, message string, conditionFn func(*v1.PersistentVolumeClaim) bool) *v1.PersistentVolumeClaim {
	var pvc *v1.PersistentVolumeClaim
	pollFunc := func() (bool, error) {
		v, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if conditionFn(v) {
			pvc = v
			return true, nil
		}
		return false, nil
	}
	if err := wait.PollImmediate(Poll, timeout, pollFunc); err != nil {
		Failf("Timed out waiting for volume claim %q to %s", pvc.Name, message)
	}
	return pvc
}

func (j *PVCTestJig) waitForConditionOrFailForPV(name string, timeout time.Duration, message string, conditionFn func(*v1.PersistentVolume) bool) *v1.PersistentVolume {
	var pv *v1.PersistentVolume
	pollFunc := func() (bool, error) {
		v, err := j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if conditionFn(v) {
			pv = v
			return true, nil
		}
		return false, nil
	}
	if err := wait.PollImmediate(Poll, timeout, pollFunc); err != nil {
		Failf("Timed out waiting for volume claim %q to %s", pv.Name, message)
	}
	return pv
}

// DeletePersistentVolumeClaim deletes the PersistentVolumeClaim with the given name / namespace.
func (j *PVCTestJig) DeletePersistentVolumeClaim(ns string, pvcName string) error {
	if j.KubeClient != nil && len(pvcName) > 0 {
		Logf("Deleting PersistentVolumeClaim %q", pvcName)
		err := j.KubeClient.CoreV1().PersistentVolumeClaims(ns).Delete(context.Background(), pvcName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("PVC delete API error: %v", err)
		}
	}
	return nil
}

// CheckVolumeCapacity verifies the Capacity of Volume provisioned.
func (j *PVCTestJig) CheckVolumeCapacity(expected string, name string, namespace string) {

	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	// Get the bound PV
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get persistent volume %q: %v", pvc.Spec.VolumeName, err)
	}

	// Check sizes
	actual := pv.Spec.Capacity[v1.ResourceStorage]

	if actual.String() != expected {
		Failf("Expected volume to be %s but got %s", expected, actual)
	}
}

// CheckVolumePerformanceLevel verifies the Performance level of Volume provisioned.
func (j *PVCTestJig) CheckVolumePerformanceLevel(bs ocicore.BlockstorageClient, namespace, name string, expectedPerformanceLevel int64) {

	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	volumeName := pvc.Spec.VolumeName
	// Get the bound PV
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), volumeName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get persistent volume %q: %v", volumeName, err)
	}
	volumeOCID := pv.Spec.CSI.VolumeHandle

	request := ocicore.GetVolumeRequest{
		VolumeId: &volumeOCID,
	}

	volume, err := bs.GetVolume(context.Background(), request)
	if err != nil {
		Failf("GetVolume %q API error: %v", volumeOCID, err)
	}
	// Check perf units vpusPerGB
	actual := volume.VpusPerGB

	if *actual != expectedPerformanceLevel {
		Failf("Expected volume performance level to be %s but got %s", expectedPerformanceLevel, actual)
	}
}

// CheckCMEKKey verifies the expected and actual CMEK key
func (j *PVCTestJig) CheckCMEKKey(bs client.BlockStorageInterface, pvcName, namespace, kmsKeyIDExpected string) {

	By("Checking is Expected and Actual CMEK key matches")
	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), pvcName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	// Get the bound PV
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get persistent volume %q: %v", pvc.Spec.VolumeName, err)
	}
	volume, err := bs.GetVolume(context.Background(), pv.Spec.CSI.VolumeHandle)
	if err != nil {
		Failf("Volume %q get API error: %v", pv.Spec.CSI.VolumeHandle, err)
	}
	Logf("Expected KMSKey:%s, Actual KMSKey:%v", kmsKeyIDExpected, volume.KmsKeyId)
	if volume.KmsKeyId == nil || *volume.KmsKeyId != kmsKeyIDExpected {
		Failf("Expected and Actual KMS key for CMEK test doesn't match. Expected KMSKey:%s, Actual KMSKey:%v", kmsKeyIDExpected, volume.KmsKeyId)
	}
}

// CheckAttachmentTypeAndEncryptionType verifies attachment type and encryption type
func (j *PVCTestJig) CheckAttachmentTypeAndEncryptionType(compute client.ComputeInterface, pvcName, namespace, podName, expectedAttachmentType string) {
	By("Checking attachment type")
	pod, err := j.KubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Logf("node is:%s", pod.Spec.NodeName)
	node, err := j.KubeClient.CoreV1().Nodes().Get(context.Background(), pod.Spec.NodeName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	// Get the bound PV
	instanceID := strings.Replace(node.Spec.ProviderID, "oci://", "", -1)
	if instanceID == "" {
		Failf("ProviderID should not be empty")
	}

	compartmentID, ok := node.Annotations["oci.oraclecloud.com/compartment-id"]
	if !ok {
		Failf("Node CompartmentID annotation should not be empty")
	}

	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(context.Background(), pvcName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	// Get the bound PV
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get persistent volume %q: %v", pvc.Spec.VolumeName, err)
	}

	attachment, err := compute.FindActiveVolumeAttachment(context.Background(), compartmentID, pv.Spec.CSI.VolumeHandle)
	if err != nil {
		Failf("VolumeAttachment %q get API error: %v", instanceID, err)
	}
	attachmentType := ""
	switch v := attachment.(type) {
	case ocicore.ParavirtualizedVolumeAttachment:
		Logf("AttachmentType is paravirtualized for attachmentID:%s", *v.GetId())
		attachmentType = AttachmentTypeParavirtualized
	case ocicore.IScsiVolumeAttachment:
		Logf("AttachmentType is iscsi for attachmentID:%s", *v.GetId())
		attachmentType = AttachmentTypeISCSI
	default:
		Logf("Display Name %s device type %s", *v.GetDisplayName(), *v.GetDevice())
		Failf("Unknown Attachment Type for attachmentID: %v", *v.GetId())
	}

	instance, err := compute.GetInstance(context.Background(), instanceID)
	if err != nil {
		Failf("instance %q get API error: %v", instanceID, err)
	}

	if *instance.LaunchOptions.IsPvEncryptionInTransitEnabled {
		expectedAttachmentType = AttachmentTypeParavirtualized
	}
	if attachmentType != expectedAttachmentType {
		Failf("expected attachmentType: %s but got %s", expectedAttachmentType, attachmentType)
	}
	By("Checking encryption type")
	Logf("instance launch option has in-transit encryption %v: volume attachment has in-transit encryption "+
		"%v", *instance.LaunchOptions.IsPvEncryptionInTransitEnabled, *attachment.GetIsPvEncryptionInTransitEnabled())
	if *instance.LaunchOptions.IsPvEncryptionInTransitEnabled != *attachment.GetIsPvEncryptionInTransitEnabled() {
		Failf("instance launch option has in-transit encryption %v, but volume attachment has in-transit "+
			"encryption %v", *instance.LaunchOptions.IsPvEncryptionInTransitEnabled, *attachment.GetIsPvEncryptionInTransitEnabled())
	}
}

// CheckEncryptionType verifies encryption type
func (j *PVCTestJig) CheckEncryptionType(namespace, podName string) {
	By("Checking encryption type")
	dfCommand := "df | grep data"

	// This test is written this way, since the only way to verify if in-transit encryption is present on FSS is by checking the df command on the pod
	// and if the IP starts with 192.x.x.x is present on the FSS mount
	output, err := RunHostCmd(namespace, podName, dfCommand)
	if err != nil || output == "" {
		Failf("kubectl exec failed or output is nil")
	}

	ipStart192 := output[0:3]
	if ipStart192 == "192" {
		Logf("FSS has in-transit encryption %s", output)
	} else {
		Failf("FSS does not have in-transit encryption")
	}
}

func (j *PVCTestJig) CheckSinglePodReadWrite(namespace string, pvcName string, checkEncryption bool) {

	By("Creating Pod that can create and write to the file")
	uid := uuid.NewUUID()
	fileName := fmt.Sprintf("out_%s.txt", uid)
	podName := j.NewPodForCSIFSSWrite(string(uid), namespace, pvcName, fileName, checkEncryption)
	time.Sleep(30 * time.Second) //waiting for pod to become up and running

	if checkEncryption {
		By("check if in transit encryption is enabled")
		j.CheckEncryptionType(namespace, podName)
	}

	By("check if the file exists")
	j.CheckFileExists(namespace, podName, "/data", fileName)

	By("Creating Pod that can read contents of existing file")
	j.NewPodForCSIFSSRead(string(uid), namespace, pvcName, fileName, checkEncryption)
}

func (j *PVCTestJig) CheckMultiplePodReadWrite(namespace string, pvcName string, checkEncryption bool) {
	uid := uuid.NewUUID()
	fileName := fmt.Sprintf("out_%s.txt", uid)
	By("Creating Pod that can create and write to the file")
	uuid1 := uuid.NewUUID()
	podName1 := j.NewPodForCSIFSSWrite(string(uuid1), namespace, pvcName, fileName, checkEncryption)
	time.Sleep(30 * time.Second) //waiting for pod to become up and running

	By("check if the file exists")
	j.CheckFileExists(namespace, podName1, "/data", fileName)

	if checkEncryption {
		By("check if in transit encryption is enabled")
		j.CheckEncryptionType(namespace, podName1)
	}

	By("Creating Pod that can create and write to the file")
	uuid2 := uuid.NewUUID()
	podName2 := j.NewPodForCSIFSSWrite(string(uuid2), namespace, pvcName, fileName, checkEncryption)
	time.Sleep(30 * time.Second) //waiting for pod to become up and running

	if checkEncryption {
		By("check if in transit encryption is enabled")
		j.CheckEncryptionType(namespace, podName2)
	}

	By("Creating Pod that can read contents of existing file")
	j.NewPodForCSIFSSRead(string(uuid1), namespace, pvcName, fileName, checkEncryption)

	By("Creating Pod that can read contents of existing file")
	j.NewPodForCSIFSSRead(string(uuid2), namespace, pvcName, fileName, checkEncryption)
}

func (j *PVCTestJig) CheckDataPersistenceWithDeployment(pvcName string, ns string) {
	nodes, err := j.KubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})

	if err != nil {
		Failf("Error getting list of nodes: %v", err)
	}

	if len(nodes.Items) == 0 {
		Failf("No worker nodes are present in the cluster")
	}

	nodeSelectorLabels := map[string]string{}
	schedulableNodeFound := false

	for _, node := range nodes.Items {
		taintIsMaster := false
		if node.Spec.Unschedulable == false {
			for _, taint := range node.Spec.Taints {
				taintIsMaster = (taint.Key == "node-role.kubernetes.io/master" || taint.Key == "node-role.kubernetes.io/control-plane")
			}
			if !taintIsMaster {
				schedulableNodeFound = true
				nodeSelectorLabels = node.Labels
				break
			}
		}
	}

	if !schedulableNodeFound {
		Failf("No schedulable nodes found")
	}

	podRunningCommand := " while true; do true; done;"

	dataWritten := "Data written"

	writeCommand := "echo \"" + dataWritten + "\" >> /data/out.txt;"
	readCommand := "cat /data/out.txt"

	By("Creating a deployment")
	deploymentName := j.createDeploymentOnNodeAndWait(podRunningCommand, pvcName, ns, "data-persistence-deployment", 1, nodeSelectorLabels)

	deployment, err := j.KubeClient.AppsV1().Deployments(ns).Get(context.Background(), deploymentName, metav1.GetOptions{})

	if err != nil {
		Failf("Error while fetching deployment %v: %v", deploymentName, err)
	}

	set := labels.Set(deployment.Spec.Selector.MatchLabels)
	pods, err := j.KubeClient.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{LabelSelector: set.AsSelector().String()})

	if err != nil {
		Failf("Error getting list of pods: %v", err)
	}

	podName := pods.Items[0].Name

	By("Writing to the volume using the pod")
	_, err = RunHostCmd(ns, podName, writeCommand)

	if err != nil {
		Failf("Error executing write command a pod: %v", err)
	}

	By("Deleting the pod used to write to the volume")
	err = j.KubeClient.CoreV1().Pods(ns).Delete(context.Background(), podName, metav1.DeleteOptions{})

	if err != nil {
		Failf("Error sending pod delete request: %v", err)
	}

	By("Waiting timeout for pod to not be found in namespace")
	err = j.waitTimeoutForPodNotFoundInNamespace(podName, ns, DefaultTimeout)

	if err != nil {
		Failf("Error deleting podt: %v", err)
	}

	By("Waiting for pod to be restarted")
	err = j.waitTimeoutForDeploymentAvailable(deploymentName, ns, deploymentAvailableTimeout, 1)

	if err != nil {
		Failf("Error waiting for deployment to become available again: %v", err)
	}

	pods, err = j.KubeClient.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{LabelSelector: set.AsSelector().String()})

	if err != nil {
		Failf("Error getting list of pods: %v", err)
	}

	podName = pods.Items[0].Name

	By("Reading from the volume using the pod and checking data integrity")
	output, err := RunHostCmd(ns, podName, readCommand)

	if err != nil {
		Failf("Error executing write command a pod: %v", err)
	}

	if dataWritten != strings.TrimSpace(output) {
		Failf("Written data not found on the volume, written: %v, found: %v", dataWritten, strings.TrimSpace(output))
	}

}

func (j *PVCTestJig) CheckISCSIQueueDepthOnNode(namespace, podName string) {
	By("Find node driver pod")
	pod, err := j.KubeClient.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Logf("node is:%s", pod.Spec.NodeName)
	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": "csi-oci-node",
		},
	}
	listOptions := metav1.ListOptions{
		FieldSelector: fields.Set{
			"spec.nodeName": pod.Spec.NodeName,
		}.AsSelector().String(),
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	}

	nodeDriverPods, err := j.KubeClient.CoreV1().Pods("kube-system").List(context.Background(), listOptions)
	Expect(err).NotTo(HaveOccurred())

	if len(nodeDriverPods.Items) != 1 {
		Failf("Failed to find node driver pod for node %s", pod.Spec.NodeName)
	}

	nodeDriverPodName := nodeDriverPods.Items[0].Name
	Logf("CSI node driver pod name is: %s", nodeDriverPodName)

	By("Check iSCSI queue depth on node")
	command := "iscsiadm -m node -o show | grep \"node.session.queue_depth = 128\" | uniq"
	output, err := RunHostCmd("kube-system", nodeDriverPodName, command)
	Expect(err).NotTo(HaveOccurred())

	Expect(strings.TrimSpace(output)).To(Equal("node.session.queue_depth = 128"))
}
