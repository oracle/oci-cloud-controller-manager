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

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/loadbalancer"
	"github.com/pkg/errors"
)

const (
	workRequestPollInterval = 5 * time.Second
)

type loadbalancerClientStruct struct {
	loadbalancer    loadBalancerClient
	requestMetadata common.RequestMetadata
	rateLimiter     RateLimiter
}

type GenericLoadBalancerInterface interface {
	CreateLoadBalancer(ctx context.Context, details *GenericCreateLoadBalancerDetails) (string, error)

	GetLoadBalancer(ctx context.Context, id string) (*GenericLoadBalancer, error)
	GetLoadBalancerByName(ctx context.Context, compartmentID, name string) (*GenericLoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, id string) (string, error)

	GetCertificateByName(ctx context.Context, lbID, name string) (*GenericCertificate, error)
	CreateCertificate(ctx context.Context, lbID string, cert *GenericCertificate) (string, error)

	CreateBackendSet(ctx context.Context, lbID, name string, details *GenericBackendSetDetails) (string, error)
	UpdateBackendSet(ctx context.Context, lbID, name string, details *GenericBackendSetDetails) (string, error)
	DeleteBackendSet(ctx context.Context, lbID, name string) (string, error)

	UpdateListener(ctx context.Context, lbID, name string, details *GenericListener) (string, error)
	CreateListener(ctx context.Context, lbID, name string, details *GenericListener) (string, error)
	DeleteListener(ctx context.Context, lbID, name string) (string, error)

	UpdateLoadBalancerShape(context.Context, string, *GenericUpdateLoadBalancerShapeDetails) (string, error)
	UpdateNetworkSecurityGroups(context.Context, string, []string) (string, error)

	AwaitWorkRequest(ctx context.Context, id string) (*GenericWorkRequest, error)
}

func (c *loadbalancerClientStruct) GetLoadBalancer(ctx context.Context, id string) (*GenericLoadBalancer, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetLoadBalancer")
	}

	resp, err := c.loadbalancer.GetLoadBalancer(ctx, loadbalancer.GetLoadBalancerRequest{
		LoadBalancerId:  &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, loadBalancerResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return c.loadbalancerToGenericLoadbalancer(&resp.LoadBalancer), nil
}

func (c *loadbalancerClientStruct) GetLoadBalancerByName(ctx context.Context, compartmentID, name string) (*GenericLoadBalancer, error) {
	var page *string
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListLoadBalancers")
		}
		resp, err := c.loadbalancer.ListLoadBalancers(ctx, loadbalancer.ListLoadBalancersRequest{
			CompartmentId:   &compartmentID,
			DisplayName:     &name,
			Page:            page,
			RequestMetadata: c.requestMetadata,
		})
		incRequestCounter(err, listVerb, loadBalancerResource)

		if err != nil {
			return nil, errors.WithStack(err)
		}
		for _, lb := range resp.Items {
			if *lb.DisplayName == name {
				return c.loadbalancerToGenericLoadbalancer(&lb), nil
			}
		}
		if page = resp.OpcNextPage; page == nil {
			break
		}
	}

	return nil, errors.WithStack(errNotFound)
}

