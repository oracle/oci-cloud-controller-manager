// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package kam

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListKamChartsRequest wrapper for the ListKamCharts operation
type ListKamChartsRequest struct {

	// The OCID of the cluster.
	ClusterId *string `mandatory:"true" contributesTo:"query" name:"clusterId"`

	// The name to filter on
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

	// The package type to restrict results to, ADD_ON or APPLICATION or *
	PackageType *string `mandatory:"false" contributesTo:"query" name:"packageType"`

	// The maximum number of items to return in a paginated "List" call.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The optional order in which to sort the results.
	SortOrder ListKamChartsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the KAM chart serach results by.
	SortBy ListKamChartsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListKamChartsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListKamChartsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListKamChartsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListKamChartsResponse wrapper for the ListKamCharts operation
type ListKamChartsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []KamChartSummary instances
	Items []KamChartSummary `presentIn:"body"`

	// For pagination of a list of items. When paging through a list, if this header appears in the response,
	// then a partial list might have been returned. Include this value as the `page` parameter for the
	// subsequent GET request to get the next batch of items.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListKamChartsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListKamChartsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListKamChartsSortOrderEnum Enum with underlying type: string
type ListKamChartsSortOrderEnum string

// Set of constants representing the allowable values for ListKamChartsSortOrderEnum
const (
	ListKamChartsSortOrderAsc  ListKamChartsSortOrderEnum = "ASC"
	ListKamChartsSortOrderDesc ListKamChartsSortOrderEnum = "DESC"
)

var mappingListKamChartsSortOrder = map[string]ListKamChartsSortOrderEnum{
	"ASC":  ListKamChartsSortOrderAsc,
	"DESC": ListKamChartsSortOrderDesc,
}

// GetListKamChartsSortOrderEnumValues Enumerates the set of values for ListKamChartsSortOrderEnum
func GetListKamChartsSortOrderEnumValues() []ListKamChartsSortOrderEnum {
	values := make([]ListKamChartsSortOrderEnum, 0)
	for _, v := range mappingListKamChartsSortOrder {
		values = append(values, v)
	}
	return values
}

// ListKamChartsSortByEnum Enum with underlying type: string
type ListKamChartsSortByEnum string

// Set of constants representing the allowable values for ListKamChartsSortByEnum
const (
	ListKamChartsSortByName    ListKamChartsSortByEnum = "NAME"
	ListKamChartsSortByType    ListKamChartsSortByEnum = "TYPE"
	ListKamChartsSortByVersion ListKamChartsSortByEnum = "VERSION"
)

var mappingListKamChartsSortBy = map[string]ListKamChartsSortByEnum{
	"NAME":    ListKamChartsSortByName,
	"TYPE":    ListKamChartsSortByType,
	"VERSION": ListKamChartsSortByVersion,
}

// GetListKamChartsSortByEnumValues Enumerates the set of values for ListKamChartsSortByEnum
func GetListKamChartsSortByEnumValues() []ListKamChartsSortByEnum {
	values := make([]ListKamChartsSortByEnum, 0)
	for _, v := range mappingListKamChartsSortBy {
		values = append(values, v)
	}
	return values
}
