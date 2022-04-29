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
	"strings"

	"github.com/oracle/oci-go-sdk/v50/core"
	"github.com/pkg/errors"
)

// ComputeInterface defines the subset of the OCI compute API utilised by the CCM.
type ComputeInterface interface {
	// GetInstance gets information about the specified instance.
	GetInstance(ctx context.Context, id string) (*core.Instance, error)

	// GetInstanceByNodeName gets the OCI instance corresponding to the given
	// Kubernetes node name.
	GetInstanceByNodeName(ctx context.Context, compartmentID, vcnID, nodeName string) (*core.Instance, error)

	GetPrimaryVNICForInstance(ctx context.Context, compartmentID, instanceID string) (*core.Vnic, error)

	VolumeAttachmentInterface
}

func (c *client) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetInstance")
	}

	resp, err := c.compute.GetInstance(ctx, core.GetInstanceRequest{
		InstanceId:      &id,
		RequestMetadata: c.requestMetadata})
	incRequestCounter(err, getVerb, instanceResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.Instance, nil
}

func (c *client) getInstanceByDisplayName(ctx context.Context, compartmentID, displayName string) (*core.Instance, error) {
	var (
		page      *string
		instances []core.Instance
	)
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListInstances")
		}
		resp, err := c.compute.ListInstances(ctx, core.ListInstancesRequest{
			CompartmentId:   &compartmentID,
			DisplayName:     &displayName,
			Page:            page,
			RequestMetadata: c.requestMetadata,
		})
		incRequestCounter(err, listVerb, instanceResource)

		if err != nil {
			return nil, errors.WithStack(err)
		}

		instances = append(instances, getNonTerminalInstances(resp.Items)...)
		if page = resp.OpcNextPage; resp.OpcNextPage == nil {
			break
		}

	}

	if len(instances) == 0 {
		return nil, errors.WithStack(errNotFound)
	}
	if len(instances) > 1 {
		return nil, errors.Errorf("too many instances returned for display name %q: %d", displayName, len(instances))
	}
	return &instances[0], nil
}

func (c *client) listVNICAttachments(ctx context.Context, req core.ListVnicAttachmentsRequest) (core.ListVnicAttachmentsResponse, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return core.ListVnicAttachmentsResponse{}, RateLimitError(false, "ListVnicAttachments")
	}

	resp, err := c.compute.ListVnicAttachments(ctx, req)
	incRequestCounter(err, listVerb, vnicAttachmentResource)

	if err != nil {
		return resp, errors.WithStack(err)
	}

	return resp, nil
}

func (c *client) GetPrimaryVNICForInstance(ctx context.Context, compartmentID, instanceID string) (*core.Vnic, error) {
	logger := c.logger.With("instanceID", instanceID, "compartmentID", compartmentID)

	var page *string
	for {
		resp, err := c.listVNICAttachments(ctx, core.ListVnicAttachmentsRequest{
			InstanceId:      &instanceID,
			CompartmentId:   &compartmentID,
			Page:            page,
			RequestMetadata: c.requestMetadata,
		})

		if err != nil {
			return nil, err
		}

		for _, attachment := range resp.Items {
			if attachment.LifecycleState != core.VnicAttachmentLifecycleStateAttached {
				logger.With("vnicAttachmentID", *attachment.Id).Info("VNIC attachment is not in attached state")
				continue
			}

			if attachment.VnicId == nil {
				// Should never happen but lets be extra cautious as field is non-mandatory in OCI API.
				logger.With("vnicAttachmentID", *attachment.Id).Error("VNIC attachment is attached but has no VNIC ID")
				continue
			}

			// TODO(apryde): Cache map[instanceID]primaryVNICID.
			vnic, err := c.GetVNIC(ctx, *attachment.VnicId)
			if err != nil {
				return nil, err
			}
			if vnic.IsPrimary != nil && *vnic.IsPrimary {
				return vnic, nil
			}
		}

		if page = resp.OpcNextPage; resp.OpcNextPage == nil {
			break
		}
	}

	return nil, errors.WithStack(errNotFound)
}

func (c *client) GetInstanceByNodeName(ctx context.Context, compartmentID, vcnID, nodeName string) (*core.Instance, error) {
	// First try lookup by display name.
	instance, err := c.getInstanceByDisplayName(ctx, compartmentID, nodeName)
	if err == nil {
		return instance, nil
	}

	logger := c.logger.With("nodeName", nodeName, "compartmentID", compartmentID)

	// Otherwise fall back to looking up via VNiC properties (hostname or public IP).
	var (
		page      *string
		instances []*core.Instance
	)
	for {
		resp, err := c.listVNICAttachments(ctx, core.ListVnicAttachmentsRequest{
			CompartmentId:   &compartmentID,
			Page:            page,
			RequestMetadata: c.requestMetadata,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, attachment := range resp.Items {
			if attachment.LifecycleState != core.VnicAttachmentLifecycleStateAttached {
				logger.With("vnicAttachmentID", *attachment.Id).Info("VNIC attachment is not in attached state")
				continue
			}

			if attachment.VnicId == nil {
				// Should never happen but lets be extra cautious as field is non-mandatory in OCI API.
				logger.With("vnicAttachmentID", *attachment.Id).Error("VNIC attachment is attached but has no VNIC ID")
				continue
			}

			vnic, err := c.GetVNIC(ctx, *attachment.VnicId)
			if err != nil {
				return nil, err
			}

			// Skip VNICs that aren't attached to the cluster's VCN.
			subnet, err := c.GetSubnet(ctx, *vnic.SubnetId)
			if err != nil {
				return nil, err
			}
			if *subnet.VcnId != vcnID {
				continue
			}

			if (vnic.PublicIp != nil && *vnic.PublicIp == nodeName) ||
				(vnic.PrivateIp != nil && *vnic.PrivateIp == nodeName) ||
				(vnic.HostnameLabel != nil && (*vnic.HostnameLabel != "" && strings.HasPrefix(nodeName, *vnic.HostnameLabel))) {
				instance, err := c.GetInstance(ctx, *attachment.InstanceId)
				if err != nil {
					return nil, err
				}

				if IsInstanceInTerminalState(instance) {
					logger.With("instanceID", *instance.Id,
						"lifecycleState", instance.LifecycleState).Warn("Instance in a terminal state")
					continue
				}

				instances = append(instances, instance)
			}
		}
		if page = resp.OpcNextPage; resp.OpcNextPage == nil {
			break
		}
	}

	if len(instances) == 0 {
		return nil, errors.WithStack(errNotFound)
	}
	if len(instances) > 1 {
		return nil, errors.Errorf("too many instances returned for node name %q: %d", nodeName, len(instances))
	}
	return instances[0], nil
}

// IsInstanceInTerminalState returns true if the instance is in a terminal state, false otherwise.
func IsInstanceInTerminalState(instance *core.Instance) bool {
	return instance.LifecycleState == core.InstanceLifecycleStateTerminated ||
		instance.LifecycleState == core.InstanceLifecycleStateTerminating
}

// IsInstanceInStoppedState returns true if the instance is in a stopped or stopping state, false otherwise.
func IsInstanceInStoppedState(instance *core.Instance) bool {
	return instance.LifecycleState == core.InstanceLifecycleStateStopped ||
		instance.LifecycleState == core.InstanceLifecycleStateStopping
}

func getNonTerminalInstances(instances []core.Instance) []core.Instance {
	var result []core.Instance
	for _, instance := range instances {
		if !IsInstanceInTerminalState(&instance) {
			result = append(result, instance)
		}
	}
	return result
}
