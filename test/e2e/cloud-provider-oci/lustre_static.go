// Copyright 2021 Oracle and/or its affiliates. All rights reserved.
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
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Lustre Static", func() {
	f := framework.NewDefaultFramework("lustre-static-e2e")
	Context("[lustre]", func() {

		It("Multiple Pods should be able consume same PVC and read, write to same file", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-lustre-e2e-test")
			pv := pvcJig.CreatePVorFailLustre(f.Namespace.Name, setupF.LustreVolumeHandle, []string{})
			pvc := pvcJig.CreateAndAwaitPVCOrFailStaticLustre(f.Namespace.Name, pv.Name, "50Gi", nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.CheckMultiplePodReadWrite(f.Namespace.Name, pvc.Name, false)
		})

		It("Multiple CSI Drivers (BV, FSS, Lustre) should work in same cluster and be able to handle mount, unmounts", func() {

			//BV
			bvPVCJig := framework.NewPVCTestJig(f.ClientSet, "csi-bv-e2e-test")
			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, bvPVCJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			bvPVC := bvPVCJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)

			//FSS
			fssPVCJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-e2e-test")
			fssPV := fssPVCJig.CreatePVorFailFSS(f.Namespace.Name, setupF.VolumeHandle, "false", []string{})
			fssPVC := fssPVCJig.CreateAndAwaitPVCOrFailStaticFSS(f.Namespace.Name, fssPV.Name, "50Gi", nil)
			f.VolumeIds = append(f.VolumeIds, fssPVC.Spec.VolumeName)

			//LUSTRE
			lusterPVCJig := framework.NewPVCTestJig(f.ClientSet, "csi-lustre-e2e-test")
			lustrePV := lusterPVCJig.CreatePVorFailLustre(f.Namespace.Name, setupF.LustreVolumeHandle, []string{})
			lustrePVC := lusterPVCJig.CreateAndAwaitPVCOrFailStaticLustre(f.Namespace.Name, lustrePV.Name, "50Gi", nil)
			f.VolumeIds = append(f.VolumeIds, lustrePVC.Spec.VolumeName)

			bvPVCJig.CheckSinglePodReadWrite(f.Namespace.Name, bvPVC.Name, false, []string{})
			fssPVCJig.CheckSinglePodReadWrite(f.Namespace.Name, fssPVC.Name, false, []string{})
			lusterPVCJig.CheckSinglePodReadWrite(f.Namespace.Name, lustrePVC.Name, false, []string{})

		})
	})
})
