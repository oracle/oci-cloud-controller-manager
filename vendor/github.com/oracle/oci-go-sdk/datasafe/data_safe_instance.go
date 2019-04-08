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

// DataSafeInstance Object of DataSafeInstance
type DataSafeInstance struct {

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
	LifecycleState DataSafeInstanceLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Simple key-value pair that is applied without any predefined name, type or scope. Exists for cross-compatibility only.
	// Example: `{"bar-key": "value"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Usage of predefined tag keys. These predefined keys are scoped to namespaces.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m DataSafeInstance) String() string {
	return common.PointerString(m)
}

// DataSafeInstanceLifecycleStateEnum Enum with underlying type: string
type DataSafeInstanceLifecycleStateEnum string

// Set of constants representing the allowable values for DataSafeInstanceLifecycleStateEnum
const (
	DataSafeInstanceLifecycleStateCreating DataSafeInstanceLifecycleStateEnum = "CREATING"
	DataSafeInstanceLifecycleStateUpdating DataSafeInstanceLifecycleStateEnum = "UPDATING"
	DataSafeInstanceLifecycleStateActive   DataSafeInstanceLifecycleStateEnum = "ACTIVE"
	DataSafeInstanceLifecycleStateDeleting DataSafeInstanceLifecycleStateEnum = "DELETING"
	DataSafeInstanceLifecycleStateDeleted  DataSafeInstanceLifecycleStateEnum = "DELETED"
	DataSafeInstanceLifecycleStateFailed   DataSafeInstanceLifecycleStateEnum = "FAILED"
)

var mappingDataSafeInstanceLifecycleState = map[string]DataSafeInstanceLifecycleStateEnum{
	"CREATING": DataSafeInstanceLifecycleStateCreating,
	"UPDATING": DataSafeInstanceLifecycleStateUpdating,
	"ACTIVE":   DataSafeInstanceLifecycleStateActive,
	"DELETING": DataSafeInstanceLifecycleStateDeleting,
	"DELETED":  DataSafeInstanceLifecycleStateDeleted,
	"FAILED":   DataSafeInstanceLifecycleStateFailed,
}

// GetDataSafeInstanceLifecycleStateEnumValues Enumerates the set of values for DataSafeInstanceLifecycleStateEnum
func GetDataSafeInstanceLifecycleStateEnumValues() []DataSafeInstanceLifecycleStateEnum {
	values := make([]DataSafeInstanceLifecycleStateEnum, 0)
	for _, v := range mappingDataSafeInstanceLifecycleState {
		values = append(values, v)
	}
	return values
}
