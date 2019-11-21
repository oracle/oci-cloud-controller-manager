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

// StorageGatewayHealth Information describing the current health of the storage gateway.
type StorageGatewayHealth struct {

	// The overall health status of the storage gateway.
	// *  ACTIVE - The storage gateway is running with no issues.
	// *  INACTIVE - The storage gateway instance has not yet been installed on a Compute instance or on premises,
	// or the control plane has not received recent heartbeat data.
	// *  WARNING - The storage gateway has warnings. The `reasons` strings describe why this health status appears.
	// *  CRITICAL - The storage gateway has critical issues. The `reasons` strings describe why this health
	// status appears.
	Status StorageGatewayHealthStatusEnum `mandatory:"true" json:"status"`

	// The version of the storage gateway instance.
	// Example: `2.0`
	Version *string `mandatory:"true" json:"version"`

	// Whether a newer storage gateway version is available.
	// Example: `false`
	IsNewerVersionAvailable *bool `mandatory:"true" json:"isNewerVersionAvailable"`

	// A timestamp identifying when the most recent heartbeat was received, in the format defined by RFC3339.
	// Example: `2019-05-17T19:55:49.263Z`
	TimeLastHeartbeatReceived *common.SDKTime `mandatory:"true" json:"timeLastHeartbeatReceived"`

	// Metrics data about the specified storage gateway.
	Metrics *Metrics `mandatory:"true" json:"metrics"`

	// Additional information about WARNING and CRITICAL health statuses.
	Reasons *StatusReasons `mandatory:"false" json:"reasons"`
}

func (m StorageGatewayHealth) String() string {
	return common.PointerString(m)
}

// StorageGatewayHealthStatusEnum Enum with underlying type: string
type StorageGatewayHealthStatusEnum string

// Set of constants representing the allowable values for StorageGatewayHealthStatusEnum
const (
	StorageGatewayHealthStatusActive   StorageGatewayHealthStatusEnum = "ACTIVE"
	StorageGatewayHealthStatusInactive StorageGatewayHealthStatusEnum = "INACTIVE"
	StorageGatewayHealthStatusWarning  StorageGatewayHealthStatusEnum = "WARNING"
	StorageGatewayHealthStatusCritical StorageGatewayHealthStatusEnum = "CRITICAL"
)

var mappingStorageGatewayHealthStatus = map[string]StorageGatewayHealthStatusEnum{
	"ACTIVE":   StorageGatewayHealthStatusActive,
	"INACTIVE": StorageGatewayHealthStatusInactive,
	"WARNING":  StorageGatewayHealthStatusWarning,
	"CRITICAL": StorageGatewayHealthStatusCritical,
}

// GetStorageGatewayHealthStatusEnumValues Enumerates the set of values for StorageGatewayHealthStatusEnum
func GetStorageGatewayHealthStatusEnumValues() []StorageGatewayHealthStatusEnum {
	values := make([]StorageGatewayHealthStatusEnum, 0)
	for _, v := range mappingStorageGatewayHealthStatus {
		values = append(values, v)
	}
	return values
}
