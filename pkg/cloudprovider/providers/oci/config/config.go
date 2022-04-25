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

package config

import (
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instance/metadata"
	"io"
	"os"

	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/common/auth"
	"github.com/pkg/errors"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// AuthConfig holds the configuration required for communicating with the OCI
// API.
type AuthConfig struct {
	Region      string `yaml:"region"`
	TenancyID   string `yaml:"tenancy"`
	UserID      string `yaml:"user"`
	PrivateKey  string `yaml:"key"`
	Fingerprint string `yaml:"fingerprint"`
	Passphrase  string `yaml:"passphrase"`

	// Used by the flex driver for OCID expansion. This should be moved to top level
	// as it doesn't strictly relate to OCI authentication.
	RegionKey string `yaml:"regionKey"`

	// The fields below are deprecated and remain purely for backwards compatibility.
	// At some point these need to be removed.

	// UseInstancePrincipals is DEPRECATED should use top-level UseInstancePrincipals
	UseInstancePrincipals bool `yaml:"useInstancePrincipals"`
	// CompartmentID is DEPRECATED and should be set on the top level Config
	// struct.
	CompartmentID string `yaml:"compartment"`
	// PrivateKeyPassphrase is DEPRECATED in favour of Passphrase.
	PrivateKeyPassphrase string `yaml:"key_passphrase"`

	//Metadata service to help fill in certain fields
	metadataSvc metadata.Interface
}

const (
	// ManagementModeAll denotes the management of security list rules for load
	// balancer ingress/egress, health checkers, and worker ingress/egress.
	ManagementModeAll = "All"
	// ManagementModeFrontend denotes the management of security list rules for load
	// balancer ingress only.
	ManagementModeFrontend = "Frontend"
	// ManagementModeNone denotes the management of no security list rules.
	ManagementModeNone = "None"
)

// LoadBalancerConfig holds the configuration options for OCI load balancers.
type LoadBalancerConfig struct {
	// Disabled disables the creation of a load balancer.
	Disabled bool `yaml:"disabled"`

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

// RateLimiterConfig holds the configuration options for OCI rate limiting.
type RateLimiterConfig struct {
	RateLimitQPSRead     float32 `yaml:"rateLimitQPSRead"`
	RateLimitBucketRead  int     `yaml:"rateLimitBucketRead"` //Read?
	RateLimitQPSWrite    float32 `yaml:"rateLimitQPSWrite"`
	RateLimitBucketWrite int     `yaml:"rateLimitBucketWrite"`
	DisableRateLimiter   bool    `yaml:"disableRateLimiter"`
}

// MetricsConfig holds the configuration for collection metrics
// which are pushed to OCI Monitoring. More details present at
// https://docs.cloud.oracle.com/en-us/iaas/Content/Monitoring/Tasks/publishingcustommetrics.htm
type MetricsConfig struct {
	CompartmentID string `yaml:"compartmentID"`
	Namespace     string `yaml:"namespace"`
	ResourceGroup string `yaml:"resourceGroup"`
	// +optional
	// This prefix is added to all the metric names
	Prefix string `yaml:"prefix"`
}

// TagConfig hold the freeform and defined tags from the cluster level
// which should be added to the LB and BV provisioned by CCM
type TagConfig struct {
	FreeformTags map[string]string                 `yaml:"freeform"`
	DefinedTags  map[string]map[string]interface{} `yaml:"defined"`
}

// initialTags are optional tags to apply to all LBs and BVs provisioned in the cluster
type InitialTags struct {
	LoadBalancer *TagConfig `yaml:"loadBalancer"`
	BlockVolume  *TagConfig `yaml:"blockVolume"`
}

// Config holds the OCI cloud-provider config passed to Kubernetes components
// via the --cloud-config option.
type Config struct {
	Auth         AuthConfig          `yaml:"auth"`
	LoadBalancer *LoadBalancerConfig `yaml:"loadBalancer"`
	RateLimiter  *RateLimiterConfig  `yaml:"rateLimiter"`
	// Metrics collection is enabled when this configuration is provided
	Metrics *MetricsConfig `yaml:"metrics"`
	// Tags to be added to managed LB and BV
	Tags *InitialTags `yaml:"tags"`

	RegionKey string `yaml:"regionKey"`

	// When set to true, clients will use an instance principal configuration provider and ignore auth fields.
	UseInstancePrincipals bool `yaml:"useInstancePrincipals"`
	// CompartmentID is the OCID of the Compartment within which the cluster
	// resides.
	CompartmentID string `yaml:"compartment"`
	// VCNID is the OCID of the Virtual Cloud Network (VCN) within which the
	// cluster resides.
	VCNID string `yaml:"vcn"`

	//Metadata service to help fill in certain fields
	metadataSvc metadata.Interface
}

// Complete the load balancer config applying defaults / overrides.
func (c *LoadBalancerConfig) Complete() {
	if c.Disabled {
		return
	}
	if len(c.SecurityListManagementMode) == 0 {
		if c.DisableSecurityListManagement {
			zap.S().Warnf("cloud-provider config: \"loadBalancer.disableSecurityListManagement\" is DEPRECATED and will be removed in a later release. Please set \"loadBalancer.SecurityListManagementMode: %s\".", ManagementModeNone)
			c.SecurityListManagementMode = ManagementModeNone
		} else {
			c.SecurityListManagementMode = ManagementModeAll
		}
	}
}

// Complete the authentication config applying defaults / overrides.
func (c *AuthConfig) Complete() {
	if len(c.Passphrase) == 0 && len(c.PrivateKeyPassphrase) > 0 {
		zap.S().Warn("cloud-provider config: auth.key_passphrase is DEPRECIATED and will be removed in a later release. Please set auth.passphrase instead.")
		c.Passphrase = c.PrivateKeyPassphrase
	}
	if c.Region == "" || c.CompartmentID == "" {
		meta, err := c.metadataSvc.Get()
		if err != nil {
			zap.S().Warn("cloud-provider config: Unable to access metadata on instance. Will not be able to complete configuration if items are missing")
			return
		}
		if c.Region == "" {
			c.Region = meta.CanonicalRegionName
		}

		if c.CompartmentID == "" {
			c.CompartmentID = meta.CompartmentID
		}
	}
}

// Complete the top-level config applying defaults / overrides.
func (c *Config) Complete() {
	if c.LoadBalancer != nil {
		c.LoadBalancer.Complete()
	}
	c.Auth.Complete()
	// Ensure backwards compatibility fields are set correctly.
	if len(c.CompartmentID) == 0 && len(c.Auth.CompartmentID) > 0 {
		zap.S().Warn("cloud-provider config: \"auth.compartment\" is DEPRECATED and will be removed in a later release. Please set \"compartment\".")
		c.CompartmentID = c.Auth.CompartmentID
	}
	if c.Auth.UseInstancePrincipals {
		zap.S().Warn("cloud-provider config: \"auth.useInstancePrincipals\" is DEPRECATED and will be removed in a later release. Please set \"useInstancePrincipals\".")
		c.UseInstancePrincipals = true
	}

	if len(c.RegionKey) == 0 {
		if len(c.Auth.RegionKey) > 0 {
			zap.S().Warn("cloud-provider config: \"auth.RegionKey\" is DEPRECATED and will be removed in a later release. Please set \"RegionKey\".")
			c.RegionKey = c.Auth.RegionKey
		} else {
			meta, err := c.metadataSvc.Get()
			if err != nil {
				zap.S().Warn("cloud-provider config: Unable to access metadata on instance. Will not be able to complete configuration if items are missing")
				return
			}
			c.RegionKey = meta.Region
		}
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

	cfg := &Config{}
	err := yaml.NewDecoder(r).Decode(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling cloud-provider config")
	}

	cfg.metadataSvc = metadata.New()
	cfg.Auth.metadataSvc = cfg.metadataSvc
	// Ensure defaults are correctly set
	cfg.Complete()

	return cfg, nil
}

// FromFile will load a cloud provider configuration file from a given file path.
func FromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	defer f.Close()

	if err != nil {
		return nil, err
	}

	return ReadConfig(f)
}

// NewConfigurationProvider takes a cloud provider config file and returns an OCI ConfigurationProvider
// to be consumed by the OCI SDK.
func NewConfigurationProvider(cfg *Config) (common.ConfigurationProvider, error) {
	var conf common.ConfigurationProvider
	if cfg != nil {
		err := cfg.Validate()
		if err != nil {
			return nil, errors.Wrap(err, "invalid client config")
		}

		if cfg.UseInstancePrincipals {
			cp, err := auth.InstancePrincipalConfigurationProvider()
			if err != nil {
				return nil, errors.Wrap(err, "failed to instantiate InstancePrincipalConfigurationProvider")
			}
			return cp, nil
		}

		conf = common.NewRawConfigurationProvider(
			cfg.Auth.TenancyID,
			cfg.Auth.UserID,
			cfg.Auth.Region,
			cfg.Auth.Fingerprint,
			cfg.Auth.PrivateKey,
			common.String(cfg.Auth.PrivateKeyPassphrase))

	} else {
		conf = common.DefaultConfigProvider()
	}

	return conf, nil
}
