// Copyright 2018-2026 Oracle and/or its affiliates. All rights reserved.
// Licensed under the Apache License, Version 2.0

package client

import (
	"context"
	"fmt"
	"time"

	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	lustre "github.com/oracle/oci-go-sdk/v65/lustrefilestorage"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	lustreDefaultCreateInterval = 30 * time.Second
	lustreDefaultDeleteInterval = 45 * time.Second
)

// LustreInterface defines the interface to OCI File Storage with Lustre consumed by the CSI controller.
type LustreInterface interface {
	// CRUD
	CreateLustreFileSystem(ctx context.Context, details lustre.CreateLustreFileSystemDetails) (*lustre.LustreFileSystem, error)
	GetLustreFileSystem(ctx context.Context, id string) (*lustre.LustreFileSystem, error)
	ListLustreFileSystems(ctx context.Context, compartmentID, ad, displayName string) ([]lustre.LustreFileSystemSummary, error)
	DeleteLustreFileSystem(ctx context.Context, id string) error

	// Waiters
	AwaitLustreFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*lustre.LustreFileSystem, error)
	AwaitLustreFileSystemDeleted(ctx context.Context, logger *zap.SugaredLogger, id string) error

	// Work requests
	ListWorkRequests(ctx context.Context, compartmentID, resourceID string) ([]lustre.WorkRequestSummary, error)
	ListWorkRequestErrors(ctx context.Context, workRequestID string, volumeID string) ([]lustre.WorkRequestError, error)
}

// CreateLustreFileSystem creates a Lustre file system and returns the created object.
func (c *client) CreateLustreFileSystem(ctx context.Context, details lustre.CreateLustreFileSystemDetails) (*lustre.LustreFileSystem, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "CreateLustreFileSystem")
	}

	resp, err := c.lustre.CreateLustreFileSystem(ctx, lustre.CreateLustreFileSystemRequest{
		CreateLustreFileSystemDetails: details,
		OpcRetryToken:                 details.DisplayName, // idempotency token (csi-lustre-<uuid>)
		RequestMetadata:               c.requestMetadata,
	})
	incRequestCounter(err, createVerb, "lustreFileSystem")

	if resp.OpcRequestId != nil {
		c.logger.With("service", "lustre", "verb", createVerb, "resource", "lustreFileSystem").
			With("volumeName", details.DisplayName, "OpcRequestId", *(resp.OpcRequestId)).
			With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for CreateLustreFileSystem call.")
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	ls := resp.LustreFileSystem
	return &ls, nil
}

// GetLustreFileSystem gets a Lustre file system by OCID.
func (c *client) GetLustreFileSystem(ctx context.Context, id string) (*lustre.LustreFileSystem, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetLustreFileSystem")
	}

	resp, err := c.lustre.GetLustreFileSystem(ctx, lustre.GetLustreFileSystemRequest{
		LustreFileSystemId: &id,
		RequestMetadata:    c.requestMetadata,
	})
	incRequestCounter(err, getVerb, "lustreFileSystem")

	if resp.OpcRequestId != nil {
		c.logger.With("service", "lustre", "verb", getVerb, "resource", "lustreFileSystem").
			With("volumeID", id, "OpcRequestId", *(resp.OpcRequestId)).
			With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for GetLustreFileSystem call.")
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}
	ls := resp.LustreFileSystem
	return &ls, nil
}

// ListLustreFileSystems lists Lustre file systems filtered by compartment, AD, and displayName.
// Only ACTIVE and CREATING states are returned; conflicting states are returned with an error flag.
func (c *client) ListLustreFileSystems(ctx context.Context, compartmentID, ad, displayName string) ([]lustre.LustreFileSystemSummary, error) {
	var page *string
	items := make([]lustre.LustreFileSystemSummary, 0)

	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListLustreFileSystems")
		}

		req := lustre.ListLustreFileSystemsRequest{
			CompartmentId:   &compartmentID,
			RequestMetadata: c.requestMetadata,
			Page:            page,
		}
		if displayName != "" {
			req.DisplayName = &displayName
		}
		if ad != "" {
			req.AvailabilityDomain = &ad
		}

		resp, err := c.lustre.ListLustreFileSystems(ctx, req)
		incRequestCounter(err, listVerb, "lustreFileSystem")

		if resp.OpcRequestId != nil {
			c.logger.With("service", "lustre", "verb", listVerb, "resource", "lustreFileSystem").
				With("compartmentID", compartmentID, "availabilityDomain", ad, "volumeName", displayName, "OpcRequestId", *(resp.OpcRequestId)).
				With("statusCode", util.GetHttpStatusCode(err)).
				Info("OPC Request ID recorded for ListLustreFileSystems call.")
		}

		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, s := range resp.LustreFileSystemCollection.Items {
			items = append(items, s)
		}

		if page = resp.OpcNextPage; page == nil {
			break
		}
	}
	return items, nil
}

