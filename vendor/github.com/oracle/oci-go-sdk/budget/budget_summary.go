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

// BudgetSummary A budget.
type BudgetSummary struct {

	// The OCID of the budget
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the compartment
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The display name of the budget.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The amount of the budget expressed in the currency of the customer's rate card.
	Amount *float32 `mandatory:"true" json:"amount"`

	// The reset period for the budget.
	ResetPeriod BudgetSummaryResetPeriodEnum `mandatory:"true" json:"resetPeriod"`

	// The current state of the budget.
	LifecycleState BudgetSummaryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Total number of alert rules in the budget
	AlertRuleCount *int `mandatory:"true" json:"alertRuleCount"`

	// Time budget was created
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// Time budget was updated
	TimeUpdated *common.SDKTime `mandatory:"true" json:"timeUpdated"`

	// This is DEPRECATED. For backwards compatability, the property will be populated when
	// targetType is "COMPARTMENT" AND targets contains EXACT ONE target compartment ocid.
	// For all other scenarios, this property will be left empty.
	TargetCompartmentId *string `mandatory:"false" json:"targetCompartmentId"`

	// The description of the budget.
	Description *string `mandatory:"false" json:"description"`

	// The type of target on which the budget is applied.
	TargetType BudgetSummaryTargetTypeEnum `mandatory:"false" json:"targetType,omitempty"`

	// The list of targets on which the budget is applied.
	//   If targetType is "COMPARTMENT", targets contains list of compartment OCIDs.
	//   If targetType is "TAG", targets contains list of tag identifiers in the form of "{tagNamespace}.{tagKey}.{tagValue}".
	Targets []string `mandatory:"false" json:"targets"`

	// Version of the budget. Starts from 1 and increments by 1.
	Version *int `mandatory:"false" json:"version"`

	// The actual spend in currency for the current budget cycle
	ActualSpend *float32 `mandatory:"false" json:"actualSpend"`

	// The forecasted spend in currency by the end of the current budget cycle
	ForecastedSpend *float32 `mandatory:"false" json:"forecastedSpend"`

	// Time budget spend was last computed
	TimeSpendComputed *common.SDKTime `mandatory:"false" json:"timeSpendComputed"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m BudgetSummary) String() string {
	return common.PointerString(m)
}

// BudgetSummaryResetPeriodEnum Enum with underlying type: string
type BudgetSummaryResetPeriodEnum string

// Set of constants representing the allowable values for BudgetSummaryResetPeriodEnum
const (
	BudgetSummaryResetPeriodMonthly BudgetSummaryResetPeriodEnum = "MONTHLY"
)

var mappingBudgetSummaryResetPeriod = map[string]BudgetSummaryResetPeriodEnum{
	"MONTHLY": BudgetSummaryResetPeriodMonthly,
}

// GetBudgetSummaryResetPeriodEnumValues Enumerates the set of values for BudgetSummaryResetPeriodEnum
func GetBudgetSummaryResetPeriodEnumValues() []BudgetSummaryResetPeriodEnum {
	values := make([]BudgetSummaryResetPeriodEnum, 0)
	for _, v := range mappingBudgetSummaryResetPeriod {
		values = append(values, v)
	}
	return values
}

// BudgetSummaryTargetTypeEnum Enum with underlying type: string
type BudgetSummaryTargetTypeEnum string

// Set of constants representing the allowable values for BudgetSummaryTargetTypeEnum
const (
	BudgetSummaryTargetTypeCompartment BudgetSummaryTargetTypeEnum = "COMPARTMENT"
	BudgetSummaryTargetTypeTag         BudgetSummaryTargetTypeEnum = "TAG"
)

var mappingBudgetSummaryTargetType = map[string]BudgetSummaryTargetTypeEnum{
	"COMPARTMENT": BudgetSummaryTargetTypeCompartment,
	"TAG":         BudgetSummaryTargetTypeTag,
}

// GetBudgetSummaryTargetTypeEnumValues Enumerates the set of values for BudgetSummaryTargetTypeEnum
func GetBudgetSummaryTargetTypeEnumValues() []BudgetSummaryTargetTypeEnum {
	values := make([]BudgetSummaryTargetTypeEnum, 0)
	for _, v := range mappingBudgetSummaryTargetType {
		values = append(values, v)
	}
	return values
}

// BudgetSummaryLifecycleStateEnum Enum with underlying type: string
type BudgetSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for BudgetSummaryLifecycleStateEnum
const (
	BudgetSummaryLifecycleStateActive   BudgetSummaryLifecycleStateEnum = "ACTIVE"
	BudgetSummaryLifecycleStateInactive BudgetSummaryLifecycleStateEnum = "INACTIVE"
)

var mappingBudgetSummaryLifecycleState = map[string]BudgetSummaryLifecycleStateEnum{
	"ACTIVE":   BudgetSummaryLifecycleStateActive,
	"INACTIVE": BudgetSummaryLifecycleStateInactive,
}

// GetBudgetSummaryLifecycleStateEnumValues Enumerates the set of values for BudgetSummaryLifecycleStateEnum
func GetBudgetSummaryLifecycleStateEnumValues() []BudgetSummaryLifecycleStateEnum {
	values := make([]BudgetSummaryLifecycleStateEnum, 0)
	for _, v := range mappingBudgetSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}
