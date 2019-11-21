// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// OraCache Public API
//
// API for the Data Caching Service. Use this service to manage Redis replicated caches.
//

package cache

// ResourceActionTypeEnum Enum with underlying type: string
type ResourceActionTypeEnum string

// Set of constants representing the allowable values for ResourceActionTypeEnum
const (
	ResourceActionTypeCreated        ResourceActionTypeEnum = "CREATED"
	ResourceActionTypeUpdated        ResourceActionTypeEnum = "UPDATED"
	ResourceActionTypeDeleted        ResourceActionTypeEnum = "DELETED"
	ResourceActionTypeInProgress     ResourceActionTypeEnum = "IN_PROGRESS"
	ResourceActionTypeCanceledCreate ResourceActionTypeEnum = "CANCELED_CREATE"
	ResourceActionTypeCanceledDelete ResourceActionTypeEnum = "CANCELED_DELETE"
	ResourceActionTypeCanceledUpdate ResourceActionTypeEnum = "CANCELED_UPDATE"
	ResourceActionTypeFailed         ResourceActionTypeEnum = "FAILED"
)

var mappingResourceActionType = map[string]ResourceActionTypeEnum{
	"CREATED":         ResourceActionTypeCreated,
	"UPDATED":         ResourceActionTypeUpdated,
	"DELETED":         ResourceActionTypeDeleted,
	"IN_PROGRESS":     ResourceActionTypeInProgress,
	"CANCELED_CREATE": ResourceActionTypeCanceledCreate,
	"CANCELED_DELETE": ResourceActionTypeCanceledDelete,
	"CANCELED_UPDATE": ResourceActionTypeCanceledUpdate,
	"FAILED":          ResourceActionTypeFailed,
}

// GetResourceActionTypeEnumValues Enumerates the set of values for ResourceActionTypeEnum
func GetResourceActionTypeEnumValues() []ResourceActionTypeEnum {
	values := make([]ResourceActionTypeEnum, 0)
	for _, v := range mappingResourceActionType {
		values = append(values, v)
	}
	return values
}
