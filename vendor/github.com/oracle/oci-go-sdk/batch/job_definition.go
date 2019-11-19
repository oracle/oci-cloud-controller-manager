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

// JobDefinition Job definition to define a job and specify docker image, command, env,
// timeout, etc
type JobDefinition struct {

	// The OCID of the job definition.
	Id *string `mandatory:"false" json:"id"`

	// The name of the job definition.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The OCID of the batch instance.
	BatchInstanceId *string `mandatory:"false" json:"batchInstanceId"`

	// The docker image of the job.
	DockerImage *string `mandatory:"false" json:"dockerImage"`

	// An OCPU is defined as the CPU capacity equivalent of one physical core of an Intel Xeon processor
	// with hyper threading enabled. OCPU for each container, for example 0.5.
	ContainerOcpu *float32 `mandatory:"false" json:"containerOcpu"`

	// MB of memory for each container, for example 512.
	ContainerMemorySizeInMbs *int `mandatory:"false" json:"containerMemorySizeInMbs"`

	// The secret name of the docker registry.
	DockerRegistrySecret *string `mandatory:"false" json:"dockerRegistrySecret"`

	// Specifies the number of retries before marking this job failed.
	RetryTimes *int `mandatory:"false" json:"retryTimes"`

	// Timeout seconds of JOB.
	TimeoutSeconds *int `mandatory:"false" json:"timeoutSeconds"`

	// The command used to run the Job. The command and args must can be combined into one runnable command serially,
	// such as "command": ["ls","-ll"], "args": ["/home"], the command run to the job will be "ls -ll /home" .
	// If you do not supply command or args, the defaults defined in the job will be used.
	Command []string `mandatory:"false" json:"command"`

	// The arguments passed to the command to run the Job, it's multiple such as
	// ["-a","-k"]. If you do not supply command or args, the defaults defined in the job will be used.
	Args []string `mandatory:"false" json:"args"`

	// Environment variables used to run the JOB - user provided data -
	// it's multiple such as [{"name": "xxx","value": "xxx"}].
	EnvironmentVariables []CreateJobDefinitionDetailsEnvironmentVariables `mandatory:"false" json:"environmentVariables"`

	// The date and time the job definition was created. Format defined by RFC3339.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The user OCID who created the job definition.
	CreatedByUserId *string `mandatory:"false" json:"createdByUserId"`

	// The user OCID who modified the job definition.
	ModifiedByUserId *string `mandatory:"false" json:"modifiedByUserId"`

	// The current state of the job definition.
	LifecycleState JobDefinitionLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Free-form tags associated with this resource. Each tag is a key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m JobDefinition) String() string {
	return common.PointerString(m)
}

// JobDefinitionLifecycleStateEnum Enum with underlying type: string
type JobDefinitionLifecycleStateEnum string

// Set of constants representing the allowable values for JobDefinitionLifecycleStateEnum
const (
	JobDefinitionLifecycleStateActive  JobDefinitionLifecycleStateEnum = "ACTIVE"
	JobDefinitionLifecycleStateDeleted JobDefinitionLifecycleStateEnum = "DELETED"
)

var mappingJobDefinitionLifecycleState = map[string]JobDefinitionLifecycleStateEnum{
	"ACTIVE":  JobDefinitionLifecycleStateActive,
	"DELETED": JobDefinitionLifecycleStateDeleted,
}

// GetJobDefinitionLifecycleStateEnumValues Enumerates the set of values for JobDefinitionLifecycleStateEnum
func GetJobDefinitionLifecycleStateEnumValues() []JobDefinitionLifecycleStateEnum {
	values := make([]JobDefinitionLifecycleStateEnum, 0)
	for _, v := range mappingJobDefinitionLifecycleState {
		values = append(values, v)
	}
	return values
}
