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
	"encoding/json"
	"github.com/oracle/oci-go-sdk/v49/common"
)

// UpdateClientVpnDetails A request to update clientVpn.
type UpdateClientVpnDetails struct {

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// A subnet for openVPN clients to access servers. Default is 172.27.224.0/20
	ClientSubnetCidr *string `mandatory:"false" json:"clientSubnetCidr"`

	// A list of accessible subnets from this clientVpnEnpoint.
	AccessibleSubnetCidrs []string `mandatory:"false" json:"accessibleSubnetCidrs"`

	// Whether re-route Internet traffic or not.
	IsRerouteEnabled *bool `mandatory:"false" json:"isRerouteEnabled"`

	// Allowed values:
	//   * `NAT`: NAT mode supports one-way access. In NAT mode, client can access the Internet from server endpoint
	//   but server endpoint cannot access the Internet from client.
	//   * `ROUTING`: ROUTING mode supports two-way access. In ROUTING mode, client and server endpoint can access the
	//   Internet to each other.
	AddressingMode UpdateClientVpnDetailsAddressingModeEnum `mandatory:"false" json:"addressingMode,omitempty"`

	RadiusConfig *RadiusConfigDetails `mandatory:"false" json:"radiusConfig"`

	LdapConfig *LdapConfigDetails `mandatory:"false" json:"ldapConfig"`

	DnsConfig *DnsConfigDetails `mandatory:"false" json:"dnsConfig"`

	SslCert SslCertDetails `mandatory:"false" json:"sslCert"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`
}

func (m UpdateClientVpnDetails) String() string {
	return common.PointerString(m)
}

// UnmarshalJSON unmarshals from json
func (m *UpdateClientVpnDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DisplayName           *string                                  `json:"displayName"`
		ClientSubnetCidr      *string                                  `json:"clientSubnetCidr"`
		AccessibleSubnetCidrs []string                                 `json:"accessibleSubnetCidrs"`
		IsRerouteEnabled      *bool                                    `json:"isRerouteEnabled"`
		AddressingMode        UpdateClientVpnDetailsAddressingModeEnum `json:"addressingMode"`
		RadiusConfig          *RadiusConfigDetails                     `json:"radiusConfig"`
		LdapConfig            *LdapConfigDetails                       `json:"ldapConfig"`
		DnsConfig             *DnsConfigDetails                        `json:"dnsConfig"`
		SslCert               sslcertdetails                           `json:"sslCert"`
		DefinedTags           map[string]map[string]interface{}        `json:"definedTags"`
		FreeformTags          map[string]string                        `json:"freeformTags"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DisplayName = model.DisplayName

	m.ClientSubnetCidr = model.ClientSubnetCidr

	m.AccessibleSubnetCidrs = make([]string, len(model.AccessibleSubnetCidrs))
	for i, n := range model.AccessibleSubnetCidrs {
		m.AccessibleSubnetCidrs[i] = n
	}

	m.IsRerouteEnabled = model.IsRerouteEnabled

	m.AddressingMode = model.AddressingMode

	m.RadiusConfig = model.RadiusConfig

	m.LdapConfig = model.LdapConfig

	m.DnsConfig = model.DnsConfig

	nn, e = model.SslCert.UnmarshalPolymorphicJSON(model.SslCert.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.SslCert = nn.(SslCertDetails)
	} else {
		m.SslCert = nil
	}

	m.DefinedTags = model.DefinedTags

	m.FreeformTags = model.FreeformTags

	return
}

// UpdateClientVpnDetailsAddressingModeEnum Enum with underlying type: string
type UpdateClientVpnDetailsAddressingModeEnum string

// Set of constants representing the allowable values for UpdateClientVpnDetailsAddressingModeEnum
const (
	UpdateClientVpnDetailsAddressingModeNat     UpdateClientVpnDetailsAddressingModeEnum = "NAT"
	UpdateClientVpnDetailsAddressingModeRouting UpdateClientVpnDetailsAddressingModeEnum = "ROUTING"
)

var mappingUpdateClientVpnDetailsAddressingMode = map[string]UpdateClientVpnDetailsAddressingModeEnum{
	"NAT":     UpdateClientVpnDetailsAddressingModeNat,
	"ROUTING": UpdateClientVpnDetailsAddressingModeRouting,
}

// GetUpdateClientVpnDetailsAddressingModeEnumValues Enumerates the set of values for UpdateClientVpnDetailsAddressingModeEnum
func GetUpdateClientVpnDetailsAddressingModeEnumValues() []UpdateClientVpnDetailsAddressingModeEnum {
	values := make([]UpdateClientVpnDetailsAddressingModeEnum, 0)
	for _, v := range mappingUpdateClientVpnDetailsAddressingMode {
		values = append(values, v)
	}
	return values
}
