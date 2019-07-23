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
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/filestorage"
	"github.com/oracle/oci-go-sdk/identity"
	"github.com/oracle/oci-go-sdk/loadbalancer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"
)

// Interface of consumed OCI API functionality.
type Interface interface {
	Compute() ComputeInterface
	LoadBalancer() LoadBalancerInterface
	Networking() NetworkingInterface
	BlockStorage() BlockStorageInterface
	FSS() FileStorageInterface
	Identity() IdentityInterface
}

// RateLimiter reader and writer.
type RateLimiter struct {
	Reader flowcontrol.RateLimiter
	Writer flowcontrol.RateLimiter
}

type computeClient interface {
	GetInstance(ctx context.Context, request core.GetInstanceRequest) (response core.GetInstanceResponse, err error)
	ListInstances(ctx context.Context, request core.ListInstancesRequest) (response core.ListInstancesResponse, err error)
	ListVnicAttachments(ctx context.Context, request core.ListVnicAttachmentsRequest) (response core.ListVnicAttachmentsResponse, err error)

	GetVolumeAttachment(ctx context.Context, request core.GetVolumeAttachmentRequest) (response core.GetVolumeAttachmentResponse, err error)
	ListVolumeAttachments(ctx context.Context, request core.ListVolumeAttachmentsRequest) (response core.ListVolumeAttachmentsResponse, err error)
	AttachVolume(ctx context.Context, request core.AttachVolumeRequest) (response core.AttachVolumeResponse, err error)
	DetachVolume(ctx context.Context, request core.DetachVolumeRequest) (response core.DetachVolumeResponse, err error)
}

type virtualNetworkClient interface {
	GetVnic(ctx context.Context, request core.GetVnicRequest) (response core.GetVnicResponse, err error)
	GetSubnet(ctx context.Context, request core.GetSubnetRequest) (response core.GetSubnetResponse, err error)
	GetVcn(ctx context.Context, request core.GetVcnRequest) (response core.GetVcnResponse, err error)
	GetSecurityList(ctx context.Context, request core.GetSecurityListRequest) (response core.GetSecurityListResponse, err error)
	UpdateSecurityList(ctx context.Context, request core.UpdateSecurityListRequest) (response core.UpdateSecurityListResponse, err error)

	GetPrivateIp(ctx context.Context, request core.GetPrivateIpRequest) (response core.GetPrivateIpResponse, err error)
}

type loadBalancerClient interface {
	GetLoadBalancer(ctx context.Context, request loadbalancer.GetLoadBalancerRequest) (response loadbalancer.GetLoadBalancerResponse, err error)
	ListLoadBalancers(ctx context.Context, request loadbalancer.ListLoadBalancersRequest) (response loadbalancer.ListLoadBalancersResponse, err error)
	CreateLoadBalancer(ctx context.Context, request loadbalancer.CreateLoadBalancerRequest) (response loadbalancer.CreateLoadBalancerResponse, err error)
	DeleteLoadBalancer(ctx context.Context, request loadbalancer.DeleteLoadBalancerRequest) (response loadbalancer.DeleteLoadBalancerResponse, err error)
	ListCertificates(ctx context.Context, request loadbalancer.ListCertificatesRequest) (response loadbalancer.ListCertificatesResponse, err error)
	CreateCertificate(ctx context.Context, request loadbalancer.CreateCertificateRequest) (response loadbalancer.CreateCertificateResponse, err error)
	GetWorkRequest(ctx context.Context, request loadbalancer.GetWorkRequestRequest) (response loadbalancer.GetWorkRequestResponse, err error)
	CreateBackendSet(ctx context.Context, request loadbalancer.CreateBackendSetRequest) (response loadbalancer.CreateBackendSetResponse, err error)
	UpdateBackendSet(ctx context.Context, request loadbalancer.UpdateBackendSetRequest) (response loadbalancer.UpdateBackendSetResponse, err error)
	DeleteBackendSet(ctx context.Context, request loadbalancer.DeleteBackendSetRequest) (response loadbalancer.DeleteBackendSetResponse, err error)
	CreateBackend(ctx context.Context, request loadbalancer.CreateBackendRequest) (response loadbalancer.CreateBackendResponse, err error)
	DeleteBackend(ctx context.Context, request loadbalancer.DeleteBackendRequest) (response loadbalancer.DeleteBackendResponse, err error)
	CreateListener(ctx context.Context, request loadbalancer.CreateListenerRequest) (response loadbalancer.CreateListenerResponse, err error)
	UpdateListener(ctx context.Context, request loadbalancer.UpdateListenerRequest) (response loadbalancer.UpdateListenerResponse, err error)
	DeleteListener(ctx context.Context, request loadbalancer.DeleteListenerRequest) (response loadbalancer.DeleteListenerResponse, err error)
}

