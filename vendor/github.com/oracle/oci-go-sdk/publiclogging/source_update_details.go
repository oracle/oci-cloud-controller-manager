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

// SourceUpdateDetails Source updated configuration.
type SourceUpdateDetails struct {

	// Log category parameters are stored here.
	Parameters map[string]string `mandatory:"false" json:"parameters"`
}

func (m SourceUpdateDetails) String() string {
	return common.PointerString(m)
}
