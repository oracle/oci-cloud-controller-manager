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

// MigrationInfo Optional info that is present if a vnic is getting migrated as part of instance live migration
type MigrationInfo struct {

	// Live migration session ID
	MigrationSessionId *string `mandatory:"true" json:"migrationSessionId"`

	SourceSmartNicVnicAttachmentInfo *SmartNicVnicAttachmentInfo `mandatory:"true" json:"sourceSmartNicVnicAttachmentInfo"`

	DestinationSmartNicVnicAttachmentInfo *SmartNicVnicAttachmentInfo `mandatory:"true" json:"destinationSmartNicVnicAttachmentInfo"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VNIC on source hypervisor that will be used for hypervisor - VCN DP communication.
	SourceHypervisorVnicId *string `mandatory:"false" json:"sourceHypervisorVnicId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the VNIC on destination hypervisor that will be used for hypervisor - VCN DP communication.
	DestinationHypervisorVnicId *string `mandatory:"false" json:"destinationHypervisorVnicId"`

	// Indicates if VNIC's ingress traffic is routed to destination smart NIC
	IsTrafficRoutedToDestination *bool `mandatory:"false" json:"isTrafficRoutedToDestination"`
}

func (m MigrationInfo) String() string {
	return common.PointerString(m)
}
