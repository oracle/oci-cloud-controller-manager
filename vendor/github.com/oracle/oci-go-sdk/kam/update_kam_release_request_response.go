// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package kam

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// UpdateKamReleaseRequest wrapper for the UpdateKamRelease operation
type UpdateKamReleaseRequest struct {

	// The OCID of the cluster.
	ClusterId *string `mandatory:"true" contributesTo:"query" name:"clusterId"`

	// The OCID of the kam release.
	ReleaseId *string `mandatory:"true" contributesTo:"path" name:"releaseId"`

	// Details of an upgrade request. The KAM chart OCID can be looked up
	// using the ListKamCharts API
	UpdateKamReleaseDetails `contributesTo:"body"`

	// A token you supply to uniquely identify the request and provide idempotency if
	// the request is retried. Idempotency tokens expire after 24 hours.
	OpcRetryToken *string `mandatory:"false" contributesTo:"header" name:"opc-retry-token"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request UpdateKamReleaseRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request UpdateKamReleaseRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request UpdateKamReleaseRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// UpdateKamReleaseResponse wrapper for the UpdateKamRelease operation
type UpdateKamReleaseResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// Unique Oracle-assigned identifier for the asynchronous request. You can use this to query status of the asynchronous operation.
	OpcWorkRequestId *string `presentIn:"header" name:"opc-work-request-id"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response UpdateKamReleaseResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response UpdateKamReleaseResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
