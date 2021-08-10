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

package config

import (
	"strings"
	"testing"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instance/metadata"
)

func TestReadConfigShouldFailWhenNoConfigProvided(t *testing.T) {
	_, err := ReadConfig(nil)
	if err == nil {
		t.Fatalf("should fail with when given no config")
	}
}

const validConfig = `
auth:
  region: us-phoenix-1
  tenancy: ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  user: ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  key: |
    -----BEGIN RSA PRIVATE KEY-----
    -----END RSA PRIVATE KEY-----
  fingerprint: 97:84:f7:26:a3:7b:74:d0:bd:4e:08:a7:79:c9:d0:1d

useInstancePrincipals: false
vcn: ocid1.vcn.oc1..
compartment: ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa

loadBalancer:
  disableSecurityListManagement: false
  subnet1: ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  subnet2: ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa

tags:
  loadBalancer:
    freeform:
      test: tags
    defined:
      namespace:
        key: value
  blockVolume:
    freeform:
      test: tags
    defined:
      namespace:
        key: value
`

const validConfigNoLoadbalancing = `
useInstancePrincipals: true
vcn: ocid1.vcn.oc1..

loadBalancer:
  disabled: true
`

const validConfigLegacyFormat = `
auth:
  region: us-phoenix-1
  regionKey: phx
  tenancy: ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  user: ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  key: |
    -----BEGIN RSA PRIVATE KEY-----
    -----END RSA PRIVATE KEY-----
  fingerprint: 97:84:f7:26:a3:7b:74:d0:bd:4e:08:a7:79:c9:d0:1d

  key_passphrase: secretpassphrase
  useInstancePrincipals: true
  compartment: ocid1.compartment.oc1
loadBalancer:
  disableSecurityListManagement: true
`

const validConfigNoRegion = `
auth:
  tenancy: ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  compartment: ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  user: ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  key: |
    -----BEGIN RSA PRIVATE KEY-----
    -----END RSA PRIVATE KEY-----
  fingerprint: 97:84:f7:26:a3:7b:74:d0:bd:4e:08:a7:79:c9:d0:1d

loadBalancer:
  disableSecurityListManagement: false
  subnet1: ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  subnet2: ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
`

func TestReadConfigShouldSucceedWhenProvidedValidConfig(t *testing.T) {
	_, err := ReadConfig(strings.NewReader(validConfig))
	if err != nil {
		t.Fatalf("expected no error but got '%+v'", err)
	}
}

func TestReadConfigShouldHaveNoDefaultRegionIfNoneSpecified(t *testing.T) {
	config, err := ReadConfig(strings.NewReader(validConfigNoRegion))

	if err != nil {
		t.Fatalf("expected no error but got '%+v'", err)
	}
	if config.Auth.Region != "" {
		t.Errorf("expected no region but got %s", config.Auth.Region)
	}
}

func TestReadConfigShouldSetCompartmentIDWhenProvidedValidConfig(t *testing.T) {
	cfg, err := ReadConfig(strings.NewReader(validConfig))
	if err != nil {
		t.Fatalf("expected no error but got '%+v'", err)
	}
	expected := "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	if cfg.CompartmentID != expected {
		t.Errorf("Got CompartmentID = %s; want CompartmentID = %s",
			cfg.CompartmentID, expected)
	}
}

func TestBackwardsCompatibilityFieldsAreSetCorrectly(t *testing.T) {
	cfg, err := ReadConfig(strings.NewReader(validConfigLegacyFormat))
	if err != nil {
		t.Fatalf("expected no error but got '%v'", err)
	}

	if cfg.CompartmentID != "ocid1.compartment.oc1" {
		t.Errorf("Compartment ID was not set correctly: cfg.CompartmentID = %v", cfg.CompartmentID)
	}

	if cfg.Auth.Passphrase != "secretpassphrase" {
		t.Errorf("Passphrase was not set correctly: cfg.Auth.Passphrase = %v", cfg.Auth.Passphrase)
	}

	if cfg.LoadBalancer.SecurityListManagementMode != ManagementModeNone {
		t.Errorf("Management mode should be set to None. It was set to %v", cfg.LoadBalancer.SecurityListManagementMode)
	}

	if cfg.RegionKey != "phx" {
		t.Errorf("Region key was not set correctly: cfg.RegionKey = %v", cfg.RegionKey)
	}
}

func TestLoadBalancingDisabled(t *testing.T) {
	cfg, err := ReadConfig(strings.NewReader(validConfigNoLoadbalancing))
	if err != nil {
		t.Fatalf("expected no error but got '%v'", err)
	}

	if !cfg.LoadBalancer.Disabled {
		t.Errorf("Load balancing should be disabled")
	}
}

func TestMetadataSvcSetsOmittedFields(t *testing.T) {
	mockMetadataSvc := metadata.NewMock(&metadata.InstanceMetadata{
		Region:              "mockRegion",
		CanonicalRegionName: "mockCanonicalRegionName",
		CompartmentID:       "mockCompartmentId",
	})
	cfg := &Config{
		Auth: AuthConfig{},
	}
	cfg.metadataSvc = mockMetadataSvc
	cfg.Auth.metadataSvc = mockMetadataSvc
	cfg.Complete()

	if cfg.CompartmentID != "mockCompartmentId" {
		t.Errorf("Metadata service does not set compartmentID")
	}
	if cfg.RegionKey != "mockRegion" {
		t.Errorf("Metadata service does not set the Region Key")
	}
	if cfg.Auth.Region != "mockCanonicalRegionName" {
		t.Errorf("Metadata service does not set the Region")
	}
}
