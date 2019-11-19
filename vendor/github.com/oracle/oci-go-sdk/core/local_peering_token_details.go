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

// LocalPeeringTokenDetails An object containing a generated peering token to be given to a peer who then accepts the token as part of the peering handshake process.
type LocalPeeringTokenDetails struct {

	// An opaque token to be shared with a peer.
	TokenForPeer *string `mandatory:"true" json:"tokenForPeer"`
}

func (m LocalPeeringTokenDetails) String() string {
	return common.PointerString(m)
}
