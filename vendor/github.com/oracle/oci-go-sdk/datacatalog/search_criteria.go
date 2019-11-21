// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DataCatalog API
//
// A description of the DataCatalog API
//

package datacatalog

import (
	"github.com/oracle/oci-go-sdk/common"
)

// SearchCriteria Search Query object that allows complex search predicates that cannot be expressed through simple query params.
type SearchCriteria struct {

	// The set of attributes that is part of the search query criteria. Used along with Search Criteria query parameter.
	Fields []string `mandatory:"false" json:"fields"`

	// Search query dsl that defines the query components including fields and predicates.
	Query *string `mandatory:"false" json:"query"`
}

func (m SearchCriteria) String() string {
	return common.PointerString(m)
}
