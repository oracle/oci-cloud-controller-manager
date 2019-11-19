package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/dts"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createTransferJobClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := dts.NewTransferJobClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferJobClientChangeTransferJobCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "ChangeTransferJobCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeTransferJobCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferJob", "ChangeTransferJobCompartment", createTransferJobClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferJobClient)

	body, err := testClient.getRequests("dts", "ChangeTransferJobCompartment")
	assert.NoError(t, err)

	type ChangeTransferJobCompartmentRequestInfo struct {
		ContainerId string
		Request     dts.ChangeTransferJobCompartmentRequest
	}

	var requests []ChangeTransferJobCompartmentRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.ChangeTransferJobCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferJobClientCreateTransferJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "CreateTransferJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateTransferJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferJob", "CreateTransferJob", createTransferJobClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferJobClient)

	body, err := testClient.getRequests("dts", "CreateTransferJob")
	assert.NoError(t, err)

	type CreateTransferJobRequestInfo struct {
		ContainerId string
		Request     dts.CreateTransferJobRequest
	}

	var requests []CreateTransferJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.CreateTransferJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferJobClientDeleteTransferJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "DeleteTransferJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteTransferJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferJob", "DeleteTransferJob", createTransferJobClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferJobClient)

	body, err := testClient.getRequests("dts", "DeleteTransferJob")
	assert.NoError(t, err)

	type DeleteTransferJobRequestInfo struct {
		ContainerId string
		Request     dts.DeleteTransferJobRequest
	}

	var requests []DeleteTransferJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.DeleteTransferJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferJobClientGetTransferJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "GetTransferJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetTransferJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferJob", "GetTransferJob", createTransferJobClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferJobClient)

	body, err := testClient.getRequests("dts", "GetTransferJob")
	assert.NoError(t, err)

	type GetTransferJobRequestInfo struct {
		ContainerId string
		Request     dts.GetTransferJobRequest
	}

	var requests []GetTransferJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetTransferJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferJobClientListTransferJobs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "ListTransferJobs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListTransferJobs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferJob", "ListTransferJobs", createTransferJobClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferJobClient)

	body, err := testClient.getRequests("dts", "ListTransferJobs")
	assert.NoError(t, err)

	type ListTransferJobsRequestInfo struct {
		ContainerId string
		Request     dts.ListTransferJobsRequest
	}

	var requests []ListTransferJobsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.ListTransferJobs(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferJobClientUpdateTransferJob(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "UpdateTransferJob")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateTransferJob is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferJob", "UpdateTransferJob", createTransferJobClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferJobClient)

	body, err := testClient.getRequests("dts", "UpdateTransferJob")
	assert.NoError(t, err)

	type UpdateTransferJobRequestInfo struct {
		ContainerId string
		Request     dts.UpdateTransferJobRequest
	}

	var requests []UpdateTransferJobRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.UpdateTransferJob(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
