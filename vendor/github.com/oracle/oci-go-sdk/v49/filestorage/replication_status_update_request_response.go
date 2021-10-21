// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package filestorage

import (
	"github.com/oracle/oci-go-sdk/v49/common"
	"net/http"
)

// ReplicationStatusUpdateRequest wrapper for the ReplicationStatusUpdate operation
type ReplicationStatusUpdateRequest struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the Replication
	ReplicationId *string `mandatory:"true" contributesTo:"query" name:"replicationId"`

	// The deltaState of the snapshot in-flight.
	DeltaState ReplicationStatusUpdateDeltaStateEnum `mandatory:"false" contributesTo:"query" name:"deltaState" omitEmpty:"true"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ReplicationStatusUpdateRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ReplicationStatusUpdateRequest) HTTPRequest(method, path string, binaryRequestBody *common.OCIReadSeekCloser, extraHeaders map[string]string) (http.Request, error) {

	return common.MakeDefaultHTTPRequestWithTaggedStructAndExtraHeaders(method, path, request, extraHeaders)
}

// BinaryRequestBody implements the OCIRequest interface
func (request ReplicationStatusUpdateRequest) BinaryRequestBody() (*common.OCIReadSeekCloser, bool) {

	return nil, false

}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ReplicationStatusUpdateRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ReplicationStatusUpdateResponse wrapper for the ReplicationStatusUpdate operation
type ReplicationStatusUpdateResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Replication instance
	Replication `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ReplicationStatusUpdateResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ReplicationStatusUpdateResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ReplicationStatusUpdateDeltaStateEnum Enum with underlying type: string
type ReplicationStatusUpdateDeltaStateEnum string

// Set of constants representing the allowable values for ReplicationStatusUpdateDeltaStateEnum
const (
	ReplicationStatusUpdateDeltaStateReadyToReplicate     ReplicationStatusUpdateDeltaStateEnum = "READY_TO_REPLICATE"
	ReplicationStatusUpdateDeltaStateReplicating          ReplicationStatusUpdateDeltaStateEnum = "REPLICATING"
	ReplicationStatusUpdateDeltaStateReplicated           ReplicationStatusUpdateDeltaStateEnum = "REPLICATED"
	ReplicationStatusUpdateDeltaStateReplicatingFailed    ReplicationStatusUpdateDeltaStateEnum = "REPLICATING_FAILED"
	ReplicationStatusUpdateDeltaStateAbortReplication     ReplicationStatusUpdateDeltaStateEnum = "ABORT_REPLICATION"
	ReplicationStatusUpdateDeltaStateAbortReplicationDone ReplicationStatusUpdateDeltaStateEnum = "ABORT_REPLICATION_DONE"
	ReplicationStatusUpdateDeltaStateDone                 ReplicationStatusUpdateDeltaStateEnum = "DONE"
	ReplicationStatusUpdateDeltaStateReadyToGc            ReplicationStatusUpdateDeltaStateEnum = "READY_TO_GC"
	ReplicationStatusUpdateDeltaStateDeleted              ReplicationStatusUpdateDeltaStateEnum = "DELETED"
)

var mappingReplicationStatusUpdateDeltaState = map[string]ReplicationStatusUpdateDeltaStateEnum{
	"READY_TO_REPLICATE":     ReplicationStatusUpdateDeltaStateReadyToReplicate,
	"REPLICATING":            ReplicationStatusUpdateDeltaStateReplicating,
	"REPLICATED":             ReplicationStatusUpdateDeltaStateReplicated,
	"REPLICATING_FAILED":     ReplicationStatusUpdateDeltaStateReplicatingFailed,
	"ABORT_REPLICATION":      ReplicationStatusUpdateDeltaStateAbortReplication,
	"ABORT_REPLICATION_DONE": ReplicationStatusUpdateDeltaStateAbortReplicationDone,
	"DONE":                   ReplicationStatusUpdateDeltaStateDone,
	"READY_TO_GC":            ReplicationStatusUpdateDeltaStateReadyToGc,
	"DELETED":                ReplicationStatusUpdateDeltaStateDeleted,
}

// GetReplicationStatusUpdateDeltaStateEnumValues Enumerates the set of values for ReplicationStatusUpdateDeltaStateEnum
func GetReplicationStatusUpdateDeltaStateEnumValues() []ReplicationStatusUpdateDeltaStateEnum {
	values := make([]ReplicationStatusUpdateDeltaStateEnum, 0)
	for _, v := range mappingReplicationStatusUpdateDeltaState {
		values = append(values, v)
	}
	return values
}
