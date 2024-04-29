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

	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/pkg/errors"
	"k8s.io/utils/pointer"
)

// NetworkingInterface defines the subset of the OCI compute API utilised by the CCM.
type NetworkingInterface interface {
	GetSubnet(ctx context.Context, id string) (*core.Subnet, error)
	GetSubnetFromCacheByIP(ip string) (*core.Subnet, error)
	IsRegionalSubnet(ctx context.Context, id string) (bool, error)

	GetVcn(ctx context.Context, id string) (*core.Vcn, error)
	GetVNIC(ctx context.Context, id string) (*core.Vnic, error)

	GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error)
	UpdateSecurityList(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) (core.UpdateSecurityListResponse, error)

	ListPrivateIps(ctx context.Context, vnicId string) ([]core.PrivateIp, error)
	GetPrivateIp(ctx context.Context, id string) (*core.PrivateIp, error)
	CreatePrivateIp(ctx context.Context, vnicID string) (*core.PrivateIp, error)
	GetIpv6(ctx context.Context, id string) (*core.Ipv6, error)

	GetPublicIpByIpAddress(ctx context.Context, id string) (*core.PublicIp, error)

	CreateNetworkSecurityGroup(ctx context.Context, compartmentId, vcnId, displayName, serviceUid string) (*core.NetworkSecurityGroup, error)
	GetNetworkSecurityGroup(ctx context.Context, id string) (*core.NetworkSecurityGroup, *string, error)
	ListNetworkSecurityGroups(ctx context.Context, displayName, compartmentId, vcnId string) ([]core.NetworkSecurityGroup, error)
	UpdateNetworkSecurityGroup(ctx context.Context, id, etag string, freeformTags map[string]string) (*core.NetworkSecurityGroup, error)
	DeleteNetworkSecurityGroup(ctx context.Context, id, etag string) (*string, error)

	AddNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.AddNetworkSecurityGroupSecurityRulesDetails) (*core.AddNetworkSecurityGroupSecurityRulesResponse, error)
	RemoveNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.RemoveNetworkSecurityGroupSecurityRulesDetails) (*core.RemoveNetworkSecurityGroupSecurityRulesResponse, error)
	ListNetworkSecurityGroupSecurityRules(ctx context.Context, id string, direction core.ListNetworkSecurityGroupSecurityRulesDirectionEnum) ([]core.SecurityRule, error)
	UpdateNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.UpdateNetworkSecurityGroupSecurityRulesDetails) (*core.UpdateNetworkSecurityGroupSecurityRulesResponse, error)
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
		c.logger.With(id).Infof("GetVNIC failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
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
		c.logger.With(id).Infof("GetSubnet failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
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
		c.logger.With(id).Infof("GetVcn failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
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

func (c *client) GetPrivateIp(ctx context.Context, id string) (*core.PrivateIp, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetPrivateIp")
	}

	resp, err := c.network.GetPrivateIp(ctx, core.GetPrivateIpRequest{
		PrivateIpId:     &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, privateIPResource)

	if err != nil {
		c.logger.With(id).Infof("GetPrivateIp failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
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
		c.logger.With(ip).Infof("GetPublicIpByIpAddress failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
		return nil, errors.WithStack(err)
	}

	return &resp.PublicIp, nil
}

func (c *client) ListPrivateIps(ctx context.Context, vnicId string) ([]core.PrivateIp, error) {
	privateIps := []core.PrivateIp{}
	// Walk through all pages to get all private IPs for VNIC
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListPrivateIp")
		}

		resp, err := c.network.ListPrivateIps(ctx, core.ListPrivateIpsRequest{
			VnicId:          &vnicId,
			RequestMetadata: c.requestMetadata,
		})
		incRequestCounter(err, listVerb, privateIPResource)
		if err != nil {
			c.logger.With(vnicId).Infof("ListPrivateIps failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
			return nil, errors.WithStack(err)
		}
		privateIps = append(privateIps, resp.Items...)
		if page := resp.OpcNextPage; page == nil {
			break
		}
	}

	return privateIps, nil
}

func (c *client) CreatePrivateIp(ctx context.Context, vnicId string) (*core.PrivateIp, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "CreatePrivateIp")
	}
	requestMetadata := getDefaultRequestMetadata(c.requestMetadata)
	resp, err := c.network.CreatePrivateIp(ctx, core.CreatePrivateIpRequest{
		CreatePrivateIpDetails: core.CreatePrivateIpDetails{
			VnicId: &vnicId,
		},
		RequestMetadata: requestMetadata,
	})
	incRequestCounter(err, createVerb, privateIPResource)
	if err != nil {
		c.logger.With(vnicId).Infof("CreatePrivateIp failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
		return nil, errors.WithStack(err)
	}

	return &resp.PrivateIp, nil
}

func (c *client) GetIpv6(ctx context.Context, id string) (*core.Ipv6, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetIpv6")
	}
	resp, err := c.network.GetIpv6(ctx, core.GetIpv6Request{
		Ipv6Id:          &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, ipv6IPResource)

	if err != nil {
		c.logger.With(id).Infof("GetIpv6 failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
		return nil, errors.WithStack(err)
	}

	return &resp.Ipv6, nil
}

func (c *client) CreateNetworkSecurityGroup(ctx context.Context, compartmentId, vcnId, displayName, serviceUid string) (*core.NetworkSecurityGroup, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "CreateNetworkSecurityGroup")
	}
	requestMetadata := getDefaultRequestMetadata(c.requestMetadata)

	resp, err := c.network.CreateNetworkSecurityGroup(ctx, core.CreateNetworkSecurityGroupRequest{
		CreateNetworkSecurityGroupDetails: core.CreateNetworkSecurityGroupDetails{
			CompartmentId: &compartmentId,
			VcnId:         &vcnId,
			DisplayName:   &displayName,
			FreeformTags:  map[string]string{"CreatedBy": "CCM", "ServiceUid": serviceUid},
		},
		OpcRetryToken:   &serviceUid,
		RequestMetadata: requestMetadata,
	})

	incRequestCounter(err, createVerb, nsgResource)
	if err != nil {
		c.logger.With(serviceUid).Infof("CreateNetworkSecurityGroup failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
		return nil, errors.WithStack(err)
	}

	return &resp.NetworkSecurityGroup, nil
}

func (c *client) GetNetworkSecurityGroup(ctx context.Context, id string) (*core.NetworkSecurityGroup, *string, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, nil, RateLimitError(false, "GetNSG")
	}

	resp, err := c.network.GetNetworkSecurityGroup(ctx, core.GetNetworkSecurityGroupRequest{
		NetworkSecurityGroupId: &id,
		RequestMetadata:        c.requestMetadata,
	})
	incRequestCounter(err, getVerb, nsgResource)

	if err != nil {
		c.logger.With(id).Infof("GetNetworkSecurityGroup failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
		return nil, nil, errors.WithStack(err)
	}

	return &resp.NetworkSecurityGroup, resp.Etag, nil
}

func (c *client) ListNetworkSecurityGroups(ctx context.Context, displayName, compartmentId, vcnId string) ([]core.NetworkSecurityGroup, error) {
	var page *string
	nsgList := make([]core.NetworkSecurityGroup, 0)
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListNSG")
		}

		resp, err := c.network.ListNetworkSecurityGroups(ctx, core.ListNetworkSecurityGroupsRequest{
			CompartmentId:   &compartmentId,
			VcnId:           &vcnId,
			Page:            page,
			DisplayName:     &displayName,
			SortBy:          core.ListNetworkSecurityGroupsSortByTimecreated,
			SortOrder:       core.ListNetworkSecurityGroupsSortOrderDesc,
			RequestMetadata: c.requestMetadata,
		})
		incRequestCounter(err, listVerb, nsgResource)

		if err != nil {
			c.logger.With(displayName).Infof("ListNetworkSecurityGroups failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
			return nil, errors.WithStack(err)
		}
		nsgList = append(nsgList, resp.Items...)
		if page = resp.OpcNextPage; resp.OpcNextPage == nil {
			break
		}
	}

	return nsgList, nil
}

func (c *client) UpdateNetworkSecurityGroup(ctx context.Context, id string, etag string, freeformTags map[string]string) (*core.NetworkSecurityGroup, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "UpdateNSG")
	}

	resp, err := c.network.UpdateNetworkSecurityGroup(ctx, core.UpdateNetworkSecurityGroupRequest{
		NetworkSecurityGroupId: &id,
		UpdateNetworkSecurityGroupDetails: core.UpdateNetworkSecurityGroupDetails{
			FreeformTags: freeformTags,
		},
		IfMatch:         &etag,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, updateVerb, nsgResource)

	if err != nil {
		c.logger.With(id).Infof("UpdateNetworkSecurityGroup failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
		return nil, errors.WithStack(err)
	}

	return &resp.NetworkSecurityGroup, nil
}

func (c *client) DeleteNetworkSecurityGroup(ctx context.Context, id, etag string) (*string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "DeleteNetworkSecurityGroup")
	}
	requestMetadata := getDefaultRequestMetadata(c.requestMetadata)

	resp, err := c.network.DeleteNetworkSecurityGroup(ctx, core.DeleteNetworkSecurityGroupRequest{
		NetworkSecurityGroupId: &id,
		IfMatch:                &etag,
		RequestMetadata:        requestMetadata,
	})

	incRequestCounter(err, deleteVerb, nsgResource)
	if err != nil {
		c.logger.With(id).Infof("DeleteNetworkSecurityGroup failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
		return nil, errors.WithStack(err)
	}

	return resp.OpcRequestId, nil
}

func (c *client) AddNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.AddNetworkSecurityGroupSecurityRulesDetails) (*core.AddNetworkSecurityGroupSecurityRulesResponse, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "AddNSGRules")
	}

	resp, err := c.network.AddNetworkSecurityGroupSecurityRules(ctx, core.AddNetworkSecurityGroupSecurityRulesRequest{
		NetworkSecurityGroupId:                      &id,
		AddNetworkSecurityGroupSecurityRulesDetails: details,
		RequestMetadata:                             c.requestMetadata,
	})
	incRequestCounter(err, createVerb, nsgRuleResource)

	if err != nil {
		c.logger.With(id).Infof("AddNetworkSecurityGroupSecurityRules failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
		return nil, errors.WithStack(err)
	}
	return &resp, nil
}

