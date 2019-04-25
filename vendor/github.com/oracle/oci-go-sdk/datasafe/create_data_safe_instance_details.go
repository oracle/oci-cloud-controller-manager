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

// CreateDataSafeInstanceDetails The information about new data safe instance.
type CreateDataSafeInstanceDetails struct {

	// data safe instance name
	DisplayName *string `mandatory:"true" json:"displayName"`

	// data safe instance Compartment Identifier
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// data safe instance Description
	Description *string `mandatory:"false" json:"description"`

	// Simple key-value pair that is applied without any predefined name, type or scope. Exists for cross-compatibility only.
	// Example: `{"bar-key": "value"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Usage of predefined tag keys. These predefined keys are scoped to namespaces.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m CreateDataSafeInstanceDetails) String() string {
	return common.PointerString(m)
}
