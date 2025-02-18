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

package e2e

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

const (
	WriteCommand                    = "echo 'Hello World' > /usr/share/nginx/html/testdata.txt; while true; do echo $(date -u) >> /data/out.txt; sleep 5; done"
	WriteCommandBlock               = "echo 'Hello World' > /var/test.txt; dd if=/var/test.txt of=/dev/xvda count=8; while true; do sleep 5; done"
	KeepAliveCommand                = "while true; do echo 'hello world' >> /usr/share/nginx/html/out.txt; sleep 5; done"
	KeepAliveCommandBlock           = "while true; do echo 'hello world' >> /var/newpod.txt; sleep 5; done"
	BVDriverName                    = "blockvolume.csi.oraclecloud.com"
	BindingModeWaitForFirstConsumer = "WaitForFirstConsumer"
	ReclaimPolicyDelete             = "Delete"
	ReclaimPolicyRetain             = "Retain"
)

var _ = Describe("Snapshot Creation and Restore", func() {
	f := framework.NewBackupFramework("snapshot-restore")

	Context("[cloudprovider][storage][csi][snapshot][restore]", func() {
		tests := []struct{
			attachmentType 	string
			backupType     	string
			fsType 			string
		}{
			{framework.AttachmentTypeParavirtualized, framework.BackupTypeIncremental, ""},
			{framework.AttachmentTypeParavirtualized, framework.BackupTypeFull, ""},
			{framework.AttachmentTypeISCSI, framework.BackupTypeIncremental, ""},
			{framework.AttachmentTypeISCSI, framework.BackupTypeFull, ""},
			{framework.AttachmentTypeISCSI, framework.BackupTypeIncremental, "xfs"},
			{framework.AttachmentTypeParavirtualized, framework.BackupTypeFull, "ext3"},
		}
		for _, entry := range tests {
			entry := entry
			testName := "Should be able to create and restore " + entry.backupType + " snapshot from " + entry.attachmentType + " volume "
			if entry.fsType != "" {
				testName += " with " + entry.fsType + " fsType"
			}
			It(testName, func() {
				scParams  := map[string]string{framework.AttachmentType: entry.attachmentType}
				vscParams := map[string]string{framework.BackupType: entry.backupType}
				scParams[framework.FstypeKey] = entry.fsType
				testSnapshotAndRestore(f, scParams, vscParams, v1.PersistentVolumeBlock)
			})
		}
		It("FS should get expanded when a PVC is restored with a lesser size backup (iscsi)", func() {
			checkOrInstallCRDs(f)
			scParams  := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			vscParams := map[string]string{framework.BackupType: framework.BackupTypeFull}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			_ = pvcJig.NewPodForCSI("pod-original", f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyDelete)
			vs  := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

			pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MaxVolumeBlock, scName, vs.Name, v1.ClaimPending, false, nil)
			podRestoreName := pvcJig.NewPodForCSI("pod-restored", f.Namespace.Name, pvcRestore.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podRestoreName, "99G")

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("FS should get expanded when a PVC is restored with a lesser size backup (paravirtualized)", func() {
			checkOrInstallCRDs(f)
			scParams  := map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized}
			vscParams := map[string]string{framework.BackupType: framework.BackupTypeFull}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			_ = pvcJig.NewPodForCSI("pod-original", f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyDelete)
			vs  := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

			pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MaxVolumeBlock, scName, vs.Name, v1.ClaimPending, false, nil)
			podRestoreName := pvcJig.NewPodForCSI("pod-restored", f.Namespace.Name, pvcRestore.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podRestoreName, "99G")

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Should be able to create and restore a snapshot from a backup(static case)", func() {
			checkOrInstallCRDs(f)
			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			vscParams := map[string]string{framework.BackupType: framework.BackupTypeFull}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			//creating a snapshot dynamically
			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			_ = pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvc, WriteCommand)
			vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyDelete)
			vs := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

			//Waiting for volume snapshot content to be created and status field to be populated
			time.Sleep(1 * time.Minute)

			vsName := vs.Name

			backupOCID := pvcJig.GetBackupIDFromSnapshot(vsName, f.Namespace.Name)

			//creating a snapshot statically using the backup provisioned dynamically
			restoreVsName := "e2e-restore-vs"
			vscontentName := pvcJig.CreateVolumeSnapshotContentOrFail(f.Namespace.Name+"-e2e-snapshot-vsc", BVDriverName, backupOCID, ReclaimPolicyDelete, restoreVsName, f.Namespace.Name, v1.PersistentVolumeFilesystem)

			pvcJig.CreateAndAwaitVolumeSnapshotStaticOrFail(restoreVsName, f.Namespace.Name, vscontentName)

			pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MinVolumeBlock, scName, restoreVsName, v1.ClaimPending, false, nil)
			podRestoreName := pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvcRestore, KeepAliveCommand)

			pvcJig.CheckFileExists(f.Namespace.Name, podRestoreName, "/usr/share/nginx/html", "testdata.txt")

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Should be able to create a snapshot and restore from a backup in another compartment", func() {
			checkOrInstallCRDs(f)
			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			volId := pvcJig.CreateVolume(f.BlockStorageClient, setupF.AdLocation, setupF.StaticSnapshotCompartmentOcid, "test-volume", 10)
			//wait for volume to become available
			time.Sleep(15 * time.Second)

			backupOCID := pvcJig.CreateVolumeBackup(f.BlockStorageClient, setupF.AdLabel, setupF.StaticSnapshotCompartmentOcid, *volId, "test-backup")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)

			//creating a snapshot statically using the backup provisioned dynamically
			restoreVsName := "e2e-restore-vs"
			vscontentName := pvcJig.CreateVolumeSnapshotContentOrFail(f.Namespace.Name+"-e2e-snapshot-vsc", BVDriverName, *backupOCID, ReclaimPolicyDelete, restoreVsName, f.Namespace.Name, v1.PersistentVolumeFilesystem)

			pvcJig.CreateAndAwaitVolumeSnapshotStaticOrFail(restoreVsName, f.Namespace.Name, vscontentName)

			pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MinVolumeBlock, scName, restoreVsName, v1.ClaimPending, false, nil)
			pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvcRestore, KeepAliveCommand)

			//wait for volume to be restored before starting cleanup
			time.Sleep(30 * time.Second)

			//cleanup
			pvcJig.DeleteVolume(f.BlockStorageClient, *volId)
			pvcJig.DeleteVolumeBackup(f.BlockStorageClient, *backupOCID)

			f.VolumeIds = append(f.VolumeIds, *volId)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})

