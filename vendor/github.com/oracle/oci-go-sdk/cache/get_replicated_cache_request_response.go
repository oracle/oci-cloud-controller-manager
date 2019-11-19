// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package cache

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetReplicatedCacheRequest wrapper for the GetReplicatedCache operation
type GetReplicatedCacheRequest struct {

	// OCID that uniquely identifies the Redis replicated cache.
	Id *string `mandatory:"true" contributesTo:"path" name:"id"`

	// A unique Oracle-assigned identifier for the request. If you need
	// to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetReplicatedCacheRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetReplicatedCacheRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetReplicatedCacheRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// GetReplicatedCacheResponse wrapper for the GetReplicatedCache operation
type GetReplicatedCacheResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The ReplicatedCache instance
	ReplicatedCache `presentIn:"body"`

	// A unique Oracle-assigned identifier for the request. If
	// you need to contact Oracle about a particular request, please provide
	// the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For optimistic concurrency control. See if-match.
	Etag *string `presentIn:"header" name:"etag"`
}

func (response GetReplicatedCacheResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetReplicatedCacheResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
