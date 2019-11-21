package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/functions"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createFunctionsManagementClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := functions.NewFunctionsManagementClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientChangeApplicationCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "ChangeApplicationCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeApplicationCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "ChangeApplicationCompartment", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "ChangeApplicationCompartment")
	assert.NoError(t, err)

	type ChangeApplicationCompartmentRequestInfo struct {
		ContainerId string
		Request     functions.ChangeApplicationCompartmentRequest
	}

	var requests []ChangeApplicationCompartmentRequestInfo
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

			response, err := c.ChangeApplicationCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientCreateApplication(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "CreateApplication")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateApplication is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "CreateApplication", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "CreateApplication")
	assert.NoError(t, err)

	type CreateApplicationRequestInfo struct {
		ContainerId string
		Request     functions.CreateApplicationRequest
	}

	var requests []CreateApplicationRequestInfo
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

			response, err := c.CreateApplication(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientCreateFunction(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "CreateFunction")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateFunction is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "CreateFunction", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "CreateFunction")
	assert.NoError(t, err)

	type CreateFunctionRequestInfo struct {
		ContainerId string
		Request     functions.CreateFunctionRequest
	}

	var requests []CreateFunctionRequestInfo
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

			response, err := c.CreateFunction(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientDeleteApplication(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "DeleteApplication")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteApplication is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "DeleteApplication", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "DeleteApplication")
	assert.NoError(t, err)

	type DeleteApplicationRequestInfo struct {
		ContainerId string
		Request     functions.DeleteApplicationRequest
	}

	var requests []DeleteApplicationRequestInfo
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

			response, err := c.DeleteApplication(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientDeleteFunction(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "DeleteFunction")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteFunction is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "DeleteFunction", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "DeleteFunction")
	assert.NoError(t, err)

	type DeleteFunctionRequestInfo struct {
		ContainerId string
		Request     functions.DeleteFunctionRequest
	}

	var requests []DeleteFunctionRequestInfo
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

			response, err := c.DeleteFunction(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientGetApplication(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "GetApplication")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetApplication is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "GetApplication", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "GetApplication")
	assert.NoError(t, err)

	type GetApplicationRequestInfo struct {
		ContainerId string
		Request     functions.GetApplicationRequest
	}

	var requests []GetApplicationRequestInfo
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

			response, err := c.GetApplication(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientGetFunction(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "GetFunction")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetFunction is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "GetFunction", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "GetFunction")
	assert.NoError(t, err)

	type GetFunctionRequestInfo struct {
		ContainerId string
		Request     functions.GetFunctionRequest
	}

	var requests []GetFunctionRequestInfo
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

			response, err := c.GetFunction(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientListApplications(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "ListApplications")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListApplications is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "ListApplications", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "ListApplications")
	assert.NoError(t, err)

	type ListApplicationsRequestInfo struct {
		ContainerId string
		Request     functions.ListApplicationsRequest
	}

	var requests []ListApplicationsRequestInfo
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
				r := req.(*functions.ListApplicationsRequest)
				return c.ListApplications(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]functions.ListApplicationsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(functions.ListApplicationsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientListFunctions(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "ListFunctions")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListFunctions is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "ListFunctions", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "ListFunctions")
	assert.NoError(t, err)

	type ListFunctionsRequestInfo struct {
		ContainerId string
		Request     functions.ListFunctionsRequest
	}

	var requests []ListFunctionsRequestInfo
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
				r := req.(*functions.ListFunctionsRequest)
				return c.ListFunctions(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]functions.ListFunctionsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(functions.ListFunctionsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientUpdateApplication(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "UpdateApplication")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateApplication is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "UpdateApplication", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "UpdateApplication")
	assert.NoError(t, err)

	type UpdateApplicationRequestInfo struct {
		ContainerId string
		Request     functions.UpdateApplicationRequest
	}

	var requests []UpdateApplicationRequestInfo
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

			response, err := c.UpdateApplication(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsManagementClientUpdateFunction(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "UpdateFunction")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateFunction is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsManagement", "UpdateFunction", createFunctionsManagementClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsManagementClient)

	body, err := testClient.getRequests("functions", "UpdateFunction")
	assert.NoError(t, err)

	type UpdateFunctionRequestInfo struct {
		ContainerId string
		Request     functions.UpdateFunctionRequest
	}

	var requests []UpdateFunctionRequestInfo
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

			response, err := c.UpdateFunction(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
