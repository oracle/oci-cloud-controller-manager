// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package batch

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListBatchInstancesRequest wrapper for the ListBatchInstances operation
type ListBatchInstancesRequest struct {

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide
	// the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// The OCID of the batch instance.
	BatchInstanceId *string `mandatory:"false" contributesTo:"query" name:"batchInstanceId"`

	// The field to sort by. You can provide one sort order (`sortOrder`).
	// Default order for TIMECREATED is descending. Default order for DISPLAYNAME is ascending.
	// The DISPLAYNAME sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListBatchInstances`) let you
	// optionally filter by availability domain if the scope of the resource type
	// is within a single availability domain. If you call one of these "List" operations
	// without specifying an availability domain, the resources are grouped by availability domain,
	// then sorted.
	SortBy ListBatchInstancesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`).
	// The DISPLAYNAME sort order is case sensitive.
	SortOrder ListBatchInstancesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only resources that match the given display name
	// exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to only return resources that match the given lifecycle state.
	// The state value is case-insensitive.
	// - ACTIVE state means the batch instance is ready for customer to use.
	// - DISABLING is in process of disable, it is a transient state on the way to INACTIVE, the batch instance is in
	// read-only mode, not allow any resource creation (compute environment, job definition, job).
	// - INACTIVE means the batch instance is in read-only mode, all job finished in the batch instance,
	// ready for delete.
	// - DELETED means cascade delete the batch instance's resources.
	LifecycleState ListBatchInstancesLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

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

func (request ListBatchInstancesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListBatchInstancesRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListBatchInstancesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListBatchInstancesResponse wrapper for the ListBatchInstances operation
type ListBatchInstancesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []BatchInstanceSummary instances
	Items []BatchInstanceSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of
	// results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need
	// to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListBatchInstancesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListBatchInstancesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListBatchInstancesSortByEnum Enum with underlying type: string
type ListBatchInstancesSortByEnum string

// Set of constants representing the allowable values for ListBatchInstancesSortByEnum
const (
	ListBatchInstancesSortByTimecreated ListBatchInstancesSortByEnum = "TIMECREATED"
	ListBatchInstancesSortByName        ListBatchInstancesSortByEnum = "NAME"
)

var mappingListBatchInstancesSortBy = map[string]ListBatchInstancesSortByEnum{
	"TIMECREATED": ListBatchInstancesSortByTimecreated,
	"NAME":        ListBatchInstancesSortByName,
}

// GetListBatchInstancesSortByEnumValues Enumerates the set of values for ListBatchInstancesSortByEnum
func GetListBatchInstancesSortByEnumValues() []ListBatchInstancesSortByEnum {
	values := make([]ListBatchInstancesSortByEnum, 0)
	for _, v := range mappingListBatchInstancesSortBy {
		values = append(values, v)
	}
	return values
}

// ListBatchInstancesSortOrderEnum Enum with underlying type: string
type ListBatchInstancesSortOrderEnum string

// Set of constants representing the allowable values for ListBatchInstancesSortOrderEnum
const (
	ListBatchInstancesSortOrderAsc  ListBatchInstancesSortOrderEnum = "ASC"
	ListBatchInstancesSortOrderDesc ListBatchInstancesSortOrderEnum = "DESC"
)

var mappingListBatchInstancesSortOrder = map[string]ListBatchInstancesSortOrderEnum{
	"ASC":  ListBatchInstancesSortOrderAsc,
	"DESC": ListBatchInstancesSortOrderDesc,
}

// GetListBatchInstancesSortOrderEnumValues Enumerates the set of values for ListBatchInstancesSortOrderEnum
func GetListBatchInstancesSortOrderEnumValues() []ListBatchInstancesSortOrderEnum {
	values := make([]ListBatchInstancesSortOrderEnum, 0)
	for _, v := range mappingListBatchInstancesSortOrder {
		values = append(values, v)
	}
	return values
}

// ListBatchInstancesLifecycleStateEnum Enum with underlying type: string
type ListBatchInstancesLifecycleStateEnum string

// Set of constants representing the allowable values for ListBatchInstancesLifecycleStateEnum
const (
	ListBatchInstancesLifecycleStateActive    ListBatchInstancesLifecycleStateEnum = "ACTIVE"
	ListBatchInstancesLifecycleStateDisabling ListBatchInstancesLifecycleStateEnum = "DISABLING"
	ListBatchInstancesLifecycleStateInactive  ListBatchInstancesLifecycleStateEnum = "INACTIVE"
	ListBatchInstancesLifecycleStateDeleted   ListBatchInstancesLifecycleStateEnum = "DELETED"
)

var mappingListBatchInstancesLifecycleState = map[string]ListBatchInstancesLifecycleStateEnum{
	"ACTIVE":    ListBatchInstancesLifecycleStateActive,
	"DISABLING": ListBatchInstancesLifecycleStateDisabling,
	"INACTIVE":  ListBatchInstancesLifecycleStateInactive,
	"DELETED":   ListBatchInstancesLifecycleStateDeleted,
}

// GetListBatchInstancesLifecycleStateEnumValues Enumerates the set of values for ListBatchInstancesLifecycleStateEnum
func GetListBatchInstancesLifecycleStateEnumValues() []ListBatchInstancesLifecycleStateEnum {
	values := make([]ListBatchInstancesLifecycleStateEnum, 0)
	for _, v := range mappingListBatchInstancesLifecycleState {
		values = append(values, v)
	}
	return values
}
