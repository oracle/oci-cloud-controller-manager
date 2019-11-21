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

// MetricsResourceFilesystems File systems resource information.
type MetricsResourceFilesystems struct {

	// The maximum number of file systems that can be created per storage gateway.
	// Example: `10`
	MaxCount *float32 `mandatory:"false" json:"maxCount"`
}

func (m MetricsResourceFilesystems) String() string {
	return common.PointerString(m)
}
