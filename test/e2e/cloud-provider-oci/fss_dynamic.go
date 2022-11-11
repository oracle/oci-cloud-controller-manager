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
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	defaultExportOptionsJsonString = "[{\"source\":\"10.0.0.0/16\",\"requirePrivilegedSourcePort\":true,\"access\":\"READ_WRITE\",\"identitySquash\":\"NONE\",\"anonymousUid\":0,\"anonymousGid\":0}]"
)

var _ = Describe("Dynamic FSS test in cluster compartment", func() {
	f := framework.NewDefaultFramework("fss-dynamic")

	Context("[cloudprovider][storage][csi][fss][mtexist]", func() {
		It("Basic Create PVC and POD for CSI-FSS", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with exportPath", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["exportPath"] = "/csi-fss-e2e-export-path-mt-exist-in-compartment"
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with exportPath and exportOptions", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["exportPath"] = "/csi-fss-e2e-export-path-export-options-mt-exist-in-compartment"
			scParameters["exportOptions"] = defaultExportOptionsJsonString
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with kmsKey", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["kmsKey"] = setupF.CMEKKMSKey
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with in-transit encryption", func() {
			checkNodeAvailability(f)
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["encryptInTransit"] = "true"
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, true, []string{})
		})
	})
	Context("[cloudprovider][storage][csi][fss][mtcreate]", func() {
		It("Basic Create PVC and POD for CSI-FSS with new mount-target creation", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetSubnetOcid": setupF.MntTargetSubnetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with exportPath and with new mount-target creation", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetSubnetOcid": setupF.MntTargetSubnetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["exportPath"] = "/csi-fss-e2e-export-path-mt-create-in-compartment"
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with exportPath and exportOptions and with new mount-target creation", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetSubnetOcid": setupF.MntTargetSubnetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["exportPath"] = "/csi-fss-e2e-export-path-export-options-mt-create-in-compartment"
			scParameters["exportOptions"] = defaultExportOptionsJsonString
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with kmsKey and with new mount-target creation", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetSubnetOcid": setupF.MntTargetSubnetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["kmsKey"] = setupF.CMEKKMSKey
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with in-transit encryption and with new mount-target creation", func() {
			checkNodeAvailability(f)
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetSubnetOcid": setupF.MntTargetSubnetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["encryptInTransit"] = "true"
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, true, []string{})
		})
	})
})

var _ = Describe("Dynamic FSS test in different compartment", func() {
	f := framework.NewDefaultFramework("fss-dynamic")

	Context("[cloudprovider][storage][csi][fss][mtexist]", func() {
		It("Basic Create PVC and POD for CSI-FSS with file-system compartment set", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with exportPath with file-system compartment set", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["exportPath"] = "/csi-fss-e2e-export-path-mt-exist-diff-compartment"
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with exportPath and exportOptions with file-system compartment set", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["exportPath"] = "/csi-fss-e2e-export-path-export-options-mt-exist-diff-compartment"
			scParameters["exportOptions"] = defaultExportOptionsJsonString
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with kmsKey and with file-system compartment set", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["kmsKey"] = setupF.CMEKKMSKey
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with in-transit encryption and with file-system compartment set", func() {
			checkNodeAvailability(f)
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["encryptInTransit"] = "true"
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, true, []string{})
		})
	})
	Context("[cloudprovider][storage][csi][fss][mtcreate]", func() {
		It("Basic Create PVC and POD for CSI-FSS with file-system compartment set and with new mount-target creation", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetSubnetOcid": setupF.MntTargetSubnetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with exportPath and with file-system compartment set and with new mount-target creation", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetSubnetOcid": setupF.MntTargetSubnetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["exportPath"] = "/csi-fss-e2e-export-path-mt-create-diff-compartment"
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with exportPath and exportOptions and with file-system compartment set and with new mount-target creation", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetSubnetOcid": setupF.MntTargetSubnetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["exportPath"] = "/csi-fss-e2e-export-path-export-options-mt-create-diff-compartment"
			scParameters["exportOptions"] = defaultExportOptionsJsonString
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with kmsKey and with file-system compartment set and with new mount-target creation", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetSubnetOcid": setupF.MntTargetSubnetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["kmsKey"] = setupF.CMEKKMSKey
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
		It("Create PVC and POD for CSI-FSS with in-transit encryption", func() {
			checkNodeAvailability(f)
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetSubnetOcid": setupF.MntTargetSubnetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["encryptInTransit"] = "true"
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, true, []string{})
		})
	})
})

