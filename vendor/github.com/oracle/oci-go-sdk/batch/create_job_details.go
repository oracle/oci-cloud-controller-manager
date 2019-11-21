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

// CreateJobDetails Details for creating a new job.
// When submit a job, user can select a defined Compute Environment and Job Definition,
// job properties will override those of referenced job definition.
type CreateJobDetails struct {

	// The OCID of the compute environment.
	ComputeEnvironmentId *string `mandatory:"true" json:"computeEnvironmentId"`

	// The OCID of the job definition.
	JobDefinitionId *string `mandatory:"true" json:"jobDefinitionId"`

	// The OCID of the batch instance.
	BatchInstanceId *string `mandatory:"true" json:"batchInstanceId"`

	// The name of the job, job name must consist of lower case alphanumeric characters,
	// '-' or '.', and must start and end with an alphanumeric character.
	// When not provided, the system generate value using the format
	// "<resourceType><timestamp>", example: job20181211220642.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The docker image of the job.
	DockerImage *string `mandatory:"false" json:"dockerImage"`

	// The command used to run the Job. The command and args must can be combined into one runnable command serially,
	// such as "command": ["ls","-ll"], "args": ["/home"], the command run to the job will be "ls -ll /home" .
	// If you do not supply command or args, the defined in the job definition will be used.
	Command []string `mandatory:"false" json:"command"`

	// The arguments passed to the command to run the Job, it's multiple such as
	// ["-a","-k"]. If you do not supply command or args, the defined in the job definition will be used.
	Args []string `mandatory:"false" json:"args"`

	// Environment variables used to run the JOB - user provided data -
	// it's multiple such as [{"name": "xxx","value": "xxx"}].
	EnvironmentVariables []CreateJobDefinitionDetailsEnvironmentVariables `mandatory:"false" json:"environmentVariables"`

	// An OCPU is defined as the CPU capacity equivalent of one physical core of an Intel Xeon processor
	// with hyper threading enabled. OCPU for each container, for example 0.5.
	ContainerOcpu *float32 `mandatory:"false" json:"containerOcpu"`

	// MB of memory for each container, for example 512.
	ContainerMemorySizeInMbs *int `mandatory:"false" json:"containerMemorySizeInMbs"`

	// Number of pods running on a job concurrently.
	Concurrency *int `mandatory:"false" json:"concurrency"`

	// The number of pods that must successfully terminate before a job can reach the SUCCEEDED state.
	// Pods that terminate unsuccessfully are retried until retryTimes has been exhausted or a job times out.
	Count *int `mandatory:"false" json:"count"`

	// Specifies the number of retries before marking this job failed.
	RetryTimes *int `mandatory:"false" json:"retryTimes"`

	// Timeout seconds of JOB. The default value of timeoutSeconds is 3600.
	TimeoutSeconds *int `mandatory:"false" json:"timeoutSeconds"`

	// The priority of the job, higher values take precedence.
	Priority *int `mandatory:"false" json:"priority"`

	// The secret name of the docker registry.
	DockerRegistrySecret *string `mandatory:"false" json:"dockerRegistrySecret"`

	// Free-form tags associated with this resource. Each tag is a key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m CreateJobDetails) String() string {
	return common.PointerString(m)
}
