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

func createTransferApplianceClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := dts.NewTransferApplianceClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceClientCreateTransferAppliance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "CreateTransferAppliance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateTransferAppliance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferAppliance", "CreateTransferAppliance", createTransferApplianceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceClient)

	body, err := testClient.getRequests("dts", "CreateTransferAppliance")
	assert.NoError(t, err)

	type CreateTransferApplianceRequestInfo struct {
		ContainerId string
		Request     dts.CreateTransferApplianceRequest
	}

	var requests []CreateTransferApplianceRequestInfo
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

			response, err := c.CreateTransferAppliance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceClientCreateTransferApplianceAdminCredentials(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "CreateTransferApplianceAdminCredentials")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateTransferApplianceAdminCredentials is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferAppliance", "CreateTransferApplianceAdminCredentials", createTransferApplianceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceClient)

	body, err := testClient.getRequests("dts", "CreateTransferApplianceAdminCredentials")
	assert.NoError(t, err)

	type CreateTransferApplianceAdminCredentialsRequestInfo struct {
		ContainerId string
		Request     dts.CreateTransferApplianceAdminCredentialsRequest
	}

	var requests []CreateTransferApplianceAdminCredentialsRequestInfo
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

			response, err := c.CreateTransferApplianceAdminCredentials(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceClientDeleteTransferAppliance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "DeleteTransferAppliance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteTransferAppliance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferAppliance", "DeleteTransferAppliance", createTransferApplianceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceClient)

	body, err := testClient.getRequests("dts", "DeleteTransferAppliance")
	assert.NoError(t, err)

	type DeleteTransferApplianceRequestInfo struct {
		ContainerId string
		Request     dts.DeleteTransferApplianceRequest
	}

	var requests []DeleteTransferApplianceRequestInfo
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

			response, err := c.DeleteTransferAppliance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceClientGetTransferAppliance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "GetTransferAppliance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetTransferAppliance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferAppliance", "GetTransferAppliance", createTransferApplianceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceClient)

	body, err := testClient.getRequests("dts", "GetTransferAppliance")
	assert.NoError(t, err)

	type GetTransferApplianceRequestInfo struct {
		ContainerId string
		Request     dts.GetTransferApplianceRequest
	}

	var requests []GetTransferApplianceRequestInfo
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

			response, err := c.GetTransferAppliance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceClientGetTransferApplianceCertificateAuthorityCertificate(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "GetTransferApplianceCertificateAuthorityCertificate")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetTransferApplianceCertificateAuthorityCertificate is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferAppliance", "GetTransferApplianceCertificateAuthorityCertificate", createTransferApplianceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceClient)

	body, err := testClient.getRequests("dts", "GetTransferApplianceCertificateAuthorityCertificate")
	assert.NoError(t, err)

	type GetTransferApplianceCertificateAuthorityCertificateRequestInfo struct {
		ContainerId string
		Request     dts.GetTransferApplianceCertificateAuthorityCertificateRequest
	}

	var requests []GetTransferApplianceCertificateAuthorityCertificateRequestInfo
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

			response, err := c.GetTransferApplianceCertificateAuthorityCertificate(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceClientGetTransferApplianceEncryptionPassphrase(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "GetTransferApplianceEncryptionPassphrase")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetTransferApplianceEncryptionPassphrase is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferAppliance", "GetTransferApplianceEncryptionPassphrase", createTransferApplianceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceClient)

	body, err := testClient.getRequests("dts", "GetTransferApplianceEncryptionPassphrase")
	assert.NoError(t, err)

	type GetTransferApplianceEncryptionPassphraseRequestInfo struct {
		ContainerId string
		Request     dts.GetTransferApplianceEncryptionPassphraseRequest
	}

	var requests []GetTransferApplianceEncryptionPassphraseRequestInfo
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

			response, err := c.GetTransferApplianceEncryptionPassphrase(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceClientListTransferAppliances(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "ListTransferAppliances")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListTransferAppliances is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferAppliance", "ListTransferAppliances", createTransferApplianceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceClient)

	body, err := testClient.getRequests("dts", "ListTransferAppliances")
	assert.NoError(t, err)

	type ListTransferAppliancesRequestInfo struct {
		ContainerId string
		Request     dts.ListTransferAppliancesRequest
	}

	var requests []ListTransferAppliancesRequestInfo
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

			response, err := c.ListTransferAppliances(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceClientUpdateTransferAppliance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "UpdateTransferAppliance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateTransferAppliance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferAppliance", "UpdateTransferAppliance", createTransferApplianceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceClient)

	body, err := testClient.getRequests("dts", "UpdateTransferAppliance")
	assert.NoError(t, err)

	type UpdateTransferApplianceRequestInfo struct {
		ContainerId string
		Request     dts.UpdateTransferApplianceRequest
	}

	var requests []UpdateTransferApplianceRequestInfo
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

			response, err := c.UpdateTransferAppliance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
