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

// ChangeLogGroupCompartmentDetails Contains details indicating which compartment the resource should move to
type ChangeLogGroupCompartmentDetails struct {

	// The of the compartment into which the resource should be moved.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`
}

func (m ChangeLogGroupCompartmentDetails) String() string {
	return common.PointerString(m)
}
