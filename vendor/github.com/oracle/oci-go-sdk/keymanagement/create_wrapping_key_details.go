// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Key Management Service API
//
// API for managing and performing operations with keys and vaults.
//

package keymanagement

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateWrappingKeyDetails The representation of CreateWrappingKeyDetails
type CreateWrappingKeyDetails struct {

	// The OCID of the compartment that contains this key.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`
}

func (m CreateWrappingKeyDetails) String() string {
	return common.PointerString(m)
}
