// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package usage

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetSubscriptionInfoRequest wrapper for the GetSubscriptionInfo operation
type GetSubscriptionInfoRequest struct {

	// The OCID of the tenancy for which usage data is fetched.
	TenancyId *string `mandatory:"true" contributesTo:"path" name:"tenancyId"`

	// Unique, Oracle-assigned identifier for the request. If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetSubscriptionInfoRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetSubscriptionInfoRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetSubscriptionInfoRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// GetSubscriptionInfoResponse wrapper for the GetSubscriptionInfo operation
type GetSubscriptionInfoResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The SubscriptionInfo instance
	SubscriptionInfo `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetSubscriptionInfoResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetSubscriptionInfoResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
