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

// PrivateAccessGatewaySummary A summary of private access gateway (PAG) information. This object is returned when listing
// PAGs.
type PrivateAccessGatewaySummary struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the PAG.
	Id *string `mandatory:"true" json:"id"`

	// The PAG's friendly name.
	DisplayName *string `mandatory:"true" json:"displayName"`
}

func (m PrivateAccessGatewaySummary) String() string {
	return common.PointerString(m)
}
