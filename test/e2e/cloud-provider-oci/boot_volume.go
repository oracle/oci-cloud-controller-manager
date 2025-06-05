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
	"time"

	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"

	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var _ = Describe("Boot volume tests", func() {
	f := framework.NewBackupFramework("csi-basic")
	Context("[cloudprovider][storage][csi][boot-volume]", func() {
		It("Boot volume as CSI data volume", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-boot-volume")
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

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName,
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI},
				pvcJig.Labels, "WaitForFirstConsumer", false, "Retain", nil)
			opts := framework.Options{
				BlockProvisionerName: setupF.BlockProvisionerName,
			}
			pvc, bootvolumeId := pvcJig.CreateAndAwaitStaticBootVolumePVCOrFailCSI(f.ComputeClient, f.BlockStorageClient, f.Namespace.Name, compartmentId, setupF.AdLocation, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteOnce, v1.ClaimPending, opts)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)

			podName := pvcJig.NewPodForCSI("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			pvcJig.DeletePod(f.Namespace.Name, podName, 7*time.Minute)
			pvcJig.DeleteBootVolume(f.BlockStorageClient, bootvolumeId, 5*time.Minute)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})

var _ = Describe("Boot volume gating test", func() {
	f := framework.NewBackupFramework("csi-basic")
	Context("[cloudprovider][storage][csi][boot-volume]", func() {
		It("Attach boot volume fails with volumeMode set to Filesystem", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-boot-vol-e2e-tests")
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

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, "blockvolume.csi.oraclecloud.com",
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI},
				pvcJig.Labels, "WaitForFirstConsumer", false, "Retain", nil)
			pvc, bootvolumeId := pvcJig.CreateAndAwaitStaticBootVolumePVCOrFailCSI(f.ComputeClient, f.BlockStorageClient, f.Namespace.Name, compartmentId, setupF.AdLocation, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeFilesystem, v1.ReadWriteOnce, v1.ClaimPending)
			pvcJig.NewPodForCSIWithoutWait("app1", f.Namespace.Name, pvc.Name, setupF.AdLabel)
			err := pvcJig.WaitTimeoutForPodRunningInNamespace("app1", f.Namespace.Name, 7*time.Minute)
			if err == nil {
				framework.Failf("Pod went to running state for gated condition")
			}

			pvcJig.DeletePod(f.Namespace.Name, "app1", 7*time.Minute)
			pvcJig.DeleteBootVolume(f.BlockStorageClient, bootvolumeId, 5*time.Minute)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})
