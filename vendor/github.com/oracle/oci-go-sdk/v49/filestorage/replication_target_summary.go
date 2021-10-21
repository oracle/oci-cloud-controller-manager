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

// ReplicationTargetSummary Summary information for replication target.
type ReplicationTargetSummary struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the replication target.
	Id *string `mandatory:"true" json:"id"`

	// The current state of this replication.
	LifecycleState ReplicationTargetSummaryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// A user-friendly name. This name is same as the replication display name for the associated resource.
	// Example: `My replication`
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The date and time the replication was created
	// in RFC 3339 (https://tools.ietf.org/rfc/rfc3339) timestamp format.
	// Example: `2021-02-02T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The availability domain the replication target is in. Must be in the availability domain as in target filesystem.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment that contains the replication.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

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

	// The snapshotTime of the most recent recoverable replication snapshot
	// in RFC 3339 (https://tools.ietf.org/rfc/rfc3339) timestamp format.
	// Example: `2021-04-04T20:01:29.100Z`
	RecoveryPointTime *common.SDKTime `mandatory:"false" json:"recoveryPointTime"`
}

func (m ReplicationTargetSummary) String() string {
	return common.PointerString(m)
}

// ReplicationTargetSummaryLifecycleStateEnum Enum with underlying type: string
type ReplicationTargetSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for ReplicationTargetSummaryLifecycleStateEnum
const (
	ReplicationTargetSummaryLifecycleStateCreating ReplicationTargetSummaryLifecycleStateEnum = "CREATING"
	ReplicationTargetSummaryLifecycleStateActive   ReplicationTargetSummaryLifecycleStateEnum = "ACTIVE"
	ReplicationTargetSummaryLifecycleStateDeleting ReplicationTargetSummaryLifecycleStateEnum = "DELETING"
	ReplicationTargetSummaryLifecycleStateDeleted  ReplicationTargetSummaryLifecycleStateEnum = "DELETED"
	ReplicationTargetSummaryLifecycleStateFailed   ReplicationTargetSummaryLifecycleStateEnum = "FAILED"
)

var mappingReplicationTargetSummaryLifecycleState = map[string]ReplicationTargetSummaryLifecycleStateEnum{
	"CREATING": ReplicationTargetSummaryLifecycleStateCreating,
	"ACTIVE":   ReplicationTargetSummaryLifecycleStateActive,
	"DELETING": ReplicationTargetSummaryLifecycleStateDeleting,
	"DELETED":  ReplicationTargetSummaryLifecycleStateDeleted,
	"FAILED":   ReplicationTargetSummaryLifecycleStateFailed,
}

// GetReplicationTargetSummaryLifecycleStateEnumValues Enumerates the set of values for ReplicationTargetSummaryLifecycleStateEnum
func GetReplicationTargetSummaryLifecycleStateEnumValues() []ReplicationTargetSummaryLifecycleStateEnum {
	values := make([]ReplicationTargetSummaryLifecycleStateEnum, 0)
	for _, v := range mappingReplicationTargetSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}
