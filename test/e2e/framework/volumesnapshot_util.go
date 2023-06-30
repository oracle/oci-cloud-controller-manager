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
	"time"

	ocicore "github.com/oracle/oci-go-sdk/v65/core"

	snapshot "github.com/kubernetes-csi/external-snapshotter/client/v6/apis/volumesnapshot/v1"
	. "github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	BackupType            = "backupType"
	BackupTypeIncremental = "incremental"
	BackupTypeFull        = "full"
)

// CreateAndAwaitVolumeSnapshotOrFail creates a new VS based on the
// jig's defaults, waits for it to become ready, and then sanity checks it and
// its dependant resources. Callers can provide a function to tweak the
// VS object before it is created.
func (j *PVCTestJig) CreateAndAwaitVolumeSnapshotOrFail(namespace, vscName string, pvcName string,
	tweak func(vs *snapshot.VolumeSnapshot)) *snapshot.VolumeSnapshot {
	vs := j.CreateVolumeSnapshotOrFail(namespace, vscName, pvcName, tweak)
	return j.CheckAndAwaitVolumeSnapshotOrFail(vs, namespace)
}

// CreateVolumeSnapshotOrFail creates a new volume snapshot based on the jig's
// defaults. Callers can provide a function to tweak the volume snapshot object
// before it is created.
func (j *PVCTestJig) CreateVolumeSnapshotOrFail(namespace string, vscName string, pvcName string,
	tweak func(vs *snapshot.VolumeSnapshot)) *snapshot.VolumeSnapshot {
	vs := j.newVSTemplateCSI(namespace, vscName, pvcName)
	return j.CheckVSorFail(vs, tweak, namespace, pvcName)
}

// newVSTemplateCSI returns the default template for this jig, but
// does not actually create the Volume Snapshot. The default Volume Snapshot has the same name
// as the jig
func (j *PVCTestJig) newVSTemplateCSI(namespace string, vscName string, pvcName string) *snapshot.VolumeSnapshot {
	vs := j.CreateVSTemplate(namespace, vscName, pvcName)
	return vs
}

func (j *PVCTestJig) CreateVSTemplate(namespace string, vscName string, pvcName string) *snapshot.VolumeSnapshot {
	return &snapshot.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: j.Name,
			Labels:       j.Labels,
		},
		Spec: snapshot.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &vscName,
			Source: snapshot.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvcName,
			},
		},
	}
}

func (j *PVCTestJig) CheckVSorFail(vs *snapshot.VolumeSnapshot, tweak func(vs *snapshot.VolumeSnapshot),
	namespace string, pvcName string) *snapshot.VolumeSnapshot {
	if tweak != nil {
		tweak(vs)
	}

	name := types.NamespacedName{Namespace: namespace, Name: j.Name}
	By(fmt.Sprintf("Creating a Volume Snapshot %q using pvc %q", name, pvcName))

	result, err := j.SnapClient.SnapshotV1().VolumeSnapshots(namespace).Create(context.Background(), vs, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create persistent volume claim %q: %v", name, err)
	}
	return result
}

func (j *PVCTestJig) CheckAndAwaitVolumeSnapshotOrFail(vs *snapshot.VolumeSnapshot, namespace string) *snapshot.VolumeSnapshot {
	vs = j.waitForConditionOrFailVS(namespace, vs.Name, DefaultTimeout, "to be provisioned",
		func(*snapshot.VolumeSnapshot) bool {
			err := j.WaitForVSReadyToUse(namespace, vs.Name)
			if err != nil {
				Failf("Volume Snapshot %q did not reach readyToUse true state : %v", vs.Name, err)
				return false
			}
			return true
		})
	return vs
}

