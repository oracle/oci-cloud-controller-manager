// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DataCatalog API
//
// A description of the DataCatalog API
//

package datacatalog

// JobExecutionStateEnum Enum with underlying type: string
type JobExecutionStateEnum string

// Set of constants representing the allowable values for JobExecutionStateEnum
const (
	JobExecutionStateCreated    JobExecutionStateEnum = "CREATED"
	JobExecutionStateInProgress JobExecutionStateEnum = "IN_PROGRESS"
	JobExecutionStateInactive   JobExecutionStateEnum = "INACTIVE"
	JobExecutionStateFailed     JobExecutionStateEnum = "FAILED"
	JobExecutionStateSucceeded  JobExecutionStateEnum = "SUCCEEDED"
	JobExecutionStateCanceled   JobExecutionStateEnum = "CANCELED"
)

var mappingJobExecutionState = map[string]JobExecutionStateEnum{
	"CREATED":     JobExecutionStateCreated,
	"IN_PROGRESS": JobExecutionStateInProgress,
	"INACTIVE":    JobExecutionStateInactive,
	"FAILED":      JobExecutionStateFailed,
	"SUCCEEDED":   JobExecutionStateSucceeded,
	"CANCELED":    JobExecutionStateCanceled,
}

// GetJobExecutionStateEnumValues Enumerates the set of values for JobExecutionStateEnum
func GetJobExecutionStateEnumValues() []JobExecutionStateEnum {
	values := make([]JobExecutionStateEnum, 0)
	for _, v := range mappingJobExecutionState {
		values = append(values, v)
	}
	return values
}
