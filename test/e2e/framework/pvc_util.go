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

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"

	"go.uber.org/zap"
	"golang.org/x/net/context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/plugin"
	ocicore "github.com/oracle/oci-go-sdk/v31/core"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	KmsKey                        = "kms-key-id"
	AttachmentTypeISCSI           = "iscsi"
	AttachmentTypeParavirtualized = "paravirtualized"
	SCName                        = "oci-kms"
	AttachmentType                = "attachment-type"
)

// PVCTestJig is a jig to help create PVC tests.
type PVCTestJig struct {
	ID                 string
	Name               string
	Labels             map[string]string
	BlockStorageClient *ocicore.BlockstorageClient
	KubeClient         clientset.Interface
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

// newPVCTemplate returns the default template for this jig, but
// does not actually create the PVC.  The default PVC has the same name
// as the jig
func (j *PVCTestJig) newPVCTemplate(namespace string, volumeSize string, scName string, adLabel string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: j.Name,
			Labels:       j.Labels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					plugin.LabelZoneFailureDomain: adLabel,
				},
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse(volumeSize),
				},
			},
			StorageClassName: &scName,
		},
	}
}

// CreatePVCorFail creates a new claim based on the jig's
// defaults. Callers can provide a function to tweak the claim object
// before it is created.
func (j *PVCTestJig) CreatePVCorFail(namespace string, volumeSize string, scName string, adLabel string, tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	pvc := j.newPVCTemplate(namespace, volumeSize, scName, adLabel)
	if tweak != nil {
		tweak(pvc)
	}

	name := types.NamespacedName{Namespace: namespace, Name: j.Name}
	By(fmt.Sprintf("Creating a PVC %q of volume size %q", name, volumeSize))

	result, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Create(pvc)
	if err != nil {
		Failf("Failed to create persistent volume claim %q: %v", name, err)
	}
	return result
}

// CreateAndAwaitPVCOrFail creates a new PVC based on the
// jig's defaults, waits for it to become ready, and then sanity checks it and
// its dependant resources. Callers can provide a function to tweak the
// PVC object before it is created.
func (j *PVCTestJig) CreateAndAwaitPVCOrFail(namespace string, volumeSize string, scName string, adLabel string, tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	pvc := j.CreatePVCorFail(namespace, volumeSize, scName, adLabel, tweak)
	pvc = j.waitForConditionOrFail(namespace, pvc.Name, DefaultTimeout, "to be dynamically provisioned", func(pvc *v1.PersistentVolumeClaim) bool {
		err := j.WaitForPVCPhase(v1.ClaimBound, namespace, pvc.Name)
		if err != nil {
			Failf("PVC %q did not become Bound: %v", pvc.Name, err)
			return false
		}
		return true
	})
	j.SanityCheckPV(pvc)
	return pvc
}

// newPVCTemplateCSI returns the default template for this jig, but
// does not actually create the PVC.  The default PVC has the same name
// as the jig
func (j *PVCTestJig) newPVCTemplateCSI(namespace string, volumeSize string, scName string) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: j.Name,
			Labels:       j.Labels,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse(volumeSize),
				},
			},
			StorageClassName: &scName,
		},
	}
}

// newPVTemplateCSI returns the default template for this jig, but
// does not actually create the PV.  The default PV has the same name
// as the jig
func (j *PVCTestJig) newPVTemplateCSI(namespace string, scName string, ocid string) *v1.PersistentVolume {
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: j.Name,
			Labels:       j.Labels,
			Annotations: map[string]string{
				"pv.kubernetes.io/provisioned-by": "blockvolume.csi.oraclecloud.com",
			},
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("50Gi"),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: &v1.CSIPersistentVolumeSource{
					Driver:       "blockvolume.csi.oraclecloud.com",
					FSType:       "ext4",
					VolumeHandle: ocid,
				},
			},
			PersistentVolumeReclaimPolicy: "Delete",
			StorageClassName:              scName,
		},
	}
}

// CreatePVorFail creates a new claim based on the jig's
// defaults. Callers can provide a function to tweak the claim object
// before it is created.
func (j *PVCTestJig) CreatePVorFailCSI(namespace string, scName string, ocid string) *v1.PersistentVolume {
	pv := j.newPVTemplateCSI(namespace, scName, ocid)

	result, err := j.KubeClient.CoreV1().PersistentVolumes().Create(pv)
	if err != nil {
		Failf("Failed to create persistent volume claim %q: %v", pv.Name, err)
	}
	return result
}

// CreatePVCorFail creates a new claim based on the jig's
// defaults. Callers can provide a function to tweak the claim object
// before it is created.
func (j *PVCTestJig) CreatePVCorFailCSI(namespace string, volumeSize string, scName string, tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	pvc := j.newPVCTemplateCSI(namespace, volumeSize, scName)
	if tweak != nil {
		tweak(pvc)
	}

	name := types.NamespacedName{Namespace: namespace, Name: j.Name}
	By(fmt.Sprintf("Creating a PVC %q of volume size %q", name, volumeSize))

	result, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Create(pvc)
	if err != nil {
		Failf("Failed to create persistent volume claim %q: %v", name, err)
	}
	return result
}

