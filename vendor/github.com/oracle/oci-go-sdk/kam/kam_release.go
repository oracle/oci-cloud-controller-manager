// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// KAM API
//
// description: |
//   Kubernetes Add-on Manager API for installing, uninstalling and upgrading
//   OKE add-ons or Marketplace applications on an OKE cluster
//

package kam

import (
	"github.com/oracle/oci-go-sdk/common"
)

// KamRelease KAM Release
type KamRelease struct {

	// The OCID of the kam release.
	Id *string `mandatory:"false" json:"id"`

	// The OCID of the kam chart
	KamChartId *string `mandatory:"false" json:"kamChartId"`

	// The name of the package
	PackageName *string `mandatory:"false" json:"packageName"`

	// The type of package, like OKE Add-on or Marketplace application
	PackageType KamReleasePackageTypeEnum `mandatory:"false" json:"packageType,omitempty"`

	// The version of the package
	PackageVersion *string `mandatory:"false" json:"packageVersion"`

	// The current state of the release.
	LifecycleState ReleaseStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// A message describing the current state in more detail.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// The time the the release was created. An RFC3339 formatted datetime string
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The time the release was updated. An RFC3339 formatted datetime string
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// Simple key-value pair that is applied without any predefined name, type or scope. Exists for cross-compatibility only.
	// Example: `{"bar-key": "value"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m KamRelease) String() string {
	return common.PointerString(m)
}

// KamReleasePackageTypeEnum Enum with underlying type: string
type KamReleasePackageTypeEnum string

// Set of constants representing the allowable values for KamReleasePackageTypeEnum
const (
	KamReleasePackageTypeOkeAddon    KamReleasePackageTypeEnum = "OKE_ADDON"
	KamReleasePackageTypeApplication KamReleasePackageTypeEnum = "APPLICATION"
)

var mappingKamReleasePackageType = map[string]KamReleasePackageTypeEnum{
	"OKE_ADDON":   KamReleasePackageTypeOkeAddon,
	"APPLICATION": KamReleasePackageTypeApplication,
}

// GetKamReleasePackageTypeEnumValues Enumerates the set of values for KamReleasePackageTypeEnum
func GetKamReleasePackageTypeEnumValues() []KamReleasePackageTypeEnum {
	values := make([]KamReleasePackageTypeEnum, 0)
	for _, v := range mappingKamReleasePackageType {
		values = append(values, v)
	}
	return values
}
