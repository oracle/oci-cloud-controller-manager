package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/publiclogging"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createLumLoggingClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := publiclogging.NewLumLoggingClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientChangeLogGroupCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "ChangeLogGroupCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeLogGroupCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "ChangeLogGroupCompartment", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "ChangeLogGroupCompartment")
	assert.NoError(t, err)

	type ChangeLogGroupCompartmentRequestInfo struct {
		ContainerId string
		Request     publiclogging.ChangeLogGroupCompartmentRequest
	}

	var requests []ChangeLogGroupCompartmentRequestInfo
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

			response, err := c.ChangeLogGroupCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientChangeLogLogGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "ChangeLogLogGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeLogLogGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "ChangeLogLogGroup", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "ChangeLogLogGroup")
	assert.NoError(t, err)

	type ChangeLogLogGroupRequestInfo struct {
		ContainerId string
		Request     publiclogging.ChangeLogLogGroupRequest
	}

	var requests []ChangeLogLogGroupRequestInfo
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

			response, err := c.ChangeLogLogGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientCreateLog(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "CreateLog")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateLog is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "CreateLog", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "CreateLog")
	assert.NoError(t, err)

	type CreateLogRequestInfo struct {
		ContainerId string
		Request     publiclogging.CreateLogRequest
	}

	var requests []CreateLogRequestInfo
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

			response, err := c.CreateLog(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientCreateLogGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "CreateLogGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateLogGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "CreateLogGroup", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "CreateLogGroup")
	assert.NoError(t, err)

	type CreateLogGroupRequestInfo struct {
		ContainerId string
		Request     publiclogging.CreateLogGroupRequest
	}

	var requests []CreateLogGroupRequestInfo
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

			response, err := c.CreateLogGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientDeleteLog(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "DeleteLog")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteLog is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "DeleteLog", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "DeleteLog")
	assert.NoError(t, err)

	type DeleteLogRequestInfo struct {
		ContainerId string
		Request     publiclogging.DeleteLogRequest
	}

	var requests []DeleteLogRequestInfo
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

			response, err := c.DeleteLog(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientDeleteLogGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "DeleteLogGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteLogGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "DeleteLogGroup", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "DeleteLogGroup")
	assert.NoError(t, err)

	type DeleteLogGroupRequestInfo struct {
		ContainerId string
		Request     publiclogging.DeleteLogGroupRequest
	}

	var requests []DeleteLogGroupRequestInfo
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

			response, err := c.DeleteLogGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientGetLog(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "GetLog")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetLog is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "GetLog", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "GetLog")
	assert.NoError(t, err)

	type GetLogRequestInfo struct {
		ContainerId string
		Request     publiclogging.GetLogRequest
	}

	var requests []GetLogRequestInfo
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

			response, err := c.GetLog(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientGetLogGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "GetLogGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetLogGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "GetLogGroup", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "GetLogGroup")
	assert.NoError(t, err)

	type GetLogGroupRequestInfo struct {
		ContainerId string
		Request     publiclogging.GetLogGroupRequest
	}

	var requests []GetLogGroupRequestInfo
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

			response, err := c.GetLogGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientListLogGroups(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "ListLogGroups")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListLogGroups is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "ListLogGroups", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "ListLogGroups")
	assert.NoError(t, err)

	type ListLogGroupsRequestInfo struct {
		ContainerId string
		Request     publiclogging.ListLogGroupsRequest
	}

	var requests []ListLogGroupsRequestInfo
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
				r := req.(*publiclogging.ListLogGroupsRequest)
				return c.ListLogGroups(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]publiclogging.ListLogGroupsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(publiclogging.ListLogGroupsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientListLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "ListLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "ListLogs", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "ListLogs")
	assert.NoError(t, err)

	type ListLogsRequestInfo struct {
		ContainerId string
		Request     publiclogging.ListLogsRequest
	}

	var requests []ListLogsRequestInfo
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
				r := req.(*publiclogging.ListLogsRequest)
				return c.ListLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]publiclogging.ListLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(publiclogging.ListLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientUpdateLog(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "UpdateLog")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateLog is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "UpdateLog", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "UpdateLog")
	assert.NoError(t, err)

	type UpdateLogRequestInfo struct {
		ContainerId string
		Request     publiclogging.UpdateLogRequest
	}

	var requests []UpdateLogRequestInfo
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

			response, err := c.UpdateLog(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLoggingClientUpdateLogGroup(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publiclogging", "UpdateLogGroup")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateLogGroup is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publiclogging", "LumLogging", "UpdateLogGroup", createLumLoggingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publiclogging.LumLoggingClient)

	body, err := testClient.getRequests("publiclogging", "UpdateLogGroup")
	assert.NoError(t, err)

	type UpdateLogGroupRequestInfo struct {
		ContainerId string
		Request     publiclogging.UpdateLogGroupRequest
	}

	var requests []UpdateLogGroupRequestInfo
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

			response, err := c.UpdateLogGroup(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
