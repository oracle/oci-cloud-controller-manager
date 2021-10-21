// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package filestorage

import (
	"github.com/oracle/oci-go-sdk/v49/common"
	"net/http"
)

// DeleteReplicationRequest wrapper for the DeleteReplication operation
type DeleteReplicationRequest struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the replication.
	ReplicationId *string `mandatory:"true" contributesTo:"path" name:"replicationId"`

	// For optimistic concurrency control. In the PUT or DELETE call
	// for a resource, set the `if-match` parameter to the value of the
	// etag from a previous GET or POST response for that resource.
	// The resource will be updated or deleted only if the etag you
	// provide matches the resource's current etag value.
	IfMatch *string `mandatory:"false" contributesTo:"header" name:"if-match"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Customer can choose a mode while deleting the replication in source region.
	// - 'FINISH_CYCLE_IF_CAPTURING_OR_APPLYING' lets current snapshot which has already started, finish gracefully and does not start a new snapshot before deleting.
	// - 'ONE_MORE_CYCLE' brings the target filesystem up-to-date with source filsystem by doing one last replication cycle immediately and then deleting.
	// - 'FINISH_CYCLE_IF_APPLYING' completes the application to the target of any snapshot in the APPLYING state, otherwise delete the replication immediately.
	// 'FINISH_APPLY' is the fastest way to make target available while deleting resource in source region.
	DeleteMode DeleteReplicationDeleteModeEnum `mandatory:"false" contributesTo:"query" name:"deleteMode" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request DeleteReplicationRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request DeleteReplicationRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request DeleteReplicationRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request DeleteReplicationRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// DeleteReplicationResponse wrapper for the DeleteReplication operation
type DeleteReplicationResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// Unique Oracle-assigned identifier for the request. If
	// you need to contact Oracle about a particular request,
	// please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response DeleteReplicationResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response DeleteReplicationResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// DeleteReplicationDeleteModeEnum Enum with underlying type: string
type DeleteReplicationDeleteModeEnum string

// Set of constants representing the allowable values for DeleteReplicationDeleteModeEnum
const (
	DeleteReplicationDeleteModeFinishCycleIfCapturingOrApplying DeleteReplicationDeleteModeEnum = "FINISH_CYCLE_IF_CAPTURING_OR_APPLYING"
	DeleteReplicationDeleteModeOneMoreCycle                     DeleteReplicationDeleteModeEnum = "ONE_MORE_CYCLE"
	DeleteReplicationDeleteModeFinishCycleIfApplying            DeleteReplicationDeleteModeEnum = "FINISH_CYCLE_IF_APPLYING"
)

var mappingDeleteReplicationDeleteMode = map[string]DeleteReplicationDeleteModeEnum{
	"FINISH_CYCLE_IF_CAPTURING_OR_APPLYING": DeleteReplicationDeleteModeFinishCycleIfCapturingOrApplying,
	"ONE_MORE_CYCLE":                        DeleteReplicationDeleteModeOneMoreCycle,
	"FINISH_CYCLE_IF_APPLYING":              DeleteReplicationDeleteModeFinishCycleIfApplying,
}

// GetDeleteReplicationDeleteModeEnumValues Enumerates the set of values for DeleteReplicationDeleteModeEnum
func GetDeleteReplicationDeleteModeEnumValues() []DeleteReplicationDeleteModeEnum {
	values := make([]DeleteReplicationDeleteModeEnum, 0)
	for _, v := range mappingDeleteReplicationDeleteMode {
		values = append(values, v)
	}
	return values
}
