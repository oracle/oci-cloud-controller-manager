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

	"github.com/oracle/oci-go-sdk/v50/core"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	_ = iota
	// KiB is 1024 bytes
	KiB = 1 << (10 * iota)
	// MiB is 1024KB
	MiB
	// GiB is 1024 MB
	GiB
	// TiB is 1024 GB
	TiB
)

const (
	volumePollInterval = 5 * time.Second
	// OCIVolumeID is the name of the oci volume id.
	OCIVolumeID = "ociVolumeID"
	// OCIVolumeBackupID is the name of the oci volume backup id annotation.
	OCIVolumeBackupID = "volume.beta.kubernetes.io/oci-volume-source"
	// FSType is the name of the file storage type parameter for storage classes.
	FSType         = "fsType"
	configFilePath = "/etc/oci/config.yaml"
)

// BlockStorageInterface defines the interface to OCI block storage utilised
// by the volume provisioner.
type BlockStorageInterface interface {
	AwaitVolumeAvailableORTimeout(ctx context.Context, id string) (*core.Volume, error)
	CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error)
	DeleteVolume(ctx context.Context, id string) error
	GetVolume(ctx context.Context, id string) (*core.Volume, error)
	GetVolumesByName(ctx context.Context, volumeName, compartmentID string) ([]core.Volume, error)
	UpdateVolume(ctx context.Context, volumeId string, details core.UpdateVolumeDetails) (*core.Volume, error)
}

func (c *client) GetVolume(ctx context.Context, id string) (*core.Volume, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetVolume")
	}

	resp, err := c.bs.GetVolume(ctx, core.GetVolumeRequest{
		VolumeId:        &id,
		RequestMetadata: c.requestMetadata})
	incRequestCounter(err, getVerb, volumeResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.Volume, nil

}

//AwaitVolumeAvailableORTimeout takes context as timeout
func (c *client) AwaitVolumeAvailableORTimeout(ctx context.Context, id string) (*core.Volume, error) {
	var vol *core.Volume
	if err := wait.PollImmediateUntil(volumePollInterval, func() (bool, error) {
		var err error
		vol, err = c.GetVolume(ctx, id)
		if err != nil {
			if !IsRetryable(err) {
				return false, err
			}
			return false, nil
		}

		switch state := vol.LifecycleState; state {
		case core.VolumeLifecycleStateAvailable:
			return true, nil
		case core.VolumeLifecycleStateFaulty,
			core.VolumeLifecycleStateTerminated,
			core.VolumeLifecycleStateTerminating:
			return false, errors.Errorf("volume did not become available (lifecycleState=%q)", state)
		}
		return false, nil
	}, ctx.Done()); err != nil {
		return nil, err
	}

	return vol, nil
}

func (c *client) CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(true, "CreateVolume")
	}

	resp, err := c.bs.CreateVolume(ctx, core.CreateVolumeRequest{CreateVolumeDetails: details,
		RequestMetadata: c.requestMetadata})
	incRequestCounter(err, createVerb, volumeResource)

	if err == nil {
		logger := c.logger.With("AvailabilityDomain", *(details.AvailabilityDomain),
			"CompartmentId", *(details.CompartmentId), "volumeName", *(details.DisplayName),
			"OpcRequestId", *(resp.OpcRequestId))
		logger.Info("OPC Request ID recorded while creating volume.")
	} else {
		return nil, errors.WithStack(err)
	}
	return &resp.Volume, nil
}

func (c *client) UpdateVolume(ctx context.Context, volumeId string, details core.UpdateVolumeDetails) (*core.Volume, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(true, "UpdateVolume")
	}

	resp, err := c.bs.UpdateVolume(ctx, core.UpdateVolumeRequest{
		VolumeId:            &volumeId,
		UpdateVolumeDetails: details,
		RequestMetadata:     c.requestMetadata,
	})
	incRequestCounter(err, updateVerb, volumeResource)

	if err == nil {
		logger := c.logger.With("volumeName", *(details.DisplayName),
			"OpcRequestId", *(resp.OpcRequestId))
		logger.Info("OPC Request ID recorded while updating volume.")
	} else {
		return nil, errors.WithStack(err)
	}
	return &resp.Volume, nil
}

func (c *client) DeleteVolume(ctx context.Context, id string) error {
	if !c.rateLimiter.Writer.TryAccept() {
		return RateLimitError(true, "DeleteVolume")
	}

	_, err := c.bs.DeleteVolume(ctx, core.DeleteVolumeRequest{
		VolumeId:        &id,
		RequestMetadata: c.requestMetadata})
	incRequestCounter(err, deleteVerb, volumeResource)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

/*
 * TODO: Expand the API to be generic 'GetVolumes' with the following features as necessary
 * 1. Option to sort by display name or creation timestamp.
 * 2. Option to filter by one or more lifecycle states.
 * 3. Option to filter by CompartmentID
 */
func (c *client) GetVolumesByName(ctx context.Context, volumeName, compartmentID string) ([]core.Volume, error) {
	var page *string
	volumeList := make([]core.Volume, 0)
	for {
		if !c.rateLimiter.Writer.TryAccept() {
			return nil, RateLimitError(true, "CreateVolume")
		}

		listVolumeResponse, err := c.bs.ListVolumes(ctx,
			core.ListVolumesRequest{
				CompartmentId:   &compartmentID,
				Page:            page,
				DisplayName:     &volumeName,
				RequestMetadata: c.requestMetadata,
			})

		if err != nil {
			return nil, errors.WithStack(err)
		}

		logger := c.logger.With("volumeName", volumeName, "CompartmentID", compartmentID,
			"OpcRequestId", *(listVolumeResponse.OpcRequestId))
		logger.Info("OPC Request ID recorded while fetching volumes by name.")

		for _, volume := range listVolumeResponse.Items {
			volumeState := volume.LifecycleState
			if volumeState == core.VolumeLifecycleStateAvailable ||
				volumeState == core.VolumeLifecycleStateProvisioning {
				volumeList = append(volumeList, volume)
			}
		}

		if page = listVolumeResponse.OpcNextPage; page == nil {
			break
		}
	}

	return volumeList, nil
}
