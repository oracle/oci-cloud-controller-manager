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

// SearchContext The search context, which contains information like highlights for matching resource.
type SearchContext struct {

	// Provides highlighted values based on which fields matched the search criteria. The key is the field name
	// (provided as a path into the resource).  The value is a list of strings that represent fragments of the field value that matched.
	// There may be more than one fragment per field if multiple sections of the value matched the search.  The matching parts of each
	// fragment are wrapped with <hl>..</hl> (highlight) tags. All values are HTML encoded (except <hl> tags). This works only when
	// FreeTextSearchDetails is used, or if the query in a StructuredSearchDetails contains a 'matching' clause.
	Highlights map[string][]string `mandatory:"false" json:"highlights"`
}

func (m SearchContext) String() string {
	return common.PointerString(m)
}
