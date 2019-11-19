// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package datacatalog

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListGlossaryPermissionsRequest wrapper for the ListGlossaryPermissions operation
type ListGlossaryPermissionsRequest struct {

	// unique Catalog identifier
	CatalogId *string `mandatory:"true" contributesTo:"path" name:"catalogId"`

	// Unique Glossary key.
	GlossaryKey *string `mandatory:"true" contributesTo:"path" name:"glossaryKey"`

	// Immutable resource name.
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

	// Used to control which fields are returned in a Glossary Permission response.
	Fields []ListGlossaryPermissionsFieldsEnum `contributesTo:"query" name:"fields" omitEmpty:"true" collectionFormat:"multi"`

	// The field to sort by. Only one sort order may be provided. Default order for TIMECREATED is descending. Default order for DISPLAYNAME is ascending. If no value is specified TIMECREATED is default.
	SortBy ListGlossaryPermissionsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either 'asc' or 'desc'.
	SortOrder ListGlossaryPermissionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

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

func (request ListGlossaryPermissionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListGlossaryPermissionsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListGlossaryPermissionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListGlossaryPermissionsResponse wrapper for the ListGlossaryPermissions operation
type ListGlossaryPermissionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []GlossaryPermissionsSummary instances
	Items []GlossaryPermissionsSummary `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListGlossaryPermissionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListGlossaryPermissionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListGlossaryPermissionsFieldsEnum Enum with underlying type: string
type ListGlossaryPermissionsFieldsEnum string

// Set of constants representing the allowable values for ListGlossaryPermissionsFieldsEnum
const (
	ListGlossaryPermissionsFieldsGlossarykey     ListGlossaryPermissionsFieldsEnum = "glossaryKey"
	ListGlossaryPermissionsFieldsUserpermissions ListGlossaryPermissionsFieldsEnum = "userPermissions"
)

var mappingListGlossaryPermissionsFields = map[string]ListGlossaryPermissionsFieldsEnum{
	"glossaryKey":     ListGlossaryPermissionsFieldsGlossarykey,
	"userPermissions": ListGlossaryPermissionsFieldsUserpermissions,
}

// GetListGlossaryPermissionsFieldsEnumValues Enumerates the set of values for ListGlossaryPermissionsFieldsEnum
func GetListGlossaryPermissionsFieldsEnumValues() []ListGlossaryPermissionsFieldsEnum {
	values := make([]ListGlossaryPermissionsFieldsEnum, 0)
	for _, v := range mappingListGlossaryPermissionsFields {
		values = append(values, v)
	}
	return values
}

// ListGlossaryPermissionsSortByEnum Enum with underlying type: string
type ListGlossaryPermissionsSortByEnum string

// Set of constants representing the allowable values for ListGlossaryPermissionsSortByEnum
const (
	ListGlossaryPermissionsSortByTimecreated ListGlossaryPermissionsSortByEnum = "TIMECREATED"
	ListGlossaryPermissionsSortByDisplayname ListGlossaryPermissionsSortByEnum = "DISPLAYNAME"
)

var mappingListGlossaryPermissionsSortBy = map[string]ListGlossaryPermissionsSortByEnum{
	"TIMECREATED": ListGlossaryPermissionsSortByTimecreated,
	"DISPLAYNAME": ListGlossaryPermissionsSortByDisplayname,
}

// GetListGlossaryPermissionsSortByEnumValues Enumerates the set of values for ListGlossaryPermissionsSortByEnum
func GetListGlossaryPermissionsSortByEnumValues() []ListGlossaryPermissionsSortByEnum {
	values := make([]ListGlossaryPermissionsSortByEnum, 0)
	for _, v := range mappingListGlossaryPermissionsSortBy {
		values = append(values, v)
	}
	return values
}

// ListGlossaryPermissionsSortOrderEnum Enum with underlying type: string
type ListGlossaryPermissionsSortOrderEnum string

// Set of constants representing the allowable values for ListGlossaryPermissionsSortOrderEnum
const (
	ListGlossaryPermissionsSortOrderAsc  ListGlossaryPermissionsSortOrderEnum = "ASC"
	ListGlossaryPermissionsSortOrderDesc ListGlossaryPermissionsSortOrderEnum = "DESC"
)

var mappingListGlossaryPermissionsSortOrder = map[string]ListGlossaryPermissionsSortOrderEnum{
	"ASC":  ListGlossaryPermissionsSortOrderAsc,
	"DESC": ListGlossaryPermissionsSortOrderDesc,
}

// GetListGlossaryPermissionsSortOrderEnumValues Enumerates the set of values for ListGlossaryPermissionsSortOrderEnum
func GetListGlossaryPermissionsSortOrderEnumValues() []ListGlossaryPermissionsSortOrderEnum {
	values := make([]ListGlossaryPermissionsSortOrderEnum, 0)
	for _, v := range mappingListGlossaryPermissionsSortOrder {
		values = append(values, v)
	}
	return values
}
