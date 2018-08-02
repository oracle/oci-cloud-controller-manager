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

	"github.com/golang/glog"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	"github.com/pkg/errors"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	wait "k8s.io/apimachinery/pkg/util/wait"
	informers "k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	cache "k8s.io/client-go/tools/cache"
	cloudprovider "k8s.io/kubernetes/pkg/cloudprovider"
	controller "k8s.io/kubernetes/pkg/controller"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instancemeta"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/util"
)

// ProviderName uniquely identifies the Oracle Bare Metal Cloud Services (OCI)
// cloud-provider.
func ProviderName() string {
	return util.ProviderName
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
	config                     *Config
}

// Compile time check that CloudProvider implements the cloudprovider.Interface
// interface.
var _ cloudprovider.Interface = &CloudProvider{}

// NewCloudProvider creates a new oci.CloudProvider.
func NewCloudProvider(config *Config) (cloudprovider.Interface, error) {
	cp, err := buildConfigurationProvider(config)
	if err != nil {
		return nil, err
	}
	c, err := client.New(cp)
	if err != nil {
		return nil, err
	}

	if config.CompartmentID == "" {
		glog.Info("Compartment not supplied in config: attempting to infer from instance metadata")
		metadata, err := instancemeta.New().Get()
		if err != nil {
			return nil, err
		}
		config.CompartmentID = metadata.CompartmentOCID
	}

	if !config.LoadBalancer.Disabled && config.VCNID == "" {
		glog.Infof("No vcn provided in cloud provider config. Falling back to looking up VCN via LB subnet.")
		subnet, err := c.Networking().GetSubnet(context.Background(), config.LoadBalancer.Subnet1)
		if err != nil {
			return nil, errors.Wrap(err, "get subnet for loadBalancer.subnet1")
		}
		config.VCNID = *subnet.VcnId
	}

	return &CloudProvider{
		client: c,
		config: config,
	}, nil
}

func init() {
	cloudprovider.RegisterCloudProvider(ProviderName(), func(config io.Reader) (cloudprovider.Interface, error) {
		cfg, err := ReadConfig(config)
		if err != nil {
			return nil, err
		}
		cfg.Complete()

		if err = cfg.Validate(); err != nil {
			return nil, err
		}

		return NewCloudProvider(cfg)
	})
}

// Initialize passes a Kubernetes clientBuilder interface to the cloud provider.
func (cp *CloudProvider) Initialize(clientBuilder controller.ControllerClientBuilder) {
	var err error
	cp.kubeclient, err = clientBuilder.Client("cloud-controller-manager")
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to create kubeclient: %v", err))
	}

	factory := informers.NewSharedInformerFactory(cp.kubeclient, 5*time.Minute)

	nodeInformer := factory.Core().V1().Nodes()
	go nodeInformer.Informer().Run(wait.NeverStop)
	serviceInformer := factory.Core().V1().Services()
	go serviceInformer.Informer().Run(wait.NeverStop)

	glog.Info("Waiting for node informer cache to sync")
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
		return newSecurityListManager(cp.client, serviceInformer, cp.config.LoadBalancer.SecurityLists, mode)
	}
}

// ProviderName returns the cloud-provider ID.
func (cp *CloudProvider) ProviderName() string {
	return ProviderName()
}

// LoadBalancer returns a balancer interface. Also returns true if the interface
// is supported, false otherwise.
func (cp *CloudProvider) LoadBalancer() (cloudprovider.LoadBalancer, bool) {
	glog.V(6).Info("Claiming to support Load Balancers")
	return cp, !cp.config.LoadBalancer.Disabled
}

// Instances returns an instances interface. Also returns true if the interface
// is supported, false otherwise.
func (cp *CloudProvider) Instances() (cloudprovider.Instances, bool) {
	glog.V(6).Info("Claiming to support instances")
	return cp, true
}

// Zones returns a zones interface. Also returns true if the interface is
// supported, false otherwise.
func (cp *CloudProvider) Zones() (cloudprovider.Zones, bool) {
	glog.V(6).Info("Claiming to support Zones")
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

func buildConfigurationProvider(config *Config) (common.ConfigurationProvider, error) {
	if config.Auth.UseInstancePrincipals {
		glog.V(2).Info("Using instance principals configuration provider")
		cp, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, errors.Wrap(err, "InstancePrincipalConfigurationProvider")
		}
		return cp, nil
	}
	glog.V(2).Info("Using raw configuration provider")
	cp := common.NewRawConfigurationProvider(
		config.Auth.TenancyID,
		config.Auth.UserID,
		config.Auth.Region,
		config.Auth.Fingerprint,
		config.Auth.PrivateKey,
		&config.Auth.Passphrase,
	)
	return cp, nil
}