func (c *loadbalancerClientStruct) CreateLoadBalancer(ctx context.Context, details *GenericCreateLoadBalancerDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateLoadBalancer")
	}

	resp, err := c.loadbalancer.CreateLoadBalancer(ctx, loadbalancer.CreateLoadBalancerRequest{
		CreateLoadBalancerDetails: loadbalancer.CreateLoadBalancerDetails{
			CompartmentId:           details.CompartmentId,
			DisplayName:             details.DisplayName,
			SubnetIds:               details.SubnetIds,
			ShapeName:               details.ShapeName,
			ShapeDetails:            c.genericShapeDetailsToShapeDetails(details.ShapeDetails),
			ReservedIps:             c.genericReservedIpToReservedIps(details.ReservedIps),
			Certificates:            c.genericCertificatesToCertificates(details.Certificates),
			IsPrivate:               details.IsPrivate,
			NetworkSecurityGroupIds: details.NetworkSecurityGroupIds,
			Listeners:               c.genericListenerDetailsToListenerDetails(details.Listeners),
			BackendSets:             c.genericBackendSetDetailsToBackendSets(details.BackendSets),
			FreeformTags:            details.FreeformTags,
			DefinedTags:             details.DefinedTags,
		},
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, createVerb, loadBalancerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "DeleteLoadBalancer")
	}

	resp, err := c.loadbalancer.DeleteLoadBalancer(ctx, loadbalancer.DeleteLoadBalancerRequest{
		LoadBalancerId:  &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, loadBalancerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) GetCertificateByName(ctx context.Context, lbID, name string) (*GenericCertificate, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "ListCertificates")
	}

	resp, err := c.loadbalancer.ListCertificates(ctx, loadbalancer.ListCertificatesRequest{
		LoadBalancerId:  &lbID,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, listVerb, certificateResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, cert := range resp.Items {
		if *cert.CertificateName == name {
			return certificateToGenericCertificate(&cert), nil
		}
	}
	return nil, errors.WithStack(errNotFound)
}

func (c *loadbalancerClientStruct) CreateCertificate(ctx context.Context, lbID string, cert *GenericCertificate) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateCertificate")
	}

	resp, err := c.loadbalancer.CreateCertificate(ctx, loadbalancer.CreateCertificateRequest{
		LoadBalancerId: &lbID,
		CreateCertificateDetails: loadbalancer.CreateCertificateDetails{
			CertificateName:   cert.CertificateName,
			CaCertificate:     cert.CaCertificate,
			PublicCertificate: cert.PublicCertificate,
			PrivateKey:        cert.PrivateKey,
			Passphrase:        cert.Passphrase,
		},
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, createVerb, certificateResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) GetWorkRequest(ctx context.Context, id string) (*loadbalancer.WorkRequest, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetWorkRequest")
	}

	resp, err := c.loadbalancer.GetWorkRequest(ctx, loadbalancer.GetWorkRequestRequest{
		WorkRequestId:   &id,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, getVerb, workRequestResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.WorkRequest, nil
}

func (c *loadbalancerClientStruct) CreateBackendSet(ctx context.Context, lbID string, name string, details *GenericBackendSetDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateBackendSet")
	}
	createBackendSetRequest := loadbalancer.CreateBackendSetRequest{
		LoadBalancerId: &lbID,
		CreateBackendSetDetails: loadbalancer.CreateBackendSetDetails{
			Name:     &name,
			Backends: c.genericBackendDetailsToBackendDetails(details.Backends),
			HealthChecker: &loadbalancer.HealthCheckerDetails{
				Protocol:         &details.HealthChecker.Protocol,
				Port:             details.HealthChecker.Port,
				UrlPath:          details.HealthChecker.UrlPath,
				Retries:          details.HealthChecker.Retries,
				ReturnCode:       details.HealthChecker.ReturnCode,
				TimeoutInMillis:  details.HealthChecker.TimeoutInMillis,
				IntervalInMillis: details.HealthChecker.IntervalInMillis,
			},
			Policy:                          details.Policy,
			SessionPersistenceConfiguration: getSessionPersistenceConfiguration(details.SessionPersistenceConfiguration),
		},
		RequestMetadata: c.requestMetadata,
	}
	if details.SslConfiguration != nil {
		createBackendSetRequest.SslConfiguration = genericSslConfigurationToSslConfiguration(details.SslConfiguration)
	}
	resp, err := c.loadbalancer.CreateBackendSet(ctx, createBackendSetRequest)

	incRequestCounter(err, createVerb, backendSetResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) UpdateBackendSet(ctx context.Context, lbID string, name string, details *GenericBackendSetDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "UpdateBackendSet")
	}

	updateBackendSetRequest := loadbalancer.UpdateBackendSetRequest{
		LoadBalancerId: &lbID,
		BackendSetName: &name,
		UpdateBackendSetDetails: loadbalancer.UpdateBackendSetDetails{
			Backends: c.genericBackendDetailsToBackendDetails(details.Backends),
			HealthChecker: &loadbalancer.HealthCheckerDetails{
				Protocol:         &details.HealthChecker.Protocol,
				Port:             details.HealthChecker.Port,
				UrlPath:          details.HealthChecker.UrlPath,
				Retries:          details.HealthChecker.Retries,
				ReturnCode:       details.HealthChecker.ReturnCode,
				TimeoutInMillis:  details.HealthChecker.TimeoutInMillis,
				IntervalInMillis: details.HealthChecker.IntervalInMillis,
			},
			Policy:                          details.Policy,
			SessionPersistenceConfiguration: getSessionPersistenceConfiguration(details.SessionPersistenceConfiguration),
		},
		RequestMetadata: c.requestMetadata,
	}

	if details.SslConfiguration != nil {
		updateBackendSetRequest.SslConfiguration = genericSslConfigurationToSslConfiguration(details.SslConfiguration)
	}
	resp, err := c.loadbalancer.UpdateBackendSet(ctx, updateBackendSetRequest)

	incRequestCounter(err, updateVerb, backendSetResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) DeleteBackendSet(ctx context.Context, lbID, name string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "DeleteBackendSet")
	}

	resp, err := c.loadbalancer.DeleteBackendSet(ctx, loadbalancer.DeleteBackendSetRequest{
		LoadBalancerId:  &lbID,
		BackendSetName:  &name,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, backendSetResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) CreateListener(ctx context.Context, lbID string, name string, details *GenericListener) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateListener")
	}

	createListener := loadbalancer.CreateListenerRequest{
		LoadBalancerId: &lbID,
		CreateListenerDetails: loadbalancer.CreateListenerDetails{
			Name:                    &name,
			DefaultBackendSetName:   details.DefaultBackendSetName,
			Port:                    details.Port,
			Protocol:                details.Protocol,
			ConnectionConfiguration: getListenerConnectionConfiguration(details.ConnectionConfiguration),
		},
		RequestMetadata: c.requestMetadata,
	}
	if details.SslConfiguration != nil {
		createListener.SslConfiguration = genericSslConfigurationToSslConfiguration(details.SslConfiguration)
	}

	resp, err := c.loadbalancer.CreateListener(ctx, createListener)

	incRequestCounter(err, createVerb, listenerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) UpdateListener(ctx context.Context, lbID string, name string, details *GenericListener) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "UpdateListener")
	}

	updateListenerRequest := loadbalancer.UpdateListenerRequest{
		LoadBalancerId: &lbID,
		ListenerName:   &name,
		UpdateListenerDetails: loadbalancer.UpdateListenerDetails{
			DefaultBackendSetName: details.DefaultBackendSetName,
			Port:                  details.Port,
			Protocol:              details.Protocol,
		},
		RequestMetadata: c.requestMetadata,
	}

	if details.SslConfiguration != nil {
		updateListenerRequest.SslConfiguration = genericSslConfigurationToSslConfiguration(details.SslConfiguration)
	}

	if details.ConnectionConfiguration != nil {
		updateListenerRequest.ConnectionConfiguration = getListenerConnectionConfiguration(details.ConnectionConfiguration)
	}

	resp, err := c.loadbalancer.UpdateListener(ctx, updateListenerRequest)

	incRequestCounter(err, updateVerb, listenerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) AwaitWorkRequest(ctx context.Context, id string) (*GenericWorkRequest, error) {
	var wr *loadbalancer.WorkRequest
	contextWithTimeout, cancel := context.WithTimeout(ctx, defaultSynchronousAPIContextTimeout)
	defer cancel()
	requestId, _ := generateRandUUID()
	err := wait.PollUntil(workRequestPollInterval, func() (done bool, err error) {
		twr, err := c.GetWorkRequest(contextWithTimeout, id)
		if err != nil {
			if IsRetryable(err) {
				return false, nil
			}
			return true, errors.Wrapf(errors.WithStack(err), "failed to get workrequest. opc-request-id: %s", requestId)
		}
		switch twr.LifecycleState {
		case loadbalancer.WorkRequestLifecycleStateSucceeded:
			wr = twr
			return true, nil
		case loadbalancer.WorkRequestLifecycleStateFailed:
			return false, errors.Errorf("WorkRequest %q failed: %s", id, *twr.Message)
		}
		return false, nil
	}, ctx.Done())

	return c.workRequestToGenericWorkRequest(wr), err
}

