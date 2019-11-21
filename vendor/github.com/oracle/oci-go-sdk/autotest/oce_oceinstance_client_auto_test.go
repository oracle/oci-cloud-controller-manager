package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/oce"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createOceInstanceClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := oce.NewOceInstanceClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="cec_devops_ww_grp@oracle.com" jiraProject="CEC" opsJiraProject="CECA"
func TestOceInstanceClientChangeOceInstanceCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oce", "ChangeOceInstanceCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeOceInstanceCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oce", "OceInstance", "ChangeOceInstanceCompartment", createOceInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oce.OceInstanceClient)

	body, err := testClient.getRequests("oce", "ChangeOceInstanceCompartment")
	assert.NoError(t, err)

	type ChangeOceInstanceCompartmentRequestInfo struct {
		ContainerId string
		Request     oce.ChangeOceInstanceCompartmentRequest
	}

	var requests []ChangeOceInstanceCompartmentRequestInfo
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

			response, err := c.ChangeOceInstanceCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="cec_devops_ww_grp@oracle.com" jiraProject="CEC" opsJiraProject="CECA"
func TestOceInstanceClientCreateOceInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oce", "CreateOceInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateOceInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oce", "OceInstance", "CreateOceInstance", createOceInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oce.OceInstanceClient)

	body, err := testClient.getRequests("oce", "CreateOceInstance")
	assert.NoError(t, err)

	type CreateOceInstanceRequestInfo struct {
		ContainerId string
		Request     oce.CreateOceInstanceRequest
	}

	var requests []CreateOceInstanceRequestInfo
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

			response, err := c.CreateOceInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="cec_devops_ww_grp@oracle.com" jiraProject="CEC" opsJiraProject="CECA"
func TestOceInstanceClientDeleteOceInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oce", "DeleteOceInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteOceInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oce", "OceInstance", "DeleteOceInstance", createOceInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oce.OceInstanceClient)

	body, err := testClient.getRequests("oce", "DeleteOceInstance")
	assert.NoError(t, err)

	type DeleteOceInstanceRequestInfo struct {
		ContainerId string
		Request     oce.DeleteOceInstanceRequest
	}

	var requests []DeleteOceInstanceRequestInfo
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

			response, err := c.DeleteOceInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="cec_devops_ww_grp@oracle.com" jiraProject="CEC" opsJiraProject="CECA"
func TestOceInstanceClientGetOceInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oce", "GetOceInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetOceInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oce", "OceInstance", "GetOceInstance", createOceInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oce.OceInstanceClient)

	body, err := testClient.getRequests("oce", "GetOceInstance")
	assert.NoError(t, err)

	type GetOceInstanceRequestInfo struct {
		ContainerId string
		Request     oce.GetOceInstanceRequest
	}

	var requests []GetOceInstanceRequestInfo
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

			response, err := c.GetOceInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="cec_devops_ww_grp@oracle.com" jiraProject="CEC" opsJiraProject="CECA"
func TestOceInstanceClientGetWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oce", "GetWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oce", "OceInstance", "GetWorkRequest", createOceInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oce.OceInstanceClient)

	body, err := testClient.getRequests("oce", "GetWorkRequest")
	assert.NoError(t, err)

	type GetWorkRequestRequestInfo struct {
		ContainerId string
		Request     oce.GetWorkRequestRequest
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

// IssueRoutingInfo tag="default" email="cec_devops_ww_grp@oracle.com" jiraProject="CEC" opsJiraProject="CECA"
func TestOceInstanceClientListOceInstances(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oce", "ListOceInstances")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListOceInstances is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oce", "OceInstance", "ListOceInstances", createOceInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oce.OceInstanceClient)

	body, err := testClient.getRequests("oce", "ListOceInstances")
	assert.NoError(t, err)

	type ListOceInstancesRequestInfo struct {
		ContainerId string
		Request     oce.ListOceInstancesRequest
	}

	var requests []ListOceInstancesRequestInfo
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
				r := req.(*oce.ListOceInstancesRequest)
				return c.ListOceInstances(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]oce.ListOceInstancesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(oce.ListOceInstancesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="cec_devops_ww_grp@oracle.com" jiraProject="CEC" opsJiraProject="CECA"
func TestOceInstanceClientListWorkRequestErrors(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oce", "ListWorkRequestErrors")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestErrors is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oce", "OceInstance", "ListWorkRequestErrors", createOceInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oce.OceInstanceClient)

	body, err := testClient.getRequests("oce", "ListWorkRequestErrors")
	assert.NoError(t, err)

	type ListWorkRequestErrorsRequestInfo struct {
		ContainerId string
		Request     oce.ListWorkRequestErrorsRequest
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
				r := req.(*oce.ListWorkRequestErrorsRequest)
				return c.ListWorkRequestErrors(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]oce.ListWorkRequestErrorsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(oce.ListWorkRequestErrorsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="cec_devops_ww_grp@oracle.com" jiraProject="CEC" opsJiraProject="CECA"
func TestOceInstanceClientListWorkRequestLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oce", "ListWorkRequestLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oce", "OceInstance", "ListWorkRequestLogs", createOceInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oce.OceInstanceClient)

	body, err := testClient.getRequests("oce", "ListWorkRequestLogs")
	assert.NoError(t, err)

	type ListWorkRequestLogsRequestInfo struct {
		ContainerId string
		Request     oce.ListWorkRequestLogsRequest
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
				r := req.(*oce.ListWorkRequestLogsRequest)
				return c.ListWorkRequestLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]oce.ListWorkRequestLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(oce.ListWorkRequestLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="cec_devops_ww_grp@oracle.com" jiraProject="CEC" opsJiraProject="CECA"
func TestOceInstanceClientListWorkRequests(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oce", "ListWorkRequests")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequests is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oce", "OceInstance", "ListWorkRequests", createOceInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oce.OceInstanceClient)

	body, err := testClient.getRequests("oce", "ListWorkRequests")
	assert.NoError(t, err)

	type ListWorkRequestsRequestInfo struct {
		ContainerId string
		Request     oce.ListWorkRequestsRequest
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
				r := req.(*oce.ListWorkRequestsRequest)
				return c.ListWorkRequests(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]oce.ListWorkRequestsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(oce.ListWorkRequestsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="cec_devops_ww_grp@oracle.com" jiraProject="CEC" opsJiraProject="CECA"
func TestOceInstanceClientUpdateOceInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("oce", "UpdateOceInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateOceInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("oce", "OceInstance", "UpdateOceInstance", createOceInstanceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(oce.OceInstanceClient)

	body, err := testClient.getRequests("oce", "UpdateOceInstance")
	assert.NoError(t, err)

	type UpdateOceInstanceRequestInfo struct {
		ContainerId string
		Request     oce.UpdateOceInstanceRequest
	}

	var requests []UpdateOceInstanceRequestInfo
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

			response, err := c.UpdateOceInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
