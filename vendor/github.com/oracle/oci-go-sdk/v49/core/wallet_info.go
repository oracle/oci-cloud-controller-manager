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

// WalletInfo Oracle Wallet that serves as a container to carry certificates, public key, private key,
// and list of CAs to be trusted
type WalletInfo struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the Vault Service secret, holding
	// the Oracle Wallet to be used for SCAN proxy.
	ScanWalletSecretId *string `mandatory:"false" json:"scanWalletSecretId"`

	// The version of secret that SCAN proxy should use to fetch the Oracle Wallet content for.
	ScanWalletSecretVersion *string `mandatory:"false" json:"scanWalletSecretVersion"`
}

func (m WalletInfo) String() string {
	return common.PointerString(m)
}
