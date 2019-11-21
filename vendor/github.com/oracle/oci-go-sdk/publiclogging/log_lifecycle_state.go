// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// PublicLoggingControlplane API
//
// PublicLoggingControlplane API specification
//

package publiclogging

// LogLifecycleStateEnum Enum with underlying type: string
type LogLifecycleStateEnum string

// Set of constants representing the allowable values for LogLifecycleStateEnum
const (
	LogLifecycleStateCreating LogLifecycleStateEnum = "CREATING"
	LogLifecycleStateActive   LogLifecycleStateEnum = "ACTIVE"
	LogLifecycleStateUpdating LogLifecycleStateEnum = "UPDATING"
	LogLifecycleStateInactive LogLifecycleStateEnum = "INACTIVE"
	LogLifecycleStateDeleting LogLifecycleStateEnum = "DELETING"
)

var mappingLogLifecycleState = map[string]LogLifecycleStateEnum{
	"CREATING": LogLifecycleStateCreating,
	"ACTIVE":   LogLifecycleStateActive,
	"UPDATING": LogLifecycleStateUpdating,
	"INACTIVE": LogLifecycleStateInactive,
	"DELETING": LogLifecycleStateDeleting,
}

// GetLogLifecycleStateEnumValues Enumerates the set of values for LogLifecycleStateEnum
func GetLogLifecycleStateEnumValues() []LogLifecycleStateEnum {
	values := make([]LogLifecycleStateEnum, 0)
	for _, v := range mappingLogLifecycleState {
		values = append(values, v)
	}
	return values
}
