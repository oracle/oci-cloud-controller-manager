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
	"time"

	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

const attachmentPollInterval = 10 * time.Second

// VolumeAttachmentInterface defines the interface to the OCI volume attachement
// API.
type VolumeAttachmentInterface interface {
	// FindVolumeAttachment searches for a volume attachment in either the state
	// ATTACHING or ATTACHED and returns the first volume attachment found.
	FindVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error)

	// AttachVolume attaches a block storage volume to the specified instance.
	// See https://docs.us-phoenix-1.oraclecloud.com/api/#/en/iaas/20160918/VolumeAttachment/AttachVolume
	AttachVolume(ctx context.Context, instanceID, volumeID string) (core.VolumeAttachment, error)

	AttachParavirtualizedVolume(ctx context.Context, instanceID, volumeID string, isPvEncryptionInTransitEnabled bool) (core.VolumeAttachment, error)

	// WaitForVolumeAttached polls waiting for a OCI block volume to be in the
	// ATTACHED state.
	WaitForVolumeAttached(ctx context.Context, attachmentID string) (core.VolumeAttachment, error)

	// DetachVolume detaches a storage volume from the specified instance.
	// See: https://docs.us-phoenix-1.oraclecloud.com/api/#/en/iaas/20160918/Volume/DetachVolume
	DetachVolume(ctx context.Context, id string) error

	// WaitForVolumeDetached polls waiting for a OCI block volume to be in the
	// DETACHED state.
	WaitForVolumeDetached(ctx context.Context, attachmentID string) error

	FindActiveVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error)
}

var _ VolumeAttachmentInterface = &client{}

func (c *client) FindVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	var page *string
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListVolumeAttachments")
		}

		resp, err := c.compute.ListVolumeAttachments(ctx, core.ListVolumeAttachmentsRequest{
			CompartmentId:   &compartmentID,
			VolumeId:        &volumeID,
			Page:            page,
			RequestMetadata: c.requestMetadata,
		})
		incRequestCounter(err, listVerb, volumeAttachmentResource)

		if resp.OpcRequestId != nil {
			c.logger.With("service", "compute", "verb", listVerb, "resource", volumeAttachmentResource).
				With("volumeID", volumeID, "OpcRequestId", *(resp.OpcRequestId)).With("statusCode", util.GetHttpStatusCode(err)).
				Info("OPC Request ID recorded for ListVolumeAttachments call.")
		}

		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, attachment := range resp.Items {
			state := attachment.GetLifecycleState()
			if state == core.VolumeAttachmentLifecycleStateAttaching ||
				state == core.VolumeAttachmentLifecycleStateAttached {
				return attachment, nil
			}
			if state == core.VolumeAttachmentLifecycleStateDetaching {
				return attachment, errors.WithStack(errNotFound)
			}
		}

		if page = resp.OpcNextPage; page == nil {
			break
		}
	}

	return nil, errors.WithStack(errNotFound)
}

func (c *client) GetVolumeAttachment(ctx context.Context, id string) (core.VolumeAttachment, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetVolumeAttachment")
	}

	resp, err := c.compute.GetVolumeAttachment(ctx, core.GetVolumeAttachmentRequest{
		VolumeAttachmentId: &id,
		RequestMetadata:    c.requestMetadata,
	})
	incRequestCounter(err, getVerb, volumeAttachmentResource)

	if resp.OpcRequestId != nil {
		c.logger.With("service", "compute", "verb", getVerb, "resource", volumeAttachmentResource).
			With("volumeAttachedId", id, "OpcRequestId", *(resp.OpcRequestId)).With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for GetVolumeAttachment call.")
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.VolumeAttachment, nil
}

