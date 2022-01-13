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
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/core"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var _ = Describe("flex volume driver", func() {
	f := framework.NewBackupFramework("fvd")
	Context("[cloudprovider][storage][fvd]", func() {
		It("should be possible to mount a volume", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "fvd-e2e-tests-pvc")

			scName := f.CreateStorageClassOrFail(framework.ClassOCI, core.ProvisionerNameDefault, nil, pvcJig.Labels, "", false)

			By("Provisioning volume to mount")
			pvc := pvcJig.CreateAndAwaitPVCOrFail(f.Namespace.Name, framework.MinVolumeBlock, scName, setupF.AdLabel, nil)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)

			By("Creating pod to check read and write to volume")
			pvcJig.CheckVolumeMount(f.Namespace.Name, pvc)
		})
	})
})