func (c *loadbalancerClientStruct) DeleteListener(ctx context.Context, lbID, name string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "DeleteListener")
	}

	resp, err := c.loadbalancer.DeleteListener(ctx, loadbalancer.DeleteListenerRequest{
		LoadBalancerId:  &lbID,
		ListenerName:    &name,
		RequestMetadata: c.requestMetadata,
	})
	incRequestCounter(err, deleteVerb, listenerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) UpdateLoadBalancerShape(ctx context.Context, lbID string, lbShapeDetails *GenericUpdateLoadBalancerShapeDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "UpdateListener")
	}

	resp, err := c.loadbalancer.UpdateLoadBalancerShape(ctx, loadbalancer.UpdateLoadBalancerShapeRequest{
		LoadBalancerId: &lbID,
		UpdateLoadBalancerShapeDetails: loadbalancer.UpdateLoadBalancerShapeDetails{
			ShapeName:    lbShapeDetails.ShapeName,
			ShapeDetails: c.genericShapeDetailsToShapeDetails(lbShapeDetails.ShapeDetails),
		},
	})
	incRequestCounter(err, updateVerb, shapeResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) UpdateNetworkSecurityGroups(ctx context.Context, lbID string, lbNetworkSecurityGroupDetails []string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "UpdateNetworkSecurityGroups")
	}

	resp, err := c.loadbalancer.UpdateNetworkSecurityGroups(ctx, loadbalancer.UpdateNetworkSecurityGroupsRequest{
		LoadBalancerId: &lbID,
		UpdateNetworkSecurityGroupsDetails: loadbalancer.UpdateNetworkSecurityGroupsDetails{
			NetworkSecurityGroupIds: lbNetworkSecurityGroupDetails,
		},
	})
	incRequestCounter(err, updateVerb, nsgResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *loadbalancerClientStruct) loadbalancerToGenericLoadbalancer(lb *loadbalancer.LoadBalancer) *GenericLoadBalancer {
	if lb == nil {
		return nil
	}
	lifecycleState := string(lb.LifecycleState)
	return &GenericLoadBalancer{
		Id:                      lb.Id,
		CompartmentId:           lb.CompartmentId,
		DisplayName:             lb.DisplayName,
		LifecycleState:          &lifecycleState,
		ShapeName:               lb.ShapeName,
		IpAddresses:             c.ipAddressesToGenericIpAddress(lb.IpAddresses),
		ShapeDetails:            shapeDetailsToGenericShapeDetails(lb.ShapeDetails),
		IsPrivate:               lb.IsPrivate,
		SubnetIds:               lb.SubnetIds,
		NetworkSecurityGroupIds: lb.NetworkSecurityGroupIds,
		Listeners:               c.listenersToGenericListenerDetails(lb.Listeners),
		Certificates:            c.certificateToGenericCertificateDetails(lb.Certificates),
		BackendSets:             c.backendSetsToGenericBackendSetDetails(lb.BackendSets),
		FreeformTags:            lb.FreeformTags,
		DefinedTags:             lb.DefinedTags,
	}
}

