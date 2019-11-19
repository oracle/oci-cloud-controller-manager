package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/streaming"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createConnectHarnessAdminClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := streaming.NewConnectHarnessAdminClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestConnectHarnessAdminClientChangeConnectHarnessCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "ChangeConnectHarnessCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeConnectHarnessCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "ConnectHarnessAdmin", "ChangeConnectHarnessCompartment", createConnectHarnessAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.ConnectHarnessAdminClient)

	body, err := testClient.getRequests("streaming", "ChangeConnectHarnessCompartment")
	assert.NoError(t, err)

	type ChangeConnectHarnessCompartmentRequestInfo struct {
		ContainerId string
		Request     streaming.ChangeConnectHarnessCompartmentRequest
	}

	var requests []ChangeConnectHarnessCompartmentRequestInfo
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

			response, err := c.ChangeConnectHarnessCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestConnectHarnessAdminClientCreateConnectHarness(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "CreateConnectHarness")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateConnectHarness is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "ConnectHarnessAdmin", "CreateConnectHarness", createConnectHarnessAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.ConnectHarnessAdminClient)

	body, err := testClient.getRequests("streaming", "CreateConnectHarness")
	assert.NoError(t, err)

	type CreateConnectHarnessRequestInfo struct {
		ContainerId string
		Request     streaming.CreateConnectHarnessRequest
	}

	var requests []CreateConnectHarnessRequestInfo
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

			response, err := c.CreateConnectHarness(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestConnectHarnessAdminClientDeleteConnectHarness(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "DeleteConnectHarness")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteConnectHarness is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "ConnectHarnessAdmin", "DeleteConnectHarness", createConnectHarnessAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.ConnectHarnessAdminClient)

	body, err := testClient.getRequests("streaming", "DeleteConnectHarness")
	assert.NoError(t, err)

	type DeleteConnectHarnessRequestInfo struct {
		ContainerId string
		Request     streaming.DeleteConnectHarnessRequest
	}

	var requests []DeleteConnectHarnessRequestInfo
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

			response, err := c.DeleteConnectHarness(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestConnectHarnessAdminClientGetConnectHarness(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "GetConnectHarness")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetConnectHarness is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "ConnectHarnessAdmin", "GetConnectHarness", createConnectHarnessAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.ConnectHarnessAdminClient)

	body, err := testClient.getRequests("streaming", "GetConnectHarness")
	assert.NoError(t, err)

	type GetConnectHarnessRequestInfo struct {
		ContainerId string
		Request     streaming.GetConnectHarnessRequest
	}

	var requests []GetConnectHarnessRequestInfo
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

			response, err := c.GetConnectHarness(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestConnectHarnessAdminClientListConnectHarnesses(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "ListConnectHarnesses")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListConnectHarnesses is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "ConnectHarnessAdmin", "ListConnectHarnesses", createConnectHarnessAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.ConnectHarnessAdminClient)

	body, err := testClient.getRequests("streaming", "ListConnectHarnesses")
	assert.NoError(t, err)

	type ListConnectHarnessesRequestInfo struct {
		ContainerId string
		Request     streaming.ListConnectHarnessesRequest
	}

	var requests []ListConnectHarnessesRequestInfo
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
				r := req.(*streaming.ListConnectHarnessesRequest)
				return c.ListConnectHarnesses(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]streaming.ListConnectHarnessesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(streaming.ListConnectHarnessesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestConnectHarnessAdminClientUpdateConnectHarness(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "UpdateConnectHarness")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateConnectHarness is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "ConnectHarnessAdmin", "UpdateConnectHarness", createConnectHarnessAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.ConnectHarnessAdminClient)

	body, err := testClient.getRequests("streaming", "UpdateConnectHarness")
	assert.NoError(t, err)

	type UpdateConnectHarnessRequestInfo struct {
		ContainerId string
		Request     streaming.UpdateConnectHarnessRequest
	}

	var requests []UpdateConnectHarnessRequestInfo
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

			response, err := c.UpdateConnectHarness(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
