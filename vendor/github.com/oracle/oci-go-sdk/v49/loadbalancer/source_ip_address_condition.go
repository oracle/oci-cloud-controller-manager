// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Load Balancing API
//
// API for the Load Balancing service. Use this API to manage load balancers, backend sets, and related items. For more
// information, see Overview of Load Balancing (https://docs.cloud.oracle.com/iaas/Content/Balance/Concepts/balanceoverview.htm).
//

package loadbalancer

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/v49/common"
)

// SourceIpAddressCondition A rule condition that checks client source IP against specified IP address or address range.
// The IP Address can be specified either as a single CIDR by value or referenced via a named CidrBlocks.
// If latter is used, then CidrBlocks must be created in the context of this Load Balancer.
// Condition evaluation depends on operator:
// *  **IN_RANGE** - The condition is determined to be true if the source IP address belongs to the CIDR or CidrBlocks referenced in *attributeValue*.
// *  **NOT_IN_RANGE** - The condition is determined to be true if the source IP address does not belong to the CIDR or CidrBlocks referenced in *attributeValue*.
type SourceIpAddressCondition struct {

	// An IPv4 or IPv6 address range that the source IP address of an incoming packet must match.
	// The service accepts only classless inter-domain routing (CIDR) format (x.x.x.x/y or x:x::x/y) strings.
	// Specify 0.0.0.0/0 or ::/0 to match all incoming traffic.
	// Besides IP ranges or IPs you can match against multiple CIDR blocks by creating a CidrBlocks resource with
	// a list of CIDR blocks and mentioning the name of CidrBlocks resource.
	// example: "192.168.0.0/16 or MySourceIPCidrBlocks"
	AttributeValue *string `mandatory:"true" json:"attributeValue"`

	// Operator which has to be appplied to this condition.
	// *  **IN_RANGE** - The condition is determined to be true if the source IP address belongs to the CIDR or CidrBlocks referenced in *attributeValue*.
	// *  **NOT_IN_RANGE** - The condition is determined to be true if the source IP address does not belong to the CIDR or CidrBlocks referenced in *attributeValue*.
	Operator SourceIpAddressConditionOperatorEnum `mandatory:"false" json:"operator,omitempty"`
}

func (m SourceIpAddressCondition) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m SourceIpAddressCondition) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeSourceIpAddressCondition SourceIpAddressCondition
	s := struct {
		DiscriminatorParam string `json:"attributeName"`
		MarshalTypeSourceIpAddressCondition
	}{
		"SOURCE_IP_ADDRESS",
		(MarshalTypeSourceIpAddressCondition)(m),
	}

	return json.Marshal(&s)
}

// SourceIpAddressConditionOperatorEnum Enum with underlying type: string
type SourceIpAddressConditionOperatorEnum string

// Set of constants representing the allowable values for SourceIpAddressConditionOperatorEnum
const (
	SourceIpAddressConditionOperatorInRange    SourceIpAddressConditionOperatorEnum = "IN_RANGE"
	SourceIpAddressConditionOperatorNotInRange SourceIpAddressConditionOperatorEnum = "NOT_IN_RANGE"
)

var mappingSourceIpAddressConditionOperator = map[string]SourceIpAddressConditionOperatorEnum{
	"IN_RANGE":     SourceIpAddressConditionOperatorInRange,
	"NOT_IN_RANGE": SourceIpAddressConditionOperatorNotInRange,
}

// GetSourceIpAddressConditionOperatorEnumValues Enumerates the set of values for SourceIpAddressConditionOperatorEnum
func GetSourceIpAddressConditionOperatorEnumValues() []SourceIpAddressConditionOperatorEnum {
	values := make([]SourceIpAddressConditionOperatorEnum, 0)
	for _, v := range mappingSourceIpAddressConditionOperator {
		values = append(values, v)
	}
	return values
}
