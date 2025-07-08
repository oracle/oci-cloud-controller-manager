// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
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

// To be added to csi_colume_creation.go when complete

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var _ = Describe("CSI RWX Raw Block Volume Creation - Immediate Volume Binding", func() {
	f := framework.NewDefaultFramework("csi-immediate")
	Context("[cloudprovider][storage][csi][raw-block][rwx]", func() {
		It("Create PVC without pod and wait to be bound.", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-immediate-bind")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "Immediate", true, "Delete", nil)
			pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimBound)
			err := f.DeleteStorageClass(f.Namespace.Name)
			if err != nil {
				Fail(fmt.Sprintf("deleting storage class failed %s", f.Namespace.Name))
			}
		})
	})
})

var _ = Describe("CSI RWX Raw Block Volume Creation", func() {
	f := framework.NewDefaultFramework("csi-basic")
	Context("[cloudprovider][storage][csi][system-tags][raw-block][rwx]", func() {
		It("Create RWX raw block PVC and POD for CSI.", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests")
			ctx := context.TODO()
			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			pvcJig.NewPodForCSI("app1", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			volumeName := pvcJig.GetVolumeNameFromPVC(pvc.GetName(), f.Namespace.Name)
			compartmentId := f.GetCompartmentId(*setupF)
			// read created BV
			volumes, err := f.Client.BlockStorage().GetVolumesByName(ctx, volumeName, compartmentId)
			framework.ExpectNoError(err)
			// volume name duplicate should not exist
			for _, volume := range volumes {
				framework.Logf("volume details %v :", volume)
				//framework.Logf("cluster ocid from setup is %s", setupF.ClusterOcid)
				if setupF.AddOkeSystemTags && !framework.HasOkeSystemTags(volume.SystemTags) && setupF.CustomDriverHandle == "" {
					framework.Failf("the resource %s is expected to have oke system tags", *volume.Id)
				}
			}
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})

		It("Create RWX raw block PVC with VolumeSize 1Gi but should use default 50Gi", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-1gi")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.VolumeFss, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			pvcJig.NewPodForCSI("app2", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
		})

		It("Create RWX raw block PVC with VolumeSize 100Gi should use 100Gi", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-100gi")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MaxVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			pvcJig.NewPodForCSI("app3", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			pvcJig.CheckVolumeCapacity("100Gi", pvc.Name, f.Namespace.Name)
		})

		It("Data should persist on CSI RWX raw block volume on pod restart", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-pod-restart-data-persistence")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)

			pvcJig.CheckDataPersistenceForRawBlockVolumeWithDeployment(pvc.Name, f.Namespace.Name)
		})
	})
})

