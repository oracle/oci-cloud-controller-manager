// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"github.com/oracle/oci-go-sdk/v49/common"
)

// CaptureFilter A capture filter contains a set of *CaptureFilterRuleDetails* governing what traffic a *Vtap* mirrors.
// The capture filter is created with no rules defined, and it must have at least one rule for the VTAP to start mirroring traffic.
type CaptureFilter struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment containing the capture filter.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the VCN containing the capture filter.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// The capture filter's Oracle ID (OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The capture filter's current administrative state.
	LifecycleState CaptureFilterLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The date and time the capture filter was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2021-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The set of rules governing what traffic a VTAP mirrors.
	CaptureFilterRules []CaptureFilterRuleDetails `mandatory:"false" json:"captureFilterRules"`
}

func (m CaptureFilter) String() string {
	return common.PointerString(m)
}

// CaptureFilterLifecycleStateEnum Enum with underlying type: string
type CaptureFilterLifecycleStateEnum string

// Set of constants representing the allowable values for CaptureFilterLifecycleStateEnum
const (
	CaptureFilterLifecycleStateProvisioning CaptureFilterLifecycleStateEnum = "PROVISIONING"
	CaptureFilterLifecycleStateAvailable    CaptureFilterLifecycleStateEnum = "AVAILABLE"
	CaptureFilterLifecycleStateUpdating     CaptureFilterLifecycleStateEnum = "UPDATING"
	CaptureFilterLifecycleStateTerminating  CaptureFilterLifecycleStateEnum = "TERMINATING"
	CaptureFilterLifecycleStateTerminated   CaptureFilterLifecycleStateEnum = "TERMINATED"
)

var mappingCaptureFilterLifecycleState = map[string]CaptureFilterLifecycleStateEnum{
	"PROVISIONING": CaptureFilterLifecycleStateProvisioning,
	"AVAILABLE":    CaptureFilterLifecycleStateAvailable,
	"UPDATING":     CaptureFilterLifecycleStateUpdating,
	"TERMINATING":  CaptureFilterLifecycleStateTerminating,
	"TERMINATED":   CaptureFilterLifecycleStateTerminated,
}

// GetCaptureFilterLifecycleStateEnumValues Enumerates the set of values for CaptureFilterLifecycleStateEnum
func GetCaptureFilterLifecycleStateEnumValues() []CaptureFilterLifecycleStateEnum {
	values := make([]CaptureFilterLifecycleStateEnum, 0)
	for _, v := range mappingCaptureFilterLifecycleState {
		values = append(values, v)
	}
	return values
}
