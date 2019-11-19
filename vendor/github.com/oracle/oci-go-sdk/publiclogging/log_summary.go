// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// PublicLoggingControlplane API
//
// PublicLoggingControlplane API specification
//

package publiclogging

import (
	"github.com/oracle/oci-go-sdk/common"
)

// LogSummary Log object configuratio summary.
type LogSummary struct {

	// The OCID of the resource.
	Id *string `mandatory:"true" json:"id"`

	// Log group OCID.
	LogGroupId *string `mandatory:"true" json:"logGroupId"`

	// The display name of a user-friendly name. It has to be unique within enclosing resource,
	// and it's changeable. Avoid entering confidential information.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The state of an pipeline.
	LifecycleState LogLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The logType that the log object is for, custom or service.
	LogType LogSummaryLogTypeEnum `mandatory:"true" json:"logType"`

	// Whether or not this resource is currently enabled.
	IsEnabled *bool `mandatory:"false" json:"isEnabled"`

	Configuration *Configuration `mandatory:"false" json:"configuration"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Time the resource was created.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Time the resource was last modified.
	TimeLastModified *common.SDKTime `mandatory:"false" json:"timeLastModified"`
}

func (m LogSummary) String() string {
	return common.PointerString(m)
}

// LogSummaryLogTypeEnum Enum with underlying type: string
type LogSummaryLogTypeEnum string

// Set of constants representing the allowable values for LogSummaryLogTypeEnum
const (
	LogSummaryLogTypeCustom  LogSummaryLogTypeEnum = "CUSTOM"
	LogSummaryLogTypeService LogSummaryLogTypeEnum = "SERVICE"
)

var mappingLogSummaryLogType = map[string]LogSummaryLogTypeEnum{
	"CUSTOM":  LogSummaryLogTypeCustom,
	"SERVICE": LogSummaryLogTypeService,
}

// GetLogSummaryLogTypeEnumValues Enumerates the set of values for LogSummaryLogTypeEnum
func GetLogSummaryLogTypeEnumValues() []LogSummaryLogTypeEnum {
	values := make([]LogSummaryLogTypeEnum, 0)
	for _, v := range mappingLogSummaryLogType {
		values = append(values, v)
	}
	return values
}
