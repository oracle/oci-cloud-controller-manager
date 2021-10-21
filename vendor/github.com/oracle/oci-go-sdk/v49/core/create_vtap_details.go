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

// CreateVtapDetails These details are included in a request to create a virtual test access point (VTAP).
type CreateVtapDetails struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment containing the `Vtap` resource.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the VCN containing the `Vtap` resource.
	VcnId *string `mandatory:"true" json:"vcnId"`

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

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the destination resource where mirrored packets are sent.
	TargetId *string `mandatory:"false" json:"targetId"`

	// The IP address of the destination resource where mirrored packets are sent.
	TargetIp *string `mandatory:"false" json:"targetIp"`

	// Defines an encapsulation header type for the VTAP's mirrored traffic.
	EncapsulationProtocol CreateVtapDetailsEncapsulationProtocolEnum `mandatory:"false" json:"encapsulationProtocol,omitempty"`

	// The virtual extensible LAN (VXLAN) network identifier (or VXLAN segment ID) that uniquely identifies the VXLAN.
	VxlanNetworkIdentifier *int64 `mandatory:"false" json:"vxlanNetworkIdentifier"`

	// Used to start or stop a `Vtap` resource.
	// * `TRUE` directs the VTAP to start mirroring traffic.
	// * `FALSE` (Default) directs the VTAP to stop mirroring traffic.
	IsVtapEnabled *bool `mandatory:"false" json:"isVtapEnabled"`

	// Used to encapsulate or decapsulate mirrored traffic prior to ingestion at target.
	IsTargetEncapsulationEnabled *bool `mandatory:"false" json:"isTargetEncapsulationEnabled"`

	// The source type for the VTAP.
	SourceType CreateVtapDetailsSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`

	ExclusionFilter *ExclusionFilterDetails `mandatory:"false" json:"exclusionFilter"`
}

func (m CreateVtapDetails) String() string {
	return common.PointerString(m)
}

// CreateVtapDetailsEncapsulationProtocolEnum Enum with underlying type: string
type CreateVtapDetailsEncapsulationProtocolEnum string

// Set of constants representing the allowable values for CreateVtapDetailsEncapsulationProtocolEnum
const (
	CreateVtapDetailsEncapsulationProtocolVxlan CreateVtapDetailsEncapsulationProtocolEnum = "VXLAN"
)

var mappingCreateVtapDetailsEncapsulationProtocol = map[string]CreateVtapDetailsEncapsulationProtocolEnum{
	"VXLAN": CreateVtapDetailsEncapsulationProtocolVxlan,
}

// GetCreateVtapDetailsEncapsulationProtocolEnumValues Enumerates the set of values for CreateVtapDetailsEncapsulationProtocolEnum
func GetCreateVtapDetailsEncapsulationProtocolEnumValues() []CreateVtapDetailsEncapsulationProtocolEnum {
	values := make([]CreateVtapDetailsEncapsulationProtocolEnum, 0)
	for _, v := range mappingCreateVtapDetailsEncapsulationProtocol {
		values = append(values, v)
	}
	return values
}

// CreateVtapDetailsSourceTypeEnum Enum with underlying type: string
type CreateVtapDetailsSourceTypeEnum string

// Set of constants representing the allowable values for CreateVtapDetailsSourceTypeEnum
const (
	CreateVtapDetailsSourceTypeVnic   CreateVtapDetailsSourceTypeEnum = "VNIC"
	CreateVtapDetailsSourceTypeSubnet CreateVtapDetailsSourceTypeEnum = "SUBNET"
)

var mappingCreateVtapDetailsSourceType = map[string]CreateVtapDetailsSourceTypeEnum{
	"VNIC":   CreateVtapDetailsSourceTypeVnic,
	"SUBNET": CreateVtapDetailsSourceTypeSubnet,
}

// GetCreateVtapDetailsSourceTypeEnumValues Enumerates the set of values for CreateVtapDetailsSourceTypeEnum
func GetCreateVtapDetailsSourceTypeEnumValues() []CreateVtapDetailsSourceTypeEnum {
	values := make([]CreateVtapDetailsSourceTypeEnum, 0)
	for _, v := range mappingCreateVtapDetailsSourceType {
		values = append(values, v)
	}
	return values
}
