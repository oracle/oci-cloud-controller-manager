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

// StockResponseBackend Send the request to a mocked backend
type StockResponseBackend struct {

	// the mocked response's status code
	Status *int `mandatory:"true" json:"status"`

	// the mocked response's body
	Body *string `mandatory:"false" json:"body"`

	// the mocked reponse's headers
	Headers []HeaderFieldSpecification `mandatory:"false" json:"headers"`
}

func (m StockResponseBackend) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m StockResponseBackend) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeStockResponseBackend StockResponseBackend
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeStockResponseBackend
	}{
		"STOCK_RESPONSE_BACKEND",
		(MarshalTypeStockResponseBackend)(m),
	}

	return json.Marshal(&s)
}
