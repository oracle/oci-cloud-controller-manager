// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// UsageApi API
//
// A description of the UsageApi API.
//

package usage

import (
	"github.com/oracle/oci-go-sdk/common"
)

// UsageRecord A record of a specific range of usage, including cost for the specified resource type for the specified time interval.
type UsageRecord struct {

	// The start time (UTC) of the interval for which this usage record belongs (inclusive).
	StartTime *common.SDKTime `mandatory:"false" json:"startTime"`

	// The end time (UTC) of the interval for which this usage record belongs (exclusive).
	EndTime *common.SDKTime `mandatory:"false" json:"endTime"`

	// The name of the internal resource type.
	InternalResourceName *string `mandatory:"false" json:"internalResourceName"`

	// The human-readable, friendly name of the resource type.
	DisplayResourceName *string `mandatory:"false" json:"displayResourceName"`

	// The unit type of the resource that is being measured.
	DisplayUnitName *string `mandatory:"false" json:"displayUnitName"`

	// The human-readable, friendly name of the service.
	ServiceName *string `mandatory:"false" json:"serviceName"`

	// The currency of the cost value, in the format specified by ISO-4217 (https://www.iso.org/iso-4217-currency-codes.html).
	Currency *string `mandatory:"false" json:"currency"`

	// The product identifier, or skuId.
	GsiProductId *string `mandatory:"false" json:"gsiProductId"`

	// A list of the actual costs and usage amounts for the target resource type within the
	// defined time interval (startTime to endTime).
	Costs []CostRecord `mandatory:"false" json:"costs"`
}

func (m UsageRecord) String() string {
	return common.PointerString(m)
}
