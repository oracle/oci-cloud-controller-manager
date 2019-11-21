// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// OraCache Public API
//
// API for the Data Caching Service. Use this service to manage Redis replicated caches.
//

package cache

// ReplicatedCacheSortFieldEnum Enum with underlying type: string
type ReplicatedCacheSortFieldEnum string

// Set of constants representing the allowable values for ReplicatedCacheSortFieldEnum
const (
	ReplicatedCacheSortFieldName           ReplicatedCacheSortFieldEnum = "NAME"
	ReplicatedCacheSortFieldTimeCreated    ReplicatedCacheSortFieldEnum = "TIME_CREATED"
	ReplicatedCacheSortFieldLifecycleState ReplicatedCacheSortFieldEnum = "LIFECYCLE_STATE"
)

var mappingReplicatedCacheSortField = map[string]ReplicatedCacheSortFieldEnum{
	"NAME":            ReplicatedCacheSortFieldName,
	"TIME_CREATED":    ReplicatedCacheSortFieldTimeCreated,
	"LIFECYCLE_STATE": ReplicatedCacheSortFieldLifecycleState,
}

// GetReplicatedCacheSortFieldEnumValues Enumerates the set of values for ReplicatedCacheSortFieldEnum
func GetReplicatedCacheSortFieldEnumValues() []ReplicatedCacheSortFieldEnum {
	values := make([]ReplicatedCacheSortFieldEnum, 0)
	for _, v := range mappingReplicatedCacheSortField {
		values = append(values, v)
	}
	return values
}
