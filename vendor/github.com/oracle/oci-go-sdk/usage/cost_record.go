// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// UsageApi API
//
// A description of the UsageApi API.
//

package usage

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CostRecord Object describing the cost and the usage for a specific resource type within the defined time interval.
type CostRecord struct {

	// The type of cost. Note that this property is not available for filtered queries (by compartmentId or cost tracking tag).
	ComputeType *string `mandatory:"false" json:"computeType"`

	// The amount of usage of the target resource (measured in {displayUnitName} units).
	ComputedQuantity *float32 `mandatory:"false" json:"computedQuantity"`

	// The cost of the target resource (measured in {currency} currency).
	ComputedAmount *float32 `mandatory:"false" json:"computedAmount"`

	// The unit price. Note that this property is not available for filtered queries (by compartmentId or cost tracking tag).
	UnitPrice *float32 `mandatory:"false" json:"unitPrice"`
}

func (m CostRecord) String() string {
	return common.PointerString(m)
}
