// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
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
	ActionType WorkRequestResourceActionTypeEnum `mandatory:"true" json:"actionType"`

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

// WorkRequestResourceActionTypeEnum Enum with underlying type: string
type WorkRequestResourceActionTypeEnum string

// Set of constants representing the allowable values for WorkRequestResourceActionTypeEnum
const (
	WorkRequestResourceActionTypeCreated        WorkRequestResourceActionTypeEnum = "CREATED"
	WorkRequestResourceActionTypeUpdated        WorkRequestResourceActionTypeEnum = "UPDATED"
	WorkRequestResourceActionTypeDeleted        WorkRequestResourceActionTypeEnum = "DELETED"
	WorkRequestResourceActionTypeInProgress     WorkRequestResourceActionTypeEnum = "IN_PROGRESS"
	WorkRequestResourceActionTypeCanceledCreate WorkRequestResourceActionTypeEnum = "CANCELED_CREATE"
	WorkRequestResourceActionTypeCanceledDelete WorkRequestResourceActionTypeEnum = "CANCELED_DELETE"
	WorkRequestResourceActionTypeCanceledUpdate WorkRequestResourceActionTypeEnum = "CANCELED_UPDATE"
	WorkRequestResourceActionTypeFailed         WorkRequestResourceActionTypeEnum = "FAILED"
)

var mappingWorkRequestResourceActionType = map[string]WorkRequestResourceActionTypeEnum{
	"CREATED":         WorkRequestResourceActionTypeCreated,
	"UPDATED":         WorkRequestResourceActionTypeUpdated,
	"DELETED":         WorkRequestResourceActionTypeDeleted,
	"IN_PROGRESS":     WorkRequestResourceActionTypeInProgress,
	"CANCELED_CREATE": WorkRequestResourceActionTypeCanceledCreate,
	"CANCELED_DELETE": WorkRequestResourceActionTypeCanceledDelete,
	"CANCELED_UPDATE": WorkRequestResourceActionTypeCanceledUpdate,
	"FAILED":          WorkRequestResourceActionTypeFailed,
}

// GetWorkRequestResourceActionTypeEnumValues Enumerates the set of values for WorkRequestResourceActionTypeEnum
func GetWorkRequestResourceActionTypeEnumValues() []WorkRequestResourceActionTypeEnum {
	values := make([]WorkRequestResourceActionTypeEnum, 0)
	for _, v := range mappingWorkRequestResourceActionType {
		values = append(values, v)
	}
	return values
}
