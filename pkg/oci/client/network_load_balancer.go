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

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/networkloadbalancer"
	"github.com/pkg/errors"
)

type networkLoadbalancer struct {
	networkloadbalancer networkLoadBalancerClient
	requestMetadata     common.RequestMetadata
	rateLimiter         RateLimiter
}

const (
	NetworkLoadBalancerEntityType = "NetworkLoadBalancer"
)

func (c *networkLoadbalancer) GetLoadBalancer(ctx context.Context, id string) (*GenericLoadBalancer, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetLoadBalancer")
	}

	resp, err := c.networkloadbalancer.GetNetworkLoadBalancer(ctx, networkloadbalancer.GetNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: &id,
		RequestMetadata:       c.requestMetadata,
	})
	incRequestCounter(err, getVerb, networkLoadBalancerResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return c.networkLoadbalancerToGenericLoadbalancer(&resp.NetworkLoadBalancer), nil
}

func (c *networkLoadbalancer) GetLoadBalancerByName(ctx context.Context, compartmentID string, name string) (*GenericLoadBalancer, error) {
	var page *string
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListLoadBalancers")
		}
		resp, err := c.networkloadbalancer.ListNetworkLoadBalancers(ctx, networkloadbalancer.ListNetworkLoadBalancersRequest{
			CompartmentId:   &compartmentID,
			DisplayName:     &name,
			Page:            page,
			RequestMetadata: c.requestMetadata,
		})
		incRequestCounter(err, listVerb, networkLoadBalancerResource)

		if err != nil {
			return nil, errors.WithStack(err)
		}
		for _, lb := range resp.Items {
			if *lb.DisplayName == name {
				return c.networkLoadbalancerSummaryToGenericLoadbalancer(&lb), nil
			}
		}
		if page = resp.OpcNextPage; page == nil {
			break
		}
	}

	return nil, errors.WithStack(errNotFound)
}

