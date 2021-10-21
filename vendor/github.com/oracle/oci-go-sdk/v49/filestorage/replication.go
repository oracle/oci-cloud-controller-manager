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

// Replication resource that governs the policy of cross region replication between source
// and target filsystems.
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type Replication struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment that contains the replication.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the replication.
	Id *string `mandatory:"true" json:"id"`

	// The current state of this replication.
	LifecycleState ReplicationLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// A user-friendly name. It does not have to be unique, and it is changeable.
	// Avoid entering confidential information.
	// Example: `My replication`
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The date and time the replication was created
	// in RFC 3339 (https://tools.ietf.org/rfc/rfc3339) timestamp format.
	// Example: `2021-01-04T20:01:29.100Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of source filesystem.
	SourceId *string `mandatory:"true" json:"sourceId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of target filesystem.
	TargetId *string `mandatory:"true" json:"targetId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of replication target.
	ReplicationTargetId *string `mandatory:"true" json:"replicationTargetId"`

	// The availability domain the replication is in. May be unset as a blank or NULL value.
	// Example: `Uocm:PHX-AD-2`
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// Duration in minutes between replication snapshots.
	ReplicationInterval *int64 `mandatory:"false" json:"replicationInterval"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the last snapshot that has been replicated completely.
	// Empty if the copy of the first snapshot has not yet completed.
	LastSnapshotId *string `mandatory:"false" json:"lastSnapshotId"`

	// The snapshotTime of the most recent recoverable replication snapshot
	// in RFC 3339 (https://tools.ietf.org/rfc/rfc3339) timestamp format.
	// Example: `2021-04-04T20:01:29.100Z`
	RecoveryPointTime *common.SDKTime `mandatory:"false" json:"recoveryPointTime"`

	// The current state of the snapshot in-flight.
	DeltaStatus ReplicationDeltaStatusEnum `mandatory:"false" json:"deltaStatus,omitempty"`

	// Additional information about the current 'lifecycleState'.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// Free-form tags for this resource. Each tag is a simple key-value pair
	//  with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m Replication) String() string {
	return common.PointerString(m)
}

// ReplicationLifecycleStateEnum Enum with underlying type: string
type ReplicationLifecycleStateEnum string

// Set of constants representing the allowable values for ReplicationLifecycleStateEnum
const (
	ReplicationLifecycleStateCreating ReplicationLifecycleStateEnum = "CREATING"
	ReplicationLifecycleStateActive   ReplicationLifecycleStateEnum = "ACTIVE"
	ReplicationLifecycleStateDeleting ReplicationLifecycleStateEnum = "DELETING"
	ReplicationLifecycleStateDeleted  ReplicationLifecycleStateEnum = "DELETED"
	ReplicationLifecycleStateFailed   ReplicationLifecycleStateEnum = "FAILED"
)

var mappingReplicationLifecycleState = map[string]ReplicationLifecycleStateEnum{
	"CREATING": ReplicationLifecycleStateCreating,
	"ACTIVE":   ReplicationLifecycleStateActive,
	"DELETING": ReplicationLifecycleStateDeleting,
	"DELETED":  ReplicationLifecycleStateDeleted,
	"FAILED":   ReplicationLifecycleStateFailed,
}

// GetReplicationLifecycleStateEnumValues Enumerates the set of values for ReplicationLifecycleStateEnum
func GetReplicationLifecycleStateEnumValues() []ReplicationLifecycleStateEnum {
	values := make([]ReplicationLifecycleStateEnum, 0)
	for _, v := range mappingReplicationLifecycleState {
		values = append(values, v)
	}
	return values
}

// ReplicationDeltaStatusEnum Enum with underlying type: string
type ReplicationDeltaStatusEnum string

// Set of constants representing the allowable values for ReplicationDeltaStatusEnum
const (
	ReplicationDeltaStatusIdle         ReplicationDeltaStatusEnum = "IDLE"
	ReplicationDeltaStatusCapturing    ReplicationDeltaStatusEnum = "CAPTURING"
	ReplicationDeltaStatusApplying     ReplicationDeltaStatusEnum = "APPLYING"
	ReplicationDeltaStatusServiceError ReplicationDeltaStatusEnum = "SERVICE_ERROR"
	ReplicationDeltaStatusUserError    ReplicationDeltaStatusEnum = "USER_ERROR"
	ReplicationDeltaStatusFailed       ReplicationDeltaStatusEnum = "FAILED"
)

var mappingReplicationDeltaStatus = map[string]ReplicationDeltaStatusEnum{
	"IDLE":          ReplicationDeltaStatusIdle,
	"CAPTURING":     ReplicationDeltaStatusCapturing,
	"APPLYING":      ReplicationDeltaStatusApplying,
	"SERVICE_ERROR": ReplicationDeltaStatusServiceError,
	"USER_ERROR":    ReplicationDeltaStatusUserError,
	"FAILED":        ReplicationDeltaStatusFailed,
}

// GetReplicationDeltaStatusEnumValues Enumerates the set of values for ReplicationDeltaStatusEnum
func GetReplicationDeltaStatusEnumValues() []ReplicationDeltaStatusEnum {
	values := make([]ReplicationDeltaStatusEnum, 0)
	for _, v := range mappingReplicationDeltaStatus {
		values = append(values, v)
	}
	return values
}
