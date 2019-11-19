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

// ReverseConnectionNatIp Current allocation of NAT IP address for a specific customer IP address.
// For service providers to establish a reverse connection to a customer IP address,
// reverse connection NAT IP address should be used as the destination.
type ReverseConnectionNatIp struct {

	// The Reverse Connection NAT IP's current state.
	LifecycleState ReverseConnectionNatIpLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the Reverse Connection NAT IP was created, in the format defined by RFC3339.
	// Example: '2016-08-25T21:10:29.600Z'
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// Customer's IP address to which the reverse connection is going to be established.
	ReverseConnectionCustomerIp *string `mandatory:"true" json:"reverseConnectionCustomerIp"`

	// The Reverse Connection NAT IP address associated with the customer IP and peId.
	ReverseConnectionNatIp *string `mandatory:"true" json:"reverseConnectionNatIp"`

	// The OCID of the customer's Private Endpoint associated with the Reverse Connection.
	PrivateEndpointId *string `mandatory:"true" json:"privateEndpointId"`
}

func (m ReverseConnectionNatIp) String() string {
	return common.PointerString(m)
}

// ReverseConnectionNatIpLifecycleStateEnum Enum with underlying type: string
type ReverseConnectionNatIpLifecycleStateEnum string

// Set of constants representing the allowable values for ReverseConnectionNatIpLifecycleStateEnum
const (
	ReverseConnectionNatIpLifecycleStateProvisioning ReverseConnectionNatIpLifecycleStateEnum = "PROVISIONING"
	ReverseConnectionNatIpLifecycleStateAvailable    ReverseConnectionNatIpLifecycleStateEnum = "AVAILABLE"
	ReverseConnectionNatIpLifecycleStateTerminating  ReverseConnectionNatIpLifecycleStateEnum = "TERMINATING"
	ReverseConnectionNatIpLifecycleStateTerminated   ReverseConnectionNatIpLifecycleStateEnum = "TERMINATED"
)

var mappingReverseConnectionNatIpLifecycleState = map[string]ReverseConnectionNatIpLifecycleStateEnum{
	"PROVISIONING": ReverseConnectionNatIpLifecycleStateProvisioning,
	"AVAILABLE":    ReverseConnectionNatIpLifecycleStateAvailable,
	"TERMINATING":  ReverseConnectionNatIpLifecycleStateTerminating,
	"TERMINATED":   ReverseConnectionNatIpLifecycleStateTerminated,
}

// GetReverseConnectionNatIpLifecycleStateEnumValues Enumerates the set of values for ReverseConnectionNatIpLifecycleStateEnum
func GetReverseConnectionNatIpLifecycleStateEnumValues() []ReverseConnectionNatIpLifecycleStateEnum {
	values := make([]ReverseConnectionNatIpLifecycleStateEnum, 0)
	for _, v := range mappingReverseConnectionNatIpLifecycleState {
		values = append(values, v)
	}
	return values
}
