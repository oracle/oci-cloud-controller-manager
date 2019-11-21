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

// ConnectionAliasSummary Summary representation of database aliases parsed from the file metadata
type ConnectionAliasSummary struct {

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	AliasName *string `mandatory:"true" json:"aliasName"`

	// The description about the database alias parsed from the file metadata
	AliasDetails *string `mandatory:"false" json:"aliasDetails"`
}

func (m ConnectionAliasSummary) String() string {
	return common.PointerString(m)
}