func (j *PVCTestJig) waitForConditionOrFailVS(namespace, name string, timeout time.Duration, message string, conditionFn func(volumeSnapshot *snapshot.VolumeSnapshot) bool) *snapshot.VolumeSnapshot {
	var vs *snapshot.VolumeSnapshot
	pollFunc := func() (bool, error) {
		v, err := j.SnapClient.SnapshotV1().VolumeSnapshots(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if conditionFn(v) {
			vs = v
			return true, nil
		}
		return false, nil
	}
	if err := wait.PollImmediate(Poll, timeout, pollFunc); err != nil {
		Failf("Timed out waiting for volume snapshot %q to %s", vs.Name, message)
	}
	return vs
}

// WaitForVSReadyToUse waits for a Volume Snapshot to be in a specific state or until timeout occurs, whichever comes first.
func (j *PVCTestJig) WaitForVSReadyToUse(ns string, vsName string) error {
	Logf("Waiting up to %v for VolumeSnapshot %s to have readyToUse true", DefaultTimeout, vsName)
	for start := time.Now(); time.Since(start) < DefaultTimeout; time.Sleep(Poll) {
		vs, err := j.SnapClient.SnapshotV1().VolumeSnapshots(ns).Get(context.Background(), vsName, metav1.GetOptions{})
		if err != nil {
			Logf("Failed to get volume snapshot %q, retrying in %v. Error: %v", vsName, Poll, err)
			continue
		} else {
			if vs.Status == nil {
				Logf("Failed to read volume snapshot %q status field, retrying in %v", vsName, Poll)
				continue
			} else if vs.Status.ReadyToUse == nil {
				Logf("Failed to read volume snapshot %q status ReadyToUse field, retrying in %v", vsName, Poll)
				continue
			} else if *vs.Status.ReadyToUse == true {
				Logf("VolumeSnapshot %s found and readyToUse=true (%v)", vsName, time.Since(start))
				return nil
			} else {
				Logf("VolumeSnapshot %s found but readyToUse is false instead of true.", vsName)
			}
		}
	}
	return fmt.Errorf("VolumeSnapshot %s not in readyToUse true within %v", vsName, DefaultTimeout)
}

// GetBackupIDFromSnapshot gets the backup OCID using the VolumeSnapshot
func (j *PVCTestJig) GetBackupIDFromSnapshot(vsName string, ns string) string {
	var vscontent *snapshot.VolumeSnapshotContent
	var err error

	vs, err := j.SnapClient.SnapshotV1().VolumeSnapshots(ns).Get(context.Background(), vsName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		Failf("VolumeSnapshot %q not found", vsName)
	}
	if err != nil {
		Failf("Error fetching volume snapshot %q: %v", vsName, err)
	}

	if vs != nil {
		if vs.Status != nil {
			if vs.Status.BoundVolumeSnapshotContentName != nil {
				vscontent, err = j.SnapClient.SnapshotV1().VolumeSnapshotContents().Get(context.Background(), *vs.Status.BoundVolumeSnapshotContentName, metav1.GetOptions{})
				if err != nil {
					Failf("Error getting volumesnapshotcontent object: %+v", err)
				}
			} else {
				Failf("Volume snapshot object BoundVolumeSnapshotContentName field empty when trying to get backupID from snapshot")
			}
		} else {
			Failf("Volume snapshot object status field empty when trying to get backupID from snapshot")
		}
	} else {
		Failf("Volume snapshot object empty when trying to get backupID from snapshot")
	}
	if vscontent != nil {
		if vscontent.Status != nil {
			if vscontent.Status.SnapshotHandle == nil {
				Failf("Volume snapshot content object SnapshotHandle field empty when trying to get backupID from snapshot content")
			}
		} else {
			Failf("Volume snapshot content object status field empty when trying to get backupID from snapshot content")
		}
	} else {
		Failf("Volume snapshot content object empty when trying to get backupID from snapshot content")
	}
	return *vscontent.Status.SnapshotHandle
}

// CreateVolumeSnapshotContentOrFail creates a new volume snapshot content based on the jig's defaults.
func (j *PVCTestJig) CreateVolumeSnapshotContentOrFail(name string, driverType string,
	backupOCID string, deletionPolicy string, vsName string, ns string) string {

	snapshotDeletionPolicy := snapshot.VolumeSnapshotContentDelete
	if deletionPolicy == "Retain" {
		snapshotDeletionPolicy = snapshot.VolumeSnapshotContentRetain
	}

	contentTemp := j.NewVolumeSnapshotContentTemplate(name, backupOCID, driverType, snapshotDeletionPolicy, vsName, ns)

	content, err := j.SnapClient.SnapshotV1().VolumeSnapshotContents().Create(context.Background(), contentTemp, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Logf("Volume Snapshot Content already exists.")
			return name
		}
		Failf("Failed to create volume snapshot content %q: %v", name, err)
	}
	return content.Name
}

// NewVolumeSnapshotContentTemplate returns the default template for this jig, but
// does not actually create the storage content. The default storage content has the same name
// as the jig
func (j *PVCTestJig) NewVolumeSnapshotContentTemplate(name string, backupOCID string,
	driverType string, deletionPolicy snapshot.DeletionPolicy, vsName string, ns string) *snapshot.VolumeSnapshotContent {
	return &snapshot.VolumeSnapshotContent{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VolumeSnapshotContent",
			APIVersion: "snapshot.storage.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: snapshot.VolumeSnapshotContentSpec{
			DeletionPolicy: deletionPolicy,
			Driver:         driverType,
			Source: snapshot.VolumeSnapshotContentSource{
				SnapshotHandle: &backupOCID,
			},
			VolumeSnapshotRef: v1.ObjectReference{
				Name:      vsName,
				Namespace: ns,
			},
		},
	}
}

