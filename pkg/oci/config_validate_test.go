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
	"reflect"
	"testing"

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
				Auth: AuthConfig{
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{},
		},
		{
			name: "valid with instance principals enabled",
			in: &Config{
				Auth: AuthConfig{
					UseInstancePrincipals: true,
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "mixing instance principals with other auth flags",
			in: &Config{
				Auth: AuthConfig{
					UseInstancePrincipals: true,
					Region:                "us-phoenix-1",
					TenancyID:             "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					UserID:                "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:            "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:           "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.region", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.tenancy", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.user", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.key", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
				&field.Error{Type: field.ErrorTypeForbidden, Field: "auth.fingerprint", Detail: "cannot be used when useInstancePrincipals is enabled", BadValue: ""},
			},
		}, {
			name: "valid_with_non_default_security_list_management_mode",
			in: &Config{
				Auth: AuthConfig{
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1:                    "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2:                    "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
					SecurityListManagementMode: ManagementModeFrontend,
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "missing_region",
			in: &Config{
				Auth: AuthConfig{
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.region", BadValue: ""},
			},
		}, {
			name: "missing_tenancy",
			in: &Config{
				Auth: AuthConfig{
					Region:        "us-phoenix-1",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.tenancy", BadValue: ""},
			},
		}, {
			name: "missing_compartment",
			in: &Config{
				Auth: AuthConfig{
					Region:      "us-phoenix-1",
					TenancyID:   "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					UserID:      "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint: "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "missing_user",
			in: &Config{
				Auth: AuthConfig{
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.user", BadValue: ""},
			},
		}, {
			name: "missing_key",
			in: &Config{
				Auth: AuthConfig{
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.key", BadValue: ""},
			},
		}, {
			name: "missing_figerprint",
			in: &Config{
				Auth: AuthConfig{
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.fingerprint", BadValue: ""},
			},
		}, {
			name: "missing_subnet1",
			in: &Config{
				Auth: AuthConfig{
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet2: "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "loadBalancer.subnet1", BadValue: ""},
			},
		}, {
			name: "missing_subnet2",
			in: &Config{
				Auth: AuthConfig{
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "loadBalancer.subnet2", BadValue: ""},
			},
		}, {
			name: "invalid_security_list_management_mode",
			in: &Config{
				Auth: AuthConfig{
					Region:        "us-phoenix-1",
					TenancyID:     "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					CompartmentID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
					UserID:        "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:    "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint:   "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				LoadBalancer: LoadBalancerConfig{
					Subnet1:                    "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					Subnet2:                    "ocid1.subnet.oc1.phx.aaaaaaaahuxrgvs65iwdz7ekwgg3l5gyah7ww5klkwjcso74u3e4i64hvtvq",
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
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tt.in.Complete()
			result := ValidateConfig(tt.in)
			if !reflect.DeepEqual(result, tt.errs) {
				t.Errorf("ValidateConfig(%#v)\n=>        %q \nExpected: %q", tt.in, result, tt.errs)
			}
		})
	}
}
