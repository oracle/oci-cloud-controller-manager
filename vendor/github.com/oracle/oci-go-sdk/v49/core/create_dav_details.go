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

// CreateDavDetails Details to create a Direct Attached Vnic.
type CreateDavDetails struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the Direct Attached Vnic's compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Index of NIC for Direct Attached Vnic.
	NicIndex *int `mandatory:"true" json:"nicIndex"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Substrate IP address of the remote endpoint. This is a required property if
	// vcnxAttachmentIds property is not defined.
	RemoteEndpointSubstrateIp *string `mandatory:"false" json:"remoteEndpointSubstrateIp"`

	// The label type for Direct Attached Vnic. This is used to determine the
	// label forwarding to be used by the Direct Attached Vnic.
	LabelType CreateDavDetailsLabelTypeEnum `mandatory:"false" json:"labelType,omitempty"`

	// List of VCNx Attachments to a DRG. This is a required property
	// if remoteEndpointSubstrateIp property is not defined.
	VcnxAttachmentIds []string `mandatory:"false" json:"vcnxAttachmentIds"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m CreateDavDetails) String() string {
	return common.PointerString(m)
}

// CreateDavDetailsLabelTypeEnum Enum with underlying type: string
type CreateDavDetailsLabelTypeEnum string

// Set of constants representing the allowable values for CreateDavDetailsLabelTypeEnum
const (
	CreateDavDetailsLabelTypeMpls CreateDavDetailsLabelTypeEnum = "MPLS"
)

var mappingCreateDavDetailsLabelType = map[string]CreateDavDetailsLabelTypeEnum{
	"MPLS": CreateDavDetailsLabelTypeMpls,
}

// GetCreateDavDetailsLabelTypeEnumValues Enumerates the set of values for CreateDavDetailsLabelTypeEnum
func GetCreateDavDetailsLabelTypeEnumValues() []CreateDavDetailsLabelTypeEnum {
	values := make([]CreateDavDetailsLabelTypeEnum, 0)
	for _, v := range mappingCreateDavDetailsLabelType {
		values = append(values, v)
	}
	return values
}
