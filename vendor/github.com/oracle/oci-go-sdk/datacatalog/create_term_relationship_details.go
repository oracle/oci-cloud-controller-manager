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

// CreateTermRelationshipDetails Properties used in Term Relationship create operations.
type CreateTermRelationshipDetails struct {

	// The display name of a user-friendly name. Is changeable. The combination of displayName and parentTermKey
	// must be unique. Avoid entering confidential information.This is the same as relationshipType for termRelationship
	DisplayName *string `mandatory:"true" json:"displayName"`

	// Unique id of the related term.
	RelatedTermKey *string `mandatory:"true" json:"relatedTermKey"`

	// Detailed description of the Term Relationship usually defined at the time of creation.
	Description *string `mandatory:"false" json:"description"`
}

func (m CreateTermRelationshipDetails) String() string {
	return common.PointerString(m)
}
