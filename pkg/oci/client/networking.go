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

	"github.com/oracle/oci-go-sdk/v50/core"
	"github.com/pkg/errors"
)

// NetworkingInterface defines the subset of the OCI compute API utilised by the CCM.
type NetworkingInterface interface {
	GetSubnet(ctx context.Context, id string) (*core.Subnet, error)
	GetSubnetFromCacheByIP(ip string) (*core.Subnet, error)
	IsRegionalSubnet(ctx context.Context, id string) (bool, error)

	GetVcn(ctx context.Context, id string) (*core.Vcn, error)

	GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error)
	UpdateSecurityList(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) (core.UpdateSecurityListResponse, error)

	GetPrivateIP(ctx context.Context, id string) (*core.PrivateIp, error)

	GetPublicIpByIpAddress(ctx context.Context, id string) (*core.PublicIp, error)
}

func (c *client) GetVNIC(ctx context.Context, id string) (*core.Vnic, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetVNIC")
	}

	resp, err := c.network.GetVnic(ctx, core.GetVnicRequest{
		VnicId:          &id,
		RequestMetadata: c.requestMetadata,
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

	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetSubnet")
	}

	resp, err := c.network.GetSubnet(ctx, core.GetSubnetRequest{
		SubnetId:        &id,
		RequestMetadata: c.requestMetadata,
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

func (c *client) IsRegionalSubnet(ctx context.Context, id string) (bool, error) {
	subnet, err := c.GetSubnet(ctx, id)
	if err != nil {
		return false, err
	}
	return subnet.AvailabilityDomain == nil, nil
}

func (c *client) GetVcn(ctx context.Context, id string) (*core.Vcn, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetVcn")
	}
	resp, err := c.network.GetVcn(ctx, core.GetVcnRequest{
		VcnId:           &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, vcnResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	vcn := &resp.Vcn
	return vcn, nil
}

func (c *client) GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return core.GetSecurityListResponse{}, RateLimitError(false, "GetSecurityList")
	}

	resp, err := c.network.GetSecurityList(ctx, core.GetSecurityListRequest{
		SecurityListId:  &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, securityListResource)

	return resp, errors.WithStack(err)
}

func (c *client) UpdateSecurityList(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) (core.UpdateSecurityListResponse, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return core.UpdateSecurityListResponse{}, RateLimitError(true, "UpdateSecurityList")
	}

	resp, err := c.network.UpdateSecurityList(ctx, core.UpdateSecurityListRequest{
		SecurityListId: &id,
		IfMatch:        &etag,
		UpdateSecurityListDetails: core.UpdateSecurityListDetails{
			IngressSecurityRules: ingressRules,
			EgressSecurityRules:  egressRules,
		},
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, updateVerb, securityListResource)
	return resp, errors.WithStack(err)
}

func subnetCacheKeyFn(obj interface{}) (string, error) {
	return *obj.(*core.Subnet).Id, nil
}

func (c *client) GetPrivateIP(ctx context.Context, id string) (*core.PrivateIp, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetPrivateIp")
	}

	resp, err := c.network.GetPrivateIp(ctx, core.GetPrivateIpRequest{
		PrivateIpId:     &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, privateIPResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.PrivateIp, nil
}

func (c *client) GetPublicIpByIpAddress(ctx context.Context, ip string) (*core.PublicIp, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetPublicIpByIpAddress")
	}
	resp, err := c.network.GetPublicIpByIpAddress(ctx, core.GetPublicIpByIpAddressRequest{
		GetPublicIpByIpAddressDetails: core.GetPublicIpByIpAddressDetails{
			IpAddress: &ip,
		},
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, publicReservedIPResource)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.PublicIp, nil
}
