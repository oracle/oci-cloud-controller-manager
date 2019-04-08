// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Budgets API
//
// Use the Budgets API to manage budgets and budget alerts.
//

package budget

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateBudgetDetails The create budget details.
// Client should use 'targetType' & 'targets' to specify the target type and list of targets on which the budget is applied.
// For backwards compatibility, 'targetCompartmentId' will still be supported for all existing clients.
// However, this is considered deprecreated and all clients be upgraded to use 'targetType' & 'targets'.
// Specifying both 'targetCompartmentId' and 'targets' will cause a Bad Request.
type CreateBudgetDetails struct {

	// The OCID of the compartment
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The amount of the budget expressed as a whole number in the currency of the customer's rate card.
	Amount *float32 `mandatory:"true" json:"amount"`

	// The reset period for the budget.
	ResetPeriod CreateBudgetDetailsResetPeriodEnum `mandatory:"true" json:"resetPeriod"`

	// This is DEPRECTAED. Set the target compartment id in targets instead.
	TargetCompartmentId *string `mandatory:"false" json:"targetCompartmentId"`

	// The displayName of the budget.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The description of the budget.
	Description *string `mandatory:"false" json:"description"`

	// The type of target on which the budget is applied.
	TargetType CreateBudgetDetailsTargetTypeEnum `mandatory:"false" json:"targetType,omitempty"`

	// The list of targets on which the budget is applied.
	//   If targetType is "COMPARTMENT", targets contains list of compartment OCIDs.
	//   If targetType is "TAG", targets contains list of tag identifiers in the form of "{tagNamespace}.{tagKey}.{tagValue}".
	// Curerntly, the array should contain EXACT ONE item.
	Targets []string `mandatory:"false" json:"targets"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m CreateBudgetDetails) String() string {
	return common.PointerString(m)
}

// CreateBudgetDetailsResetPeriodEnum Enum with underlying type: string
type CreateBudgetDetailsResetPeriodEnum string

// Set of constants representing the allowable values for CreateBudgetDetailsResetPeriodEnum
const (
	CreateBudgetDetailsResetPeriodMonthly CreateBudgetDetailsResetPeriodEnum = "MONTHLY"
)

var mappingCreateBudgetDetailsResetPeriod = map[string]CreateBudgetDetailsResetPeriodEnum{
	"MONTHLY": CreateBudgetDetailsResetPeriodMonthly,
}

// GetCreateBudgetDetailsResetPeriodEnumValues Enumerates the set of values for CreateBudgetDetailsResetPeriodEnum
func GetCreateBudgetDetailsResetPeriodEnumValues() []CreateBudgetDetailsResetPeriodEnum {
	values := make([]CreateBudgetDetailsResetPeriodEnum, 0)
	for _, v := range mappingCreateBudgetDetailsResetPeriod {
		values = append(values, v)
	}
	return values
}

// CreateBudgetDetailsTargetTypeEnum Enum with underlying type: string
type CreateBudgetDetailsTargetTypeEnum string

// Set of constants representing the allowable values for CreateBudgetDetailsTargetTypeEnum
const (
	CreateBudgetDetailsTargetTypeCompartment CreateBudgetDetailsTargetTypeEnum = "COMPARTMENT"
	CreateBudgetDetailsTargetTypeTag         CreateBudgetDetailsTargetTypeEnum = "TAG"
)

var mappingCreateBudgetDetailsTargetType = map[string]CreateBudgetDetailsTargetTypeEnum{
	"COMPARTMENT": CreateBudgetDetailsTargetTypeCompartment,
	"TAG":         CreateBudgetDetailsTargetTypeTag,
}

// GetCreateBudgetDetailsTargetTypeEnumValues Enumerates the set of values for CreateBudgetDetailsTargetTypeEnum
func GetCreateBudgetDetailsTargetTypeEnumValues() []CreateBudgetDetailsTargetTypeEnum {
	values := make([]CreateBudgetDetailsTargetTypeEnum, 0)
	for _, v := range mappingCreateBudgetDetailsTargetType {
		values = append(values, v)
	}
	return values
}
