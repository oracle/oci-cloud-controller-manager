// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package functions

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListTriggersRequest wrapper for the ListTriggers operation
type ListTriggersRequest struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the application to which this function belongs.
	ApplicationId *string `mandatory:"true" contributesTo:"query" name:"applicationId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the function to which this trigger belongs.
	FunctionId *string `mandatory:"false" contributesTo:"query" name:"functionId"`

	// The maximum number of items to return. 1 is the minimum, 50 is the maximum.
	// Default: 10
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The pagination token for a list query returned by a previous operation
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// A filter to return only functions that match the lifecycle state in this parameter.
	// Example: `Creating`
	LifecycleState TriggerLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// A filter to return only triggers with display names that match the display name string. Matching is exact.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to return only triggers with the specified OCID.
	Id *string `mandatory:"false" contributesTo:"query" name:"id"`

	// Specifies sort order.
	// * **ASC:** Ascending sort order.
	// * **DESC:** Descending sort order.
	SortOrder ListTriggersSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Specifies the attribute with which to sort the rules.
	// Default: `timeCreated`
	// * **timeCreated:** Sorts by timeCreated.
	// * **displayName:** Sorts by displayName.
	// * **id:** Sorts by id.
	SortBy ListTriggersSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListTriggersRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListTriggersRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListTriggersRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListTriggersResponse wrapper for the ListTriggers operation
type ListTriggersResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []TriggerSummary instances
	Items []TriggerSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of
	// results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListTriggersResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListTriggersResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListTriggersSortOrderEnum Enum with underlying type: string
type ListTriggersSortOrderEnum string

// Set of constants representing the allowable values for ListTriggersSortOrderEnum
const (
	ListTriggersSortOrderAsc  ListTriggersSortOrderEnum = "ASC"
	ListTriggersSortOrderDesc ListTriggersSortOrderEnum = "DESC"
)

var mappingListTriggersSortOrder = map[string]ListTriggersSortOrderEnum{
	"ASC":  ListTriggersSortOrderAsc,
	"DESC": ListTriggersSortOrderDesc,
}

// GetListTriggersSortOrderEnumValues Enumerates the set of values for ListTriggersSortOrderEnum
func GetListTriggersSortOrderEnumValues() []ListTriggersSortOrderEnum {
	values := make([]ListTriggersSortOrderEnum, 0)
	for _, v := range mappingListTriggersSortOrder {
		values = append(values, v)
	}
	return values
}

// ListTriggersSortByEnum Enum with underlying type: string
type ListTriggersSortByEnum string

// Set of constants representing the allowable values for ListTriggersSortByEnum
const (
	ListTriggersSortByTimecreated ListTriggersSortByEnum = "timeCreated"
	ListTriggersSortById          ListTriggersSortByEnum = "id"
	ListTriggersSortByDisplayname ListTriggersSortByEnum = "displayName"
)

var mappingListTriggersSortBy = map[string]ListTriggersSortByEnum{
	"timeCreated": ListTriggersSortByTimecreated,
	"id":          ListTriggersSortById,
	"displayName": ListTriggersSortByDisplayname,
}

// GetListTriggersSortByEnumValues Enumerates the set of values for ListTriggersSortByEnum
func GetListTriggersSortByEnumValues() []ListTriggersSortByEnum {
	values := make([]ListTriggersSortByEnum, 0)
	for _, v := range mappingListTriggersSortBy {
		values = append(values, v)
	}
	return values
}
