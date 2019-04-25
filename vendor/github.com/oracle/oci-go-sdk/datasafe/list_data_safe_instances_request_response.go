// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package datasafe

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListDataSafeInstancesRequest wrapper for the ListDataSafeInstances operation
type ListDataSafeInstancesRequest struct {

	// The ID of the compartment in which to list resources.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Example: `My new resource`
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to return only resources that match the given lifecycle state exactly.
	LifecycleState DataSafeInstanceSummaryLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// The maximum number of items to return.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The page at which to start retrieving results.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The sort order to use, either 'asc' or 'desc'.
	SortOrder ListDataSafeInstancesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The field to sort by. Only one sort order may be provided. Default order for TIMECREATED is descending. Default order for DISPLAYNAME is ascending. If no value is specified TIMECREATED is default.
	SortBy ListDataSafeInstancesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The client request ID for tracing.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListDataSafeInstancesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListDataSafeInstancesRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListDataSafeInstancesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListDataSafeInstancesResponse wrapper for the ListDataSafeInstances operation
type ListDataSafeInstancesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []DataSafeInstanceSummary instances
	Items []DataSafeInstanceSummary `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If
	// you need to contact Oracle about a particular request,
	// please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For pagination of a list of `DataSafeInstance`s. If this header appears in the response, then this
	// is a partial list of data safe instances. Include this value as the `page` parameter in a subsequent
	// GET request to get the next batch of data safe nstances.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListDataSafeInstancesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListDataSafeInstancesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListDataSafeInstancesSortOrderEnum Enum with underlying type: string
type ListDataSafeInstancesSortOrderEnum string

// Set of constants representing the allowable values for ListDataSafeInstancesSortOrderEnum
const (
	ListDataSafeInstancesSortOrderAsc  ListDataSafeInstancesSortOrderEnum = "ASC"
	ListDataSafeInstancesSortOrderDesc ListDataSafeInstancesSortOrderEnum = "DESC"
)

var mappingListDataSafeInstancesSortOrder = map[string]ListDataSafeInstancesSortOrderEnum{
	"ASC":  ListDataSafeInstancesSortOrderAsc,
	"DESC": ListDataSafeInstancesSortOrderDesc,
}

// GetListDataSafeInstancesSortOrderEnumValues Enumerates the set of values for ListDataSafeInstancesSortOrderEnum
func GetListDataSafeInstancesSortOrderEnumValues() []ListDataSafeInstancesSortOrderEnum {
	values := make([]ListDataSafeInstancesSortOrderEnum, 0)
	for _, v := range mappingListDataSafeInstancesSortOrder {
		values = append(values, v)
	}
	return values
}

// ListDataSafeInstancesSortByEnum Enum with underlying type: string
type ListDataSafeInstancesSortByEnum string

// Set of constants representing the allowable values for ListDataSafeInstancesSortByEnum
const (
	ListDataSafeInstancesSortByTimecreated ListDataSafeInstancesSortByEnum = "TIMECREATED"
	ListDataSafeInstancesSortByDisplayname ListDataSafeInstancesSortByEnum = "DISPLAYNAME"
)

var mappingListDataSafeInstancesSortBy = map[string]ListDataSafeInstancesSortByEnum{
	"TIMECREATED": ListDataSafeInstancesSortByTimecreated,
	"DISPLAYNAME": ListDataSafeInstancesSortByDisplayname,
}

// GetListDataSafeInstancesSortByEnumValues Enumerates the set of values for ListDataSafeInstancesSortByEnum
func GetListDataSafeInstancesSortByEnumValues() []ListDataSafeInstancesSortByEnum {
	values := make([]ListDataSafeInstancesSortByEnum, 0)
	for _, v := range mappingListDataSafeInstancesSortBy {
		values = append(values, v)
	}
	return values
}
