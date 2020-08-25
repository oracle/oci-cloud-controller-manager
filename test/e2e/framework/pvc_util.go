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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/plugin"
	ocicore "github.com/oracle/oci-go-sdk/core"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
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
