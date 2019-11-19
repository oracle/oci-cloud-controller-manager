package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createBlockstorageClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := core.NewBlockstorageClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientChangeBootVolumeBackupCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ChangeBootVolumeBackupCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeBootVolumeBackupCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ChangeBootVolumeBackupCompartment", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ChangeBootVolumeBackupCompartment")
	assert.NoError(t, err)

	type ChangeBootVolumeBackupCompartmentRequestInfo struct {
		ContainerId string
		Request     core.ChangeBootVolumeBackupCompartmentRequest
	}

	var requests []ChangeBootVolumeBackupCompartmentRequestInfo
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

			response, err := c.ChangeBootVolumeBackupCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientChangeBootVolumeCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ChangeBootVolumeCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeBootVolumeCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ChangeBootVolumeCompartment", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ChangeBootVolumeCompartment")
	assert.NoError(t, err)

	type ChangeBootVolumeCompartmentRequestInfo struct {
		ContainerId string
		Request     core.ChangeBootVolumeCompartmentRequest
	}

	var requests []ChangeBootVolumeCompartmentRequestInfo
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

			response, err := c.ChangeBootVolumeCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientChangeVolumeBackupCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ChangeVolumeBackupCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeVolumeBackupCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ChangeVolumeBackupCompartment", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ChangeVolumeBackupCompartment")
	assert.NoError(t, err)

	type ChangeVolumeBackupCompartmentRequestInfo struct {
		ContainerId string
		Request     core.ChangeVolumeBackupCompartmentRequest
	}

	var requests []ChangeVolumeBackupCompartmentRequestInfo
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

			response, err := c.ChangeVolumeBackupCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientChangeVolumeCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ChangeVolumeCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeVolumeCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ChangeVolumeCompartment", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ChangeVolumeCompartment")
	assert.NoError(t, err)

	type ChangeVolumeCompartmentRequestInfo struct {
		ContainerId string
		Request     core.ChangeVolumeCompartmentRequest
	}

	var requests []ChangeVolumeCompartmentRequestInfo
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

			response, err := c.ChangeVolumeCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientChangeVolumeGroupBackupCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ChangeVolumeGroupBackupCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeVolumeGroupBackupCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ChangeVolumeGroupBackupCompartment", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ChangeVolumeGroupBackupCompartment")
	assert.NoError(t, err)

	type ChangeVolumeGroupBackupCompartmentRequestInfo struct {
		ContainerId string
		Request     core.ChangeVolumeGroupBackupCompartmentRequest
	}

	var requests []ChangeVolumeGroupBackupCompartmentRequestInfo
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

			response, err := c.ChangeVolumeGroupBackupCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientChangeVolumeGroupCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ChangeVolumeGroupCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeVolumeGroupCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ChangeVolumeGroupCompartment", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ChangeVolumeGroupCompartment")
	assert.NoError(t, err)

	type ChangeVolumeGroupCompartmentRequestInfo struct {
		ContainerId string
		Request     core.ChangeVolumeGroupCompartmentRequest
	}

	var requests []ChangeVolumeGroupCompartmentRequestInfo
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

			response, err := c.ChangeVolumeGroupCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientCopyBootVolumeBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CopyBootVolumeBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CopyBootVolumeBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "CopyBootVolumeBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "CopyBootVolumeBackup")
	assert.NoError(t, err)

	type CopyBootVolumeBackupRequestInfo struct {
		ContainerId string
		Request     core.CopyBootVolumeBackupRequest
	}

	var requests []CopyBootVolumeBackupRequestInfo
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

			response, err := c.CopyBootVolumeBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientCopyVolumeBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CopyVolumeBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CopyVolumeBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "CopyVolumeBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "CopyVolumeBackup")
	assert.NoError(t, err)

	type CopyVolumeBackupRequestInfo struct {
		ContainerId string
		Request     core.CopyVolumeBackupRequest
	}

	var requests []CopyVolumeBackupRequestInfo
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

			response, err := c.CopyVolumeBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientCreateBootVolume(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateBootVolume")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateBootVolume is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "CreateBootVolume", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "CreateBootVolume")
	assert.NoError(t, err)

	type CreateBootVolumeRequestInfo struct {
		ContainerId string
		Request     core.CreateBootVolumeRequest
	}

	var requests []CreateBootVolumeRequestInfo
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

			response, err := c.CreateBootVolume(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientCreateBootVolumeBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateBootVolumeBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateBootVolumeBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "CreateBootVolumeBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "CreateBootVolumeBackup")
	assert.NoError(t, err)

	type CreateBootVolumeBackupRequestInfo struct {
		ContainerId string
		Request     core.CreateBootVolumeBackupRequest
	}

	var requests []CreateBootVolumeBackupRequestInfo
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

			response, err := c.CreateBootVolumeBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientCreateVolume(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateVolume")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateVolume is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "CreateVolume", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "CreateVolume")
	assert.NoError(t, err)

	type CreateVolumeRequestInfo struct {
		ContainerId string
		Request     core.CreateVolumeRequest
	}

	var requests []CreateVolumeRequestInfo
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

			response, err := c.CreateVolume(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientCreateVolumeBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateVolumeBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateVolumeBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "CreateVolumeBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "CreateVolumeBackup")
	assert.NoError(t, err)

	type CreateVolumeBackupRequestInfo struct {
		ContainerId string
		Request     core.CreateVolumeBackupRequest
	}

	var requests []CreateVolumeBackupRequestInfo
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

			response, err := c.CreateVolumeBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientCreateVolumeBackupPolicy(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateVolumeBackupPolicy")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateVolumeBackupPolicy is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "CreateVolumeBackupPolicy", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "CreateVolumeBackupPolicy")
	assert.NoError(t, err)

	type CreateVolumeBackupPolicyRequestInfo struct {
		ContainerId string
		Request     core.CreateVolumeBackupPolicyRequest
	}

	var requests []CreateVolumeBackupPolicyRequestInfo
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

			response, err := c.CreateVolumeBackupPolicy(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientCreateVolumeBackupPolicyAssignment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateVolumeBackupPolicyAssignment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateVolumeBackupPolicyAssignment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "CreateVolumeBackupPolicyAssignment", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "CreateVolumeBackupPolicyAssignment")
	assert.NoError(t, err)

	type CreateVolumeBackupPolicyAssignmentRequestInfo struct {
		ContainerId string
		Request     core.CreateVolumeBackupPolicyAssignmentRequest
	}

	var requests []CreateVolumeBackupPolicyAssignmentRequestInfo
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

			response, err := c.CreateVolumeBackupPolicyAssignment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientCreateVolumeGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateVolumeGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateVolumeGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "CreateVolumeGroup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "CreateVolumeGroup")
	assert.NoError(t, err)

	type CreateVolumeGroupRequestInfo struct {
		ContainerId string
		Request     core.CreateVolumeGroupRequest
	}

	var requests []CreateVolumeGroupRequestInfo
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

			response, err := c.CreateVolumeGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientCreateVolumeGroupBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "CreateVolumeGroupBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateVolumeGroupBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "CreateVolumeGroupBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "CreateVolumeGroupBackup")
	assert.NoError(t, err)

	type CreateVolumeGroupBackupRequestInfo struct {
		ContainerId string
		Request     core.CreateVolumeGroupBackupRequest
	}

	var requests []CreateVolumeGroupBackupRequestInfo
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

			response, err := c.CreateVolumeGroupBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientDeleteBootVolume(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteBootVolume")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteBootVolume is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "DeleteBootVolume", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "DeleteBootVolume")
	assert.NoError(t, err)

	type DeleteBootVolumeRequestInfo struct {
		ContainerId string
		Request     core.DeleteBootVolumeRequest
	}

	var requests []DeleteBootVolumeRequestInfo
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

			response, err := c.DeleteBootVolume(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientDeleteBootVolumeBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteBootVolumeBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteBootVolumeBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "DeleteBootVolumeBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "DeleteBootVolumeBackup")
	assert.NoError(t, err)

	type DeleteBootVolumeBackupRequestInfo struct {
		ContainerId string
		Request     core.DeleteBootVolumeBackupRequest
	}

	var requests []DeleteBootVolumeBackupRequestInfo
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

			response, err := c.DeleteBootVolumeBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientDeleteBootVolumeKmsKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteBootVolumeKmsKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteBootVolumeKmsKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "DeleteBootVolumeKmsKey", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "DeleteBootVolumeKmsKey")
	assert.NoError(t, err)

	type DeleteBootVolumeKmsKeyRequestInfo struct {
		ContainerId string
		Request     core.DeleteBootVolumeKmsKeyRequest
	}

	var requests []DeleteBootVolumeKmsKeyRequestInfo
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

			response, err := c.DeleteBootVolumeKmsKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientDeleteVolume(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteVolume")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteVolume is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "DeleteVolume", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "DeleteVolume")
	assert.NoError(t, err)

	type DeleteVolumeRequestInfo struct {
		ContainerId string
		Request     core.DeleteVolumeRequest
	}

	var requests []DeleteVolumeRequestInfo
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

			response, err := c.DeleteVolume(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientDeleteVolumeBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteVolumeBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteVolumeBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "DeleteVolumeBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "DeleteVolumeBackup")
	assert.NoError(t, err)

	type DeleteVolumeBackupRequestInfo struct {
		ContainerId string
		Request     core.DeleteVolumeBackupRequest
	}

	var requests []DeleteVolumeBackupRequestInfo
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

			response, err := c.DeleteVolumeBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientDeleteVolumeBackupPolicy(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteVolumeBackupPolicy")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteVolumeBackupPolicy is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "DeleteVolumeBackupPolicy", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "DeleteVolumeBackupPolicy")
	assert.NoError(t, err)

	type DeleteVolumeBackupPolicyRequestInfo struct {
		ContainerId string
		Request     core.DeleteVolumeBackupPolicyRequest
	}

	var requests []DeleteVolumeBackupPolicyRequestInfo
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

			response, err := c.DeleteVolumeBackupPolicy(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientDeleteVolumeBackupPolicyAssignment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteVolumeBackupPolicyAssignment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteVolumeBackupPolicyAssignment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "DeleteVolumeBackupPolicyAssignment", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "DeleteVolumeBackupPolicyAssignment")
	assert.NoError(t, err)

	type DeleteVolumeBackupPolicyAssignmentRequestInfo struct {
		ContainerId string
		Request     core.DeleteVolumeBackupPolicyAssignmentRequest
	}

	var requests []DeleteVolumeBackupPolicyAssignmentRequestInfo
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

			response, err := c.DeleteVolumeBackupPolicyAssignment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientDeleteVolumeGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteVolumeGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteVolumeGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "DeleteVolumeGroup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "DeleteVolumeGroup")
	assert.NoError(t, err)

	type DeleteVolumeGroupRequestInfo struct {
		ContainerId string
		Request     core.DeleteVolumeGroupRequest
	}

	var requests []DeleteVolumeGroupRequestInfo
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

			response, err := c.DeleteVolumeGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientDeleteVolumeGroupBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteVolumeGroupBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteVolumeGroupBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "DeleteVolumeGroupBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "DeleteVolumeGroupBackup")
	assert.NoError(t, err)

	type DeleteVolumeGroupBackupRequestInfo struct {
		ContainerId string
		Request     core.DeleteVolumeGroupBackupRequest
	}

	var requests []DeleteVolumeGroupBackupRequestInfo
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

			response, err := c.DeleteVolumeGroupBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientDeleteVolumeKmsKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "DeleteVolumeKmsKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteVolumeKmsKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "DeleteVolumeKmsKey", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "DeleteVolumeKmsKey")
	assert.NoError(t, err)

	type DeleteVolumeKmsKeyRequestInfo struct {
		ContainerId string
		Request     core.DeleteVolumeKmsKeyRequest
	}

	var requests []DeleteVolumeKmsKeyRequestInfo
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

			response, err := c.DeleteVolumeKmsKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetBootVolume(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetBootVolume")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetBootVolume is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetBootVolume", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetBootVolume")
	assert.NoError(t, err)

	type GetBootVolumeRequestInfo struct {
		ContainerId string
		Request     core.GetBootVolumeRequest
	}

	var requests []GetBootVolumeRequestInfo
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

			response, err := c.GetBootVolume(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetBootVolumeBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetBootVolumeBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetBootVolumeBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetBootVolumeBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetBootVolumeBackup")
	assert.NoError(t, err)

	type GetBootVolumeBackupRequestInfo struct {
		ContainerId string
		Request     core.GetBootVolumeBackupRequest
	}

	var requests []GetBootVolumeBackupRequestInfo
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

			response, err := c.GetBootVolumeBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetBootVolumeKmsKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetBootVolumeKmsKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetBootVolumeKmsKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetBootVolumeKmsKey", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetBootVolumeKmsKey")
	assert.NoError(t, err)

	type GetBootVolumeKmsKeyRequestInfo struct {
		ContainerId string
		Request     core.GetBootVolumeKmsKeyRequest
	}

	var requests []GetBootVolumeKmsKeyRequestInfo
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

			response, err := c.GetBootVolumeKmsKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetVolume(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetVolume")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetVolume is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetVolume", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetVolume")
	assert.NoError(t, err)

	type GetVolumeRequestInfo struct {
		ContainerId string
		Request     core.GetVolumeRequest
	}

	var requests []GetVolumeRequestInfo
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

			response, err := c.GetVolume(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetVolumeBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetVolumeBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetVolumeBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetVolumeBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetVolumeBackup")
	assert.NoError(t, err)

	type GetVolumeBackupRequestInfo struct {
		ContainerId string
		Request     core.GetVolumeBackupRequest
	}

	var requests []GetVolumeBackupRequestInfo
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

			response, err := c.GetVolumeBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetVolumeBackupPolicy(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetVolumeBackupPolicy")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetVolumeBackupPolicy is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetVolumeBackupPolicy", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetVolumeBackupPolicy")
	assert.NoError(t, err)

	type GetVolumeBackupPolicyRequestInfo struct {
		ContainerId string
		Request     core.GetVolumeBackupPolicyRequest
	}

	var requests []GetVolumeBackupPolicyRequestInfo
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

			response, err := c.GetVolumeBackupPolicy(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetVolumeBackupPolicyAssetAssignment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetVolumeBackupPolicyAssetAssignment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetVolumeBackupPolicyAssetAssignment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetVolumeBackupPolicyAssetAssignment", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetVolumeBackupPolicyAssetAssignment")
	assert.NoError(t, err)

	type GetVolumeBackupPolicyAssetAssignmentRequestInfo struct {
		ContainerId string
		Request     core.GetVolumeBackupPolicyAssetAssignmentRequest
	}

	var requests []GetVolumeBackupPolicyAssetAssignmentRequestInfo
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
				r := req.(*core.GetVolumeBackupPolicyAssetAssignmentRequest)
				return c.GetVolumeBackupPolicyAssetAssignment(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.GetVolumeBackupPolicyAssetAssignmentResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.GetVolumeBackupPolicyAssetAssignmentResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetVolumeBackupPolicyAssignment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetVolumeBackupPolicyAssignment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetVolumeBackupPolicyAssignment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetVolumeBackupPolicyAssignment", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetVolumeBackupPolicyAssignment")
	assert.NoError(t, err)

	type GetVolumeBackupPolicyAssignmentRequestInfo struct {
		ContainerId string
		Request     core.GetVolumeBackupPolicyAssignmentRequest
	}

	var requests []GetVolumeBackupPolicyAssignmentRequestInfo
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

			response, err := c.GetVolumeBackupPolicyAssignment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetVolumeGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetVolumeGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetVolumeGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetVolumeGroup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetVolumeGroup")
	assert.NoError(t, err)

	type GetVolumeGroupRequestInfo struct {
		ContainerId string
		Request     core.GetVolumeGroupRequest
	}

	var requests []GetVolumeGroupRequestInfo
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

			response, err := c.GetVolumeGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetVolumeGroupBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetVolumeGroupBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetVolumeGroupBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetVolumeGroupBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetVolumeGroupBackup")
	assert.NoError(t, err)

	type GetVolumeGroupBackupRequestInfo struct {
		ContainerId string
		Request     core.GetVolumeGroupBackupRequest
	}

	var requests []GetVolumeGroupBackupRequestInfo
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

			response, err := c.GetVolumeGroupBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientGetVolumeKmsKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "GetVolumeKmsKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetVolumeKmsKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "GetVolumeKmsKey", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "GetVolumeKmsKey")
	assert.NoError(t, err)

	type GetVolumeKmsKeyRequestInfo struct {
		ContainerId string
		Request     core.GetVolumeKmsKeyRequest
	}

	var requests []GetVolumeKmsKeyRequestInfo
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

			response, err := c.GetVolumeKmsKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientListBootVolumeBackups(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListBootVolumeBackups")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListBootVolumeBackups is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ListBootVolumeBackups", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ListBootVolumeBackups")
	assert.NoError(t, err)

	type ListBootVolumeBackupsRequestInfo struct {
		ContainerId string
		Request     core.ListBootVolumeBackupsRequest
	}

	var requests []ListBootVolumeBackupsRequestInfo
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
				r := req.(*core.ListBootVolumeBackupsRequest)
				return c.ListBootVolumeBackups(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListBootVolumeBackupsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListBootVolumeBackupsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientListBootVolumes(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListBootVolumes")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListBootVolumes is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ListBootVolumes", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ListBootVolumes")
	assert.NoError(t, err)

	type ListBootVolumesRequestInfo struct {
		ContainerId string
		Request     core.ListBootVolumesRequest
	}

	var requests []ListBootVolumesRequestInfo
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
				r := req.(*core.ListBootVolumesRequest)
				return c.ListBootVolumes(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListBootVolumesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListBootVolumesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientListVolumeBackupPolicies(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListVolumeBackupPolicies")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListVolumeBackupPolicies is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ListVolumeBackupPolicies", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ListVolumeBackupPolicies")
	assert.NoError(t, err)

	type ListVolumeBackupPoliciesRequestInfo struct {
		ContainerId string
		Request     core.ListVolumeBackupPoliciesRequest
	}

	var requests []ListVolumeBackupPoliciesRequestInfo
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
				r := req.(*core.ListVolumeBackupPoliciesRequest)
				return c.ListVolumeBackupPolicies(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListVolumeBackupPoliciesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListVolumeBackupPoliciesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientListVolumeBackups(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListVolumeBackups")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListVolumeBackups is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ListVolumeBackups", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ListVolumeBackups")
	assert.NoError(t, err)

	type ListVolumeBackupsRequestInfo struct {
		ContainerId string
		Request     core.ListVolumeBackupsRequest
	}

	var requests []ListVolumeBackupsRequestInfo
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
				r := req.(*core.ListVolumeBackupsRequest)
				return c.ListVolumeBackups(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListVolumeBackupsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListVolumeBackupsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientListVolumeGroupBackups(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListVolumeGroupBackups")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListVolumeGroupBackups is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ListVolumeGroupBackups", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ListVolumeGroupBackups")
	assert.NoError(t, err)

	type ListVolumeGroupBackupsRequestInfo struct {
		ContainerId string
		Request     core.ListVolumeGroupBackupsRequest
	}

	var requests []ListVolumeGroupBackupsRequestInfo
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
				r := req.(*core.ListVolumeGroupBackupsRequest)
				return c.ListVolumeGroupBackups(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListVolumeGroupBackupsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListVolumeGroupBackupsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientListVolumeGroups(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListVolumeGroups")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListVolumeGroups is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ListVolumeGroups", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ListVolumeGroups")
	assert.NoError(t, err)

	type ListVolumeGroupsRequestInfo struct {
		ContainerId string
		Request     core.ListVolumeGroupsRequest
	}

	var requests []ListVolumeGroupsRequestInfo
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
				r := req.(*core.ListVolumeGroupsRequest)
				return c.ListVolumeGroups(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListVolumeGroupsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListVolumeGroupsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientListVolumes(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "ListVolumes")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListVolumes is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "ListVolumes", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "ListVolumes")
	assert.NoError(t, err)

	type ListVolumesRequestInfo struct {
		ContainerId string
		Request     core.ListVolumesRequest
	}

	var requests []ListVolumesRequestInfo
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
				r := req.(*core.ListVolumesRequest)
				return c.ListVolumes(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]core.ListVolumesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(core.ListVolumesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientUpdateBootVolume(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateBootVolume")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateBootVolume is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "UpdateBootVolume", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "UpdateBootVolume")
	assert.NoError(t, err)

	type UpdateBootVolumeRequestInfo struct {
		ContainerId string
		Request     core.UpdateBootVolumeRequest
	}

	var requests []UpdateBootVolumeRequestInfo
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

			response, err := c.UpdateBootVolume(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientUpdateBootVolumeBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateBootVolumeBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateBootVolumeBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "UpdateBootVolumeBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "UpdateBootVolumeBackup")
	assert.NoError(t, err)

	type UpdateBootVolumeBackupRequestInfo struct {
		ContainerId string
		Request     core.UpdateBootVolumeBackupRequest
	}

	var requests []UpdateBootVolumeBackupRequestInfo
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

			response, err := c.UpdateBootVolumeBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientUpdateBootVolumeKmsKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateBootVolumeKmsKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateBootVolumeKmsKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "UpdateBootVolumeKmsKey", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "UpdateBootVolumeKmsKey")
	assert.NoError(t, err)

	type UpdateBootVolumeKmsKeyRequestInfo struct {
		ContainerId string
		Request     core.UpdateBootVolumeKmsKeyRequest
	}

	var requests []UpdateBootVolumeKmsKeyRequestInfo
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

			response, err := c.UpdateBootVolumeKmsKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientUpdateVolume(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateVolume")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateVolume is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "UpdateVolume", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "UpdateVolume")
	assert.NoError(t, err)

	type UpdateVolumeRequestInfo struct {
		ContainerId string
		Request     core.UpdateVolumeRequest
	}

	var requests []UpdateVolumeRequestInfo
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

			response, err := c.UpdateVolume(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientUpdateVolumeBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateVolumeBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateVolumeBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "UpdateVolumeBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "UpdateVolumeBackup")
	assert.NoError(t, err)

	type UpdateVolumeBackupRequestInfo struct {
		ContainerId string
		Request     core.UpdateVolumeBackupRequest
	}

	var requests []UpdateVolumeBackupRequestInfo
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

			response, err := c.UpdateVolumeBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientUpdateVolumeBackupPolicy(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateVolumeBackupPolicy")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateVolumeBackupPolicy is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "UpdateVolumeBackupPolicy", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "UpdateVolumeBackupPolicy")
	assert.NoError(t, err)

	type UpdateVolumeBackupPolicyRequestInfo struct {
		ContainerId string
		Request     core.UpdateVolumeBackupPolicyRequest
	}

	var requests []UpdateVolumeBackupPolicyRequestInfo
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

			response, err := c.UpdateVolumeBackupPolicy(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientUpdateVolumeGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateVolumeGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateVolumeGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "UpdateVolumeGroup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "UpdateVolumeGroup")
	assert.NoError(t, err)

	type UpdateVolumeGroupRequestInfo struct {
		ContainerId string
		Request     core.UpdateVolumeGroupRequest
	}

	var requests []UpdateVolumeGroupRequestInfo
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

			response, err := c.UpdateVolumeGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientUpdateVolumeGroupBackup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateVolumeGroupBackup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateVolumeGroupBackup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "UpdateVolumeGroupBackup", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "UpdateVolumeGroupBackup")
	assert.NoError(t, err)

	type UpdateVolumeGroupBackupRequestInfo struct {
		ContainerId string
		Request     core.UpdateVolumeGroupBackupRequest
	}

	var requests []UpdateVolumeGroupBackupRequestInfo
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

			response, err := c.UpdateVolumeGroupBackup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="blockStorage" email="sic_block_storage_cp_us_grp@oracle.com" jiraProject="BLOCK" opsJiraProject="BSCP"
func TestBlockstorageClientUpdateVolumeKmsKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("core", "UpdateVolumeKmsKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateVolumeKmsKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("core", "Blockstorage", "UpdateVolumeKmsKey", createBlockstorageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(core.BlockstorageClient)

	body, err := testClient.getRequests("core", "UpdateVolumeKmsKey")
	assert.NoError(t, err)

	type UpdateVolumeKmsKeyRequestInfo struct {
		ContainerId string
		Request     core.UpdateVolumeKmsKeyRequest
	}

	var requests []UpdateVolumeKmsKeyRequestInfo
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

			response, err := c.UpdateVolumeKmsKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
