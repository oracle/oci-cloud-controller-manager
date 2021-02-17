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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// validateAuthConfig provides basic validation of AuthConfig instances.
func validateAuthConfig(c *AuthConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if c == nil {
		return append(allErrs, field.Required(fldPath, ""))
	}
	checkFields := map[string]string{
		"region":      c.Region,
		"tenancy":     c.TenancyID,
		"user":        c.UserID,
		"key":         c.PrivateKey,
		"fingerprint": c.Fingerprint,
	}
	for fieldName, fieldValue := range checkFields {
		if fieldValue == "" {
			if fieldName == "region" {
				allErrs = append(allErrs, field.InternalError(fldPath.Child(fieldName), errors.New("This value is normally discovered automatically if omitted. Continue checking the logs to see if something else is wrong")))
			} else {
				allErrs = append(allErrs, field.Required(fldPath.Child(fieldName), ""))
			}
		}
	}
	return allErrs
}

// SecurityListManagementModeChoices are the supported security list management
// modes.
var SecurityListManagementModeChoices = []string{ManagementModeAll, ManagementModeFrontend, ManagementModeNone}

// IsValidSecurityListManagementMode checks if a given security list management
// mode is valid.
func IsValidSecurityListManagementMode(mode string) bool {
	return sets.NewString(SecurityListManagementModeChoices...).Has(mode)
}

// validateLoadBalancerConfig provides basic validation of LoadBalancerConfig
// instances.
func validateLoadBalancerConfig(c *Config, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	lbConfig := c.LoadBalancer
	if lbConfig.Subnet1 == "" && c.VCNID == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("vcn"), "VCNID configuration must be provided if configuration for subnet1 is not provided"))
	}
	if !IsValidSecurityListManagementMode(lbConfig.SecurityListManagementMode) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("securityListManagementMode"),
			lbConfig.SecurityListManagementMode, "invalid security list management mode"))
	}
	return allErrs
}

// ValidateConfig validates the OCI Cloud Provider config file.
func ValidateConfig(c *Config) field.ErrorList {
	allErrs := field.ErrorList{}
	if len(c.CompartmentID) == 0 {
		allErrs = append(allErrs, field.InternalError(field.NewPath("compartment"), errors.New("This value is normally discovered automatically if omitted. Continue checking the logs to see if something else is wrong")))
	}
	if !c.UseInstancePrincipals {
		allErrs = append(allErrs, validateAuthConfig(&c.Auth, field.NewPath("auth"))...)
	}
	if c.LoadBalancer != nil && !c.LoadBalancer.Disabled {
		allErrs = append(allErrs, validateLoadBalancerConfig(c, field.NewPath("loadBalancer"))...)
	}

	// Validate metric config if the metrics block is not empty
	if c.Metrics != nil {
		if len(c.Metrics.CompartmentID) == 0 {
			// The compartment is required for pushing metrics to OCI Monitoring
			allErrs = append(allErrs, field.Required(field.NewPath("metrics", "compartment"), "Compartment is required for pushing custom metrics"))
		}

		if len(c.Metrics.Namespace) == 0 {
			allErrs = append(allErrs, field.Required(field.NewPath("metrics", "namespace"), "Metric namespace is required for pushing custom metrics"))
		}

		if len(c.Metrics.ResourceGroup) == 0 {
			allErrs = append(allErrs, field.Required(field.NewPath("metrics", "resourceGroup"), "Metric resource group is required for pushing custom metrics"))
		}
	}
	return allErrs
}
