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

// ReplicatedCache Details of the Redis replicated cache. Redis replicated caches are comprised of Oracle-managed Redis nodes that each contain a replica of the cached data. The cache is accessible from a tenant's compartment using a published endpoint.
type ReplicatedCache struct {

	// The OCID of the Redis replicated cache.
	Id *string `mandatory:"true" json:"id"`

	// The compartment OCID from which the Redis replicated cache is accessible.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The VCN OCID that contains the network resources and subnets to which the Redis nodes are attached.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// The name of the Redis replicated cache
	Name *string `mandatory:"true" json:"name"`

	// The number of replica nodes that make up the Redis replicated cache.
	ReplicaCount *int `mandatory:"true" json:"replicaCount"`

	// The amount of memory allocated to the Redis replicated cache.
	Shape *string `mandatory:"true" json:"shape"`

	// The endpoints of the replicas that make up the Redis replicated cache.
	RedisNodes []EndPoint `mandatory:"true" json:"redisNodes"`

	// Cache creation timestamp. Format defined by RFC3339.
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The `lifecycleState` of the Redis replicated cache.
	LifecycleState LifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// A brief description of the Redis replicated cache
	Description *string `mandatory:"false" json:"description"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m ReplicatedCache) String() string {
	return common.PointerString(m)
}

// ReplicatedCacheLifecycleStateEnum is an alias to type: LifecycleStateEnum
// Consider using LifecycleStateEnum instead
// Deprecated
type ReplicatedCacheLifecycleStateEnum = LifecycleStateEnum

// Set of constants representing the allowable values for LifecycleStateEnum
// Deprecated
const (
	ReplicatedCacheLifecycleStateCreating LifecycleStateEnum = "CREATING"
	ReplicatedCacheLifecycleStateActive   LifecycleStateEnum = "ACTIVE"
	ReplicatedCacheLifecycleStateUpdating LifecycleStateEnum = "UPDATING"
	ReplicatedCacheLifecycleStateDeleting LifecycleStateEnum = "DELETING"
	ReplicatedCacheLifecycleStateDeleted  LifecycleStateEnum = "DELETED"
	ReplicatedCacheLifecycleStateFailed   LifecycleStateEnum = "FAILED"
)

// GetReplicatedCacheLifecycleStateEnumValues Enumerates the set of values for LifecycleStateEnum
// Consider using GetLifecycleStateEnumValue
// Deprecated
var GetReplicatedCacheLifecycleStateEnumValues = GetLifecycleStateEnumValues
