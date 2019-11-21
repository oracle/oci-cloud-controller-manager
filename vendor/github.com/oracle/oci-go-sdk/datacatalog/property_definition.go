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

// PropertyDefinition Details of a single Type property.
type PropertyDefinition struct {

	// Name of the property.
	Name *string `mandatory:"false" json:"name"`

	// The properties value type.
	Type *string `mandatory:"false" json:"type"`

	// Whether instances of the Type are required to set this property.
	IsRequired *bool `mandatory:"false" json:"isRequired"`

	// Indicates if this property value can be updated.
	IsUpdatable *bool `mandatory:"false" json:"isUpdatable"`
}

func (m PropertyDefinition) String() string {
	return common.PointerString(m)
}
