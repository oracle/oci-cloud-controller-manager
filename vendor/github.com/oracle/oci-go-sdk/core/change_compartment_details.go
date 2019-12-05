// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"github.com/oracle/oci-go-sdk/common"
)

// ChangeCompartmentDetails Contains details indicating which compartment the resource should move to
type ChangeCompartmentDetails struct {

	// The OCID of the new compartment
	CompartmentId *string `mandatory:"false" json:"compartmentId"`
}

func (m ChangeCompartmentDetails) String() string {
	return common.PointerString(m)
}
