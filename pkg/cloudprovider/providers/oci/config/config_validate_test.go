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
	"reflect"
	"testing"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instance/metadata"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestValidateConfig(t *testing.T) {
	testCases := []struct {
		name string
		in   *Config
		errs field.ErrorList
	}{
		{
			name: "valid",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{},
		},
		{
			name: "valid with instance principals enabled",
			in: &Config{
				metadataSvc: metadata.NewMock(&metadata.InstanceMetadata{CompartmentID: "compartment"}),
				Auth: AuthConfig{
					metadataSvc:           metadata.NewMock(&metadata.InstanceMetadata{CompartmentID: "compartment"}),
					UseInstancePrincipals: true,
					TenancyID:             "not empty",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "valid_with_non_default_security_list_management_mode",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1:                    "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2:                    "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					SecurityListManagementMode: ManagementModeFrontend,
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "missing_region",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeInternal, Field: "auth.region", Detail: "This value is normally discovered automatically if omitted. Continue checking the logs to see if something else is wrong"},
			},
		}, {
			name: "missing_tenancy",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.tenancy", BadValue: ""},
			},
		}, {
			name: "missing_compartment",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc: metadata.NewErrorMock(),
					Region:      "us-phoenix-1",
					TenancyID:   "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:      "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint: "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeInternal, Field: "compartment", Detail: "This value is normally discovered automatically if omitted. Continue checking the logs to see if something else is wrong"},
			},
		}, {
			name: "missing_user",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.user", BadValue: ""},
			},
		}, {
			name: "missing_key",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.key", BadValue: ""},
			},
		}, {
			name: "missing_fingerprint",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.fingerprint", BadValue: ""},
			},
		}, {
			name: "missing_vcnid",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "vcn", BadValue: "", Detail: "VCNID configuration must be provided if configuration for subnet1 is not provided"},
			},
		}, {
			name: "missing_lbconfig",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: nil,
			},
			errs: field.ErrorList{},
		}, {
			name: "invalid_security_list_management_mode",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1:                    "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2:                    "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					SecurityListManagementMode: "invalid",
				},
			},
			errs: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeInvalid,
					Field:    "loadBalancer.securityListManagementMode",
					BadValue: "invalid",
					Detail:   "invalid security list management mode",
				},
			},
		},
		{
			name: "valid config for Auth, LB and metrics",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
				Metrics: &MetricsConfig{
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Namespace:     "test",
					ResourceGroup: "test-rg",
					Prefix:        "Prefix.",
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "valid config for Auth, LB and invalid config for metrics",
			in: &Config{
				metadataSvc: metadata.NewErrorMock(),
				Auth: AuthConfig{
					metadataSvc:   metadata.NewErrorMock(),
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					UserID:        "ocid1.user.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: &LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				},
				Metrics: &MetricsConfig{
					CompartmentID: "",
					Namespace:     "",
					ResourceGroup: "",
					Prefix:        "Prefix.",
				},
			},
			errs: field.ErrorList{
				&field.Error{
					Type:     field.ErrorTypeRequired,
					Field:    "metrics.compartment",
					BadValue: "",
					Detail:   "Compartment is required for pushing custom metrics",
				},
				&field.Error{
					Type:     field.ErrorTypeRequired,
					Field:    "metrics.namespace",
					BadValue: "",
					Detail:   "Metric namespace is required for pushing custom metrics",
				},
				&field.Error{
					Type:     field.ErrorTypeRequired,
					Field:    "metrics.resourceGroup",
					BadValue: "",
					Detail:   "Metric resource group is required for pushing custom metrics",
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.in.Complete()
			result := ValidateConfig(tt.in)
			if !reflect.DeepEqual(result, tt.errs) {
				t.Errorf("ValidateConfig (%s) \n(%#v)\n=>        %q \nExpected: %q", tt.name, tt.in, result, tt.errs)
			}
		})
	}
}
