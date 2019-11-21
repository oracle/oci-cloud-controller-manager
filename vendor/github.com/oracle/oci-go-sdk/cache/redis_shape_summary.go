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

// RedisShapeSummary The amount of memory allocated to the Redis replicated cache.
type RedisShapeSummary struct {

	// Redis shape
	Shape *string `mandatory:"true" json:"shape"`
}

func (m RedisShapeSummary) String() string {
	return common.PointerString(m)
}
