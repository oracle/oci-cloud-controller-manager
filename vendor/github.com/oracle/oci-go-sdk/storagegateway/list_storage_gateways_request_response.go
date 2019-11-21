// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package storagegateway

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListStorageGatewaysRequest wrapper for the ListStorageGateways operation
type ListStorageGatewaysRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the relevant compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List" call. For important
	// details about how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated "List" call.
	// For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for TIMECREATED is descending.
	// Default order for DISPLAYNAME is ascending. The DISPLAYNAME sort order is case sensitive.
	SortBy ListStorageGatewaysSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order is case sensitive.
	SortOrder ListStorageGatewaysSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only resources that match the specified display name exactly.
	// Example: `example_resource_display_name`
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to return only resources that match the specified lifecycle state.
	// Example: `ACTIVE`
	LifecycleState LifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListStorageGatewaysRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListStorageGatewaysRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListStorageGatewaysRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListStorageGatewaysResponse wrapper for the ListStorageGateways operation
type ListStorageGatewaysResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []StorageGatewaySummary instances
	Items []StorageGatewaySummary `presentIn:"body"`

	// The unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For list pagination. When this header appears in the response, previous pages of results exist. When
	// paging through a list, provide this value as the `page` parameter for the subsequent request to page
	// backwards. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcPrevPage *string `presentIn:"header" name:"opc-prev-page"`

	// For list pagination. When this header appears in the response, additional pages of results remain.
	// Include this value as the `page` parameter for the subsequent GET request. For important details about
	// how pagination works, see List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListStorageGatewaysResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListStorageGatewaysResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListStorageGatewaysSortByEnum Enum with underlying type: string
type ListStorageGatewaysSortByEnum string

// Set of constants representing the allowable values for ListStorageGatewaysSortByEnum
const (
	ListStorageGatewaysSortByTimecreated ListStorageGatewaysSortByEnum = "TIMECREATED"
	ListStorageGatewaysSortByDisplayname ListStorageGatewaysSortByEnum = "DISPLAYNAME"
)

var mappingListStorageGatewaysSortBy = map[string]ListStorageGatewaysSortByEnum{
	"TIMECREATED": ListStorageGatewaysSortByTimecreated,
	"DISPLAYNAME": ListStorageGatewaysSortByDisplayname,
}

// GetListStorageGatewaysSortByEnumValues Enumerates the set of values for ListStorageGatewaysSortByEnum
func GetListStorageGatewaysSortByEnumValues() []ListStorageGatewaysSortByEnum {
	values := make([]ListStorageGatewaysSortByEnum, 0)
	for _, v := range mappingListStorageGatewaysSortBy {
		values = append(values, v)
	}
	return values
}

// ListStorageGatewaysSortOrderEnum Enum with underlying type: string
type ListStorageGatewaysSortOrderEnum string

// Set of constants representing the allowable values for ListStorageGatewaysSortOrderEnum
const (
	ListStorageGatewaysSortOrderAsc  ListStorageGatewaysSortOrderEnum = "ASC"
	ListStorageGatewaysSortOrderDesc ListStorageGatewaysSortOrderEnum = "DESC"
)

var mappingListStorageGatewaysSortOrder = map[string]ListStorageGatewaysSortOrderEnum{
	"ASC":  ListStorageGatewaysSortOrderAsc,
	"DESC": ListStorageGatewaysSortOrderDesc,
}

// GetListStorageGatewaysSortOrderEnumValues Enumerates the set of values for ListStorageGatewaysSortOrderEnum
func GetListStorageGatewaysSortOrderEnumValues() []ListStorageGatewaysSortOrderEnum {
	values := make([]ListStorageGatewaysSortOrderEnum, 0)
	for _, v := range mappingListStorageGatewaysSortOrder {
		values = append(values, v)
	}
	return values
}
