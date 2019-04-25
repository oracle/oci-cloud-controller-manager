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

// Item The model for an item within an array of filter values.
type Item struct {

	// The name of the item.
	Name *string `mandatory:"false" json:"name"`

	// The code of the item.
	Code *string `mandatory:"false" json:"code"`
}

func (m Item) String() string {
	return common.PointerString(m)
}
