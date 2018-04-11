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

package oci

import (
	"io"
	"io/ioutil"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// AuthConfig holds the configuration required for communicating with the OCI
// API.
type AuthConfig struct {
	UseInstancePrincipals bool   `yaml:"useInstancePrincipals"`
	Region                string `yaml:"region"`
	TenancyID             string `yaml:"tenancy"`
	// CompartmentID is DEPRECIATED and should be set on the top level Config
	// struct.
	CompartmentID string `yaml:"compartment"`
	UserID        string `yaml:"user"`
	PrivateKey    string `yaml:"key"`
	Fingerprint   string `yaml:"fingerprint"`
	// PrivateKeyPassphrase is DEPRECIATED in favour of Passphrase.
	PrivateKeyPassphrase string `yaml:"key_passphrase"`
	Passphrase           string `yaml:"passphrase"`
}

// LoadBalancerConfig holds the configuration options for OCI load balancers.
type LoadBalancerConfig struct {
	// DisableSecurityListManagement disables the automatic creation of ingress
	// rules for the node subnets and egress rules for the load balancers to the node subnets.
	//
	// If security list management is disabled, then it requires that the user
	// has setup a rule that allows inbound traffic to the appropriate ports
	// for kube proxy health port, node port ranges, and health check port ranges.
	// E.g. 10.82.0.0/16 30000-32000
	DisableSecurityListManagement bool `yaml:"disableSecurityListManagement"`

	// SecurityListManagementMode defines how the CCM manages security lists
	// when provisioning load balancers. Available modes are All, Frontend,
	// and None.
	SecurityListManagementMode string `yaml:"securityListManagementMode"`

	Subnet1 string `yaml:"subnet1"`
	Subnet2 string `yaml:"subnet2"`

	// SecurityLists defines the Security List to mutate for each Subnet (
	// both load balancer and worker).
	SecurityLists map[string]string `yaml:"securityLists"`
}

// Config holds the OCI cloud-provider config passed to Kubernetes compontents
// via the --cloud-config option.
type Config struct {
	Auth         AuthConfig         `yaml:"auth"`
	LoadBalancer LoadBalancerConfig `yaml:"loadBalancer"`

	// CompartmentID is the OCID of the Compartment within which the cluster
	// resides.
	CompartmentID string `yaml:"compartment"`
	// VCNID is the OCID of the Virtual Cloud Network (VCN) within which the
	// cluster resides.
	VCNID string `yaml:"vcn"`
}

// Complete the config applying defaults / overrides.
func (c *Config) Complete() {
	if c.LoadBalancer.SecurityListManagementMode == "" {
		c.LoadBalancer.SecurityListManagementMode = ManagementModeAll // default
		if c.LoadBalancer.DisableSecurityListManagement {
			glog.Warningf("cloud-provider config: \"loadBalancer.disableSecurityListManagement\" is DEPRECIATED and will be removed in a later release. Please set \"loadBalancer.SecurityListManagementMode: %s\".", ManagementModeNone)
			c.LoadBalancer.SecurityListManagementMode = ManagementModeNone
		}
	}
	if c.CompartmentID == "" && c.Auth.CompartmentID != "" {
		glog.Warning("cloud-provider config: \"auth.compartment\" is DEPRECIATED and will be removed in a later release. Please set \"compartment\".")
		c.CompartmentID = c.Auth.CompartmentID
	}
	if c.Auth.Passphrase == "" && c.Auth.PrivateKeyPassphrase != "" {
		glog.Warning("cloud-provider config: \"auth.key_passphrase\" is DEPRECIATED and will be removed in a later release. Please set \"auth.passphrase\".")
		c.Auth.Passphrase = c.Auth.PrivateKeyPassphrase
	}
}

// Validate validates the OCI cloud-provider config.
func (c *Config) Validate() error {
	return ValidateConfig(c).ToAggregate()
}

// ReadConfig consumes the config Reader and constructs a Config object.
func ReadConfig(r io.Reader) (*Config, error) {
	if r == nil {
		return nil, errors.New("no cloud-provider config file given")
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "reading cloud-provider config")
	}

	cfg := &Config{}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling cloud-provider config")
	}

	return cfg, nil
}
