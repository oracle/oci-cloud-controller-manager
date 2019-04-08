// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package batch

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListJobsRequest wrapper for the ListJobs operation
type ListJobsRequest struct {

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
	SortBy ListJobsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`).
	// The DISPLAYNAME sort order is case sensitive.
	SortOrder ListJobsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The OCID of the batch instance.
	BatchInstanceId *string `mandatory:"false" contributesTo:"query" name:"batchInstanceId"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// The OCID of the Job.
	JobId *string `mandatory:"false" contributesTo:"query" name:"jobId"`

	// A filter to return only resources that match the given display name
	// exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The current work request status of the job.
	LifecycleState ListJobsLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

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

func (request ListJobsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListJobsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListJobsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListJobsResponse wrapper for the ListJobs operation
type ListJobsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []JobSummary instances
	Items []JobSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages of
	// results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need
	// to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListJobsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListJobsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListJobsSortByEnum Enum with underlying type: string
type ListJobsSortByEnum string

// Set of constants representing the allowable values for ListJobsSortByEnum
const (
	ListJobsSortByTimecreated ListJobsSortByEnum = "TIMECREATED"
	ListJobsSortByDisplayname ListJobsSortByEnum = "DISPLAYNAME"
)

var mappingListJobsSortBy = map[string]ListJobsSortByEnum{
	"TIMECREATED": ListJobsSortByTimecreated,
	"DISPLAYNAME": ListJobsSortByDisplayname,
}

// GetListJobsSortByEnumValues Enumerates the set of values for ListJobsSortByEnum
func GetListJobsSortByEnumValues() []ListJobsSortByEnum {
	values := make([]ListJobsSortByEnum, 0)
	for _, v := range mappingListJobsSortBy {
		values = append(values, v)
	}
	return values
}

// ListJobsSortOrderEnum Enum with underlying type: string
type ListJobsSortOrderEnum string

// Set of constants representing the allowable values for ListJobsSortOrderEnum
const (
	ListJobsSortOrderAsc  ListJobsSortOrderEnum = "ASC"
	ListJobsSortOrderDesc ListJobsSortOrderEnum = "DESC"
)

var mappingListJobsSortOrder = map[string]ListJobsSortOrderEnum{
	"ASC":  ListJobsSortOrderAsc,
	"DESC": ListJobsSortOrderDesc,
}

// GetListJobsSortOrderEnumValues Enumerates the set of values for ListJobsSortOrderEnum
func GetListJobsSortOrderEnumValues() []ListJobsSortOrderEnum {
	values := make([]ListJobsSortOrderEnum, 0)
	for _, v := range mappingListJobsSortOrder {
		values = append(values, v)
	}
	return values
}

// ListJobsLifecycleStateEnum Enum with underlying type: string
type ListJobsLifecycleStateEnum string

// Set of constants representing the allowable values for ListJobsLifecycleStateEnum
const (
	ListJobsLifecycleStateAccepted   ListJobsLifecycleStateEnum = "ACCEPTED"
	ListJobsLifecycleStateInProgress ListJobsLifecycleStateEnum = "IN_PROGRESS"
	ListJobsLifecycleStateFailed     ListJobsLifecycleStateEnum = "FAILED"
	ListJobsLifecycleStateSucceeded  ListJobsLifecycleStateEnum = "SUCCEEDED"
	ListJobsLifecycleStateCanceling  ListJobsLifecycleStateEnum = "CANCELING"
	ListJobsLifecycleStateCanceled   ListJobsLifecycleStateEnum = "CANCELED"
	ListJobsLifecycleStateDeleted    ListJobsLifecycleStateEnum = "DELETED"
)

var mappingListJobsLifecycleState = map[string]ListJobsLifecycleStateEnum{
	"ACCEPTED":    ListJobsLifecycleStateAccepted,
	"IN_PROGRESS": ListJobsLifecycleStateInProgress,
	"FAILED":      ListJobsLifecycleStateFailed,
	"SUCCEEDED":   ListJobsLifecycleStateSucceeded,
	"CANCELING":   ListJobsLifecycleStateCanceling,
	"CANCELED":    ListJobsLifecycleStateCanceled,
	"DELETED":     ListJobsLifecycleStateDeleted,
}

// GetListJobsLifecycleStateEnumValues Enumerates the set of values for ListJobsLifecycleStateEnum
func GetListJobsLifecycleStateEnumValues() []ListJobsLifecycleStateEnum {
	values := make([]ListJobsLifecycleStateEnum, 0)
	for _, v := range mappingListJobsLifecycleState {
		values = append(values, v)
	}
	return values
}