type filestorageClient interface {
	CreateFileSystem(ctx context.Context, request filestorage.CreateFileSystemRequest) (response filestorage.CreateFileSystemResponse, err error)
	GetFileSystem(ctx context.Context, request filestorage.GetFileSystemRequest) (response filestorage.GetFileSystemResponse, err error)
	ListFileSystems(ctx context.Context, request filestorage.ListFileSystemsRequest) (response filestorage.ListFileSystemsResponse, err error)
	DeleteFileSystem(ctx context.Context, request filestorage.DeleteFileSystemRequest) (response filestorage.DeleteFileSystemResponse, err error)

	CreateExport(ctx context.Context, request filestorage.CreateExportRequest) (response filestorage.CreateExportResponse, err error)
	ListExports(ctx context.Context, request filestorage.ListExportsRequest) (response filestorage.ListExportsResponse, err error)
	GetExport(ctx context.Context, request filestorage.GetExportRequest) (response filestorage.GetExportResponse, err error)
	DeleteExport(ctx context.Context, request filestorage.DeleteExportRequest) (response filestorage.DeleteExportResponse, err error)

	GetMountTarget(ctx context.Context, request filestorage.GetMountTargetRequest) (response filestorage.GetMountTargetResponse, err error)
}

type blockstorageClient interface {
	GetVolume(ctx context.Context, request core.GetVolumeRequest) (response core.GetVolumeResponse, err error)
	CreateVolume(ctx context.Context, request core.CreateVolumeRequest) (response core.CreateVolumeResponse, err error)
	DeleteVolume(ctx context.Context, request core.DeleteVolumeRequest) (response core.DeleteVolumeResponse, err error)
	ListVolumes(ctx context.Context, request core.ListVolumesRequest) (response core.ListVolumesResponse, err error)
}

type identityClient interface {
	ListAvailabilityDomains(ctx context.Context, request identity.ListAvailabilityDomainsRequest) (identity.ListAvailabilityDomainsResponse, error)
}

type client struct {
	compute      computeClient
	network      virtualNetworkClient
	loadbalancer loadBalancerClient
	filestorage  filestorageClient
	bs           blockstorageClient
	identity     identityClient

	rateLimiter RateLimiter

	subnetCache cache.Store
	logger      *zap.SugaredLogger
}

// New constructs an OCI API client.
func New(logger *zap.SugaredLogger, cp common.ConfigurationProvider, opRateLimiter *RateLimiter) (Interface, error) {
	compute, err := core.NewComputeClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewComputeClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &compute.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring load balancer client custom transport")
	}

	network, err := core.NewVirtualNetworkClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewVirtualNetworkClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &network.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring load balancer client custom transport")
	}

	lb, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewLoadBalancerClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &lb.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring loadbalancer client custom transport")
	}

	identity, err := identity.NewIdentityClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewIdentityClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &identity.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring identity service client custom transport")
	}

	bs, err := core.NewBlockstorageClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewBlockstorageClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &bs.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring block storage service client custom transport")
	}

	fss, err := filestorage.NewFileStorageClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewFileStorageClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &fss.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring file storage service client custom transport")
	}

	c := &client{
		compute:      &compute,
		network:      &network,
		identity:     &identity,
		loadbalancer: &lb,
		bs:           &bs,
		filestorage:  &fss,

		rateLimiter: *opRateLimiter,

		subnetCache: cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
		logger:      logger,
	}

	return c, nil
}

func (c *client) LoadBalancer() LoadBalancerInterface {
	return c
}

func (c *client) Networking() NetworkingInterface {
	return c
}

func (c *client) Compute() ComputeInterface {
	return c
}

func (c *client) Identity() IdentityInterface {
	return c
}

func (c *client) BlockStorage() BlockStorageInterface {
	return c
}

func (c *client) FSS() FileStorageInterface {
	return c
}

func configureCustomTransport(logger *zap.SugaredLogger, baseClient *common.BaseClient) error {
	httpClient := baseClient.HTTPClient.(*http.Client)

	var transport *http.Transport
	if httpClient.Transport == nil {
		transport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	} else {
		transport = httpClient.Transport.(*http.Transport)
	}

	ociProxy := os.Getenv("OCI_PROXY")
	if ociProxy != "" {
		proxyURL, err := url.Parse(ociProxy)
		if err != nil {
			return errors.Wrapf(err, "failed to parse OCI proxy url: %s", ociProxy)
		}
		transport.Proxy = func(req *http.Request) (*url.URL, error) {
			return proxyURL, nil
		}
	}

	trustedCACertPath := os.Getenv("TRUSTED_CA_CERT_PATH")
	if trustedCACertPath != "" {
		logger.With("path", trustedCACertPath).Infof("Configuring OCI client with a new trusted ca")
		trustedCACert, err := ioutil.ReadFile(trustedCACertPath)
		if err != nil {
			return errors.Wrapf(err, "failed to read root certificate: %s", trustedCACertPath)
		}
		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(trustedCACert)
		if !ok {
			return errors.Wrapf(err, "failed to parse root certificate: %s", trustedCACertPath)
		}
		transport.TLSClientConfig = &tls.Config{RootCAs: caCertPool}
	}

	httpClient.Transport = transport
	return nil
}
