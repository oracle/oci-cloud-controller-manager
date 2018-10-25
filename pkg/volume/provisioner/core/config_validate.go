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

package core

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func validateAuthConfig(c Config, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if c.UseInstancePrincipals {
		if c.Auth.Region != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("region"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.Auth.TenancyOCID != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("tenancy"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.Auth.UserOCID != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("user"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.Auth.PrivateKey != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("key"), "cannot be used when useInstancePrincipals is enabled"))
		}
		if c.Auth.Fingerprint != "" {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("fingerprint"), "cannot be used when useInstancePrincipals is enabled"))
		}
		return allErrs
	}

	if c.Auth.Region == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("region"), ""))
	}
	if c.Auth.TenancyOCID == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("tenancy"), ""))
	}
	if c.Auth.UserOCID == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("user"), ""))
	}
	if c.Auth.PrivateKey == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("key"), ""))
	}
	if c.Auth.Fingerprint == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("fingerprint"), ""))
	}
	return allErrs
}

// ValidateConfig validates the OCI Volume Provisioner config file.
func ValidateConfig(c *Config) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateAuthConfig(*c, field.NewPath("auth"))...)
	return allErrs
}
