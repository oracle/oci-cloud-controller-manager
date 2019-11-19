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

// SearchResponse Search response object.
type SearchResponse struct {
	Summary *SearchResultSummary `mandatory:"true" json:"summary"`

	// List of search results
	Results []SearchResult `mandatory:"false" json:"results"`

	// List of log field schema information.
	Fields []FieldInfo `mandatory:"false" json:"fields"`
}

func (m SearchResponse) String() string {
	return common.PointerString(m)
}
