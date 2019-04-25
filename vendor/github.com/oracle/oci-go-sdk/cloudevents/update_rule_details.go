// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// CloudEvents API
//
// API for the CloudEvents Service. Use this API to manage rules and actions that create automation
// in your tenancy. For more information, see Overview of Events (https://docs.cloud.oracle.com/iaas/Content/Events/Concepts/eventsoverview.htm).
//

package cloudevents

import (
	"github.com/oracle/oci-go-sdk/common"
)

// UpdateRuleDetails The rule attributes that you can update.
type UpdateRuleDetails struct {

	// A string that describes the rule. It does not have to be unique, and you can change it. Avoid entering
	// confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// A string that describes the details of the rule. It does not have to be unique, and you can change it. Avoid entering
	// confidential information.
	Description *string `mandatory:"false" json:"description"`

	// Whether or not this rule is currently enabled.
	// Example: `true`
	IsEnabled *bool `mandatory:"false" json:"isEnabled"`

	// Specifies the event that will trigger the actions associated with this rule.
	// Example: `"eventType": "com.oraclecloud.dbaas.autonomous.database.backup.end"`
	Condition *string `mandatory:"false" json:"condition"`

	// Action object.
	Actions *ActionDetailsList `mandatory:"false" json:"actions"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace. Exists for cross-compatibility only.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}'
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m UpdateRuleDetails) String() string {
	return common.PointerString(m)
}
