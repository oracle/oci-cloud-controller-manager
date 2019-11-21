// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// OraCache Public API
//
// API for the Data Caching Service. Use this service to manage Redis replicated caches.
//

package cache

// WorkRequestSortFieldEnum Enum with underlying type: string
type WorkRequestSortFieldEnum string

// Set of constants representing the allowable values for WorkRequestSortFieldEnum
const (
	WorkRequestSortFieldTimeAccepted WorkRequestSortFieldEnum = "TIME_ACCEPTED"
	WorkRequestSortFieldTimeStarted  WorkRequestSortFieldEnum = "TIME_STARTED"
	WorkRequestSortFieldTimeFinished WorkRequestSortFieldEnum = "TIME_FINISHED"
	WorkRequestSortFieldStatus       WorkRequestSortFieldEnum = "STATUS"
)

var mappingWorkRequestSortField = map[string]WorkRequestSortFieldEnum{
	"TIME_ACCEPTED": WorkRequestSortFieldTimeAccepted,
	"TIME_STARTED":  WorkRequestSortFieldTimeStarted,
	"TIME_FINISHED": WorkRequestSortFieldTimeFinished,
	"STATUS":        WorkRequestSortFieldStatus,
}

// GetWorkRequestSortFieldEnumValues Enumerates the set of values for WorkRequestSortFieldEnum
func GetWorkRequestSortFieldEnumValues() []WorkRequestSortFieldEnum {
	values := make([]WorkRequestSortFieldEnum, 0)
	for _, v := range mappingWorkRequestSortField {
		values = append(values, v)
	}
	return values
}
