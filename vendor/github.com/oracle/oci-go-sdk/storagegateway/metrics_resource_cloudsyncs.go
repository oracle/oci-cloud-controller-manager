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

// MetricsResourceCloudsyncs Cloud syncs resource information.
type MetricsResourceCloudsyncs struct {

	// The maximum number of cloud syncs that can be created per storage gateway.
	// Example: `20`
	MaxCount *float32 `mandatory:"false" json:"maxCount"`
}

func (m MetricsResourceCloudsyncs) String() string {
	return common.PointerString(m)
}
