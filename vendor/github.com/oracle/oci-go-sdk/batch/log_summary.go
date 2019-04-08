// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Oracle Batch Service
//
// This is a Oracle Batch Service. You can find out more about at
// wiki (https://confluence.oraclecorp.com/confluence/display/C9QA/OCI+Batch+Service+-+Core+Functionality+Technical+Design+and+Explanation+for+Phase+I).
//

package batch

import (
	"github.com/oracle/oci-go-sdk/common"
)

// LogSummary Log file for a job's execution.
type LogSummary struct {

	// Log id.
	Id *string `mandatory:"false" json:"id"`

	// URL for the log content.
	JobLogUrl *string `mandatory:"false" json:"jobLogUrl"`
}

func (m LogSummary) String() string {
	return common.PointerString(m)
}
