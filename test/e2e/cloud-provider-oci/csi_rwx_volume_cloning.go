package e2e

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("CSI RWX Block Volume Creation with PVC datasource", func() {
	f := framework.NewDefaultFramework("csi-volume-cloning")
	Context("[cloudprovider][storage][csi][cloning][raw-block][rwx]", func() {
		It("Check RWX Funtionality with Cloned raw block PVC", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-cloning-basic")

			nodes := pvcJig.ListSchedulableNodesInAD(setupF.AdLabel)
			if len(nodes) < 2 {
				Skip(fmt.Sprintf("at least 2 schedulable nodes in a AD is required to test MULTI_NODE %s", f.Namespace.Name))
			}

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			srcPvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			srcPod := pvcJig.NewPodForCSI("app1", f.Namespace.Name, srcPvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			clonePvc := pvcJig.CreateAndAwaitClonePVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, srcPvc.Name, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			var clonePodList []string
			// schedule a pod on each available node
			for i := range nodes {
				clonePod := pvcJig.NewPodForCSIwAntiAffinity(fmt.Sprintf("pod-%d", i), f.Namespace.Name, clonePvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)
				clonePodList = append(clonePodList, clonePod)

				pvcJig.CheckDataInBlockDevice(f.Namespace.Name, clonePod, "Hello World")
				pvcJig.ExtractDataFromBlockDevice(f.Namespace.Name, clonePod, "/dev/xvda", "/tmp/testdata.txt")
				pvcJig.CheckFileCorruption(f.Namespace.Name, clonePod, "/tmp", "testdata.txt")
			}

			// clean up
			pvcJig.DeleteAndAwaitPod(f.Namespace.Name, srcPod)
			pvcJig.DeleteAndAwaitPVC(f.Namespace.Name, srcPvc.Name)

			for _, pod := range clonePodList {
				pvcJig.DeleteAndAwaitPod(f.Namespace.Name, pod)
			}
			f.VolumeIds = append(f.VolumeIds, srcPvc.Spec.VolumeName, clonePvc.Spec.VolumeName)
			pvcJig.DeleteAndAwaitPVC(f.Namespace.Name, clonePvc.Name)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})

		It("Check RWO Funtionality with Cloned RWO raw block PVC from RWX source PVC", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-cloning-basic")

			nodes := pvcJig.ListSchedulableNodesInAD(setupF.AdLabel)
			if len(nodes) < 2 {
				Skip(fmt.Sprintf("at least 2 schedulable nodes in a AD is required to test MULTI_NODE %s", f.Namespace.Name))
			}

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			srcPvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			srcPod := pvcJig.NewPodForCSI("app1", f.Namespace.Name, srcPvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			clonePvc := pvcJig.CreateAndAwaitClonePVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, srcPvc.Name, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending)
			clonePod := pvcJig.NewPodForCSIClone("app2", f.Namespace.Name, clonePvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			pvcJig.CheckDataInBlockDevice(f.Namespace.Name, clonePod, "Hello World")
			pvcJig.ExtractDataFromBlockDevice(f.Namespace.Name, clonePod, "/dev/xvda", "/tmp/testdata.txt")
			pvcJig.CheckFileCorruption(f.Namespace.Name, clonePod, "/tmp", "testdata.txt")

			// clean up
			pvcJig.DeleteAndAwaitPod(f.Namespace.Name, srcPod)
			pvcJig.DeleteAndAwaitPVC(f.Namespace.Name, srcPvc.Name)
			pvcJig.DeleteAndAwaitPod(f.Namespace.Name, clonePod)
			pvcJig.DeleteAndAwaitPVC(f.Namespace.Name, clonePvc.Name)
			f.VolumeIds = append(f.VolumeIds, srcPvc.Spec.VolumeName, clonePvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})
