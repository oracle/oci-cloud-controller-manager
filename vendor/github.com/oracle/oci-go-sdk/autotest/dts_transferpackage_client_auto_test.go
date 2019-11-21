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

func createTransferPackageClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := dts.NewTransferPackageClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferPackageClientAttachDevicesToTransferPackage(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "AttachDevicesToTransferPackage")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("AttachDevicesToTransferPackage is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferPackage", "AttachDevicesToTransferPackage", createTransferPackageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferPackageClient)

	body, err := testClient.getRequests("dts", "AttachDevicesToTransferPackage")
	assert.NoError(t, err)

	type AttachDevicesToTransferPackageRequestInfo struct {
		ContainerId string
		Request     dts.AttachDevicesToTransferPackageRequest
	}

	var requests []AttachDevicesToTransferPackageRequestInfo
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

			response, err := c.AttachDevicesToTransferPackage(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferPackageClientCreateTransferPackage(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "CreateTransferPackage")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateTransferPackage is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferPackage", "CreateTransferPackage", createTransferPackageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferPackageClient)

	body, err := testClient.getRequests("dts", "CreateTransferPackage")
	assert.NoError(t, err)

	type CreateTransferPackageRequestInfo struct {
		ContainerId string
		Request     dts.CreateTransferPackageRequest
	}

	var requests []CreateTransferPackageRequestInfo
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

			response, err := c.CreateTransferPackage(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferPackageClientDeleteTransferPackage(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "DeleteTransferPackage")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteTransferPackage is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferPackage", "DeleteTransferPackage", createTransferPackageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferPackageClient)

	body, err := testClient.getRequests("dts", "DeleteTransferPackage")
	assert.NoError(t, err)

	type DeleteTransferPackageRequestInfo struct {
		ContainerId string
		Request     dts.DeleteTransferPackageRequest
	}

	var requests []DeleteTransferPackageRequestInfo
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

			response, err := c.DeleteTransferPackage(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferPackageClientDetachDevicesFromTransferPackage(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "DetachDevicesFromTransferPackage")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DetachDevicesFromTransferPackage is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferPackage", "DetachDevicesFromTransferPackage", createTransferPackageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferPackageClient)

	body, err := testClient.getRequests("dts", "DetachDevicesFromTransferPackage")
	assert.NoError(t, err)

	type DetachDevicesFromTransferPackageRequestInfo struct {
		ContainerId string
		Request     dts.DetachDevicesFromTransferPackageRequest
	}

	var requests []DetachDevicesFromTransferPackageRequestInfo
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

			response, err := c.DetachDevicesFromTransferPackage(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferPackageClientGetTransferPackage(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "GetTransferPackage")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetTransferPackage is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferPackage", "GetTransferPackage", createTransferPackageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferPackageClient)

	body, err := testClient.getRequests("dts", "GetTransferPackage")
	assert.NoError(t, err)

	type GetTransferPackageRequestInfo struct {
		ContainerId string
		Request     dts.GetTransferPackageRequest
	}

	var requests []GetTransferPackageRequestInfo
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

			response, err := c.GetTransferPackage(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferPackageClientListTransferPackages(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "ListTransferPackages")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListTransferPackages is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferPackage", "ListTransferPackages", createTransferPackageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferPackageClient)

	body, err := testClient.getRequests("dts", "ListTransferPackages")
	assert.NoError(t, err)

	type ListTransferPackagesRequestInfo struct {
		ContainerId string
		Request     dts.ListTransferPackagesRequest
	}

	var requests []ListTransferPackagesRequestInfo
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

			response, err := c.ListTransferPackages(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferPackageClientUpdateTransferPackage(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "UpdateTransferPackage")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateTransferPackage is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferPackage", "UpdateTransferPackage", createTransferPackageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferPackageClient)

	body, err := testClient.getRequests("dts", "UpdateTransferPackage")
	assert.NoError(t, err)

	type UpdateTransferPackageRequestInfo struct {
		ContainerId string
		Request     dts.UpdateTransferPackageRequest
	}

	var requests []UpdateTransferPackageRequestInfo
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

			response, err := c.UpdateTransferPackage(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
