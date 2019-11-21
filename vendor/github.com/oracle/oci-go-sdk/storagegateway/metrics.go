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

// Metrics The most current telemetry data posted from the specified storage gateway instance.
type Metrics struct {

	// Provides metrics on storage gateway resource capacity, such as the number of vCPUs, memory size, and so forth.
	Resource *MetricsResource `mandatory:"false" json:"resource"`

	// Provides metrics on resource usage, such as CPU and memory utilization, created or connected file systems
	// count, and so forth.
	Stats *MetricsStats `mandatory:"false" json:"stats"`

	// Provides metrics on errors, warnings, and rejected IO on the storage gateway.
	Issues *MetricsIssues `mandatory:"false" json:"issues"`

	// Provides metrics on each of the file systems in the storage gateway. Each key represents a file system name.
	Filesystems map[string]MetricsFilesystem `mandatory:"false" json:"filesystems"`

	// Provides metrics on each of the cloud syncs in the storage gateway. Each key represents a cloud sync name.
	Cloudsyncs map[string]MetricsCloudsync `mandatory:"false" json:"cloudsyncs"`
}

func (m Metrics) String() string {
	return common.PointerString(m)
}
