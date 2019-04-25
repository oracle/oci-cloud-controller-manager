// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Marketplace Service API
//
// Manage applications in Oracle Cloud Infrastructure Marketplace.
//

package marketplace

import (
	"github.com/oracle/oci-go-sdk/common"
)

// DocumentationLink A link to a documentation web resource.
type DocumentationLink struct {

	// The text describing the resource.
	Name *string `mandatory:"false" json:"name"`

	// The url of the resource.
	Url *string `mandatory:"false" json:"url"`

	// The category of the document.
	DocumentCategory *string `mandatory:"false" json:"documentCategory"`
}

func (m DocumentationLink) String() string {
	return common.PointerString(m)
}
