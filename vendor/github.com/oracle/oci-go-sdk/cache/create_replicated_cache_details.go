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

// CreateReplicatedCacheDetails The properties that are required to create the Redis replicated cache.
type CreateReplicatedCacheDetails struct {

	// The OCID of a compartment in the customer's tenancy. A Redis replicated cache object is created in this compartment. The object represents the logical set of Redis server instances that are deployed in an Oracle-managed tenancy. The Redis server instances are network accessible from the customer's tenancy.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. Avoid entering confidential information.
	Name *string `mandatory:"true" json:"name"`

	// The OCID of a VCN in the customer's tenancy. The VCN must be located in the specified compartment. The VCN contains the network resources and subnets that allow access to the Redis nodes that are deployed in an Oracle-managed tenancy.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// The number of Redis replication nodes.
	ReplicaCount *int `mandatory:"true" json:"replicaCount"`

	// The physical characteristics (memory, network bandwidth, OCPUs, and so on) of the virtual machine on which the Redis node runs. The shape determines the amount of memory allocated to the Redis replicated cache.
	Shape *string `mandatory:"true" json:"shape"`

	// A description of the Redis replicated cache. Avoid entering confidential information.
	Description *string `mandatory:"false" json:"description"`

	// The primary Redis node and up to 5 replication nodes. Each node hosts a Redis server instance and is associated with a specific availability domain and subnet.
	RedisNodes []RedisNodeDetails `mandatory:"false" json:"redisNodes"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m CreateReplicatedCacheDetails) String() string {
	return common.PointerString(m)
}
