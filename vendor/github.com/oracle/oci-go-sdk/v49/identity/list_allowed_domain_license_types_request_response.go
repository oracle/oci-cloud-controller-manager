// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package identity

import (
	"github.com/oracle/oci-go-sdk/v49/common"
	"net/http"
)

// ListAllowedDomainLicenseTypesRequest wrapper for the ListAllowedDomainLicenseTypes operation
type ListAllowedDomainLicenseTypesRequest struct {

	// The domain license type
	CurrentLicenseTypeName *string `mandatory:"false" contributesTo:"query" name:"currentLicenseTypeName"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The maximum number of items to return in a paginated "List" call.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// A filter to only return resources that match the given name exactly.
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for NAME is ascending. The NAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by Availability Domain if the scope of the resource type is within a
	// single Availability Domain. If you call one of these "List" operations without specifying
	// an Availability Domain, the resources are grouped by Availability Domain, then sorted.
	SortBy ListAllowedDomainLicenseTypesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The NAME sort order
	// is case sensitive.
	SortOrder ListAllowedDomainLicenseTypesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListAllowedDomainLicenseTypesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListAllowedDomainLicenseTypesRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ListAllowedDomainLicenseTypesRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListAllowedDomainLicenseTypesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListAllowedDomainLicenseTypesResponse wrapper for the ListAllowedDomainLicenseTypes operation
type ListAllowedDomainLicenseTypesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []AllowedDomainLicenseTypeSummary instances
	Items []AllowedDomainLicenseTypeSummary `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For pagination of a list of items. When paging through a list, if this header appears in the response,
	// then a partial list might have been returned. Include this value as the `page` parameter for the
	// subsequent GET request to get the next batch of items.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListAllowedDomainLicenseTypesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListAllowedDomainLicenseTypesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListAllowedDomainLicenseTypesSortByEnum Enum with underlying type: string
type ListAllowedDomainLicenseTypesSortByEnum string

// Set of constants representing the allowable values for ListAllowedDomainLicenseTypesSortByEnum
const (
	ListAllowedDomainLicenseTypesSortByTimecreated ListAllowedDomainLicenseTypesSortByEnum = "TIMECREATED"
	ListAllowedDomainLicenseTypesSortByName        ListAllowedDomainLicenseTypesSortByEnum = "NAME"
)

var mappingListAllowedDomainLicenseTypesSortBy = map[string]ListAllowedDomainLicenseTypesSortByEnum{
	"TIMECREATED": ListAllowedDomainLicenseTypesSortByTimecreated,
	"NAME":        ListAllowedDomainLicenseTypesSortByName,
}

// GetListAllowedDomainLicenseTypesSortByEnumValues Enumerates the set of values for ListAllowedDomainLicenseTypesSortByEnum
func GetListAllowedDomainLicenseTypesSortByEnumValues() []ListAllowedDomainLicenseTypesSortByEnum {
	values := make([]ListAllowedDomainLicenseTypesSortByEnum, 0)
	for _, v := range mappingListAllowedDomainLicenseTypesSortBy {
		values = append(values, v)
	}
	return values
}

// ListAllowedDomainLicenseTypesSortOrderEnum Enum with underlying type: string
type ListAllowedDomainLicenseTypesSortOrderEnum string

// Set of constants representing the allowable values for ListAllowedDomainLicenseTypesSortOrderEnum
const (
	ListAllowedDomainLicenseTypesSortOrderAsc  ListAllowedDomainLicenseTypesSortOrderEnum = "ASC"
	ListAllowedDomainLicenseTypesSortOrderDesc ListAllowedDomainLicenseTypesSortOrderEnum = "DESC"
)

var mappingListAllowedDomainLicenseTypesSortOrder = map[string]ListAllowedDomainLicenseTypesSortOrderEnum{
	"ASC":  ListAllowedDomainLicenseTypesSortOrderAsc,
	"DESC": ListAllowedDomainLicenseTypesSortOrderDesc,
}

// GetListAllowedDomainLicenseTypesSortOrderEnumValues Enumerates the set of values for ListAllowedDomainLicenseTypesSortOrderEnum
func GetListAllowedDomainLicenseTypesSortOrderEnumValues() []ListAllowedDomainLicenseTypesSortOrderEnum {
	values := make([]ListAllowedDomainLicenseTypesSortOrderEnum, 0)
	for _, v := range mappingListAllowedDomainLicenseTypesSortOrder {
		values = append(values, v)
	}
	return values
}
