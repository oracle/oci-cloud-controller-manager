package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/kam"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createKamClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := kam.NewKamClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="kam-service-support_ww_grp@oracle.com" jiraProject="CNP" opsJiraProject="SMESH"
func TestKamClientCreateKamRelease(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("kam", "CreateKamRelease")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateKamRelease is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("kam", "Kam", "CreateKamRelease", createKamClientWithProvider)
	assert.NoError(t, err)
	c := cc.(kam.KamClient)

	body, err := testClient.getRequests("kam", "CreateKamRelease")
	assert.NoError(t, err)

	type CreateKamReleaseRequestInfo struct {
		ContainerId string
		Request     kam.CreateKamReleaseRequest
	}

	var requests []CreateKamReleaseRequestInfo
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

			response, err := c.CreateKamRelease(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="kam-service-support_ww_grp@oracle.com" jiraProject="CNP" opsJiraProject="SMESH"
func TestKamClientDeleteKamRelease(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("kam", "DeleteKamRelease")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteKamRelease is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("kam", "Kam", "DeleteKamRelease", createKamClientWithProvider)
	assert.NoError(t, err)
	c := cc.(kam.KamClient)

	body, err := testClient.getRequests("kam", "DeleteKamRelease")
	assert.NoError(t, err)

	type DeleteKamReleaseRequestInfo struct {
		ContainerId string
		Request     kam.DeleteKamReleaseRequest
	}

	var requests []DeleteKamReleaseRequestInfo
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

			response, err := c.DeleteKamRelease(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="kam-service-support_ww_grp@oracle.com" jiraProject="CNP" opsJiraProject="SMESH"
func TestKamClientGetWorkRequest(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("kam", "GetWorkRequest")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetWorkRequest is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("kam", "Kam", "GetWorkRequest", createKamClientWithProvider)
	assert.NoError(t, err)
	c := cc.(kam.KamClient)

	body, err := testClient.getRequests("kam", "GetWorkRequest")
	assert.NoError(t, err)

	type GetWorkRequestRequestInfo struct {
		ContainerId string
		Request     kam.GetWorkRequestRequest
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

// IssueRoutingInfo tag="default" email="kam-service-support_ww_grp@oracle.com" jiraProject="CNP" opsJiraProject="SMESH"
func TestKamClientListKamCharts(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("kam", "ListKamCharts")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListKamCharts is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("kam", "Kam", "ListKamCharts", createKamClientWithProvider)
	assert.NoError(t, err)
	c := cc.(kam.KamClient)

	body, err := testClient.getRequests("kam", "ListKamCharts")
	assert.NoError(t, err)

	type ListKamChartsRequestInfo struct {
		ContainerId string
		Request     kam.ListKamChartsRequest
	}

	var requests []ListKamChartsRequestInfo
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
				r := req.(*kam.ListKamChartsRequest)
				return c.ListKamCharts(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]kam.ListKamChartsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(kam.ListKamChartsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="kam-service-support_ww_grp@oracle.com" jiraProject="CNP" opsJiraProject="SMESH"
func TestKamClientListKamReleases(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("kam", "ListKamReleases")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListKamReleases is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("kam", "Kam", "ListKamReleases", createKamClientWithProvider)
	assert.NoError(t, err)
	c := cc.(kam.KamClient)

	body, err := testClient.getRequests("kam", "ListKamReleases")
	assert.NoError(t, err)

	type ListKamReleasesRequestInfo struct {
		ContainerId string
		Request     kam.ListKamReleasesRequest
	}

	var requests []ListKamReleasesRequestInfo
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
				r := req.(*kam.ListKamReleasesRequest)
				return c.ListKamReleases(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]kam.ListKamReleasesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(kam.ListKamReleasesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="kam-service-support_ww_grp@oracle.com" jiraProject="CNP" opsJiraProject="SMESH"
func TestKamClientListWorkRequestErrors(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("kam", "ListWorkRequestErrors")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestErrors is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("kam", "Kam", "ListWorkRequestErrors", createKamClientWithProvider)
	assert.NoError(t, err)
	c := cc.(kam.KamClient)

	body, err := testClient.getRequests("kam", "ListWorkRequestErrors")
	assert.NoError(t, err)

	type ListWorkRequestErrorsRequestInfo struct {
		ContainerId string
		Request     kam.ListWorkRequestErrorsRequest
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
				r := req.(*kam.ListWorkRequestErrorsRequest)
				return c.ListWorkRequestErrors(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]kam.ListWorkRequestErrorsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(kam.ListWorkRequestErrorsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="kam-service-support_ww_grp@oracle.com" jiraProject="CNP" opsJiraProject="SMESH"
func TestKamClientListWorkRequestLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("kam", "ListWorkRequestLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequestLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("kam", "Kam", "ListWorkRequestLogs", createKamClientWithProvider)
	assert.NoError(t, err)
	c := cc.(kam.KamClient)

	body, err := testClient.getRequests("kam", "ListWorkRequestLogs")
	assert.NoError(t, err)

	type ListWorkRequestLogsRequestInfo struct {
		ContainerId string
		Request     kam.ListWorkRequestLogsRequest
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
				r := req.(*kam.ListWorkRequestLogsRequest)
				return c.ListWorkRequestLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]kam.ListWorkRequestLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(kam.ListWorkRequestLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="kam-service-support_ww_grp@oracle.com" jiraProject="CNP" opsJiraProject="SMESH"
func TestKamClientListWorkRequests(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("kam", "ListWorkRequests")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListWorkRequests is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("kam", "Kam", "ListWorkRequests", createKamClientWithProvider)
	assert.NoError(t, err)
	c := cc.(kam.KamClient)

	body, err := testClient.getRequests("kam", "ListWorkRequests")
	assert.NoError(t, err)

	type ListWorkRequestsRequestInfo struct {
		ContainerId string
		Request     kam.ListWorkRequestsRequest
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
				r := req.(*kam.ListWorkRequestsRequest)
				return c.ListWorkRequests(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]kam.ListWorkRequestsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(kam.ListWorkRequestsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="kam-service-support_ww_grp@oracle.com" jiraProject="CNP" opsJiraProject="SMESH"
func TestKamClientUpdateKamRelease(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("kam", "UpdateKamRelease")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateKamRelease is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("kam", "Kam", "UpdateKamRelease", createKamClientWithProvider)
	assert.NoError(t, err)
	c := cc.(kam.KamClient)

	body, err := testClient.getRequests("kam", "UpdateKamRelease")
	assert.NoError(t, err)

	type UpdateKamReleaseRequestInfo struct {
		ContainerId string
		Request     kam.UpdateKamReleaseRequest
	}

	var requests []UpdateKamReleaseRequestInfo
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

			response, err := c.UpdateKamRelease(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
