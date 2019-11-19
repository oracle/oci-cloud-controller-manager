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

// MetricsStatsCloudsyncs Cloud syncs statistics.
type MetricsStatsCloudsyncs struct {

	// Total number of cloud syncs.
	// Example: `7`
	Count *float32 `mandatory:"false" json:"count"`

	// Number of running cloud syncs.
	// Example: `1`
	RunningCount *float32 `mandatory:"false" json:"runningCount"`

	// Number of created cloud syncs.
	// Example: `2`
	CreatedCount *float32 `mandatory:"false" json:"createdCount"`

	// Number of completed cloud syncs.
	// Example: `3`
	CompletedCount *float32 `mandatory:"false" json:"completedCount"`

	// Number of canceled cloud syncs.
	// Example: `1`
	CanceledCount *float32 `mandatory:"false" json:"canceledCount"`

	// Number of failed cloud syncs.
	// Example: `0`
	FailedCount *float32 `mandatory:"false" json:"failedCount"`

	// Number of upload cloud syncs.
	// Example: `4`
	UploadCount *float32 `mandatory:"false" json:"uploadCount"`

	// Number of download cloud syncs.
	// Example: `3`
	DownloadCount *float32 `mandatory:"false" json:"downloadCount"`
}

func (m MetricsStatsCloudsyncs) String() string {
	return common.PointerString(m)
}