func (j *PVCTestJig) CreateAndAwaitVolumeSnapshotStaticOrFail(name string, ns string, vscontentName string) *snapshot.VolumeSnapshot {
	vsTemp := &snapshot.VolumeSnapshot{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VolumeSnapshotContent",
			APIVersion: "snapshot.storage.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: snapshot.VolumeSnapshotSpec{
			Source: snapshot.VolumeSnapshotSource{
				VolumeSnapshotContentName: &vscontentName,
			},
		},
	}

	vs, err := j.SnapClient.SnapshotV1().VolumeSnapshots(ns).Create(context.Background(), vsTemp, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Logf("Volume Snapshot Content already exists.")
			return vsTemp
		}
		Failf("Failed to create volume snapshot content %q: %v", name, err)
	}
	return vs
}

// DeleteVolumeSnapshot deletes the VolumeSnapshot with the given name / namespace.
func (j *PVCTestJig) DeleteVolumeSnapshot(ns string, vsName string) error {
	if j.SnapClient != nil && len(vsName) > 0 {
		Logf("Deleting VolumeSnapshot %q", vsName)
		err := j.SnapClient.SnapshotV1().VolumeSnapshots(ns).Delete(context.Background(), vsName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("VolumeSnapshot delete API error: %v", err)
		}
	}
	return nil
}

// WaitTimeoutForVSContentNotFound waits default amount of time for the specified Volume Snapshot Content to be terminated.
// If the VSContent Get api returns IsNotFound then the wait stops and nil is returned. If the Get api returns
// an error other than "not found" then that error is returned and the wait stops.
func (j *PVCTestJig) WaitTimeoutForVSContentNotFound(vscontentName string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, timeout, j.vscontentNotFound(vscontentName))
}

func (j *PVCTestJig) vscontentNotFound(vscontentName string) wait.ConditionFunc {
	return func() (bool, error) {
		_, err := j.SnapClient.SnapshotV1().VolumeSnapshotContents().Get(context.Background(), vscontentName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return true, nil // done
		}
		if err != nil {
			return true, err // stop wait with error
		}
		return false, nil
	}
}

func (j *PVCTestJig) CheckVSContentExists(pvName string) bool {
	_, err := j.SnapClient.SnapshotV1().VolumeSnapshotContents().Get(context.Background(), pvName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false
	}
	if err != nil {
		return false
	}
	return true
}

// CreateVolumeBackup is a function to create a block volume backup
func (j *PVCTestJig) CreateVolumeBackup(bs ocicore.BlockstorageClient, adLabel string, compartmentId string, volumeId string, backupName string) *string {
	request := ocicore.CreateVolumeBackupRequest{
		CreateVolumeBackupDetails: ocicore.CreateVolumeBackupDetails{
			VolumeId:    &volumeId,
			DisplayName: &backupName,
			Type:        ocicore.CreateVolumeBackupDetailsTypeFull,
		},
	}

	newVolumeBackup, err := bs.CreateVolumeBackup(context.Background(), request)
	if err != nil {
		Failf("VolumeBackup %q creation API error: %v", backupName, err)
	}
	return newVolumeBackup.Id
}

// DeleteVolumeSnapshotContent deletes the VolumeSnapshotContent object with the given name
func (j *PVCTestJig) DeleteVolumeSnapshotContent(vscontentName string) error {
	if j.SnapClient != nil && len(vscontentName) > 0 {
		Logf("Deleting VolumeSnapshotContent %q", vscontentName)
		err := j.SnapClient.SnapshotV1().VolumeSnapshotContents().Delete(context.Background(), vscontentName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("VolumeSnapshotContent delete API error: %v", err)
		}
	}
	return nil
}

// DeleteVolumeBackup is a function to delete a block volume backup
func (j *PVCTestJig) DeleteVolumeBackup(bs ocicore.BlockstorageClient, backupId string) {
	request := ocicore.DeleteVolumeBackupRequest{
		VolumeBackupId: &backupId,
	}

	_, err := bs.DeleteVolumeBackup(context.Background(), request)
	if err != nil {
		Failf("VolumeBackup %q creation API error: %v", backupId, err)
	}
}

// GetVsContentNameFromVS is a function to get VS Content Name from VS
func (j *PVCTestJig) GetVsContentNameFromVS(vsName string, ns string) *string {
	var vscontentName *string

	vs, err := j.SnapClient.SnapshotV1().VolumeSnapshots(ns).Get(context.Background(), vsName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		Failf("VolumeSnapshot %q not found", vsName)
	}
	if err != nil {
		Failf("Error fetching volume snapshot %q: %v", vsName, err)
	}
	if vs != nil {
		if vs.Status != nil {
			if vs.Status.BoundVolumeSnapshotContentName != nil {
				vscontentName = vs.Status.BoundVolumeSnapshotContentName
			} else {
				Failf("Volume snapshot object BoundVolumeSnapshotContentName field empty when trying to get volume snapshot content name from snapshot")
			}
		} else {
			Failf("Volume snapshot object status field empty when trying to get volume snapshot content name from snapshot")
		}
	} else {
		Failf("Volume snapshot object empty when trying to get volume snapshot content name from snapshot")
	}

	return vscontentName
}
