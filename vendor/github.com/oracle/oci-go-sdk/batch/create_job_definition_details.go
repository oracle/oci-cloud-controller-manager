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

// CreateJobDefinitionDetails Details for creating a new job definition.
type CreateJobDefinitionDetails struct {

	// The OCID of the batch instance.
	BatchInstanceId *string `mandatory:"true" json:"batchInstanceId"`

	// The docker image of the job.
	DockerImage *string `mandatory:"true" json:"dockerImage"`

	// An OCPU is defined as the CPU capacity equivalent of one physical core of an Intel Xeon processor
	// with hyper threading enabled. OCPU for each container, for example 0.5.
	ContainerOcpu *float32 `mandatory:"true" json:"containerOcpu"`

	// MB of memory for each container, for example 512.
	ContainerMemorySizeInMbs *int `mandatory:"true" json:"containerMemorySizeInMbs"`

	// The name of the job definition. When not provided, the system generate value using the format
	// "<resourceType><timestamp>", example: jobDefinition20181211220642.
	DisplayName *string `mandatory:"false" json:"displayName"`

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

	// Free-form tags associated with this resource. Each tag is a key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m CreateJobDefinitionDetails) String() string {
	return common.PointerString(m)
}
