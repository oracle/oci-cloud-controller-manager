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

package client

import (
	"context"
	"fmt"
	"time"

	fss "github.com/oracle/oci-go-sdk/v50/filestorage"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	defaultTimeout  = 5 * time.Minute
	defaultInterval = 5 * time.Second
)

// FileStorageInterface defines the interface to OCI File Storage Service
// consumed by the volume provisioner and fss-controller.
type FileStorageInterface interface {
	AwaitMountTargetActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*fss.MountTarget, error)

	GetFileSystem(ctx context.Context, id string) (*fss.FileSystem, error)
	GetFileSystemSummaryByDisplayName(ctx context.Context, compartmentID, ad, displayName string) (bool, []fss.FileSystemSummary, error)
	AwaitFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*fss.FileSystem, error)
	CreateFileSystem(ctx context.Context, details fss.CreateFileSystemDetails) (*fss.FileSystem, error)
	DeleteFileSystem(ctx context.Context, id string) error

	CreateExport(ctx context.Context, details fss.CreateExportDetails) (*fss.Export, error)
	FindExport(ctx context.Context, fsID, path, exportSetID string) (*fss.ExportSummary, error)
	AwaitExportActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*fss.Export, error)
	DeleteExport(ctx context.Context, id string) error

	GetMountTarget(ctx context.Context, id string) (*fss.MountTarget, error)
	CreateMountTarget(ctx context.Context, details fss.CreateMountTargetDetails) (*fss.MountTarget, error)
	DeleteMountTarget(ctx context.Context, id string) error
	GetMountTargetSummaryByDisplayName(ctx context.Context, compartmentID, ad, mountTargetName string) (bool, []fss.MountTargetSummary, error)
}

func (c *client) CreateFileSystem(ctx context.Context, details fss.CreateFileSystemDetails) (*fss.FileSystem, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "CreateFileSystem")
	}

	resp, err := c.filestorage.CreateFileSystem(ctx, fss.CreateFileSystemRequest{
		CreateFileSystemDetails: details,
		RequestMetadata:         c.requestMetadata,
	})

	incRequestCounter(err, createVerb, fileSystemResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.FileSystem, nil
}

func (c *client) GetFileSystem(ctx context.Context, id string) (*fss.FileSystem, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetFileSystem")
	}

	resp, err := c.filestorage.GetFileSystem(ctx, fss.GetFileSystemRequest{
		FileSystemId:    &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, fileSystemResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.FileSystem, nil
}

func (c *client) AwaitFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*fss.FileSystem, error) {
	logger.Infof("Waiting for FileSystem to be in lifecycle state %q", fss.FileSystemLifecycleStateActive)

	var fs *fss.FileSystem
	err := wait.PollImmediate(defaultInterval, defaultTimeout, func() (bool, error) {
		logger.Debug("Polling FileSystem lifecycle state")

		var err error
		fs, err = c.GetFileSystem(ctx, id)
		if err != nil {
			return false, err
		}

		switch state := fs.LifecycleState; state {
		case fss.FileSystemLifecycleStateActive:
			logger.Infof("FileSystem is in lifecycle state %q", state)
			return true, nil
		case fss.FileSystemLifecycleStateDeleting, fss.FileSystemLifecycleStateDeleted, fss.FileSystemLifecycleStateFailed:
			return false, errors.Errorf("file system %q is in lifecycle state %q", *fs.Id, state)
		default:
			logger.Debugf("FileSystem is in lifecycle state %q", state)
			return false, nil
		}
	})
	if err != nil {
		return nil, err
	}

	return fs, nil
}

