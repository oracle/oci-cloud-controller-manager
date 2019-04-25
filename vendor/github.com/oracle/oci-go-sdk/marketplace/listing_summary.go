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

// ListingSummary The model for a summary of an Oracle Cloud Infrastructure Marketplace listing.
type ListingSummary struct {

	// The unique identifier for the listing in the Oracle Cloud Infrastructure Marketplace.
	Id *string `mandatory:"false" json:"id"`

	// The name of the listing.
	Name *string `mandatory:"false" json:"name"`

	// Short description of the listing.
	ShortDescription *string `mandatory:"false" json:"shortDescription"`

	// The tagline of the listing.
	Tagline *string `mandatory:"false" json:"tagline"`

	// The URL of the listing icon.
	Icon *UploadData `mandatory:"false" json:"icon"`
}

func (m ListingSummary) String() string {
	return common.PointerString(m)
}
