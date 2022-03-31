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
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewStorageClassTemplate returns the default template for this jig, but
// does not actually create the storage class. The default storage class has the same name
// as the jig
func (f *CloudProviderFramework) newStorageClassTemplate(name string, provisionerType string,
	parameters map[string]string, testLabels map[string]string, volumeBindingMode *storagev1.VolumeBindingMode,
	allowVolumeExpansion bool) *storagev1.StorageClass {
	return &storagev1.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StorageClass",
			APIVersion: "storage.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: testLabels,
		},
		Provisioner:          provisionerType,
		Parameters:           parameters,
		VolumeBindingMode:    volumeBindingMode,
		AllowVolumeExpansion: &allowVolumeExpansion,
	}
}

// DeleteStorageClass deletes a storage class given the name
func (f *CloudProviderFramework) DeleteStorageClass(name string) error {
	err := f.ClientSet.StorageV1().StorageClasses().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// CreateStorageClassOrFail creates a new storage class based on the jig's defaults.
func (f *CloudProviderFramework) CreateStorageClassOrFail(name string, provisionerType string,
	parameters map[string]string, testLabels map[string]string, bindingMode string, allowVolumeExpansion bool) string {
	volumeBindingMode := storagev1.VolumeBindingImmediate
	if bindingMode == "WaitForFirstConsumer" {
		volumeBindingMode = storagev1.VolumeBindingWaitForFirstConsumer
	}
	classTemp := f.newStorageClassTemplate(name, provisionerType, parameters, testLabels, &volumeBindingMode, allowVolumeExpansion)

	class, err := f.ClientSet.StorageV1().StorageClasses().Create(context.Background(), classTemp, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Logf("Storage Class already exists. Updating existing storage class.")
			class, err = f.UpdateStorageClassOrFail(classTemp, allowVolumeExpansion, nil)
			if err != nil {
				Logf("Error: %v", err)
			}
			return name
		}
		Failf("Failed to create storage class %q: %v", name, err)
	}
	f.StorageClasses = append(f.StorageClasses, class.Name)
	return class.Name
}

func (f *CloudProviderFramework) UpdateStorageClassOrFail(storageClass *storagev1.StorageClass, allowVolumeExpansion bool,
	tweak func(sc *storagev1.StorageClass)) (*storagev1.StorageClass, error) {

	if tweak != nil {
		tweak(storageClass)
	}

	Logf("Updating a SC %q", storageClass.Name)

	oldSC, err := f.ClientSet.StorageV1().StorageClasses().Get(context.Background(), storageClass.Name,
		metav1.GetOptions{})
	if err != nil {
		Failf("Failed to find StorageClass %v : %q", storageClass.Name, err)
		return storageClass, err
	}
	newSC := oldSC.DeepCopy()
	newSC.AllowVolumeExpansion = &allowVolumeExpansion

	class, err := f.ClientSet.StorageV1().StorageClasses().Update(context.Background(), newSC,
		metav1.UpdateOptions{})
	return class, err
}

// NewCSIDriverTemplate returns the default template for this jig, but
// does not actually create the CSI Driver.
func (f *CloudProviderFramework) newCSIDriverTemplate(name string, testLabels map[string]string, fsGroupPolicy storagev1.FSGroupPolicy) *storagev1.CSIDriver {
	return &storagev1.CSIDriver{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CSIDriver",
			APIVersion: "storage.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: testLabels,
		},
		Spec: storagev1.CSIDriverSpec{
			FSGroupPolicy: &fsGroupPolicy,
		},
	}
}

// CreateCSIDriverOrFail creates a new storage class based on the jig's defaults.
func (f *CloudProviderFramework) CreateCSIDriverOrFail(name string,
	testLabels map[string]string, fsGroupPolicy storagev1.FSGroupPolicy) string {

	csiDriverTemp := f.newCSIDriverTemplate(name, testLabels, fsGroupPolicy)

	csiDriver, err := f.ClientSet.StorageV1().CSIDrivers().Create(context.Background(), csiDriverTemp, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Logf("CSI Driver already exists. Continuing with the existing one.")
			return csiDriver.Name
		}
		Failf("Failed to create CSI Driver %q: %v", name, err)
	}
	return csiDriver.Name
}

// DeleteCSIDriver deletes a CSI Driver given the name
func (f *CloudProviderFramework) DeleteCSIDriver(name string) error {
	err := f.ClientSet.StorageV1().CSIDrivers().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
