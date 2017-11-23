// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.
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
	"errors"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// AuthConfig holds the configuration required for communicating with the OCI
// API.
type AuthConfig struct {
	Region          string `yaml:"region"`
	TenancyOCID     string `yaml:"tenancy"`
	CompartmentOCID string `yaml:"compartment"`
	UserOCID        string `yaml:"user"`
	PrivateKey      string `yaml:"key"`
	Fingerprint     string `yaml:"fingerprint"`
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

	Subnet1 string `yaml:"subnet1"`
	Subnet2 string `yaml:"subnet2"`
}

// Config holds the OCI cloud-provider config passed to Kubernetes compontents
// via the --cloud-config option.
type Config struct {
	Auth         AuthConfig         `yaml:"auth"`
	LoadBalancer LoadBalancerConfig `yaml:"loadBalancer"`
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

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
