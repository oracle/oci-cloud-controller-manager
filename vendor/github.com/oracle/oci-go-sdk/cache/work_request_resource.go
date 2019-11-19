// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// OraCache Public API
//
// API for the Data Caching Service. Use this service to manage Redis replicated caches.
//

package cache

import (
	"github.com/oracle/oci-go-sdk/common"
)

// WorkRequestResource The details of a resource that a work request affects.
type WorkRequestResource struct {

	// The way in which the resource is affected.
	ActionType ResourceActionTypeEnum `mandatory:"true" json:"actionType"`

	// The type of the resource.
	EntityType *string `mandatory:"true" json:"entityType"`

	// The OCID of the resource.
	Identifier *string `mandatory:"true" json:"identifier"`

	// The URI path to the resource.
	EntityUri *string `mandatory:"true" json:"entityUri"`
}

func (m WorkRequestResource) String() string {
	return common.PointerString(m)
}

// WorkRequestResourceActionTypeEnum is an alias to type: ResourceActionTypeEnum
// Consider using ResourceActionTypeEnum instead
// Deprecated
type WorkRequestResourceActionTypeEnum = ResourceActionTypeEnum

// Set of constants representing the allowable values for ResourceActionTypeEnum
// Deprecated
const (
	WorkRequestResourceActionTypeCreated        ResourceActionTypeEnum = "CREATED"
	WorkRequestResourceActionTypeUpdated        ResourceActionTypeEnum = "UPDATED"
	WorkRequestResourceActionTypeDeleted        ResourceActionTypeEnum = "DELETED"
	WorkRequestResourceActionTypeInProgress     ResourceActionTypeEnum = "IN_PROGRESS"
	WorkRequestResourceActionTypeCanceledCreate ResourceActionTypeEnum = "CANCELED_CREATE"
	WorkRequestResourceActionTypeCanceledDelete ResourceActionTypeEnum = "CANCELED_DELETE"
	WorkRequestResourceActionTypeCanceledUpdate ResourceActionTypeEnum = "CANCELED_UPDATE"
	WorkRequestResourceActionTypeFailed         ResourceActionTypeEnum = "FAILED"
)

// GetWorkRequestResourceActionTypeEnumValues Enumerates the set of values for ResourceActionTypeEnum
// Consider using GetResourceActionTypeEnumValue
// Deprecated
var GetWorkRequestResourceActionTypeEnumValues = GetResourceActionTypeEnumValues
