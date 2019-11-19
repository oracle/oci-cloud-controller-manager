// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Public Logging Search API
//
// A description of the Public Logging Search API
//

package publicloggingsearch

import (
	"github.com/oracle/oci-go-sdk/common"
)

// SearchLogsDetails Search request object.
type SearchLogsDetails struct {

	// Start filter log's date and time, in the format defined by RFC3339.
	TimeStart *common.SDKTime `mandatory:"true" json:"timeStart"`

	// End filter log's date and time, in the format defined by RFC3339.
	TimeEnd *common.SDKTime `mandatory:"true" json:"timeEnd"`

	// Query corresponding to the search operation. This query is parsed and validated before execution and |
	// should follow the spec. TODO: Add link to query language specs in doc release when ready.
	SearchQuery *string `mandatory:"true" json:"searchQuery"`

	// Whether to return field schema information for the log stream specified in searchQuery.
	IsReturnFieldInfo *bool `mandatory:"false" json:"isReturnFieldInfo"`
}

func (m SearchLogsDetails) String() string {
	return common.PointerString(m)
}
