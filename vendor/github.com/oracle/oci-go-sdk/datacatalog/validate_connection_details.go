// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DataCatalog API
//
// A description of the DataCatalog API
//

package datacatalog

import (
	"github.com/oracle/oci-go-sdk/common"
)

// ValidateConnectionDetails Validate Connection from the connection metadata or oracle wallet file
type ValidateConnectionDetails struct {
	ConnectionDetail *CreateConnectionDetails `mandatory:"false" json:"connectionDetail"`

	// The information used to validate the connection
	ConnectionPayload []byte `mandatory:"false" json:"connectionPayload"`
}

func (m ValidateConnectionDetails) String() string {
	return common.PointerString(m)
}
