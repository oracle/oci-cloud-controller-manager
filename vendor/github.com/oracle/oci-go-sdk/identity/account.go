// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Identity and Access Management Service API
//
// APIs for managing users, groups, compartments, and policies.
//

package identity

import (
	"github.com/oracle/oci-go-sdk/common"
)

// Account The basic information for an account object.
type Account struct {

	// The name of the tenant.
	AccountName *string `mandatory:"true" json:"accountName"`

	// The lifecycle state of the tenant.
	State *string `mandatory:"true" json:"state"`

	// The tenant home region.
	HomeRegion *string `mandatory:"true" json:"homeRegion"`

	// The tenant active regions.
	ActiveRegions []string `mandatory:"true" json:"activeRegions"`

	// The OCID of the tenancy.
	TenancyOcid *string `mandatory:"false" json:"tenancyOcid"`
}

func (m Account) String() string {
	return common.PointerString(m)
}
