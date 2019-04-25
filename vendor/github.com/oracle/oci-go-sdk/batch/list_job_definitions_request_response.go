// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package batch

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListJobDefinitionsRequest wrapper for the ListJobDefinitions operation
type ListJobDefinitionsRequest struct {

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide
	// the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The field to sort by. You can provide one sort order (`sortOrder`).
	// Default order for TIMECREATED is descending. Default order for DISPLAYNAME is ascending.
	// The DISPLAYNAME sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListJobs`) let you
	// optionally filter by availability domain if the scope of the resource type
	// is within a single availability domain. If you call one of these "List" operations
	// without specifying an availability domain, the resources are grouped by availability domain,
	// then sorted.
	SortBy ListJobDefinitionsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`).
	// The DISPLAYNAME sort order is case sensitive.
	SortOrder ListJobDefinitionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The OCID of the batch instance.
	BatchInstanceId *string `mandatory:"false" contributesTo:"query" name:"batchInstanceId"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// The OCID of the Job definition.
	JobDefinitionId *string `mandatory:"false" contributesTo:"query" name:"jobDefinitionId"`

	// A filter to return only resources that match the given display name
	// exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to only return resources that match the given lifecycle state.
	// The state value is case-insensitive.
	LifecycleState ListJobDefinitionsLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

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

func (request ListJobDefinitionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListJobDefinitionsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListJobDefinitionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListJobDefinitionsResponse wrapper for the ListJobDefinitions operation
type ListJobDefinitionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []JobDefinitionSummary instances
	Items []JobDefinitionSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of
	// results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need
	// to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListJobDefinitionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListJobDefinitionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListJobDefinitionsSortByEnum Enum with underlying type: string
type ListJobDefinitionsSortByEnum string

// Set of constants representing the allowable values for ListJobDefinitionsSortByEnum
const (
	ListJobDefinitionsSortByTimecreated ListJobDefinitionsSortByEnum = "TIMECREATED"
	ListJobDefinitionsSortByDisplayname ListJobDefinitionsSortByEnum = "DISPLAYNAME"
)

var mappingListJobDefinitionsSortBy = map[string]ListJobDefinitionsSortByEnum{
	"TIMECREATED": ListJobDefinitionsSortByTimecreated,
	"DISPLAYNAME": ListJobDefinitionsSortByDisplayname,
}

// GetListJobDefinitionsSortByEnumValues Enumerates the set of values for ListJobDefinitionsSortByEnum
func GetListJobDefinitionsSortByEnumValues() []ListJobDefinitionsSortByEnum {
	values := make([]ListJobDefinitionsSortByEnum, 0)
	for _, v := range mappingListJobDefinitionsSortBy {
		values = append(values, v)
	}
	return values
}

// ListJobDefinitionsSortOrderEnum Enum with underlying type: string
type ListJobDefinitionsSortOrderEnum string

// Set of constants representing the allowable values for ListJobDefinitionsSortOrderEnum
const (
	ListJobDefinitionsSortOrderAsc  ListJobDefinitionsSortOrderEnum = "ASC"
	ListJobDefinitionsSortOrderDesc ListJobDefinitionsSortOrderEnum = "DESC"
)

var mappingListJobDefinitionsSortOrder = map[string]ListJobDefinitionsSortOrderEnum{
	"ASC":  ListJobDefinitionsSortOrderAsc,
	"DESC": ListJobDefinitionsSortOrderDesc,
}

// GetListJobDefinitionsSortOrderEnumValues Enumerates the set of values for ListJobDefinitionsSortOrderEnum
func GetListJobDefinitionsSortOrderEnumValues() []ListJobDefinitionsSortOrderEnum {
	values := make([]ListJobDefinitionsSortOrderEnum, 0)
	for _, v := range mappingListJobDefinitionsSortOrder {
		values = append(values, v)
	}
	return values
}

// ListJobDefinitionsLifecycleStateEnum Enum with underlying type: string
type ListJobDefinitionsLifecycleStateEnum string

// Set of constants representing the allowable values for ListJobDefinitionsLifecycleStateEnum
const (
	ListJobDefinitionsLifecycleStateActive  ListJobDefinitionsLifecycleStateEnum = "ACTIVE"
	ListJobDefinitionsLifecycleStateDeleted ListJobDefinitionsLifecycleStateEnum = "DELETED"
)

var mappingListJobDefinitionsLifecycleState = map[string]ListJobDefinitionsLifecycleStateEnum{
	"ACTIVE":  ListJobDefinitionsLifecycleStateActive,
	"DELETED": ListJobDefinitionsLifecycleStateDeleted,
}

// GetListJobDefinitionsLifecycleStateEnumValues Enumerates the set of values for ListJobDefinitionsLifecycleStateEnum
func GetListJobDefinitionsLifecycleStateEnumValues() []ListJobDefinitionsLifecycleStateEnum {
	values := make([]ListJobDefinitionsLifecycleStateEnum, 0)
	for _, v := range mappingListJobDefinitionsLifecycleState {
		values = append(values, v)
	}
	return values
}
