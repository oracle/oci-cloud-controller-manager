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
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var _ = Describe("CSI Volume Creation", func() {
	f := framework.NewDefaultFramework("csi-basic")
	Context("[cloudprovider][storage][csi][system-tags]", func() {
		It("Create PVC and POD for CSI.", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests")
			ctx := context.TODO()
			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			volumeName := pvcJig.GetVolumeNameFromPVC(pvc.GetName(), f.Namespace.Name)
			compartmentId := f.GetCompartmentId(*setupF)
			// read created BV
			volumes, err := f.Client.BlockStorage().GetVolumesByName(ctx, volumeName, compartmentId)
			framework.ExpectNoError(err)
			// volume name duplicate should not exist
			for _, volume := range volumes {
				framework.Logf("volume details %v :", volume)
				if setupF.AddOkeSystemTags && !framework.HasOkeSystemTags(volume.SystemTags) {
					framework.Failf("the resource %s is expected to have oke system tags", *volume.Id)
				}
			}

		})

		It("Create PVC with VolumeSize 1Gi but should use default 50Gi", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-1gi")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.VolumeFss, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("app2", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(15 * time.Second) //waiting for pod to up and running

			pvcObj := pvcJig.GetPVCByName(pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, pvcObj.Spec.VolumeName)

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
		})

		It("Create PVC with VolumeSize 100Gi should use 100Gi", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-100gi")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MaxVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("app3", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(15 * time.Second) //waiting for pod to up and running

			pvcObj := pvcJig.GetPVCByName(pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, pvcObj.Spec.VolumeName)

			pvcJig.CheckVolumeCapacity("100Gi", pvc.Name, f.Namespace.Name)
		})

		It("Data should persist on CSI volume on pod restart", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-pod-restart-data-persistence")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			pvcJig.CheckDataPersistenceWithDeployment(pvc.Name, f.Namespace.Name)

			pvcObj := pvcJig.GetPVCByName(pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, pvcObj.Spec.VolumeName)
		})

		It("FsGroup test for CSI", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-pod-nginx")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			pvcJig.CheckVolumeDirectoryOwnership(f.Namespace.Name, pvc)

			pvcObj := pvcJig.GetPVCByName(pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, pvcObj.Spec.VolumeName)
		})
	})

	Context("[cloudprovider][storage][csi][block]", func() {
		It("Create PVC and POD for CSI in Raw Block Volume Mode", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSIBlock("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(15 * time.Second)
			pvcObj := pvcJig.GetPVCByName(pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, pvcObj.Spec.VolumeName)
		})

		It("Data should persist on CSI raw block volume on pod restart", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-pod-restart-data-persistence")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)

			pvcJig.CheckDataPersistenceForRawBlockVolumeWithDeployment(pvc.Name, f.Namespace.Name)

			pvcObj := pvcJig.GetPVCByName(pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, pvcObj.Spec.VolumeName)
		})
	})
})

