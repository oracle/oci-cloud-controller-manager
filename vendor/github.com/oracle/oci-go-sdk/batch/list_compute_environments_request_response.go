// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package batch

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListComputeEnvironmentsRequest wrapper for the ListComputeEnvironments operation
type ListComputeEnvironmentsRequest struct {

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide
	// the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The OCID of the batch instance.
	BatchInstanceId *string `mandatory:"false" contributesTo:"query" name:"batchInstanceId"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// The OCID of the compute environment.
	ComputeEnvironmentId *string `mandatory:"false" contributesTo:"query" name:"computeEnvironmentId"`

	// The field to sort by. You can provide one sort order (`sortOrder`).
	// Default order for TIMECREATED is descending. Default order for DISPLAYNAME is ascending.
	// The DISPLAYNAME sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListJobs`) let you
	// optionally filter by availability domain if the scope of the resource type
	// is within a single availability domain. If you call one of these "List" operations
	// without specifying an availability domain, the resources are grouped by availability domain,
	// then sorted.
	SortBy ListComputeEnvironmentsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`).
	// The DISPLAYNAME sort order is case sensitive.
	SortOrder ListComputeEnvironmentsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only resources that match the given display name
	// exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to only return resources that match the given lifecycle state.
	// The state value is case-insensitive.
	// - ACTIVE state means the compute environment is ready to use, user can select
	// the compute environment when submitting job.
	// - DISABLING is in process of disable, it is a transient state on the way to INACTIVE, waiting for all running
	// job complete in the compute environment, user can not select the compute environment when submitting job.
	// - INACTIVE means user can not select the compute environment when submitting job,
	// all job finished in the compute environment, ready for delete.
	// - DELETED means cascade delete the compute environment's resource, including node pool, worker node.
	LifecycleState ListComputeEnvironmentsLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

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

func (request ListComputeEnvironmentsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeEnvironmentsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeEnvironmentsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListComputeEnvironmentsResponse wrapper for the ListComputeEnvironments operation
type ListComputeEnvironmentsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ComputeEnvironmentSummary instances
	Items []ComputeEnvironmentSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of
	// results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need
	// to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeEnvironmentsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeEnvironmentsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeEnvironmentsSortByEnum Enum with underlying type: string
type ListComputeEnvironmentsSortByEnum string

// Set of constants representing the allowable values for ListComputeEnvironmentsSortByEnum
const (
	ListComputeEnvironmentsSortByTimecreated ListComputeEnvironmentsSortByEnum = "TIMECREATED"
	ListComputeEnvironmentsSortByDisplayname ListComputeEnvironmentsSortByEnum = "DISPLAYNAME"
)

var mappingListComputeEnvironmentsSortBy = map[string]ListComputeEnvironmentsSortByEnum{
	"TIMECREATED": ListComputeEnvironmentsSortByTimecreated,
	"DISPLAYNAME": ListComputeEnvironmentsSortByDisplayname,
}

// GetListComputeEnvironmentsSortByEnumValues Enumerates the set of values for ListComputeEnvironmentsSortByEnum
func GetListComputeEnvironmentsSortByEnumValues() []ListComputeEnvironmentsSortByEnum {
	values := make([]ListComputeEnvironmentsSortByEnum, 0)
	for _, v := range mappingListComputeEnvironmentsSortBy {
		values = append(values, v)
	}
	return values
}

// ListComputeEnvironmentsSortOrderEnum Enum with underlying type: string
type ListComputeEnvironmentsSortOrderEnum string

// Set of constants representing the allowable values for ListComputeEnvironmentsSortOrderEnum
const (
	ListComputeEnvironmentsSortOrderAsc  ListComputeEnvironmentsSortOrderEnum = "ASC"
	ListComputeEnvironmentsSortOrderDesc ListComputeEnvironmentsSortOrderEnum = "DESC"
)

var mappingListComputeEnvironmentsSortOrder = map[string]ListComputeEnvironmentsSortOrderEnum{
	"ASC":  ListComputeEnvironmentsSortOrderAsc,
	"DESC": ListComputeEnvironmentsSortOrderDesc,
}

// GetListComputeEnvironmentsSortOrderEnumValues Enumerates the set of values for ListComputeEnvironmentsSortOrderEnum
func GetListComputeEnvironmentsSortOrderEnumValues() []ListComputeEnvironmentsSortOrderEnum {
	values := make([]ListComputeEnvironmentsSortOrderEnum, 0)
	for _, v := range mappingListComputeEnvironmentsSortOrder {
		values = append(values, v)
	}
	return values
}

// ListComputeEnvironmentsLifecycleStateEnum Enum with underlying type: string
type ListComputeEnvironmentsLifecycleStateEnum string

// Set of constants representing the allowable values for ListComputeEnvironmentsLifecycleStateEnum
const (
	ListComputeEnvironmentsLifecycleStateActive    ListComputeEnvironmentsLifecycleStateEnum = "ACTIVE"
	ListComputeEnvironmentsLifecycleStateDisabling ListComputeEnvironmentsLifecycleStateEnum = "DISABLING"
	ListComputeEnvironmentsLifecycleStateInactive  ListComputeEnvironmentsLifecycleStateEnum = "INACTIVE"
	ListComputeEnvironmentsLifecycleStateDeleted   ListComputeEnvironmentsLifecycleStateEnum = "DELETED"
)

var mappingListComputeEnvironmentsLifecycleState = map[string]ListComputeEnvironmentsLifecycleStateEnum{
	"ACTIVE":    ListComputeEnvironmentsLifecycleStateActive,
	"DISABLING": ListComputeEnvironmentsLifecycleStateDisabling,
	"INACTIVE":  ListComputeEnvironmentsLifecycleStateInactive,
	"DELETED":   ListComputeEnvironmentsLifecycleStateDeleted,
}

// GetListComputeEnvironmentsLifecycleStateEnumValues Enumerates the set of values for ListComputeEnvironmentsLifecycleStateEnum
func GetListComputeEnvironmentsLifecycleStateEnumValues() []ListComputeEnvironmentsLifecycleStateEnum {
	values := make([]ListComputeEnvironmentsLifecycleStateEnum, 0)
	for _, v := range mappingListComputeEnvironmentsLifecycleState {
		values = append(values, v)
	}
	return values
}
