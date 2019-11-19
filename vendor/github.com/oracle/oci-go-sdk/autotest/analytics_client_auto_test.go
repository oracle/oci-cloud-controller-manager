package autotest

import (
	"github.com/oracle/oci-go-sdk/analytics"
	"github.com/oracle/oci-go-sdk/common"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createAnalyticsClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := analytics.NewAnalyticsClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientChangeAnalyticsInstanceCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "ChangeAnalyticsInstanceCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeAnalyticsInstanceCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "ChangeAnalyticsInstanceCompartment", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "ChangeAnalyticsInstanceCompartment")
	assert.NoError(t, err)

	type ChangeAnalyticsInstanceCompartmentRequestInfo struct {
		ContainerId string
		Request     analytics.ChangeAnalyticsInstanceCompartmentRequest
	}

	var requests []ChangeAnalyticsInstanceCompartmentRequestInfo
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

			response, err := c.ChangeAnalyticsInstanceCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientCreateAnalyticsInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "CreateAnalyticsInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateAnalyticsInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "CreateAnalyticsInstance", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "CreateAnalyticsInstance")
	assert.NoError(t, err)

	type CreateAnalyticsInstanceRequestInfo struct {
		ContainerId string
		Request     analytics.CreateAnalyticsInstanceRequest
	}

	var requests []CreateAnalyticsInstanceRequestInfo
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

			response, err := c.CreateAnalyticsInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientDeleteAnalyticsInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "DeleteAnalyticsInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteAnalyticsInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "DeleteAnalyticsInstance", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "DeleteAnalyticsInstance")
	assert.NoError(t, err)

	type DeleteAnalyticsInstanceRequestInfo struct {
		ContainerId string
		Request     analytics.DeleteAnalyticsInstanceRequest
	}

	var requests []DeleteAnalyticsInstanceRequestInfo
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

			response, err := c.DeleteAnalyticsInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientDeleteWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "DeleteWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "DeleteWorkRequest", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "DeleteWorkRequest")
	assert.NoError(t, err)

	type DeleteWorkRequestRequestInfo struct {
		ContainerId string
		Request     analytics.DeleteWorkRequestRequest
	}

	var requests []DeleteWorkRequestRequestInfo
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

			response, err := c.DeleteWorkRequest(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientGetAnalyticsInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "GetAnalyticsInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetAnalyticsInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "GetAnalyticsInstance", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "GetAnalyticsInstance")
	assert.NoError(t, err)

	type GetAnalyticsInstanceRequestInfo struct {
		ContainerId string
		Request     analytics.GetAnalyticsInstanceRequest
	}

	var requests []GetAnalyticsInstanceRequestInfo
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

			response, err := c.GetAnalyticsInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientGetWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "GetWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "GetWorkRequest", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "GetWorkRequest")
	assert.NoError(t, err)

	type GetWorkRequestRequestInfo struct {
		ContainerId string
		Request     analytics.GetWorkRequestRequest
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

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientListAnalyticsInstances(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "ListAnalyticsInstances")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListAnalyticsInstances is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "ListAnalyticsInstances", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "ListAnalyticsInstances")
	assert.NoError(t, err)

	type ListAnalyticsInstancesRequestInfo struct {
		ContainerId string
		Request     analytics.ListAnalyticsInstancesRequest
	}

	var requests []ListAnalyticsInstancesRequestInfo
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
				r := req.(*analytics.ListAnalyticsInstancesRequest)
				return c.ListAnalyticsInstances(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]analytics.ListAnalyticsInstancesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(analytics.ListAnalyticsInstancesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientListWorkRequestErrors(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "ListWorkRequestErrors")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestErrors is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "ListWorkRequestErrors", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "ListWorkRequestErrors")
	assert.NoError(t, err)

	type ListWorkRequestErrorsRequestInfo struct {
		ContainerId string
		Request     analytics.ListWorkRequestErrorsRequest
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
				r := req.(*analytics.ListWorkRequestErrorsRequest)
				return c.ListWorkRequestErrors(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]analytics.ListWorkRequestErrorsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(analytics.ListWorkRequestErrorsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientListWorkRequestLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "ListWorkRequestLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "ListWorkRequestLogs", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "ListWorkRequestLogs")
	assert.NoError(t, err)

	type ListWorkRequestLogsRequestInfo struct {
		ContainerId string
		Request     analytics.ListWorkRequestLogsRequest
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
				r := req.(*analytics.ListWorkRequestLogsRequest)
				return c.ListWorkRequestLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]analytics.ListWorkRequestLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(analytics.ListWorkRequestLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientListWorkRequests(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "ListWorkRequests")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequests is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "ListWorkRequests", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "ListWorkRequests")
	assert.NoError(t, err)

	type ListWorkRequestsRequestInfo struct {
		ContainerId string
		Request     analytics.ListWorkRequestsRequest
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
				r := req.(*analytics.ListWorkRequestsRequest)
				return c.ListWorkRequests(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]analytics.ListWorkRequestsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(analytics.ListWorkRequestsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientScaleAnalyticsInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "ScaleAnalyticsInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ScaleAnalyticsInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "ScaleAnalyticsInstance", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "ScaleAnalyticsInstance")
	assert.NoError(t, err)

	type ScaleAnalyticsInstanceRequestInfo struct {
		ContainerId string
		Request     analytics.ScaleAnalyticsInstanceRequest
	}

	var requests []ScaleAnalyticsInstanceRequestInfo
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

			response, err := c.ScaleAnalyticsInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientStartAnalyticsInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "StartAnalyticsInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("StartAnalyticsInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "StartAnalyticsInstance", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "StartAnalyticsInstance")
	assert.NoError(t, err)

	type StartAnalyticsInstanceRequestInfo struct {
		ContainerId string
		Request     analytics.StartAnalyticsInstanceRequest
	}

	var requests []StartAnalyticsInstanceRequestInfo
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

			response, err := c.StartAnalyticsInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientStopAnalyticsInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "StopAnalyticsInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("StopAnalyticsInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "StopAnalyticsInstance", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "StopAnalyticsInstance")
	assert.NoError(t, err)

	type StopAnalyticsInstanceRequestInfo struct {
		ContainerId string
		Request     analytics.StopAnalyticsInstanceRequest
	}

	var requests []StopAnalyticsInstanceRequestInfo
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

			response, err := c.StopAnalyticsInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_oac_ww_grp@oracle.com" jiraProject="OB" opsJiraProject="AOAC"
func TestAnalyticsClientUpdateAnalyticsInstance(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("analytics", "UpdateAnalyticsInstance")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateAnalyticsInstance is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("analytics", "Analytics", "UpdateAnalyticsInstance", createAnalyticsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(analytics.AnalyticsClient)

	body, err := testClient.getRequests("analytics", "UpdateAnalyticsInstance")
	assert.NoError(t, err)

	type UpdateAnalyticsInstanceRequestInfo struct {
		ContainerId string
		Request     analytics.UpdateAnalyticsInstanceRequest
	}

	var requests []UpdateAnalyticsInstanceRequestInfo
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

			response, err := c.UpdateAnalyticsInstance(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
