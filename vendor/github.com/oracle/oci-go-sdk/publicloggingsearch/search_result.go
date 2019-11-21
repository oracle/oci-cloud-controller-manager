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

// SearchResult A log search result entry
type SearchResult struct {

	// JSON blob containing the search entry with projected fields.
	Data *interface{} `mandatory:"true" json:"data"`
}

func (m SearchResult) String() string {
	return common.PointerString(m)
}
