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

	"github.com/oracle/oci-go-sdk/core"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

const volumePollInterval = 5 * time.Second

// BlockStorageInterface defines the interface to OCI block storage utilisted
// by the volume provisioner.
type BlockStorageInterface interface {
	AwaitVolumeAvailable(ctx context.Context, id string) (*core.Volume, error)
	CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error)
	DeleteVolume(ctx context.Context, id string) error
}

func (c *client) GetVolume(ctx context.Context, id string) (*core.Volume, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetVolume")
	}

	resp, err := c.bs.GetVolume(ctx, core.GetVolumeRequest{VolumeId: &id})
	incRequestCounter(err, getVerb, volumeResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.Volume, nil

}

func (c *client) AwaitVolumeAvailable(ctx context.Context, id string) (*core.Volume, error) {
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

	resp, err := c.bs.CreateVolume(ctx, core.CreateVolumeRequest{CreateVolumeDetails: details})
	incRequestCounter(err, createVerb, volumeResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.Volume, nil
}

func (c *client) DeleteVolume(ctx context.Context, id string) error {
	if !c.rateLimiter.Writer.TryAccept() {
		return RateLimitError(true, "DeleteVolume")
	}

	_, err := c.bs.DeleteVolume(ctx, core.DeleteVolumeRequest{VolumeId: &id})
	incRequestCounter(err, deleteVerb, volumeResource)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
