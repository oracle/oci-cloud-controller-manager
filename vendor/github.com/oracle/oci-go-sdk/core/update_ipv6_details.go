// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"github.com/oracle/oci-go-sdk/common"
)

// UpdateIpv6Details The representation of UpdateIpv6Details
type UpdateIpv6Details struct {

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable. Avoid
	// entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see
	// Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Whether the IPv6 can be used for internet communication. Allowed by default for an IPv6 in
	// a public subnet. Never allowed for an IPv6 in a private subnet. If the value is `true`, the
	// IPv6 uses its public IP address for internet communication.
	// If you switch this from `true` to `false`, the `publicIpAddress` attribute for the IPv6
	// becomes null.
	// Example: `false`
	IsInternetAccessAllowed *bool `mandatory:"false" json:"isInternetAccessAllowed"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the VNIC to reassign the IPv6 to.
	// The VNIC must be in the same subnet as the current VNIC.
	VnicId *string `mandatory:"false" json:"vnicId"`
}

func (m UpdateIpv6Details) String() string {
	return common.PointerString(m)
}
