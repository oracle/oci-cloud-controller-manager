// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// KAM API
//
// description: |
//   Kubernetes Add-on Manager API for installing, uninstalling and upgrading
//   OKE add-ons or Marketplace applications on an OKE cluster
//

package kam

// SortKamChartsByEnum Enum with underlying type: string
type SortKamChartsByEnum string

// Set of constants representing the allowable values for SortKamChartsByEnum
const (
	SortKamChartsByName    SortKamChartsByEnum = "NAME"
	SortKamChartsByType    SortKamChartsByEnum = "TYPE"
	SortKamChartsByVersion SortKamChartsByEnum = "VERSION"
)

var mappingSortKamChartsBy = map[string]SortKamChartsByEnum{
	"NAME":    SortKamChartsByName,
	"TYPE":    SortKamChartsByType,
	"VERSION": SortKamChartsByVersion,
}

// GetSortKamChartsByEnumValues Enumerates the set of values for SortKamChartsByEnum
func GetSortKamChartsByEnumValues() []SortKamChartsByEnum {
	values := make([]SortKamChartsByEnum, 0)
	for _, v := range mappingSortKamChartsBy {
		values = append(values, v)
	}
	return values
}