var _ = Describe("CSI RWX Raw Block Volume MULTI_NODE", func() {
	f := framework.NewDefaultFramework("csi-basic")
	Context("[cloudprovider][storage][csi][raw-block][rwx]", func() {
		It("Create RWX raw block PVC, schedule a pod on each worker node", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests")

			nodes := pvcJig.ListSchedulableNodesInAD(setupF.AdLabel)
			if len(nodes) < 2 {
				Skip(fmt.Sprintf("at least 2 schedulable nodes in %s (found %d) are required to test MULTI_NODE - %s", setupF.AdLabel, len(nodes), f.Namespace.Name))
			}

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)

			// schedule a pod on each available node
			for i := range nodes {
				pvcJig.NewPodForCSIwAntiAffinity(fmt.Sprintf("pod-%d", i), f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)
			}
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})

		It("Create RWX raw block PVC, schedule a pod on each worker node, then attempt to schedule another", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests")

			nodes := pvcJig.ListSchedulableNodesInAD(setupF.AdLabel)
			if len(nodes) < 2 {
				Skip(fmt.Sprintf("at least 2 schedulable nodes in %s (found %d) are required to test MULTI_NODE - %s", setupF.AdLabel, len(nodes), f.Namespace.Name))
			}

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)

			// schedule a pod on each available node
			for i := range nodes {
				pvcJig.NewPodForCSIwAntiAffinity(fmt.Sprintf("pod-%d", i), f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)
			}
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)

			// TODO: try to schedule another pod (expect failure)
		})

		It("Create RWX raw block PVC, schedule a pod on each worker node, delete one of the pods", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests")

			nodes := pvcJig.ListSchedulableNodesInAD(setupF.AdLabel)
			if len(nodes) < 2 {
				Skip(fmt.Sprintf("at least 2 schedulable nodes in %s (found %d) are required to test MULTI_NODE - %s", setupF.AdLabel, len(nodes), f.Namespace.Name))
			}

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)

			pods := []string{}

			// schedule a pod on each available node
			for i := range nodes {
				pods = append(pods, pvcJig.NewPodForCSIwAntiAffinity(fmt.Sprintf("pod-%d", i), f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock))
			}

			// delete one pod of many
			pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, pods[0])

			// expect volume and remaining pods to remain functional
			pvcJig.CheckVolumeCapacity(framework.MinVolumeBlock, pvc.Name, f.Namespace.Name)
			pvcJig.CheckFileExists(f.Namespace.Name, pods[1], "/dev", "xvda")
			pvcJig.CheckExpandedRawBlockVolumeReadWrite(f.Namespace.Name, pods[1])
			pvcJig.ExtractDataFromBlockDevice(f.Namespace.Name, pods[1], "/dev/xvda", "/tmp/testdata.txt")
			// pvcJig.CheckFileCorruption(f.Namespace.Name, pods[1], "/tmp", "testdata.txt")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})
	})
})

