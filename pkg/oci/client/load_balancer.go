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

	"github.com/oracle/oci-go-sdk/loadbalancer"
	"github.com/pkg/errors"
)

const workRequestPollInterval = 5 * time.Second

// LoadBalancerInterface for consumed LB functionality.
type LoadBalancerInterface interface {
	CreateLoadBalancer(ctx context.Context, details loadbalancer.CreateLoadBalancerDetails) (string, error)

	GetLoadBalancer(ctx context.Context, id string) (*loadbalancer.LoadBalancer, error)
	GetLoadBalancerByName(ctx context.Context, compartmentID, name string) (*loadbalancer.LoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, id string) (string, error)

	GetCertificateByName(ctx context.Context, lbID, name string) (*loadbalancer.Certificate, error)
	CreateCertificate(ctx context.Context, lbID string, cert loadbalancer.CertificateDetails) (string, error)

	CreateBackendSet(ctx context.Context, lbID, name string, details loadbalancer.BackendSetDetails) (string, error)
	UpdateBackendSet(ctx context.Context, lbID, name string, details loadbalancer.BackendSetDetails) (string, error)
	DeleteBackendSet(ctx context.Context, lbID, name string) (string, error)

	CreateBackend(ctx context.Context, lbID, bsName string, details loadbalancer.BackendDetails) (string, error)
	DeleteBackend(ctx context.Context, lbID, bsName, name string) (string, error)

	UpdateListener(ctx context.Context, lbID, name string, details loadbalancer.ListenerDetails) (string, error)
	CreateListener(ctx context.Context, lbID, name string, details loadbalancer.ListenerDetails) (string, error)
	DeleteListener(ctx context.Context, lbID, name string) (string, error)

	AwaitWorkRequest(ctx context.Context, id string) (*loadbalancer.WorkRequest, error)
}

func (c *client) GetLoadBalancer(ctx context.Context, id string) (*loadbalancer.LoadBalancer, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetLoadBalancer")
	}

	resp, err := c.loadbalancer.GetLoadBalancer(ctx, loadbalancer.GetLoadBalancerRequest{
		LoadBalancerId: &id,
	})
	incRequestCounter(err, getVerb, loadBalancerResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.LoadBalancer, nil
}

func (c *client) GetLoadBalancerByName(ctx context.Context, compartmentID, name string) (*loadbalancer.LoadBalancer, error) {
	var page *string
	for {
		if !c.rateLimiter.Reader.TryAccept() {
			return nil, RateLimitError(false, "ListLoadBalancers")
		}
		resp, err := c.loadbalancer.ListLoadBalancers(ctx, loadbalancer.ListLoadBalancersRequest{
			CompartmentId: &compartmentID,
			DisplayName:   &name,
			Page:          page,
		})
		incRequestCounter(err, listVerb, loadBalancerResource)

		if err != nil {
			return nil, errors.WithStack(err)
		}
		for _, lb := range resp.Items {
			if *lb.DisplayName == name {
				return &lb, nil
			}
		}
		if page = resp.OpcNextPage; page == nil {
			break
		}
	}

	return nil, errors.WithStack(errNotFound)
}

func (c *client) CreateLoadBalancer(ctx context.Context, details loadbalancer.CreateLoadBalancerDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateLoadBalancer")
	}

	resp, err := c.loadbalancer.CreateLoadBalancer(ctx, loadbalancer.CreateLoadBalancerRequest{
		CreateLoadBalancerDetails: details,
	})
	incRequestCounter(err, createVerb, loadBalancerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "DeleteLoadBalancer")
	}

	resp, err := c.loadbalancer.DeleteLoadBalancer(ctx, loadbalancer.DeleteLoadBalancerRequest{
		LoadBalancerId: &id,
	})
	incRequestCounter(err, deleteVerb, loadBalancerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) GetCertificateByName(ctx context.Context, lbID, name string) (*loadbalancer.Certificate, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "ListCertificates")
	}

	resp, err := c.loadbalancer.ListCertificates(ctx, loadbalancer.ListCertificatesRequest{
		LoadBalancerId: &lbID,
	})
	incRequestCounter(err, listVerb, certificateResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	for _, cert := range resp.Items {
		if *cert.CertificateName == name {
			return &cert, nil
		}
	}
	return nil, errors.WithStack(errNotFound)
}

func (c *client) CreateCertificate(ctx context.Context, lbID string, cert loadbalancer.CertificateDetails) (string, error) {
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
	})
	incRequestCounter(err, createVerb, certificateResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) GetWorkRequest(ctx context.Context, id string) (*loadbalancer.WorkRequest, error) {
	if !c.rateLimiter.Reader.TryAccept() {
		return nil, RateLimitError(false, "GetWorkRequest")
	}

	resp, err := c.loadbalancer.GetWorkRequest(ctx, loadbalancer.GetWorkRequestRequest{
		WorkRequestId: &id,
	})
	incRequestCounter(err, getVerb, workRequestResource)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &resp.WorkRequest, nil
}

