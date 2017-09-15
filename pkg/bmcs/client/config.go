// Copyright 2017 The Oracle Kubernetes Cloud Controller Manager Authors
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

	gcfg "gopkg.in/gcfg.v1"
)

// Config holds the BMCS cloud-provider config passed to Kubernetes compontents
// via the --cloud-config option.
type Config struct {
	Global struct {
		UserOCID        string `gcfg:"user"`
		CompartmentOCID string `gcfg:"compartment"`
		TenancyOCID     string `gcfg:"tenancy"`
		Fingerprint     string `gcfg:"fingerprint"`
		PrivateKeyFile  string `gcfg:"key-file"`
		Region          string `gcfg:"region"`
		// DisableSecurityListManagement disables the automatic creation of ingress
		// rules for the node subnets and egress rules for the load balancers to the node subnets.
		//
		// If security list management is disabled, then it requires that the user
		// has setup a rule that allows inbound traffic to the appropriate ports
		// for kube proxy health port, node port ranges, and health check port ranges.
		// E.g. 10.82.0.0/16 30000-32000
		DisableSecurityListManagement bool `gcfg:"disableSecurityListManagement"`
	}
	LoadBalancer struct {
		Subnet1 string `gcfg:"subnet1"`
		Subnet2 string `gcfg:"subnet2"`
	}
}

// ReadConfig consumes the config Reader and constructs a Config object.
func ReadConfig(r io.Reader) (*Config, error) {
	if r == nil {
		return nil, errors.New("no cloud-provider config file given")
	}

	cfg := &Config{}

	err := gcfg.ReadInto(cfg, r)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
