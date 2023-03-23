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

	snapshot "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateVolumeSnapshotClassOrFail creates a new volume snapshot class based on the jig's defaults.
func (f *CloudProviderFramework) CreateVolumeSnapshotClassOrFail(name string, driverType string,
	parameters map[string]string, deletionPolicy string) string {

	snapshotDeletionPolicy := snapshot.VolumeSnapshotContentDelete
	if deletionPolicy == "Retain" {
		snapshotDeletionPolicy = snapshot.VolumeSnapshotContentRetain
	}

	classTemp := f.NewVolumeSnapshotClassTemplate(name, parameters, driverType, snapshotDeletionPolicy)

	class, err := f.SnapClientSet.SnapshotV1().VolumeSnapshotClasses().Create(context.Background(), classTemp, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Logf("Volume Snapshot Class already exists.")
			return name
		}
		Failf("Failed to create volume snapshot class %q: %v", name, err)
	}
	f.VolumeSnapshotClasses = append(f.VolumeSnapshotClasses, class.Name)
	return class.Name
}

// NewVolumeSnapshotClassTemplate returns the default template for this jig, but
// does not actually create the storage class. The default storage class has the same name
// as the jig
func (f *CloudProviderFramework) NewVolumeSnapshotClassTemplate(name string, parameters map[string]string,
	driverType string, deletionPolicy snapshot.DeletionPolicy) *snapshot.VolumeSnapshotClass {
	return &snapshot.VolumeSnapshotClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VolumeSnapshotClass",
			APIVersion: "snapshot.storage.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Driver:         driverType,
		Parameters:     parameters,
		DeletionPolicy: deletionPolicy,
	}
}

// DeleteVolumeSnapshotClass deletes a volume snapshot class given the name
func (f *CloudProviderFramework) DeleteVolumeSnapshotClass(name string) error {
	err := f.SnapClientSet.SnapshotV1().VolumeSnapshotClasses().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
