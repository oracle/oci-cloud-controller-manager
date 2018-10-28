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

package core

import (
	"io"
	"io/ioutil"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// Config holds the OCI cloud-provider config passed to Kubernetes compontents.
type Config struct {
	providercfg.Config `yaml:",inline"`
}

// Validate validates the OCI config.
func (c *Config) Validate() error {
	return ValidateConfig(c).ToAggregate()
}

// LoadConfig consumes the config Reader and constructs a Config object.
func LoadConfig(r io.Reader) (*Config, error) {
	if r == nil {
		return nil, errors.New("no configuration file given")
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

	cfg.Complete()

	return cfg, nil
}

// Complete the config applying defaults / overrides.
func (c *Config) Complete() {
	if c.CompartmentID == "" && c.Auth.CompartmentID != "" {
		zap.S().Warn("cloud-provider config: \"auth.compartment\" is DEPRECATED and will be removed in a later release. Please set \"compartment\".")
		c.CompartmentID = c.Auth.CompartmentID
	}
}

func newConfigurationProvider(logger *zap.SugaredLogger, cfg *Config) (common.ConfigurationProvider, error) {
	var conf common.ConfigurationProvider
	if cfg != nil {
		err := cfg.Validate()
		if err != nil {
			return nil, errors.Wrap(err, "invalid client config")
		}
		if cfg.UseInstancePrincipals {
			logger.Info("Using instance principals configuration provider.")
			cp, err := auth.InstancePrincipalConfigurationProvider()
			if err != nil {
				return nil, errors.Wrap(err, "InstancePrincipalConfigurationProvider")
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