var _ = Describe("Dynamic FSS deletion test", func() {
	f := framework.NewDefaultFramework("fss-dynamic")

	Context("[cloudprovider][storage][csi][fss][mtexist]", func() {
		It("Basic Delete POD and PVC for CSI-FSS", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvcObject := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			By("Creating Pod that can create and write to the file")
			uid := uuid.NewUUID()
			fileName := fmt.Sprintf("out_%s.txt", uid)
			podName := pvcJig.NewPodForCSIFSSWrite(string(uid), f.Namespace.Name, pvcObject.Name, fileName, false)
			time.Sleep(30 * time.Second) //waiting for pod to become up and running
			pvc := pvcJig.GetPVCByName(pvcObject.Name, f.Namespace.Name)
			volumeName := pvc.Spec.VolumeName
			framework.Logf("Pod name : %s", podName)
			framework.Logf("Persistent volume name : %s", volumeName)
			pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName)
			err := pvcJig.DeletePersistentVolumeClaim(f.Namespace.Name, pvc.Name)
			if err != nil {
				framework.Failf("Failed to delete persistent volume claim: %s", err.Error())
			}
			err = pvcJig.WaitTimeoutForPVNotFound(volumeName, 10*time.Minute)
			if err != nil {
				framework.Failf("Persistent volume did not terminate : %s", err.Error())
			}
		})
		It("Test PV not deleted when reclaim policy is Retain", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Retain", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvcObject := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			By("Creating Pod that can create and write to the file")
			uid := uuid.NewUUID()
			fileName := fmt.Sprintf("out_%s.txt", uid)
			podName := pvcJig.NewPodForCSIFSSWrite(string(uid), f.Namespace.Name, pvcObject.Name, fileName, false)
			time.Sleep(30 * time.Second) //waiting for pod to become up and running
			pvc := pvcJig.GetPVCByName(pvcObject.Name, f.Namespace.Name)
			volumeName := pvc.Spec.VolumeName
			framework.Logf("Pod name : %s", podName)
			framework.Logf("Persistent volume name : %s", volumeName)
			pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName)
			err := pvcJig.DeletePersistentVolumeClaim(f.Namespace.Name, pvc.Name)
			if err != nil {
				framework.Failf("Failed to delete persistent volume claim: %s", err.Error())
			}
			time.Sleep(1 * time.Minute)
			pvExists := pvcJig.CheckPVExists(volumeName)
			if pvExists != true {
				framework.Failf("Persistent volume was deleted")
			}
			f.VolumeIds = append(f.VolumeIds, volumeName)
		})
		It("Test export is deleted in cluster compartment when export path is not set", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvcObject := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			By("Creating Pod that can create and write to the file")
			uid := uuid.NewUUID()
			fileName := fmt.Sprintf("out_%s.txt", uid)
			podName := pvcJig.NewPodForCSIFSSWrite(string(uid), f.Namespace.Name, pvcObject.Name, fileName, false)
			time.Sleep(30 * time.Second) //waiting for pod to become up and running
			pvc := pvcJig.GetPVCByName(pvcObject.Name, f.Namespace.Name)
			volumeName := pvc.Spec.VolumeName
			framework.Logf("Pod name : %s", podName)
			framework.Logf("Persistent volume name : %s", volumeName)
			fsId, err := f.GetFSIdByDisplayName(context.Background(), f.CloudProviderConfig.CompartmentID, setupF.AdLocation, volumeName)
			if err != nil {
				framework.Failf("Failed to get FS Id by display name: %s", err.Error())
			}
			framework.Logf("FS OCID : %s", fsId)
			exportSetId, err := f.GetExportsSetIdByMountTargetId(context.Background(), setupF.MntTargetOcid)
			if err != nil {
				framework.Failf("Failed to get export set Id : %s", err.Error())
			}
			framework.Logf("Export Set Id : %s", exportSetId)
			pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName)
			err = pvcJig.DeletePersistentVolumeClaim(f.Namespace.Name, pvc.Name)
			if err != nil {
				framework.Failf("Failed to delete persistent volume claim: %s", err.Error())
			}
			time.Sleep(2 * time.Minute)
			exportExists := f.CheckExportExists(context.Background(), fsId, "/"+volumeName, exportSetId)
			if exportExists {
				framework.Failf("Failed to delete export")
			}
			volumeExists := f.CheckFSVolumeExist(context.Background(), fsId)
			if volumeExists {
				framework.Failf("Failed to delete FS volume")
			}
		})
		It("Test export is deleted in cluster compartment when export path is set", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["exportPath"] = "/csi-fss-e2e-delete-export-mt-exist-in-compartment"
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvcObject := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			By("Creating Pod that can create and write to the file")
			uid := uuid.NewUUID()
			fileName := fmt.Sprintf("out_%s.txt", uid)
			podName := pvcJig.NewPodForCSIFSSWrite(string(uid), f.Namespace.Name, pvcObject.Name, fileName, false)
			time.Sleep(30 * time.Second) //waiting for pod to become up and running
			pvc := pvcJig.GetPVCByName(pvcObject.Name, f.Namespace.Name)
			volumeName := pvc.Spec.VolumeName
			framework.Logf("Pod name : %s", podName)
			framework.Logf("Persistent volume name : %s", volumeName)
			fsId, err := f.GetFSIdByDisplayName(context.Background(), f.CloudProviderConfig.CompartmentID, setupF.AdLocation, volumeName)
			if err != nil {
				framework.Failf("Failed to get FS Id by display name: %s", err.Error())
			}
			framework.Logf("FS OCID : %s", fsId)
			exportSetId, err := f.GetExportsSetIdByMountTargetId(context.Background(), setupF.MntTargetOcid)
			if err != nil {
				framework.Failf("Failed to get export set Id : %s", err.Error())
			}
			framework.Logf("Export Set Id : %s", exportSetId)
			pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName)
			err = pvcJig.DeletePersistentVolumeClaim(f.Namespace.Name, pvc.Name)
			if err != nil {
				framework.Failf("Failed to delete persistent volume claim: %s", err.Error())
			}
			time.Sleep(2 * time.Minute)
			exportExists := f.CheckExportExists(context.Background(), fsId, scParameters["exportPath"], exportSetId)
			if exportExists {
				framework.Failf("Failed to delete export")
			}
			volumeExists := f.CheckFSVolumeExist(context.Background(), fsId)
			if volumeExists {
				framework.Failf("Failed to delete FS volume")
			}
		})
		It("Test export is deleted in different compartment when export path is not set", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvcObject := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			By("Creating Pod that can create and write to the file")
			uid := uuid.NewUUID()
			fileName := fmt.Sprintf("out_%s.txt", uid)
			podName := pvcJig.NewPodForCSIFSSWrite(string(uid), f.Namespace.Name, pvcObject.Name, fileName, false)
			time.Sleep(30 * time.Second) //waiting for pod to become up and running
			pvc := pvcJig.GetPVCByName(pvcObject.Name, f.Namespace.Name)
			volumeName := pvc.Spec.VolumeName
			framework.Logf("Pod name : %s", podName)
			framework.Logf("Persistent volume name : %s", volumeName)
			fsId, err := f.GetFSIdByDisplayName(context.Background(), setupF.MntTargetCompartmentOcid, setupF.AdLocation, volumeName)
			if err != nil {
				framework.Failf("Failed to get FS Id by display name: %s", err.Error())
			}
			framework.Logf("FS OCID : %s", fsId)
			exportSetId, err := f.GetExportsSetIdByMountTargetId(context.Background(), setupF.MntTargetOcid)
			if err != nil {
				framework.Failf("Failed to get export set Id : %s", err.Error())
			}
			framework.Logf("Export Set Id : %s", exportSetId)
			pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName)
			err = pvcJig.DeletePersistentVolumeClaim(f.Namespace.Name, pvc.Name)
			if err != nil {
				framework.Failf("Failed to delete persistent volume claim: %s", err.Error())
			}
			time.Sleep(2 * time.Minute)
			exportExists := f.CheckExportExists(context.Background(), fsId, "/"+volumeName, exportSetId)
			if exportExists {
				framework.Failf("Failed to delete export")
			}
			volumeExists := f.CheckFSVolumeExist(context.Background(), fsId)
			if volumeExists {
				framework.Failf("Failed to delete FS volume")
			}
		})
		It("Test export is deleted in different compartment when export path is set", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid, "compartmentOcid": setupF.MntTargetCompartmentOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scParameters["exportPath"] = "/csi-fss-e2e-delete-export-mt-exist-diff-compartment"
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvcObject := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			By("Creating Pod that can create and write to the file")
			uid := uuid.NewUUID()
			fileName := fmt.Sprintf("out_%s.txt", uid)
			podName := pvcJig.NewPodForCSIFSSWrite(string(uid), f.Namespace.Name, pvcObject.Name, fileName, false)
			time.Sleep(30 * time.Second) //waiting for pod to become up and running
			pvc := pvcJig.GetPVCByName(pvcObject.Name, f.Namespace.Name)
			volumeName := pvc.Spec.VolumeName
			framework.Logf("Pod name : %s", podName)
			framework.Logf("Persistent volume name : %s", volumeName)
			fsId, err := f.GetFSIdByDisplayName(context.Background(), setupF.MntTargetCompartmentOcid, setupF.AdLocation, volumeName)
			if err != nil {
				framework.Failf("Failed to get FS Id by display name: %s", err.Error())
			}
			framework.Logf("FS OCID : %s", fsId)
			exportSetId, err := f.GetExportsSetIdByMountTargetId(context.Background(), setupF.MntTargetOcid)
			if err != nil {
				framework.Failf("Failed to get export set Id : %s", err.Error())
			}
			framework.Logf("Export Set Id : %s", exportSetId)
			pvcJig.DeleteAndAwaitPodOrFail(f.Namespace.Name, podName)
			err = pvcJig.DeletePersistentVolumeClaim(f.Namespace.Name, pvc.Name)
			if err != nil {
				framework.Failf("Failed to delete persistent volume claim: %s", err.Error())
			}
			time.Sleep(2 * time.Minute)
			exportExists := f.CheckExportExists(context.Background(), fsId, scParameters["exportPath"], exportSetId)
			if exportExists {
				framework.Failf("Failed to delete export")
			}
			volumeExists := f.CheckFSVolumeExist(context.Background(), fsId)
			if volumeExists {
				framework.Failf("Failed to delete FS volume")
			}
		})
	})
})