var _ = Describe("CSI Raw Block Volume Expansion iSCSI", func() {
	f := framework.NewDefaultFramework("csi-expansion")
	Context("[cloudprovider][storage][csi][expand][iSCSI][raw-block][rwx]", func() {
		It("Expand raw block PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for iSCSI volumes", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-iscsi")

			nodes := pvcJig.ListSchedulableNodesInAD(setupF.AdLabel)
			if len(nodes) < 2 {
				Skip(fmt.Sprintf("at least 2 schedulable nodes in %s (found %d) are required to test MULTI_NODE - %s", setupF.AdLabel, len(nodes), f.Namespace.Name))
			}

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName,
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)

			podNameA := pvcJig.NewPodForCSIwAntiAffinity("expanded-pvc-pod-1", f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)
			podNameB := pvcJig.NewPodForCSIwAntiAffinity("expanded-pvc-pod-2", f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			pvcJig.CheckVolumeCapacity(size, expandedPvc.Name, f.Namespace.Name)

			pvcJig.CheckFileExists(f.Namespace.Name, podNameA, "/dev", "xvda")
			pvcJig.CheckExpandedRawBlockVolumeReadWrite(f.Namespace.Name, podNameA)
			pvcJig.ExtractDataFromBlockDevice(f.Namespace.Name, podNameA, "/dev/xvda", "/tmp/testdata.txt")
			// pvcJig.CheckFileCorruption(f.Namespace.Name, podNameA, "/tmp", "testdata.txt")

			pvcJig.CheckFileExists(f.Namespace.Name, podNameB, "/dev", "xvda")
			pvcJig.CheckExpandedRawBlockVolumeReadWrite(f.Namespace.Name, podNameB)
			pvcJig.ExtractDataFromBlockDevice(f.Namespace.Name, podNameB, "/dev/xvda", "/tmp/testdata.txt")
			// pvcJig.CheckFileCorruption(f.Namespace.Name, podNameA, "/tmp", "testdata.txt")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})
	})
})

var _ = Describe("CSI Raw Block Volume Expansion Paravirtualized", func() {
	f := framework.NewDefaultFramework("csi-expansion")
	Context("[cloudprovider][storage][csi][expand][paravirtualized][raw-block][rwx]", func() {
		It("Expand raw block PVC VolumeSize from 50Gi to 100Gi and asserts size, file existence and file corruptions for paravirtualized volumes", func() {
			var size = "100Gi"
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-resizer-pvc-expand-to-100gi-paravirtualized")

			nodes := pvcJig.ListSchedulableNodesInAD(setupF.AdLabel)
			if len(nodes) < 2 {
				Skip(fmt.Sprintf("at least 2 schedulable nodes in %s (found %d) are required to test MULTI_NODE - %s", setupF.AdLabel, len(nodes), f.Namespace.Name))
			}

			scParameter := map[string]string{
				framework.KmsKey:         setupF.CMEKKMSKey,
				framework.AttachmentType: framework.AttachmentTypeParavirtualized,
			}
			scName := f.CreateStorageClassOrFail(f.Namespace.Name,
				setupF.BlockProvisionerName, scParameter, pvcJig.Labels,
				"WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)

			podNameA := pvcJig.NewPodForCSIwAntiAffinity("expanded-pvc-pod-1", f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)
			podNameB := pvcJig.NewPodForCSIwAntiAffinity("expanded-pvc-pod-2", f.Namespace.Name, pvc.Name, setupF.AdLabel, v1.PersistentVolumeBlock)

			expandedPvc := pvcJig.UpdateAndAwaitPVCOrFailCSI(pvc, pvc.Namespace, size, nil)

			pvcJig.CheckVolumeCapacity(size, expandedPvc.Name, f.Namespace.Name)

			pvcJig.CheckFileExists(f.Namespace.Name, podNameA, "/dev", "xvda")
			pvcJig.CheckExpandedRawBlockVolumeReadWrite(f.Namespace.Name, podNameA)
			pvcJig.ExtractDataFromBlockDevice(f.Namespace.Name, podNameA, "/dev/xvda", "/tmp/testdata.txt")
			// pvcJig.CheckFileCorruption(f.Namespace.Name, podNameA, "/tmp", "testdata.txt")

			pvcJig.CheckFileExists(f.Namespace.Name, podNameB, "/dev", "xvda")
			pvcJig.CheckExpandedRawBlockVolumeReadWrite(f.Namespace.Name, podNameB)
			pvcJig.ExtractDataFromBlockDevice(f.Namespace.Name, podNameB, "/dev/xvda", "/tmp/testdata.txt")
			// pvcJig.CheckFileCorruption(f.Namespace.Name, podNameA, "/tmp", "testdata.txt")
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})
	})
})

var _ = Describe("CSI Raw Block Volume Performance Level", func() {
	f := framework.NewBackupFramework("csi-perf-level")
	Context("[cloudprovider][storage][csi][perf][iSCSI][raw-block][rwx]", func() {
		It("Create CSI raw block volume with Performance Level as Low Cost", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-lowcost")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName,
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "0"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			pvcJig.NewPodForCSI("low-cost-pvc-app", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.LowCostPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Create CSI raw block volume with no Performance Level and verify default", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-default")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName, nil, pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			pvcJig.NewPodForCSI("default-pvc-app", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.BalancedPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
		})
		It("Create CSI raw block volume with Performance Level as High", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-iscsi-high")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName,
				map[string]string{framework.AttachmentType: framework.AttachmentTypeISCSI, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			podName := pvcJig.NewPodForCSI("high-perf-pvc-app", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.HigherPerformanceOption)
			pvcJig.CheckISCSIQueueDepthOnNode(f.Namespace.Name, podName)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
	Context("[cloudprovider][storage][csi][perf][paravirtualized][raw-block][rwx]", func() {
		It("Create CSI raw block volume with Performance Level as Low Cost", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-lowcost")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName,
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "0"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			pvcJig.NewPodForCSI("low-cost-pvc-app", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.LowCostPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Create CSI raw block volume with no Performance Level and verify default", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-balanced")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName,
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "10"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			pvcJig.NewPodForCSI("default-pvc-app", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.BalancedPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
		It("Create CSI raw block volume with Performance Level as High", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-paravirtual-high")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName,
				map[string]string{framework.AttachmentType: framework.AttachmentTypeParavirtualized, csi_util.VpusPerGB: "20"},
				pvcJig.Labels, "WaitForFirstConsumer", true, "Delete", nil)
			pvc := pvcJig.CreateAndAwaitPVCOrFailCSI(f.Namespace.Name, framework.MinVolumeBlock, scName, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending)
			pvcJig.NewPodForCSI("high-perf-pvc-app", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			pvcJig.CheckVolumePerformanceLevel(f.BlockStorageClient, pvc.Namespace, pvc.Name, csi_util.HigherPerformanceOption)
			f.VolumeIds = append(f.VolumeIds, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
	Context("[cloudprovider][storage][csi][perf][static][raw-block][rwx]", func() {
		It("High Performance CSI raw block volume Static Provisioning", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-perf-static-high")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName,
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
			opts := framework.Options{
				BlockProvisionerName: setupF.BlockProvisionerName,
			}
			pvc, volumeId := pvcJig.CreateAndAwaitStaticPVCOrFailCSI(f.BlockStorageClient, f.Namespace.Name, framework.MinVolumeBlock, csi_util.HigherPerformanceOption, scName, setupF.AdLocation, compartmentId, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending, opts)
			podName := pvcJig.NewPodForCSI("app4", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
			pvcJig.CheckISCSIQueueDepthOnNode(pvc.Namespace, podName)
			f.VolumeIds = append(f.VolumeIds, volumeId, pvc.Spec.VolumeName)
			_ = f.DeleteStorageClass(f.Namespace.Name)
		})
	})
})

var _ = Describe("CSI Static Raw Block Volume Creation", func() {
	f := framework.NewBackupFramework("csi-static")
	Context("[cloudprovider][storage][csi][static][raw-block][rwx]", func() {
		It("Static Provisioning of a raw block CSI", func() {
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-provisioner-e2e-tests-pvc-with-static")

			scName := f.CreateStorageClassOrFail(f.Namespace.Name, setupF.BlockProvisionerName,
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
			opts := framework.Options{
				BlockProvisionerName: setupF.BlockProvisionerName,
			}
			pvc, volumeId := pvcJig.CreateAndAwaitStaticPVCOrFailCSI(f.BlockStorageClient, f.Namespace.Name, framework.MinVolumeBlock, 10, scName, setupF.AdLocation, compartmentId, nil, v1.PersistentVolumeBlock, v1.ReadWriteMany, v1.ClaimPending, opts)
			pvcJig.NewPodForCSI("app4", f.Namespace.Name, pvc.Name, "", v1.PersistentVolumeBlock)

			pvcObj := pvcJig.GetPVCByName(pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, pvcObj.Spec.VolumeName)
			pvcJig.CheckVolumeCapacity("50Gi", pvc.Name, f.Namespace.Name)
			f.VolumeIds = append(f.VolumeIds, volumeId)
		})
	})
})

var _ = Describe("CSI CMEK,PV attachment and in-transit encryption test", func() {
	f := framework.NewDefaultFramework("csi-basic")
	Context("[cloudprovider][storage][csi][cmek][paravirtualized][raw-block][rwx]", func() {
		It("Create raw block PVC and POD for CSI with CMEK,PV attachment and in-transit encryption", func() {
			TestCMEKAttachmentTypeAndEncryptionType(f, framework.AttachmentTypeParavirtualized, true, v1.ReadWriteMany)
		})
	})

	Context("[cloudprovider][storage][csi][cmek][iscsi][raw-block][rwx]", func() {
		It("Create raw block PVC and POD for CSI with CMEK,ISCSI attachment and in-transit encryption", func() {
			TestCMEKAttachmentTypeAndEncryptionType(f, framework.AttachmentTypeISCSI, true, v1.ReadWriteMany)
		})
	})

})
