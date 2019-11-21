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

// UpdateGlossaryDetails Properties used in Glossary update operations.
type UpdateGlossaryDetails struct {

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Detailed description of the Glossary.
	Description *string `mandatory:"false" json:"description"`

	// Id (OCID) of the user who is the owner of the glossary.
	Owner *string `mandatory:"false" json:"owner"`
}

func (m UpdateGlossaryDetails) String() string {
	return common.PointerString(m)
}
