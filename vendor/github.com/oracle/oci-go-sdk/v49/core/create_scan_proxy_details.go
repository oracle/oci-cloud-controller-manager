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

// CreateScanProxyDetails Details for adding a RAC's scan listener information.
type CreateScanProxyDetails struct {

	// Type indicating whether Scan listener is specified by its FQDN or list of IPs
	ScanListenerType ScanProxyScanListenerTypeEnum `mandatory:"true" json:"scanListenerType"`

	// The FQDN/IPs and port information of customer's Real Application Cluster (RAC)'s SCAN
	// listeners.
	ScanListenerInfo []ScanListenerInfo `mandatory:"true" json:"scanListenerInfo"`

	// The protocol to be used for communication between client, scanProxy and RAC's scan
	// listeners
	Protocol ScanProxyProtocolEnum `mandatory:"false" json:"protocol,omitempty"`

	ScanListenerWallet *WalletInfo `mandatory:"false" json:"scanListenerWallet"`
}

func (m CreateScanProxyDetails) String() string {
	return common.PointerString(m)
}
