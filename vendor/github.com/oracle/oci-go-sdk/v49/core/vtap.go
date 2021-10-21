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

// Vtap A virtual test access point (VTAP) provides a way to mirror all traffic from a designated source to a selected target in order to facilitate troubleshooting, security analysis, and data monitoring.
// A VTAP is functionally similar to a test access point (TAP) you might deploy in your on-premises network.
// A *CaptureFilter* contains a set of *CaptureFilterRuleDetails* governing what traffic a VTAP mirrors.
type Vtap struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment containing the `Vtap` resource.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the VCN containing the `Vtap` resource.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// The VTAP's Oracle ID (OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm)).
	Id *string `mandatory:"true" json:"id"`

	// The VTAP's administrative lifecycle state.
	LifecycleState VtapLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the source point where packets are captured.
	SourceId *string `mandatory:"true" json:"sourceId"`

	// The capture filter's Oracle ID (OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm)).
	CaptureFilterId *string `mandatory:"true" json:"captureFilterId"`

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

	// The VTAP's current running state.
	LifecycleStateDetails VtapLifecycleStateDetailsEnum `mandatory:"false" json:"lifecycleStateDetails,omitempty"`

	// The date and time the VTAP was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2020-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the destination resource where mirrored packets are sent.
	TargetId *string `mandatory:"false" json:"targetId"`

	// The IP address of the destination resource where mirrored packets are sent.
	TargetIp *string `mandatory:"false" json:"targetIp"`

	// Defines an encapsulation header type for the VTAP's mirrored traffic.
	EncapsulationProtocol VtapEncapsulationProtocolEnum `mandatory:"false" json:"encapsulationProtocol,omitempty"`

	// The virtual extensible LAN (VXLAN) network identifier (or VXLAN segment ID) that uniquely identifies the VXLAN.
	VxlanNetworkIdentifier *int64 `mandatory:"false" json:"vxlanNetworkIdentifier"`

	// Used to start or stop a `Vtap` resource.
	// * `TRUE` directs the VTAP to start mirroring traffic.
	// * `FALSE` (Default) directs the VTAP to stop mirroring traffic.
	IsVtapEnabled *bool `mandatory:"false" json:"isVtapEnabled"`

	// Used to encapsulate or decapsulate mirrored traffic prior to ingestion at target.
	IsTargetEncapsulationEnabled *bool `mandatory:"false" json:"isTargetEncapsulationEnabled"`

	// The source type for the VTAP.
	SourceType VtapSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`

	ExclusionFilter *ExclusionFilterDetails `mandatory:"false" json:"exclusionFilter"`
}

func (m Vtap) String() string {
	return common.PointerString(m)
}

// VtapLifecycleStateEnum Enum with underlying type: string
type VtapLifecycleStateEnum string

// Set of constants representing the allowable values for VtapLifecycleStateEnum
const (
	VtapLifecycleStateProvisioning VtapLifecycleStateEnum = "PROVISIONING"
	VtapLifecycleStateAvailable    VtapLifecycleStateEnum = "AVAILABLE"
	VtapLifecycleStateUpdating     VtapLifecycleStateEnum = "UPDATING"
	VtapLifecycleStateTerminating  VtapLifecycleStateEnum = "TERMINATING"
	VtapLifecycleStateTerminated   VtapLifecycleStateEnum = "TERMINATED"
)

var mappingVtapLifecycleState = map[string]VtapLifecycleStateEnum{
	"PROVISIONING": VtapLifecycleStateProvisioning,
	"AVAILABLE":    VtapLifecycleStateAvailable,
	"UPDATING":     VtapLifecycleStateUpdating,
	"TERMINATING":  VtapLifecycleStateTerminating,
	"TERMINATED":   VtapLifecycleStateTerminated,
}

// GetVtapLifecycleStateEnumValues Enumerates the set of values for VtapLifecycleStateEnum
func GetVtapLifecycleStateEnumValues() []VtapLifecycleStateEnum {
	values := make([]VtapLifecycleStateEnum, 0)
	for _, v := range mappingVtapLifecycleState {
		values = append(values, v)
	}
	return values
}

// VtapLifecycleStateDetailsEnum Enum with underlying type: string
type VtapLifecycleStateDetailsEnum string

// Set of constants representing the allowable values for VtapLifecycleStateDetailsEnum
const (
	VtapLifecycleStateDetailsRunning VtapLifecycleStateDetailsEnum = "RUNNING"
	VtapLifecycleStateDetailsStopped VtapLifecycleStateDetailsEnum = "STOPPED"
)

var mappingVtapLifecycleStateDetails = map[string]VtapLifecycleStateDetailsEnum{
	"RUNNING": VtapLifecycleStateDetailsRunning,
	"STOPPED": VtapLifecycleStateDetailsStopped,
}

// GetVtapLifecycleStateDetailsEnumValues Enumerates the set of values for VtapLifecycleStateDetailsEnum
func GetVtapLifecycleStateDetailsEnumValues() []VtapLifecycleStateDetailsEnum {
	values := make([]VtapLifecycleStateDetailsEnum, 0)
	for _, v := range mappingVtapLifecycleStateDetails {
		values = append(values, v)
	}
	return values
}

// VtapEncapsulationProtocolEnum Enum with underlying type: string
type VtapEncapsulationProtocolEnum string

// Set of constants representing the allowable values for VtapEncapsulationProtocolEnum
const (
	VtapEncapsulationProtocolVxlan VtapEncapsulationProtocolEnum = "VXLAN"
)

var mappingVtapEncapsulationProtocol = map[string]VtapEncapsulationProtocolEnum{
	"VXLAN": VtapEncapsulationProtocolVxlan,
}

// GetVtapEncapsulationProtocolEnumValues Enumerates the set of values for VtapEncapsulationProtocolEnum
func GetVtapEncapsulationProtocolEnumValues() []VtapEncapsulationProtocolEnum {
	values := make([]VtapEncapsulationProtocolEnum, 0)
	for _, v := range mappingVtapEncapsulationProtocol {
		values = append(values, v)
	}
	return values
}

// VtapSourceTypeEnum Enum with underlying type: string
type VtapSourceTypeEnum string

// Set of constants representing the allowable values for VtapSourceTypeEnum
const (
	VtapSourceTypeVnic   VtapSourceTypeEnum = "VNIC"
	VtapSourceTypeSubnet VtapSourceTypeEnum = "SUBNET"
)

var mappingVtapSourceType = map[string]VtapSourceTypeEnum{
	"VNIC":   VtapSourceTypeVnic,
	"SUBNET": VtapSourceTypeSubnet,
}

// GetVtapSourceTypeEnumValues Enumerates the set of values for VtapSourceTypeEnum
func GetVtapSourceTypeEnumValues() []VtapSourceTypeEnum {
	values := make([]VtapSourceTypeEnum, 0)
	for _, v := range mappingVtapSourceType {
		values = append(values, v)
	}
	return values
}
