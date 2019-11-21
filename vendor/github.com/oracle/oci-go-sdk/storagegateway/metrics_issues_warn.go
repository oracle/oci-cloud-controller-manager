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

// MetricsIssuesWarn Warning information.
type MetricsIssuesWarn struct {

	// Total number of warnings found in the logs.
	// Example: `6`
	Count *float32 `mandatory:"false" json:"count"`
}

func (m MetricsIssuesWarn) String() string {
	return common.PointerString(m)
}
