// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// PublicLoggingControlplane API
//
// PublicLoggingControlplane API specification
//

package publiclogging

import (
	"github.com/oracle/oci-go-sdk/common"
)

// Id Resource unique identifier.
type Id struct {

	// The OCID of the resource.
	Id *string `mandatory:"false" json:"id"`
}

func (m Id) String() string {
	return common.PointerString(m)
}
