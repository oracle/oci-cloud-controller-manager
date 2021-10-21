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

// ClientVpnSummary a summary of ClientVpn.
type ClientVpnSummary struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment that user sent request.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of clientVPNEndpoint.
	Id *string `mandatory:"false" json:"id"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// A limit that allows the maximum number of VPN concurrent connections.
	MaxConnections *int `mandatory:"false" json:"maxConnections"`

	// The current state of the ClientVpn.
	LifecycleState ClientVpnLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// A subnet for openVPN clients to access servers.
	ClientSubnetCidr *string `mandatory:"false" json:"clientSubnetCidr"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the attachedSubnet (VNIC) in customer tenancy.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// The IP address in attached subnet.
	IpAddressInAttachedSubnet *string `mandatory:"false" json:"ipAddressInAttachedSubnet"`

	// Whether re-route Internet traffic or not.
	IsRerouteEnabled *bool `mandatory:"false" json:"isRerouteEnabled"`

	// The date and time the ClientVpn was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Allowed values:
	//   * `NAT`: NAT mode supports the one-way access. In NAT mode, client can access the Internet from server endpoint
	//   but server endpoint cannot access the Internet from client.
	//   * `ROUTING`: ROUTING mode supports the two-way access. In ROUTING mode, client and server endpoint can access the
	//   Internet to each other.
	AddressingMode ClientVpnSummaryAddressingModeEnum `mandatory:"false" json:"addressingMode,omitempty"`

	// Allowed values:
	//   * `LOCAL`: Local authentication mode that applies users and password to get authentication through the server.
	//   * `RADIUS`: RADIUS authentication mode applies users and password to get authentication through the RADIUS server.
	//   * `LDAP`: LDAP authentication mode that applies users and passwords to get authentication through the LDAP server.
	AuthenticationMode ClientVpnSummaryAuthenticationModeEnum `mandatory:"false" json:"authenticationMode,omitempty"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m ClientVpnSummary) String() string {
	return common.PointerString(m)
}

// ClientVpnSummaryAddressingModeEnum Enum with underlying type: string
type ClientVpnSummaryAddressingModeEnum string

// Set of constants representing the allowable values for ClientVpnSummaryAddressingModeEnum
const (
	ClientVpnSummaryAddressingModeNat     ClientVpnSummaryAddressingModeEnum = "NAT"
	ClientVpnSummaryAddressingModeRouting ClientVpnSummaryAddressingModeEnum = "ROUTING"
)

var mappingClientVpnSummaryAddressingMode = map[string]ClientVpnSummaryAddressingModeEnum{
	"NAT":     ClientVpnSummaryAddressingModeNat,
	"ROUTING": ClientVpnSummaryAddressingModeRouting,
}

// GetClientVpnSummaryAddressingModeEnumValues Enumerates the set of values for ClientVpnSummaryAddressingModeEnum
func GetClientVpnSummaryAddressingModeEnumValues() []ClientVpnSummaryAddressingModeEnum {
	values := make([]ClientVpnSummaryAddressingModeEnum, 0)
	for _, v := range mappingClientVpnSummaryAddressingMode {
		values = append(values, v)
	}
	return values
}

// ClientVpnSummaryAuthenticationModeEnum Enum with underlying type: string
type ClientVpnSummaryAuthenticationModeEnum string

// Set of constants representing the allowable values for ClientVpnSummaryAuthenticationModeEnum
const (
	ClientVpnSummaryAuthenticationModeLocal  ClientVpnSummaryAuthenticationModeEnum = "LOCAL"
	ClientVpnSummaryAuthenticationModeRadius ClientVpnSummaryAuthenticationModeEnum = "RADIUS"
	ClientVpnSummaryAuthenticationModeLdap   ClientVpnSummaryAuthenticationModeEnum = "LDAP"
)

var mappingClientVpnSummaryAuthenticationMode = map[string]ClientVpnSummaryAuthenticationModeEnum{
	"LOCAL":  ClientVpnSummaryAuthenticationModeLocal,
	"RADIUS": ClientVpnSummaryAuthenticationModeRadius,
	"LDAP":   ClientVpnSummaryAuthenticationModeLdap,
}

// GetClientVpnSummaryAuthenticationModeEnumValues Enumerates the set of values for ClientVpnSummaryAuthenticationModeEnum
func GetClientVpnSummaryAuthenticationModeEnumValues() []ClientVpnSummaryAuthenticationModeEnum {
	values := make([]ClientVpnSummaryAuthenticationModeEnum, 0)
	for _, v := range mappingClientVpnSummaryAuthenticationMode {
		values = append(values, v)
	}
	return values
}
