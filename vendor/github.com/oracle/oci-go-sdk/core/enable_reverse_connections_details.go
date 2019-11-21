// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
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
	"github.com/oracle/oci-go-sdk/common"
)

// EnableReverseConnectionsDetails Details for enabling reverse connections functionality on the Private Endpoint
type EnableReverseConnectionsDetails struct {

	// List of DNS zones to exclude from the default DNS resolution context.
	ExcludedDnsZones []string `mandatory:"true" json:"excludedDnsZones"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the service provider's subnet, in which DNS proxy endpoint will be spwaned.
	ServiceSubnetId *string `mandatory:"true" json:"serviceSubnetId"`

	// IP addresses in customer VCN which will be used as source IPs for reverse connections from the service provider's VCN to the customer VCN.
	// If no list is specified or an empty list is provided, one IP address will be picked from customer subnet's CIDR.
	ReverseConnectionsOriginIps []ReverseConnectionsSourceIpDetails `mandatory:"false" json:"reverseConnectionsOriginIps"`

	// IP address in service provider VCN that should be used as DNS proxy for resolving the DNS FQDN to destination IP addresses for reverse connections.
	// If not provided, some available IP address will be picked from service provider subnet's CIDR.
	DnsProxyIp *string `mandatory:"false" json:"dnsProxyIp"`

	// Context in which the DNS proxy will resolve the DNS queries in. The default is `SERVICE`.
	// Allowed values:
	//  * `SERVICE` : All DNS queries will be resolved within the service VCN's DNS context, unless the FQDN belongs to one of zones in the `excludedDnsZones` list.
	//  * `CUSTOMER` : All DNS queries will be resolved within the customer VCN's DNS context, unless the FQDN belongs to one of zones in the `excludedDnsZones` list.
	DefaultDnsResolutionContext EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum `mandatory:"false" json:"defaultDnsResolutionContext,omitempty"`
}

func (m EnableReverseConnectionsDetails) String() string {
	return common.PointerString(m)
}

// EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum Enum with underlying type: string
type EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum string

// Set of constants representing the allowable values for EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum
const (
	EnableReverseConnectionsDetailsDefaultDnsResolutionContextService  EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum = "SERVICE"
	EnableReverseConnectionsDetailsDefaultDnsResolutionContextCustomer EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum = "CUSTOMER"
)

var mappingEnableReverseConnectionsDetailsDefaultDnsResolutionContext = map[string]EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum{
	"SERVICE":  EnableReverseConnectionsDetailsDefaultDnsResolutionContextService,
	"CUSTOMER": EnableReverseConnectionsDetailsDefaultDnsResolutionContextCustomer,
}

// GetEnableReverseConnectionsDetailsDefaultDnsResolutionContextEnumValues Enumerates the set of values for EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum
func GetEnableReverseConnectionsDetailsDefaultDnsResolutionContextEnumValues() []EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum {
	values := make([]EnableReverseConnectionsDetailsDefaultDnsResolutionContextEnum, 0)
	for _, v := range mappingEnableReverseConnectionsDetailsDefaultDnsResolutionContext {
		values = append(values, v)
	}
	return values
}
