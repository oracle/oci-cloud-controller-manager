package e2e

import (
	"time"

	. "github.com/onsi/ginkgo"

	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("CSI Volume Creation with PVC datasource", func() {
	f := framework.NewDefaultFramework("csi-volume-cloning")
	Context("[cloudprovider][storage][csi][cloning]", func() {
		It("Create PVC with source PVC name specified in dataSource", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-cloning-basic")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			srcPvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			srcPod := pvcJig.NewPodForCSI("app1", f.Namespace.Name, srcPvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) // Wait for data to be written to source PV

			clonePvc := pvcJig.CreateAndAwaitClonePVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, srcPvc.Name, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			clonePod := pvcJig.NewPodForCSIClone("app2", f.Namespace.Name, clonePvc.Name, setupF.AdLabel)
			pvcJig.CheckFileExists(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.DeleteAndAwaitPod(f.Namespace.Name, srcPod)
			pvcJig.DeleteAndAwaitPVC(f.Namespace.Name, srcPvc.Name)
			pvcJig.DeleteAndAwaitPod(f.Namespace.Name, clonePod)
			pvcJig.DeleteAndAwaitPVC(f.Namespace.Name, clonePvc.Name)
		})
	})

	Context("[cloudprovider][storage][csi][cloning][expand]", func() {
		It("Create Clone PVC with size greater than the source PVC", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-volume-size-test")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			srcPvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("app1", f.Namespace.Name, srcPvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) // Wait for data to be written to source PV

			clonePvc := pvcJig.CreateAndAwaitClonePVCOrFailCSI(f.Namespace.Name, framework.MaxVolumeBlock, scName, srcPvc.Name, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			clonePod := pvcJig.NewPodForCSIClone("app2", f.Namespace.Name, clonePvc.Name, setupF.AdLabel)

			pvcJig.CheckFileExists(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckVolumeCapacity(framework.MaxVolumeBlock, clonePvc.Name, f.Namespace.Name)
			pvcJig.CheckFileCorruption(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckExpandedVolumeReadWrite(f.Namespace.Name, clonePod)
			pvcJig.CheckUsableVolumeSizeInsidePod(f.Namespace.Name, clonePod, "98")
		})
	})

	Context("[cloudprovider][storage][csi][cloning]", func() {
		It("Should be able to create a clone volume with in-transit encryption", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-cloning-cmek-iscsi-in-transit-e2e-tests")
			scParameter := map[string]string{
				framework.KmsKey:         setupF.CMEKKMSKey,
				framework.AttachmentType: framework.AttachmentTypeParavirtualized,
			}
			scName := f.CreateStorageClassOrFail(framework.ClassOCIKMS, "blockvolume.csi.oraclecloud.com", scParameter, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			srcPvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			srcPod := pvcJig.NewPodForCSI("app1", f.Namespace.Name, srcPvc.Name, setupF.AdLabel)
			pvcJig.CheckCMEKKey(f.Client.BlockStorage(), srcPvc.Name, f.Namespace.Name, setupF.CMEKKMSKey)
			pvcJig.CheckAttachmentTypeAndEncryptionType(f.Client.Compute(), srcPvc.Name, f.Namespace.Name, srcPod, framework.AttachmentTypeParavirtualized)

			time.Sleep(60 * time.Second) // Wait for data to be written to source PV

			clonePvc := pvcJig.CreateAndAwaitClonePVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, srcPvc.Name, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			clonePod := pvcJig.NewPodForCSIClone("app2", f.Namespace.Name, clonePvc.Name, setupF.AdLabel)

			pvcJig.CheckFileExists(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckCMEKKey(f.Client.BlockStorage(), clonePvc.Name, f.Namespace.Name, setupF.CMEKKMSKey)
			pvcJig.CheckAttachmentTypeAndEncryptionType(f.Client.Compute(), clonePvc.Name, f.Namespace.Name, clonePod, framework.AttachmentTypeParavirtualized)

			f.VolumeIds = append(f.VolumeIds, srcPvc.Spec.VolumeName, clonePvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCIKMS)
		})
	})

	Context("[cloudprovider][storage][csi][cloning]", func() {
		It("Create PVC with source PVC name specified in dataSource - ISCSI", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-cloning-iscsi-test")

			scParameter := map[string]string{
				framework.KmsKey:         setupF.CMEKKMSKey,
				framework.AttachmentType: framework.AttachmentTypeISCSI,
			}
			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", scParameter, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			srcPvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("app1", f.Namespace.Name, srcPvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) // Wait for data to be written to source PV

			clonePvc := pvcJig.CreateAndAwaitClonePVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, srcPvc.Name, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			clonePod := pvcJig.NewPodForCSIClone("app2", f.Namespace.Name, clonePvc.Name, setupF.AdLabel)
			pvcJig.CheckFileExists(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, clonePod, "/data", "testdata.txt")
		})
	})

	Context("[cloudprovider][storage][csi][cloning]", func() {
		It("Create PVC with source PVC name specified in dataSource - ParaVirtualized", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-cloning-pv")

			scParameter := map[string]string{
				framework.KmsKey:         setupF.CMEKKMSKey,
				framework.AttachmentType: framework.AttachmentTypeParavirtualized,
			}
			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", scParameter, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			srcPvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("app1", f.Namespace.Name, srcPvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) // Wait for data to be written to source PV

			clonePvc := pvcJig.CreateAndAwaitClonePVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, srcPvc.Name, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			clonePod := pvcJig.NewPodForCSIClone("app2", f.Namespace.Name, clonePvc.Name, setupF.AdLabel)
			pvcJig.CheckFileExists(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, clonePod, "/data", "testdata.txt")
		})
	})
})

var _ = Describe("CSI Volume Cloning with different storage classes", func() {
	f := framework.NewBackupFramework("csi-cloning-sc")
	Context("[cloudprovider][storage][csi][cloning]", func() {
		It("Should be able to create a clone with different storage class than the source volume - different vpusPerGB", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-volume-sc-test")

			scParameters1 := map[string]string{
				framework.AttachmentType: framework.AttachmentTypeISCSI,
				csi_util.VpusPerGB:       "20",
			}
			scName1 := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", scParameters1, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			srcPvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName1, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("app1", f.Namespace.Name, srcPvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) // Wait for data to be written to source PV

			scParameters2 := map[string]string{
				framework.AttachmentType: framework.AttachmentTypeISCSI,
				csi_util.VpusPerGB:       "0",
			}
			scName2 := f.CreateStorageClassOrFail(framework.ClassOCILowCost, "blockvolume.csi.oraclecloud.com", scParameters2, pvcJig.Labels, "Immediate", true, "Delete", nil)
			clonePvc := pvcJig.CreateAndAwaitClonePVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName2, srcPvc.Name, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			clonePod := pvcJig.NewPodForCSIClone("app2", f.Namespace.Name, clonePvc.Name, setupF.AdLabel)
			pvcJig.CheckFileExists(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckAttachmentTypeAndEncryptionType(f.Client.Compute(), clonePvc.Name, f.Namespace.Name, clonePod, framework.AttachmentTypeISCSI)
			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, clonePvc.Namespace, clonePvc.Name, csi_util.LowCostPerformanceOption)
		})
	})
})

var _ = Describe("CSI Volume Cloning with static source Volume", func() {
	f := framework.NewBackupFramework("csi-static-cloning")
	Context("[cloudprovider][storage][csi][static][cloning]", func() {
		It("Create Clone PVC from a statically created source volume", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-static-cloning-test")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com",
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
			srcPvc, volumeId := pvcJig.CreateAndAwaitStaticPVCOrFailCSI(f.BlockStorageClient, f.Namespace.Name, framework.MinVolumeBlock, 10, scName, setupF.AdLocation, compartmentId, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			f.VolumeIds = append(f.VolumeIds, srcPvc.Spec.VolumeName)
			pvcJig.NewPodForCSI("app1", f.Namespace.Name, srcPvc.Name, setupF.AdLabel)

			time.Sleep(90 * time.Second) //waiting for pod to up and running
			pvcJig.CheckVolumeCapacity("50Gi", srcPvc.Name, f.Namespace.Name)

			clonePvc := pvcJig.CreateAndAwaitClonePVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, srcPvc.Name, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			clonePod := pvcJig.NewPodForCSIClone("app2", f.Namespace.Name, clonePvc.Name, setupF.AdLabel)
			pvcJig.CheckFileExists(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, clonePod, "/data", "testdata.txt")

			f.VolumeIds = append(f.VolumeIds, volumeId)
		})
	})
})

var _ = Describe("CSI Volume Cloning Performance Level", func() {
	f := framework.NewBackupFramework("csi-cloning-perf")
	Context("[cloudprovider][storage][csi][cloning]", func() {
		It("Create high performance clone from a low performance source volume", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-cloning-perf-test")

			scName1 := f.CreateStorageClassOrFail(framework.ClassOCILowCost, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "0"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			srcPvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName1, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSI("low-cost-source-pvc-app", f.Namespace.Name, srcPvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, srcPvc.Namespace, srcPvc.Name, csi_util.LowCostPerformanceOption)

			scName2 := f.CreateStorageClassOrFail(framework.ClassOCIHigh, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			clonePvc := pvcJig.CreateAndAwaitClonePVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName2, srcPvc.Name, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			clonePod := pvcJig.NewPodForCSIClone("high-cost-clone-pvc-app", f.Namespace.Name, clonePvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckFileExists(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, clonePod, "/data", "testdata.txt")
			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, clonePvc.Namespace, clonePvc.Name, csi_util.HigherPerformanceOption)

			f.VolumeIds = append(f.VolumeIds, srcPvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCILowCost)
		})
	})
})
