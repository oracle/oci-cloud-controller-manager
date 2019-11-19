// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// KAM API
//
// description: |
//   Kubernetes Add-on Manager API for installing, uninstalling and upgrading
//   OKE add-ons or Marketplace applications on an OKE cluster
//

package kam

// PackageTypeEnum Enum with underlying type: string
type PackageTypeEnum string

// Set of constants representing the allowable values for PackageTypeEnum
const (
	PackageTypeOkeAddon    PackageTypeEnum = "OKE_ADDON"
	PackageTypeApplication PackageTypeEnum = "APPLICATION"
)

var mappingPackageType = map[string]PackageTypeEnum{
	"OKE_ADDON":   PackageTypeOkeAddon,
	"APPLICATION": PackageTypeApplication,
}

// GetPackageTypeEnumValues Enumerates the set of values for PackageTypeEnum
func GetPackageTypeEnumValues() []PackageTypeEnum {
	values := make([]PackageTypeEnum, 0)
	for _, v := range mappingPackageType {
		values = append(values, v)
	}
	return values
}
