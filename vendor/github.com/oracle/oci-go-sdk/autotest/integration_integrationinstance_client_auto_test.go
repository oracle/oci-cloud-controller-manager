package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/integration"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createIntegrationInstanceClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := integration.NewIntegrationInstanceClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="&lt;tbd&gt;_ww@oracle.com" jiraProject="&lt;tbc&gt;" opsJiraProject="&lt;tbd&gt;"
func TestIntegrationInstanceClientChangeIntegrationInstanceCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("integration", "ChangeIntegrationInstanceCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeIntegrationInstanceCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("integration", "IntegrationInstance", "ChangeIntegrationInstanceCompartment", createIntegrationInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(integration.IntegrationInstanceClient)

	body, err := testClient.getRequests("integration", "ChangeIntegrationInstanceCompartment")
	assert.NoError(t, err)

	type ChangeIntegrationInstanceCompartmentRequestInfo struct {
		ContainerId string
		Request     integration.ChangeIntegrationInstanceCompartmentRequest
	}

	var requests []ChangeIntegrationInstanceCompartmentRequestInfo
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

			response, err := c.ChangeIntegrationInstanceCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="&lt;tbd&gt;_ww@oracle.com" jiraProject="&lt;tbc&gt;" opsJiraProject="&lt;tbd&gt;"
func TestIntegrationInstanceClientCreateIntegrationInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("integration", "CreateIntegrationInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateIntegrationInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("integration", "IntegrationInstance", "CreateIntegrationInstance", createIntegrationInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(integration.IntegrationInstanceClient)

	body, err := testClient.getRequests("integration", "CreateIntegrationInstance")
	assert.NoError(t, err)

	type CreateIntegrationInstanceRequestInfo struct {
		ContainerId string
		Request     integration.CreateIntegrationInstanceRequest
	}

	var requests []CreateIntegrationInstanceRequestInfo
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

			response, err := c.CreateIntegrationInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="&lt;tbd&gt;_ww@oracle.com" jiraProject="&lt;tbc&gt;" opsJiraProject="&lt;tbd&gt;"
func TestIntegrationInstanceClientDeleteIntegrationInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("integration", "DeleteIntegrationInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteIntegrationInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("integration", "IntegrationInstance", "DeleteIntegrationInstance", createIntegrationInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(integration.IntegrationInstanceClient)

	body, err := testClient.getRequests("integration", "DeleteIntegrationInstance")
	assert.NoError(t, err)

	type DeleteIntegrationInstanceRequestInfo struct {
		ContainerId string
		Request     integration.DeleteIntegrationInstanceRequest
	}

	var requests []DeleteIntegrationInstanceRequestInfo
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

			response, err := c.DeleteIntegrationInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="&lt;tbd&gt;_ww@oracle.com" jiraProject="&lt;tbc&gt;" opsJiraProject="&lt;tbd&gt;"
func TestIntegrationInstanceClientGetIntegrationInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("integration", "GetIntegrationInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetIntegrationInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("integration", "IntegrationInstance", "GetIntegrationInstance", createIntegrationInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(integration.IntegrationInstanceClient)

	body, err := testClient.getRequests("integration", "GetIntegrationInstance")
	assert.NoError(t, err)

	type GetIntegrationInstanceRequestInfo struct {
		ContainerId string
		Request     integration.GetIntegrationInstanceRequest
	}

	var requests []GetIntegrationInstanceRequestInfo
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

			response, err := c.GetIntegrationInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="&lt;tbd&gt;_ww@oracle.com" jiraProject="&lt;tbc&gt;" opsJiraProject="&lt;tbd&gt;"
func TestIntegrationInstanceClientGetWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("integration", "GetWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("integration", "IntegrationInstance", "GetWorkRequest", createIntegrationInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(integration.IntegrationInstanceClient)

	body, err := testClient.getRequests("integration", "GetWorkRequest")
	assert.NoError(t, err)

	type GetWorkRequestRequestInfo struct {
		ContainerId string
		Request     integration.GetWorkRequestRequest
	}

	var requests []GetWorkRequestRequestInfo
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

			response, err := c.GetWorkRequest(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="&lt;tbd&gt;_ww@oracle.com" jiraProject="&lt;tbc&gt;" opsJiraProject="&lt;tbd&gt;"
func TestIntegrationInstanceClientListIntegrationInstances(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("integration", "ListIntegrationInstances")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListIntegrationInstances is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("integration", "IntegrationInstance", "ListIntegrationInstances", createIntegrationInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(integration.IntegrationInstanceClient)

	body, err := testClient.getRequests("integration", "ListIntegrationInstances")
	assert.NoError(t, err)

	type ListIntegrationInstancesRequestInfo struct {
		ContainerId string
		Request     integration.ListIntegrationInstancesRequest
	}

	var requests []ListIntegrationInstancesRequestInfo
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
				r := req.(*integration.ListIntegrationInstancesRequest)
				return c.ListIntegrationInstances(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]integration.ListIntegrationInstancesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(integration.ListIntegrationInstancesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="&lt;tbd&gt;_ww@oracle.com" jiraProject="&lt;tbc&gt;" opsJiraProject="&lt;tbd&gt;"
func TestIntegrationInstanceClientListWorkRequestErrors(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("integration", "ListWorkRequestErrors")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestErrors is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("integration", "IntegrationInstance", "ListWorkRequestErrors", createIntegrationInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(integration.IntegrationInstanceClient)

	body, err := testClient.getRequests("integration", "ListWorkRequestErrors")
	assert.NoError(t, err)

	type ListWorkRequestErrorsRequestInfo struct {
		ContainerId string
		Request     integration.ListWorkRequestErrorsRequest
	}

	var requests []ListWorkRequestErrorsRequestInfo
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
				r := req.(*integration.ListWorkRequestErrorsRequest)
				return c.ListWorkRequestErrors(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]integration.ListWorkRequestErrorsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(integration.ListWorkRequestErrorsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="&lt;tbd&gt;_ww@oracle.com" jiraProject="&lt;tbc&gt;" opsJiraProject="&lt;tbd&gt;"
func TestIntegrationInstanceClientListWorkRequestLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("integration", "ListWorkRequestLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("integration", "IntegrationInstance", "ListWorkRequestLogs", createIntegrationInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(integration.IntegrationInstanceClient)

	body, err := testClient.getRequests("integration", "ListWorkRequestLogs")
	assert.NoError(t, err)

	type ListWorkRequestLogsRequestInfo struct {
		ContainerId string
		Request     integration.ListWorkRequestLogsRequest
	}

	var requests []ListWorkRequestLogsRequestInfo
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
				r := req.(*integration.ListWorkRequestLogsRequest)
				return c.ListWorkRequestLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]integration.ListWorkRequestLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(integration.ListWorkRequestLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="&lt;tbd&gt;_ww@oracle.com" jiraProject="&lt;tbc&gt;" opsJiraProject="&lt;tbd&gt;"
func TestIntegrationInstanceClientListWorkRequests(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("integration", "ListWorkRequests")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequests is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("integration", "IntegrationInstance", "ListWorkRequests", createIntegrationInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(integration.IntegrationInstanceClient)

	body, err := testClient.getRequests("integration", "ListWorkRequests")
	assert.NoError(t, err)

	type ListWorkRequestsRequestInfo struct {
		ContainerId string
		Request     integration.ListWorkRequestsRequest
	}

	var requests []ListWorkRequestsRequestInfo
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
				r := req.(*integration.ListWorkRequestsRequest)
				return c.ListWorkRequests(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]integration.ListWorkRequestsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(integration.ListWorkRequestsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="&lt;tbd&gt;_ww@oracle.com" jiraProject="&lt;tbc&gt;" opsJiraProject="&lt;tbd&gt;"
func TestIntegrationInstanceClientUpdateIntegrationInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("integration", "UpdateIntegrationInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateIntegrationInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("integration", "IntegrationInstance", "UpdateIntegrationInstance", createIntegrationInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(integration.IntegrationInstanceClient)

	body, err := testClient.getRequests("integration", "UpdateIntegrationInstance")
	assert.NoError(t, err)

	type UpdateIntegrationInstanceRequestInfo struct {
		ContainerId string
		Request     integration.UpdateIntegrationInstanceRequest
	}

	var requests []UpdateIntegrationInstanceRequestInfo
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

			response, err := c.UpdateIntegrationInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
