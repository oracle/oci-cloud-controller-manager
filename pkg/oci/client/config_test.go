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
	"strings"
	"testing"
)

func TestReadConfigShouldFailWhenNoConfigProvided(t *testing.T) {
	_, err := ReadConfig(nil)
	if err == nil {
		t.Fatalf("should fail with when given no config")
	}
}

const validConfig = `
[Global]
user = ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
fingerprint = 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00
key-file = test.pem
tenancy = ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
compartment = ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
region = us-phoenix-1
`

const validConfigNoRegion = `
[Global]
user = ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
fingerprint = 00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00
key-file = test.pem
tenancy = ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
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
	if config.Global.Region != "" {
		t.Errorf("expected no region but got %s", config.Global.Region)
	}
}

func TestReadConfigShouldSetCompartmentOCIDWhenProvidedValidConfig(t *testing.T) {
	cfg, err := ReadConfig(strings.NewReader(validConfig))
	if err != nil {
		t.Fatalf("expected no error but got '%+v'", err)
	}
	expected := "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	if cfg.Global.CompartmentOCID != expected {
		t.Errorf("Got Global.CompartmentOCID = %s; want Global.CompartmentOCID = %s",
			cfg.Global.CompartmentOCID, expected)
	}
}
