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

// MetricsCloudsync cloud sync-specific metrics.
type MetricsCloudsync struct {

	// The total number of cloud sync errors.
	// Example: `1`
	ErrorCount *float32 `mandatory:"false" json:"errorCount"`

	// The total number of cloud sync warnings.
	// Example: `3`
	WarnCount *float32 `mandatory:"false" json:"warnCount"`

	// Cloud sync cache utilization (percent).
	// Example: `45`
	CacheUtilPercent *float32 `mandatory:"false" json:"cacheUtilPercent"`

	// Cloud sync cache hits (percent).
	// Example: `80`
	CacheHitPercent *float32 `mandatory:"false" json:"cacheHitPercent"`

	// Uploaded cloud sync data in bytes.
	// Example: `22369621334`
	UploadedDataInBytes *float32 `mandatory:"false" json:"uploadedDataInBytes"`

	// Downloaded cloud sync data in bytes.
	// Example: `13421772800`
	DownloadedDataInBytes *float32 `mandatory:"false" json:"downloadedDataInBytes"`

	// Cloud sync write (ingestion) data in bytes.
	// Example: `8947848534`
	WriteDataInBytes *float32 `mandatory:"false" json:"writeDataInBytes"`

	// Cloud sync read data in bytes.
	// Example: `10737418240`
	ReadDataInBytes *float32 `mandatory:"false" json:"readDataInBytes"`

	// Cloud sync pending data in bytes.
	// Example: `894784854`
	PendingDataInBytes *float32 `mandatory:"false" json:"pendingDataInBytes"`

	// Source data size in bytes.
	// Example: `13421772800`
	SourceDataInBytes *float32 `mandatory:"false" json:"sourceDataInBytes"`
}

func (m MetricsCloudsync) String() string {
	return common.PointerString(m)
}
