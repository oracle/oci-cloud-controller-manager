// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package marketplace

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListListingsRequest wrapper for the ListListings operation
type ListListingsRequest struct {

	// The name of the listing.
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

	// The unique identifier of the listing.
	ListingId *string `mandatory:"false" contributesTo:"query" name:"listingId"`

	// Limit listings to just this publisher.
	PublisherId *string `mandatory:"false" contributesTo:"query" name:"publisherId"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a particular request,
	// please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// How many records to return. Specify a value greater than zero and less than or equal to 1000. The default is 30.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The field that is used to sort listed results. You can only specify one field to sort by.
	// `TIMERELEASED` displays results in descending order by default. `NAME` displays results in
	// ascending order by default. You can change your preference by specifying a different sort order.
	SortBy ListListingsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either 'asc' or 'desc'.
	SortOrder ListListingsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListListingsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListListingsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListListingsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListListingsResponse wrapper for the ListListings operation
type ListListingsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ListingSummary instances
	Items []ListingSummary `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For list pagination. When this header appears in the response, additional pages of results remain.
	// Include this value as the `page` parameter for the subsequent GET request. For important details about
	// how pagination works, see List Pagination (https://docs.cloud.oracle.com/Content/API/Concepts/usingapi.htm#List_Pagination).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListListingsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListListingsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListListingsSortByEnum Enum with underlying type: string
type ListListingsSortByEnum string

// Set of constants representing the allowable values for ListListingsSortByEnum
const (
	ListListingsSortByName         ListListingsSortByEnum = "NAME"
	ListListingsSortByTimereleased ListListingsSortByEnum = "TIMERELEASED"
)

var mappingListListingsSortBy = map[string]ListListingsSortByEnum{
	"NAME":         ListListingsSortByName,
	"TIMERELEASED": ListListingsSortByTimereleased,
}

// GetListListingsSortByEnumValues Enumerates the set of values for ListListingsSortByEnum
func GetListListingsSortByEnumValues() []ListListingsSortByEnum {
	values := make([]ListListingsSortByEnum, 0)
	for _, v := range mappingListListingsSortBy {
		values = append(values, v)
	}
	return values
}

// ListListingsSortOrderEnum Enum with underlying type: string
type ListListingsSortOrderEnum string

// Set of constants representing the allowable values for ListListingsSortOrderEnum
const (
	ListListingsSortOrderAsc  ListListingsSortOrderEnum = "ASC"
	ListListingsSortOrderDesc ListListingsSortOrderEnum = "DESC"
)

var mappingListListingsSortOrder = map[string]ListListingsSortOrderEnum{
	"ASC":  ListListingsSortOrderAsc,
	"DESC": ListListingsSortOrderDesc,
}

// GetListListingsSortOrderEnumValues Enumerates the set of values for ListListingsSortOrderEnum
func GetListListingsSortOrderEnumValues() []ListListingsSortOrderEnum {
	values := make([]ListListingsSortOrderEnum, 0)
	for _, v := range mappingListListingsSortOrder {
		values = append(values, v)
	}
	return values
}
