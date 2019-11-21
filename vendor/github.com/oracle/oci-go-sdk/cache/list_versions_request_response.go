// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package cache

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListVersionsRequest wrapper for the ListVersions operation
type ListVersionsRequest struct {

	// The OCID of the compartment for which to list the Redis versions.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The order of sorting (ASC or DESC).
	SortOrder ListVersionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

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

func (request ListVersionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListVersionsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListVersionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListVersionsResponse wrapper for the ListVersions operation
type ListVersionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []RedisVersionSummary instances
	Items []RedisVersionSummary `presentIn:"body"`

	// A unique Oracle-assigned identifier for the request. If
	// you need to contact Oracle about a particular request, please provide
	// the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// The next page value to provide for the page header in the next request.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListVersionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListVersionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListVersionsSortOrderEnum Enum with underlying type: string
type ListVersionsSortOrderEnum string

// Set of constants representing the allowable values for ListVersionsSortOrderEnum
const (
	ListVersionsSortOrderAsc  ListVersionsSortOrderEnum = "ASC"
	ListVersionsSortOrderDesc ListVersionsSortOrderEnum = "DESC"
)

var mappingListVersionsSortOrder = map[string]ListVersionsSortOrderEnum{
	"ASC":  ListVersionsSortOrderAsc,
	"DESC": ListVersionsSortOrderDesc,
}

// GetListVersionsSortOrderEnumValues Enumerates the set of values for ListVersionsSortOrderEnum
func GetListVersionsSortOrderEnumValues() []ListVersionsSortOrderEnum {
	values := make([]ListVersionsSortOrderEnum, 0)
	for _, v := range mappingListVersionsSortOrder {
		values = append(values, v)
	}
	return values
}
