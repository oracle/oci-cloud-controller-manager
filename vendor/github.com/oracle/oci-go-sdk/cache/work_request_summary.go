// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// OraCache Public API
//
// API for the Data Caching Service. Use this service to manage Redis replicated caches.
//

package cache

import (
	"github.com/oracle/oci-go-sdk/common"
)

// WorkRequestSummary A summary of a work request.
type WorkRequestSummary struct {

	// The type of operation that is being perfomed by the work request.
	OperationType *string `mandatory:"true" json:"operationType"`

	// The current status of the work request.
	Status WorkRequestSummaryStatusEnum `mandatory:"true" json:"status"`

	// The OCID of the work request.
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the compartment that initiated the work request.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The list of resources the work request affects.
	Resources []WorkRequestResource `mandatory:"true" json:"resources"`

	// The current progress of the work request.
	PercentComplete *float32 `mandatory:"true" json:"percentComplete"`

	// The time the work request was created.
	TimeAccepted *common.SDKTime `mandatory:"true" json:"timeAccepted"`

	// The time the work request was moved from `ACCEPTED` status to `IN_PROGRESS` status.
	TimeStarted *common.SDKTime `mandatory:"false" json:"timeStarted"`

	// The time this work request reached a terminal status - `SUCCEEDED`, `CANCELED` or `FAILED`.
	TimeFinished *common.SDKTime `mandatory:"false" json:"timeFinished"`
}

func (m WorkRequestSummary) String() string {
	return common.PointerString(m)
}

// WorkRequestSummaryStatusEnum Enum with underlying type: string
type WorkRequestSummaryStatusEnum string

// Set of constants representing the allowable values for WorkRequestSummaryStatusEnum
const (
	WorkRequestSummaryStatusCreated        WorkRequestSummaryStatusEnum = "CREATED"
	WorkRequestSummaryStatusUpdated        WorkRequestSummaryStatusEnum = "UPDATED"
	WorkRequestSummaryStatusDeleted        WorkRequestSummaryStatusEnum = "DELETED"
	WorkRequestSummaryStatusInProgress     WorkRequestSummaryStatusEnum = "IN_PROGRESS"
	WorkRequestSummaryStatusCanceledCreate WorkRequestSummaryStatusEnum = "CANCELED_CREATE"
	WorkRequestSummaryStatusCanceledDelete WorkRequestSummaryStatusEnum = "CANCELED_DELETE"
	WorkRequestSummaryStatusCanceledUpdate WorkRequestSummaryStatusEnum = "CANCELED_UPDATE"
	WorkRequestSummaryStatusFailed         WorkRequestSummaryStatusEnum = "FAILED"
)

var mappingWorkRequestSummaryStatus = map[string]WorkRequestSummaryStatusEnum{
	"CREATED":         WorkRequestSummaryStatusCreated,
	"UPDATED":         WorkRequestSummaryStatusUpdated,
	"DELETED":         WorkRequestSummaryStatusDeleted,
	"IN_PROGRESS":     WorkRequestSummaryStatusInProgress,
	"CANCELED_CREATE": WorkRequestSummaryStatusCanceledCreate,
	"CANCELED_DELETE": WorkRequestSummaryStatusCanceledDelete,
	"CANCELED_UPDATE": WorkRequestSummaryStatusCanceledUpdate,
	"FAILED":          WorkRequestSummaryStatusFailed,
}

// GetWorkRequestSummaryStatusEnumValues Enumerates the set of values for WorkRequestSummaryStatusEnum
func GetWorkRequestSummaryStatusEnumValues() []WorkRequestSummaryStatusEnum {
	values := make([]WorkRequestSummaryStatusEnum, 0)
	for _, v := range mappingWorkRequestSummaryStatus {
		values = append(values, v)
	}
	return values
}
