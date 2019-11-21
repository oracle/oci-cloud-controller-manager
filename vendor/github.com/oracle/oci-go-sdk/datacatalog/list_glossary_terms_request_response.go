// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package datacatalog

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListGlossaryTermsRequest wrapper for the ListGlossaryTerms operation
type ListGlossaryTermsRequest struct {

	// unique Catalog identifier
	CatalogId *string `mandatory:"true" contributesTo:"path" name:"catalogId"`

	// Unique Glossary key.
	GlossaryKey *string `mandatory:"true" contributesTo:"path" name:"glossaryKey"`

	// A filter to return only resources that match the entire display name given. The match is not case sensitive.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to return only resources that match the specified lifecycle state. The value is case insensitive.
	LifecycleState LifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique key of the parent term.
	ParentTermKey *string `mandatory:"false" contributesTo:"query" name:"parentTermKey"`

	// Indicates whether a term may contain child terms.
	IsAllowedToHaveChildTerms *bool `mandatory:"false" contributesTo:"query" name:"isAllowedToHaveChildTerms"`

	// Status of the approval workflow for this business term in the glossary
	WorkflowStatus TermWorkflowStatusEnum `mandatory:"false" contributesTo:"query" name:"workflowStatus" omitEmpty:"true"`

	// Full path of the resource for resources that support paths.
	Path *string `mandatory:"false" contributesTo:"query" name:"path"`

	// Used to control which fields are returned in a Term summary response.
	Fields []ListGlossaryTermsFieldsEnum `contributesTo:"query" name:"fields" omitEmpty:"true" collectionFormat:"multi"`

	// The field to sort by. Only one sort order may be provided. Default order for TIMECREATED is descending. Default order for DISPLAYNAME is ascending. If no value is specified TIMECREATED is default.
	SortBy ListGlossaryTermsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either 'asc' or 'desc'.
	SortOrder ListGlossaryTermsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

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

func (request ListGlossaryTermsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListGlossaryTermsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListGlossaryTermsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListGlossaryTermsResponse wrapper for the ListGlossaryTerms operation
type ListGlossaryTermsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []TermSummary instances
	Items []TermSummary `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListGlossaryTermsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListGlossaryTermsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListGlossaryTermsFieldsEnum Enum with underlying type: string
type ListGlossaryTermsFieldsEnum string

// Set of constants representing the allowable values for ListGlossaryTermsFieldsEnum
const (
	ListGlossaryTermsFieldsKey                       ListGlossaryTermsFieldsEnum = "key"
	ListGlossaryTermsFieldsDisplayname               ListGlossaryTermsFieldsEnum = "displayName"
	ListGlossaryTermsFieldsDescription               ListGlossaryTermsFieldsEnum = "description"
	ListGlossaryTermsFieldsGlossarykey               ListGlossaryTermsFieldsEnum = "glossaryKey"
	ListGlossaryTermsFieldsParenttermkey             ListGlossaryTermsFieldsEnum = "parentTermKey"
	ListGlossaryTermsFieldsIsallowedtohavechildterms ListGlossaryTermsFieldsEnum = "isAllowedToHaveChildTerms"
	ListGlossaryTermsFieldsPath                      ListGlossaryTermsFieldsEnum = "path"
	ListGlossaryTermsFieldsLifecyclestate            ListGlossaryTermsFieldsEnum = "lifecycleState"
	ListGlossaryTermsFieldsTimecreated               ListGlossaryTermsFieldsEnum = "timeCreated"
	ListGlossaryTermsFieldsWorkflowstatus            ListGlossaryTermsFieldsEnum = "workflowStatus"
	ListGlossaryTermsFieldsAssociatedobjectcount     ListGlossaryTermsFieldsEnum = "associatedObjectCount"
	ListGlossaryTermsFieldsUri                       ListGlossaryTermsFieldsEnum = "uri"
)

var mappingListGlossaryTermsFields = map[string]ListGlossaryTermsFieldsEnum{
	"key":                       ListGlossaryTermsFieldsKey,
	"displayName":               ListGlossaryTermsFieldsDisplayname,
	"description":               ListGlossaryTermsFieldsDescription,
	"glossaryKey":               ListGlossaryTermsFieldsGlossarykey,
	"parentTermKey":             ListGlossaryTermsFieldsParenttermkey,
	"isAllowedToHaveChildTerms": ListGlossaryTermsFieldsIsallowedtohavechildterms,
	"path":                  ListGlossaryTermsFieldsPath,
	"lifecycleState":        ListGlossaryTermsFieldsLifecyclestate,
	"timeCreated":           ListGlossaryTermsFieldsTimecreated,
	"workflowStatus":        ListGlossaryTermsFieldsWorkflowstatus,
	"associatedObjectCount": ListGlossaryTermsFieldsAssociatedobjectcount,
	"uri": ListGlossaryTermsFieldsUri,
}

// GetListGlossaryTermsFieldsEnumValues Enumerates the set of values for ListGlossaryTermsFieldsEnum
func GetListGlossaryTermsFieldsEnumValues() []ListGlossaryTermsFieldsEnum {
	values := make([]ListGlossaryTermsFieldsEnum, 0)
	for _, v := range mappingListGlossaryTermsFields {
		values = append(values, v)
	}
	return values
}

// ListGlossaryTermsSortByEnum Enum with underlying type: string
type ListGlossaryTermsSortByEnum string

// Set of constants representing the allowable values for ListGlossaryTermsSortByEnum
const (
	ListGlossaryTermsSortByTimecreated ListGlossaryTermsSortByEnum = "TIMECREATED"
	ListGlossaryTermsSortByDisplayname ListGlossaryTermsSortByEnum = "DISPLAYNAME"
)

var mappingListGlossaryTermsSortBy = map[string]ListGlossaryTermsSortByEnum{
	"TIMECREATED": ListGlossaryTermsSortByTimecreated,
	"DISPLAYNAME": ListGlossaryTermsSortByDisplayname,
}

// GetListGlossaryTermsSortByEnumValues Enumerates the set of values for ListGlossaryTermsSortByEnum
func GetListGlossaryTermsSortByEnumValues() []ListGlossaryTermsSortByEnum {
	values := make([]ListGlossaryTermsSortByEnum, 0)
	for _, v := range mappingListGlossaryTermsSortBy {
		values = append(values, v)
	}
	return values
}

// ListGlossaryTermsSortOrderEnum Enum with underlying type: string
type ListGlossaryTermsSortOrderEnum string

// Set of constants representing the allowable values for ListGlossaryTermsSortOrderEnum
const (
	ListGlossaryTermsSortOrderAsc  ListGlossaryTermsSortOrderEnum = "ASC"
	ListGlossaryTermsSortOrderDesc ListGlossaryTermsSortOrderEnum = "DESC"
)

var mappingListGlossaryTermsSortOrder = map[string]ListGlossaryTermsSortOrderEnum{
	"ASC":  ListGlossaryTermsSortOrderAsc,
	"DESC": ListGlossaryTermsSortOrderDesc,
}

// GetListGlossaryTermsSortOrderEnumValues Enumerates the set of values for ListGlossaryTermsSortOrderEnum
func GetListGlossaryTermsSortOrderEnumValues() []ListGlossaryTermsSortOrderEnum {
	values := make([]ListGlossaryTermsSortOrderEnum, 0)
	for _, v := range mappingListGlossaryTermsSortOrder {
		values = append(values, v)
	}
	return values
}
