// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// API Gateway API
//
// API for the API Gateway service. Use this API to manage gateways, deployments, and related items.
//

package apigateway

import (
	"github.com/oracle/oci-go-sdk/common"
)

// ChangeGatewayCompartmentDetails The new compartment details for the gateway
type ChangeGatewayCompartmentDetails struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment in which
	// resource is created.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`
}

func (m ChangeGatewayCompartmentDetails) String() string {
	return common.PointerString(m)
}
