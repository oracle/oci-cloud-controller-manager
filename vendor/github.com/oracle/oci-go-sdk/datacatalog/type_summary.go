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

// TypeSummary Summary Data Catalog Type Information. All types are statically defined in the system and are immutable.
// It isn't possible to create new types or update existing types via the api.
type TypeSummary struct {

	// Unique Type key that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The immutable name of the type.
	Name *string `mandatory:"false" json:"name"`

	// Detailed description of the Type.
	Description *string `mandatory:"false" json:"description"`

	// The Catalog's Oracle ID (OCID).
	CatalogId *string `mandatory:"false" json:"catalogId"`

	// Indicates the category this type belongs to. For instance , data assets , connections.
	TypeCategory *string `mandatory:"false" json:"typeCategory"`

	// URI to the Type instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// State of the Folder.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`
}

func (m TypeSummary) String() string {
	return common.PointerString(m)
}
