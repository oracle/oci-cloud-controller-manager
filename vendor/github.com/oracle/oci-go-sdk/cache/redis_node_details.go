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

// RedisNodeDetails The Redis nodes that host the Redis servers. The nodes are created in the specified availability domain.
type RedisNodeDetails struct {

	// The name of the availability domain where the Redis node should be located.
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The subnet id to which this Redis node is attached.
	SubnetId *string `mandatory:"true" json:"subnetId"`

	// Whether this node should be the primary Redis node. The default value is `false`.
	IsPrimary *bool `mandatory:"false" json:"isPrimary"`
}

func (m RedisNodeDetails) String() string {
	return common.PointerString(m)
}
