// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package usage

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListUsageRecordsRequest wrapper for the ListUsageRecords operation
type ListUsageRecordsRequest struct {

	// The OCID of the tenancy for which usage is being fetched.
	TenancyId *string `mandatory:"true" contributesTo:"path" name:"tenancyId"`

	// Start time (UTC) of the target date range for which to fetch usage data (inclusive).
	StartTime *common.SDKTime `mandatory:"true" contributesTo:"query" name:"startTime"`

	// End time (UTC) of the target date range for which to fetch usage data (exclusive).
	EndTime *common.SDKTime `mandatory:"true" contributesTo:"query" name:"endTime"`

	// The aggregation period for the usage data.
	Granularity ListUsageRecordsGranularityEnum `mandatory:"true" contributesTo:"query" name:"granularity" omitEmpty:"true"`

	// Optional parameter to filter the data to a specific compartment only.
	// Note that this parameter cannot be combined with the `tag` parameter.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// Optional parameter to filter the data to a specific cost-tracking tag only.
	// Note that this parameter cannot be combined with the `compartmentId` parameter.
	Tag *string `mandatory:"false" contributesTo:"query" name:"tag"`

	// Unique, Oracle-assigned identifier for the request. If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Reserved parameter for supporting pagination in the future.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListUsageRecordsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListUsageRecordsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListUsageRecordsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListUsageRecordsResponse wrapper for the ListUsageRecords operation
type ListUsageRecordsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []UsageRecord instances
	Items []UsageRecord `presentIn:"body"`

	// Reserved header for supporting pagination in the future.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListUsageRecordsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListUsageRecordsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListUsageRecordsGranularityEnum Enum with underlying type: string
type ListUsageRecordsGranularityEnum string

// Set of constants representing the allowable values for ListUsageRecordsGranularityEnum
const (
	ListUsageRecordsGranularityHourly  ListUsageRecordsGranularityEnum = "HOURLY"
	ListUsageRecordsGranularityDaily   ListUsageRecordsGranularityEnum = "DAILY"
	ListUsageRecordsGranularityMonthly ListUsageRecordsGranularityEnum = "MONTHLY"
)

var mappingListUsageRecordsGranularity = map[string]ListUsageRecordsGranularityEnum{
	"HOURLY":  ListUsageRecordsGranularityHourly,
	"DAILY":   ListUsageRecordsGranularityDaily,
	"MONTHLY": ListUsageRecordsGranularityMonthly,
}

// GetListUsageRecordsGranularityEnumValues Enumerates the set of values for ListUsageRecordsGranularityEnum
func GetListUsageRecordsGranularityEnumValues() []ListUsageRecordsGranularityEnum {
	values := make([]ListUsageRecordsGranularityEnum, 0)
	for _, v := range mappingListUsageRecordsGranularity {
		values = append(values, v)
	}
	return values
}
