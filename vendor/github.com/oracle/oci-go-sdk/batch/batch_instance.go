// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Oracle Batch Service
//
// This is a Oracle Batch Service. You can find out more about at
// wiki (https://confluence.oraclecorp.com/confluence/display/C9QA/OCI+Batch+Service+-+Core+Functionality+Technical+Design+and+Explanation+for+Phase+I).
//

package batch

import (
	"github.com/oracle/oci-go-sdk/common"
)

// BatchInstance batch instance
type BatchInstance struct {

	// The OCID of the batch instance.
	Id *string `mandatory:"false" json:"id"`

	// The name of the batch instance.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The OCID of the cluster.
	ClusterId *string `mandatory:"false" json:"clusterId"`

	// The user OCID who created the batch instance.
	CreatedByUserId *string `mandatory:"false" json:"createdByUserId"`

	// The user OCID who deleted the batch instance.
	DeletedByUserId *string `mandatory:"false" json:"deletedByUserId"`

	// The Kubernetes namespace containing the batch instance.
	Namespace *string `mandatory:"false" json:"namespace"`

	// The date and time the batch instance was created. Format defined by RFC3339.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The date and time the batch instance was deleted. Format defined by RFC3339.
	TimeDeleted *common.SDKTime `mandatory:"false" json:"timeDeleted"`

	// The current state of the batch instance.
	// - ACTIVE state means the batch instance is ready for customer to use.
	// - DISABLING is in process of disable, it is a transient state on the way to INACTIVE, the batch instance is
	// in read-only mode, not allow any resource creation (compute environment, job definition, job).
	// - INACTIVE means the batch instance is in read-only mode, all job finished in the batch instance,
	// ready for delete.
	// - DELETED means cascade delete the batch instance's resources.
	LifecycleState BatchInstanceLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The user OCID who modified the batch instance.
	ModifiedByUserId *string `mandatory:"false" json:"modifiedByUserId"`

	// Free-form tags associated with this resource. Each tag is a key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m BatchInstance) String() string {
	return common.PointerString(m)
}

// BatchInstanceLifecycleStateEnum Enum with underlying type: string
type BatchInstanceLifecycleStateEnum string

// Set of constants representing the allowable values for BatchInstanceLifecycleStateEnum
const (
	BatchInstanceLifecycleStateActive    BatchInstanceLifecycleStateEnum = "ACTIVE"
	BatchInstanceLifecycleStateDisabling BatchInstanceLifecycleStateEnum = "DISABLING"
	BatchInstanceLifecycleStateInactive  BatchInstanceLifecycleStateEnum = "INACTIVE"
	BatchInstanceLifecycleStateDeleted   BatchInstanceLifecycleStateEnum = "DELETED"
)

var mappingBatchInstanceLifecycleState = map[string]BatchInstanceLifecycleStateEnum{
	"ACTIVE":    BatchInstanceLifecycleStateActive,
	"DISABLING": BatchInstanceLifecycleStateDisabling,
	"INACTIVE":  BatchInstanceLifecycleStateInactive,
	"DELETED":   BatchInstanceLifecycleStateDeleted,
}

// GetBatchInstanceLifecycleStateEnumValues Enumerates the set of values for BatchInstanceLifecycleStateEnum
func GetBatchInstanceLifecycleStateEnumValues() []BatchInstanceLifecycleStateEnum {
	values := make([]BatchInstanceLifecycleStateEnum, 0)
	for _, v := range mappingBatchInstanceLifecycleState {
		values = append(values, v)
	}
	return values
}
