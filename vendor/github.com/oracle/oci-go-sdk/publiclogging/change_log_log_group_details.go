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

// ChangeLogLogGroupDetails Contains details indicating which log group the log should move to
type ChangeLogLogGroupDetails struct {

	// Log group OCID.
	TargetLogGroupId *string `mandatory:"false" json:"targetLogGroupId"`
}

func (m ChangeLogLogGroupDetails) String() string {
	return common.PointerString(m)
}
