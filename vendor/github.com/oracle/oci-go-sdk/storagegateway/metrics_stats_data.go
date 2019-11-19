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

// MetricsStatsData Data activity statistics.
type MetricsStatsData struct {

	// Uploaded data in bytes.
	// Example: `134217728000`
	UploadedDataInBytes *float32 `mandatory:"false" json:"uploadedDataInBytes"`

	// Downloaded data in bytes.
	// Example: `80530636800`
	DownloadedDataInBytes *float32 `mandatory:"false" json:"downloadedDataInBytes"`

	// Write (ingestion) data in bytes.
	// Example: `53687091200`
	WriteDataInBytes *float32 `mandatory:"false" json:"writeDataInBytes"`

	// Read data in bytes.
	// Example: `64424509440`
	ReadDataInBytes *float32 `mandatory:"false" json:"readDataInBytes"`

	// Pending data in bytes.
	// Example: `5368709120`
	PendingDataInBytes *float32 `mandatory:"false" json:"pendingDataInBytes"`

	// Upload throughput in megabytes per second.
	UploadThroughputInMBps *float32 `mandatory:"false" json:"uploadThroughputInMBps"`

	// Download throughput in megabytes per second.
	DownloadThroughputInMBps *float32 `mandatory:"false" json:"downloadThroughputInMBps"`
}

func (m MetricsStatsData) String() string {
	return common.PointerString(m)
}
