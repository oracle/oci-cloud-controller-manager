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

package block

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instance/metadata"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var ociRegions = map[string]string{
	"phx": "us-phoenix-1",
	"iad": "us-ashburn-1",
	"fra": "eu-frankfurt-1",
	"lhr": "uk-london-1",
}

// Config holds the configuration for the OCI flexvolume driver.
type Config struct {
	providercfg.Config `yaml:",inline"`

	metadata metadata.Interface
}

// NewConfig creates a new Config based on the contents of the given io.Reader.
func NewConfig(r io.Reader) (*Config, error) {
	if r == nil {
		return nil, errors.New("no config provided")
	}

	c := &Config{}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return nil, err
	}

	c.metadata = metadata.New()

	if !c.UseInstancePrincipals {
		if err := c.setDefaults(); err != nil {
			return nil, err
		}
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// ConfigFromFile reads the file at the given path and marshals it into a Config
// object.
func ConfigFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	return NewConfig(f)
}

func (c *Config) setDefaults() error {
	if c.Auth.Region == "" || c.Auth.CompartmentID == "" {
		meta, err := c.metadata.Get()
		if err != nil {
			return err
		}

		if c.Auth.Region == "" {
			c.Auth.Region = meta.Region
		}
		if c.Auth.CompartmentID == "" {
			c.Auth.CompartmentID = meta.CompartmentOCID
		}
	}

	err := c.setRegionFields(c.Auth.Region)
	if err != nil {
		return fmt.Errorf("setting config region fields: %v", err)
	}

	if c.Auth.Passphrase == "" && c.Auth.PrivateKeyPassphrase != "" {
		log.Print("config: auth.key_passphrase is DEPRECIATED and will be removed in a later release. Please set auth.passphrase instead.")
		c.Auth.Passphrase = c.Auth.PrivateKeyPassphrase
	}

	return nil
}

// setRegionFields accepts either a region short name or a region long name and
// sets both the Region and RegionKey fields.
func (c *Config) setRegionFields(region string) error {
	input := region
	region = strings.ToLower(region)

	var name, key string
	name, ok := ociRegions[region]
	if !ok {
		for key, name = range ociRegions {
			if name == region {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("tried to connect to unsupported OCI region %q", input)
		}
	} else {
		key = region
	}

	c.Auth.Region = name
	c.Auth.RegionKey = key

	return nil
}

// validate checks that all required fields are populated.
func (c *Config) validate() error {
	return ValidateConfig(c).ToAggregate()
}

func validateAuthConfig(c *Config, fldPath *field.Path) field.ErrorList {
	errList := field.ErrorList{}

	if c.UseInstancePrincipals {
		if c.Auth.Region != "" {
			errList = append(errList, field.Forbidden(fldPath.Child("region"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.Auth.CompartmentID != "" {
			errList = append(errList, field.Forbidden(fldPath.Child("compartment"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.Auth.TenancyID != "" {
			errList = append(errList, field.Forbidden(fldPath.Child("tenancy"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.Auth.UserID != "" {
			errList = append(errList, field.Forbidden(fldPath.Child("user"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.Auth.PrivateKey != "" {
			errList = append(errList, field.Forbidden(fldPath.Child("key"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.Auth.Fingerprint != "" {
			errList = append(errList, field.Forbidden(fldPath.Child("fingerprint"), "cannot be used when useInstancePrincipals is enabled"))
		}
	} else {
		if c.Auth.Region == "" {
			errList = append(errList, field.Required(fldPath.Child("region"), ""))
		}
		if c.Auth.TenancyID == "" {
			errList = append(errList, field.Required(fldPath.Child("tenancy"), ""))
		}
		if c.Auth.UserID == "" {
			errList = append(errList, field.Required(fldPath.Child("user"), ""))
		}
		if c.Auth.PrivateKey == "" {
			errList = append(errList, field.Required(fldPath.Child("key"), ""))
		}
		if c.Auth.Fingerprint == "" {
			errList = append(errList, field.Required(fldPath.Child("fingerprint"), ""))
		}
	}

	if c.Auth.RegionKey == "" {
		errList = append(errList, field.Required(fldPath.Child("region_key"), ""))
	}

	if c.Auth.VCNID == "" {
		errList = append(errList, field.Required(fldPath.Child("vcn"), ""))
	}

	return errList
}

// ValidateConfig validates the OCI Flexible Volume Provisioner config file.
func ValidateConfig(c *Config) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateAuthConfig(c, field.NewPath("auth"))...)
	return allErrs
}

func configurationProviderFromConfig(config *Config) (common.ConfigurationProvider, error) {
	if config.UseInstancePrincipals {
		cp, err := auth.InstancePrincipalConfigurationProvider()
		if err != nil {
			return nil, err
		}
		return cp, nil
	}

	return common.NewRawConfigurationProvider(
		config.Auth.TenancyID,
		config.Auth.UserID,
		config.Auth.Region,
		config.Auth.Fingerprint,
		config.Auth.PrivateKey,
		&config.Auth.Passphrase,
	), nil
}
