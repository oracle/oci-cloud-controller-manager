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

func createTransferDeviceClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := dts.NewTransferDeviceClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferDeviceClientCreateTransferDevice(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "CreateTransferDevice")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateTransferDevice is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferDevice", "CreateTransferDevice", createTransferDeviceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferDeviceClient)

	body, err := testClient.getRequests("dts", "CreateTransferDevice")
	assert.NoError(t, err)

	type CreateTransferDeviceRequestInfo struct {
		ContainerId string
		Request     dts.CreateTransferDeviceRequest
	}

	var requests []CreateTransferDeviceRequestInfo
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

			response, err := c.CreateTransferDevice(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferDeviceClientDeleteTransferDevice(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "DeleteTransferDevice")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteTransferDevice is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferDevice", "DeleteTransferDevice", createTransferDeviceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferDeviceClient)

	body, err := testClient.getRequests("dts", "DeleteTransferDevice")
	assert.NoError(t, err)

	type DeleteTransferDeviceRequestInfo struct {
		ContainerId string
		Request     dts.DeleteTransferDeviceRequest
	}

	var requests []DeleteTransferDeviceRequestInfo
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

			response, err := c.DeleteTransferDevice(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferDeviceClientGetTransferDevice(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "GetTransferDevice")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetTransferDevice is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferDevice", "GetTransferDevice", createTransferDeviceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferDeviceClient)

	body, err := testClient.getRequests("dts", "GetTransferDevice")
	assert.NoError(t, err)

	type GetTransferDeviceRequestInfo struct {
		ContainerId string
		Request     dts.GetTransferDeviceRequest
	}

	var requests []GetTransferDeviceRequestInfo
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

			response, err := c.GetTransferDevice(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferDeviceClientListTransferDevices(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "ListTransferDevices")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListTransferDevices is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferDevice", "ListTransferDevices", createTransferDeviceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferDeviceClient)

	body, err := testClient.getRequests("dts", "ListTransferDevices")
	assert.NoError(t, err)

	type ListTransferDevicesRequestInfo struct {
		ContainerId string
		Request     dts.ListTransferDevicesRequest
	}

	var requests []ListTransferDevicesRequestInfo
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

			response, err := c.ListTransferDevices(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferDeviceClientUpdateTransferDevice(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "UpdateTransferDevice")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateTransferDevice is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferDevice", "UpdateTransferDevice", createTransferDeviceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferDeviceClient)

	body, err := testClient.getRequests("dts", "UpdateTransferDevice")
	assert.NoError(t, err)

	type UpdateTransferDeviceRequestInfo struct {
		ContainerId string
		Request     dts.UpdateTransferDeviceRequest
	}

	var requests []UpdateTransferDeviceRequestInfo
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

			response, err := c.UpdateTransferDevice(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
