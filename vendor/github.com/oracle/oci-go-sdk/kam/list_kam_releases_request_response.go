// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package kam

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListKamReleasesRequest wrapper for the ListKamReleases operation
type ListKamReleasesRequest struct {

	// The OCID of the cluster.
	ClusterId *string `mandatory:"true" contributesTo:"query" name:"clusterId"`

	// The releaseId to filter on.
	ReleaseId *string `mandatory:"false" contributesTo:"query" name:"releaseId"`

	// The name to filter on
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

	// The package type to restrict results to, ADD_ON or APPLICATION or *
	PackageType *string `mandatory:"false" contributesTo:"query" name:"packageType"`

	// The package name to filter on
	PackageName *string `mandatory:"false" contributesTo:"query" name:"packageName"`

	// The maximum number of items to return in a paginated "List" call.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The optional order in which to sort the results.
	SortOrder ListKamReleasesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the results by.
	SortBy ListKamReleasesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// To limit the list to releases with given lifecycle state
	LifecycleState ListKamReleasesLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListKamReleasesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListKamReleasesRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListKamReleasesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListKamReleasesResponse wrapper for the ListKamReleases operation
type ListKamReleasesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []KamReleaseSummary instances
	Items []KamReleaseSummary `presentIn:"body"`

	// For pagination of a list of items. When paging through a list, if this header appears in the response,
	// then a partial list might have been returned. Include this value as the `page` parameter for the
	// subsequent GET request to get the next batch of items.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListKamReleasesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListKamReleasesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListKamReleasesSortOrderEnum Enum with underlying type: string
type ListKamReleasesSortOrderEnum string

// Set of constants representing the allowable values for ListKamReleasesSortOrderEnum
const (
	ListKamReleasesSortOrderAsc  ListKamReleasesSortOrderEnum = "ASC"
	ListKamReleasesSortOrderDesc ListKamReleasesSortOrderEnum = "DESC"
)

var mappingListKamReleasesSortOrder = map[string]ListKamReleasesSortOrderEnum{
	"ASC":  ListKamReleasesSortOrderAsc,
	"DESC": ListKamReleasesSortOrderDesc,
}

// GetListKamReleasesSortOrderEnumValues Enumerates the set of values for ListKamReleasesSortOrderEnum
func GetListKamReleasesSortOrderEnumValues() []ListKamReleasesSortOrderEnum {
	values := make([]ListKamReleasesSortOrderEnum, 0)
	for _, v := range mappingListKamReleasesSortOrder {
		values = append(values, v)
	}
	return values
}

// ListKamReleasesSortByEnum Enum with underlying type: string
type ListKamReleasesSortByEnum string

// Set of constants representing the allowable values for ListKamReleasesSortByEnum
const (
	ListKamReleasesSortByTimeaccepted ListKamReleasesSortByEnum = "TIMEACCEPTED"
	ListKamReleasesSortByTimeupdated  ListKamReleasesSortByEnum = "TIMEUPDATED"
)

var mappingListKamReleasesSortBy = map[string]ListKamReleasesSortByEnum{
	"TIMEACCEPTED": ListKamReleasesSortByTimeaccepted,
	"TIMEUPDATED":  ListKamReleasesSortByTimeupdated,
}

// GetListKamReleasesSortByEnumValues Enumerates the set of values for ListKamReleasesSortByEnum
func GetListKamReleasesSortByEnumValues() []ListKamReleasesSortByEnum {
	values := make([]ListKamReleasesSortByEnum, 0)
	for _, v := range mappingListKamReleasesSortBy {
		values = append(values, v)
	}
	return values
}

// ListKamReleasesLifecycleStateEnum Enum with underlying type: string
type ListKamReleasesLifecycleStateEnum string

// Set of constants representing the allowable values for ListKamReleasesLifecycleStateEnum
const (
	ListKamReleasesLifecycleStateCreating ListKamReleasesLifecycleStateEnum = "CREATING"
	ListKamReleasesLifecycleStateUpdating ListKamReleasesLifecycleStateEnum = "UPDATING"
	ListKamReleasesLifecycleStateActive   ListKamReleasesLifecycleStateEnum = "ACTIVE"
	ListKamReleasesLifecycleStateDeleting ListKamReleasesLifecycleStateEnum = "DELETING"
	ListKamReleasesLifecycleStateDeleted  ListKamReleasesLifecycleStateEnum = "DELETED"
	ListKamReleasesLifecycleStateFailed   ListKamReleasesLifecycleStateEnum = "FAILED"
)

var mappingListKamReleasesLifecycleState = map[string]ListKamReleasesLifecycleStateEnum{
	"CREATING": ListKamReleasesLifecycleStateCreating,
	"UPDATING": ListKamReleasesLifecycleStateUpdating,
	"ACTIVE":   ListKamReleasesLifecycleStateActive,
	"DELETING": ListKamReleasesLifecycleStateDeleting,
	"DELETED":  ListKamReleasesLifecycleStateDeleted,
	"FAILED":   ListKamReleasesLifecycleStateFailed,
}

// GetListKamReleasesLifecycleStateEnumValues Enumerates the set of values for ListKamReleasesLifecycleStateEnum
func GetListKamReleasesLifecycleStateEnumValues() []ListKamReleasesLifecycleStateEnum {
	values := make([]ListKamReleasesLifecycleStateEnum, 0)
	for _, v := range mappingListKamReleasesLifecycleState {
		values = append(values, v)
	}
	return values
}
