// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Data Safe Control Plane API
//
// The API to manage data safe instance creation and deletion
//

package datasafe

import (
	"github.com/oracle/oci-go-sdk/common"
)

// WorkRequestSummary A description of workrequest status
type WorkRequestSummary struct {

	// type of the work request
	OperationType WorkRequestSummaryOperationTypeEnum `mandatory:"true" json:"operationType"`

	// status of current work request.
	Status WorkRequestSummaryStatusEnum `mandatory:"true" json:"status"`

	// The id of the work request.
	Id *string `mandatory:"true" json:"id"`

	// The ocid of the compartment that contains the work request. Work requests should be scoped to
	// the same compartment as the resource the work request affects. If the work request affects multiple resources,
	// and those resources are not in the same compartment, it is up to the service team to pick the primary
	// resource whose compartment should be used
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The resources affected by this work request.
	Resources []WorkRequestResource `mandatory:"true" json:"resources"`

	// Percentage of the request completed.
	PercentComplete *float32 `mandatory:"true" json:"percentComplete"`

	// The date and time the request was created, as described in
	// RFC 3339 (https://tools.ietf.org/rfc/rfc3339), section 14.29.
	TimeAccepted *common.SDKTime `mandatory:"true" json:"timeAccepted"`

	// The date and time the request was started, as described in RFC 3339 (https://tools.ietf.org/rfc/rfc3339),
	// section 14.29.
	TimeStarted *common.SDKTime `mandatory:"false" json:"timeStarted"`

	// The date and time the object was finished, as described in RFC 3339 (https://tools.ietf.org/rfc/rfc3339).
	TimeFinished *common.SDKTime `mandatory:"false" json:"timeFinished"`
}

func (m WorkRequestSummary) String() string {
	return common.PointerString(m)
}

// WorkRequestSummaryOperationTypeEnum Enum with underlying type: string
type WorkRequestSummaryOperationTypeEnum string

// Set of constants representing the allowable values for WorkRequestSummaryOperationTypeEnum
const (
	WorkRequestSummaryOperationTypeCreateDataSafeInstance WorkRequestSummaryOperationTypeEnum = "CREATE_DATA_SAFE_INSTANCE"
	WorkRequestSummaryOperationTypeUpdateDataSafeInstance WorkRequestSummaryOperationTypeEnum = "UPDATE_DATA_SAFE_INSTANCE"
	WorkRequestSummaryOperationTypeDeleteDataSafeInstance WorkRequestSummaryOperationTypeEnum = "DELETE_DATA_SAFE_INSTANCE"
)

var mappingWorkRequestSummaryOperationType = map[string]WorkRequestSummaryOperationTypeEnum{
	"CREATE_DATA_SAFE_INSTANCE": WorkRequestSummaryOperationTypeCreateDataSafeInstance,
	"UPDATE_DATA_SAFE_INSTANCE": WorkRequestSummaryOperationTypeUpdateDataSafeInstance,
	"DELETE_DATA_SAFE_INSTANCE": WorkRequestSummaryOperationTypeDeleteDataSafeInstance,
}

// GetWorkRequestSummaryOperationTypeEnumValues Enumerates the set of values for WorkRequestSummaryOperationTypeEnum
func GetWorkRequestSummaryOperationTypeEnumValues() []WorkRequestSummaryOperationTypeEnum {
	values := make([]WorkRequestSummaryOperationTypeEnum, 0)
	for _, v := range mappingWorkRequestSummaryOperationType {
		values = append(values, v)
	}
	return values
}

// WorkRequestSummaryStatusEnum Enum with underlying type: string
type WorkRequestSummaryStatusEnum string

// Set of constants representing the allowable values for WorkRequestSummaryStatusEnum
const (
	WorkRequestSummaryStatusAccepted   WorkRequestSummaryStatusEnum = "ACCEPTED"
	WorkRequestSummaryStatusInProgress WorkRequestSummaryStatusEnum = "IN_PROGRESS"
	WorkRequestSummaryStatusFailed     WorkRequestSummaryStatusEnum = "FAILED"
	WorkRequestSummaryStatusSucceeded  WorkRequestSummaryStatusEnum = "SUCCEEDED"
)

var mappingWorkRequestSummaryStatus = map[string]WorkRequestSummaryStatusEnum{
	"ACCEPTED":    WorkRequestSummaryStatusAccepted,
	"IN_PROGRESS": WorkRequestSummaryStatusInProgress,
	"FAILED":      WorkRequestSummaryStatusFailed,
	"SUCCEEDED":   WorkRequestSummaryStatusSucceeded,
}

// GetWorkRequestSummaryStatusEnumValues Enumerates the set of values for WorkRequestSummaryStatusEnum
func GetWorkRequestSummaryStatusEnumValues() []WorkRequestSummaryStatusEnum {
	values := make([]WorkRequestSummaryStatusEnum, 0)
	for _, v := range mappingWorkRequestSummaryStatus {
		values = append(values, v)
	}
	return values
}