func (c *client) RemoveNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.RemoveNetworkSecurityGroupSecurityRulesDetails) (*core.RemoveNetworkSecurityGroupSecurityRulesResponse, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "RemoveNSGRules")
	}

	resp, err := c.network.RemoveNetworkSecurityGroupSecurityRules(ctx, core.RemoveNetworkSecurityGroupSecurityRulesRequest{
		NetworkSecurityGroupId:                         &id,
		RemoveNetworkSecurityGroupSecurityRulesDetails: details,
		RequestMetadata:                                c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, nsgRuleResource)

	if err != nil {
		c.logger.With(id).Infof("RemoveNetworkSecurityGroupSecurityRules failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
		return nil, errors.WithStack(err)
	}
	return &resp, nil
}

func (c *client) ListNetworkSecurityGroupSecurityRules(ctx context.Context, id string, direction core.ListNetworkSecurityGroupSecurityRulesDirectionEnum) ([]core.SecurityRule, error) {
	var page *string
	nsgRules := make([]core.SecurityRule, 0)
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListNetworkSecurityGroupSecurityRules")
		}
		resp, err := c.network.ListNetworkSecurityGroupSecurityRules(ctx, core.ListNetworkSecurityGroupSecurityRulesRequest{
			NetworkSecurityGroupId: &id,
			Direction:              direction,
			Page:                   page,
			RequestMetadata:        c.requestMetadata,
		})
		incRequestCounter(err, listVerb, nsgRuleResource)

		if err != nil {
			c.logger.With(id).Infof("ListNetworkSecurityGroupSecurityRules failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
			return nil, errors.WithStack(err)
		}
		for _, rule := range resp.Items {
			nsgRules = append(nsgRules, rule)
		}
		if page = resp.OpcNextPage; resp.OpcNextPage == nil {
			break
		}
	}
	return nsgRules, nil
}

func (c *client) UpdateNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.UpdateNetworkSecurityGroupSecurityRulesDetails) (*core.UpdateNetworkSecurityGroupSecurityRulesResponse, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return nil, RateLimitError(false, "UpdateNSGSecurityRules")
	}

	resp, err := c.network.UpdateNetworkSecurityGroupSecurityRules(ctx, core.UpdateNetworkSecurityGroupSecurityRulesRequest{
		NetworkSecurityGroupId:                         &id,
		UpdateNetworkSecurityGroupSecurityRulesDetails: details,
		RequestMetadata:                                c.requestMetadata,
	})
	incRequestCounter(err, updateVerb, nsgRuleResource)

	if err != nil {
		c.logger.With(id).Infof("UpdateNetworkSecurityGroupSecurityRules failed %s", pointer.StringDeref(resp.OpcRequestId, ""))
		return nil, errors.WithStack(err)
	}
	return &resp, nil
}