// CreateAndAwaitPVCOrFail creates a new PVC based on the
// jig's defaults, waits for it to become ready, and then sanity checks it and
// its dependant resources. Callers can provide a function to tweak the
// PVC object before it is created.
func (j *PVCTestJig) CreateAndAwaitPVCOrFailCSI(namespace string, volumeSize string, scName string, tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {
	pvc := j.CreatePVCorFailCSI(namespace, volumeSize, scName, tweak)
	pvc = j.waitForConditionOrFail(namespace, pvc.Name, DefaultTimeout, "to be dynamically provisioned", func(pvc *v1.PersistentVolumeClaim) bool {
		err := j.WaitForPVCPhase(v1.ClaimPending, namespace, pvc.Name)
		if err != nil {
			Failf("PVC %q did not become Bound: %v", pvc.Name, err)
			return false
		}
		return true
	})
	zap.S().With(pvc.Namespace).With(pvc.Name).Info("PVC is created.")
	return pvc
}

// CreateAndAwaitStaticPVCOrFailCSI creates a new PV and PVC based on the
// jig's defaults, waits for it to become ready, and then sanity checks it and
// its dependant resources. Callers can provide a function to tweak the
// PVC object before it is created.
func (j *PVCTestJig) CreateAndAwaitStaticPVCOrFailCSI(bs ocicore.BlockstorageClient, namespace string, volumeSize string, scName string, adLabel string, compartmentId string, tweak func(pvc *v1.PersistentVolumeClaim)) *v1.PersistentVolumeClaim {

	volumeOcid := j.CreateVolume(bs, adLabel, compartmentId, "test-volume")

	pv := j.CreatePVorFailCSI(namespace, scName, *volumeOcid)
	pv = j.waitForConditionOrFailForPV(pv.Name, DefaultTimeout, "to be dynamically provisioned", func(pvc *v1.PersistentVolume) bool {
		err := j.WaitForPVPhase(v1.VolumeAvailable, pv.Name)
		if err != nil {
			Failf("PV %q did not created: %v", pv.Name, err)
			return false
		}
		return true
	})

	pvc := j.CreatePVCorFailCSI(namespace, volumeSize, scName, tweak)
	pvc = j.waitForConditionOrFail(namespace, pvc.Name, DefaultTimeout, "to be dynamically provisioned", func(pvc *v1.PersistentVolumeClaim) bool {
		err := j.WaitForPVCPhase(v1.ClaimPending, namespace, pvc.Name)
		if err != nil {
			Failf("PVC %q did not become Bound: %v", pvc.Name, err)
			return false
		}
		return true
	})
	zap.S().With(pvc.Namespace).With(pvc.Name).Info("PVC is created.")
	return pvc
}

// CreateVolume is a function to create the block volume
func (j *PVCTestJig) CreateVolume(bs ocicore.BlockstorageClient, adLabel string, compartmentId string, volName string) *string {
	var size int64 = 50
	request := ocicore.CreateVolumeRequest{
		CreateVolumeDetails: ocicore.CreateVolumeDetails{
			AvailabilityDomain: &adLabel,
			DisplayName:        &volName,
			SizeInGBs:          &size,
			CompartmentId:      &compartmentId,
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
func (j *PVCTestJig) NewPODForCSI(name string, namespace string, claimName string, adLabel string) string {
	By("Creating a pod with the claiming PVC created by CSI")
	pod, err := j.KubeClient.CoreV1().Pods(namespace).Create(&v1.Pod{
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
					Args:    []string{"-c", "while true; do echo $(date -u) >> /data/out.txt; sleep 5; done"},
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
	})
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

// WaitForPVCPhase waits for a PersistentVolumeClaim to be in a specific phase or until timeout occurs, whichever comes first.
func (j *PVCTestJig) WaitForPVCPhase(phase v1.PersistentVolumeClaimPhase, ns string, pvcName string) error {
	Logf("Waiting up to %v for PersistentVolumeClaim %s to have phase %s", DefaultTimeout, pvcName, phase)
	for start := time.Now(); time.Since(start) < DefaultTimeout; time.Sleep(Poll) {
		pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(ns).Get(pvcName, metav1.GetOptions{})
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
		pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
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
	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(pvc.Name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	// Get the bound PV
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get persistent volume %q: %v", pvc.Spec.VolumeName, err)
	}

	// Check sizes
	pvCapacity := pv.Spec.Capacity[v1.ResourceName(v1.ResourceStorage)]
	claimCapacity := pvc.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	Expect(pvCapacity.Value()).To(Equal(claimCapacity.Value()), "pvCapacity is not equal to expectedCapacity")

	// Check PV properties
	expectedAccessModes := []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
	Expect(pv.Spec.AccessModes).To(Equal(expectedAccessModes))
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
		v, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
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
		v, err := j.KubeClient.CoreV1().PersistentVolumes().Get(name, metav1.GetOptions{})
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
		err := j.KubeClient.CoreV1().PersistentVolumeClaims(ns).Delete(pvcName, nil)
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("PVC delete API error: %v", err)
		}
	}
	return nil
}

// CheckVolumeCapacity verifies the Capacity of Volume provisioned.
func (j *PVCTestJig) CheckVolumeCapacity(expected string, name string, namespace string) {

	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())

	// Get the bound PV
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		Failf("Failed to get persistent volume %q: %v", pvc.Spec.VolumeName, err)
	}

	// Check sizes
	actual := pv.Spec.Capacity[v1.ResourceName(v1.ResourceStorage)]

	if actual.String() != expected {
		Failf("Expected volume to be %s but got %s", expected, actual)
	}
}

// CheckCMEKKey verifies the expected and actual CMEK key
func (j *PVCTestJig) CheckCMEKKey(bs client.BlockStorageInterface, pvcName, namespace, kmsKeyIDExpected string) {

	By("Checking is Expected and Actual CMEK key matches")
	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(pvcName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	// Get the bound PV
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
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
	pod, err := j.KubeClient.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Logf("node is:%s", pod.Spec.NodeName)
	node, err := j.KubeClient.CoreV1().Nodes().Get(pod.Spec.NodeName, metav1.GetOptions{})
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

	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(pvcName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	// Get the bound PV
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
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
