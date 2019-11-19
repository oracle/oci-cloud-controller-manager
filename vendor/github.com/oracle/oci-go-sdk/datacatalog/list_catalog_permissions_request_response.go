// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package datacatalog

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListCatalogPermissionsRequest wrapper for the ListCatalogPermissions operation
type ListCatalogPermissionsRequest struct {

	// unique Catalog identifier
	CatalogId *string `mandatory:"true" contributesTo:"path" name:"catalogId"`

	// A filter to return only resources that match the entire display name given. The match is not case sensitive.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// Used to control which fields are returned in a response.
	Fields []ListCatalogPermissionsFieldsEnum `contributesTo:"query" name:"fields" omitEmpty:"true" collectionFormat:"multi"`

	// The field to sort by. Only one sort order may be provided. Default order for TIMECREATED is descending. Default order for DISPLAYNAME is ascending. If no value is specified TIMECREATED is default.
	SortBy ListCatalogPermissionsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either 'asc' or 'desc'.
	SortOrder ListCatalogPermissionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

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

func (request ListCatalogPermissionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListCatalogPermissionsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListCatalogPermissionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListCatalogPermissionsResponse wrapper for the ListCatalogPermissions operation
type ListCatalogPermissionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []CatalogPermissionsSummary instances
	Items []CatalogPermissionsSummary `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListCatalogPermissionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListCatalogPermissionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListCatalogPermissionsFieldsEnum Enum with underlying type: string
type ListCatalogPermissionsFieldsEnum string

// Set of constants representing the allowable values for ListCatalogPermissionsFieldsEnum
const (
	ListCatalogPermissionsFieldsCatalogid       ListCatalogPermissionsFieldsEnum = "catalogId"
	ListCatalogPermissionsFieldsUserpermissions ListCatalogPermissionsFieldsEnum = "userPermissions"
)

var mappingListCatalogPermissionsFields = map[string]ListCatalogPermissionsFieldsEnum{
	"catalogId":       ListCatalogPermissionsFieldsCatalogid,
	"userPermissions": ListCatalogPermissionsFieldsUserpermissions,
}

// GetListCatalogPermissionsFieldsEnumValues Enumerates the set of values for ListCatalogPermissionsFieldsEnum
func GetListCatalogPermissionsFieldsEnumValues() []ListCatalogPermissionsFieldsEnum {
	values := make([]ListCatalogPermissionsFieldsEnum, 0)
	for _, v := range mappingListCatalogPermissionsFields {
		values = append(values, v)
	}
	return values
}

// ListCatalogPermissionsSortByEnum Enum with underlying type: string
type ListCatalogPermissionsSortByEnum string

// Set of constants representing the allowable values for ListCatalogPermissionsSortByEnum
const (
	ListCatalogPermissionsSortByTimecreated ListCatalogPermissionsSortByEnum = "TIMECREATED"
	ListCatalogPermissionsSortByDisplayname ListCatalogPermissionsSortByEnum = "DISPLAYNAME"
)

var mappingListCatalogPermissionsSortBy = map[string]ListCatalogPermissionsSortByEnum{
	"TIMECREATED": ListCatalogPermissionsSortByTimecreated,
	"DISPLAYNAME": ListCatalogPermissionsSortByDisplayname,
}

// GetListCatalogPermissionsSortByEnumValues Enumerates the set of values for ListCatalogPermissionsSortByEnum
func GetListCatalogPermissionsSortByEnumValues() []ListCatalogPermissionsSortByEnum {
	values := make([]ListCatalogPermissionsSortByEnum, 0)
	for _, v := range mappingListCatalogPermissionsSortBy {
		values = append(values, v)
	}
	return values
}

// ListCatalogPermissionsSortOrderEnum Enum with underlying type: string
type ListCatalogPermissionsSortOrderEnum string

// Set of constants representing the allowable values for ListCatalogPermissionsSortOrderEnum
const (
	ListCatalogPermissionsSortOrderAsc  ListCatalogPermissionsSortOrderEnum = "ASC"
	ListCatalogPermissionsSortOrderDesc ListCatalogPermissionsSortOrderEnum = "DESC"
)

var mappingListCatalogPermissionsSortOrder = map[string]ListCatalogPermissionsSortOrderEnum{
	"ASC":  ListCatalogPermissionsSortOrderAsc,
	"DESC": ListCatalogPermissionsSortOrderDesc,
}

// GetListCatalogPermissionsSortOrderEnumValues Enumerates the set of values for ListCatalogPermissionsSortOrderEnum
func GetListCatalogPermissionsSortOrderEnumValues() []ListCatalogPermissionsSortOrderEnum {
	values := make([]ListCatalogPermissionsSortOrderEnum, 0)
	for _, v := range mappingListCatalogPermissionsSortOrder {
		values = append(values, v)
	}
	return values
}
