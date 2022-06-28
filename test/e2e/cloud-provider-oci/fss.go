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
	"context"
	. "github.com/onsi/ginkgo"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Basic FSS test", func() {
	f := framework.NewDefaultFramework("fss-basic")
	Context("[cloudprovider][storage][csi][fss]", func() {
		It("Create PVC and POD for CSI-FSS", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-e2e-test")
			pv := pvcJig.CreatePVorFailFSS(f.Namespace.Name, setupF.VolumeHandle, "false")
			pvc := pvcJig.CreateAndAwaitPVCOrFailFSS(f.Namespace.Name, pv.Name, "50Gi", nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false)
		})
	})
})

var _ = Describe("FSS in-transit encryption test", func() {
	f := framework.NewDefaultFramework("fss-basic")
	Context("[cloudprovider][storage][csi][fss]", func() {
		It("Create PVC and POD for FSS in-transit encryption", func() {
			TestEncryptionType(f)
		})
	})
})

func TestEncryptionType(f *framework.CloudProviderFramework) {
	checkNodeAvailability(f)
	pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-e2e-test-intransit")
	pv := pvcJig.CreatePVorFailFSS(f.Namespace.Name, setupF.VolumeHandle, "true")
	pvc := pvcJig.CreateAndAwaitPVCOrFailFSS(f.Namespace.Name, pv.Name, "50Gi", nil)
	f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
	pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, true)
}

var _ = Describe("Multiple Pods FSS test", func() {
	f := framework.NewDefaultFramework("multiple-pod")
	Context("[cloudprovider][storage][csi][fss]", func() {
		It("Multiple Pods should be able to read write same file", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-e2e-test")
			pv := pvcJig.CreatePVorFailFSS(f.Namespace.Name, setupF.VolumeHandle, "false")
			pvc := pvcJig.CreateAndAwaitPVCOrFailFSS(f.Namespace.Name, pv.Name, "50Gi", nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.CheckMultiplePodReadWrite(f.Namespace.Name, pvc.Name, false)
		})

		It("Multiple Pods should be able to read write same file with InTransit encryption enabled", func() {
			checkNodeAvailability(f)
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-e2e-test")
			pv := pvcJig.CreatePVorFailFSS(f.Namespace.Name, setupF.VolumeHandle, "true")
			pvc := pvcJig.CreateAndAwaitPVCOrFailFSS(f.Namespace.Name, pv.Name, "50Gi", nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.CheckMultiplePodReadWrite(f.Namespace.Name, pvc.Name, true)
		})
	})
})

func checkNodeAvailability(f *framework.CloudProviderFramework) {
	pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-e2e-test")
	nodeList, err := pvcJig.KubeClient.CoreV1().Nodes().List(context.Background(), v12.ListOptions{LabelSelector: "oke.oraclecloud.com/e2e.oci-fss-util"})
	if err != nil {
		framework.Logf("Error getting applicable nodes: %v", err)
	}
	nodesAvailable := false
	for _, node := range nodeList.Items {
		if node.Spec.Unschedulable == false {
			nodesAvailable = true
			break
		}
	}
	if !nodesAvailable {
		Skip("Skipping test due to non-availability of nodes with label \"oke.oraclecloud.com/e2e.oci-fss-util\"")
	}
}
