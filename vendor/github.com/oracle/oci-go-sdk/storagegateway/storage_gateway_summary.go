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

// StorageGatewaySummary Summary view of the specified storage gateway.
type StorageGatewaySummary struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the storage gateway.
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the storage gateway.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The user-friendly name of the storage gateway. It does not have to be unique, and it is changeable.
	// Example: `example_storage_gateway`
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The overall health status of the storage gateway.
	// *  ACTIVE - The storage gateway is running with no issues.
	// *  INACTIVE - The storage gateway instance has not yet been installed on a Compute instance or on premises,
	// or the control plane has not received recent heartbeat data.
	// *  WARNING - The storage gateway has warnings.
	// *  CRITICAL - The storage gateway has critical issues.
	Status StorageGatewaySummaryStatusEnum `mandatory:"true" json:"status"`

	// The version of the storage gateway instance.
	// Example: `2.0`
	Version *string `mandatory:"true" json:"version"`

	// Whether a newer storage gateway version is available.
	// Example: `false`
	IsNewerVersionAvailable *bool `mandatory:"true" json:"isNewerVersionAvailable"`

	// The date and time the storage gateway was created, in the format defined by RFC3339.
	// Example: `2019-05-16T21:52:40.793Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The current lifecycle state of the storage gateway. You cannot use the storage gateway before the state is ACTIVE.
	LifecycleState LifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The changeable description assigned to the storage gateway during creation. It does not have to be unique.
	//  Example: `my first storage gateway`
	Description *string `mandatory:"false" json:"description"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information,
	// see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m StorageGatewaySummary) String() string {
	return common.PointerString(m)
}

// StorageGatewaySummaryStatusEnum Enum with underlying type: string
type StorageGatewaySummaryStatusEnum string

// Set of constants representing the allowable values for StorageGatewaySummaryStatusEnum
const (
	StorageGatewaySummaryStatusActive   StorageGatewaySummaryStatusEnum = "ACTIVE"
	StorageGatewaySummaryStatusInactive StorageGatewaySummaryStatusEnum = "INACTIVE"
	StorageGatewaySummaryStatusWarning  StorageGatewaySummaryStatusEnum = "WARNING"
	StorageGatewaySummaryStatusCritical StorageGatewaySummaryStatusEnum = "CRITICAL"
)

var mappingStorageGatewaySummaryStatus = map[string]StorageGatewaySummaryStatusEnum{
	"ACTIVE":   StorageGatewaySummaryStatusActive,
	"INACTIVE": StorageGatewaySummaryStatusInactive,
	"WARNING":  StorageGatewaySummaryStatusWarning,
	"CRITICAL": StorageGatewaySummaryStatusCritical,
}

// GetStorageGatewaySummaryStatusEnumValues Enumerates the set of values for StorageGatewaySummaryStatusEnum
func GetStorageGatewaySummaryStatusEnumValues() []StorageGatewaySummaryStatusEnum {
	values := make([]StorageGatewaySummaryStatusEnum, 0)
	for _, v := range mappingStorageGatewaySummaryStatus {
		values = append(values, v)
	}
	return values
}