func (c *loadbalancerClientStruct) ipAddressesToGenericIpAddress(ipAddresses []loadbalancer.IpAddress) []GenericIpAddress {
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

func shapeDetailsToGenericShapeDetails(shapeDetails *loadbalancer.ShapeDetails) *GenericShapeDetails {
	genericShapeDetails := &GenericShapeDetails{}
	if shapeDetails == nil {
		return genericShapeDetails
	}
	if shapeDetails.MinimumBandwidthInMbps != nil {
		genericShapeDetails.MinimumBandwidthInMbps = shapeDetails.MinimumBandwidthInMbps
	}
	if shapeDetails.MaximumBandwidthInMbps != nil {
		genericShapeDetails.MaximumBandwidthInMbps = shapeDetails.MaximumBandwidthInMbps
	}
	return genericShapeDetails
}

func (c *loadbalancerClientStruct) listenersToGenericListenerDetails(details map[string]loadbalancer.Listener) map[string]GenericListener {
	genericListenerDetails := make(map[string]GenericListener)

	for k, v := range details {
		listenerDetailsStruct := GenericListener{
			Name:                  v.Name,
			DefaultBackendSetName: v.DefaultBackendSetName,
			Port:                  v.Port,
			Protocol:              v.Protocol,
			HostnameNames:         v.HostnameNames,
			PathRouteSetName:      v.PathRouteSetName,
			RoutingPolicyName:     v.RoutingPolicyName,
			RuleSetNames:          v.RuleSetNames,
		}

		if v.SslConfiguration != nil {
			listenerDetailsStruct.SslConfiguration = sslConfigurationToGenericSslConfiguration(v.SslConfiguration)
		}

		if v.ConnectionConfiguration != nil {
			listenerDetailsStruct.ConnectionConfiguration = &GenericConnectionConfiguration{
				IdleTimeout:                    v.ConnectionConfiguration.IdleTimeout,
				BackendTcpProxyProtocolVersion: v.ConnectionConfiguration.BackendTcpProxyProtocolVersion,
			}
		}
		genericListenerDetails[k] = listenerDetailsStruct

	}
	return genericListenerDetails
}

func (c *loadbalancerClientStruct) genericListenerDetailsToListenerDetails(details map[string]GenericListener) map[string]loadbalancer.ListenerDetails {
	listenerDetails := make(map[string]loadbalancer.ListenerDetails)

	for k, v := range details {
		listenerDetailsStruct := loadbalancer.ListenerDetails{
			DefaultBackendSetName: v.DefaultBackendSetName,
			Port:                  v.Port,
			Protocol:              v.Protocol,
			HostnameNames:         v.HostnameNames,
		}

		if v.RuleSetNames != nil {
			listenerDetailsStruct.RuleSetNames = v.RuleSetNames
		}
		if v.RoutingPolicyName != nil {
			listenerDetailsStruct.RoutingPolicyName = v.RoutingPolicyName
		}

		if v.PathRouteSetName != nil {
			listenerDetailsStruct.PathRouteSetName = v.PathRouteSetName
		}

		if v.SslConfiguration != nil {
			listenerDetailsStruct.SslConfiguration = genericSslConfigurationToSslConfiguration(v.SslConfiguration)
		}

		if v.ConnectionConfiguration != nil {
			listenerDetailsStruct.ConnectionConfiguration = &loadbalancer.ConnectionConfiguration{
				IdleTimeout:                    v.ConnectionConfiguration.IdleTimeout,
				BackendTcpProxyProtocolVersion: v.ConnectionConfiguration.BackendTcpProxyProtocolVersion,
			}
		}
		listenerDetails[k] = listenerDetailsStruct

	}
	return listenerDetails
}

func (c *loadbalancerClientStruct) certificateToGenericCertificateDetails(certificates map[string]loadbalancer.Certificate) map[string]GenericCertificate {
	genericCertificateDetails := make(map[string]GenericCertificate)

	for k, v := range certificates {
		genericCertificateDetails[k] = GenericCertificate{
			CertificateName:   v.CertificateName,
			PublicCertificate: v.PublicCertificate,
			CaCertificate:     v.CaCertificate,
		}
	}
	return genericCertificateDetails
}

func (c *loadbalancerClientStruct) backendSetsToGenericBackendSetDetails(backendSets map[string]loadbalancer.BackendSet) map[string]GenericBackendSetDetails {
	genericBackendSetDetails := make(map[string]GenericBackendSetDetails)

	for k, v := range backendSets {
		backendDetailsStruct := GenericBackendSetDetails{
			HealthChecker: &GenericHealthChecker{
				Protocol:         *v.HealthChecker.Protocol,
				Port:             v.HealthChecker.Port,
				UrlPath:          v.HealthChecker.UrlPath,
				Retries:          v.HealthChecker.Retries,
				ReturnCode:       v.HealthChecker.ReturnCode,
				TimeoutInMillis:  v.HealthChecker.TimeoutInMillis,
				IntervalInMillis: v.HealthChecker.IntervalInMillis,
			},
			Policy:   v.Policy,
			Name:     v.Name,
			Backends: backendDetailsToGenericBackendDetails(v.Backends),
		}

		if v.SslConfiguration != nil {
			backendDetailsStruct.SslConfiguration = sslConfigurationToGenericSslConfiguration(v.SslConfiguration)
		}

		if v.SessionPersistenceConfiguration != nil {
			backendDetailsStruct.SessionPersistenceConfiguration = getGenericSessionPersistenceConfiguration(v.SessionPersistenceConfiguration)
		}
		genericBackendSetDetails[k] = backendDetailsStruct
	}

	return genericBackendSetDetails
}

func (c *loadbalancerClientStruct) genericBackendSetDetailsToBackendSets(backendSets map[string]GenericBackendSetDetails) map[string]loadbalancer.BackendSetDetails {
	backendSetDetails := make(map[string]loadbalancer.BackendSetDetails)

	for k, v := range backendSets {
		backendSetDetailsStruct := loadbalancer.BackendSetDetails{
			HealthChecker: &loadbalancer.HealthCheckerDetails{
				Protocol:         &v.HealthChecker.Protocol,
				Port:             v.HealthChecker.Port,
				UrlPath:          v.HealthChecker.UrlPath,
				Retries:          v.HealthChecker.Retries,
				ReturnCode:       v.HealthChecker.ReturnCode,
				TimeoutInMillis:  v.HealthChecker.TimeoutInMillis,
				IntervalInMillis: v.HealthChecker.IntervalInMillis,
			},
			Policy:   v.Policy,
			Backends: c.genericBackendDetailsToBackendDetails(v.Backends),
		}

		if v.SslConfiguration != nil {
			backendSetDetailsStruct.SslConfiguration = genericSslConfigurationToSslConfiguration(v.SslConfiguration)
		}

		if v.SessionPersistenceConfiguration != nil {
			backendSetDetailsStruct.SessionPersistenceConfiguration = getSessionPersistenceConfiguration(v.SessionPersistenceConfiguration)
		}
		backendSetDetails[k] = backendSetDetailsStruct
	}
	return backendSetDetails
}

func (c *loadbalancerClientStruct) workRequestToGenericWorkRequest(request *loadbalancer.WorkRequest) *GenericWorkRequest {
	if request == nil {
		return nil
	}
	lifecycleState := string(request.LifecycleState)
	genericWorkRequest := &GenericWorkRequest{
		Id:             request.Id,
		LoadBalancerId: request.LoadBalancerId,
		LifecycleState: &lifecycleState,
		CompartmentId:  request.CompartmentId,
		Message:        request.Message,
		OperationType:  *request.Type,
	}
	return genericWorkRequest
}

func certificateToGenericCertificate(certificate *loadbalancer.Certificate) *GenericCertificate {
	if certificate == nil {
		return nil
	}
	genericCertificateDetails := &GenericCertificate{
		CertificateName:   certificate.CertificateName,
		PublicCertificate: certificate.PublicCertificate,
		CaCertificate:     certificate.CaCertificate,
	}
	return genericCertificateDetails
}

func (c *loadbalancerClientStruct) genericCertificatesToCertificates(genericCertificates map[string]GenericCertificate) map[string]loadbalancer.CertificateDetails {
	certificates := make(map[string]loadbalancer.CertificateDetails)

	for k, cert := range genericCertificates {
		certStruct := loadbalancer.CertificateDetails{
			CertificateName:   cert.CertificateName,
			Passphrase:        cert.Passphrase,
			PrivateKey:        cert.PrivateKey,
			PublicCertificate: cert.PublicCertificate,
			CaCertificate:     cert.CaCertificate,
		}
		certificates[k] = certStruct
	}

	return certificates
}

func (c *loadbalancerClientStruct) genericShapeDetailsToShapeDetails(details *GenericShapeDetails) *loadbalancer.ShapeDetails {
	if details == nil {
		return nil
	}
	return &loadbalancer.ShapeDetails{
		MinimumBandwidthInMbps: details.MinimumBandwidthInMbps,
		MaximumBandwidthInMbps: details.MaximumBandwidthInMbps,
	}
}

func (c *loadbalancerClientStruct) genericReservedIpToReservedIps(genericReservedIps []GenericReservedIp) []loadbalancer.ReservedIp {
	reservedIps := make([]loadbalancer.ReservedIp, 0)
	for _, address := range genericReservedIps {
		reservedIps = append(reservedIps, loadbalancer.ReservedIp{
			Id: address.Id,
		})
	}
	return reservedIps
}

func (c *loadbalancerClientStruct) genericBackendDetailsToBackendDetails(details []GenericBackend) []loadbalancer.BackendDetails {
	backendDetails := make([]loadbalancer.BackendDetails, 0)

	for _, backends := range details {
		backendDetails = append(backendDetails, loadbalancer.BackendDetails{
			IpAddress: backends.IpAddress,
			Port:      backends.Port,
			Weight:    backends.Weight,
		})
	}
	return backendDetails
}

func backendDetailsToGenericBackendDetails(details []loadbalancer.Backend) []GenericBackend {
	genericBackendDetails := make([]GenericBackend, 0)

	for _, backends := range details {
		genericBackendDetails = append(genericBackendDetails, GenericBackend{
			IpAddress: backends.IpAddress,
			Port:      backends.Port,
			Weight:    backends.Weight,
		})
	}
	return genericBackendDetails
}

func genericSslConfigurationToSslConfiguration(details *GenericSslConfigurationDetails) *loadbalancer.SslConfigurationDetails {
	if details == nil {
		return nil
	}
	return &loadbalancer.SslConfigurationDetails{
		VerifyDepth:                    details.VerifyDepth,
		VerifyPeerCertificate:          details.VerifyPeerCertificate,
		TrustedCertificateAuthorityIds: details.TrustedCertificateAuthorityIds,
		CertificateIds:                 details.CertificateIds,
		CertificateName:                details.CertificateName,
		ServerOrderPreference:          loadbalancer.SslConfigurationDetailsServerOrderPreferenceEnum(details.ServerOrderPreference),
		CipherSuiteName:                details.CipherSuiteName,
		Protocols:                      details.Protocols,
	}
}

func sslConfigurationToGenericSslConfiguration(details *loadbalancer.SslConfiguration) *GenericSslConfigurationDetails {
	if details == nil {
		return nil
	}
	return &GenericSslConfigurationDetails{
		VerifyDepth:                    details.VerifyDepth,
		VerifyPeerCertificate:          details.VerifyPeerCertificate,
		TrustedCertificateAuthorityIds: details.TrustedCertificateAuthorityIds,
		CertificateIds:                 details.CertificateIds,
		CertificateName:                details.CertificateName,
		ServerOrderPreference:          string(details.ServerOrderPreference),
		CipherSuiteName:                details.CipherSuiteName,
		Protocols:                      details.Protocols,
	}
}

func getSessionPersistenceConfiguration(details *GenericSessionPersistenceConfiguration) *loadbalancer.SessionPersistenceConfigurationDetails {
	if details == nil {
		return nil
	}
	return &loadbalancer.SessionPersistenceConfigurationDetails{
		CookieName:      details.CookieName,
		DisableFallback: details.DisableFallback,
	}
}

func getGenericSessionPersistenceConfiguration(details *loadbalancer.SessionPersistenceConfigurationDetails) *GenericSessionPersistenceConfiguration {
	if details == nil {
		return nil
	}

	return &GenericSessionPersistenceConfiguration{
		CookieName:      details.CookieName,
		DisableFallback: details.DisableFallback,
	}
}

func getListenerConnectionConfiguration(details *GenericConnectionConfiguration) *loadbalancer.ConnectionConfiguration {
	var connectionConfiguration *loadbalancer.ConnectionConfiguration

	if details == nil {
		connectionConfiguration = nil
	} else {
		connectionConfiguration = &loadbalancer.ConnectionConfiguration{
			IdleTimeout:                    details.IdleTimeout,
			BackendTcpProxyProtocolVersion: details.BackendTcpProxyProtocolVersion,
		}
	}
	return connectionConfiguration
}
