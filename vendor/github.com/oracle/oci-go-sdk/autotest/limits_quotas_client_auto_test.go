package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/limits"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createQuotasClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := limits.NewQuotasClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="platform_limits_grp@oracle.com" jiraProject="LIM" opsJiraProject="LIM"
func TestQuotasClientCreateQuota(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("limits", "CreateQuota")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateQuota is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("limits", "Quotas", "CreateQuota", createQuotasClientWithProvider)
	assert.NoError(t, err)
	c := cc.(limits.QuotasClient)

	body, err := testClient.getRequests("limits", "CreateQuota")
	assert.NoError(t, err)

	type CreateQuotaRequestInfo struct {
		ContainerId string
		Request     limits.CreateQuotaRequest
	}

	var requests []CreateQuotaRequestInfo
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

			response, err := c.CreateQuota(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="platform_limits_grp@oracle.com" jiraProject="LIM" opsJiraProject="LIM"
func TestQuotasClientDeleteQuota(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("limits", "DeleteQuota")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteQuota is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("limits", "Quotas", "DeleteQuota", createQuotasClientWithProvider)
	assert.NoError(t, err)
	c := cc.(limits.QuotasClient)

	body, err := testClient.getRequests("limits", "DeleteQuota")
	assert.NoError(t, err)

	type DeleteQuotaRequestInfo struct {
		ContainerId string
		Request     limits.DeleteQuotaRequest
	}

	var requests []DeleteQuotaRequestInfo
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

			response, err := c.DeleteQuota(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="platform_limits_grp@oracle.com" jiraProject="LIM" opsJiraProject="LIM"
func TestQuotasClientGetQuota(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("limits", "GetQuota")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetQuota is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("limits", "Quotas", "GetQuota", createQuotasClientWithProvider)
	assert.NoError(t, err)
	c := cc.(limits.QuotasClient)

	body, err := testClient.getRequests("limits", "GetQuota")
	assert.NoError(t, err)

	type GetQuotaRequestInfo struct {
		ContainerId string
		Request     limits.GetQuotaRequest
	}

	var requests []GetQuotaRequestInfo
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

			response, err := c.GetQuota(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="platform_limits_grp@oracle.com" jiraProject="LIM" opsJiraProject="LIM"
func TestQuotasClientListQuotas(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("limits", "ListQuotas")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListQuotas is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("limits", "Quotas", "ListQuotas", createQuotasClientWithProvider)
	assert.NoError(t, err)
	c := cc.(limits.QuotasClient)

	body, err := testClient.getRequests("limits", "ListQuotas")
	assert.NoError(t, err)

	type ListQuotasRequestInfo struct {
		ContainerId string
		Request     limits.ListQuotasRequest
	}

	var requests []ListQuotasRequestInfo
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
				r := req.(*limits.ListQuotasRequest)
				return c.ListQuotas(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]limits.ListQuotasResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(limits.ListQuotasResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="platform_limits_grp@oracle.com" jiraProject="LIM" opsJiraProject="LIM"
func TestQuotasClientUpdateQuota(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("limits", "UpdateQuota")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateQuota is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("limits", "Quotas", "UpdateQuota", createQuotasClientWithProvider)
	assert.NoError(t, err)
	c := cc.(limits.QuotasClient)

	body, err := testClient.getRequests("limits", "UpdateQuota")
	assert.NoError(t, err)

	type UpdateQuotaRequestInfo struct {
		ContainerId string
		Request     limits.UpdateQuotaRequest
	}

	var requests []UpdateQuotaRequestInfo
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

			response, err := c.UpdateQuota(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
