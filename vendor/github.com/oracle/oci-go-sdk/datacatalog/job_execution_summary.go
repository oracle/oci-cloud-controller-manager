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

// JobExecutionSummary A list of Job Executions. A Job Execution is a unit of work being executed on behalf of a Job.
type JobExecutionSummary struct {

	// Unique key of the Job Execution resource.
	Key *string `mandatory:"true" json:"key"`

	// The unique key of the parent Job.
	JobKey *string `mandatory:"false" json:"jobKey"`

	// Type of the Job Execution.
	JobType JobTypeEnum `mandatory:"false" json:"jobType,omitempty"`

	// The unique key of the parent execution or null if this Job Execution has no parent.
	ParentKey *string `mandatory:"false" json:"parentKey"`

	// The unique key of the triggering external scheduler resource or null if this Job Execution is not externally triggered.
	ScheduleInstanceKey *string `mandatory:"false" json:"scheduleInstanceKey"`

	// Status of the Job Execution. For eg: Running, Paused, Completed etc
	LifecycleState JobExecutionStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the JobExecution was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Time that Job Execution started. An RFC3339 formatted datetime string.
	TimeStarted *common.SDKTime `mandatory:"false" json:"timeStarted"`

	// Time that the Job Execution ended or null if it hasn't yet completed.
	// An RFC3339 formatted datetime string.
	TimeEnded *common.SDKTime `mandatory:"false" json:"timeEnded"`

	// URI to the Job Execution instance in the API.
	Uri *string `mandatory:"false" json:"uri"`
}

func (m JobExecutionSummary) String() string {
	return common.PointerString(m)
}
