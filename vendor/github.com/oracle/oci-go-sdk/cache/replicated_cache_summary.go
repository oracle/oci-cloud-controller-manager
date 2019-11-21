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

// ReplicatedCacheSummary Summary information of the Redis replicated cache.
type ReplicatedCacheSummary struct {

	// The OCID of the Redis replicated cache.
	Id *string `mandatory:"true" json:"id"`

	// The name of the Redis replicated cache.
	Name *string `mandatory:"true" json:"name"`

	// The number of replicas that make up the Redis replicated cache.
	ReplicaCount *int `mandatory:"true" json:"replicaCount"`

	// The `lifecycleState` of the Redis replicated cache.
	LifecycleState LifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The URI to access detailed information about the Redis replicated cache.
	ResourceUri *string `mandatory:"true" json:"resourceUri"`

	// The amount of memory allocated to the Redis replicated cache.
	Shape *string `mandatory:"true" json:"shape"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"true" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"true" json:"freeformTags"`

	// A brief description of the Redis replicated cache.
	Description *string `mandatory:"false" json:"description"`
}

func (m ReplicatedCacheSummary) String() string {
	return common.PointerString(m)
}

// ReplicatedCacheSummaryLifecycleStateEnum is an alias to type: LifecycleStateEnum
// Consider using LifecycleStateEnum instead
// Deprecated
type ReplicatedCacheSummaryLifecycleStateEnum = LifecycleStateEnum

// Set of constants representing the allowable values for LifecycleStateEnum
// Deprecated
const (
	ReplicatedCacheSummaryLifecycleStateCreating LifecycleStateEnum = "CREATING"
	ReplicatedCacheSummaryLifecycleStateActive   LifecycleStateEnum = "ACTIVE"
	ReplicatedCacheSummaryLifecycleStateUpdating LifecycleStateEnum = "UPDATING"
	ReplicatedCacheSummaryLifecycleStateDeleting LifecycleStateEnum = "DELETING"
	ReplicatedCacheSummaryLifecycleStateDeleted  LifecycleStateEnum = "DELETED"
	ReplicatedCacheSummaryLifecycleStateFailed   LifecycleStateEnum = "FAILED"
)

// GetReplicatedCacheSummaryLifecycleStateEnumValues Enumerates the set of values for LifecycleStateEnum
// Consider using GetLifecycleStateEnumValue
// Deprecated
var GetReplicatedCacheSummaryLifecycleStateEnumValues = GetLifecycleStateEnumValues
