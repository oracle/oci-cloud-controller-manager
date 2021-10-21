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

// TunnelPhaseTwoDetails Tunnel Detail Information specific to IPSec Phase 2
type TunnelPhaseTwoDetails struct {

	// The total configured lifetime of an IKE security association
	Lifetime *int64 `mandatory:"false" json:"lifetime"`

	// The remaining lifetime before rekey
	RemainingLifetime *int64 `mandatory:"false" json:"remainingLifetime"`

	// The inbound Security Parameter Index (SPI) identification tag
	InboundToOracleSpi *string `mandatory:"false" json:"inboundToOracleSpi"`

	// The outbound Security Parameter Index (SPI) identification tag
	OutboundFromOracleSpi *string `mandatory:"false" json:"outboundFromOracleSpi"`

	// List of supported Phase Two authentication algorithms supported during tunnel negotiation.
	ProposedAuthenticationAlgorithms []string `mandatory:"false" json:"proposedAuthenticationAlgorithms"`

	// The negotiated Authentication Algorithm
	NegotiatedAuthenticationAlgorithm *string `mandatory:"false" json:"negotiatedAuthenticationAlgorithm"`

	// List of proposed Encryption Algorithms
	ProposedEncryptionAlgorithms []string `mandatory:"false" json:"proposedEncryptionAlgorithms"`

	// The negotiated Encryption Algorithm
	NegotiatedEncryptionAlgorithm *string `mandatory:"false" json:"negotiatedEncryptionAlgorithm"`

	// Proposed DH Group
	ProposedDhGroup *string `mandatory:"false" json:"proposedDhGroup"`

	// The negotiated DH Group
	NegotiatedDhGroup *string `mandatory:"false" json:"negotiatedDhGroup"`

	// ESP Phase 2 established
	IsEspEstablished *bool `mandatory:"false" json:"isEspEstablished"`

	// Is PFS (perfect forward secrecy) enabled
	IsPfsEnabled *bool `mandatory:"false" json:"isPfsEnabled"`

	// The date and time we retrieved the remaining lifetime, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	RemainingLifetimeLastRetrieved *common.SDKTime `mandatory:"false" json:"remainingLifetimeLastRetrieved"`
}

func (m TunnelPhaseTwoDetails) String() string {
	return common.PointerString(m)
}
