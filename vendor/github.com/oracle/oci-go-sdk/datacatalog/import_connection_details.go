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

// ImportConnectionDetails Import Connection from the connection metadata and oracle wallet file
type ImportConnectionDetails struct {

	// The information used to import the connection
	ConnectionPayload []byte `mandatory:"true" json:"connectionPayload"`

	ConnectionDetail *CreateConnectionDetails `mandatory:"false" json:"connectionDetail"`
}

func (m ImportConnectionDetails) String() string {
	return common.PointerString(m)
}
