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

// JobSummary Condensed Job data when listing jobs.
type JobSummary struct {

	// The OCID of the job.
	Id *string `mandatory:"false" json:"id"`

	// The name of the job, job name must consist of lower case alphanumeric characters,
	// '-' or '.', and must start and end with an alphanumeric character.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The OCID of the compute environment.
	ComputeEnvironmentId *string `mandatory:"false" json:"computeEnvironmentId"`

	// The OCID of the job definition.
	JobDefinitionId *string `mandatory:"false" json:"jobDefinitionId"`

	// The OCID of the batch instance.
	BatchInstanceId *string `mandatory:"false" json:"batchInstanceId"`

	// The docker image of the job.
	DockerImage *string `mandatory:"false" json:"dockerImage"`

	// The command used to run the Job. The command and args must can be combined into one runnable command serially,
	// such as "command": ["ls","-ll"], "args": ["/home"], the command run to the job will be "ls -ll /home" .
	// If you do not supply command or args, the defined in the job definition will be used.
	Command []string `mandatory:"false" json:"command"`

	// An OCPU is defined as the CPU capacity equivalent of one physical core of an Intel Xeon processor
	// with hyper threading enabled. OCPU for each container, for example 0.5. If set
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

	// Timeout seconds of JOB.
	TimeoutSeconds *int `mandatory:"false" json:"timeoutSeconds"`

	// The priority of the job, higher values take precedence.
	Priority *int `mandatory:"false" json:"priority"`

	// The secret name of the docker registry.
	DockerRegistrySecret *string `mandatory:"false" json:"dockerRegistrySecret"`

	// The user OCID who created the job.
	CreatedByUserId *string `mandatory:"false" json:"createdByUserId"`

	// The current work request status of the job.
	LifecycleState JobSummaryLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Describe the backend operation details.
	StatusDescription *string `mandatory:"false" json:"statusDescription"`

	// Describe the error message if the backend operation encounter error.
	ErrorCode *string `mandatory:"false" json:"errorCode"`

	// Job name in kubernetes.
	JobKubeName *string `mandatory:"false" json:"jobKubeName"`

	// The date and time the job was submitted. Format defined by RFC3339.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The date and time the job was started. Format defined by RFC3339.
	TimeStarted *common.SDKTime `mandatory:"false" json:"timeStarted"`

	// The date and time the job was completed. Format defined by RFC3339.
	TimeCompleted *common.SDKTime `mandatory:"false" json:"timeCompleted"`

	// The OCID of the node pool.
	NodePoolId *string `mandatory:"false" json:"nodePoolId"`

	// LOG PATH for job stderr path.
	JobLogStderrPath *string `mandatory:"false" json:"jobLogStderrPath"`

	// LOG PATH for job stdout path.
	JobLogStdoutPath *string `mandatory:"false" json:"jobLogStdoutPath"`

	// Free-form tags associated with this resource. Each tag is a key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m JobSummary) String() string {
	return common.PointerString(m)
}

// JobSummaryLifecycleStateEnum Enum with underlying type: string
type JobSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for JobSummaryLifecycleStateEnum
const (
	JobSummaryLifecycleStateAccepted   JobSummaryLifecycleStateEnum = "ACCEPTED"
	JobSummaryLifecycleStateInProgress JobSummaryLifecycleStateEnum = "IN_PROGRESS"
	JobSummaryLifecycleStateFailed     JobSummaryLifecycleStateEnum = "FAILED"
	JobSummaryLifecycleStateSucceeded  JobSummaryLifecycleStateEnum = "SUCCEEDED"
	JobSummaryLifecycleStateCanceling  JobSummaryLifecycleStateEnum = "CANCELING"
	JobSummaryLifecycleStateCanceled   JobSummaryLifecycleStateEnum = "CANCELED"
	JobSummaryLifecycleStateDeleted    JobSummaryLifecycleStateEnum = "DELETED"
)

var mappingJobSummaryLifecycleState = map[string]JobSummaryLifecycleStateEnum{
	"ACCEPTED":    JobSummaryLifecycleStateAccepted,
	"IN_PROGRESS": JobSummaryLifecycleStateInProgress,
	"FAILED":      JobSummaryLifecycleStateFailed,
	"SUCCEEDED":   JobSummaryLifecycleStateSucceeded,
	"CANCELING":   JobSummaryLifecycleStateCanceling,
	"CANCELED":    JobSummaryLifecycleStateCanceled,
	"DELETED":     JobSummaryLifecycleStateDeleted,
}

// GetJobSummaryLifecycleStateEnumValues Enumerates the set of values for JobSummaryLifecycleStateEnum
func GetJobSummaryLifecycleStateEnumValues() []JobSummaryLifecycleStateEnum {
	values := make([]JobSummaryLifecycleStateEnum, 0)
	for _, v := range mappingJobSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}
