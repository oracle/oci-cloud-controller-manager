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

// SearchTagSummary Represents the association of an object to a term. Returned as part of search result.
type SearchTagSummary struct {

	// Name of the tag which matches the term name.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// Unique tag key that is immutable.
	Key *string `mandatory:"false" json:"key"`
}

func (m SearchTagSummary) String() string {
	return common.PointerString(m)
}
