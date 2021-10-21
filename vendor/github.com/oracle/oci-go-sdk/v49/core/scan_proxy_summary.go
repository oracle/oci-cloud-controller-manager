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

// ScanProxySummary A summary of scan proxy information. This object is returned when listing scan proxies of a
// private endpoint.
type ScanProxySummary struct {

	// An Oracle-assigned unique identifier for a scanProxy within a privateEndpoint. You specify
	// this ID when you want to get or update or delete a scan ListScanProxiesProxy
	Id *string `mandatory:"true" json:"id"`

	// The FQDN/IPs and port information of customer's Real Application Cluster (RAC)'s SCAN
	// listeners.
	ScanListenerInfo []ScanListenerInfo `mandatory:"true" json:"scanListenerInfo"`

	// The port to which service DB client has to connect on scan proxy to initiate scan
	// connections.
	ScanProxyPort *int `mandatory:"true" json:"scanProxyPort"`

	// Type indicating whether Scan listener is specified by its FQDN or list of IPs
	ScanListenerType ScanProxyScanListenerTypeEnum `mandatory:"false" json:"scanListenerType,omitempty"`

	// The protocol used for communication between client, scanProxy and RAC's scan
	// listeners
	Protocol ScanProxyProtocolEnum `mandatory:"false" json:"protocol,omitempty"`

	ScanListenerWallet *WalletInfo `mandatory:"false" json:"scanListenerWallet"`

	// The scan proxy instance's current state.
	LifecycleState ScanProxyLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`
}

func (m ScanProxySummary) String() string {
	return common.PointerString(m)
}
