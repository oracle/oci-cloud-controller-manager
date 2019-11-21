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

// ReverseConnectionConfiguration Reverse Connections Configuration details for the Private Endpoint.
// When reverse connections functionality is enabled, the Private Endpoint can allow reverse connections to be established to the customer VCN.
// Reverse connections will use different source IP addresses than the Private Endpoints' IP address.
type ReverseConnectionConfiguration struct {

	// The Reverse Connections Configuration's current state.
	LifecycleState ReverseConnectionConfigurationLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// IP addresses in customer VCN which will be used as source IPs for reverse connections from the service provider's VCN to the customer VCN.
	ReverseConnectionsSourceIps []ReverseConnectionsSourceIpDetails `mandatory:"false" json:"reverseConnectionsSourceIps"`

	// IP address in service provider VCN that should be used as DNS proxy for resolving DNS FQDN to the destination IP addresses for reverse connections.
	DnsProxyIp *string `mandatory:"false" json:"dnsProxyIp"`

	// Context in which the DNS proxy will resolve the DNS queries in. The default is `SERVICE`.
	// Allowed values:
	//  * `SERVICE` : All DNS queries will be resolved within the service VCN's DNS context, unless the FQDN belongs to one of zones in the `excludedDnsZones` list.
	//  * `CUSTOMER` : All DNS queries will be resolved within the customer VCN's DNS context, unless the FQDN belongs to one of zones in the `excludedDnsZones` list.
	DefaultDnsResolutionContext ReverseConnectionConfigurationDefaultDnsResolutionContextEnum `mandatory:"false" json:"defaultDnsResolutionContext,omitempty"`

	// List of DNS zones to exclude from the default DNS resolution context.
	ExcludedDnsZones []string `mandatory:"false" json:"excludedDnsZones"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the service provider's subnet, in which DNS proxy endpoint will be spwaned.
	ServiceSubnetId *string `mandatory:"false" json:"serviceSubnetId"`
}

func (m ReverseConnectionConfiguration) String() string {
	return common.PointerString(m)
}

// ReverseConnectionConfigurationLifecycleStateEnum Enum with underlying type: string
type ReverseConnectionConfigurationLifecycleStateEnum string

// Set of constants representing the allowable values for ReverseConnectionConfigurationLifecycleStateEnum
const (
	ReverseConnectionConfigurationLifecycleStateProvisioning ReverseConnectionConfigurationLifecycleStateEnum = "PROVISIONING"
	ReverseConnectionConfigurationLifecycleStateAvailable    ReverseConnectionConfigurationLifecycleStateEnum = "AVAILABLE"
	ReverseConnectionConfigurationLifecycleStateUpdating     ReverseConnectionConfigurationLifecycleStateEnum = "UPDATING"
	ReverseConnectionConfigurationLifecycleStateTerminating  ReverseConnectionConfigurationLifecycleStateEnum = "TERMINATING"
	ReverseConnectionConfigurationLifecycleStateTerminated   ReverseConnectionConfigurationLifecycleStateEnum = "TERMINATED"
	ReverseConnectionConfigurationLifecycleStateFailed       ReverseConnectionConfigurationLifecycleStateEnum = "FAILED"
)

var mappingReverseConnectionConfigurationLifecycleState = map[string]ReverseConnectionConfigurationLifecycleStateEnum{
	"PROVISIONING": ReverseConnectionConfigurationLifecycleStateProvisioning,
	"AVAILABLE":    ReverseConnectionConfigurationLifecycleStateAvailable,
	"UPDATING":     ReverseConnectionConfigurationLifecycleStateUpdating,
	"TERMINATING":  ReverseConnectionConfigurationLifecycleStateTerminating,
	"TERMINATED":   ReverseConnectionConfigurationLifecycleStateTerminated,
	"FAILED":       ReverseConnectionConfigurationLifecycleStateFailed,
}

// GetReverseConnectionConfigurationLifecycleStateEnumValues Enumerates the set of values for ReverseConnectionConfigurationLifecycleStateEnum
func GetReverseConnectionConfigurationLifecycleStateEnumValues() []ReverseConnectionConfigurationLifecycleStateEnum {
	values := make([]ReverseConnectionConfigurationLifecycleStateEnum, 0)
	for _, v := range mappingReverseConnectionConfigurationLifecycleState {
		values = append(values, v)
	}
	return values
}

// ReverseConnectionConfigurationDefaultDnsResolutionContextEnum Enum with underlying type: string
type ReverseConnectionConfigurationDefaultDnsResolutionContextEnum string

// Set of constants representing the allowable values for ReverseConnectionConfigurationDefaultDnsResolutionContextEnum
const (
	ReverseConnectionConfigurationDefaultDnsResolutionContextService  ReverseConnectionConfigurationDefaultDnsResolutionContextEnum = "SERVICE"
	ReverseConnectionConfigurationDefaultDnsResolutionContextCustomer ReverseConnectionConfigurationDefaultDnsResolutionContextEnum = "CUSTOMER"
)

var mappingReverseConnectionConfigurationDefaultDnsResolutionContext = map[string]ReverseConnectionConfigurationDefaultDnsResolutionContextEnum{
	"SERVICE":  ReverseConnectionConfigurationDefaultDnsResolutionContextService,
	"CUSTOMER": ReverseConnectionConfigurationDefaultDnsResolutionContextCustomer,
}

// GetReverseConnectionConfigurationDefaultDnsResolutionContextEnumValues Enumerates the set of values for ReverseConnectionConfigurationDefaultDnsResolutionContextEnum
func GetReverseConnectionConfigurationDefaultDnsResolutionContextEnumValues() []ReverseConnectionConfigurationDefaultDnsResolutionContextEnum {
	values := make([]ReverseConnectionConfigurationDefaultDnsResolutionContextEnum, 0)
	for _, v := range mappingReverseConnectionConfigurationDefaultDnsResolutionContext {
		values = append(values, v)
	}
	return values
}
