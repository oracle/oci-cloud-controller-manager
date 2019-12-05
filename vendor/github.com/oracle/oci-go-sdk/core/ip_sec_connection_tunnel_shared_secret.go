// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"github.com/oracle/oci-go-sdk/common"
)

// IpSecConnectionTunnelSharedSecret The tunnel's shared secret (pre-shared key).
type IpSecConnectionTunnelSharedSecret struct {

	// The tunnel's shared secret (pre-shared key).
	// Example: `EXAMPLEToUis6j1c.p8G.dVQxcmdfMO0yXMLi.lZTbYCMDGu4V8o`
	SharedSecret *string `mandatory:"true" json:"sharedSecret"`
}

func (m IpSecConnectionTunnelSharedSecret) String() string {
	return common.PointerString(m)
}
