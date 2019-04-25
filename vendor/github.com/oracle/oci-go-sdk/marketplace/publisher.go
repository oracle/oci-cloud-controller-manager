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

// Publisher The model for a publisher.
type Publisher struct {

	// Unique identifier for the publisher.
	Id *string `mandatory:"false" json:"id"`

	// The name of the publisher.
	Name *string `mandatory:"false" json:"name"`

	// A description of the publisher.
	Description *string `mandatory:"false" json:"description"`

	// The year the publisher was founded.
	YearFounded *int64 `mandatory:"false" json:"yearFounded"`

	// The publisher's website.
	WebsiteUrl *string `mandatory:"false" json:"websiteUrl"`

	// The email address of the publisher.
	ContactEmail *string `mandatory:"false" json:"contactEmail"`

	// The phone number of the publisher.
	ContactPhone *string `mandatory:"false" json:"contactPhone"`

	// The address of the publisher's headquarters.
	HqAddress *string `mandatory:"false" json:"hqAddress"`

	// The URL of the publisher's logo.
	Logo *UploadData `mandatory:"false" json:"logo"`

	// Reference links.
	Links []Link `mandatory:"false" json:"links"`
}

func (m Publisher) String() string {
	return common.PointerString(m)
}
