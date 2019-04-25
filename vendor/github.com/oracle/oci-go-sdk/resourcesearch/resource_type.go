// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Resource Search Service
//
// Search for resources across your cloud infrastructure
//

package resourcesearch

import (
	"github.com/oracle/oci-go-sdk/common"
)

// ResourceType Defines a type of resource that may be queried for.
type ResourceType struct {

	// The unique name of the resource type, matches the value returned as part of the ResourceSummary object.
	Name *string `mandatory:"true" json:"name"`

	// List all of the fields that can be used for querying, along with their value type.
	Fields []QueryableFieldDescription `mandatory:"true" json:"fields"`
}

func (m ResourceType) String() string {
	return common.PointerString(m)
}