func (c *client) CreateBackendSet(ctx context.Context, lbID, name string, details loadbalancer.BackendSetDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateBackendSet")
	}

	resp, err := c.loadbalancer.CreateBackendSet(ctx, loadbalancer.CreateBackendSetRequest{
		LoadBalancerId: &lbID,
		CreateBackendSetDetails: loadbalancer.CreateBackendSetDetails{
			Name:                            &name,
			Backends:                        details.Backends,
			HealthChecker:                   details.HealthChecker,
			Policy:                          details.Policy,
			SessionPersistenceConfiguration: details.SessionPersistenceConfiguration,
			SslConfiguration:                details.SslConfiguration,
		},
	})
	incRequestCounter(err, createVerb, backendSetResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) UpdateBackendSet(ctx context.Context, lbID, name string, details loadbalancer.BackendSetDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "UpdateBackendSet")
	}

	resp, err := c.loadbalancer.UpdateBackendSet(ctx, loadbalancer.UpdateBackendSetRequest{
		LoadBalancerId: &lbID,
		BackendSetName: &name,
		UpdateBackendSetDetails: loadbalancer.UpdateBackendSetDetails{
			Backends:                        details.Backends,
			HealthChecker:                   details.HealthChecker,
			Policy:                          details.Policy,
			SessionPersistenceConfiguration: details.SessionPersistenceConfiguration,
			SslConfiguration:                details.SslConfiguration,
		},
	})
	incRequestCounter(err, updateVerb, backendSetResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) DeleteBackendSet(ctx context.Context, lbID, name string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "DeleteBackendSet")
	}

	resp, err := c.loadbalancer.DeleteBackendSet(ctx, loadbalancer.DeleteBackendSetRequest{
		LoadBalancerId: &lbID,
		BackendSetName: &name,
	})
	incRequestCounter(err, deleteVerb, backendSetResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) CreateBackend(ctx context.Context, lbID, bsName string, details loadbalancer.BackendDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateBackend")
	}

	resp, err := c.loadbalancer.CreateBackend(ctx, loadbalancer.CreateBackendRequest{
		LoadBalancerId: &lbID,
		BackendSetName: &bsName,
		CreateBackendDetails: loadbalancer.CreateBackendDetails{
			IpAddress: details.IpAddress,
			Port:      details.Port,
		},
	})
	incRequestCounter(err, createVerb, backendResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) DeleteBackend(ctx context.Context, lbID, bsName, name string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "DeleteBackend")
	}

	resp, err := c.loadbalancer.DeleteBackend(ctx, loadbalancer.DeleteBackendRequest{
		LoadBalancerId: &lbID,
		BackendSetName: &bsName,
		BackendName:    &name,
	})
	incRequestCounter(err, deleteVerb, backendResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) CreateListener(ctx context.Context, lbID, name string, details loadbalancer.ListenerDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "CreateListener")
	}

	resp, err := c.loadbalancer.CreateListener(ctx, loadbalancer.CreateListenerRequest{
		LoadBalancerId: &lbID,
		CreateListenerDetails: loadbalancer.CreateListenerDetails{
			Name:                    &name,
			DefaultBackendSetName:   details.DefaultBackendSetName,
			Port:                    details.Port,
			Protocol:                details.Protocol,
			SslConfiguration:        details.SslConfiguration,
			ConnectionConfiguration: details.ConnectionConfiguration,
		},
	})
	incRequestCounter(err, createVerb, listenerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) UpdateListener(ctx context.Context, lbID, name string, details loadbalancer.ListenerDetails) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "UpdateListener")
	}

	resp, err := c.loadbalancer.UpdateListener(ctx, loadbalancer.UpdateListenerRequest{
		LoadBalancerId: &lbID,
		ListenerName:   &name,
		UpdateListenerDetails: loadbalancer.UpdateListenerDetails{
			DefaultBackendSetName:   details.DefaultBackendSetName,
			Port:                    details.Port,
			Protocol:                details.Protocol,
			SslConfiguration:        details.SslConfiguration,
			ConnectionConfiguration: details.ConnectionConfiguration,
		},
	})
	incRequestCounter(err, updateVerb, listenerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}

func (c *client) AwaitWorkRequest(ctx context.Context, id string) (*loadbalancer.WorkRequest, error) {
	var wr *loadbalancer.WorkRequest
	err := wait.PollUntil(workRequestPollInterval, func() (done bool, err error) {
		twr, err := c.GetWorkRequest(ctx, id)
		if err != nil {
			if IsRetryable(err) {
				return false, nil
			}
			return true, errors.WithStack(err)
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
	return wr, err
}

func (c *client) DeleteListener(ctx context.Context, lbID, name string) (string, error) {
	if !c.rateLimiter.Writer.TryAccept() {
		return "", RateLimitError(true, "DeleteListener")
	}

	resp, err := c.loadbalancer.DeleteListener(ctx, loadbalancer.DeleteListenerRequest{
		LoadBalancerId: &lbID,
		ListenerName:   &name,
	})
	incRequestCounter(err, deleteVerb, listenerResource)

	if err != nil {
		return "", errors.WithStack(err)
	}

	return *resp.OpcWorkRequestId, nil
}
