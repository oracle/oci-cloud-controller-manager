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
	"net"

	"github.com/oracle/oci-go-sdk/core"
	"github.com/pkg/errors"
)

type NetworkingInterface interface {
	GetSubnet(ctx context.Context, id string) (*core.Subnet, error)
	GetSubnetFromCacheByIP(ip string) (*core.Subnet, error)
}

func (c *client) getVNIC(ctx context.Context, id string) (*core.Vnic, error) {
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

// GetSubnetFromCacheByIP checks to see if the given IP is contained by any subnet CIDR block in the subnet cache
// If no hits were found then no subnet and no error will be returned (nil, nil)
func (c *client) GetSubnetFromCacheByIP(ip string) (*core.Subnet, error) {
	ipAddr := net.ParseIP(ip)
	for _, subnetItem := range c.subnetCache.List() {
		subnet := subnetItem.(*core.Subnet)
		_, cidr, err := net.ParseCIDR(*subnet.CidrBlock)
		if err != nil {
			// This should never actually error but just in case
			return nil, fmt.Errorf("unable to parse CIDR block %q for subnet %q: %v", *subnet.CidrBlock, *subnet.Id, err)
		}

		if cidr.Contains(ipAddr) {
			return subnet, nil
		}
	}
	return nil, nil
}

func subnetCacheKeyFn(obj interface{}) (string, error) {
	return *obj.(*core.Subnet).Id, nil
}
