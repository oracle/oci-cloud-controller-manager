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

func createTransferApplianceEntitlementClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := dts.NewTransferApplianceEntitlementClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceEntitlementClientCreateTransferApplianceEntitlement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "CreateTransferApplianceEntitlement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateTransferApplianceEntitlement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferApplianceEntitlement", "CreateTransferApplianceEntitlement", createTransferApplianceEntitlementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceEntitlementClient)

	body, err := testClient.getRequests("dts", "CreateTransferApplianceEntitlement")
	assert.NoError(t, err)

	type CreateTransferApplianceEntitlementRequestInfo struct {
		ContainerId string
		Request     dts.CreateTransferApplianceEntitlementRequest
	}

	var requests []CreateTransferApplianceEntitlementRequestInfo
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

			response, err := c.CreateTransferApplianceEntitlement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceEntitlementClientGetTransferApplianceEntitlement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "GetTransferApplianceEntitlement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetTransferApplianceEntitlement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferApplianceEntitlement", "GetTransferApplianceEntitlement", createTransferApplianceEntitlementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceEntitlementClient)

	body, err := testClient.getRequests("dts", "GetTransferApplianceEntitlement")
	assert.NoError(t, err)

	type GetTransferApplianceEntitlementRequestInfo struct {
		ContainerId string
		Request     dts.GetTransferApplianceEntitlementRequest
	}

	var requests []GetTransferApplianceEntitlementRequestInfo
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

			response, err := c.GetTransferApplianceEntitlement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestTransferApplianceEntitlementClientListTransferApplianceEntitlement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "ListTransferApplianceEntitlement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListTransferApplianceEntitlement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "TransferApplianceEntitlement", "ListTransferApplianceEntitlement", createTransferApplianceEntitlementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.TransferApplianceEntitlementClient)

	body, err := testClient.getRequests("dts", "ListTransferApplianceEntitlement")
	assert.NoError(t, err)

	type ListTransferApplianceEntitlementRequestInfo struct {
		ContainerId string
		Request     dts.ListTransferApplianceEntitlementRequest
	}

	var requests []ListTransferApplianceEntitlementRequestInfo
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

			response, err := c.ListTransferApplianceEntitlement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
