// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetReverseConnectionNatIpRequest wrapper for the GetReverseConnectionNatIp operation
type GetReverseConnectionNatIpRequest struct {

	// The IP address associated with a customer
	ReverseConnectionCustomerIp *string `mandatory:"true" contributesTo:"path" name:"reverseConnectionCustomerIp"`

	// The private endpoint's OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm).
	PrivateEndpointId *string `mandatory:"true" contributesTo:"path" name:"privateEndpointId"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetReverseConnectionNatIpRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetReverseConnectionNatIpRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetReverseConnectionNatIpRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// GetReverseConnectionNatIpResponse wrapper for the GetReverseConnectionNatIp operation
type GetReverseConnectionNatIpResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The ReverseConnectionNatIp instance
	ReverseConnectionNatIp `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetReverseConnectionNatIpResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetReverseConnectionNatIpResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
