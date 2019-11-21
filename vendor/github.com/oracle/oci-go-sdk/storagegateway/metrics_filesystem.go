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

// MetricsFilesystem File system-specific metrics.
type MetricsFilesystem struct {

	// The total number of file system errors.
	// Example: `1`
	ErrorCount *float32 `mandatory:"false" json:"errorCount"`

	// The total number of file system warnings.
	// Example: `3`
	WarnCount *float32 `mandatory:"false" json:"warnCount"`

	// File system cache utilization (percent).
	// Example: `45`
	CacheUtilPercent *float32 `mandatory:"false" json:"cacheUtilPercent"`

	// File system cache hits (percent).
	// Example: `80`
	CacheHitPercent *float32 `mandatory:"false" json:"cacheHitPercent"`

	// Uploaded file system data in bytes.
	// Example: `44739242667`
	UploadedDataInBytes *float32 `mandatory:"false" json:"uploadedDataInBytes"`

	// Downloaded file system data in bytes.
	// Example: `26843545600`
	DownloadedDataInBytes *float32 `mandatory:"false" json:"downloadedDataInBytes"`

	// File system write (ingestion) data in bytes.
	// Example: `17895697067`
	WriteDataInBytes *float32 `mandatory:"false" json:"writeDataInBytes"`

	// File system read data in bytes.
	// Example: `21474836480`
	ReadDataInBytes *float32 `mandatory:"false" json:"readDataInBytes"`

	// File system pending data in bytes.
	// Example: `1789569707`
	PendingDataInBytes *float32 `mandatory:"false" json:"pendingDataInBytes"`

	// File system object storage usage in bytes.
	ObjectStorageUsageInBytes *float32 `mandatory:"false" json:"objectStorageUsageInBytes"`

	// Zero if the file system is not connected.  Non-zero if the file system is connected.
	// Example: `0`
	IsConnected *float32 `mandatory:"false" json:"isConnected"`

	// Zero if the file system is not in refresh mode. Non-zero if the file system is in refresh mode.
	// Example: `0`
	IsInRefreshMode *float32 `mandatory:"false" json:"isInRefreshMode"`
}

func (m MetricsFilesystem) String() string {
	return common.PointerString(m)
}
