// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Data Safe Control Plane API
//
// The API to manage data safe instance creation and deletion
//

package datasafe

import (
	"github.com/oracle/oci-go-sdk/common"
)

// ChangeCompartmentDetails The compartment to be changed to
type ChangeCompartmentDetails struct {

	// The new compartment identifier for the data safe instance
	CompartmentId *string `mandatory:"false" json:"compartmentId"`
}

func (m ChangeCompartmentDetails) String() string {
	return common.PointerString(m)
}
