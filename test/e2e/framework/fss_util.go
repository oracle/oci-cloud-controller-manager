package framework

import (
	"context"
	"github.com/oracle/oci-go-sdk/v65/filestorage"
	client "k8s.io/apimachinery/pkg/api/errors"
)

func (f *CloudProviderFramework) GetFSIdByDisplayName(ctx context.Context, compartmentId, adLocation, pvName string) (string, error) {
	_, fsVolumeSummaryList, err := f.Client.FSS().GetFileSystemSummaryByDisplayName(ctx, compartmentId, adLocation, pvName)
	if client.IsNotFound(err) {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return *fsVolumeSummaryList[0].Id, nil
}

func (f *CloudProviderFramework) GetExportsSetIdByMountTargetId(ctx context.Context, mountTargetId string) (string, error) {
	mountTarget, err := f.Client.FSS().GetMountTarget(ctx, mountTargetId)
	if client.IsNotFound(err) {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return *mountTarget.ExportSetId, nil
}

func (f *CloudProviderFramework) CheckFSVolumeExist(ctx context.Context, fsId string) bool {
	fs, err := f.Client.FSS().GetFileSystem(ctx, fsId)
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
	export, err := f.Client.FSS().FindExport(ctx, fsId, exportPath, exportSetId)
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