func (c *networkLoadbalancer) CreateLoadBalancer(ctx context.Context, details *GenericCreateLoadBalancerDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateLoadBalancer")
	}

	resp, err := c.networkloadbalancer.CreateNetworkLoadBalancer(ctx, networkloadbalancer.CreateNetworkLoadBalancerRequest{
		CreateNetworkLoadBalancerDetails: networkloadbalancer.CreateNetworkLoadBalancerDetails{
			CompartmentId:               details.CompartmentId,
			DisplayName:                 details.DisplayName,
			SubnetId:                    &details.SubnetIds[0],
			IsPreserveSourceDestination: details.IsPreserveSourceDestination,
			ReservedIps:                 c.genericReservedIpToReservedIps(details.ReservedIps),
			IsPrivate:                   details.IsPrivate,
			NetworkSecurityGroupIds:     details.NetworkSecurityGroupIds,
			Listeners:                   c.genericListenerDetailsToListenerDetails(details.Listeners),
			BackendSets:                 c.genericBackendSetDetailsToBackendSets(details.BackendSets),
			FreeformTags:                details.FreeformTags,
			DefinedTags:                 details.DefinedTags,
		},
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, createVerb, networkLoadBalancerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *networkLoadbalancer) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "DeleteLoadBalancer")
	}

	resp, err := c.networkloadbalancer.DeleteNetworkLoadBalancer(ctx, networkloadbalancer.DeleteNetworkLoadBalancerRequest{
		NetworkLoadBalancerId: &id,
		RequestMetadata:       c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, networkLoadBalancerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *networkLoadbalancer) CreateCertificate(ctx context.Context, lbID string, cert *GenericCertificate) (string, error) {
	return "", nil
}

func (c *networkLoadbalancer) GetWorkRequest(ctx context.Context, id string) (*networkloadbalancer.WorkRequest, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetWorkRequest")
	}

	resp, err := c.networkloadbalancer.GetWorkRequest(ctx, networkloadbalancer.GetWorkRequestRequest{
		WorkRequestId:   &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, workRequestResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.WorkRequest, nil
}

func (c *networkLoadbalancer) CreateBackendSet(ctx context.Context, lbID string, name string, details *GenericBackendSetDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateBackendSet")
	}

	resp, err := c.networkloadbalancer.CreateBackendSet(ctx, networkloadbalancer.CreateBackendSetRequest{
		NetworkLoadBalancerId: &lbID,
		CreateBackendSetDetails: networkloadbalancer.CreateBackendSetDetails{
			Name:             &name,
			Backends:         backendsToBackendDetails(details.Backends),
			IsPreserveSource: details.IsPreserveSource,
			HealthChecker:    healthCheckerToHealthCheckerDetails(details.HealthChecker),
			Policy:           networkloadbalancer.NetworkLoadBalancingPolicyEnum(*details.Policy),
		},
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, createVerb, backendSetResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *networkLoadbalancer) GetCertificateByName(ctx context.Context, lbID string, name string) (*GenericCertificate, error) {
	return nil, nil
}

func (c *networkLoadbalancer) UpdateBackendSet(ctx context.Context, lbID string, name string, details *GenericBackendSetDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "UpdateBackendSet")
	}

	stringPolicy := details.Policy
	resp, err := c.networkloadbalancer.UpdateBackendSet(ctx, networkloadbalancer.UpdateBackendSetRequest{
		NetworkLoadBalancerId: &lbID,
		BackendSetName:        &name,
		UpdateBackendSetDetails: networkloadbalancer.UpdateBackendSetDetails{
			Backends:         backendsToBackendDetails(details.Backends),
			IsPreserveSource: details.IsPreserveSource,
			HealthChecker:    healthCheckerToHealthCheckerDetails(details.HealthChecker),
			Policy:           stringPolicy,
		},
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, updateVerb, backendSetResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *networkLoadbalancer) DeleteBackendSet(ctx context.Context, lbID, name string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "DeleteBackendSet")
	}

	resp, err := c.networkloadbalancer.DeleteBackendSet(ctx, networkloadbalancer.DeleteBackendSetRequest{
		NetworkLoadBalancerId: &lbID,
		BackendSetName:        &name,
		RequestMetadata:       c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, backendSetResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *networkLoadbalancer) CreateListener(ctx context.Context, lbID string, name string, details *GenericListener) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateListener")
	}

	resp, err := c.networkloadbalancer.CreateListener(ctx, networkloadbalancer.CreateListenerRequest{
		NetworkLoadBalancerId: &lbID,
		CreateListenerDetails: networkloadbalancer.CreateListenerDetails{
			Name:                  &name,
			DefaultBackendSetName: details.DefaultBackendSetName,
			Port:                  details.Port,
			Protocol:              networkloadbalancer.ListenerProtocolsEnum(*details.Protocol),
		},
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, createVerb, listenerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *networkLoadbalancer) UpdateListener(ctx context.Context, lbID string, name string, details *GenericListener) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "UpdateListener")
	}

	resp, err := c.networkloadbalancer.UpdateListener(ctx, networkloadbalancer.UpdateListenerRequest{
		NetworkLoadBalancerId: &lbID,
		ListenerName:          &name,
		UpdateListenerDetails: networkloadbalancer.UpdateListenerDetails{
			DefaultBackendSetName: details.DefaultBackendSetName,
			Port:                  details.Port,
			Protocol:              networkloadbalancer.ListenerProtocolsEnum(*details.Protocol),
		},
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, updateVerb, listenerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *networkLoadbalancer) AwaitWorkRequest(ctx context.Context, id string) (*GenericWorkRequest, error) {
	var wr *networkloadbalancer.WorkRequest
	contextWithTimeout, cancel := context.WithTimeout(ctx, defaultSynchronousAPIContextTimeout)
	defer cancel()
	err := wait.PollUntil(workRequestPollInterval, func() (done bool, err error) {
		twr, err := c.GetWorkRequest(contextWithTimeout, id)
		if err != nil {
			if IsRetryable(err) {
				return false, nil
			}
			return true, errors.WithStack(err)
		}
		switch twr.Status {
		case networkloadbalancer.OperationStatusSucceeded:
			wr = twr
			return true, nil
		case networkloadbalancer.OperationStatusFailed:
			return false, errors.Errorf("WorkRequest %q failed. PercentComplete: %f", id, *twr.PercentComplete)
		}
		return false, nil
	}, ctx.Done())

	return c.workRequestToGenericWorkRequest(wr), err
}

func (c *networkLoadbalancer) DeleteListener(ctx context.Context, lbID, name string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "DeleteListener")
	}

	resp, err := c.networkloadbalancer.DeleteListener(ctx, networkloadbalancer.DeleteListenerRequest{
		NetworkLoadBalancerId: &lbID,
		ListenerName:          &name,
		RequestMetadata:       c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, listenerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *networkLoadbalancer) UpdateLoadBalancerShape(context.Context, string, *GenericUpdateLoadBalancerShapeDetails) (string, error) {
	return "", nil
}

func (c *networkLoadbalancer) UpdateNetworkSecurityGroups(ctx context.Context, lbID string, lbNetworkSecurityGroupDetails []string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "UpdateNetworkSecurityGroups")
	}

	resp, err := c.networkloadbalancer.UpdateNetworkSecurityGroups(ctx, networkloadbalancer.UpdateNetworkSecurityGroupsRequest{
		NetworkLoadBalancerId: &lbID,
		UpdateNetworkSecurityGroupsDetails: networkloadbalancer.UpdateNetworkSecurityGroupsDetails{
			NetworkSecurityGroupIds: lbNetworkSecurityGroupDetails,
		},
	})
	incRequestCounter(err, updateVerb, nsgResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func backendsToBackendDetails(backends []GenericBackend) []networkloadbalancer.BackendDetails {
	backendDetails := make([]networkloadbalancer.BackendDetails, 0)
	for _, backend := range backends {
		backendDetails = append(backendDetails, networkloadbalancer.BackendDetails{
			Port:      backend.Port,
			Name:      backend.Name,
			IpAddress: backend.IpAddress,
			TargetId:  backend.TargetId,
			Weight:    backend.Weight,
		})

	}
	return backendDetails
}

func healthCheckerToHealthCheckerDetails(healthChecker *GenericHealthChecker) *networkloadbalancer.HealthCheckerDetails {
	healthCheckerDetails := networkloadbalancer.HealthCheckerDetails{
		Port:              healthChecker.Port,
		Protocol:          networkloadbalancer.HealthCheckProtocolsEnum(healthChecker.Protocol),
		Retries:           healthChecker.Retries,
		ReturnCode:        healthChecker.ReturnCode,
		ResponseBodyRegex: healthChecker.ResponseBodyRegex,
		TimeoutInMillis:   healthChecker.TimeoutInMillis,
		IntervalInMillis:  healthChecker.IntervalInMillis,
		UrlPath:           healthChecker.UrlPath,
	}

	return &healthCheckerDetails
}

func (c *networkLoadbalancer) networkLoadbalancerToGenericLoadbalancer(nlb *networkloadbalancer.NetworkLoadBalancer) *GenericLoadBalancer {
	lifecycleState := string(nlb.LifecycleState)
	return &GenericLoadBalancer{
		Id:                      nlb.Id,
		CompartmentId:           nlb.CompartmentId,
		DisplayName:             nlb.DisplayName,
		LifecycleState:          &lifecycleState,
		IpAddresses:             c.ipAddressesToGenericIpAddress(nlb.IpAddresses),
		IsPrivate:               nlb.IsPrivate,
		SubnetIds:               []string{*nlb.SubnetId},
		NetworkSecurityGroupIds: nlb.NetworkSecurityGroupIds,
		Listeners:               c.listenersToGenericListenerDetails(nlb.Listeners),
		BackendSets:             c.backendSetsToGenericBackendSetDetails(nlb.BackendSets),
		FreeformTags:            nlb.FreeformTags,
		DefinedTags:             nlb.DefinedTags,
	}
}

func (c *networkLoadbalancer) networkLoadbalancerSummaryToGenericLoadbalancer(nlb *networkloadbalancer.NetworkLoadBalancerSummary) *GenericLoadBalancer {
	lifecycleState := string(nlb.LifecycleState)
	return &GenericLoadBalancer{
		Id:                      nlb.Id,
		CompartmentId:           nlb.CompartmentId,
		DisplayName:             nlb.DisplayName,
		LifecycleState:          &lifecycleState,
		IpAddresses:             c.ipAddressesToGenericIpAddress(nlb.IpAddresses),
		IsPrivate:               nlb.IsPrivate,
		SubnetIds:               []string{*nlb.SubnetId},
		NetworkSecurityGroupIds: nlb.NetworkSecurityGroupIds,
		Listeners:               c.listenersToGenericListenerDetails(nlb.Listeners),
		BackendSets:             c.backendSetsToGenericBackendSetDetails(nlb.BackendSets),
		FreeformTags:            nlb.FreeformTags,
		DefinedTags:             nlb.DefinedTags,
	}
}

func (c *networkLoadbalancer) ipAddressesToGenericIpAddress(ipAddresses []networkloadbalancer.IpAddress) []GenericIpAddress {
	genericIPAddresses := make([]GenericIpAddress, 0)
	for _, address := range ipAddresses {
		genericIPAddresses = append(genericIPAddresses, GenericIpAddress{
			IpAddress:  address.IpAddress,
			IsPublic:   address.IsPublic,
			ReservedIp: (*GenericReservedIp)(address.ReservedIp),
		})
	}
	return genericIPAddresses
}

func (c *networkLoadbalancer) listenersToGenericListenerDetails(details map[string]networkloadbalancer.Listener) map[string]GenericListener {
	genericListenerDetails := make(map[string]GenericListener)

	for k, v := range details {
		protocol := string(v.Protocol)
		genericListenerDetails[k] = GenericListener{
			Name:                  v.Name,
			DefaultBackendSetName: v.DefaultBackendSetName,
			Port:                  v.Port,
			Protocol:              &protocol,
		}
	}
	return genericListenerDetails
}

func (c *networkLoadbalancer) backendSetsToGenericBackendSetDetails(backendSets map[string]networkloadbalancer.BackendSet) map[string]GenericBackendSetDetails {
	genericBackendSetDetails := make(map[string]GenericBackendSetDetails)

	for k, v := range backendSets {
		policyString := string(v.Policy)
		genericBackendSetDetails[k] = GenericBackendSetDetails{
			HealthChecker: &GenericHealthChecker{
				Protocol:         string(v.HealthChecker.Protocol),
				Port:             v.HealthChecker.Port,
				UrlPath:          v.HealthChecker.UrlPath,
				Retries:          v.HealthChecker.Retries,
				ReturnCode:       v.HealthChecker.ReturnCode,
				TimeoutInMillis:  v.HealthChecker.TimeoutInMillis,
				IntervalInMillis: v.HealthChecker.IntervalInMillis,
			},
			Name:             v.Name,
			Policy:           &policyString,
			Backends:         c.backendDetailsToGenericBackendDetails(v.Backends),
			IsPreserveSource: v.IsPreserveSource,
		}
	}

	return genericBackendSetDetails
}

func (c *networkLoadbalancer) backendDetailsToGenericBackendDetails(details []networkloadbalancer.Backend) []GenericBackend {
	genericBackendDetails := make([]GenericBackend, 0)

	for _, backends := range details {
		genericBackendDetails = append(genericBackendDetails, GenericBackend{
			IpAddress: backends.IpAddress,
			Port:      backends.Port,
			Weight:    backends.Weight,
			TargetId:  backends.TargetId,
		})
	}
	return genericBackendDetails
}

func getNetworkLoadBalancerID(request *networkloadbalancer.WorkRequest) *string {
	var networkLoadBalancerID *string
	for _, resource := range request.Resources {
		if *resource.EntityType == NetworkLoadBalancerEntityType {
			networkLoadBalancerID = resource.Identifier
			break
		}
	}
	return networkLoadBalancerID
}

func (c *networkLoadbalancer) workRequestToGenericWorkRequest(request *networkloadbalancer.WorkRequest) *GenericWorkRequest {
	if request == nil {
		return nil
	}
	genericWorkRequest := &GenericWorkRequest{
		Id:             request.Id,
		LoadBalancerId: getNetworkLoadBalancerID(request),
		OperationType:  string(request.OperationType),
		Status:         string(request.Status),
		CompartmentId:  request.CompartmentId,
	}
	return genericWorkRequest
}

func (c *networkLoadbalancer) genericReservedIpToReservedIps(genericReservedIps []GenericReservedIp) []networkloadbalancer.ReservedIp {
	reservedIps := make([]networkloadbalancer.ReservedIp, 0)
	for _, address := range genericReservedIps {
		reservedIps = append(reservedIps, networkloadbalancer.ReservedIp{
			Id: address.Id,
		})
	}
	return reservedIps
}

func (c *networkLoadbalancer) genericListenerDetailsToListenerDetails(details map[string]GenericListener) map[string]networkloadbalancer.ListenerDetails {
	listenerDetails := make(map[string]networkloadbalancer.ListenerDetails)

	for k, v := range details {
		listenerDetails[k] = networkloadbalancer.ListenerDetails{
			Name:                  v.Name,
			DefaultBackendSetName: v.DefaultBackendSetName,
			Port:                  v.Port,
			Protocol:              networkloadbalancer.ListenerProtocolsEnum(*v.Protocol),
		}
	}
	return listenerDetails
}

func (c *networkLoadbalancer) genericBackendSetDetailsToBackendSets(backendSets map[string]GenericBackendSetDetails) map[string]networkloadbalancer.BackendSetDetails {
	backendSetDetails := make(map[string]networkloadbalancer.BackendSetDetails)

	for k, v := range backendSets {
		backendSetDetails[k] = networkloadbalancer.BackendSetDetails{
			HealthChecker: &networkloadbalancer.HealthChecker{
				Protocol:         networkloadbalancer.HealthCheckProtocolsEnum(v.HealthChecker.Protocol),
				Port:             v.HealthChecker.Port,
				UrlPath:          v.HealthChecker.UrlPath,
				Retries:          v.HealthChecker.Retries,
				ReturnCode:       v.HealthChecker.ReturnCode,
				TimeoutInMillis:  v.HealthChecker.TimeoutInMillis,
				IntervalInMillis: v.HealthChecker.IntervalInMillis,
			},
			Policy:           networkloadbalancer.NetworkLoadBalancingPolicyEnum(*v.Policy),
			Backends:         c.genericBackendDetailsToBackendDetails(v.Backends),
			IsPreserveSource: v.IsPreserveSource,
		}
	}
	return backendSetDetails
}

func (c *networkLoadbalancer) genericBackendDetailsToBackendDetails(details []GenericBackend) []networkloadbalancer.Backend {
	backendDetails := make([]networkloadbalancer.Backend, 0)

	for _, backends := range details {
		backendDetails = append(backendDetails, networkloadbalancer.Backend{
			IpAddress: backends.IpAddress,
			Port:      backends.Port,
			Weight:    backends.Weight,
			TargetId:  backends.TargetId,
		})
	}
	return backendDetails
}
