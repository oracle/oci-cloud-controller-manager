// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DataCatalog API
//
// A description of the DataCatalog API
//

package datacatalog

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateDataAssetDetails Properties used in Data Asset create operations.
type CreateDataAssetDetails struct {

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The key of the Data Asset type. This can be obtained via the '/types' endpoint.
	TypeKey *string `mandatory:"true" json:"typeKey"`

	// Detailed description of the Data Asset.
	Description *string `mandatory:"false" json:"description"`

	// A map of maps which contains the properties which are specific to the asset type. Each Data Asset type
	// definition defines it's set of required and optional properties. The map keys are category names and the
	// values are maps of property name to property value. Every property is contained inside of a category. Most
	// Data Assets have required properties within the "default" category. To determine the set of optional and
	// required properties for a Data Asset type, a query can be done on '/types?type=dataAsset' which returns a
	// collection of all Data Asset types. The appropriate Data Asset type, which includes definitions of all of
	// it's properties, can be identified from this collection.
	// Example: `{"properties": { "default": { "host": "host1", "port": "1521", "database": "orcl"}}}`
	Properties map[string]map[string]string `mandatory:"false" json:"properties"`
}

func (m CreateDataAssetDetails) String() string {
	return common.PointerString(m)
}
