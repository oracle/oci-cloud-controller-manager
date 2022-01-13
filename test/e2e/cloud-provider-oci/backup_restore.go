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
	. "github.com/onsi/ginkgo"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/block"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Backup/Restore", func() {
	f := framework.NewBackupFramework("backup-restore")
	Context("[cloudprovider][storage][fvp]", func() {
		It("should be possible to backup a volume and restore the created backup", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "volume-provisioner-e2e-tests-pvc")

			scName := f.CreateStorageClassOrFail(framework.ClassOCI, core.ProvisionerNameDefault, nil, pvcJig.Labels, "", false)

			By("Provisioning volume to backup")
			pvc := pvcJig.CreateAndAwaitPVCOrFail(f.Namespace.Name, framework.MinVolumeBlock, scName, setupF.AdLabel, nil)
			backupID, err := pvcJig.CreateBackupVolume(f.BlockStorageClient, pvc)
			if err != nil {
				framework.Failf("Failed to created backup for pvc %q: %v", pvc.Name, err)
			}
			f.BackupIDs = append(f.BackupIDs, backupID)
			framework.Logf("PVC %q has been backed up with the following id %q", pvc.Name, backupID)

			By("Teardown volume")
			pvcJig.DeletePersistentVolumeClaim(pvc.Name, f.Namespace.Name)

			By("Restoring the backup")
			pvcRestored := pvcJig.CreateAndAwaitPVCOrFail(f.Namespace.Name, framework.MinVolumeBlock, scName, setupF.AdLabel, func(pvcRestore *v1.PersistentVolumeClaim) {
				pvcRestore.Name = pvc.Name + "-restored"
				pvcRestore.ObjectMeta.Annotations = map[string]string{
					block.OCIVolumeBackupID: backupID,
				}
			})

			By("Creating pod to check read and write to volume")
			pvcJig.CheckVolumeReadWrite(f.Namespace.Name, pvcRestored)
		})
	})
})
