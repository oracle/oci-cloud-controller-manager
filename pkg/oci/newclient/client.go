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
	"time"

	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/pkg/errors"

	oldclient "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
)

// Interface of consumed OCI API functionality.
type Interface interface {
	// GetInstance gets information about the specified instance.
	GetInstance(ctx context.Context, id string) (*core.Instance, error)
	// GetInstanceByDisplayName gets information about the named instance.
	GetInstanceByDisplayName(ctx context.Context, displayName string) (*core.Instance, error)
	// GetInstanceByNodeName gets the OCI instance corresponding to the given
	// Kubernetes node name.
	GetInstanceByNodeName(ctx context.Context, nodeName string) (*core.Instance, error)

	// GetVNIC gets information about the specified VNIC.
	GetVNIC(ctx context.Context, id string) (*core.Vnic, error)
	GetPrimaryVNICForInstance(ctx context.Context, instanceID string) (*core.Vnic, error)

	// GetSubnet gets information about the specified subnet.
	GetSubnet(ctx context.Context, id string) (*core.Subnet, error)
}

type client struct {
	compute *core.ComputeClient
	network *core.VirtualNetworkClient
	config  *oldclient.Config

	vcnID string

	subnetCache cache.Store
}

// New constructs an OCI API client.
func New(config *oldclient.Config) (Interface, error) {
	cp := common.NewRawConfigurationProvider(
		config.Auth.TenancyOCID,
		config.Auth.UserOCID,
		config.Auth.Region,
		config.Auth.Fingerprint,
		config.Auth.PrivateKey,
		&config.Auth.PrivateKeyPassphrase,
	)
	compute, err := core.NewComputeClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewComputeClientWithConfigurationProvider")
	}

	network, err := core.NewVirtualNetworkClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewVirtualNetworkClientWithConfigurationProvider")
	}

	c := &client{
		config:  config,
		compute: &compute,
		network: &network,

		subnetCache: cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
	}

	vcnID := config.VCNID
	if vcnID == "" {
		glog.Infof("No vcn provided in cloud provider config. Falling back to looking up VCN via LB subnet.")
		subnet, err := c.GetSubnet(context.Background(), config.LoadBalancer.Subnet1)
		if err != nil {
			return nil, errors.Wrap(err, "get subnet for loadBalancer.subnet1")
		}
		vcnID = *subnet.VcnId
	}
	c.vcnID = vcnID

	return c, nil
}

func (c *client) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	resp, err := c.compute.GetInstance(ctx, core.GetInstanceRequest{
		InstanceId: &id,
	})
	incRequestCounter(err, getVerb, instanceResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.Instance, nil
}

func (c *client) GetInstanceByDisplayName(ctx context.Context, displayName string) (*core.Instance, error) {
	var (
		page      *string
		instances []core.Instance
	)
	for {
		resp, err := c.compute.ListInstances(ctx, core.ListInstancesRequest{
			CompartmentId: &c.config.Auth.CompartmentOCID,
			DisplayName:   &displayName,
			Page:          page,
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
	resp, err := c.compute.ListVnicAttachments(ctx, req)
	incRequestCounter(err, listVerb, vnicAttachmentResource)

	if err != nil {
		return resp, errors.WithStack(err)
	}

	return resp, nil
}

func (c *client) GetPrimaryVNICForInstance(ctx context.Context, instanceID string) (*core.Vnic, error) {
	var page *string
	for {
		resp, err := c.listVNICAttachments(ctx, core.ListVnicAttachmentsRequest{
			InstanceId:    &instanceID,
			CompartmentId: &c.config.Auth.CompartmentOCID,
			Page:          page,
		})

		if err != nil {
			return nil, err
		}

		for _, attachment := range resp.Items {
			if attachment.LifecycleState != core.VnicAttachmentLifecycleStateAttached {
				glog.Infof("VNIC attachment %q for instance %q has a state of %q (not %q)", attachment.Id, instanceID, attachment.LifecycleState, core.VnicAttachmentLifecycleStateAttached)
				continue
			}

			// TODO(apryde): Cache map[instanceID]primaryVNICID.
			vnic, err := c.GetVNIC(ctx, *attachment.VnicId)
			if err != nil {
				return nil, err
			}
			if *vnic.IsPrimary {
				return vnic, nil
			}
		}

		if page = resp.OpcNextPage; resp.OpcNextPage == nil {
			break
		}
	}

	return nil, errors.Errorf("no primary VNIC associated with instance %q")
}

func (c *client) GetVNIC(ctx context.Context, id string) (*core.Vnic, error) {
	resp, err := c.network.GetVnic(ctx, core.GetVnicRequest{
		VnicId: &id,
	})
	incRequestCounter(err, getVerb, vnicResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.Vnic, nil
}

func (c *client) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	item, exists, err := c.subnetCache.GetByKey(id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if exists {
		return item.(*core.Subnet), nil
	}

	resp, err := c.network.GetSubnet(ctx, core.GetSubnetRequest{
		SubnetId: &id,
	})
	incRequestCounter(err, getVerb, subnetResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	subnet := &resp.Subnet
	_ = c.subnetCache.Add(subnet)
	return subnet, nil
}

func (c *client) GetInstanceByNodeName(ctx context.Context, nodeName string) (*core.Instance, error) {
	// First try lookup by display name.
	instance, err := c.GetInstanceByDisplayName(ctx, nodeName)
	if err == nil {
		return instance, nil
	}

	// Otherwise fall back to looking up via VNiC properties (hostname or public IP).
	var (
		page      *string
		instances []*core.Instance
	)
	for {
		resp, err := c.listVNICAttachments(ctx, core.ListVnicAttachmentsRequest{
			CompartmentId: &c.config.Auth.CompartmentOCID,
			Page:          page,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, attachment := range resp.Items {
			if attachment.LifecycleState != core.VnicAttachmentLifecycleStateAttached {
				glog.Infof("VNIC attachment %q for instance %q has a life cycle state of %q (not %q)",
					attachment.Id, nodeName, attachment.LifecycleState, core.VnicAttachmentLifecycleStateAttached)
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
			if *subnet.VcnId != c.vcnID {
				continue
			}

			if *vnic.PublicIp == nodeName || (*vnic.HostnameLabel != "" && strings.HasPrefix(nodeName, *vnic.HostnameLabel)) {
				instance, err := c.GetInstance(ctx, *attachment.InstanceId)
				if err != nil {
					return nil, err
				}

				if IsInstanceInTerminalState(instance) {
					glog.Warningf("Instance %q is in state %q which is a terminal state", instance.Id, instance.LifecycleState)
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
		instance.LifecycleState == core.InstanceLifecycleStateTerminating ||
		instance.LifecycleState == "UNKNOWN"
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

func subnetCacheKeyFn(obj interface{}) (string, error) {
	return *obj.(*core.Subnet).Id, nil
}
