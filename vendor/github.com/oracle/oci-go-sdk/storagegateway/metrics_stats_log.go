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

// MetricsStatsLog Log storage statistics.
type MetricsStatsLog struct {

	// Log storage utilization (percent).
	// Example: `65`
	UtilPercent *float32 `mandatory:"false" json:"utilPercent"`
}

func (m MetricsStatsLog) String() string {
	return common.PointerString(m)
}