var _ = Describe("Dynamic FSS test with mount options", func() {
	f := framework.NewDefaultFramework("fss-dynamic")

	Context("[cloudprovider][storage][csi][fss][mtexist]", func() {
		It("Basic Dynamic FSS test with mount options", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			mountOptions := []string{"hard"} // TODO : Add other mount options once OKE-22216 is Done.
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "WaitForFirstConsumer", false, "Delete", mountOptions)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvc := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimPending, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvc.Name, false, []string{})
		})
	})
})

var _ = Describe("Dynamic FSS test with immediate binding mode", func() {
	f := framework.NewDefaultFramework("fss-dynamic")

	Context("[cloudprovider][storage][csi][fss][mtexist]", func() {
		It("Basic Dynamic FSS test with immediate binding mode", func() {
			scParameters := map[string]string{"availabilityDomain": setupF.AdLabel, "mountTargetOcid": setupF.MntTargetOcid}
			pvcJig := framework.NewPVCTestJig(f.ClientSet, "csi-fss-dyn-e2e-test")
			scName := f.CreateStorageClassOrFail(framework.ClassFssDynamic, framework.FssProvisionerType, scParameters, pvcJig.Labels, "Immediate", false, "Delete", nil)
			f.StorageClasses = append(f.StorageClasses, scName)
			pvcObject := pvcJig.CreateAndAwaitPVCOrFailDynamicFSS(f.Namespace.Name, "50Gi", scName, v1.ClaimBound, nil)
			pvcJig.CheckSinglePodReadWrite(f.Namespace.Name, pvcObject.Name, false, []string{})
		})
	})
})
