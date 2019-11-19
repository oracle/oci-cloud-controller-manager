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

// GlossaryPermissionsSummary Permissions object for Glosssaries
type GlossaryPermissionsSummary struct {

	// An array of permissions
	UserPermissions []string `mandatory:"false" json:"userPermissions"`

	// The unique key of the parent Glossary.
	GlossaryKey *string `mandatory:"false" json:"glossaryKey"`
}

func (m GlossaryPermissionsSummary) String() string {
	return common.PointerString(m)
}
