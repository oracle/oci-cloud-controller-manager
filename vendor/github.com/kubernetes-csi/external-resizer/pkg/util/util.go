/*
Copyright 2018 The Kubernetes Authors.

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

package util

import (
	"encoding/json"
	"fmt"
	"regexp"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

var (
	knownResizeConditions = map[v1.PersistentVolumeClaimConditionType]bool{
		v1.PersistentVolumeClaimResizing:                true,
		v1.PersistentVolumeClaimFileSystemResizePending: true,
	}

	// AnnPreResizeCapacity annotation is added to a PV when expanding volume.
	// Its value is status capacity of the PVC prior to the volume expansion
	// Its value will be set by the external-resizer when it deems that filesystem resize is required after resizing volume.
	// Its value will be used by pv_controller to determine pvc's status capacity when binding pvc and pv.
	AnnPreResizeCapacity = "volume.alpha.kubernetes.io/pre-resize-capacity"
)

// PVCKey returns an unique key of a PVC object,
func PVCKey(pvc *v1.PersistentVolumeClaim) string {
	return fmt.Sprintf("%s/%s", pvc.Namespace, pvc.Name)
}

// MergeResizeConditionsOfPVC updates pvc with requested resize conditions
// leaving other conditions untouched.
func MergeResizeConditionsOfPVC(oldConditions, newConditions []v1.PersistentVolumeClaimCondition) []v1.PersistentVolumeClaimCondition {
	newConditionSet := make(map[v1.PersistentVolumeClaimConditionType]v1.PersistentVolumeClaimCondition, len(newConditions))
	for _, condition := range newConditions {
		newConditionSet[condition.Type] = condition
	}

	var resultConditions []v1.PersistentVolumeClaimCondition
	for _, condition := range oldConditions {
		// If Condition is of not resize type, we keep it.
		if _, ok := knownResizeConditions[condition.Type]; !ok {
			newConditions = append(newConditions, condition)
			continue
		}
		if newCondition, ok := newConditionSet[condition.Type]; ok {
			// Use the new condition to replace old condition with same type.
			resultConditions = append(resultConditions, newCondition)
			delete(newConditionSet, condition.Type)
		}

		// Drop old conditions whose type not exist in new conditions.
	}

	// Append remains resize conditions.
	for _, condition := range newConditionSet {
		resultConditions = append(resultConditions, condition)
	}

	return resultConditions
}

func GetPVCPatchData(oldPVC, newPVC *v1.PersistentVolumeClaim, addResourceVersionCheck bool) ([]byte, error) {
	patchBytes, err := GetPatchData(oldPVC, newPVC)
	if err != nil {
		return patchBytes, err
	}

	if addResourceVersionCheck {
		patchBytes, err = addResourceVersion(patchBytes, oldPVC.ResourceVersion)
		if err != nil {
			return nil, fmt.Errorf("apply ResourceVersion to patch data failed: %v", err)
		}
	}
	return patchBytes, nil
}

func addResourceVersion(patchBytes []byte, resourceVersion string) ([]byte, error) {
	var patchMap map[string]interface{}
	err := json.Unmarshal(patchBytes, &patchMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling patch with %v", err)
	}
	u := unstructured.Unstructured{Object: patchMap}
	a, err := meta.Accessor(&u)
	if err != nil {
		return nil, fmt.Errorf("error creating accessor with  %v", err)
	}
	a.SetResourceVersion(resourceVersion)
	versionBytes, err := json.Marshal(patchMap)
	if err != nil {
		return nil, fmt.Errorf("error marshalling json patch with %v", err)
	}
	return versionBytes, nil
}

func GetPatchData(oldObj, newObj interface{}) ([]byte, error) {
	oldData, err := json.Marshal(oldObj)
	if err != nil {
		return nil, fmt.Errorf("marshal old object failed: %v", err)
	}
	newData, err := json.Marshal(newObj)
	if err != nil {
		return nil, fmt.Errorf("marshal new object failed: %v", err)
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, oldObj)
	if err != nil {
		return nil, fmt.Errorf("CreateTwoWayMergePatch failed: %v", err)
	}
	return patchBytes, nil
}

// HasFileSystemResizePendingCondition returns true if a pvc has a FileSystemResizePending condition.
// This means the controller side resize operation is finished, and kubelet side operation is in progress.
func HasFileSystemResizePendingCondition(pvc *v1.PersistentVolumeClaim) bool {
	for _, condition := range pvc.Status.Conditions {
		if condition.Type == v1.PersistentVolumeClaimFileSystemResizePending && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// SanitizeName changes any name to a sanitized name which can be accepted by kubernetes.
func SanitizeName(name string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9-]")
	name = re.ReplaceAllString(name, "-")
	if name[len(name)-1] == '-' {
		// name must not end with '-'
		name = name + "X"
	}
	return name
}
