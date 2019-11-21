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

// JobLog Job Log details. A Job Log is an audit log record inserted during the lifecycle of a job execution instance.
type JobLog struct {

	// Unique key of the Job Log which is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The unique key of the parent Job Execution for which the log resource was created.
	JobExecutionKey *string `mandatory:"false" json:"jobExecutionKey"`

	// Id (OCID) of the user who created the log record for this job. Usually the executor of the job instance.
	CreatedById *string `mandatory:"false" json:"createdById"`

	// Id (OCID) of the user who created the log record for this job. Usually the executor of the job instance.
	UpdatedById *string `mandatory:"false" json:"updatedById"`

	// Job Log update time. An RFC3339 formatted datetime string.
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// The date and time the JobLog was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Severity Level for this Log
	Severity *string `mandatory:"false" json:"severity"`

	// Message for this Job Log
	LogMessage *string `mandatory:"false" json:"logMessage"`

	// URI to the Job Log instance in the API.
	Uri *string `mandatory:"false" json:"uri"`
}

func (m JobLog) String() string {
	return common.PointerString(m)
}
