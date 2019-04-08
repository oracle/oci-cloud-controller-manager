package autotest

import (
	"github.com/oracle/oci-go-sdk/batch"
	"github.com/oracle/oci-go-sdk/common"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createBatchServiceClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := batch.NewBatchServiceClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientCancelJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "CancelJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CancelJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "CancelJob", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "CancelJob")
	assert.NoError(t, err)

	type CancelJobRequestInfo struct {
		ContainerId string
		Request     batch.CancelJobRequest
	}

	var requests []CancelJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.CancelJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientCreateBatchInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "CreateBatchInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateBatchInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "CreateBatchInstance", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "CreateBatchInstance")
	assert.NoError(t, err)

	type CreateBatchInstanceRequestInfo struct {
		ContainerId string
		Request     batch.CreateBatchInstanceRequest
	}

	var requests []CreateBatchInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.CreateBatchInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientCreateComputeEnvironment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "CreateComputeEnvironment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateComputeEnvironment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "CreateComputeEnvironment", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "CreateComputeEnvironment")
	assert.NoError(t, err)

	type CreateComputeEnvironmentRequestInfo struct {
		ContainerId string
		Request     batch.CreateComputeEnvironmentRequest
	}

	var requests []CreateComputeEnvironmentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.CreateComputeEnvironment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientCreateJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "CreateJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "CreateJob", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "CreateJob")
	assert.NoError(t, err)

	type CreateJobRequestInfo struct {
		ContainerId string
		Request     batch.CreateJobRequest
	}

	var requests []CreateJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.CreateJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientCreateJobDefinition(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "CreateJobDefinition")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateJobDefinition is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "CreateJobDefinition", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "CreateJobDefinition")
	assert.NoError(t, err)

	type CreateJobDefinitionRequestInfo struct {
		ContainerId string
		Request     batch.CreateJobDefinitionRequest
	}

	var requests []CreateJobDefinitionRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.CreateJobDefinition(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientDeleteBatchInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "DeleteBatchInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteBatchInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "DeleteBatchInstance", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "DeleteBatchInstance")
	assert.NoError(t, err)

	type DeleteBatchInstanceRequestInfo struct {
		ContainerId string
		Request     batch.DeleteBatchInstanceRequest
	}

	var requests []DeleteBatchInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DeleteBatchInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientDeleteComputeEnvironment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "DeleteComputeEnvironment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteComputeEnvironment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "DeleteComputeEnvironment", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "DeleteComputeEnvironment")
	assert.NoError(t, err)

	type DeleteComputeEnvironmentRequestInfo struct {
		ContainerId string
		Request     batch.DeleteComputeEnvironmentRequest
	}

	var requests []DeleteComputeEnvironmentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DeleteComputeEnvironment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientDeleteJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "DeleteJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "DeleteJob", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "DeleteJob")
	assert.NoError(t, err)

	type DeleteJobRequestInfo struct {
		ContainerId string
		Request     batch.DeleteJobRequest
	}

	var requests []DeleteJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DeleteJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientDeleteJobDefinition(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "DeleteJobDefinition")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteJobDefinition is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "DeleteJobDefinition", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "DeleteJobDefinition")
	assert.NoError(t, err)

	type DeleteJobDefinitionRequestInfo struct {
		ContainerId string
		Request     batch.DeleteJobDefinitionRequest
	}

	var requests []DeleteJobDefinitionRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DeleteJobDefinition(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientDisableBatchInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "DisableBatchInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DisableBatchInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "DisableBatchInstance", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "DisableBatchInstance")
	assert.NoError(t, err)

	type DisableBatchInstanceRequestInfo struct {
		ContainerId string
		Request     batch.DisableBatchInstanceRequest
	}

	var requests []DisableBatchInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DisableBatchInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientDisableComputeEnvironment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "DisableComputeEnvironment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DisableComputeEnvironment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "DisableComputeEnvironment", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "DisableComputeEnvironment")
	assert.NoError(t, err)

	type DisableComputeEnvironmentRequestInfo struct {
		ContainerId string
		Request     batch.DisableComputeEnvironmentRequest
	}

	var requests []DisableComputeEnvironmentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DisableComputeEnvironment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientEnableBatchInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "EnableBatchInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("EnableBatchInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "EnableBatchInstance", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "EnableBatchInstance")
	assert.NoError(t, err)

	type EnableBatchInstanceRequestInfo struct {
		ContainerId string
		Request     batch.EnableBatchInstanceRequest
	}

	var requests []EnableBatchInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.EnableBatchInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientEnableComputeEnvironment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "EnableComputeEnvironment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("EnableComputeEnvironment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "EnableComputeEnvironment", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "EnableComputeEnvironment")
	assert.NoError(t, err)

	type EnableComputeEnvironmentRequestInfo struct {
		ContainerId string
		Request     batch.EnableComputeEnvironmentRequest
	}

	var requests []EnableComputeEnvironmentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.EnableComputeEnvironment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientGetBatchInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "GetBatchInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetBatchInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "GetBatchInstance", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "GetBatchInstance")
	assert.NoError(t, err)

	type GetBatchInstanceRequestInfo struct {
		ContainerId string
		Request     batch.GetBatchInstanceRequest
	}

	var requests []GetBatchInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetBatchInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientGetComputeEnvironment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "GetComputeEnvironment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetComputeEnvironment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "GetComputeEnvironment", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "GetComputeEnvironment")
	assert.NoError(t, err)

	type GetComputeEnvironmentRequestInfo struct {
		ContainerId string
		Request     batch.GetComputeEnvironmentRequest
	}

	var requests []GetComputeEnvironmentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetComputeEnvironment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientGetJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "GetJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "GetJob", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "GetJob")
	assert.NoError(t, err)

	type GetJobRequestInfo struct {
		ContainerId string
		Request     batch.GetJobRequest
	}

	var requests []GetJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientGetJobDefinition(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "GetJobDefinition")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetJobDefinition is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "GetJobDefinition", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "GetJobDefinition")
	assert.NoError(t, err)

	type GetJobDefinitionRequestInfo struct {
		ContainerId string
		Request     batch.GetJobDefinitionRequest
	}

	var requests []GetJobDefinitionRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetJobDefinition(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientGetJobLog(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "GetJobLog")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetJobLog is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "GetJobLog", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "GetJobLog")
	assert.NoError(t, err)

	type GetJobLogRequestInfo struct {
		ContainerId string
		Request     batch.GetJobLogRequest
	}

	var requests []GetJobLogRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetJobLog(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientGetJobLogContent(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "GetJobLogContent")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetJobLogContent is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "GetJobLogContent", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "GetJobLogContent")
	assert.NoError(t, err)

	type GetJobLogContentRequestInfo struct {
		ContainerId string
		Request     batch.GetJobLogContentRequest
	}

	var requests []GetJobLogContentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetJobLogContent(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientListBatchInstances(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "ListBatchInstances")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListBatchInstances is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "ListBatchInstances", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "ListBatchInstances")
	assert.NoError(t, err)

	type ListBatchInstancesRequestInfo struct {
		ContainerId string
		Request     batch.ListBatchInstancesRequest
	}

	var requests []ListBatchInstancesRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*batch.ListBatchInstancesRequest)
				return c.ListBatchInstances(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]batch.ListBatchInstancesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(batch.ListBatchInstancesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientListComputeEnvironments(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "ListComputeEnvironments")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListComputeEnvironments is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "ListComputeEnvironments", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "ListComputeEnvironments")
	assert.NoError(t, err)

	type ListComputeEnvironmentsRequestInfo struct {
		ContainerId string
		Request     batch.ListComputeEnvironmentsRequest
	}

	var requests []ListComputeEnvironmentsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*batch.ListComputeEnvironmentsRequest)
				return c.ListComputeEnvironments(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]batch.ListComputeEnvironmentsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(batch.ListComputeEnvironmentsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientListJobDefinitions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "ListJobDefinitions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListJobDefinitions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "ListJobDefinitions", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "ListJobDefinitions")
	assert.NoError(t, err)

	type ListJobDefinitionsRequestInfo struct {
		ContainerId string
		Request     batch.ListJobDefinitionsRequest
	}

	var requests []ListJobDefinitionsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*batch.ListJobDefinitionsRequest)
				return c.ListJobDefinitions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]batch.ListJobDefinitionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(batch.ListJobDefinitionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientListJobLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "ListJobLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListJobLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "ListJobLogs", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "ListJobLogs")
	assert.NoError(t, err)

	type ListJobLogsRequestInfo struct {
		ContainerId string
		Request     batch.ListJobLogsRequest
	}

	var requests []ListJobLogsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*batch.ListJobLogsRequest)
				return c.ListJobLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]batch.ListJobLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(batch.ListJobLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientListJobs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "ListJobs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListJobs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "ListJobs", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "ListJobs")
	assert.NoError(t, err)

	type ListJobsRequestInfo struct {
		ContainerId string
		Request     batch.ListJobsRequest
	}

	var requests []ListJobsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*batch.ListJobsRequest)
				return c.ListJobs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]batch.ListJobsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(batch.ListJobsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientUpdateBatchInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "UpdateBatchInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateBatchInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "UpdateBatchInstance", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "UpdateBatchInstance")
	assert.NoError(t, err)

	type UpdateBatchInstanceRequestInfo struct {
		ContainerId string
		Request     batch.UpdateBatchInstanceRequest
	}

	var requests []UpdateBatchInstanceRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.UpdateBatchInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientUpdateComputeEnvironment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "UpdateComputeEnvironment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateComputeEnvironment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "UpdateComputeEnvironment", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "UpdateComputeEnvironment")
	assert.NoError(t, err)

	type UpdateComputeEnvironmentRequestInfo struct {
		ContainerId string
		Request     batch.UpdateComputeEnvironmentRequest
	}

	var requests []UpdateComputeEnvironmentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.UpdateComputeEnvironment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientUpdateJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "UpdateJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "UpdateJob", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "UpdateJob")
	assert.NoError(t, err)

	type UpdateJobRequestInfo struct {
		ContainerId string
		Request     batch.UpdateJobRequest
	}

	var requests []UpdateJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.UpdateJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_batch_service_ww_grp@oracle.com" jiraProject="BCS (Batch Cloud Service)" opsJiraProject="BCS"
func TestBatchServiceClientUpdateJobDefinition(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("batch", "UpdateJobDefinition")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateJobDefinition is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("batch", "BatchService", "UpdateJobDefinition", createBatchServiceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(batch.BatchServiceClient)

	body, err := testClient.getRequests("batch", "UpdateJobDefinition")
	assert.NoError(t, err)

	type UpdateJobDefinitionRequestInfo struct {
		ContainerId string
		Request     batch.UpdateJobDefinitionRequest
	}

	var requests []UpdateJobDefinitionRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.UpdateJobDefinition(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
