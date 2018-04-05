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

	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/loadbalancer"
	"github.com/pkg/errors"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instancemeta"
)

// Interface of consumed OCI API functionality.
type Interface interface {
	Compute() ComputeInterface
	LoadBalancer() LoadBalancerInterface
	Networking() NetworkingInterface
}

type client struct {
	config *Config

	compute      *core.ComputeClient
	network      *core.VirtualNetworkClient
	loadbalancer *loadbalancer.LoadBalancerClient

	vcnID string

	subnetCache cache.Store
}

// New constructs an OCI API client.
func New(config *Config) (Interface, error) {
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

	err = configureCustomTransport(&compute.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring load balancer client custom transport")
	}

	network, err := core.NewVirtualNetworkClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewVirtualNetworkClientWithConfigurationProvider")
	}

	err = configureCustomTransport(&network.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring load balancer client custom transport")
	}

	lb, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewLoadBalancerClientWithConfigurationProvider")
	}

	err = configureCustomTransport(&lb.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring load balancer client custom transport")
	}

	c := &client{
		config: config,

		compute:      &compute,
		network:      &network,
		loadbalancer: &lb,

		subnetCache: cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
	}

	if config.Auth.CompartmentOCID == "" {
		glog.Info("Compartment not supplied in config: attempting to infer from instance metadata")
		metadata, err := instancemeta.New().Get()
		if err != nil {
			return nil, err
		}
		config.Auth.CompartmentOCID = metadata.CompartmentOCID
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

func (c *client) LoadBalancer() LoadBalancerInterface {
	return c
}

func (c *client) Networking() NetworkingInterface {
	return c
}

func (c *client) Compute() ComputeInterface {
	return c
}

func configureCustomTransport(baseClient *common.BaseClient) error {
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
		glog.Infof("configuring OCI client with a new trusted ca: %s", trustedCACertPath)
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
