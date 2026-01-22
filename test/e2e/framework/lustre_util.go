// Copyright 2026 Oracle and/or its affiliates. All rights reserved.
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
	"strings"
	"time"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/lustrefilestorage"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (f *CloudProviderFramework) CheckLustreVolumeExist(ctx context.Context, fsId string) bool {
	fs, err := f.Client.Lustre().GetLustreFileSystem(ctx, fsId)
	if client.IsNotFound(err) {
		return false
	}
	if err != nil {
		return false
	}
	if fs.LifecycleState == lustrefilestorage.LustreFileSystemLifecycleStateDeleting || fs.LifecycleState == lustrefilestorage.LustreFileSystemLifecycleStateDeleted {
		return false
	}
	return true
}

func (f *CloudProviderFramework) WaitForLustreFSDeleted(ctx context.Context, compartmentId, adLocation, volumeHandle string, pollInterval, timeout time.Duration) bool {
	if volumeHandle == "" {
		return true
	}
	fsId := volumeHandle[:strings.Index(volumeHandle, ":")]
	Logf("Waiting for lustre filesystem %v to be deleted", fsId)
	var deleted bool
	err := wait.Poll(pollInterval, timeout, func() (done bool, err error) {
		exists := f.CheckLustreVolumeExist(ctx, fsId)
		if !exists {
			deleted = true
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		Logf("Error waiting for Lustre FS deletion: %v", err)
		return false
	}
	return deleted
}

// CleanupLustreFileSystems deletes any existing Lustre file systems in the given compartment.
// Intended for test hygiene before starting Lustre E2E tests.
func (f *CloudProviderFramework) CleanupLustreFileSystems(ctx context.Context, compartmentId string) {
	Logf("Scanning for pre-existing Lustre File Systems in compartment: %s", compartmentId)
	fsList, err := f.Client.Lustre().ListLustreFileSystems(ctx, compartmentId, "", "")
	if err != nil {
		Logf("Failed to list Lustre file systems for cleanup: %v", err)
		return
	}
	if len(fsList) == 0 {
		Logf("No pre-existing Lustre File Systems found to cleanup")
		return
	}
	for _, s := range fsList {
		// Skip already deleting/deleted
		if s.LifecycleState == lustrefilestorage.LustreFileSystemLifecycleStateDeleting ||
			s.LifecycleState == lustrefilestorage.LustreFileSystemLifecycleStateDeleted ||
			s.LifecycleState == lustrefilestorage.LustreFileSystemLifecycleStateCreating {
			continue
		}
		id := ""
		if s.Id != nil {
			id = *s.Id
		}
		name := ""
		if s.DisplayName != nil {
			name = *s.DisplayName
		}
		Logf("Deleting leftover Lustre FS: name=%s id=%s", name, id)
		if err := f.Client.Lustre().DeleteLustreFileSystem(ctx, id); err != nil {
			Logf("DeleteLustreFileSystem failed for %s: %v", id, err)
			continue
		}
	}
}
