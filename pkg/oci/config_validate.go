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
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// validateAuthConfig provides basic validation of AuthConfig instances.
func validateAuthConfig(c *AuthConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if c == nil {
		return append(allErrs, field.Required(fldPath, ""))
	}
	if c.UseInstancePrincipals {
		if c.Region != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("region"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.TenancyID != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("tenancy"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.UserID != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("user"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.PrivateKey != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("key"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.Fingerprint != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("fingerprint"), "cannot be used when useInstancePrincipals is enabled"))
		}

		return allErrs
	}
	if c.Region == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("region"), ""))
	}
	if c.TenancyID == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("tenancy"), ""))
	}
	if c.UserID == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("user"), ""))
	}
	if c.PrivateKey == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("key"), ""))
	}
	if c.Fingerprint == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("fingerprint"), ""))
	}
	return allErrs
}

// validateLoadBalancerConfig provides basic validation of LoadBalancerConfig
// instances.
func validateLoadBalancerConfig(c *LoadBalancerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if c == nil {
		return append(allErrs, field.Required(fldPath, ""))
	}
	if c.Subnet1 == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("subnet1"), ""))
	}
	if c.Subnet2 == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("subnet2"), ""))
	}
	if !IsValidSecurityListManagementMode(c.SecurityListManagementMode) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("securityListManagementMode"), c.SecurityListManagementMode, "invalid security list management mode"))
	}
	return allErrs
}

// ValidateConfig validates the OCI Cloud Provider config file.
func ValidateConfig(c *Config) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateAuthConfig(&c.Auth, field.NewPath("auth"))...)
	allErrs = append(allErrs, validateLoadBalancerConfig(&c.LoadBalancer, field.NewPath("loadBalancer"))...)
	return allErrs
}
