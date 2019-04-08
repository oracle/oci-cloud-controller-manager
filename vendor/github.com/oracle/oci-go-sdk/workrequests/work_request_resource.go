// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Work Requests API
//
// A description of the work requests API
//

package workrequests

import (
	"github.com/oracle/oci-go-sdk/common"
)

// WorkRequestResource A resource created or operated on by a work request.
type WorkRequestResource struct {

	// The way in which this resource is affected by the work tracked in the work request.
	// A resource being created, updated, or deleted will remain in the IN_PROGRESS state until
	// work is complete for that resource at which point it will transition to CREATED, UPDATED,
	// or DELETED, respectively.
	ActionType *string `mandatory:"true" json:"actionType"`

	// The resource type the work request affects.
	EntityType *string `mandatory:"true" json:"entityType"`

	// An OCID or other unique identifier for the resource.
	Identifier *string `mandatory:"true" json:"identifier"`

	// The URI path used to access the resource metadata.
	EntityUri *string `mandatory:"false" json:"entityUri"`
}

func (m WorkRequestResource) String() string {
	return common.PointerString(m)
}
