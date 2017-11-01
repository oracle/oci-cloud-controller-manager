package client

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	netsets "k8s.io/kubernetes/pkg/util/net/sets"
)

// validateAuthConfig provides basic validation of AuthConfig instances.
func validateAuthConfig(c AuthConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if c.Region == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("region"), ""))
	}
	if c.TenancyOCID == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("tenancy"), ""))
	}
	if c.UserOCID == "" {
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
func validateLoadBalancerConfig(c LoadBalancerConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if c.Subnet1 == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("subnet1"), ""))
	}
	if c.Subnet2 == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("subnet2"), ""))
	}
	if len(c.DefaultSourceCIDRs) > 0 {
		_, err := netsets.ParseIPNets(c.DefaultSourceCIDRs...)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(
				fldPath.Child("defaultSourceCIDRs"),
				c.DefaultSourceCIDRs,
				"expecting a list of IP ranges. For example, 10.0.0.0/24.",
			))
		}
	}
	return allErrs
}

// ValidateConfig validates the OCI Cloud Provider config file.
func ValidateConfig(c *Config) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, validateAuthConfig(c.Auth, field.NewPath("auth"))...)
	allErrs = append(allErrs, validateLoadBalancerConfig(c.LoadBalancer, field.NewPath("loadBalancer"))...)
	return allErrs
}
