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

// AccessLogPolicy Configures the pushing of access logs to OCI Public Logging.
type AccessLogPolicy struct {

	// Enables pushing of access logs to OCI Public logging.
	IsEnabled *bool `mandatory:"false" json:"isEnabled"`
}

func (m AccessLogPolicy) String() string {
	return common.PointerString(m)
}
