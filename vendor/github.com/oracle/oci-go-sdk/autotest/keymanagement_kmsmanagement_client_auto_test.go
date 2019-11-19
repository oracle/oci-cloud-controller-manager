package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/keymanagement"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createKmsManagementClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {
	client, err := keymanagement.NewKmsManagementClientWithConfigurationProvider(p, testConfig.Endpoint)
	return client, err
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientBackupKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "BackupKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("BackupKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "BackupKey", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "BackupKey")
	assert.NoError(t, err)

	type BackupKeyRequestInfo struct {
		ContainerId string
		Request     keymanagement.BackupKeyRequest
	}

	var requests []BackupKeyRequestInfo
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

			response, err := c.BackupKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientCancelKeyDeletion(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "CancelKeyDeletion")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CancelKeyDeletion is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "CancelKeyDeletion", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "CancelKeyDeletion")
	assert.NoError(t, err)

	type CancelKeyDeletionRequestInfo struct {
		ContainerId string
		Request     keymanagement.CancelKeyDeletionRequest
	}

	var requests []CancelKeyDeletionRequestInfo
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

			response, err := c.CancelKeyDeletion(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientCancelKeyVersionDeletion(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "CancelKeyVersionDeletion")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CancelKeyVersionDeletion is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "CancelKeyVersionDeletion", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "CancelKeyVersionDeletion")
	assert.NoError(t, err)

	type CancelKeyVersionDeletionRequestInfo struct {
		ContainerId string
		Request     keymanagement.CancelKeyVersionDeletionRequest
	}

	var requests []CancelKeyVersionDeletionRequestInfo
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

			response, err := c.CancelKeyVersionDeletion(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientChangeKeyCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "ChangeKeyCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeKeyCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "ChangeKeyCompartment", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "ChangeKeyCompartment")
	assert.NoError(t, err)

	type ChangeKeyCompartmentRequestInfo struct {
		ContainerId string
		Request     keymanagement.ChangeKeyCompartmentRequest
	}

	var requests []ChangeKeyCompartmentRequestInfo
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

			response, err := c.ChangeKeyCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientCreateKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "CreateKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "CreateKey", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "CreateKey")
	assert.NoError(t, err)

	type CreateKeyRequestInfo struct {
		ContainerId string
		Request     keymanagement.CreateKeyRequest
	}

	var requests []CreateKeyRequestInfo
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

			response, err := c.CreateKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientCreateKeyVersion(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "CreateKeyVersion")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateKeyVersion is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "CreateKeyVersion", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "CreateKeyVersion")
	assert.NoError(t, err)

	type CreateKeyVersionRequestInfo struct {
		ContainerId string
		Request     keymanagement.CreateKeyVersionRequest
	}

	var requests []CreateKeyVersionRequestInfo
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

			response, err := c.CreateKeyVersion(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientCreateWrappingKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "CreateWrappingKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateWrappingKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "CreateWrappingKey", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "CreateWrappingKey")
	assert.NoError(t, err)

	type CreateWrappingKeyRequestInfo struct {
		ContainerId string
		Request     keymanagement.CreateWrappingKeyRequest
	}

	var requests []CreateWrappingKeyRequestInfo
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

			response, err := c.CreateWrappingKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientDeleteWrappingKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "DeleteWrappingKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteWrappingKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "DeleteWrappingKey", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "DeleteWrappingKey")
	assert.NoError(t, err)

	type DeleteWrappingKeyRequestInfo struct {
		ContainerId string
		Request     keymanagement.DeleteWrappingKeyRequest
	}

	var requests []DeleteWrappingKeyRequestInfo
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

			response, err := c.DeleteWrappingKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientDisableKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "DisableKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DisableKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "DisableKey", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "DisableKey")
	assert.NoError(t, err)

	type DisableKeyRequestInfo struct {
		ContainerId string
		Request     keymanagement.DisableKeyRequest
	}

	var requests []DisableKeyRequestInfo
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

			response, err := c.DisableKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientEnableKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "EnableKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("EnableKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "EnableKey", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "EnableKey")
	assert.NoError(t, err)

	type EnableKeyRequestInfo struct {
		ContainerId string
		Request     keymanagement.EnableKeyRequest
	}

	var requests []EnableKeyRequestInfo
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

			response, err := c.EnableKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientGetKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "GetKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "GetKey", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "GetKey")
	assert.NoError(t, err)

	type GetKeyRequestInfo struct {
		ContainerId string
		Request     keymanagement.GetKeyRequest
	}

	var requests []GetKeyRequestInfo
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

			response, err := c.GetKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientGetKeyVersion(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "GetKeyVersion")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetKeyVersion is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "GetKeyVersion", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "GetKeyVersion")
	assert.NoError(t, err)

	type GetKeyVersionRequestInfo struct {
		ContainerId string
		Request     keymanagement.GetKeyVersionRequest
	}

	var requests []GetKeyVersionRequestInfo
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

			response, err := c.GetKeyVersion(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientGetWrappingKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "GetWrappingKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetWrappingKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "GetWrappingKey", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "GetWrappingKey")
	assert.NoError(t, err)

	type GetWrappingKeyRequestInfo struct {
		ContainerId string
		Request     keymanagement.GetWrappingKeyRequest
	}

	var requests []GetWrappingKeyRequestInfo
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

			response, err := c.GetWrappingKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientImportKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "ImportKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ImportKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "ImportKey", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "ImportKey")
	assert.NoError(t, err)

	type ImportKeyRequestInfo struct {
		ContainerId string
		Request     keymanagement.ImportKeyRequest
	}

	var requests []ImportKeyRequestInfo
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

			response, err := c.ImportKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientImportKeyVersion(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "ImportKeyVersion")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ImportKeyVersion is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "ImportKeyVersion", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "ImportKeyVersion")
	assert.NoError(t, err)

	type ImportKeyVersionRequestInfo struct {
		ContainerId string
		Request     keymanagement.ImportKeyVersionRequest
	}

	var requests []ImportKeyVersionRequestInfo
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

			response, err := c.ImportKeyVersion(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientListKeyVersions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "ListKeyVersions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListKeyVersions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "ListKeyVersions", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "ListKeyVersions")
	assert.NoError(t, err)

	type ListKeyVersionsRequestInfo struct {
		ContainerId string
		Request     keymanagement.ListKeyVersionsRequest
	}

	var requests []ListKeyVersionsRequestInfo
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
				r := req.(*keymanagement.ListKeyVersionsRequest)
				return c.ListKeyVersions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]keymanagement.ListKeyVersionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(keymanagement.ListKeyVersionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientListKeys(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "ListKeys")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListKeys is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "ListKeys", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "ListKeys")
	assert.NoError(t, err)

	type ListKeysRequestInfo struct {
		ContainerId string
		Request     keymanagement.ListKeysRequest
	}

	var requests []ListKeysRequestInfo
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
				r := req.(*keymanagement.ListKeysRequest)
				return c.ListKeys(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]keymanagement.ListKeysResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(keymanagement.ListKeysResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientRestoreKeyFromFile(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "RestoreKeyFromFile")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("RestoreKeyFromFile is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "RestoreKeyFromFile", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "RestoreKeyFromFile")
	assert.NoError(t, err)

	type RestoreKeyFromFileRequestInfo struct {
		ContainerId string
		Request     keymanagement.RestoreKeyFromFileRequest
	}

	var requests []RestoreKeyFromFileRequestInfo
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

			response, err := c.RestoreKeyFromFile(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientRestoreKeyFromObjectStore(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "RestoreKeyFromObjectStore")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("RestoreKeyFromObjectStore is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "RestoreKeyFromObjectStore", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "RestoreKeyFromObjectStore")
	assert.NoError(t, err)

	type RestoreKeyFromObjectStoreRequestInfo struct {
		ContainerId string
		Request     keymanagement.RestoreKeyFromObjectStoreRequest
	}

	var requests []RestoreKeyFromObjectStoreRequestInfo
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

			response, err := c.RestoreKeyFromObjectStore(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientScheduleKeyDeletion(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "ScheduleKeyDeletion")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ScheduleKeyDeletion is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "ScheduleKeyDeletion", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "ScheduleKeyDeletion")
	assert.NoError(t, err)

	type ScheduleKeyDeletionRequestInfo struct {
		ContainerId string
		Request     keymanagement.ScheduleKeyDeletionRequest
	}

	var requests []ScheduleKeyDeletionRequestInfo
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

			response, err := c.ScheduleKeyDeletion(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientScheduleKeyVersionDeletion(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "ScheduleKeyVersionDeletion")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ScheduleKeyVersionDeletion is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "ScheduleKeyVersionDeletion", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "ScheduleKeyVersionDeletion")
	assert.NoError(t, err)

	type ScheduleKeyVersionDeletionRequestInfo struct {
		ContainerId string
		Request     keymanagement.ScheduleKeyVersionDeletionRequest
	}

	var requests []ScheduleKeyVersionDeletionRequestInfo
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

			response, err := c.ScheduleKeyVersionDeletion(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="sparta_kms_us_grp@oracle.com" jiraProject="KMS" opsJiraProject="KMS"
func TestKmsManagementClientUpdateKey(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("keymanagement", "UpdateKey")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateKey is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("keymanagement", "KmsManagement", "UpdateKey", createKmsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(keymanagement.KmsManagementClient)

	body, err := testClient.getRequests("keymanagement", "UpdateKey")
	assert.NoError(t, err)

	type UpdateKeyRequestInfo struct {
		ContainerId string
		Request     keymanagement.UpdateKeyRequest
	}

	var requests []UpdateKeyRequestInfo
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

			response, err := c.UpdateKey(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
