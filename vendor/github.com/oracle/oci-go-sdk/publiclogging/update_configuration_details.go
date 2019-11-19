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

// UpdateConfigurationDetails The updateable configuration properties
type UpdateConfigurationDetails struct {
	Source *SourceUpdateDetails `mandatory:"true" json:"source"`

	Indexing *Indexing `mandatory:"false" json:"indexing"`
}

func (m UpdateConfigurationDetails) String() string {
	return common.PointerString(m)
}
