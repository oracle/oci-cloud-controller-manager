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

// CaptureFilterRuleDetails This resource contains the rules governing what traffic a VTAP mirrors.
type CaptureFilterRuleDetails struct {

	// The traffic direction the VTAP is configured to mirror.
	TrafficDirection CaptureFilterRuleDetailsTrafficDirectionEnum `mandatory:"true" json:"trafficDirection"`

	// Include or exclude packets meeting this definition from mirrored traffic.
	RuleAction CaptureFilterRuleDetailsRuleActionEnum `mandatory:"false" json:"ruleAction,omitempty"`

	// Traffic from this CIDR block to the VTAP source will be mirrored to the VTAP target.
	SourceCidr *string `mandatory:"false" json:"sourceCidr"`

	// Traffic sent to this CIDR block through the VTAP source will be mirrored to the VTAP target.
	DestinationCidr *string `mandatory:"false" json:"destinationCidr"`

	// The transport protocol used in the filter. If do not choose a protocol, all protocols will be used in the filter.
	// Supported options are:
	//   * 1 = ICMP
	//   * 6 = TCP
	//   * 17 = UDP
	Protocol *string `mandatory:"false" json:"protocol"`

	IcmpOptions *IcmpOptions `mandatory:"false" json:"icmpOptions"`

	TcpOptions *TcpOptions `mandatory:"false" json:"tcpOptions"`

	UdpOptions *UdpOptions `mandatory:"false" json:"udpOptions"`
}

func (m CaptureFilterRuleDetails) String() string {
	return common.PointerString(m)
}

// CaptureFilterRuleDetailsTrafficDirectionEnum Enum with underlying type: string
type CaptureFilterRuleDetailsTrafficDirectionEnum string

// Set of constants representing the allowable values for CaptureFilterRuleDetailsTrafficDirectionEnum
const (
	CaptureFilterRuleDetailsTrafficDirectionIngress CaptureFilterRuleDetailsTrafficDirectionEnum = "INGRESS"
	CaptureFilterRuleDetailsTrafficDirectionEgress  CaptureFilterRuleDetailsTrafficDirectionEnum = "EGRESS"
)

var mappingCaptureFilterRuleDetailsTrafficDirection = map[string]CaptureFilterRuleDetailsTrafficDirectionEnum{
	"INGRESS": CaptureFilterRuleDetailsTrafficDirectionIngress,
	"EGRESS":  CaptureFilterRuleDetailsTrafficDirectionEgress,
}

// GetCaptureFilterRuleDetailsTrafficDirectionEnumValues Enumerates the set of values for CaptureFilterRuleDetailsTrafficDirectionEnum
func GetCaptureFilterRuleDetailsTrafficDirectionEnumValues() []CaptureFilterRuleDetailsTrafficDirectionEnum {
	values := make([]CaptureFilterRuleDetailsTrafficDirectionEnum, 0)
	for _, v := range mappingCaptureFilterRuleDetailsTrafficDirection {
		values = append(values, v)
	}
	return values
}

// CaptureFilterRuleDetailsRuleActionEnum Enum with underlying type: string
type CaptureFilterRuleDetailsRuleActionEnum string

// Set of constants representing the allowable values for CaptureFilterRuleDetailsRuleActionEnum
const (
	CaptureFilterRuleDetailsRuleActionInclude CaptureFilterRuleDetailsRuleActionEnum = "INCLUDE"
	CaptureFilterRuleDetailsRuleActionExclude CaptureFilterRuleDetailsRuleActionEnum = "EXCLUDE"
)

var mappingCaptureFilterRuleDetailsRuleAction = map[string]CaptureFilterRuleDetailsRuleActionEnum{
	"INCLUDE": CaptureFilterRuleDetailsRuleActionInclude,
	"EXCLUDE": CaptureFilterRuleDetailsRuleActionExclude,
}

// GetCaptureFilterRuleDetailsRuleActionEnumValues Enumerates the set of values for CaptureFilterRuleDetailsRuleActionEnum
func GetCaptureFilterRuleDetailsRuleActionEnumValues() []CaptureFilterRuleDetailsRuleActionEnum {
	values := make([]CaptureFilterRuleDetailsRuleActionEnum, 0)
	for _, v := range mappingCaptureFilterRuleDetailsRuleAction {
		values = append(values, v)
	}
	return values
}