var _ = Describe("Raw Block Volume Snapshot Creation and Restore", func() {
	f := framework.NewBackupFramework("snapshot-restore")

	Context("[cloudprovider][storage][csi][snapshot][restore][raw-block]", func() {
		testsBlock := []struct {
			attachmentType string
			backupType     string
			fsType         string
		}{
			{framework.AttachmentTypeParavirtualized, framework.BackupTypeIncremental, ""},
			{framework.AttachmentTypeParavirtualized, framework.BackupTypeFull, ""},
			{framework.AttachmentTypeISCSI, framework.BackupTypeIncremental, ""},
			{framework.AttachmentTypeISCSI, framework.BackupTypeFull, ""},
		}

		for _, entry := range testsBlock {
			entry := entry
			testName := "Should be able to create and restore " + entry.backupType + " snapshot from " + entry.attachmentType + " raw block volume"
			if entry.fsType != "" {
				testName += " with " + entry.fsType + " fsType"
			}
			It(testName, func() {
				scParams := map[string]string{framework.AttachmentType: entry.attachmentType}
				vscParams := map[string]string{framework.BackupType: entry.backupType}
				scParams[framework.FstypeKey] = entry.fsType
				testSnapshotAndRestore(f, scParams, vscParams, v1.PersistentVolumeBlock)
			})
		}
		It("FS should get expanded when a raw block PVC is restored with a lesser size backup (iscsi)", func() {
			checkOrInstallCRDs(f)
			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			vscParams := map[string]string{framework.BackupType: framework.BackupTypeFull}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)

			_ = pvcJig.NewPodForCSI("pod-original", f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyDelete)
			vs := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

			pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MaxVolumeBlock, scName, vs.Name, v1.ClaimPending, true, nil)
			podRestoreName := pvcJig.NewPodForCSI("pod-restored", f.Namespace.Name, pvcRestore.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckUsableVolumeSizeInsidePodBlock(f.Namespace.Name, podRestoreName, "100")

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("FS should get expanded when a raw block PVC is restored with a lesser size backup (paravirtualized", func() {
			checkOrInstallCRDs(f)
			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized}
			vscParams := map[string]string{framework.BackupType: framework.BackupTypeFull}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)

			_ = pvcJig.NewPodForCSI("pod-original", f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyDelete)
			vs := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

			pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MaxVolumeBlock, scName, vs.Name, v1.ClaimPending, true, nil)
			podRestoreName := pvcJig.NewPodForCSI("pod-restored", f.Namespace.Name, pvcRestore.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckUsableVolumeSizeInsidePodBlock(f.Namespace.Name, podRestoreName, "100")

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Should be able to create and restore a snapshot from a raw block volume backup(static case)", func() {
			checkOrInstallCRDs(f)
			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			vscParams := map[string]string{framework.BackupType: framework.BackupTypeFull}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			//creating a snapshot dynamically
			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			_ = pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvc, WriteCommandBlock)
			vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyDelete)
			vs := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

			//Waiting for volume snapshot content to be created and status field to be populated
			time.Sleep(1 * time.Minute)

			vsName := vs.Name

			backupOCID := pvcJig.GetBackupIDFromSnapshot(vsName, f.Namespace.Name)

			//creating a snapshot statically using the backup provisioned dynamically
			restoreVsName := "e2e-restore-vs"
			vscontentName := pvcJig.CreateVolumeSnapshotContentOrFail(f.Namespace.Name+"-e2e-snapshot-vsc", BVDriverName, backupOCID, ReclaimPolicyDelete, restoreVsName, f.Namespace.Name, v1.PersistentVolumeBlock)

			pvcJig.CreateAndAwaitVolumeSnapshotStaticOrFail(restoreVsName, f.Namespace.Name, vscontentName)

			pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MinVolumeBlock, scName, restoreVsName, v1.ClaimPending, true, nil)
			podRestoreName := pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvcRestore, KeepAliveCommandBlock)
			pvcJig.CheckDataInBlockDevice(f.Namespace.Name, podRestoreName, "Hello World")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Should be able to create a snapshot and restore from a raw block backup in another compartment", func() {
			checkOrInstallCRDs(f)
			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			volId := pvcJig.CreateVolume(f.BlockStorageClient, setupF.AdLocation, setupF.StaticSnapshotCompartmentOcid, "test-volume", 10)
			//wait for volume to become available
			time.Sleep(15 * time.Second)

			backupOCID := pvcJig.CreateVolumeBackup(f.BlockStorageClient, setupF.AdLabel, setupF.StaticSnapshotCompartmentOcid, *volId, "test-backup")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)

			//creating a snapshot statically using the backup provisioned dynamically
			restoreVsName := "e2e-restore-vs"
			vscontentName := pvcJig.CreateVolumeSnapshotContentOrFail(f.Namespace.Name+"-e2e-snapshot-vsc", BVDriverName, *backupOCID, ReclaimPolicyDelete, restoreVsName, f.Namespace.Name, v1.PersistentVolumeBlock)

			pvcJig.CreateAndAwaitVolumeSnapshotStaticOrFail(restoreVsName, f.Namespace.Name, vscontentName)

			pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MinVolumeBlock, scName, restoreVsName, v1.ClaimPending, true, nil)
			pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvcRestore, KeepAliveCommandBlock)

			//wait for volume to be restored before starting cleanup
			time.Sleep(30 * time.Second)

			//cleanup
			pvcJig.DeleteVolume(f.BlockStorageClient, *volId)
			pvcJig.DeleteVolumeBackup(f.BlockStorageClient, *backupOCID)

			f.VolumeIds = append(f.VolumeIds, *volId)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})

