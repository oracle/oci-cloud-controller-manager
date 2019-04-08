// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package cache

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListReplicatedCachesRequest wrapper for the ListReplicatedCaches operation
type ListReplicatedCachesRequest struct {

	// The OCID of the compartment for which to list the Redis replicated
	// caches.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID of the Redis replicated cache.
	Id *string `mandatory:"false" contributesTo:"query" name:"id"`

	// The name of the Redis replicated cache to be included in the list.
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

	// The current lifecycle state of the Redis replicated cache.
	LifecycleState ListReplicatedCachesLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// The lifecycle states that are not current for the Redis replicated cache.
	LifecycleStateNotEquals ListReplicatedCachesLifecycleStateNotEqualsEnum `mandatory:"false" contributesTo:"query" name:"lifecycleStateNotEquals" omitEmpty:"true"`

	// The field on which to sort the results.
	SortBy ListReplicatedCachesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The order of sorting (ASC or DESC).
	SortOrder ListReplicatedCachesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The value of the opc-next-page response header from the previous request.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The maximum number of items to return.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// A unique Oracle-assigned identifier for the request. If you need
	// to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListReplicatedCachesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListReplicatedCachesRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListReplicatedCachesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListReplicatedCachesResponse wrapper for the ListReplicatedCaches operation
type ListReplicatedCachesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ReplicatedCacheSummary instances
	Items []ReplicatedCacheSummary `presentIn:"body"`

	// A unique Oracle-assigned identifier for the request. If
	// you need to contact Oracle about a particular request, please provide
	// the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// The next page value to provide as a page header in next request
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListReplicatedCachesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListReplicatedCachesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListReplicatedCachesLifecycleStateEnum Enum with underlying type: string
type ListReplicatedCachesLifecycleStateEnum string

// Set of constants representing the allowable values for ListReplicatedCachesLifecycleStateEnum
const (
	ListReplicatedCachesLifecycleStateCreating ListReplicatedCachesLifecycleStateEnum = "CREATING"
	ListReplicatedCachesLifecycleStateActive   ListReplicatedCachesLifecycleStateEnum = "ACTIVE"
	ListReplicatedCachesLifecycleStateUpdating ListReplicatedCachesLifecycleStateEnum = "UPDATING"
	ListReplicatedCachesLifecycleStateDeleting ListReplicatedCachesLifecycleStateEnum = "DELETING"
	ListReplicatedCachesLifecycleStateDeleted  ListReplicatedCachesLifecycleStateEnum = "DELETED"
	ListReplicatedCachesLifecycleStateFailed   ListReplicatedCachesLifecycleStateEnum = "FAILED"
)

var mappingListReplicatedCachesLifecycleState = map[string]ListReplicatedCachesLifecycleStateEnum{
	"CREATING": ListReplicatedCachesLifecycleStateCreating,
	"ACTIVE":   ListReplicatedCachesLifecycleStateActive,
	"UPDATING": ListReplicatedCachesLifecycleStateUpdating,
	"DELETING": ListReplicatedCachesLifecycleStateDeleting,
	"DELETED":  ListReplicatedCachesLifecycleStateDeleted,
	"FAILED":   ListReplicatedCachesLifecycleStateFailed,
}

// GetListReplicatedCachesLifecycleStateEnumValues Enumerates the set of values for ListReplicatedCachesLifecycleStateEnum
func GetListReplicatedCachesLifecycleStateEnumValues() []ListReplicatedCachesLifecycleStateEnum {
	values := make([]ListReplicatedCachesLifecycleStateEnum, 0)
	for _, v := range mappingListReplicatedCachesLifecycleState {
		values = append(values, v)
	}
	return values
}

// ListReplicatedCachesLifecycleStateNotEqualsEnum Enum with underlying type: string
type ListReplicatedCachesLifecycleStateNotEqualsEnum string

// Set of constants representing the allowable values for ListReplicatedCachesLifecycleStateNotEqualsEnum
const (
	ListReplicatedCachesLifecycleStateNotEqualsCreating ListReplicatedCachesLifecycleStateNotEqualsEnum = "CREATING"
	ListReplicatedCachesLifecycleStateNotEqualsActive   ListReplicatedCachesLifecycleStateNotEqualsEnum = "ACTIVE"
	ListReplicatedCachesLifecycleStateNotEqualsUpdating ListReplicatedCachesLifecycleStateNotEqualsEnum = "UPDATING"
	ListReplicatedCachesLifecycleStateNotEqualsDeleting ListReplicatedCachesLifecycleStateNotEqualsEnum = "DELETING"
	ListReplicatedCachesLifecycleStateNotEqualsDeleted  ListReplicatedCachesLifecycleStateNotEqualsEnum = "DELETED"
	ListReplicatedCachesLifecycleStateNotEqualsFailed   ListReplicatedCachesLifecycleStateNotEqualsEnum = "FAILED"
)

var mappingListReplicatedCachesLifecycleStateNotEquals = map[string]ListReplicatedCachesLifecycleStateNotEqualsEnum{
	"CREATING": ListReplicatedCachesLifecycleStateNotEqualsCreating,
	"ACTIVE":   ListReplicatedCachesLifecycleStateNotEqualsActive,
	"UPDATING": ListReplicatedCachesLifecycleStateNotEqualsUpdating,
	"DELETING": ListReplicatedCachesLifecycleStateNotEqualsDeleting,
	"DELETED":  ListReplicatedCachesLifecycleStateNotEqualsDeleted,
	"FAILED":   ListReplicatedCachesLifecycleStateNotEqualsFailed,
}

// GetListReplicatedCachesLifecycleStateNotEqualsEnumValues Enumerates the set of values for ListReplicatedCachesLifecycleStateNotEqualsEnum
func GetListReplicatedCachesLifecycleStateNotEqualsEnumValues() []ListReplicatedCachesLifecycleStateNotEqualsEnum {
	values := make([]ListReplicatedCachesLifecycleStateNotEqualsEnum, 0)
	for _, v := range mappingListReplicatedCachesLifecycleStateNotEquals {
		values = append(values, v)
	}
	return values
}

// ListReplicatedCachesSortByEnum Enum with underlying type: string
type ListReplicatedCachesSortByEnum string

// Set of constants representing the allowable values for ListReplicatedCachesSortByEnum
const (
	ListReplicatedCachesSortByName           ListReplicatedCachesSortByEnum = "NAME"
	ListReplicatedCachesSortByTimeCreated    ListReplicatedCachesSortByEnum = "TIME_CREATED"
	ListReplicatedCachesSortByLifecycleState ListReplicatedCachesSortByEnum = "LIFECYCLE_STATE"
)

var mappingListReplicatedCachesSortBy = map[string]ListReplicatedCachesSortByEnum{
	"NAME":            ListReplicatedCachesSortByName,
	"TIME_CREATED":    ListReplicatedCachesSortByTimeCreated,
	"LIFECYCLE_STATE": ListReplicatedCachesSortByLifecycleState,
}

// GetListReplicatedCachesSortByEnumValues Enumerates the set of values for ListReplicatedCachesSortByEnum
func GetListReplicatedCachesSortByEnumValues() []ListReplicatedCachesSortByEnum {
	values := make([]ListReplicatedCachesSortByEnum, 0)
	for _, v := range mappingListReplicatedCachesSortBy {
		values = append(values, v)
	}
	return values
}

// ListReplicatedCachesSortOrderEnum Enum with underlying type: string
type ListReplicatedCachesSortOrderEnum string

// Set of constants representing the allowable values for ListReplicatedCachesSortOrderEnum
const (
	ListReplicatedCachesSortOrderAsc  ListReplicatedCachesSortOrderEnum = "ASC"
	ListReplicatedCachesSortOrderDesc ListReplicatedCachesSortOrderEnum = "DESC"
)

var mappingListReplicatedCachesSortOrder = map[string]ListReplicatedCachesSortOrderEnum{
	"ASC":  ListReplicatedCachesSortOrderAsc,
	"DESC": ListReplicatedCachesSortOrderDesc,
}

// GetListReplicatedCachesSortOrderEnumValues Enumerates the set of values for ListReplicatedCachesSortOrderEnum
func GetListReplicatedCachesSortOrderEnumValues() []ListReplicatedCachesSortOrderEnum {
	values := make([]ListReplicatedCachesSortOrderEnum, 0)
	for _, v := range mappingListReplicatedCachesSortOrder {
		values = append(values, v)
	}
	return values
}
