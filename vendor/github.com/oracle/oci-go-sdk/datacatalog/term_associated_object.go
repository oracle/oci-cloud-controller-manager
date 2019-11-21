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

// TermAssociatedObject Projection of an object that is tagged to a term.
type TermAssociatedObject struct {

	// Immutable key used to uniquely identify the associated object.
	Key *string `mandatory:"true" json:"key"`

	// Name of the associated object.
	Name *string `mandatory:"false" json:"name"`

	// URI of the associated object within the data catalog API.
	Uri *string `mandatory:"false" json:"uri"`
}

func (m TermAssociatedObject) String() string {
	return common.PointerString(m)
}