var _ = Describe("Volume Snapshot Deletion Tests", func() {
	f := framework.NewBackupFramework("snapshot-delete")

	Context("[cloudprovider][storage][csi][snapshot]", func() {
		It("Basic Delete POD and VS", func() {
			checkOrInstallCRDs(f)
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			vscParams := map[string]string{framework.BackupType: framework.BackupTypeFull}

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			_ = pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvc, WriteCommand)

			vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyDelete)
			vs := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

			//Waiting for volume snapshot content to be created and status field to be populated
			time.Sleep(1 * time.Minute)

			vsName := vs.Name
			var vscontentName *string

			vscontentName = pvcJig.GetVsContentNameFromVS(vsName, f.Namespace.Name)

			err := pvcJig.DeleteVolumeSnapshot(f.Namespace.Name, vsName)
			if err != nil {
				framework.Failf("Failed to delete volume snapshot: %s", err.Error())
			}
			err = pvcJig.WaitTimeoutForVSContentNotFound(*vscontentName, 10*time.Minute)
			if err != nil {
				framework.Failf("Volume Snapshot Content object did not terminate : %s", err.Error())
			}

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Test VSContent not deleted when reclaim policy is Retain", func() {
			checkOrInstallCRDs(f)
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			vscParams := map[string]string{framework.BackupType: framework.BackupTypeFull}

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			_ = pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvc, WriteCommand)

			vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyRetain)
			vs := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

			//Waiting for volume snapshot content to be created and status field to be populated
			time.Sleep(1 * time.Minute)

			vsName := vs.Name
			var vscontentName *string

			vscontentName = pvcJig.GetVsContentNameFromVS(vsName, f.Namespace.Name)

			//for cleanup
			backupId := pvcJig.GetBackupIDFromSnapshot(vsName, f.Namespace.Name)

			err := pvcJig.DeleteVolumeSnapshot(f.Namespace.Name, vsName)
			if err != nil {
				framework.Failf("Failed to delete volume snapshot: %s", err.Error())
			}
			time.Sleep(90 * time.Second)
			vscontentExists := pvcJig.CheckVSContentExists(*vscontentName)
			if vscontentExists != true {
				framework.Failf("Volume Snapshot Content was deleted")
			}

			//cleanup
			err = pvcJig.DeleteVolumeSnapshotContent(*vscontentName)
			if err != nil {
				framework.Failf("Failed to delete volume snapshot content: %s", err.Error())
			}
			pvcJig.DeleteVolumeBackup(f.BlockStorageClient, backupId)

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})

