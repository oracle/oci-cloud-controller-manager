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
	volumePollInterval       = 5 * time.Second
	volumeBackupPollInterval = 5 * time.Second
	volumeClonePollInterval  = 10 * time.Second
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
	AwaitVolumeCloneAvailableOrTimeout(ctx context.Context, id string) (*core.Volume, error)
	CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error)
	DeleteVolume(ctx context.Context, id string) error
	GetVolume(ctx context.Context, id string) (*core.Volume, error)
	GetVolumesByName(ctx context.Context, volumeName, compartmentID string) ([]core.Volume, error)
	UpdateVolume(ctx context.Context, volumeId string, details core.UpdateVolumeDetails) (*core.Volume, error)

	AwaitVolumeBackupAvailableOrTimeout(ctx context.Context, id string) (*core.VolumeBackup, error)
	CreateVolumeBackup(ctx context.Context, details core.CreateVolumeBackupDetails) (*core.VolumeBackup, error)
	DeleteVolumeBackup(ctx context.Context, id string) error
	GetVolumeBackup(ctx context.Context, id string) (*core.VolumeBackup, error)
	GetVolumeBackupsByName(ctx context.Context, snapshotName, compartmentID string) ([]core.VolumeBackup, error)
}

func (c *client) GetVolume(ctx context.Context, id string) (*core.Volume, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetVolume")
	}

	resp, err := c.bs.GetVolume(ctx, core.GetVolumeRequest{
		VolumeId:        &id,
		RequestMetadata: c.requestMetadata})
	incRequestCounter(err, getVerb, volumeResource)

	if resp.OpcRequestId != nil {
		c.logger.With("service", "blockstorage", "verb", getVerb, "resource", volumeResource).
			With("volumeID", id, "OpcRequestId", *(resp.OpcRequestId)).
			With("statusCode", util.GetHttpStatusCode(err)).Info("OPC Request ID recorded for GetVolume call.")
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.Volume, nil

}

func (c *client) GetVolumeBackup(ctx context.Context, id string) (*core.VolumeBackup, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetVolumeBackup")
	}

	resp, err := c.bs.GetVolumeBackup(ctx, core.GetVolumeBackupRequest{
		VolumeBackupId:  &id,
		RequestMetadata: c.requestMetadata})
	incRequestCounter(err, getVerb, volumeBackupResource)

	if resp.OpcRequestId != nil {
		c.logger.With("service", "blockstorage", "verb", getVerb, "resource", volumeBackupResource).
			With("volumeBackupId", id, "OpcRequestId", *(resp.OpcRequestId)).With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for GetVolumeBackup call.")
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.VolumeBackup, nil
}

// AwaitVolumeAvailableORTimeout takes context as timeout
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

// AwaitVolumeBackupAvailableOrTimeout takes context as timeout
func (c *client) AwaitVolumeBackupAvailableOrTimeout(ctx context.Context, id string) (*core.VolumeBackup, error) {
	var volBackup *core.VolumeBackup
	if err := wait.PollImmediateUntil(volumeBackupPollInterval, func() (bool, error) {
		var err error
		volBackup, err = c.GetVolumeBackup(ctx, id)
		if err != nil {
			if !IsRetryable(err) {
				return false, err
			}
			return false, nil
		}

		switch state := volBackup.LifecycleState; state {
		case core.VolumeBackupLifecycleStateAvailable:
			return true, nil
		case core.VolumeBackupLifecycleStateFaulty,
			core.VolumeBackupLifecycleStateTerminated,
			core.VolumeBackupLifecycleStateTerminating:
			return false, errors.Errorf("snapshot did not become available (lifecycleState=%q)", state)
		}
		return false, nil
	}, ctx.Done()); err != nil {
		return nil, err
	}

	return volBackup, nil
}

func (c *client) AwaitVolumeCloneAvailableOrTimeout(ctx context.Context, id string) (*core.Volume, error) {
	var volClone *core.Volume
	if err := wait.PollImmediateUntil(volumeClonePollInterval, func() (bool, error) {
		var err error
		volClone, err = c.GetVolume(ctx, id)
		if err != nil {
			if !IsRetryable(err) {
				return false, err
			}
			return false, nil
		}

		switch state := volClone.LifecycleState; state {
		case core.VolumeLifecycleStateAvailable:
			if *volClone.IsHydrated == true {
				return true, nil
			}
			return false, nil
		case core.VolumeLifecycleStateFaulty,
			core.VolumeLifecycleStateTerminated,
			core.VolumeLifecycleStateTerminating:
			return false, errors.Errorf("Clone volume did not become available and hydrated (lifecycleState=%q) (hydrationStatus=%v)", state, *volClone.IsHydrated)
		}
		return false, nil
	}, ctx.Done()); err != nil {
		return nil, err
	}

	return volClone, nil
}

func (c *client) CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(true, "CreateVolume")
	}

	resp, err := c.bs.CreateVolume(ctx, core.CreateVolumeRequest{CreateVolumeDetails: details,
		RequestMetadata: c.requestMetadata})
	incRequestCounter(err, createVerb, volumeResource)

	if resp.OpcRequestId != nil {
		c.logger.With("service", "blockstorage", "verb", createVerb, "resource", volumeResource).
			With("volumeName", *(details.DisplayName), "OpcRequestId", *(resp.OpcRequestId)).With("statusCode", util.GetHttpStatusCode(err)).
			With("availabilityDomain", *(details.AvailabilityDomain), "CompartmentId", *(details.CompartmentId)).
			Info("OPC Request ID recorded while creating volume.")
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &resp.Volume, nil
}

