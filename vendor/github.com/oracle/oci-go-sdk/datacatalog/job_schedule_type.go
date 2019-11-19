// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DataCatalog API
//
// A description of the DataCatalog API
//

package datacatalog

// JobScheduleTypeEnum Enum with underlying type: string
type JobScheduleTypeEnum string

// Set of constants representing the allowable values for JobScheduleTypeEnum
const (
	JobScheduleTypeScheduled JobScheduleTypeEnum = "SCHEDULED"
	JobScheduleTypeImmediate JobScheduleTypeEnum = "IMMEDIATE"
)

var mappingJobScheduleType = map[string]JobScheduleTypeEnum{
	"SCHEDULED": JobScheduleTypeScheduled,
	"IMMEDIATE": JobScheduleTypeImmediate,
}

// GetJobScheduleTypeEnumValues Enumerates the set of values for JobScheduleTypeEnum
func GetJobScheduleTypeEnumValues() []JobScheduleTypeEnum {
	values := make([]JobScheduleTypeEnum, 0)
	for _, v := range mappingJobScheduleType {
		values = append(values, v)
	}
	return values
}
