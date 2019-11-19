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

// ReverseConnectionsSourceIpDetails IP information for Reverse Connections Configuration, this will be a part of Private Endpoint object that is returned.
type ReverseConnectionsSourceIpDetails struct {

	// The IP that will be used as source for reverse connections.
	SourceIp *string `mandatory:"false" json:"sourceIp"`
}

func (m ReverseConnectionsSourceIpDetails) String() string {
	return common.PointerString(m)
}
