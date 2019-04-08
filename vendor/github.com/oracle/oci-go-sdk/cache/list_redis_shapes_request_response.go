// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package cache

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListRedisShapesRequest wrapper for the ListRedisShapes operation
type ListRedisShapesRequest struct {

	// The OCID of the compartment for which to list the Redis shapes.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The order of sorting (ASC or DESC).
	SortOrder ListRedisShapesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The value of the opc-next-page response header from the previous request.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The maximum number of items to return.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// A unique Oracle-assigned identifier for the request. If you need
	// to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListRedisShapesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListRedisShapesRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListRedisShapesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListRedisShapesResponse wrapper for the ListRedisShapes operation
type ListRedisShapesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []RedisShapeSummary instances
	Items []RedisShapeSummary `presentIn:"body"`

	// A unique Oracle-assigned identifier for the request. If
	// you need to contact Oracle about a particular request, please provide
	// the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// The next page value to provide for the page header in the next request.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListRedisShapesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListRedisShapesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListRedisShapesSortOrderEnum Enum with underlying type: string
type ListRedisShapesSortOrderEnum string

// Set of constants representing the allowable values for ListRedisShapesSortOrderEnum
const (
	ListRedisShapesSortOrderAsc  ListRedisShapesSortOrderEnum = "ASC"
	ListRedisShapesSortOrderDesc ListRedisShapesSortOrderEnum = "DESC"
)

var mappingListRedisShapesSortOrder = map[string]ListRedisShapesSortOrderEnum{
	"ASC":  ListRedisShapesSortOrderAsc,
	"DESC": ListRedisShapesSortOrderDesc,
}

// GetListRedisShapesSortOrderEnumValues Enumerates the set of values for ListRedisShapesSortOrderEnum
func GetListRedisShapesSortOrderEnumValues() []ListRedisShapesSortOrderEnum {
	values := make([]ListRedisShapesSortOrderEnum, 0)
	for _, v := range mappingListRedisShapesSortOrder {
		values = append(values, v)
	}
	return values
}
