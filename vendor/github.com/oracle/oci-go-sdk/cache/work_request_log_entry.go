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

// WorkRequestLogEntry A log entry of a work request.
type WorkRequestLogEntry struct {

	// The log message.
	Message *string `mandatory:"true" json:"message"`

	// The time the log message was written.
	TimeStamp *common.SDKTime `mandatory:"true" json:"timeStamp"`
}

func (m WorkRequestLogEntry) String() string {
	return common.PointerString(m)
}
