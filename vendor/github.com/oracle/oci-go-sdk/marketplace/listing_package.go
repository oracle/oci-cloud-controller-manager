// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Marketplace Service API
//
// Manage applications in Oracle Cloud Infrastructure Marketplace.
//

package marketplace

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// ListingPackage A base object for all types of listing packages.
type ListingPackage interface {

	// The OCID of the listing this package belongs to.
	GetListingId() *string

	// The version of this package.
	GetVersion() *string

	// Description of this package.
	GetDescription() *string

	// The unique identifier of the package resource.
	GetResourceId() *string

	// The date and time this listing package was created, in the format defined by RFC3339.
	// Example: `2016-08-25T21:10:29.600Z`
	GetTimeCreated() *common.SDKTime
}

type listingpackage struct {
	JsonData    []byte
	ListingId   *string         `mandatory:"true" json:"listingId"`
	Version     *string         `mandatory:"true" json:"version"`
	Description *string         `mandatory:"false" json:"description"`
	ResourceId  *string         `mandatory:"false" json:"resourceId"`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
	PackageType string          `json:"packageType"`
}

// UnmarshalJSON unmarshals json
func (m *listingpackage) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerlistingpackage listingpackage
	s := struct {
		Model Unmarshalerlistingpackage
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.ListingId = s.Model.ListingId
	m.Version = s.Model.Version
	m.Description = s.Model.Description
	m.ResourceId = s.Model.ResourceId
	m.TimeCreated = s.Model.TimeCreated
	m.PackageType = s.Model.PackageType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *listingpackage) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.PackageType {
	case "Image":
		mm := ImageListingPackage{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

//GetListingId returns ListingId
func (m listingpackage) GetListingId() *string {
	return m.ListingId
}

//GetVersion returns Version
func (m listingpackage) GetVersion() *string {
	return m.Version
}

//GetDescription returns Description
func (m listingpackage) GetDescription() *string {
	return m.Description
}

//GetResourceId returns ResourceId
func (m listingpackage) GetResourceId() *string {
	return m.ResourceId
}

//GetTimeCreated returns TimeCreated
func (m listingpackage) GetTimeCreated() *common.SDKTime {
	return m.TimeCreated
}

func (m listingpackage) String() string {
	return common.PointerString(m)
}
