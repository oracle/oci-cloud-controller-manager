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

// AttributeTag Represents an association of an Entity Attribute to a Term.
type AttributeTag struct {

	// Unique tag key that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// Name of the tag which matches the term name.
	Name *string `mandatory:"false" json:"name"`

	// Unique key of the related term.
	TermKey *string `mandatory:"false" json:"termKey"`

	// Path of the related term.
	TermPath *string `mandatory:"false" json:"termPath"`

	// Description of the related term.
	TermDescription *string `mandatory:"false" json:"termDescription"`

	// The current state of the Tag.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the Tag was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Id (OCID) of the user who created the Tag.
	CreatedById *string `mandatory:"false" json:"createdById"`

	// URI to the Tag instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// The unique key of the parent Attribute.
	AttributeKey *string `mandatory:"false" json:"attributeKey"`
}

func (m AttributeTag) String() string {
	return common.PointerString(m)
}
