// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package batch

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListJobLogsRequest wrapper for the ListJobLogs operation
type ListJobLogsRequest struct {

	// The OCID of the Job.
	JobId *string `mandatory:"true" contributesTo:"path" name:"jobId"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide
	// the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// A filter that returns only logs of a specified id.
	// Log Id is consist of 3 parts: Pod name, Pod namespace and container id,
	// and use '_' to connect the 3 parts according to kuberbetes naming convention.
	LogId *string `mandatory:"false" contributesTo:"query" name:"logId"`

	// The field to sort by. You can provide one sort order (`sortOrder`).
	// Default order for TIMECREATED is descending. Default order for DISPLAYNAME is ascending.
	// The DISPLAYNAME sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListJobs`) let you
	// optionally filter by availability domain if the scope of the resource type
	// is within a single availability domain. If you call one of these "List" operations
	// without specifying an availability domain, the resources are grouped by availability domain,
	// then sorted.
	SortBy ListJobLogsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`).
	// The DISPLAYNAME sort order is case sensitive.
	SortOrder ListJobLogsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// For list pagination. The maximum number of results per page, or items to
	// return in a paginated "List" call.
	// For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from
	// the previous "List" call. For important details about how pagination works,
	// see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListJobLogsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListJobLogsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListJobLogsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListJobLogsResponse wrapper for the ListJobLogs operation
type ListJobLogsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []LogSummary instances
	Items []LogSummary `presentIn:"body"`

	// Unique identifier for the request.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// Retrieves the next page of paginated list items. If the `opc-next-page`
	// header appears in the response, additional pages of results remain.
	// To receive the next page, include the header value in the `page` param.
	// If the `opc-next-page` header does not appear in the response, there
	// are no more list items to get. For more information about list pagination,
	// see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListJobLogsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListJobLogsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListJobLogsSortByEnum Enum with underlying type: string
type ListJobLogsSortByEnum string

// Set of constants representing the allowable values for ListJobLogsSortByEnum
const (
	ListJobLogsSortByTimecreated ListJobLogsSortByEnum = "TIMECREATED"
	ListJobLogsSortByDisplayname ListJobLogsSortByEnum = "DISPLAYNAME"
)

var mappingListJobLogsSortBy = map[string]ListJobLogsSortByEnum{
	"TIMECREATED": ListJobLogsSortByTimecreated,
	"DISPLAYNAME": ListJobLogsSortByDisplayname,
}

// GetListJobLogsSortByEnumValues Enumerates the set of values for ListJobLogsSortByEnum
func GetListJobLogsSortByEnumValues() []ListJobLogsSortByEnum {
	values := make([]ListJobLogsSortByEnum, 0)
	for _, v := range mappingListJobLogsSortBy {
		values = append(values, v)
	}
	return values
}

// ListJobLogsSortOrderEnum Enum with underlying type: string
type ListJobLogsSortOrderEnum string

// Set of constants representing the allowable values for ListJobLogsSortOrderEnum
const (
	ListJobLogsSortOrderAsc  ListJobLogsSortOrderEnum = "ASC"
	ListJobLogsSortOrderDesc ListJobLogsSortOrderEnum = "DESC"
)

var mappingListJobLogsSortOrder = map[string]ListJobLogsSortOrderEnum{
	"ASC":  ListJobLogsSortOrderAsc,
	"DESC": ListJobLogsSortOrderDesc,
}

// GetListJobLogsSortOrderEnumValues Enumerates the set of values for ListJobLogsSortOrderEnum
func GetListJobLogsSortOrderEnumValues() []ListJobLogsSortOrderEnum {
	values := make([]ListJobLogsSortOrderEnum, 0)
	for _, v := range mappingListJobLogsSortOrder {
		values = append(values, v)
	}
	return values
}
