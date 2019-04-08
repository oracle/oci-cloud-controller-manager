// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Functions Service API
//
// API for the Functions service.
//

package functions

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateTriggerDetails Properties to create a new trigger.
type CreateTriggerDetails struct {

	// The display name of the trigger. The display name is unique within the function containing the trigger. Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The OCID of the function the trigger belongs to.
	FunctionId *string `mandatory:"true" json:"functionId"`

	// The type of the trigger.
	Type TriggerTypeEnum `mandatory:"true" json:"type"`

	// The URI path for the trigger.
	// Example: `/sayHello`, `/say/Hello`
	Source *string `mandatory:"true" json:"source"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m CreateTriggerDetails) String() string {
	return common.PointerString(m)
}
