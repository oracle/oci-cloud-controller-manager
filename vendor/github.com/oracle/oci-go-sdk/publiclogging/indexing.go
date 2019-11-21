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

// Indexing Log indexing configuration.
type Indexing struct {

	// True if indexing enabled.
	IsEnabled *bool `mandatory:"false" json:"isEnabled"`
}

func (m Indexing) String() string {
	return common.PointerString(m)
}
