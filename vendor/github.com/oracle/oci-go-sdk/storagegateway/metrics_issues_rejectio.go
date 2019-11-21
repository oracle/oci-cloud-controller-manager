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

// MetricsIssuesRejectio IO Rejection information.
type MetricsIssuesRejectio struct {

	// Indicates whether the storage gateway is in protection mode and rejecting input and output (IO). Protection
	// mode allows no writes to the local file system cache.
	// This value is zero `0` unless the system is in protection mode. A non-zero IO rejection count means that
	// protection mode applies to all file systems in the storage gateway.
	// Example: `0`
	Count *float32 `mandatory:"false" json:"count"`
}

func (m MetricsIssuesRejectio) String() string {
	return common.PointerString(m)
}
