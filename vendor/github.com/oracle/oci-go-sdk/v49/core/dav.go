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

// Dav A Direct Attached Vnic.
type Dav struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the Direct Attached Vnic.
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the Direct Attached Vnic's compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The current state of the Direct Attached Vnic.
	LifecycleState DavLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Index of NIC for Direct Attached Vnic.
	NicIndex *int `mandatory:"true" json:"nicIndex"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the instance.
	InstanceId *string `mandatory:"false" json:"instanceId"`

	// The MAC address for the DAV. This will be a newly allocated MAC address
	// and not the one used by the instance.
	MacAddress *string `mandatory:"false" json:"macAddress"`

	// The substrate IP of DAV and primary VNIC attached to the instance.
	// This field will be null in case the DAV is not attached.
	SubstrateIp *string `mandatory:"false" json:"substrateIp"`

	// The allocated slot id for the Dav.
	SlotId *int `mandatory:"false" json:"slotId"`

	// The VLAN Tag assigned to Direct Attached Vnic.
	VlanTag *int `mandatory:"false" json:"vlanTag"`

	// The MAC address of the Virtual Router.
	VirtualRouterMac *string `mandatory:"false" json:"virtualRouterMac"`

	// Substrate IP address of the remote endpoint.
	RemoteEndpointSubstrateIp *string `mandatory:"false" json:"remoteEndpointSubstrateIp"`

	// List of VCNx Attachments to a DRG.
	VcnxAttachmentIds []string `mandatory:"false" json:"vcnxAttachmentIds"`

	// The label type for Direct Attached Vnic. This is used to determine the
	// label forwarding to be used by the Direct Attached Vnic.
	LabelType DavLabelTypeEnum `mandatory:"false" json:"labelType,omitempty"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m Dav) String() string {
	return common.PointerString(m)
}

// DavLifecycleStateEnum Enum with underlying type: string
type DavLifecycleStateEnum string

// Set of constants representing the allowable values for DavLifecycleStateEnum
const (
	DavLifecycleStateProvisioning DavLifecycleStateEnum = "PROVISIONING"
	DavLifecycleStateUpdating     DavLifecycleStateEnum = "UPDATING"
	DavLifecycleStateAvailable    DavLifecycleStateEnum = "AVAILABLE"
	DavLifecycleStateTerminating  DavLifecycleStateEnum = "TERMINATING"
	DavLifecycleStateTerminated   DavLifecycleStateEnum = "TERMINATED"
)

var mappingDavLifecycleState = map[string]DavLifecycleStateEnum{
	"PROVISIONING": DavLifecycleStateProvisioning,
	"UPDATING":     DavLifecycleStateUpdating,
	"AVAILABLE":    DavLifecycleStateAvailable,
	"TERMINATING":  DavLifecycleStateTerminating,
	"TERMINATED":   DavLifecycleStateTerminated,
}

// GetDavLifecycleStateEnumValues Enumerates the set of values for DavLifecycleStateEnum
func GetDavLifecycleStateEnumValues() []DavLifecycleStateEnum {
	values := make([]DavLifecycleStateEnum, 0)
	for _, v := range mappingDavLifecycleState {
		values = append(values, v)
	}
	return values
}

// DavLabelTypeEnum Enum with underlying type: string
type DavLabelTypeEnum string

// Set of constants representing the allowable values for DavLabelTypeEnum
const (
	DavLabelTypeMpls DavLabelTypeEnum = "MPLS"
)

var mappingDavLabelType = map[string]DavLabelTypeEnum{
	"MPLS": DavLabelTypeMpls,
}

// GetDavLabelTypeEnumValues Enumerates the set of values for DavLabelTypeEnum
func GetDavLabelTypeEnumValues() []DavLabelTypeEnum {
	values := make([]DavLabelTypeEnum, 0)
	for _, v := range mappingDavLabelType {
		values = append(values, v)
	}
	return values
}
