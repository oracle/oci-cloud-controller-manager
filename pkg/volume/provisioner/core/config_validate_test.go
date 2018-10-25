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
					Region:      "us-phoenix-1",
					TenancyOCID: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					UserOCID:    "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint: "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
			},
			errs: field.ErrorList{},
		}, {
			name: "missing_region",
			in: &Config{
				Auth: AuthConfig{
					TenancyOCID: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					UserOCID:    "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint: "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.region", BadValue: ""},
			},
		}, {
			name: "missing_tenancy",
			in: &Config{
				Auth: AuthConfig{
					Region:      "us-phoenix-1",
					UserOCID:    "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint: "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.tenancy", BadValue: ""},
			},
		}, {
			name: "missing_user",
			in: &Config{
				Auth: AuthConfig{
					Region:      "us-phoenix-1",
					TenancyOCID: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint: "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.user", BadValue: ""},
			},
		}, {
			name: "missing_key",
			in: &Config{
				Auth: AuthConfig{
					Region:      "us-phoenix-1",
					TenancyOCID: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					UserOCID:    "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					Fingerprint: "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.key", BadValue: ""},
			},
		}, {
			name: "missing_figerprint",
			in: &Config{
				Auth: AuthConfig{
					Region:      "us-phoenix-1",
					TenancyOCID: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					UserOCID:    "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
				},
			},
			errs: field.ErrorList{
				&field.Error{Type: field.ErrorTypeRequired, Field: "auth.fingerprint", BadValue: ""},
			},
		}, {
			name: "valid with instance principals enabled",
			in: &Config{
				UseInstancePrincipals: true,
			},
			errs: field.ErrorList{},
		}, {
			name: "mixing instance principals with other auth flags",
			in: &Config{
				Auth: AuthConfig{

					Region:      "us-phoenix-1",
					TenancyOCID: "ocid1.tenancy.oc1..aaaaaaaatyn7scrtwtqedvgrxgr2xunzeo6uanvyhzxqblctwkrpisvke4kq",
					UserOCID:    "ocid1.user.oc1..aaaaaaaai77mql2xerv7cn6wu3nhxang3y4jk56vo5bn5l5lysl34avnui3q",
					PrivateKey:  "-----BEGIN RSA PRIVATE KEY----- (etc)",
					Fingerprint: "8c:bf:17:7b:5f:e0:7d:13:75:11:d6:39:0d:e2:84:74",
				},
				UseInstancePrincipals: true,
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
				t.Errorf("ValidateConfig(%#v)\n=> %#v\nExpected: %#v", tt.in, result, tt.errs)
			}
		})
	}
}
