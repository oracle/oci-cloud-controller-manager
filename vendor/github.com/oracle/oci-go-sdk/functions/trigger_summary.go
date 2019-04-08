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

// TriggerSummary Summary of a trigger.
type TriggerSummary struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the trigger.
	Id *string `mandatory:"false" json:"id"`

	// The display name of the trigger. The display name is unique within the function containing the trigger.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The OCID of the function the trigger belongs to.
	FunctionId *string `mandatory:"false" json:"functionId"`

	// The OCID of the application the trigger belongs to.
	ApplicationId *string `mandatory:"false" json:"applicationId"`

	// The OCID of the compartment that contains the trigger.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The type of the trigger.
	Type TriggerTypeEnum `mandatory:"false" json:"type,omitempty"`

	// The URI path for the trigger.
	// Example: `/sayHello`, `/say/Hello`
	Source *string `mandatory:"false" json:"source"`

	// The fully qualified endpoint URL for the trigger.
	Endpoint *string `mandatory:"false" json:"endpoint"`

	// The current state of the trigger.
	LifecycleState TriggerLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// The time the trigger was created, expressed in RFC 3339 (https://tools.ietf.org/html/rfc3339)
	// timestamp format.
	// Example: `2018-09-12T22:47:12.613Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The time the trigger was updated, expressed in RFC 3339 (https://tools.ietf.org/html/rfc3339)
	// timestamp format.
	// Example: `2018-09-12T22:47:12.613Z`
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`
}

func (m TriggerSummary) String() string {
	return common.PointerString(m)
}
