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

// UpdateTermRelationshipDetails Properties used in Term Relationship update operations.
type UpdateTermRelationshipDetails struct {

	// The display name of a user-friendly name. Is changeable. The combination of displayName and parentTermKey
	// must be unique. Avoid entering confidential information.This is the same as relationshipType for termRelationship
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Detailed description of the Term Relationship usually defined at the time of creation.
	Description *string `mandatory:"false" json:"description"`
}

func (m UpdateTermRelationshipDetails) String() string {
	return common.PointerString(m)
}
