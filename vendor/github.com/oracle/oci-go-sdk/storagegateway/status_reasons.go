// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Storage Gateway API
//
// API for the Storage Gateway service. Use this API to manage storage gateways and related items. For more
// information, see Overview of Storage Gateway (https://docs.cloud.oracle.com/iaas/Content/StorageGateway/Concepts/storagegatewayoverview.htm).
//

package storagegateway

import (
	"github.com/oracle/oci-go-sdk/common"
)

// StatusReasons The reasons for the overall health status. The object can include an array of reason strings for the 'CRITICAL'
// health status and another array of reason strings for the 'WARNING' health status. The 'CRITICAL' health status
// includes both arrays if a warning also exists.
type StatusReasons struct {

	// An array of reasons for the critical status.
	// Example: `Rejecting IO due to low cache space`
	Critical []string `mandatory:"false" json:"critical"`

	// An array of reasons for the warning status.
	// Example: `A newer version is available`
	Warning []string `mandatory:"false" json:"warning"`
}

func (m StatusReasons) String() string {
	return common.PointerString(m)
}
