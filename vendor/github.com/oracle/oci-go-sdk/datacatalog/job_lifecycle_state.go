// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DataCatalog API
//
// A description of the DataCatalog API
//

package datacatalog

// JobLifecycleStateEnum Enum with underlying type: string
type JobLifecycleStateEnum string

// Set of constants representing the allowable values for JobLifecycleStateEnum
const (
	JobLifecycleStateActive   JobLifecycleStateEnum = "ACTIVE"
	JobLifecycleStateInactive JobLifecycleStateEnum = "INACTIVE"
	JobLifecycleStateExpired  JobLifecycleStateEnum = "EXPIRED"
)

var mappingJobLifecycleState = map[string]JobLifecycleStateEnum{
	"ACTIVE":   JobLifecycleStateActive,
	"INACTIVE": JobLifecycleStateInactive,
	"EXPIRED":  JobLifecycleStateExpired,
}

// GetJobLifecycleStateEnumValues Enumerates the set of values for JobLifecycleStateEnum
func GetJobLifecycleStateEnumValues() []JobLifecycleStateEnum {
	values := make([]JobLifecycleStateEnum, 0)
	for _, v := range mappingJobLifecycleState {
		values = append(values, v)
	}
	return values
}
