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

// HeaderFieldSpecification header in key/value pair
type HeaderFieldSpecification struct {

	// name of the header
	Name *string `mandatory:"false" json:"name"`

	// value of the header
	Value *string `mandatory:"false" json:"value"`
}

func (m HeaderFieldSpecification) String() string {
	return common.PointerString(m)
}
