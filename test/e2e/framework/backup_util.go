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

package framework

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/block"
	ocicore "github.com/oracle/oci-go-sdk/v50/core"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// CreateBackupVolume creates a volume backup on OCI from an exsiting volume and returns the backup volume id
func (j *PVCTestJig) CreateBackupVolume(storageClient ocicore.BlockstorageClient, pvc *v1.PersistentVolumeClaim) (string, error) {
	By("Creating backup of the volume")
	pvc, err := j.KubeClient.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(context.Background(), pvc.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get persistent volume claim %q: %v", pvc.Name, err)
	}
	pv, err := j.KubeClient.CoreV1().PersistentVolumes().Get(context.Background(), pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get persistent volume created name by claim %q: %v", pvc.Spec.VolumeName, err)
	}
	volumeID := pv.ObjectMeta.Annotations[block.OCIVolumeID]
	// TO-DO make sure this ctx works okay - changed for PR
	ctx := context.TODO()
	backupVolume, err := storageClient.CreateVolumeBackup(ctx, ocicore.CreateVolumeBackupRequest{
		CreateVolumeBackupDetails: ocicore.CreateVolumeBackupDetails{
			VolumeId:    &volumeID,
			DisplayName: &j.Name,
			Type:        ocicore.CreateVolumeBackupDetailsTypeFull,
		},
	})
	if err != nil {
		return *backupVolume.Id, fmt.Errorf("failed to backup volume with ocid %q: %v", volumeID, err)
	}

	err = j.waitForVolumeAvailable(ctx, storageClient, *backupVolume.Id, DefaultTimeout)
	if err != nil {
		// Delete the volume if it failed to get in a good state for us
		if _, err := storageClient.DeleteVolumeBackup(ctx, ocicore.DeleteVolumeBackupRequest{
			VolumeBackupId: backupVolume.Id,
		}); err != nil {
			Logf("Backup volume failed to become available. Deleting backup volume %q was not possible: %v", *backupVolume.Id, err)
		}

		return *backupVolume.Id, err
	}
	return *backupVolume.Id, nil
}

func (j *PVCTestJig) waitForVolumeAvailable(ctx context.Context, storageClient ocicore.BlockstorageClient, volumeID string, timeout time.Duration) error {
	isVolumeReady := func() (bool, error) {
		ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
		defer cancel()

		resp, err := storageClient.GetVolumeBackup(ctx, ocicore.GetVolumeBackupRequest{
			VolumeBackupId: &volumeID,
		})
		if err != nil {
			return false, err
		}

		state := resp.LifecycleState
		Logf("State: %q", state)
		switch state {
		case ocicore.VolumeBackupLifecycleStateCreating:
			return false, nil
		case ocicore.VolumeBackupLifecycleStateAvailable:
			return true, nil
		case ocicore.VolumeBackupLifecycleStateFaulty,
			ocicore.VolumeBackupLifecycleStateTerminated,
			ocicore.VolumeBackupLifecycleStateTerminating:
			return false, fmt.Errorf("volume has lifecycle state %q", state)
		}
		return false, nil
	}

	return wait.PollImmediate(time.Second*5, timeout, func() (bool, error) {
		ready, err := isVolumeReady()
		if err != nil {
			return false, fmt.Errorf("failed to provision volume %q: %v", volumeID, err)
		}
		return ready, nil
	})
}