var _ = Describe("Raw Block Volume Snapshot Deletion Tests", func() {
	f := framework.NewBackupFramework("snapshot-delete")

	Context("[cloudprovider][storage][csi][snapshot][raw-block]", func() {
		It("Basic Delete POD and VS", func() {
			checkOrInstallCRDs(f)
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			vscParams := map[string]string{framework.BackupType: framework.BackupTypeFull}

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)

			_ = pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvc, WriteCommandBlock)

			vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyDelete)
			vs := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

			//Waiting for volume snapshot content to be created and status field to be populated
			time.Sleep(1 * time.Minute)

			vsName := vs.Name
			var vscontentName *string

			vscontentName = pvcJig.GetVsContentNameFromVS(vsName, f.Namespace.Name)

			err := pvcJig.DeleteVolumeSnapshot(f.Namespace.Name, vsName)
			if err != nil {
				framework.Failf("Failed to delete volume snapshot: %s", err.Error())
			}
			err = pvcJig.WaitTimeoutForVSContentNotFound(*vscontentName, 10*time.Minute)
			if err != nil {
				framework.Failf("Volume Snapshot Content object did not terminate : %s", err.Error())
			}

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Test VSContent not deleted when reclaim policy is Retain", func() {
			checkOrInstallCRDs(f)
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			vscParams := map[string]string{framework.BackupType: framework.BackupTypeFull}

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)

			_ = pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvc, WriteCommandBlock)

			vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyRetain)
			vs := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

			//Waiting for volume snapshot content to be created and status field to be populated
			time.Sleep(1 * time.Minute)

			vsName := vs.Name
			var vscontentName *string

			vscontentName = pvcJig.GetVsContentNameFromVS(vsName, f.Namespace.Name)

			//for cleanup
			backupId := pvcJig.GetBackupIDFromSnapshot(vsName, f.Namespace.Name)

			err := pvcJig.DeleteVolumeSnapshot(f.Namespace.Name, vsName)
			if err != nil {
				framework.Failf("Failed to delete volume snapshot: %s", err.Error())
			}
			time.Sleep(90 * time.Second)
			vscontentExists := pvcJig.CheckVSContentExists(*vscontentName)
			if vscontentExists != true {
				framework.Failf("Volume Snapshot Content was deleted")
			}

			//cleanup
			err = pvcJig.DeleteVolumeSnapshotContent(*vscontentName)
			if err != nil {
				framework.Failf("Failed to delete volume snapshot content: %s", err.Error())
			}
			pvcJig.DeleteVolumeBackup(f.BlockStorageClient, backupId)

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})

