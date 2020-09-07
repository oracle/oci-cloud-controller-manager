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
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	"time"
)

var _ = Describe("CSI Volume Creation", func() {
	f := framework.NewDefaultFramework("csi-basic")
	Context("[cloudprovider][storage][csi]", func() {
		It("Create PVC and POD for CSI.", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil)

			pvcJig.NewPODForCSI("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel)
		})

		It("Create PVC with VolumeSize 1Gi but should use default 50Gi", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-1gi")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.VolumeFss, scName, nil)

			pvcJig.NewPODForCSI("app2", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
		})

		It("Create PVC with VolumeSize 100Gi should use 100Gi", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-100gi")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MaxVolumeBlock, scName, nil)

			pvcJig.NewPODForCSI("app3", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumeCapacity("100Gi", pvc.Name, f.Namespace.Name)
		})
	})
})

var _ = Describe("CSI Static Volume Creation", func() {
	f := framework.NewBackupFramework("csi-static")
	Context("[cloudprovider][storage][csi]", func() {
		It("Static Provisioning CSI", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-static")

			scName := f.CreateStorageClassOrFail(framework.ClassOCICSI, "blockvolume.csi.oraclecloud.com", nil, pvcJig.Labels)

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
			pvc := pvcJig.CreateAndAwaitStaticPVCOrFailCSI(f.BlockStorageClient, f.Namespace.Name, framework.MinVolumeBlock, scName, setupF.AdLocation, compartmentId, nil)

			pvcJig.NewPODForCSI("app4", f.Namespace.Name, pvc.Name, setupF.AdLabel)

			time.Sleep(60 * time.Second) //waiting for pod to up and running

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
		})
	})
})
