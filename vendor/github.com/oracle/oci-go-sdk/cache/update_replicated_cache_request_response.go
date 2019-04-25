// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package cache

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// UpdateReplicatedCacheRequest wrapper for the UpdateReplicatedCache operation
type UpdateReplicatedCacheRequest struct {

	// Input parameters that are used to update the Redis replicated cache.
	UpdateReplicatedCacheDetails `contributesTo:"body"`

	// The OCID that uniquely identifies the Redis replicated cache.
	Id *string `mandatory:"true" contributesTo:"path" name:"id"`

	// A unique Oracle-assigned identifier for the request. If you need
	// to contact Oracle about a particular request, please provide the request
	// ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Use the `if-match` parameter to use optimistic concurrency control. In the `PUT` or `DELETE` call
	// for a resource, set the `if-match` parameter to the value of the `etag`
	// from a previous `GET` or `POST` response for that resource. The resource
	// is updated or deleted only if the `etag` matches the resource's
	// current `etag` value.
	IfMatch *string `mandatory:"false" contributesTo:"header" name:"If-Match"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request UpdateReplicatedCacheRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request UpdateReplicatedCacheRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request UpdateReplicatedCacheRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// UpdateReplicatedCacheResponse wrapper for the UpdateReplicatedCache operation
type UpdateReplicatedCacheResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A unique Oracle-assigned identifier for the request. If
	// you need to contact Oracle about a particular request, please provide
	// the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For optimistic concurrency control, See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`
}

func (response UpdateReplicatedCacheResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response UpdateReplicatedCacheResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
