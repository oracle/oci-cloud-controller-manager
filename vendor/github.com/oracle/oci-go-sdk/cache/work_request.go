// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// OraCache Public API
//
// API for the Data Caching Service. Use this service to manage Redis replicated caches.
//

package cache

import (
	"github.com/oracle/oci-go-sdk/common"
)

// WorkRequest Details of a work request. The Data Cache service initiates and manages service operations in the context of a work request.
type WorkRequest struct {

	// The type of operation that is currently being performed.
	OperationType *string `mandatory:"true" json:"operationType"`

	// The current status of the work request.
	Status WorkRequestStatusEnum `mandatory:"true" json:"status"`

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

	// The time the work request reached a terminal status - `SUCCEEDED`, `CANCELED` or `FAILED`.
	TimeFinished *common.SDKTime `mandatory:"false" json:"timeFinished"`
}

func (m WorkRequest) String() string {
	return common.PointerString(m)
}