func (c *client) AttachVolume(ctx context.Context, instanceID, volumeID string) (core.VolumeAttachment, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "")
	}

	device, err := c.getDevicePath(ctx, instanceID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp, err := c.compute.AttachVolume(ctx, core.AttachVolumeRequest{
		AttachVolumeDetails: core.AttachIScsiVolumeDetails{
			InstanceId: &instanceID,
			VolumeId:   &volumeID,
			Device:     device,
		},
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, createVerb, volumeAttachmentResource)

	if resp.OpcRequestId != nil {
		c.logger.With("service", "compute", "verb", createVerb, "resource", volumeAttachmentResource).
			With("volumeID", volumeID, "instanceID", instanceID, "OpcRequestId", *(resp.OpcRequestId)).With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for AttachVolume call.")
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.VolumeAttachment, nil
}

func (c *client) getDevicePath(ctx context.Context, instanceID string) (*string, error) {
	//https://docs.cloud.oracle.com/en-us/iaas/Content/Block/References/consistentdevicepaths.htm. here we
	//are getting first available consistent device using ListInstanceDevices using that device in time of attachment
	limit := 1
	isAvailable := true
	listInstanceDevicesResp, err := c.compute.ListInstanceDevices(ctx, core.ListInstanceDevicesRequest{
		InstanceId:  &instanceID,
		Limit:       &limit,
		IsAvailable: &isAvailable,
	})

	incRequestCounter(err, listVerb, instanceResource)

	if listInstanceDevicesResp.OpcRequestId != nil {
		c.logger.With("service", "compute", "verb", listVerb, "resource", instanceResource).
			With("instanceID", instanceID, "OpcRequestId", *(listInstanceDevicesResp.OpcRequestId)).
			With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for ListInstanceDevices call.")
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	device := listInstanceDevicesResp.Items[0].Name

	return device, nil
}

func (c *client) AttachParavirtualizedVolume(ctx context.Context, instanceID, volumeID string, isPvEncryptionInTransitEnabled bool) (core.VolumeAttachment, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "")
	}

	device, err := c.getDevicePath(ctx, instanceID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp, err := c.compute.AttachVolume(ctx, core.AttachVolumeRequest{
		AttachVolumeDetails: core.AttachParavirtualizedVolumeDetails{
			InstanceId:                     &instanceID,
			VolumeId:                       &volumeID,
			IsPvEncryptionInTransitEnabled: &isPvEncryptionInTransitEnabled,
			Device:                         device,
		},
		RequestMetadata: c.requestMetadata,
	})

	incRequestCounter(err, createVerb, volumeAttachmentResource)

	if resp.OpcRequestId != nil {
		c.logger.With("service", "compute", "verb", createVerb, "resource", instanceResource).
			With("volumeID", volumeID, "instanceID", instanceID, "OpcRequestId", *(resp.OpcRequestId)).
			With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for AttachVolume call.")
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp.VolumeAttachment, nil
}

func (c *client) WaitForVolumeAttached(ctx context.Context, id string) (core.VolumeAttachment, error) {
	var va core.VolumeAttachment
	if err := wait.PollImmediateUntil(attachmentPollInterval, func() (done bool, err error) {
		if va, err = c.GetVolumeAttachment(ctx, id); err != nil {
			if IsRetryable(err) {
				return false, nil
			}
			return true, errors.WithStack(err)
		}
		switch state := va.GetLifecycleState(); state {
		case core.VolumeAttachmentLifecycleStateAttached:
			return true, nil
		case core.VolumeAttachmentLifecycleStateDetaching, core.VolumeAttachmentLifecycleStateDetached:
			return false, errors.Errorf("attachment %q in lifecycle state %q", *(va.GetId()), state)
		}
		return false, nil
	}, ctx.Done()); err != nil {
		return nil, errors.WithStack(err)
	}

	return va, nil
}

func (c *client) DetachVolume(ctx context.Context, id string) error {
	if !c.rateLimiter.Writer.TryAccept() {
		return RateLimitError(false, "DetachVolume")
	}
	resp, err := c.compute.DetachVolume(ctx, core.DetachVolumeRequest{
		VolumeAttachmentId: &id,
		RequestMetadata:    c.requestMetadata,
	})

	if resp.OpcRequestId != nil {
		c.logger.With("service", "compute", "verb", deleteVerb, "resource", volumeAttachmentResource).
			With("volumeAttachedId", id, "OpcRequestId", *(resp.OpcRequestId)).With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for DetachVolume call.")
	}

	incRequestCounter(err, deleteVerb, volumeAttachmentResource)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *client) WaitForVolumeDetached(ctx context.Context, id string) error {
	if err := wait.PollImmediateUntil(attachmentPollInterval, func() (done bool, err error) {
		va, err := c.GetVolumeAttachment(ctx, id)
		if err != nil {
			if IsRetryable(err) {
				return false, nil
			}
			return true, errors.WithStack(err)
		}
		if va.GetLifecycleState() == core.VolumeAttachmentLifecycleStateDetached {
			return true, nil
		}
		return false, nil
	}, ctx.Done()); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *client) FindActiveVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	var page *string
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListVolumeAttachments")
		}

		resp, err := c.compute.ListVolumeAttachments(ctx, core.ListVolumeAttachmentsRequest{
			CompartmentId:   &compartmentID,
			VolumeId:        &volumeID,
			Page:            page,
			RequestMetadata: c.requestMetadata,
		})

		if resp.OpcRequestId != nil {
			c.logger.With("service", "compute", "verb", listVerb, "resource", volumeAttachmentResource).
				With("volumeID", volumeID, "OpcRequestId", *(resp.OpcRequestId)).With("statusCode", util.GetHttpStatusCode(err)).
				Info("OPC Request ID recorded for ListVolumeAttachments call.")
		}

		incRequestCounter(err, listVerb, volumeAttachmentResource)

		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, attachment := range resp.Items {
			state := attachment.GetLifecycleState()
			if state == core.VolumeAttachmentLifecycleStateAttaching ||
				state == core.VolumeAttachmentLifecycleStateAttached ||
				state == core.VolumeAttachmentLifecycleStateDetaching {
				return attachment, nil
			}
		}

		if page = resp.OpcNextPage; page == nil {
			break
		}
	}

	return nil, errors.WithStack(errNotFound)
}
