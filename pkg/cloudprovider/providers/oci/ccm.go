// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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

// Package oci implements an external Kubernetes cloud-provider for Oracle Cloud
// Infrastructure.
package oci

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	cloudprovider "k8s.io/cloud-provider"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instance/metadata"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/core"
)

const (
	// providerName uniquely identifies the Oracle Cloud Infrastructure
	// (OCI) cloud-provider.
	providerName   = "oci"
	providerPrefix = providerName + "://"
)

// ProviderName uniquely identifies the Oracle Bare Metal Cloud Services (OCI)
// cloud-provider.
func ProviderName() string {
	return providerName
}

// CloudProvider is an implementation of the cloud-provider interface for OCI.
type CloudProvider struct {
	// NodeLister provides a cache to lookup nodes for deleting a load balancer.
	// Due to limitations in the OCI API around going from an IP to a subnet
	// we use the node lister to go from IP -> node / provider id -> ... -> subnet
	NodeLister listersv1.NodeLister

	client     client.Interface
	kubeclient clientset.Interface

	securityListManagerFactory securityListManagerFactory
	config                     *providercfg.Config

	logger        *zap.SugaredLogger
	instanceCache cache.Store
	metricPusher  *metrics.MetricPusher
}

func (cp *CloudProvider) InstancesV2() (cloudprovider.InstancesV2, bool) {
	cp.logger.Debug("Claiming to not support instancesV2")
	return nil, false
}

// Compile time check that CloudProvider implements the cloudprovider.Interface
// interface.
var _ cloudprovider.Interface = &CloudProvider{}

// NewCloudProvider creates a new oci.CloudProvider.
func NewCloudProvider(config *providercfg.Config) (cloudprovider.Interface, error) {
	// The global logger has been replaced with the logger we constructed in
	// cloud_provider_oci.go so capture it here and then pass it into all components.
	logger := zap.L()
	logger = logger.With(zap.String("component", "cloud-controller-manager"))

	cp, err := providercfg.NewConfigurationProvider(config)
	if err != nil {
		return nil, err
	}

	rateLimiter := client.NewRateLimiter(logger.Sugar(), config.RateLimiter)

	c, err := client.New(logger.Sugar(), cp, &rateLimiter)
	if err != nil {
		return nil, err
	}

	if config.CompartmentID == "" {
		logger.Info("Compartment not supplied in config: attempting to infer from instance metadata")
		metadata, err := metadata.New().Get()
		if err != nil {
			return nil, err
		}
		config.CompartmentID = metadata.CompartmentID
	}

	if !config.LoadBalancer.Disabled && config.VCNID == "" {
		logger.Info("No VCN provided in cloud provider config. Falling back to looking up VCN via LB subnet.")
		subnet, err := c.Networking().GetSubnet(context.Background(), config.LoadBalancer.Subnet1)
		if err != nil {
			return nil, errors.Wrap(err, "get subnet for loadBalancer.subnet1")
		}
		config.VCNID = *subnet.VcnId
	}

	metricPusher, err := metrics.NewMetricPusher(logger.Sugar())
	if err != nil {
		logger.Sugar().With("error", err).Error("Metrics collection could not be enabled")
		// disable metrics
		metricPusher = nil
	}

	if metricPusher != nil {
		logger.Info("Metrics collection has been enabled")
	} else {
		logger.Info("Metrics collection has not been enabled")
	}

	return &CloudProvider{
		client:        c,
		config:        config,
		logger:        logger.Sugar(),
		instanceCache: cache.NewTTLStore(instanceCacheKeyFn, time.Duration(24)*time.Hour),
		metricPusher:  metricPusher,
	}, nil
}

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName(), func(config io.Reader) (cloudprovider.Interface, error) {
		cfg, err := providercfg.ReadConfig(config)
		if err != nil {
			return nil, err
		}

		if err = cfg.Validate(); err != nil {
			return nil, err
		}

		return NewCloudProvider(cfg)
	})
	common.EnableInstanceMetadataServiceLookup()
}

// Initialize passes a Kubernetes clientBuilder interface to the cloud provider.
func (cp *CloudProvider) Initialize(clientBuilder cloudprovider.ControllerClientBuilder, stop <-chan struct{}) {
	var err error
	cp.kubeclient, err = clientBuilder.Client("cloud-controller-manager")
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to create kubeclient: %v", err))
	}

	factory := informers.NewSharedInformerFactory(cp.kubeclient, 5*time.Minute)

	nodeInfoController := NewNodeInfoController(
		factory.Core().V1().Nodes(),
		cp.kubeclient,
		cp,
		cp.logger,
		cp.instanceCache,
		cp.client)

	nodeInformer := factory.Core().V1().Nodes()
	go nodeInformer.Informer().Run(wait.NeverStop)
	serviceInformer := factory.Core().V1().Services()
	go serviceInformer.Informer().Run(wait.NeverStop)
	go nodeInfoController.Run(wait.NeverStop)

	cp.logger.Info("Waiting for node informer cache to sync")
	if !cache.WaitForCacheSync(wait.NeverStop, nodeInformer.Informer().HasSynced, serviceInformer.Informer().HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for informers to sync"))
	}
	cp.NodeLister = nodeInformer.Lister()

	cp.securityListManagerFactory = func(mode string) securityListManager {
		if cp.config.LoadBalancer.Disabled {
			return newSecurityListManagerNOOP()
		}
		if len(mode) == 0 {
			mode = cp.config.LoadBalancer.SecurityListManagementMode
		}
		return newSecurityListManager(cp.logger, cp.client, serviceInformer, cp.config.LoadBalancer.SecurityLists, mode)
	}
}

// ProviderName returns the cloud-provider ID.
func (cp *CloudProvider) ProviderName() string {
	return ProviderName()
}

// LoadBalancer returns a balancer interface. Also returns true if the interface
// is supported, false otherwise.
func (cp *CloudProvider) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	cp.logger.Debug("Claiming to support load balancers")
	return cp, !cp.config.LoadBalancer.Disabled
}

// Instances returns an instances interface. Also returns true if the interface
// is supported, false otherwise.
func (cp *CloudProvider) Instances() (cloudprovider.Instances, bool) {
	cp.logger.Debug("Claiming to support instances")
	return cp, true
}

// Zones returns a zones interface. Also returns true if the interface is
// supported, false otherwise.
func (cp *CloudProvider) Zones() (cloudprovider.Zones, bool) {
	cp.logger.Debug("Claiming to support zones")
	return cp, true
}

// Clusters returns a clusters interface.  Also returns true if the interface is
// supported, false otherwise.
func (cp *CloudProvider) Clusters() (cloudprovider.Clusters, bool) {
	return nil, false
}

// Routes returns a routes interface along with whether the interface is
// supported.
func (cp *CloudProvider) Routes() (cloudprovider.Routes, bool) {
	return nil, false
}

// ScrubDNS provides an opportunity for cloud-provider-specific code to process
// DNS settings for pods.
func (cp *CloudProvider) ScrubDNS(nameservers, searches []string) (nsOut, srchOut []string) {
	return nameservers, searches
}

// HasClusterID returns true if the cluster has a clusterID.
func (cp *CloudProvider) HasClusterID() bool {
	return true
}

func instanceCacheKeyFn(obj interface{}) (string, error) {
	return *obj.(*core.Instance).Id, nil
}
