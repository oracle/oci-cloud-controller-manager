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

// WorkRequestLogCollection Collection of the Work Request logs
type WorkRequestLogCollection struct {

	// Work Request logs
	Items []WorkRequestLog `mandatory:"true" json:"items"`
}

func (m WorkRequestLogCollection) String() string {
	return common.PointerString(m)
}
