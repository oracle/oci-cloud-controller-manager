package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/storagegateway"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createStorageGatewayClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := storagegateway.NewStorageGatewayClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientCancelCloudSync(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "CancelCloudSync")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CancelCloudSync is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "CancelCloudSync", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "CancelCloudSync")
	assert.NoError(t, err)

	type CancelCloudSyncRequestInfo struct {
		ContainerId string
		Request     storagegateway.CancelCloudSyncRequest
	}

	var requests []CancelCloudSyncRequestInfo
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

			response, err := c.CancelCloudSync(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientChangeStorageGatewayCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "ChangeStorageGatewayCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeStorageGatewayCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "ChangeStorageGatewayCompartment", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "ChangeStorageGatewayCompartment")
	assert.NoError(t, err)

	type ChangeStorageGatewayCompartmentRequestInfo struct {
		ContainerId string
		Request     storagegateway.ChangeStorageGatewayCompartmentRequest
	}

	var requests []ChangeStorageGatewayCompartmentRequestInfo
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

			response, err := c.ChangeStorageGatewayCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientConnectFileSystem(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "ConnectFileSystem")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ConnectFileSystem is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "ConnectFileSystem", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "ConnectFileSystem")
	assert.NoError(t, err)

	type ConnectFileSystemRequestInfo struct {
		ContainerId string
		Request     storagegateway.ConnectFileSystemRequest
	}

	var requests []ConnectFileSystemRequestInfo
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

			response, err := c.ConnectFileSystem(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientCreateCloudSync(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "CreateCloudSync")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateCloudSync is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "CreateCloudSync", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "CreateCloudSync")
	assert.NoError(t, err)

	type CreateCloudSyncRequestInfo struct {
		ContainerId string
		Request     storagegateway.CreateCloudSyncRequest
	}

	var requests []CreateCloudSyncRequestInfo
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

			response, err := c.CreateCloudSync(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientCreateFileSystem(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "CreateFileSystem")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateFileSystem is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "CreateFileSystem", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "CreateFileSystem")
	assert.NoError(t, err)

	type CreateFileSystemRequestInfo struct {
		ContainerId string
		Request     storagegateway.CreateFileSystemRequest
	}

	var requests []CreateFileSystemRequestInfo
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

			response, err := c.CreateFileSystem(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientCreateStorageGateway(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "CreateStorageGateway")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateStorageGateway is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "CreateStorageGateway", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "CreateStorageGateway")
	assert.NoError(t, err)

	type CreateStorageGatewayRequestInfo struct {
		ContainerId string
		Request     storagegateway.CreateStorageGatewayRequest
	}

	var requests []CreateStorageGatewayRequestInfo
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

			response, err := c.CreateStorageGateway(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientDeleteCloudSync(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "DeleteCloudSync")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteCloudSync is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "DeleteCloudSync", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "DeleteCloudSync")
	assert.NoError(t, err)

	type DeleteCloudSyncRequestInfo struct {
		ContainerId string
		Request     storagegateway.DeleteCloudSyncRequest
	}

	var requests []DeleteCloudSyncRequestInfo
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

			response, err := c.DeleteCloudSync(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientDeleteFileSystem(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "DeleteFileSystem")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteFileSystem is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "DeleteFileSystem", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "DeleteFileSystem")
	assert.NoError(t, err)

	type DeleteFileSystemRequestInfo struct {
		ContainerId string
		Request     storagegateway.DeleteFileSystemRequest
	}

	var requests []DeleteFileSystemRequestInfo
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

			response, err := c.DeleteFileSystem(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientDeleteStorageGateway(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "DeleteStorageGateway")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteStorageGateway is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "DeleteStorageGateway", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "DeleteStorageGateway")
	assert.NoError(t, err)

	type DeleteStorageGatewayRequestInfo struct {
		ContainerId string
		Request     storagegateway.DeleteStorageGatewayRequest
	}

	var requests []DeleteStorageGatewayRequestInfo
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

			response, err := c.DeleteStorageGateway(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientDisconnectFileSystem(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "DisconnectFileSystem")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DisconnectFileSystem is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "DisconnectFileSystem", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "DisconnectFileSystem")
	assert.NoError(t, err)

	type DisconnectFileSystemRequestInfo struct {
		ContainerId string
		Request     storagegateway.DisconnectFileSystemRequest
	}

	var requests []DisconnectFileSystemRequestInfo
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

			response, err := c.DisconnectFileSystem(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientGetCloudSync(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "GetCloudSync")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetCloudSync is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "GetCloudSync", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "GetCloudSync")
	assert.NoError(t, err)

	type GetCloudSyncRequestInfo struct {
		ContainerId string
		Request     storagegateway.GetCloudSyncRequest
	}

	var requests []GetCloudSyncRequestInfo
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

			response, err := c.GetCloudSync(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientGetCloudSyncHealth(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "GetCloudSyncHealth")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetCloudSyncHealth is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "GetCloudSyncHealth", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "GetCloudSyncHealth")
	assert.NoError(t, err)

	type GetCloudSyncHealthRequestInfo struct {
		ContainerId string
		Request     storagegateway.GetCloudSyncHealthRequest
	}

	var requests []GetCloudSyncHealthRequestInfo
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

			response, err := c.GetCloudSyncHealth(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientGetFileSystem(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "GetFileSystem")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetFileSystem is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "GetFileSystem", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "GetFileSystem")
	assert.NoError(t, err)

	type GetFileSystemRequestInfo struct {
		ContainerId string
		Request     storagegateway.GetFileSystemRequest
	}

	var requests []GetFileSystemRequestInfo
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

			response, err := c.GetFileSystem(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientGetFileSystemHealth(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "GetFileSystemHealth")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetFileSystemHealth is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "GetFileSystemHealth", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "GetFileSystemHealth")
	assert.NoError(t, err)

	type GetFileSystemHealthRequestInfo struct {
		ContainerId string
		Request     storagegateway.GetFileSystemHealthRequest
	}

	var requests []GetFileSystemHealthRequestInfo
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

			response, err := c.GetFileSystemHealth(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientGetStorageGateway(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "GetStorageGateway")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetStorageGateway is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "GetStorageGateway", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "GetStorageGateway")
	assert.NoError(t, err)

	type GetStorageGatewayRequestInfo struct {
		ContainerId string
		Request     storagegateway.GetStorageGatewayRequest
	}

	var requests []GetStorageGatewayRequestInfo
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

			response, err := c.GetStorageGateway(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientGetStorageGatewayHealth(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "GetStorageGatewayHealth")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetStorageGatewayHealth is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "GetStorageGatewayHealth", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "GetStorageGatewayHealth")
	assert.NoError(t, err)

	type GetStorageGatewayHealthRequestInfo struct {
		ContainerId string
		Request     storagegateway.GetStorageGatewayHealthRequest
	}

	var requests []GetStorageGatewayHealthRequestInfo
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

			response, err := c.GetStorageGatewayHealth(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientListCloudSyncs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "ListCloudSyncs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListCloudSyncs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "ListCloudSyncs", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "ListCloudSyncs")
	assert.NoError(t, err)

	type ListCloudSyncsRequestInfo struct {
		ContainerId string
		Request     storagegateway.ListCloudSyncsRequest
	}

	var requests []ListCloudSyncsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*storagegateway.ListCloudSyncsRequest)
				return c.ListCloudSyncs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]storagegateway.ListCloudSyncsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(storagegateway.ListCloudSyncsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientListFileSystems(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "ListFileSystems")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListFileSystems is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "ListFileSystems", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "ListFileSystems")
	assert.NoError(t, err)

	type ListFileSystemsRequestInfo struct {
		ContainerId string
		Request     storagegateway.ListFileSystemsRequest
	}

	var requests []ListFileSystemsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*storagegateway.ListFileSystemsRequest)
				return c.ListFileSystems(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]storagegateway.ListFileSystemsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(storagegateway.ListFileSystemsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientListStorageGateways(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "ListStorageGateways")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListStorageGateways is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "ListStorageGateways", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "ListStorageGateways")
	assert.NoError(t, err)

	type ListStorageGatewaysRequestInfo struct {
		ContainerId string
		Request     storagegateway.ListStorageGatewaysRequest
	}

	var requests []ListStorageGatewaysRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*storagegateway.ListStorageGatewaysRequest)
				return c.ListStorageGateways(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]storagegateway.ListStorageGatewaysResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(storagegateway.ListStorageGatewaysResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientReclaimFileSystem(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "ReclaimFileSystem")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ReclaimFileSystem is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "ReclaimFileSystem", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "ReclaimFileSystem")
	assert.NoError(t, err)

	type ReclaimFileSystemRequestInfo struct {
		ContainerId string
		Request     storagegateway.ReclaimFileSystemRequest
	}

	var requests []ReclaimFileSystemRequestInfo
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

			response, err := c.ReclaimFileSystem(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientRefreshFileSystem(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "RefreshFileSystem")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("RefreshFileSystem is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "RefreshFileSystem", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "RefreshFileSystem")
	assert.NoError(t, err)

	type RefreshFileSystemRequestInfo struct {
		ContainerId string
		Request     storagegateway.RefreshFileSystemRequest
	}

	var requests []RefreshFileSystemRequestInfo
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

			response, err := c.RefreshFileSystem(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientRunCloudSync(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "RunCloudSync")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("RunCloudSync is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "RunCloudSync", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "RunCloudSync")
	assert.NoError(t, err)

	type RunCloudSyncRequestInfo struct {
		ContainerId string
		Request     storagegateway.RunCloudSyncRequest
	}

	var requests []RunCloudSyncRequestInfo
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

			response, err := c.RunCloudSync(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientUpdateCloudSync(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "UpdateCloudSync")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateCloudSync is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "UpdateCloudSync", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "UpdateCloudSync")
	assert.NoError(t, err)

	type UpdateCloudSyncRequestInfo struct {
		ContainerId string
		Request     storagegateway.UpdateCloudSyncRequest
	}

	var requests []UpdateCloudSyncRequestInfo
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

			response, err := c.UpdateCloudSync(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientUpdateFileSystem(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "UpdateFileSystem")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateFileSystem is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "UpdateFileSystem", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "UpdateFileSystem")
	assert.NoError(t, err)

	type UpdateFileSystemRequestInfo struct {
		ContainerId string
		Request     storagegateway.UpdateFileSystemRequest
	}

	var requests []UpdateFileSystemRequestInfo
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

			response, err := c.UpdateFileSystem(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="csgdev_us_grp@oracle.com" jiraProject="CSG" opsJiraProject="SG"
func TestStorageGatewayClientUpdateStorageGateway(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("storagegateway", "UpdateStorageGateway")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateStorageGateway is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("storagegateway", "StorageGateway", "UpdateStorageGateway", createStorageGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(storagegateway.StorageGatewayClient)

	body, err := testClient.getRequests("storagegateway", "UpdateStorageGateway")
	assert.NoError(t, err)

	type UpdateStorageGatewayRequestInfo struct {
		ContainerId string
		Request     storagegateway.UpdateStorageGatewayRequest
	}

	var requests []UpdateStorageGatewayRequestInfo
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

			response, err := c.UpdateStorageGateway(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
