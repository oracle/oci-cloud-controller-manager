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

// KamChartSummary An OKE Add-on or Marketplace application, the OCID of this record may be
// used for installation or upgrade
type KamChartSummary struct {

	// The OCID of the kam chart.
	Id *string `mandatory:"true" json:"id"`

	// The name of the OKE Add-on or Marketplace app
	PackageName *string `mandatory:"false" json:"packageName"`

	// The description of the package
	PackageDescription *string `mandatory:"false" json:"packageDescription"`

	// The type of package, like OKE Add-on or Marketplace application
	PackageType KamChartSummaryPackageTypeEnum `mandatory:"false" json:"packageType,omitempty"`

	// The version of the OKE Add-on or Marketplace app
	Version *string `mandatory:"false" json:"version"`
}

func (m KamChartSummary) String() string {
	return common.PointerString(m)
}

// KamChartSummaryPackageTypeEnum Enum with underlying type: string
type KamChartSummaryPackageTypeEnum string

// Set of constants representing the allowable values for KamChartSummaryPackageTypeEnum
const (
	KamChartSummaryPackageTypeOkeAddon    KamChartSummaryPackageTypeEnum = "OKE_ADDON"
	KamChartSummaryPackageTypeApplication KamChartSummaryPackageTypeEnum = "APPLICATION"
)

var mappingKamChartSummaryPackageType = map[string]KamChartSummaryPackageTypeEnum{
	"OKE_ADDON":   KamChartSummaryPackageTypeOkeAddon,
	"APPLICATION": KamChartSummaryPackageTypeApplication,
}

// GetKamChartSummaryPackageTypeEnumValues Enumerates the set of values for KamChartSummaryPackageTypeEnum
func GetKamChartSummaryPackageTypeEnumValues() []KamChartSummaryPackageTypeEnum {
	values := make([]KamChartSummaryPackageTypeEnum, 0)
	for _, v := range mappingKamChartSummaryPackageType {
		values = append(values, v)
	}
	return values
}