func (c *client) GetFileSystemSummaryByDisplayName(ctx context.Context, compartmentID, ad, displayName string) (bool, []fss.FileSystemSummary, error) {

	var page *string
	fileSystemSummaries := make([]fss.FileSystemSummary, 0)
	conflictingFileSystemSummaries := make([]fss.FileSystemSummary, 0)
	foundConflicting := false
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return foundConflicting, nil, RateLimitError(false, "ListFileSystems")
		}

		resp, err := c.filestorage.ListFileSystems(ctx, fss.ListFileSystemsRequest{
			CompartmentId:      &compartmentID,
			AvailabilityDomain: &ad,
			DisplayName:        &displayName,
			RequestMetadata:    c.requestMetadata,
		})
		incRequestCounter(err, listVerb, fileSystemResource)
		if err != nil {
			return foundConflicting, nil, errors.WithStack(err)
		}

		logger := c.logger.With("volumeName", displayName, "compartmentID", compartmentID, "availabilityDomain", ad,
			"OpcRequestId", *(resp.OpcRequestId))
		logger.Info("OPC Request ID recorded while fetching file systems by name.")

		for _, fileSystemSummary := range resp.Items {
			lifecycleState := fileSystemSummary.LifecycleState
			if lifecycleState == fss.FileSystemSummaryLifecycleStateActive ||
				lifecycleState == fss.FileSystemSummaryLifecycleStateCreating {
				fileSystemSummaries = append(fileSystemSummaries, fileSystemSummary)
			} else {
				conflictingFileSystemSummaries = append(fileSystemSummaries, fileSystemSummary)
				foundConflicting = true
			}
		}

		if page = resp.OpcNextPage; page == nil {
			break
		}
	}

	if foundConflicting {
		return foundConflicting, conflictingFileSystemSummaries, errors.Errorf("Found file system summary neither active nor creating state")
	}
	return foundConflicting, fileSystemSummaries, nil
}

func (c *client) DeleteFileSystem(ctx context.Context, id string) error {
	if !c.rateLimiter.Writer.TryAccept() {
		return RateLimitError(true, "DeleteFileSystem")
	}

	_, err := c.filestorage.DeleteFileSystem(ctx, fss.DeleteFileSystemRequest{
		FileSystemId:    &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, fileSystemResource)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *client) GetMountTarget(ctx context.Context, id string) (*fss.MountTarget, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetMountTarget")
	}

	resp, err := c.filestorage.GetMountTarget(ctx, fss.GetMountTargetRequest{
		MountTargetId:   &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, mountTargetResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.MountTarget, nil
}

func (c *client) AwaitMountTargetActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*fss.MountTarget, error) {
	logger.Infof("Waiting for MountTarget to be in lifecycle state %q", fss.MountTargetLifecycleStateActive)

	var mt *fss.MountTarget
	if err := wait.PollImmediate(defaultInterval, defaultTimeout, func() (bool, error) {
		logger.Debug("Polling MountTarget lifecycle state")

		var err error
		mt, err = c.GetMountTarget(ctx, id)
		if err != nil {
			return false, err
		}

		switch state := mt.LifecycleState; state {
		case fss.MountTargetLifecycleStateActive:
			logger.Infof("Mount target is in lifecycle state %q", state)
			return true, nil
		case fss.MountTargetLifecycleStateFailed,
			fss.MountTargetLifecycleStateDeleting,
			fss.MountTargetLifecycleStateDeleted:
			logger.With("lifecycleState", state).Error("MountTarget will not become ACTIVE")
			return false, fmt.Errorf("mount target %q is in lifecycle state %q and will not become ACTIVE", *mt.Id, state)
		default:
			logger.Debugf("Mount target is in lifecycle state %q", state)
			return false, nil
		}
	}); err != nil {
		return nil, err
	}
	return mt, nil
}

func (c *client) CreateExport(ctx context.Context, details fss.CreateExportDetails) (*fss.Export, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "CreateExport")
	}

	resp, err := c.filestorage.CreateExport(ctx, fss.CreateExportRequest{
		CreateExportDetails: details,
		RequestMetadata:     c.requestMetadata,
	})
	incRequestCounter(err, createVerb, exportResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.Export, nil

}

func (c *client) GetExport(ctx context.Context, id string) (*fss.Export, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetExport")
	}

	resp, err := c.filestorage.GetExport(ctx, fss.GetExportRequest{
		ExportId:        &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, exportResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.Export, nil
}

// FindExport looks for an existing export with the same filesystem ID, export set ID, and path.
func (c *client) FindExport(ctx context.Context, fsID, path, exportSetID string) (*fss.ExportSummary, error) {
	var page *string
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListExports")
		}
		resp, err := c.filestorage.ListExports(ctx, fss.ListExportsRequest{
			FileSystemId:    &fsID,
			ExportSetId:     &exportSetID,
			Page:            page,
			RequestMetadata: c.requestMetadata,
		})
		incRequestCounter(err, listVerb, exportResource)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		for _, export := range resp.Items {
			if *export.Path == path {
				if export.LifecycleState == fss.ExportSummaryLifecycleStateCreating ||
					export.LifecycleState == fss.ExportSummaryLifecycleStateActive {
					return &export, nil
				}
				return &export, errors.Errorf("Found export in conflicting state %s: %s", *export.Id, export.LifecycleState)
			}
		}
		if page = resp.OpcNextPage; resp.OpcNextPage == nil {
			break
		}
	}

	return nil, errors.WithStack(errNotFound)
}

