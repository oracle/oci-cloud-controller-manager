// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// KAM API
//
// description: |
//   Kubernetes Add-on Manager API for installing, uninstalling and upgrading
//   OKE add-ons or Marketplace applications on an OKE cluster
//

package kam

// WorkRequestResourceMetadataKeyEnum Enum with underlying type: string
type WorkRequestResourceMetadataKeyEnum string

// Set of constants representing the allowable values for WorkRequestResourceMetadataKeyEnum
const (
	WorkRequestResourceMetadataKeyKamChartId     WorkRequestResourceMetadataKeyEnum = "KAM_CHART_ID"
	WorkRequestResourceMetadataKeyPackageName    WorkRequestResourceMetadataKeyEnum = "PACKAGE_NAME"
	WorkRequestResourceMetadataKeyPackageType    WorkRequestResourceMetadataKeyEnum = "PACKAGE_TYPE"
	WorkRequestResourceMetadataKeyPackageVersion WorkRequestResourceMetadataKeyEnum = "PACKAGE_VERSION"
	WorkRequestResourceMetadataKeyDescription    WorkRequestResourceMetadataKeyEnum = "DESCRIPTION"
)

var mappingWorkRequestResourceMetadataKey = map[string]WorkRequestResourceMetadataKeyEnum{
	"KAM_CHART_ID":    WorkRequestResourceMetadataKeyKamChartId,
	"PACKAGE_NAME":    WorkRequestResourceMetadataKeyPackageName,
	"PACKAGE_TYPE":    WorkRequestResourceMetadataKeyPackageType,
	"PACKAGE_VERSION": WorkRequestResourceMetadataKeyPackageVersion,
	"DESCRIPTION":     WorkRequestResourceMetadataKeyDescription,
}

// GetWorkRequestResourceMetadataKeyEnumValues Enumerates the set of values for WorkRequestResourceMetadataKeyEnum
func GetWorkRequestResourceMetadataKeyEnumValues() []WorkRequestResourceMetadataKeyEnum {
	values := make([]WorkRequestResourceMetadataKeyEnum, 0)
	for _, v := range mappingWorkRequestResourceMetadataKey {
		values = append(values, v)
	}
	return values
}
