// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// API Gateway API
//
// API for the API Gateway service. Use this API to manage gateways, deployments, and related items.
//

package apigateway

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// AnonymousRouteAuthorizationPolicy For an ANONYMOUS type, an authenticated API must have an "isAnonymousAccessAllowed" property set to "true"
// in the authentication policy.
type AnonymousRouteAuthorizationPolicy struct {
}

func (m AnonymousRouteAuthorizationPolicy) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m AnonymousRouteAuthorizationPolicy) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeAnonymousRouteAuthorizationPolicy AnonymousRouteAuthorizationPolicy
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeAnonymousRouteAuthorizationPolicy
	}{
		"ANONYMOUS",
		(MarshalTypeAnonymousRouteAuthorizationPolicy)(m),
	}

	return json.Marshal(&s)
}
