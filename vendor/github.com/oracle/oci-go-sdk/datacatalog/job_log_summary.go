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

// JobLogSummary A List of Job Execution Logs.
// A Job Log is an audit log record inserted during the lifecycle of a job execution instance.
// There can be one or more logs for an execution instance.
type JobLogSummary struct {

	// Unique key of the Job Log which is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The unique key of the parent Job Execution for which the log resource was created.
	JobExecutionKey *string `mandatory:"false" json:"jobExecutionKey"`

	// URI to the Job Log instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// The date and time the JobLog was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Severity Level for this Log.
	Severity *string `mandatory:"false" json:"severity"`

	// Message for this Job Log
	LogMessage *string `mandatory:"false" json:"logMessage"`
}

func (m JobLogSummary) String() string {
	return common.PointerString(m)
}
