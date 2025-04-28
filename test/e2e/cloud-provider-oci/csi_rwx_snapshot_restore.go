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
	"fmt"

	. "github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"

	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var _ = Describe("RWX Raw Block Volume Snapshot Creation and Restore", func() {
	f := framework.NewBackupFramework("snapshot-restore")

	Context("[cloudprovider][storage][csi][snapshot][restore][raw-block][rwx]", func() {
		It("Should be able to schedule a pod on each worker node after creating a snapshot and restore from a RWX raw block backup", func() {
			checkOrInstallCRDs(f)
			scParams := map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-snapshot-restore-e2e-tests")

			nodes := pvcJig.ListSchedulableNodesInAD(setupF.AdLabel)
			if len(nodes) < 2 {
				Skip(fmt.Sprintf("at least 2 schedulable nodes in a AD is required to test MULTI_NODE %s", f.Namespace.Name))
			}

			pvcJig.InitialiseSnapClient(f.SnapClientSet)

			volId := pvcJig.CreateVolume(f.BlockStorageClient, setupF.AdLocation, setupF.StaticSnapshotCompartmentOcid, "test-volume", 10)

			backupOCID := pvcJig.CreateVolumeBackup(f.BlockStorageClient, setupF.StaticSnapshotCompartmentOcid, *volId, "test-backup")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, BVDriverName, scParams, pvcJig.Labels, BindingModeWaitForFirstConsumer, true, ReclaimPolicyDelete, nil)

			//creating a snapshot statically using the backup provisioned dynamically
			restoreVsName := "e2e-restore-vs"
			vscontentName := pvcJig.CreateVolumeSnapshotContentOrFail(f.Namespace.Name+"-e2e-snapshot-vsc", BVDriverName, *backupOCID, ReclaimPolicyDelete, restoreVsName, f.Namespace.Name, v1.PersistentVolumeBlock)

			pvcJig.CreateAndAwaitVolumeSnapshotStaticOrFail(restoreVsName, f.Namespace.Name, vscontentName)

			pvcRestore := pvcJig.CreateAndAwaitPVCOrFailSnapshotSource(f.Namespace.Name, framework.MinVolumeBlock, scName, restoreVsName, v1.ReadWriteMany, v1.ClaimPending, true, nil)

			var clonePodList []string
			// schedule a pod on each available node
			for i := range nodes {
				clonePod := pvcJig.NewPodForCSIwAntiAffinity(fmt.Sprintf("pod-%d", i), f.Namespace.Name, pvcRestore.Name, setupF.AdLabel, v1.PersistentVolumeBlock)
				clonePodList = append(clonePodList, clonePod)
			}

			//cleanup
			for _, pod := range clonePodList {
				pvcJig.DeleteAndAwaitPod(f.Namespace.Name, pod)
			}
			pvcJig.DeleteVolume(f.BlockStorageClient, *volId)
			pvcJig.DeleteVolumeBackup(f.BlockStorageClient, *backupOCID)
			f.VolumeIds = append(f.VolumeIds, pvcRestore.Spec.VolumeName, *volId)
			_ = f.DeleteVolumeSnapshotClass(f.Namespace.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})