func (c *client) CreateVolumeBackup(ctx context.Context, details core.CreateVolumeBackupDetails) (*core.VolumeBackup, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(true, "CreateSnapshot")
	}

	resp, err := c.bs.CreateVolumeBackup(ctx, core.CreateVolumeBackupRequest{CreateVolumeBackupDetails: details,
		RequestMetadata: c.requestMetadata})
	incRequestCounter(err, createVerb, volumeBackupResource)

	if resp.OpcRequestId != nil {
		c.logger.With("service", "blockstorage", "verb", createVerb, "resource", volumeBackupResource).
			With("volumeBackupName", *(details.DisplayName), "OpcRequestId", *(resp.OpcRequestId)).
			With("statusCode", util.GetHttpStatusCode(err)).Info("OPC Request ID recorded for CreateVolumeBackup call.")
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.VolumeBackup, nil
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

	if resp.OpcRequestId != nil {
		c.logger.With("service", "blockstorage", "verb", updateVerb, "resource", volumeResource).
			With("volumeName", *(details.DisplayName), "volumeID", volumeId, "OpcRequestId", *(resp.OpcRequestId)).
			With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded while updating volume.")
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &resp.Volume, nil
}

func (c *client) DeleteVolume(ctx context.Context, id string) error {
	if !c.rateLimiter.Writer.TryAccept() {
		return RateLimitError(true, "DeleteVolume")
	}

	resp, err := c.bs.DeleteVolume(ctx, core.DeleteVolumeRequest{
		VolumeId:        &id,
		RequestMetadata: c.requestMetadata})
	incRequestCounter(err, deleteVerb, volumeResource)

	if resp.OpcRequestId != nil {
		c.logger.With("service", "blockstorage", "verb", deleteVerb, "resource", volumeResource).
			With("volumeID", id, "OpcRequestId", *(resp.OpcRequestId)).
			With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for DeleteVolume call.")
	}

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *client) DeleteVolumeBackup(ctx context.Context, id string) error {
	if !c.rateLimiter.Writer.TryAccept() {
		return RateLimitError(true, "DeleteSnapshot")
	}

	resp, err := c.bs.DeleteVolumeBackup(ctx, core.DeleteVolumeBackupRequest{
		VolumeBackupId:  &id,
		RequestMetadata: c.requestMetadata})
	incRequestCounter(err, deleteVerb, volumeBackupResource)

	if resp.OpcRequestId != nil {
		c.logger.With("service", "blockstorage", "verb", deleteVerb, "resource", volumeBackupResource).
			With("volumeBackupId", id, "OpcRequestId", *(resp.OpcRequestId)).With("statusCode", util.GetHttpStatusCode(err)).
			Info("OPC Request ID recorded for DeleteVolumeBackup call.")
	}

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

		if listVolumeResponse.OpcRequestId != nil {
			c.logger.With("service", "blockstorage", "verb", listVerb, "resource", volumeResource).
				With("volumeName", volumeName, "CompartmentID", compartmentID, "OpcRequestId", *(listVolumeResponse.OpcRequestId)).
				With("statusCode", util.GetHttpStatusCode(err)).
				Info("OPC Request ID recorded while fetching volumes by name.")
		}

		if err != nil {
			return nil, errors.WithStack(err)
		}

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

func (c *client) GetVolumeBackupsByName(ctx context.Context, snapshotName, compartmentID string) ([]core.VolumeBackup, error) {
	var page *string
	volumeBackupList := make([]core.VolumeBackup, 0)

	for {

		if !c.rateLimiter.Writer.TryAccept() {
			return nil, RateLimitError(true, "CreateVolumeBackup")
		}

		listVolumeBackupsResponse, err := c.bs.ListVolumeBackups(ctx,
			core.ListVolumeBackupsRequest{
				CompartmentId:   &compartmentID,
				Page:            page,
				DisplayName:     &snapshotName,
				RequestMetadata: c.requestMetadata,
			})

		if listVolumeBackupsResponse.OpcRequestId != nil {
			c.logger.With("service", "blockstorage", "verb", listVerb, "resource", volumeBackupResource).
				With("snapshotName", snapshotName, "CompartmentID", compartmentID, "OpcRequestId", *(listVolumeBackupsResponse.OpcRequestId)).
				With("statusCode", util.GetHttpStatusCode(err)).
				Info("OPC Request ID recorded while fetching volume backups by name.")
		}

		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, volumeBackup := range listVolumeBackupsResponse.Items {
			volumeState := volumeBackup.LifecycleState
			if volumeState == core.VolumeBackupLifecycleStateAvailable ||
				volumeState == core.VolumeBackupLifecycleStateCreating {
				volumeBackupList = append(volumeBackupList, volumeBackup)
			}
		}

		if page = listVolumeBackupsResponse.OpcNextPage; page == nil {
			break
		}
	}

	return volumeBackupList, nil
}