func (c *client) AwaitExportActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*fss.Export, error) {
	logger.Info("Waiting for Export to be in lifecycle state ACTIVE")

	var export *fss.Export
	if err := wait.PollImmediate(defaultInterval, defaultTimeout, func() (bool, error) {
		logger.Debug("Polling export lifecycle state")

		var err error
		export, err = c.GetExport(ctx, id)
		if err != nil {
			return false, err
		}

		switch state := export.LifecycleState; state {
		case fss.ExportLifecycleStateActive:
			logger.Infof("Export is in lifecycle state %q", state)
			return true, nil
		case fss.ExportLifecycleStateDeleting, fss.ExportLifecycleStateDeleted:
			logger.Errorf("Export is in lifecycle state %q", state)
			return false, fmt.Errorf("export %q is in lifecycle state %q", *export.Id, state)
		default:
			logger.Debugf("Export is in lifecycle state %q", state)
			return false, nil
		}
	}); err != nil {
		return nil, err
	}
	return export, nil
}

func (c *client) DeleteExport(ctx context.Context, id string) error {
	if !c.rateLimiter.Writer.TryAccept() {
		return RateLimitError(true, "DeleteExport")
	}

	_, err := c.filestorage.DeleteExport(ctx, fss.DeleteExportRequest{
		ExportId:        &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, exportResource)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *client) CreateMountTarget(ctx context.Context, details fss.CreateMountTargetDetails) (*fss.MountTarget, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "CreateMountTarget")
	}

	resp, err := c.filestorage.CreateMountTarget(ctx, fss.CreateMountTargetRequest{
		CreateMountTargetDetails: details,
		RequestMetadata:          c.requestMetadata,
	})
	incRequestCounter(err, createVerb, mountTargetResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.MountTarget, nil
}

func (c *client) DeleteMountTarget(ctx context.Context, id string) error {
	if !c.rateLimiter.Writer.TryAccept() {
		return RateLimitError(true, "DeleteMountTarget")
	}

	_, err := c.filestorage.DeleteMountTarget(ctx, fss.DeleteMountTargetRequest{
		MountTargetId:   &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, mountTargetResource)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *client) GetMountTargetSummaryByDisplayName(ctx context.Context, compartmentID, ad, mountTargetName string) (bool, []fss.MountTargetSummary, error) {

	var page *string
	mountTargetSummaries := make([]fss.MountTargetSummary, 0)
	conflictingMountTargetSummaries := make([]fss.MountTargetSummary, 0)
	foundConflicting := false
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return foundConflicting, nil, RateLimitError(false, "ListFileSystems")
		}

		resp, err := c.filestorage.ListMountTargets(ctx, fss.ListMountTargetsRequest{
			CompartmentId:      &compartmentID,
			AvailabilityDomain: &ad,
			DisplayName:        &mountTargetName,
			RequestMetadata:    c.requestMetadata,
		})
		incRequestCounter(err, listVerb, mountTargetResource)
		if err != nil {
			return foundConflicting, nil, errors.WithStack(err)
		}

		logger := c.logger.With("volumeName", mountTargetName, "compartmentID", compartmentID, "availabilityDomain", ad,
			"OpcRequestId", *(resp.OpcRequestId))
		logger.Info("OPC Request ID recorded while fetching mount targets by name.")

		for _, mountTargetSummary := range resp.Items {
			lifecycleState := mountTargetSummary.LifecycleState
			if lifecycleState == fss.MountTargetSummaryLifecycleStateActive ||
				lifecycleState == fss.MountTargetSummaryLifecycleStateCreating {
				mountTargetSummaries = append(mountTargetSummaries, mountTargetSummary)
			} else {
				conflictingMountTargetSummaries = append(conflictingMountTargetSummaries, mountTargetSummary)
				foundConflicting = true
			}
		}

		if page = resp.OpcNextPage; page == nil {
			break
		}
	}

	if foundConflicting {
		return foundConflicting, conflictingMountTargetSummaries, errors.Errorf("Found mount target summary neither active nor creating state")
	}

	return foundConflicting, mountTargetSummaries, nil
}
