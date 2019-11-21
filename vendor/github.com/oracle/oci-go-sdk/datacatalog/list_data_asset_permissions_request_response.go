// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package datacatalog

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListDataAssetPermissionsRequest wrapper for the ListDataAssetPermissions operation
type ListDataAssetPermissionsRequest struct {

	// unique Catalog identifier
	CatalogId *string `mandatory:"true" contributesTo:"path" name:"catalogId"`

	// Unique Data Asset key.
	DataAssetKey *string `mandatory:"true" contributesTo:"path" name:"dataAssetKey"`

	// Immutable resource name.
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

	// Used to control which fields are returned in a Data Asset Permission response.
	Fields []ListDataAssetPermissionsFieldsEnum `contributesTo:"query" name:"fields" omitEmpty:"true" collectionFormat:"multi"`

	// The field to sort by. Only one sort order may be provided. Default order for TIMECREATED is descending. Default order for DISPLAYNAME is ascending. If no value is specified TIMECREATED is default.
	SortBy ListDataAssetPermissionsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either 'asc' or 'desc'.
	SortOrder ListDataAssetPermissionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The maximum number of items to return.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The page token representing the page at which to start retrieving results. This is usually retrieved from a previous list call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The client request ID for tracing.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListDataAssetPermissionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListDataAssetPermissionsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListDataAssetPermissionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListDataAssetPermissionsResponse wrapper for the ListDataAssetPermissions operation
type ListDataAssetPermissionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []DataAssetPermissionsSummary instances
	Items []DataAssetPermissionsSummary `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListDataAssetPermissionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListDataAssetPermissionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListDataAssetPermissionsFieldsEnum Enum with underlying type: string
type ListDataAssetPermissionsFieldsEnum string

// Set of constants representing the allowable values for ListDataAssetPermissionsFieldsEnum
const (
	ListDataAssetPermissionsFieldsDataassetkey    ListDataAssetPermissionsFieldsEnum = "dataAssetKey"
	ListDataAssetPermissionsFieldsUserpermissions ListDataAssetPermissionsFieldsEnum = "userPermissions"
)

var mappingListDataAssetPermissionsFields = map[string]ListDataAssetPermissionsFieldsEnum{
	"dataAssetKey":    ListDataAssetPermissionsFieldsDataassetkey,
	"userPermissions": ListDataAssetPermissionsFieldsUserpermissions,
}

// GetListDataAssetPermissionsFieldsEnumValues Enumerates the set of values for ListDataAssetPermissionsFieldsEnum
func GetListDataAssetPermissionsFieldsEnumValues() []ListDataAssetPermissionsFieldsEnum {
	values := make([]ListDataAssetPermissionsFieldsEnum, 0)
	for _, v := range mappingListDataAssetPermissionsFields {
		values = append(values, v)
	}
	return values
}

// ListDataAssetPermissionsSortByEnum Enum with underlying type: string
type ListDataAssetPermissionsSortByEnum string

// Set of constants representing the allowable values for ListDataAssetPermissionsSortByEnum
const (
	ListDataAssetPermissionsSortByTimecreated ListDataAssetPermissionsSortByEnum = "TIMECREATED"
	ListDataAssetPermissionsSortByDisplayname ListDataAssetPermissionsSortByEnum = "DISPLAYNAME"
)

var mappingListDataAssetPermissionsSortBy = map[string]ListDataAssetPermissionsSortByEnum{
	"TIMECREATED": ListDataAssetPermissionsSortByTimecreated,
	"DISPLAYNAME": ListDataAssetPermissionsSortByDisplayname,
}

// GetListDataAssetPermissionsSortByEnumValues Enumerates the set of values for ListDataAssetPermissionsSortByEnum
func GetListDataAssetPermissionsSortByEnumValues() []ListDataAssetPermissionsSortByEnum {
	values := make([]ListDataAssetPermissionsSortByEnum, 0)
	for _, v := range mappingListDataAssetPermissionsSortBy {
		values = append(values, v)
	}
	return values
}

// ListDataAssetPermissionsSortOrderEnum Enum with underlying type: string
type ListDataAssetPermissionsSortOrderEnum string

// Set of constants representing the allowable values for ListDataAssetPermissionsSortOrderEnum
const (
	ListDataAssetPermissionsSortOrderAsc  ListDataAssetPermissionsSortOrderEnum = "ASC"
	ListDataAssetPermissionsSortOrderDesc ListDataAssetPermissionsSortOrderEnum = "DESC"
)

var mappingListDataAssetPermissionsSortOrder = map[string]ListDataAssetPermissionsSortOrderEnum{
	"ASC":  ListDataAssetPermissionsSortOrderAsc,
	"DESC": ListDataAssetPermissionsSortOrderDesc,
}

// GetListDataAssetPermissionsSortOrderEnumValues Enumerates the set of values for ListDataAssetPermissionsSortOrderEnum
func GetListDataAssetPermissionsSortOrderEnumValues() []ListDataAssetPermissionsSortOrderEnum {
	values := make([]ListDataAssetPermissionsSortOrderEnum, 0)
	for _, v := range mappingListDataAssetPermissionsSortOrder {
		values = append(values, v)
	}
	return values
}
