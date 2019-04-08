// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
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

// ComputeEnvironment Compute environment is definition of nodepool, does NOT mean nodepool will
// be created immediately.
type ComputeEnvironment struct {

	// The OCID of the compute environment.
	Id *string `mandatory:"false" json:"id"`

	// The name of the compute environment.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The OCID of the batch instance.
	BatchInstanceId *string `mandatory:"false" json:"batchInstanceId"`

	// The machine image name.
	MachineImageName *string `mandatory:"false" json:"machineImageName"`

	// The shape for compute environment
	ShapeName *string `mandatory:"false" json:"shapeName"`

	// The OCID for subnet.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// Kubernetes version.
	KubeVersion *string `mandatory:"false" json:"kubeVersion"`

	// The date and time the compute environment was created. Format defined by RFC3339.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The user OCID who created the compute environment.
	CreatedByUserId *string `mandatory:"false" json:"createdByUserId"`

	// The user OCID who modified the compute environment.
	ModifiedByUserId *string `mandatory:"false" json:"modifiedByUserId"`

	// The current state of the compute environment.
	// - ACTIVE state means the compute environment is ready to use, user can select
	// the compute environment when submitting job.
	// - DISABLING is in process of disable, it is a transient state on the way to INACTIVE, waiting for all running
	// job complete in the compute environment, user can not select the compute environment when submitting job.
	// - INACTIVE means user can not select the compute environment when submitting job,
	// all job finished in the compute environment, ready for delete.
	// - DELETED means cascade delete the compute environment's resource, including node pool, worker node.
	LifecycleState ComputeEnvironmentLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Describe the backend operation details.
	StatusDescription *string `mandatory:"false" json:"statusDescription"`

	// Describe the error message if the backend operation encounter error.
	ErrorCode *string `mandatory:"false" json:"errorCode"`

	// Auto Scale Down
	IsAutoScaleDown *bool `mandatory:"false" json:"isAutoScaleDown"`

	// A public key is OpenSSH .pub format that can be used for verifying digital signatures generated
	// using a corresponding private key, you need generate a new public/private rsa key pair
	// or use existing ssh key. For example:
	// ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDd25ZgCEms2Cnt922S4PZVQolmvPDLJsWG8dAGEijlqPh7vepzJvCayaIymU6C6DEDtAqRN/CPm6tcIG/TFvy4al9pseIXAngfPfwNoC1jYdBYM941cEt2legcmkBCoB/wIK69SefRbO3nfbLxh/2ebtRWTJey5658wUS3JODoE9wd22EAg87I0P2Fbpo1W3kVZqF+cj7x0+t1ewZ4Rg2Bf98+hs9U9JmnmgPdk7cpo9CfF6FoiSRMMWb1kxaqESP8Q/gleajk6g1GZQkE7hEy9OxwI1QpLaAy/557vD/wJ5C0di9h+dA5gYe0QXeBeZ6zPlllJhilWehPtJIfT5hC57ks9+fBwZPqNwE92lICq5tiU8PfpamqRb1F1KiPN88G2fNUKGJHejN5DziKw6b4+RzzLneRv5VtK/FGm9wPGRdRhLzi7Wk59um9NDvd63GDV5ebQCjYBOGd1B82S9bpZlSHoewWXL9yavL5un5X8+/fETXlUkkKB4DRuKU6/aSbe0tKynngY0ZsdyJ/OcS1UbibOAXrt/AYl2/g15gWFYIRvm7VC20immiT4wf1B2fi87o5fbHfWuViJsxhjG4Eb1/0rTkJCTPV8RnNnjiKUJ9k7SRsw+NaK88MNFye0E7sTvl3Z+5vcuKZRatSVdRuP0XztvfyjXmlx2goM/dWMw== jet_sample_ww_grp@oracle.com
	SshPublicKey *string `mandatory:"false" json:"sshPublicKey"`

	// Free-form tags associated with this resource. Each tag is a key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m ComputeEnvironment) String() string {
	return common.PointerString(m)
}

// ComputeEnvironmentLifecycleStateEnum Enum with underlying type: string
type ComputeEnvironmentLifecycleStateEnum string

// Set of constants representing the allowable values for ComputeEnvironmentLifecycleStateEnum
const (
	ComputeEnvironmentLifecycleStateActive    ComputeEnvironmentLifecycleStateEnum = "ACTIVE"
	ComputeEnvironmentLifecycleStateDisabling ComputeEnvironmentLifecycleStateEnum = "DISABLING"
	ComputeEnvironmentLifecycleStateInactive  ComputeEnvironmentLifecycleStateEnum = "INACTIVE"
	ComputeEnvironmentLifecycleStateDeleted   ComputeEnvironmentLifecycleStateEnum = "DELETED"
)

var mappingComputeEnvironmentLifecycleState = map[string]ComputeEnvironmentLifecycleStateEnum{
	"ACTIVE":    ComputeEnvironmentLifecycleStateActive,
	"DISABLING": ComputeEnvironmentLifecycleStateDisabling,
	"INACTIVE":  ComputeEnvironmentLifecycleStateInactive,
	"DELETED":   ComputeEnvironmentLifecycleStateDeleted,
}

// GetComputeEnvironmentLifecycleStateEnumValues Enumerates the set of values for ComputeEnvironmentLifecycleStateEnum
func GetComputeEnvironmentLifecycleStateEnumValues() []ComputeEnvironmentLifecycleStateEnum {
	values := make([]ComputeEnvironmentLifecycleStateEnum, 0)
	for _, v := range mappingComputeEnvironmentLifecycleState {
		values = append(values, v)
	}
	return values
}