// DeleteLustreFileSystem deletes a Lustre file system by OCID.
func (c *client) DeleteLustreFileSystem(ctx context.Context, id string) error {
	if !c.rateLimiter.Writer.TryAccept() {
		return RateLimitError(true, "DeleteLustreFileSystem")
	}

	resp, err := c.lustre.DeleteLustreFileSystem(ctx, lustre.DeleteLustreFileSystemRequest{
		LustreFileSystemId: &id,
		RequestMetadata:    c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, "lustreFileSystem")

	if resp.OpcRequestId != nil {
		c.logger.With("service", "lustre", "verb", deleteVerb, "resource", "lustreFileSystem").
			With("volumeID", id, "OpcRequestId", *(resp.OpcRequestId)).
			With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for DeleteLustreFileSystem call.")
	}

	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// AwaitLustreFileSystemActive waits for the Lustre file system to become ACTIVE or returns error on terminal states.
func (c *client) AwaitLustreFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*lustre.LustreFileSystem, error) {
	logger.Info("Waiting for LustreFileSystem to go in lifecycle state ACTIVE")

	var fs *lustre.LustreFileSystem
	err := wait.PollUntilContextCancel(ctx, lustreDefaultCreateInterval, true, func(ctx context.Context) (bool, error) {
		logger.Debug("Polling LustreFileSystem lifecycle state")

		var err error
		fs, err = c.GetLustreFileSystem(ctx, id)
		if err != nil {
			return false, err
		}

		switch state := fs.LifecycleState; state {
		case lustre.LustreFileSystemLifecycleStateActive:
			logger.Infof("LustreFileSystem is in lifecycle state %q", state)
			return true, nil
		case lustre.LustreFileSystemLifecycleStateDeleting,
			lustre.LustreFileSystemLifecycleStateDeleted,
			lustre.LustreFileSystemLifecycleStateFailed:
			return false, fmt.Errorf("LustreFileSystem is in lifecycle state %q", state)
		default:
			logger.Infof("LustreFileSystem is in lifecycle state %q", state)
			return false, nil
		}
	})
	if err != nil {
		return nil, err
	}
	return fs, nil
}

// AwaitLustreFileSystemDeleted waits until the Lustre file system is deleted (404 or DELETED).
func (c *client) AwaitLustreFileSystemDeleted(ctx context.Context, logger *zap.SugaredLogger, id string) error {
	logger.Info("Waiting for LustreFileSystem  to be DELETED")

	return wait.PollUntilContextCancel(ctx, lustreDefaultDeleteInterval, true, func(ctx context.Context) (bool, error) {
		logger.Debug("Polling LustreFileSystem deletion state")

		fs, err := c.GetLustreFileSystem(ctx, id)
		if err != nil {
			if IsNotFound(err) {
				logger.Info("LustreFileSystem is deleted (not found).")
				return true, nil
			}
			return false, err
		}
		switch state := fs.LifecycleState; state {
		case lustre.LustreFileSystemLifecycleStateDeleted:
			logger.Info("LustreFileSystem is DELETED")
			return true, nil
		case lustre.LustreFileSystemLifecycleStateDeleting:
			logger.Debugf("LustreFileSystem is still DELETING")
			return false, nil
		default:
			logger.Infof("LustreFileSystem is in state %v", state)
			return false, nil
		}
	})
}

// ListWorkRequests lists Lustre work requests filtered by compartment and resourceId, sorted by timeAccepted desc.
func (c *client) ListWorkRequests(ctx context.Context, compartmentID, resourceID string) ([]lustre.WorkRequestSummary, error) {
	var page *string
	var items []lustre.WorkRequestSummary

	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListWorkRequests")
		}
		resp, err := c.lustre.ListWorkRequests(ctx, lustre.ListWorkRequestsRequest{
			CompartmentId:   &compartmentID,
			ResourceId:      &resourceID,
			SortBy:          lustre.ListWorkRequestsSortByTimeaccepted,
			SortOrder:       lustre.ListWorkRequestsSortOrderDesc,
			Page:            page,
			RequestMetadata: c.requestMetadata,
		})
		incRequestCounter(err, listVerb, "lustreWorkRequests")

		if resp.OpcRequestId != nil {
			c.logger.With("service", "lustre", "verb", listVerb, "resource", "lustreWorkRequests").
				With("compartmentID", compartmentID, "volumeID", resourceID, "OpcRequestId", *(resp.OpcRequestId)).
				With("statusCode", util.GetHttpStatusCode(err)).
				Info("OPC Request ID recorded for lustre ListWorkRequests call.")
		}

		if err != nil {
			return nil, errors.WithStack(err)
		}

		items = append(items, resp.WorkRequestSummaryCollection.Items...)
		if page = resp.OpcNextPage; page == nil {
			break
		}
	}
	return items, nil
}

// ListWorkRequestErrors returns work request errors for a given work request id.
func (c *client) ListWorkRequestErrors(ctx context.Context, workRequestID string, volumeID string) ([]lustre.WorkRequestError, error) {
	var page *string
	var items []lustre.WorkRequestError

	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListWorkRequestErrors")
		}
		resp, err := c.lustre.ListWorkRequestErrors(ctx, lustre.ListWorkRequestErrorsRequest{
			WorkRequestId:   &workRequestID,
			SortBy:          lustre.ListWorkRequestErrorsSortByTimestamp,
			SortOrder:       lustre.ListWorkRequestErrorsSortOrderDesc,
			Page:            page,
			RequestMetadata: c.requestMetadata,
		})

		if resp.OpcRequestId != nil {
			c.logger.With("service", "lustre", "verb", listVerb, "resource", "lustreWorkRequestErrors").
				With("workRequestID", workRequestID, "volumeID", volumeID, "OpcRequestId", *(resp.OpcRequestId)).
				With("statusCode", util.GetHttpStatusCode(err)).
				Info("OPC Request ID recorded for lustre ListWorkRequestErrors call.")
		}

		incRequestCounter(err, listVerb, "lustreWorkRequestErrors")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		items = append(items, resp.WorkRequestErrorCollection.Items...)
		if page = resp.OpcNextPage; page == nil {
			break
		}
	}
	return items, nil
}
