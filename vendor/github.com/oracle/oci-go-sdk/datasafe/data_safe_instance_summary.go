// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Data Safe Control Plane API
//
// The API to manage data safe instance creation and deletion
//

package datasafe

import (
	"github.com/oracle/oci-go-sdk/common"
)

// DataSafeInstanceSummary Summary of the data safe instance.
type DataSafeInstanceSummary struct {

	// Unique identifier that is immutable on creation
	Id *string `mandatory:"true" json:"id"`

	// Compartment Identifier
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Data safe instance name, can be renamed
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Service URL of the data safe instance
	Url *string `mandatory:"false" json:"url"`

	// Description of the data safe instance
	Description *string `mandatory:"false" json:"description"`

	// The time the the data safe instance was created. An RFC3339 formatted datetime string
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The current state of the data safe instance.
	LifecycleState DataSafeInstanceSummaryLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`
}

func (m DataSafeInstanceSummary) String() string {
	return common.PointerString(m)
}

// DataSafeInstanceSummaryLifecycleStateEnum Enum with underlying type: string
type DataSafeInstanceSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for DataSafeInstanceSummaryLifecycleStateEnum
const (
	DataSafeInstanceSummaryLifecycleStateCreating DataSafeInstanceSummaryLifecycleStateEnum = "CREATING"
	DataSafeInstanceSummaryLifecycleStateUpdating DataSafeInstanceSummaryLifecycleStateEnum = "UPDATING"
	DataSafeInstanceSummaryLifecycleStateActive   DataSafeInstanceSummaryLifecycleStateEnum = "ACTIVE"
	DataSafeInstanceSummaryLifecycleStateDeleting DataSafeInstanceSummaryLifecycleStateEnum = "DELETING"
	DataSafeInstanceSummaryLifecycleStateDeleted  DataSafeInstanceSummaryLifecycleStateEnum = "DELETED"
	DataSafeInstanceSummaryLifecycleStateFailed   DataSafeInstanceSummaryLifecycleStateEnum = "FAILED"
)

var mappingDataSafeInstanceSummaryLifecycleState = map[string]DataSafeInstanceSummaryLifecycleStateEnum{
	"CREATING": DataSafeInstanceSummaryLifecycleStateCreating,
	"UPDATING": DataSafeInstanceSummaryLifecycleStateUpdating,
	"ACTIVE":   DataSafeInstanceSummaryLifecycleStateActive,
	"DELETING": DataSafeInstanceSummaryLifecycleStateDeleting,
	"DELETED":  DataSafeInstanceSummaryLifecycleStateDeleted,
	"FAILED":   DataSafeInstanceSummaryLifecycleStateFailed,
}

// GetDataSafeInstanceSummaryLifecycleStateEnumValues Enumerates the set of values for DataSafeInstanceSummaryLifecycleStateEnum
func GetDataSafeInstanceSummaryLifecycleStateEnumValues() []DataSafeInstanceSummaryLifecycleStateEnum {
	values := make([]DataSafeInstanceSummaryLifecycleStateEnum, 0)
	for _, v := range mappingDataSafeInstanceSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}
