package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/oda"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createOdaClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := oda.NewOdaClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="omce_devops_hybrid_us_grp@oracle.com" jiraProject="ODA" opsJiraProject="ODA"
func TestOdaClientChangeOdaInstanceCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oda", "ChangeOdaInstanceCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeOdaInstanceCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oda", "Oda", "ChangeOdaInstanceCompartment", createOdaClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oda.OdaClient)

	body, err := testClient.getRequests("oda", "ChangeOdaInstanceCompartment")
	assert.NoError(t, err)

	type ChangeOdaInstanceCompartmentRequestInfo struct {
		ContainerId string
		Request     oda.ChangeOdaInstanceCompartmentRequest
	}

	var requests []ChangeOdaInstanceCompartmentRequestInfo
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

			response, err := c.ChangeOdaInstanceCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="omce_devops_hybrid_us_grp@oracle.com" jiraProject="ODA" opsJiraProject="ODA"
func TestOdaClientCreateOdaInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oda", "CreateOdaInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateOdaInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oda", "Oda", "CreateOdaInstance", createOdaClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oda.OdaClient)

	body, err := testClient.getRequests("oda", "CreateOdaInstance")
	assert.NoError(t, err)

	type CreateOdaInstanceRequestInfo struct {
		ContainerId string
		Request     oda.CreateOdaInstanceRequest
	}

	var requests []CreateOdaInstanceRequestInfo
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

			response, err := c.CreateOdaInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="omce_devops_hybrid_us_grp@oracle.com" jiraProject="ODA" opsJiraProject="ODA"
func TestOdaClientDeleteOdaInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oda", "DeleteOdaInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteOdaInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oda", "Oda", "DeleteOdaInstance", createOdaClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oda.OdaClient)

	body, err := testClient.getRequests("oda", "DeleteOdaInstance")
	assert.NoError(t, err)

	type DeleteOdaInstanceRequestInfo struct {
		ContainerId string
		Request     oda.DeleteOdaInstanceRequest
	}

	var requests []DeleteOdaInstanceRequestInfo
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

			response, err := c.DeleteOdaInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="omce_devops_hybrid_us_grp@oracle.com" jiraProject="ODA" opsJiraProject="ODA"
func TestOdaClientGetOdaInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oda", "GetOdaInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetOdaInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oda", "Oda", "GetOdaInstance", createOdaClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oda.OdaClient)

	body, err := testClient.getRequests("oda", "GetOdaInstance")
	assert.NoError(t, err)

	type GetOdaInstanceRequestInfo struct {
		ContainerId string
		Request     oda.GetOdaInstanceRequest
	}

	var requests []GetOdaInstanceRequestInfo
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

			response, err := c.GetOdaInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="omce_devops_hybrid_us_grp@oracle.com" jiraProject="ODA" opsJiraProject="ODA"
func TestOdaClientGetWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oda", "GetWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oda", "Oda", "GetWorkRequest", createOdaClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oda.OdaClient)

	body, err := testClient.getRequests("oda", "GetWorkRequest")
	assert.NoError(t, err)

	type GetWorkRequestRequestInfo struct {
		ContainerId string
		Request     oda.GetWorkRequestRequest
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

// IssueRoutingInfo tag="default" email="omce_devops_hybrid_us_grp@oracle.com" jiraProject="ODA" opsJiraProject="ODA"
func TestOdaClientListOdaInstances(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oda", "ListOdaInstances")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListOdaInstances is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oda", "Oda", "ListOdaInstances", createOdaClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oda.OdaClient)

	body, err := testClient.getRequests("oda", "ListOdaInstances")
	assert.NoError(t, err)

	type ListOdaInstancesRequestInfo struct {
		ContainerId string
		Request     oda.ListOdaInstancesRequest
	}

	var requests []ListOdaInstancesRequestInfo
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
				r := req.(*oda.ListOdaInstancesRequest)
				return c.ListOdaInstances(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]oda.ListOdaInstancesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(oda.ListOdaInstancesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="omce_devops_hybrid_us_grp@oracle.com" jiraProject="ODA" opsJiraProject="ODA"
func TestOdaClientListWorkRequestErrors(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oda", "ListWorkRequestErrors")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestErrors is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oda", "Oda", "ListWorkRequestErrors", createOdaClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oda.OdaClient)

	body, err := testClient.getRequests("oda", "ListWorkRequestErrors")
	assert.NoError(t, err)

	type ListWorkRequestErrorsRequestInfo struct {
		ContainerId string
		Request     oda.ListWorkRequestErrorsRequest
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
				r := req.(*oda.ListWorkRequestErrorsRequest)
				return c.ListWorkRequestErrors(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]oda.ListWorkRequestErrorsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(oda.ListWorkRequestErrorsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="omce_devops_hybrid_us_grp@oracle.com" jiraProject="ODA" opsJiraProject="ODA"
func TestOdaClientListWorkRequestLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oda", "ListWorkRequestLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oda", "Oda", "ListWorkRequestLogs", createOdaClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oda.OdaClient)

	body, err := testClient.getRequests("oda", "ListWorkRequestLogs")
	assert.NoError(t, err)

	type ListWorkRequestLogsRequestInfo struct {
		ContainerId string
		Request     oda.ListWorkRequestLogsRequest
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
				r := req.(*oda.ListWorkRequestLogsRequest)
				return c.ListWorkRequestLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]oda.ListWorkRequestLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(oda.ListWorkRequestLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="omce_devops_hybrid_us_grp@oracle.com" jiraProject="ODA" opsJiraProject="ODA"
func TestOdaClientListWorkRequests(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oda", "ListWorkRequests")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequests is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oda", "Oda", "ListWorkRequests", createOdaClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oda.OdaClient)

	body, err := testClient.getRequests("oda", "ListWorkRequests")
	assert.NoError(t, err)

	type ListWorkRequestsRequestInfo struct {
		ContainerId string
		Request     oda.ListWorkRequestsRequest
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
				r := req.(*oda.ListWorkRequestsRequest)
				return c.ListWorkRequests(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]oda.ListWorkRequestsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(oda.ListWorkRequestsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="omce_devops_hybrid_us_grp@oracle.com" jiraProject="ODA" opsJiraProject="ODA"
func TestOdaClientUpdateOdaInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oda", "UpdateOdaInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateOdaInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oda", "Oda", "UpdateOdaInstance", createOdaClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oda.OdaClient)

	body, err := testClient.getRequests("oda", "UpdateOdaInstance")
	assert.NoError(t, err)

	type UpdateOdaInstanceRequestInfo struct {
		ContainerId string
		Request     oda.UpdateOdaInstanceRequest
	}

	var requests []UpdateOdaInstanceRequestInfo
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

			response, err := c.UpdateOdaInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
