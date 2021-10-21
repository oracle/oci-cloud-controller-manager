// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// File Storage API
//
// API for the File Storage service. Use this API to manage file systems, mount targets, and snapshots. For more information, see Overview of File Storage (https://docs.cloud.oracle.com/iaas/Content/File/Concepts/filestorageoverview.htm).
//

package filestorage

import (
	"github.com/oracle/oci-go-sdk/v49/common"
)

// ReplicationTarget resource that has information about cross region replication between source
// and target filsystems in the target region.
// All mutations(except delete) can only be done through corresponding replication resource.
// Deleting resource in target does not delete the source resource and sets resource into FAILED state.
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type ReplicationTarget struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment that contains the replication.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the replication.
	Id *string `mandatory:"true" json:"id"`

	// The current state of this replication.
	LifecycleState ReplicationTargetLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// A user-friendly name. This name is same as the replication display name for the associated resource.
	// Example: `My Replication`
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The date and time the replication target was created in target region.
	// in RFC 3339 (https://tools.ietf.org/rfc/rfc3339) timestamp format.
	// Example: `2021-01-04T20:01:29.100Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of source filesystem.
	SourceId *string `mandatory:"true" json:"sourceId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of target filesystem.
	TargetId *string `mandatory:"true" json:"targetId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of replication.
	ReplicationId *string `mandatory:"true" json:"replicationId"`

	// The availability domain the replication is in. May be unset
	// as a blank or NULL value.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of last snapshot snapshot which was completely applied to target filesystem.
	// Empty while the first snapshot is being applied.
	LastSnapshotId *string `mandatory:"false" json:"lastSnapshotId"`

	// The snapshotTime of the most recent recoverable replication snapshot
	// in RFC 3339 (https://tools.ietf.org/rfc/rfc3339) timestamp format.
	// Example: `2021-04-04T20:01:29.100Z`
	RecoveryPointTime *common.SDKTime `mandatory:"false" json:"recoveryPointTime"`

	// The current state of the snapshot in-flight.
	DeltaStatus ReplicationTargetDeltaStatusEnum `mandatory:"false" json:"deltaStatus,omitempty"`

	// Percentage progress of the snapshot in-flight which is currently being applied to the target region.
	// When nothing is in progress then it is at 100. Value is 0 before the first snapshot is applied.
	ApplyingReplicationProgress *int64 `mandatory:"false" json:"applyingReplicationProgress"`

	// Free-form tags for this resource. Each tag is a simple key-value pair
	//  with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// Additional information about the current 'lifecycleState'.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`
}

func (m ReplicationTarget) String() string {
	return common.PointerString(m)
}

// ReplicationTargetLifecycleStateEnum Enum with underlying type: string
type ReplicationTargetLifecycleStateEnum string

// Set of constants representing the allowable values for ReplicationTargetLifecycleStateEnum
const (
	ReplicationTargetLifecycleStateCreating ReplicationTargetLifecycleStateEnum = "CREATING"
	ReplicationTargetLifecycleStateActive   ReplicationTargetLifecycleStateEnum = "ACTIVE"
	ReplicationTargetLifecycleStateDeleting ReplicationTargetLifecycleStateEnum = "DELETING"
	ReplicationTargetLifecycleStateDeleted  ReplicationTargetLifecycleStateEnum = "DELETED"
	ReplicationTargetLifecycleStateFailed   ReplicationTargetLifecycleStateEnum = "FAILED"
)

var mappingReplicationTargetLifecycleState = map[string]ReplicationTargetLifecycleStateEnum{
	"CREATING": ReplicationTargetLifecycleStateCreating,
	"ACTIVE":   ReplicationTargetLifecycleStateActive,
	"DELETING": ReplicationTargetLifecycleStateDeleting,
	"DELETED":  ReplicationTargetLifecycleStateDeleted,
	"FAILED":   ReplicationTargetLifecycleStateFailed,
}

// GetReplicationTargetLifecycleStateEnumValues Enumerates the set of values for ReplicationTargetLifecycleStateEnum
func GetReplicationTargetLifecycleStateEnumValues() []ReplicationTargetLifecycleStateEnum {
	values := make([]ReplicationTargetLifecycleStateEnum, 0)
	for _, v := range mappingReplicationTargetLifecycleState {
		values = append(values, v)
	}
	return values
}

// ReplicationTargetDeltaStatusEnum Enum with underlying type: string
type ReplicationTargetDeltaStatusEnum string

// Set of constants representing the allowable values for ReplicationTargetDeltaStatusEnum
const (
	ReplicationTargetDeltaStatusIdle         ReplicationTargetDeltaStatusEnum = "IDLE"
	ReplicationTargetDeltaStatusCapturing    ReplicationTargetDeltaStatusEnum = "CAPTURING"
	ReplicationTargetDeltaStatusApplying     ReplicationTargetDeltaStatusEnum = "APPLYING"
	ReplicationTargetDeltaStatusServiceError ReplicationTargetDeltaStatusEnum = "SERVICE_ERROR"
	ReplicationTargetDeltaStatusUserError    ReplicationTargetDeltaStatusEnum = "USER_ERROR"
	ReplicationTargetDeltaStatusFailed       ReplicationTargetDeltaStatusEnum = "FAILED"
)

var mappingReplicationTargetDeltaStatus = map[string]ReplicationTargetDeltaStatusEnum{
	"IDLE":          ReplicationTargetDeltaStatusIdle,
	"CAPTURING":     ReplicationTargetDeltaStatusCapturing,
	"APPLYING":      ReplicationTargetDeltaStatusApplying,
	"SERVICE_ERROR": ReplicationTargetDeltaStatusServiceError,
	"USER_ERROR":    ReplicationTargetDeltaStatusUserError,
	"FAILED":        ReplicationTargetDeltaStatusFailed,
}

// GetReplicationTargetDeltaStatusEnumValues Enumerates the set of values for ReplicationTargetDeltaStatusEnum
func GetReplicationTargetDeltaStatusEnumValues() []ReplicationTargetDeltaStatusEnum {
	values := make([]ReplicationTargetDeltaStatusEnum, 0)
	for _, v := range mappingReplicationTargetDeltaStatus {
		values = append(values, v)
	}
	return values
}
