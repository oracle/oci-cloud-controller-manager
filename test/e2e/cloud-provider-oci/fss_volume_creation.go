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

//var _ = Describe("FSS Volume Creation", func() {
//	f := framework.NewDefaultFramework("fss-volume")
//	Context("[fss]", func() {
//		It("should be possible to create a PVC for a FSS with the MountTarget specified via the StorageClass", func() {
//			pvcJig := framework.NewPVCTestJig(f.ClientSet, "volume-provisioner-e2e-tests-pvc")
//
//			scName := f.CreateStorageClassOrFail(
//				framework.ClassOCIMntFss,
//				core.ProvisionerNameFss,
//				map[string]string{fss.MntTargetID: setupF.MntTargetOcid},
//				pvcJig.Labels)
//
//			By("Creating PVC that will dynamically provision a FSS")
//			pvc := pvcJig.CreateAndAwaitPVCOrFail(
//				f.Namespace.Name,
//				framework.VolumeFss,
//				scName,
//				setupF.AdLabel,
//				nil)
//
//			By("Creating a Pod and waiting till attaches to the volume")
//			pvcJig.CheckVolumeReadWrite(f.Namespace.Name, pvc)
//		})
//
//		It("should be possible to create a PVC for a FSS with the MountTarget specified via an annotation on the PVC", func() {
//			pvcJig := framework.NewPVCTestJig(f.ClientSet, "volume-provisioner-e2e-tests-pvc")
//
//			scName := f.CreateStorageClassOrFail(
//				framework.ClassOCISubnetFss,
//				core.ProvisionerNameFss,
//				map[string]string{},
//				pvcJig.Labels)
//
//			By("Creating PVC that will dynamically provision a FSS")
//			pvc := pvcJig.CreateAndAwaitPVCOrFail(
//				f.Namespace.Name,
//				framework.VolumeFss,
//				scName,
//				setupF.AdLabel,
//				func(pvc *v1.PersistentVolumeClaim) {
//					pvc.Annotations = map[string]string{
//						fss.AnnotationMountTargetID: setupF.MntTargetOcid,
//					}
//				})
//
//			By("Creating a Pod and waiting till attaches to the volume")
//			pvcJig.CheckVolumeReadWrite(f.Namespace.Name, pvc)
//		})
//	})
//})
