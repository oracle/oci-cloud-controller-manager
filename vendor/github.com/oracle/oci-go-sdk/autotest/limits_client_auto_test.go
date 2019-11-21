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

func createLimitsClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := limits.NewLimitsClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="platform_limits_grp@oracle.com" jiraProject="LIM" opsJiraProject="LIM"
func TestLimitsClientGetResourceAvailability(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("limits", "GetResourceAvailability")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetResourceAvailability is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("limits", "Limits", "GetResourceAvailability", createLimitsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(limits.LimitsClient)

	body, err := testClient.getRequests("limits", "GetResourceAvailability")
	assert.NoError(t, err)

	type GetResourceAvailabilityRequestInfo struct {
		ContainerId string
		Request     limits.GetResourceAvailabilityRequest
	}

	var requests []GetResourceAvailabilityRequestInfo
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

			response, err := c.GetResourceAvailability(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="platform_limits_grp@oracle.com" jiraProject="LIM" opsJiraProject="LIM"
func TestLimitsClientListLimitDefinitions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("limits", "ListLimitDefinitions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListLimitDefinitions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("limits", "Limits", "ListLimitDefinitions", createLimitsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(limits.LimitsClient)

	body, err := testClient.getRequests("limits", "ListLimitDefinitions")
	assert.NoError(t, err)

	type ListLimitDefinitionsRequestInfo struct {
		ContainerId string
		Request     limits.ListLimitDefinitionsRequest
	}

	var requests []ListLimitDefinitionsRequestInfo
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
				r := req.(*limits.ListLimitDefinitionsRequest)
				return c.ListLimitDefinitions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]limits.ListLimitDefinitionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(limits.ListLimitDefinitionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="platform_limits_grp@oracle.com" jiraProject="LIM" opsJiraProject="LIM"
func TestLimitsClientListLimitValues(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("limits", "ListLimitValues")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListLimitValues is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("limits", "Limits", "ListLimitValues", createLimitsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(limits.LimitsClient)

	body, err := testClient.getRequests("limits", "ListLimitValues")
	assert.NoError(t, err)

	type ListLimitValuesRequestInfo struct {
		ContainerId string
		Request     limits.ListLimitValuesRequest
	}

	var requests []ListLimitValuesRequestInfo
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
				r := req.(*limits.ListLimitValuesRequest)
				return c.ListLimitValues(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]limits.ListLimitValuesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(limits.ListLimitValuesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="platform_limits_grp@oracle.com" jiraProject="LIM" opsJiraProject="LIM"
func TestLimitsClientListServices(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("limits", "ListServices")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListServices is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("limits", "Limits", "ListServices", createLimitsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(limits.LimitsClient)

	body, err := testClient.getRequests("limits", "ListServices")
	assert.NoError(t, err)

	type ListServicesRequestInfo struct {
		ContainerId string
		Request     limits.ListServicesRequest
	}

	var requests []ListServicesRequestInfo
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
				r := req.(*limits.ListServicesRequest)
				return c.ListServices(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]limits.ListServicesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(limits.ListServicesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
