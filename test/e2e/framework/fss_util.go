// Copyright 2022 Oracle and/or its affiliates. All rights reserved.
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
	"reflect"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/filestorage"
)

func (f *CloudProviderFramework) GetFSIdByDisplayName(ctx context.Context, compartmentId, adLocation, pvName string) (string, error) {
	Logf("GetFileSystemSummaryByDisplayName request params")
	Logf("compartmentId: %+v", compartmentId)
	Logf("adLocation: %+v", adLocation)
	Logf("pvName: %+v", pvName)
	_, fsVolumeSummaryList, err := f.Client.FSS(nil).GetFileSystemSummaryByDisplayName(ctx, compartmentId, adLocation, pvName)
	if client.IsNotFound(err) {
		return "", err
	}
	if err != nil {
		return "", err
	}
	if len(fsVolumeSummaryList) == 0 {
		Logf("fsVolumeSummaryList is empty or nil")
		return "", fmt.Errorf("no file system volume found")
	}

	Logf("fsVolumeSummaryList length: %d", len(fsVolumeSummaryList))
	Logf("First volume summary: %+v", fsVolumeSummaryList[0])

	return *fsVolumeSummaryList[0].Id, nil
}

func (f *CloudProviderFramework) GetExportsSetIdByMountTargetId(ctx context.Context, mountTargetId string) (string, error) {
	mountTarget, err := f.Client.FSS(nil).GetMountTarget(ctx, mountTargetId)
	if client.IsNotFound(err) {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return *mountTarget.ExportSetId, nil
}

func (f *CloudProviderFramework) GetMountTargetByVolumeName(ctx context.Context, compartmentId, adLocation, pvName string) (*filestorage.MountTarget, error) {
	fsID, err := f.GetFSIdByDisplayName(ctx, compartmentId, adLocation, pvName)
	if err != nil {
		return nil, err
	}

	return f.GetMountTargetByFileSystemID(ctx, fsID)
}

// GetMountTargetByFileSystemID returns the mount target recorded by the CSI
// provisioner on a filesystem. It deliberately avoids a ListFileSystems
// display-name lookup when the caller already has the filesystem OCID.
func (f *CloudProviderFramework) GetMountTargetByFileSystemID(ctx context.Context, fsID string) (*filestorage.MountTarget, error) {
	fs, err := f.Client.FSS(nil).GetFileSystem(ctx, fsID)
	if client.IsNotFound(err) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	if fs.FreeformTags == nil {
		return nil, fmt.Errorf("filesystem %s does not contain freeform tags", fsID)
	}

	mountTargetID := fs.FreeformTags["mountTargetOCID"]
	if mountTargetID == "" {
		return nil, fmt.Errorf("filesystem %s does not contain mountTargetOCID freeform tag", fsID)
	}

	mountTarget, err := f.Client.FSS(nil).GetMountTarget(ctx, mountTargetID)
	if client.IsNotFound(err) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return mountTarget, nil
}

func (f *CloudProviderFramework) CheckFSVolumeExist(ctx context.Context, fsId string) bool {
	fs, err := f.Client.FSS(nil).GetFileSystem(ctx, fsId)
	if client.IsNotFound(err) {
		return false
	}
	if err != nil {
		return false
	}
	if fs.LifecycleState == filestorage.FileSystemLifecycleStateDeleting || fs.LifecycleState == filestorage.FileSystemLifecycleStateDeleted {
		return false
	}
	return true
}

func (f *CloudProviderFramework) CheckExportExists(ctx context.Context, fsId, exportPath, exportSetId string) bool {
	export, err := f.Client.FSS(nil).FindExport(ctx, fsId, exportPath, exportSetId)
	if client.IsNotFound(err) {
		return false
	}
	if err != nil {
		return false
	}
	if export.LifecycleState == filestorage.ExportSummaryLifecycleStateDeleting || export.LifecycleState == filestorage.ExportSummaryLifecycleStateDeleted {
		return false
	}
	return true
}

func ValidateMountTargetSecurityAttributes(mountTarget *filestorage.MountTarget, saNs, sa string, val interface{}) bool {
	if mountTarget.SecurityAttributes == nil {
		Logf("Mount target security attributes are nil")
		return saNs == ""
	}

	if saNs == "" && len(mountTarget.SecurityAttributes) == 0 {
		return true
	}

	ns, ok := mountTarget.SecurityAttributes[saNs]
	if !ok {
		Logf("Security attribute namespace %s not present on mount target", saNs)
		return false
	}

	actual, ok := ns[sa]
	if !ok {
		Logf("Security attribute %s not present in security attribute namespace %s on mount target", sa, saNs)
		return false
	}

	Logf("Mount target security attributes: current: %v, waiting for %v", actual, val)

	return reflect.DeepEqual(actual, val)
}

func (f *CloudProviderFramework) CheckMountTargetSecurityAttributesByVolumeName(ctx context.Context, volumeName string, compartment string, adlocation string, saNs string, sa string, val interface{}) (bool, error) {
	mountTarget, err := f.GetMountTargetByVolumeName(ctx, compartment, adlocation, volumeName)
	if err != nil {
		return false, err
	}

	return ValidateMountTargetSecurityAttributes(mountTarget, saNs, sa, val), nil
}

func (f *CloudProviderFramework) CheckMountTargetSecurityAttributesByFileSystemID(ctx context.Context, filesystemID string, saNs string, sa string, val interface{}) (bool, error) {
	mountTarget, err := f.GetMountTargetByFileSystemID(ctx, filesystemID)
	if err != nil {
		return false, err
	}

	return ValidateMountTargetSecurityAttributes(mountTarget, saNs, sa, val), nil
}
