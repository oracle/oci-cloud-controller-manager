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
	"reflect"
	"testing"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instance/metadata"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestConfigDefaulting(t *testing.T) {
	expectedCompartmentID := "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq"
	expectedRegion := "us-phoenix-1"
	expectedRegionKey := "phx"

	cfg := &Config{metadata: metadata.NewMock(
		&metadata.InstanceMetadata{
			CompartmentID:       expectedCompartmentID,
			CanonicalRegionName: expectedRegion,
			Region:              expectedRegionKey, // instance metadata API only returns the region key
		},
	)}

	err := cfg.setDefaults()
	if err != nil {
		t.Fatalf("cfg.setDefaults() => %v, expected no error", err)
	}

	if cfg.Auth.Region != expectedRegion {
		t.Fatalf("Expected cfg.Region = %q, got %q", cfg.Auth.Region, expectedRegion)
	}

	if cfg.Auth.RegionKey != expectedRegionKey {
		t.Fatalf("Expected cfg.RegionKey = %q, got %q", cfg.Auth.RegionKey, expectedRegionKey)
	}

	if cfg.Auth.CompartmentID != expectedCompartmentID {
		t.Fatalf("Expected cfg.CompartmentID = %q, got %q", cfg.Auth.CompartmentID, expectedCompartmentID)
	}
}

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name string
		in   *Config
		errs field.ErrorList
	}{
		{
			name: "valid",
			in: &Config{
				Config: providercfg.Config{
					Auth: providercfg.AuthConfig{
						Region:        "us-phoenix-1",
						RegionKey:     "phx",
						CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						TenancyID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
						Fingerprint:   "d4:1d:8c:d9:8f:00:b2:04:e9:80:09:98:ec:f8:42:7e",
					},
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "missing_region",
			in: &Config{
				Config: providercfg.Config{
					Auth: providercfg.AuthConfig{
						RegionKey:     "phx",
						CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						TenancyID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
						Fingerprint:   "d4:1d:8c:d9:8f:00:b2:04:e9:80:09:98:ec:f8:42:7e",
					},
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.region", BadValue: ""},
			},
		}, {
			name: "missing_region_key",
			in: &Config{
				Config: providercfg.Config{
					Auth: providercfg.AuthConfig{
						Region:        "us-phoenix-1",
						CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						TenancyID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
						Fingerprint:   "d4:1d:8c:d9:8f:00:b2:04:e9:80:09:98:ec:f8:42:7e",
					},
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.region_key", BadValue: ""},
			},
		}, {
			name: "missing_tenancy_ocid",
			in: &Config{
				Config: providercfg.Config{
					Auth: providercfg.AuthConfig{
						Region:        "us-phoenix-1",
						RegionKey:     "phx",
						CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
						Fingerprint:   "d4:1d:8c:d9:8f:00:b2:04:e9:80:09:98:ec:f8:42:7e",
					},
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.tenancy", BadValue: ""},
			},
		}, {
			name: "missing_user_ocid",
			in: &Config{
				Config: providercfg.Config{
					Auth: providercfg.AuthConfig{
						Region:        "us-phoenix-1",
						RegionKey:     "phx",
						CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						TenancyID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
						Fingerprint:   "d4:1d:8c:d9:8f:00:b2:04:e9:80:09:98:ec:f8:42:7e",
					},
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.user", BadValue: ""},
			},
		}, {
			name: "missing_key_file",
			in: &Config{
				Config: providercfg.Config{
					Auth: providercfg.AuthConfig{
						Region:        "us-phoenix-1",
						RegionKey:     "phx",
						CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						TenancyID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						Fingerprint:   "d4:1d:8c:d9:8f:00:b2:04:e9:80:09:98:ec:f8:42:7e",
					},
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.key", BadValue: ""},
			},
		}, {
			name: "missing_fingerprint",
			in: &Config{
				Config: providercfg.Config{
					Auth: providercfg.AuthConfig{
						Region:        "us-phoenix-1",
						RegionKey:     "phx",
						CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						TenancyID:     "ocid1.tennancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					},
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.fingerprint", BadValue: ""},
			},
		}, {
			name: "valid with instance principals enabled",
			in: &Config{
				Config: providercfg.Config{
					Auth: providercfg.AuthConfig{
						RegionKey: "phx",
					},
					UseInstancePrincipals: true,
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "mixing instance principals with other auth flags",
			in: &Config{
				Config: providercfg.Config{
					Auth: providercfg.AuthConfig{
						Region:      "us-phoenix-1",
						TenancyID:   "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
						UserID:      "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
						PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
						Fingerprint: "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
						RegionKey:   "phx",
					},
					UseInstancePrincipals: true,
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.region", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.tenancy", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.user", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.key", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.fingerprint", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateConfig(tt.in)
			if !reflect.DeepEqual(result, tt.errs) {
				t.Errorf("ValidateConfig(%+v)\n=> %+v\nExpected: %+v", tt.in, result, tt.errs)
			}
		})
	}
}