var _ = Describe("CSI Volume Creation with different fstypes", func() {
	f := framework.NewDefaultFramework("csi-fstypes")
	Context("[cloudprovider][storage][csi][fstypes][iSCSI]", func() {
		It("Create PVC with fstype as XFS", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-fstype-xfs")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", map[string]string{framework.FstypeKey: "xfs"}, pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MaxVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSI("app-xfs", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second)

			pvcJig.CheckFilesystemTypeOfVolumeInsidePod(f.Namespace.Name, podName, "xfs")

			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Create PVC with fstype as EXT3", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-fstype-ext3")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", map[string]string{framework.FstypeKey: "ext3"}, pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MaxVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			podName := pvcJig.NewPodForCSI("app-ext3", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckFilesystemTypeOfVolumeInsidePod(f.Namespace.Name, podName, "ext3")
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})

	Context("[cloudprovider][storage][csi][fstypes][paravirtualized]", func() {
		It("Create PVC with fstype as XFS with paravirtualized attachment type", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-fstype-xfs")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", map[string]string{framework.FstypeKey: "xfs", framework.KmsKey: setupF.CMEKKMSKey, framework.AttachmentType: framework.AttachmentTypeParavirtualized}, pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MaxVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			podName := pvcJig.NewPodForCSI("app-xfs", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			pvcJig.CheckCMEKKey(f.Client.BlockStorage(), pvc.Name, f.Namespace.Name, setupF.CMEKKMSKey)

			pvcJig.CheckAttachmentTypeAndEncryptionType(f.Client.Compute(), pvc.Name, f.Namespace.Name, podName, framework.AttachmentTypeParavirtualized)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcObj := pvcJig.GetPVCByName(pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, pvcObj.Spec.VolumeName)

			pvcJig.CheckFilesystemTypeOfVolumeInsidePod(f.Namespace.Name, podName, "xfs")
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
	Context("[cloudprovider][storage][csi][expand][fstypes][iSCSI]", func() {
		It("Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for iSCSI volumes with xfs filesystem type", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-pvc-expand-to-100gi-iscsi-xfs")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, framework.FstypeKey: "xfs"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSI("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckFilesystemTypeOfVolumeInsidePod(f.Namespace.Name, podName, "xfs")
			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName, "100G")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for iSCSI volumes with ext3 filesystem type", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-pvc-expand-to-100gi-iscsi-ext3")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, framework.FstypeKey: "ext3"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSI("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckFilesystemTypeOfVolumeInsidePod(f.Namespace.Name, podName, "ext3")
			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName, "99G")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})

var _ = Describe("CSI Volume Expansion iSCSI", func() {
	f := framework.NewDefaultFramework("csi-expansion")
	Context("[cloudprovider][storage][csi][expand][iSCSI][filesystem]", func() {
		It("Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for iSCSI volumes with existing storage class", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-iscsi")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSI("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName, "99G")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})

		It("Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for iSCSI volumes with new storage class", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-iscsi")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSI("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName, "99G")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})

	Context("[cloudprovider][storage][csi][expand][iSCSI][block]", func() {
		It("Expand Raw Block PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence for iSCSI volumes with existing storage class", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-iscsi")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSIBlock("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/dev", "xvda")
			pvcJig.CheckExpandedRawBlockVolumeReadWrite(f.Namespace.Name, podName)

			fmt.Print(podName)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})
	})

	It("Expand Raw Block PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence for iSCSI volumes with new storage class", func() {
		var size = "100Gi"
		pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-iscsi")

		scName := f.CreateStorageClassOrFail(framework.ClassOCICSIExpand, "blockvolume.csi.oraclecloud.com",
			map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI},
			pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
		pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
		podName := pvcJig.NewPodForCSIBlock("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

		time.Sleep(60 * time.Second) //waiting for pod to up and running

		expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

		time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

		pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
		pvcJig.CheckFileExists(f.Namespace.Name, podName, "/dev", "xvda")
		pvcJig.CheckExpandedRawBlockVolumeReadWrite(f.Namespace.Name, podName)
		f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		_ = f.DeleteStorageClass(framework.ClassOCICSIExpand)
	})
})

var _ = Describe("CSI Volume Performance Level", func() {
	f := framework.NewBackupFramework("csi-perf-level")
	Context("[cloudprovider][storage][csi][perf][iSCSI][filesystem]", func() {
		It("Create CSI block volume with Performance Level as Low Cost", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-lowcost")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "0"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("low-cost-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.LowCostPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Create CSI block volume with no Performance Level and verify default", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-default")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("default-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.BalancedPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})
		It("Create CSI block volume with Performance Level as High", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-high")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSI("high-perf-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running
			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.HigherPerformanceOption)
			pvcJig.CheckISCSIQueueDepthOnNode(f.Namespace.Name, podName)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
	Context("[cloudprovider][storage][csi][perf][paravirtualized][filesystem]", func() {
		It("Create CSI block volume with Performance Level as Low Cost", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-lowcost")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "0"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("low-cost-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.LowCostPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Create CSI block volume with no Performance Level and verify default", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-balanced")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "10"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("default-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.BalancedPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})

		It("Create CSI block volume with Performance Level as High", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-high")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("high-perf-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running
			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.HigherPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})

	Context("[cloudprovider][storage][csi][perf][static]", func() {
		It("High Performance Static Provisioning CSI", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-static-high")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)

			compartmentId := ""
			if setupF.Compartment1 != "" {
				compartmentId = setupF.Compartment1
			} else if f.CloudProviderConfig.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.CompartmentID
			} else if f.CloudProviderConfig.Auth.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.Auth.CompartmentID
			} else {
				framework.Failf("Compartment Id undefined.")
			}
			pvc, volumeId := pvcJig.CreateAndAwaitStaticPVCOrFailCSI(f.BlockStorageClient, f.Namespace.Name, framework.MinVolumeBlock, csi_util.HigherPerformanceOption, scName, setupF.AdLocation, compartmentId, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			podName := pvcJig.NewPodForCSI("app4", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
			pvcJig.CheckISCSIQueueDepthOnNode(pvc.Namespace, podName)
			f.VolumeIds = append(f.VolumeIds, volumeId)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})

var _ = Describe("CSI Raw Block Volume Performance Level", func() {
	f := framework.NewBackupFramework("csi-perf-level")
	Context("[cloudprovider][storage][csi][perf][iSCSI][block]", func() {
		It("Create CSI raw block volume with Performance Level as Low Cost", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-lowcost")

			scName := f.CreateStorageClassOrFail(framework.ClassOCILowCost, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "0"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSIBlock("low-cost-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.LowCostPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCILowCost)
		})

		It("Create CSI raw block volume with no Performance Level and verify default", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-default")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSIBlock("default-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.BalancedPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})

		It("Create CSI raw block volume with Performance Level as High", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-high")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIHigh, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSIBlock("high-perf-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running
			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.HigherPerformanceOption)
			pvcJig.CheckISCSIQueueDepthOnNode(f.Namespace.Name, podName)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCIHigh)
		})
	})

	Context("[cloudprovider][storage][csi][perf][paravirtualized][block]", func() {
		It("Create CSI raw block volume with Performance Level as Low Cost", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-lowcost")

			scName := f.CreateStorageClassOrFail(framework.ClassOCILowCost, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "0"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSIBlock("low-cost-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.LowCostPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCILowCost)
		})
		It("Create CSI raw block volume with no Performance Level and verify default", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-balanced")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIBalanced, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "10"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSIBlock("default-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.BalancedPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCIBalanced)
		})

		It("Create CSI raw block volume with Performance Level as High", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-high")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIHigh, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSIBlock("high-perf-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running
			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.HigherPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCIHigh)
		})
	})

	Context("[cloudprovider][storage][csi][perf][static]", func() {
		It("High Performance Static Provisioning CSI", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-static-high")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIHigh, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)

			compartmentId := ""
			if setupF.Compartment1 != "" {
				compartmentId = setupF.Compartment1
			} else if f.CloudProviderConfig.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.CompartmentID
			} else if f.CloudProviderConfig.Auth.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.Auth.CompartmentID
			} else {
				framework.Failf("Compartment Id undefined.")
			}
			pvc, volumeId := pvcJig.CreateAndAwaitStaticPVCOrFailCSI(f.BlockStorageClient, f.Namespace.Name, framework.MinVolumeBlock, csi_util.HigherPerformanceOption, scName, setupF.AdLocation, compartmentId, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			podName := pvcJig.NewPodForCSIBlock("app4", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
			pvcJig.CheckISCSIQueueDepthOnNode(pvc.Namespace, podName)
			f.VolumeIds = append(f.VolumeIds, volumeId)
			_ = f.DeleteStorageClass(framework.ClassOCIHigh)
		})
	})
})

var _ = Describe("CSI Ultra High Performance Volumes", func() {
	f := framework.NewBackupFramework("csi-uhp")
	Context("[cloudprovider][storage][csi][uhp]", func() {
		It("Create ISCSI CSI block volume with UHP Performance Level", func() {
			checkUhpPrerequisites(f)
			compartmentId := f.GetCompartmentId(*setupF)
			if compartmentId == "" {
				framework.Failf("Compartment Id undefined.")
			}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-uhp")
			ctx := context.Background()

			By("Running test: Create ISCSI CSI block volume with UHP Performance Level")
			scName := f.CreateStorageClassOrFail(framework.ClassOCIUHP+"-1", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "30"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSI("uhp-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			pvcJig.VerifyMultipathEnabled(ctx, f.ComputeClient, pvc.Name, f.Namespace.Name, compartmentId)

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			err := pvcJig.DeleteAndAwaitPod(f.Namespace.Name, podName)
			if err != nil {
				framework.Failf("Error deleting pod: %v", err)
			}
			_ = f.DeleteStorageClass(scName)
			By("Completed test: Create ISCSI CSI block volume with UHP Performance Level")

			By("Running test: Create Paravirtualized CSI block volume with UHP Performance Level")
			scName = f.CreateStorageClassOrFail(framework.ClassOCIUHP+"-2", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "30"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc = pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName = pvcJig.NewPodForCSI("uhp-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			pvcJig.VerifyMultipathEnabled(ctx, f.ComputeClient, pvc.Name, f.Namespace.Name, compartmentId)

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			err = pvcJig.DeleteAndAwaitPod(f.Namespace.Name, podName)
			if err != nil {
				framework.Failf("Error deleting pod: %v", err)
			}
			_ = f.DeleteStorageClass(scName)
			By("Completed test: Create Paravirtualized CSI block volume with UHP Performance Level")

			By("Running test: Create CSI block volume with UHP Performance Level and xfs file system")
			scName = f.CreateStorageClassOrFail(framework.ClassOCIUHP+"-3", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "30", framework.FstypeKey: "xfs"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc = pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName = pvcJig.NewPodForCSI("uhp-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			pvcJig.VerifyMultipathEnabled(ctx, f.ComputeClient, pvc.Name, f.Namespace.Name, compartmentId)

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			err = pvcJig.DeleteAndAwaitPod(f.Namespace.Name, podName)
			if err != nil {
				framework.Failf("Error deleting pod: %v", err)
			}
			_ = f.DeleteStorageClass(scName)
			By("Completed test: Create CSI block volume with UHP Performance Level and xfs file system")

			By("Running test: Static Provisioning CSI UHP")
			scName = f.CreateStorageClassOrFail(framework.ClassOCIUHP+"-4", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "30"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)

			pvc, volumeId := pvcJig.CreateAndAwaitStaticPVCOrFailCSI(f.BlockStorageClient, f.Namespace.Name, framework.MinVolumeBlock, 30, scName, setupF.AdLocation, compartmentId, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			podName = pvcJig.NewPodForCSI("app4", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			pvcJig.VerifyMultipathEnabled(ctx, f.ComputeClient, pvc.Name, f.Namespace.Name, compartmentId)

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
			err = pvcJig.DeleteAndAwaitPod(f.Namespace.Name, podName)
			if err != nil {
				framework.Failf("Error deleting pod: %v", err)
			}
			f.VolumeIds = append(f.VolumeIds, volumeId)
			_ = f.DeleteStorageClass(scName)
			By("Completed test: Static Provisioning CSI UHP")

			By("Running test: Basic Pod Delete UHP")
			scName = f.CreateStorageClassOrFail(framework.ClassOCIUHP+"-5", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "30"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc = pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName = pvcJig.NewPodForCSI("uhp-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			pvcJig.VerifyMultipathEnabled(ctx, f.ComputeClient, pvc.Name, f.Namespace.Name, compartmentId)

			volumeName := pvcJig.GetVolumeNameFromPVC(pvc.Name, f.Namespace.Name)
			framework.Logf("Pod name : %s", podName)
			framework.Logf("Persistent volume name : %s", volumeName)

			pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName)
			err = pvcJig.DeletePersistentVolumeClaim(f.Namespace.Name, pvc.Name)
			if err != nil {
				framework.Failf("Failed to delete persistent volume claim: %s", err.Error())
			}
			err = pvcJig.WaitTimeoutForPVNotFound(volumeName, 10*time.Minute)
			if err != nil {
				framework.Failf("Persistent volume did not terminate : %s", err.Error())
			}
			_ = f.DeleteStorageClass(scName)
			By("Completed test: Basic Pod Delete UHP")

			By("Running test: Create UHP PVC and POD for CSI with CMEK and in-transit encryption")
			scParameter := map[string]string{
				framework.KmsKey:         setupF.CMEKKMSKey,
				framework.AttachmentType: framework.AttachmentTypeISCSI,
				csi_util.VpusPerGB:       "30",
			}
			scName = f.CreateStorageClassOrFail(framework.ClassOCIKMS+"-1", "blockvolume.csi.oraclecloud.com", scParameter, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc = pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName = pvcJig.NewPodForCSI("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			pvcJig.VerifyMultipathEnabled(ctx, f.ComputeClient, pvc.Name, f.Namespace.Name, compartmentId)
			pvcJig.CheckCMEKKey(f.Client.BlockStorage(), pvc.Name, f.Namespace.Name, setupF.CMEKKMSKey)
			pvcJig.CheckAttachmentTypeAndEncryptionType(f.Client.Compute(), pvc.Name, f.Namespace.Name, podName, framework.AttachmentTypeISCSI)
			pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(scName)
			By("Completed test: Create UHP PVC and POD for CSI with CMEK and in-transit encryption")

			By("Running test: Create UHP and lower performance block volumes on same node")
			sc1params := map[string]string{
				framework.AttachmentType: framework.AttachmentTypeISCSI,
				csi_util.VpusPerGB:       "30",
			}
			sc2params := map[string]string{
				framework.AttachmentType: framework.AttachmentTypeISCSI,
			}
			testTwoPVCSetup(f, sc1params, sc2params)
			By("Completed test: Create UHP and lower performance block volumes on same node")

			By("Running test: Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for iSCSI UHP volume")
			pvcJig.Name = "csi-uhp-pvc-expand-to-100gi"
			var size = "100Gi"
			scName = f.CreateStorageClassOrFail(framework.ClassOCIUHP+"-6", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "30"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc = pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName = pvcJig.NewPodForCSI("expanded-uhp-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			time.Sleep(60 * time.Second) //waiting for pod to up and running
			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)
			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName, "99G")
			pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(scName)
			By("Completed test: Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for iSCSI UHP volume")

			By("Running test: Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for Paravirtualized UHP volume")
			scName = f.CreateStorageClassOrFail(framework.ClassOCIUHP+"-7", "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "30"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc = pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName = pvcJig.NewPodForCSI("expanded-uhp-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			time.Sleep(60 * time.Second) //waiting for pod to up and running
			expandedPvc = pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)
			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName, "99G")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(scName)
			By("Completed test: Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for Paravirtualized UHP volume")
		})
	})
})

var _ = Describe("CSI UHP Volumes additional e2es", func() {
	f := framework.NewBackupFramework("csi-uhp-additional")
	Context("[uhp]", func() {
		It("Create UHP paravirtual volume and lower performance ISCSI block volumes on same node", func() {
			checkUhpPrerequisites(f)
			sc1params := map[string]string{
				framework.AttachmentType: framework.AttachmentTypeParavirtualized,
				csi_util.VpusPerGB:       "30",
			}
			sc2params := map[string]string{
				framework.AttachmentType: framework.AttachmentTypeISCSI,
			}
			testTwoPVCSetup(f, sc1params, sc2params)
		})
		It("Create UHP ISCSI volume and lower performance paravirtualized block volumes on same node", func() {
			checkUhpPrerequisites(f)
			sc1params := map[string]string{
				framework.AttachmentType: framework.AttachmentTypeISCSI,
				csi_util.VpusPerGB:       "30",
			}
			sc2params := map[string]string{
				framework.AttachmentType: framework.AttachmentTypeParavirtualized,
			}
			testTwoPVCSetup(f, sc1params, sc2params)
		})
		It("Create two UHP ISCSI block volumes on same node", func() {
			checkUhpPrerequisites(f)
			sc1params := map[string]string{
				framework.AttachmentType: framework.AttachmentTypeISCSI,
				csi_util.VpusPerGB:       "30",
			}
			sc2params := map[string]string{
				framework.AttachmentType: framework.AttachmentTypeISCSI,
				csi_util.VpusPerGB:       "30",
			}
			testTwoPVCSetup(f, sc1params, sc2params)
		})
	})
})

var _ = Describe("CSI Volume Expansion Paravirtualized", func() {
	f := framework.NewDefaultFramework("csi-expansion")
	Context("[cloudprovider][storage][csi][expand][paravirtualized][filesystem]", func() {
		It("Expand PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for paravirtualized volumes with new storage class", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-paravirtualized")

			scParameter := map[string]string{
				framework.KmsKey:         setupF.CMEKKMSKey,
				framework.AttachmentType: framework.AttachmentTypeParavirtualized,
			}
			scName := f.CreateStorageClassOrFail(f.Namespace.Name,
				"blockvolume.csi.oraclecloud.com", scParameter, pvcJig.Labels,
				"WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSI("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, podName, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, podName)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, podName, "99G")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})

	Context("[cloudprovider][storage][csi][expand][paravirtualized][block]", func() {
		It("Expand Raw Block PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence for paravirtualized volumes with new storage class", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-paravirtualized")

			scParameter := map[string]string{
				framework.KmsKey:         setupF.CMEKKMSKey,
				framework.AttachmentType: framework.AttachmentTypeParavirtualized,
			}
			scName := f.CreateStorageClassOrFail(framework.ClassOCICSIExpand,
				"blockvolume.csi.oraclecloud.com", scParameter, pvcJig.Labels,
				"WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			podName := pvcJig.NewPodForCSIBlock("expanded-pvc-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			time.Sleep(120 * time.Second) //waiting for expanded pvc to be functional

			pvcJig.CheckVolumeCapacity("100Gi", expandedPvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, podName, "/dev", "xvda")
			pvcJig.CheckExpandedRawBlockVolumeReadWrite(f.Namespace.Name, podName)

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCICSIExpand)
		})
	})
})

var _ = Describe("CSI Static Volume Creation", func() {
	f := framework.NewBackupFramework("csi-static")
	Context("[cloudprovider][storage][csi][static]", func() {
		It("Static Provisioning CSI", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-static")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)

			compartmentId := ""
			if setupF.Compartment1 != "" {
				compartmentId = setupF.Compartment1
			} else if f.CloudProviderConfig.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.CompartmentID
			} else if f.CloudProviderConfig.Auth.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.Auth.CompartmentID
			} else {
				framework.Failf("Compartment Id undefined.")
			}
			pvc, volumeId := pvcJig.CreateAndAwaitStaticPVCOrFailCSI(f.BlockStorageClient, f.Namespace.Name, framework.MinVolumeBlock, 10, scName, setupF.AdLocation, compartmentId, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.NewPodForCSI("app4", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, volumeId)
		})
	})
})

var _ = Describe("CSI CMEK,PV attachment and in-transit encryption test", func() {
	f := framework.NewDefaultFramework("csi-basic")
	Context("[cloudprovider][storage][csi][cmek][paravirtualized]", func() {
		It("Create PVC and POD for CSI with CMEK,PV attachment and in-transit encryption", func() {
			TestCMEKAttachmentTypeAndEncryptionType(f, framework.AttachmentTypeParavirtualized)
		})
	})

	Context("[cloudprovider][storage][csi][cmek][iscsi]", func() {
		It("Create PVC and POD for CSI with CMEK,ISCSI attachment and in-transit encryption", func() {
			TestCMEKAttachmentTypeAndEncryptionType(f, framework.AttachmentTypeISCSI)
		})
	})

})

func TestCMEKAttachmentTypeAndEncryptionType(f *framework.CloudProviderFramework, expectedAttachmentType string) {
	pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-cmek-iscsi-in-transit-e2e-tests")
	scParameter := map[string]string{
		framework.KmsKey:         setupF.CMEKKMSKey,
		framework.AttachmentType: expectedAttachmentType,
	}
	scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", scParameter, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
	pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
	podName := pvcJig.NewPodForCSI("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel)
	pvcJig.CheckCMEKKey(f.Client.BlockStorage(), pvc.Name, f.Namespace.Name, setupF.CMEKKMSKey)
	pvcJig.CheckAttachmentTypeAndEncryptionType(f.Client.Compute(), pvc.Name, f.Namespace.Name, podName, expectedAttachmentType)
	f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
	_ = f.DeleteStorageClass(f.Namespace.Name)
}

var _ = Describe("CSI Volume Capabilites", func() {
	f := framework.NewDefaultFramework("csi-basic")
	Context("[cloudprovider][storage][csi]", func() {
		It("Create volume fails with volumeMode set to block", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.NewPodForCSIWithoutWait("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			pvcObject := pvcJig.GetPVCByName(pvc.Name, f.Namespace.Name)
			err := pvcJig.WaitTimeoutForPVCBound(pvcObject.Name, f.Namespace.Name, 8*time.Minute)
			if err == nil {
				framework.Failf("PVC volume mode is not in pending status")
			}
			if pvcObject.Status.Phase != v1.ClaimPending {
				framework.Failf("PVC volume mode is not in pending status")
			}
		})

		It("Create volume fails with accessMode set to ReadWriteMany", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteMany, v1.ClaimPending)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.NewPodForCSIWithoutWait("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			pvcObject := pvcJig.GetPVCByName(pvc.Name, f.Namespace.Name)
			err := pvcJig.WaitTimeoutForPVCBound(pvcObject.Name, f.Namespace.Name, 8*time.Minute)
			if err == nil {
				framework.Failf("PVC volume mode is not in pending status")
			}
			if pvcObject.Status.Phase != v1.ClaimPending {
				framework.Failf("PVC volume mode is not in pending status")
			}
		})
	})
})

var _ = Describe("CSI Volume Creation - Immediate Volume Binding", func() {
	f := framework.NewDefaultFramework("csi-immediate")
	Context("[cloudprovider][storage][csi]", func() {
		It("Create PVC without pod and wait to be bound.", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-immediate-bind")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "Immediate", true, "Delete", nil)
			pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimBound)
			err := f.DeleteStorageClass(f.Namespace.Name)
			if err != nil {
				Fail(fmt.Sprintf("deleting storage class failed %s", f.Namespace.Name))
			}
		})
	})
})

func testTwoPVCSetup(f *framework.CloudProviderFramework, storageclass1params map[string]string, storageclass2params map[string]string) {
	compartmentId := f.GetCompartmentId(*setupF)
	if compartmentId == "" {
		framework.Failf("Compartment Id undefined.")
	}

	pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-two-pvc-setup")

	sc1Name := f.CreateStorageClassOrFail("storage-class-one", "blockvolume.csi.oraclecloud.com",
		storageclass1params,
		pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
	pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, sc1Name, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
	podName := pvcJig.NewPodForCSI("pvc-one-app", f.Namespace.Name, pvc.Name, setupF.AdLabel)

	ctx := context.Background()
	pvcJig.VerifyMultipathEnabled(ctx, f.ComputeClient, pvc.Name, f.Namespace.Name, compartmentId)

	nodeHostname := pvcJig.GetNodeHostnameFromPod(podName, f.Namespace.Name)

	nodeLabels := map[string]string{
		v1.LabelTopologyZone:        setupF.AdLabel,
		framework.NodeHostnameLabel: nodeHostname,
	}

	lowPerfScName := f.CreateStorageClassOrFail("storage-class-two", "blockvolume.csi.oraclecloud.com",
		storageclass2params,
		pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
	pvcTwo := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, lowPerfScName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
	podName2 := pvcJig.NewPodWithLabels("pvc-two-app", f.Namespace.Name, pvcTwo.Name, nodeLabels)

	pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName)
	pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName2)
	f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
	f.VolumeIds = append(f.VolumeIds, pvcTwo.Spec.VolumeName)
	_ = f.DeleteStorageClass("storage-class-one")
	_ = f.DeleteStorageClass("storage-class-two")
}

func checkUhpPrerequisites(f *framework.CloudProviderFramework) {
	if !f.RunUhpE2E {
		Skip("Skipping test since RUN_UHP_E2E environment variable is set to \"false\"")
	}
}
