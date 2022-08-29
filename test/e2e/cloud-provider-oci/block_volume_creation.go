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
)

var _ = Describe("Block Volume Creation", func() {
	f := framework.NewDefaultFramework("block-basic")
	Context("[cloudprovider][storage][fvp]", func() {
		It("Should be possible to create a persistent volume claim for a block storage (PVC)", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "volume-provisioner-e2e-tests-pvc")

			scName := f.CreateStorageClassOrFail(framework.ClassOCI, core.ProvisionerNameDefault, nil, pvcJig.Labels, "", false)
			pvc := pvcJig.CreateAndAwaitPVCOrFail(f.Namespace.Name, framework.MinVolumeBlock, scName, setupF.AdLabel, nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})

		It("Should be possible to create a persistent volume claim (PVC) for a block storage of Ext3 file system ", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "volume-provisioner-e2e-tests-pvc")

			scName := f.CreateStorageClassOrFail(framework.ClassOCIExt3, core.ProvisionerNameDefault, map[string]string{block.FSType: "ext3"}, pvcJig.Labels, "", false)
			pvc := pvcJig.CreateAndAwaitPVCOrFail(f.Namespace.Name, framework.MinVolumeBlock, scName, setupF.AdLabel, nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(framework.ClassOCIExt3)
		})

		It("Should be possible to create a persistent volume claim (PVC) for a block storage with no AD specified ", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "volume-provisioner-e2e-tests-pvc")

			scName := f.CreateStorageClassOrFail(framework.ClassOCI, core.ProvisionerNameDefault, nil, pvcJig.Labels, "", false)
			pvc := pvcJig.CreateAndAwaitPVCOrFail(f.Namespace.Name, framework.MinVolumeBlock, scName, "", nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})
	})
})
