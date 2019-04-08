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

// WorkRequestError An error that occurred while executing a work request.
type WorkRequestError struct {

	// The code of the error that occurred.
	Code *string `mandatory:"true" json:"code"`

	// The log message.
	Message *string `mandatory:"true" json:"message"`

	// The time the log message was written.
	TimeStamp *common.SDKTime `mandatory:"true" json:"timeStamp"`
}

func (m WorkRequestError) String() string {
	return common.PointerString(m)
}
