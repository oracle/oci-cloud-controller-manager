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

// SearchResultSummary Summary of results.
type SearchResultSummary struct {

	// Total number of search results.
	ResultCount *int `mandatory:"false" json:"resultCount"`

	// Total number of field schema information.
	FieldCount *int `mandatory:"false" json:"fieldCount"`
}

func (m SearchResultSummary) String() string {
	return common.PointerString(m)
}
