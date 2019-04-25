package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/usage"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createUsageClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := usage.NewUsageClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="" email="" jiraProject="" opsJiraProject=""
func TestUsageClientGetSubscriptionInfo(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("usage", "GetSubscriptionInfo")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetSubscriptionInfo is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("usage", "Usage", "GetSubscriptionInfo", createUsageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(usage.UsageClient)

	body, err := testClient.getRequests("usage", "GetSubscriptionInfo")
	assert.NoError(t, err)

	type GetSubscriptionInfoRequestInfo struct {
		ContainerId string
		Request     usage.GetSubscriptionInfoRequest
	}

	var requests []GetSubscriptionInfoRequestInfo
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

			response, err := c.GetSubscriptionInfo(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="" email="" jiraProject="" opsJiraProject=""
func TestUsageClientListUsageRecords(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("usage", "ListUsageRecords")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListUsageRecords is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("usage", "Usage", "ListUsageRecords", createUsageClientWithProvider)
	assert.NoError(t, err)
	c := cc.(usage.UsageClient)

	body, err := testClient.getRequests("usage", "ListUsageRecords")
	assert.NoError(t, err)

	type ListUsageRecordsRequestInfo struct {
		ContainerId string
		Request     usage.ListUsageRecordsRequest
	}

	var requests []ListUsageRecordsRequestInfo
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
				r := req.(*usage.ListUsageRecordsRequest)
				return c.ListUsageRecords(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]usage.ListUsageRecordsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(usage.ListUsageRecordsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