func testSnapshotAndRestore(f *framework.CloudProviderFramework, scParams map[string]string, vscParams map[string]string, volumeMode v1.PersistentVolumeMode) {
	checkOrInstallCRDs(f)
	pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")
	pvcJig.InitialiseSnapClient(f.SnapClientSet)

	scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)

	pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, volumeMode, v1.ReadWriteOnce, v1.ClaimPending)

	if volumeMode == v1.PersistentVolumeFilesystem {
		_ = pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvc, WriteCommand)
	} else {
		_ = pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvc, WriteCommandBlock)
	}

	// Waiting to be sure write command runs
	time.Sleep(30 * time.Second)

	vscName := f.CreateVolumeSnapshotClassOrFail(f.Namespace.Name, BVDriverName, vscParams, ReclaimPolicyDelete)
	vs := pvcJig.CreateAndAwaitVolumeSnapshotOrFail(f.Namespace.Name, vscName, pvc.Name, nil)

	if volumeMode == v1.PersistentVolumeFilesystem {
		pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MinVolumeBlock, scName, vs.Name, v1.ClaimPending, false, nil)
		podRestoreName := pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvcRestore, KeepAliveCommand)

		// Check if the file exists in the restored pod
		pvcJig.CheckFileExists(f.Namespace.Name, podRestoreName, "/usr/share/nginx/html", "testdata.txt")
	} else {
		pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MinVolumeBlock, scName, vs.Name, v1.ClaimPending, true, nil)
		podRestoreName := pvcJig.CreateAndAwaitNginxPodOrFail(f.Namespace.Name, pvcRestore, KeepAliveCommandBlock)

		// Check data in block device for restored pod
		pvcJig.CheckDataInBlockDevice(f.Namespace.Name, podRestoreName, "Hello World")
	}

	// Clean up
	f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
	_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
	_ = f.DeleteStorageClass(f.Namespace.Name)
}

func checkOrInstallCRDs(f *framework.CloudProviderFramework) {
	var err error

	_, err = f.CRDClientSet.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), "volumesnapshots.snapshot.storage.k8s.io", metav1.GetOptions{})
	if err != nil {
		if setupF.EnableCreateCluster == false {
			Skip("Skipping test because VolumeSnapshot CRD is not present")
		} else {
			framework.RunKubectl("create", "-f", "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v6.2.0/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml")
		}
	}

	_, err = f.CRDClientSet.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), "volumesnapshotclasses.snapshot.storage.k8s.io", metav1.GetOptions{})
	if err != nil {
		if setupF.EnableCreateCluster == false {
			Skip("Skipping test because VolumeSnapshotClass CRD is not present")
		} else {
			framework.RunKubectl("create", "-f", "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v6.2.0/client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml")
		}
	}

	_, err = f.CRDClientSet.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), "volumesnapshotcontents.snapshot.storage.k8s.io", metav1.GetOptions{})
	if err != nil {
		if setupF.EnableCreateCluster == false {
			Skip("Skipping test because VolumeSnapshotContent CRD is not present")
		} else {
			framework.RunKubectl("create", "-f", "https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/v6.2.0/client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml")
		}
	}
}
