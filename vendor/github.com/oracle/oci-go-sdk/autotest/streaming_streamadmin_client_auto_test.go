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

func createStreamAdminClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := streaming.NewStreamAdminClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientChangeStreamCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "ChangeStreamCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeStreamCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "ChangeStreamCompartment", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "ChangeStreamCompartment")
	assert.NoError(t, err)

	type ChangeStreamCompartmentRequestInfo struct {
		ContainerId string
		Request     streaming.ChangeStreamCompartmentRequest
	}

	var requests []ChangeStreamCompartmentRequestInfo
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

			response, err := c.ChangeStreamCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientCreateArchiver(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "CreateArchiver")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateArchiver is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "CreateArchiver", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "CreateArchiver")
	assert.NoError(t, err)

	type CreateArchiverRequestInfo struct {
		ContainerId string
		Request     streaming.CreateArchiverRequest
	}

	var requests []CreateArchiverRequestInfo
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

			response, err := c.CreateArchiver(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientCreateStream(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "CreateStream")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateStream is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "CreateStream", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "CreateStream")
	assert.NoError(t, err)

	type CreateStreamRequestInfo struct {
		ContainerId string
		Request     streaming.CreateStreamRequest
	}

	var requests []CreateStreamRequestInfo
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

			response, err := c.CreateStream(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientDeleteStream(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "DeleteStream")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteStream is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "DeleteStream", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "DeleteStream")
	assert.NoError(t, err)

	type DeleteStreamRequestInfo struct {
		ContainerId string
		Request     streaming.DeleteStreamRequest
	}

	var requests []DeleteStreamRequestInfo
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

			response, err := c.DeleteStream(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientGetArchiver(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "GetArchiver")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetArchiver is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "GetArchiver", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "GetArchiver")
	assert.NoError(t, err)

	type GetArchiverRequestInfo struct {
		ContainerId string
		Request     streaming.GetArchiverRequest
	}

	var requests []GetArchiverRequestInfo
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

			response, err := c.GetArchiver(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientGetDefaultStreamPool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "GetDefaultStreamPool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetDefaultStreamPool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "GetDefaultStreamPool", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "GetDefaultStreamPool")
	assert.NoError(t, err)

	type GetDefaultStreamPoolRequestInfo struct {
		ContainerId string
		Request     streaming.GetDefaultStreamPoolRequest
	}

	var requests []GetDefaultStreamPoolRequestInfo
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

			response, err := c.GetDefaultStreamPool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientGetStream(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "GetStream")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetStream is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "GetStream", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "GetStream")
	assert.NoError(t, err)

	type GetStreamRequestInfo struct {
		ContainerId string
		Request     streaming.GetStreamRequest
	}

	var requests []GetStreamRequestInfo
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

			response, err := c.GetStream(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientListDefaultStreamPool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "ListDefaultStreamPool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListDefaultStreamPool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "ListDefaultStreamPool", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "ListDefaultStreamPool")
	assert.NoError(t, err)

	type ListDefaultStreamPoolRequestInfo struct {
		ContainerId string
		Request     streaming.ListDefaultStreamPoolRequest
	}

	var requests []ListDefaultStreamPoolRequestInfo
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

			response, err := c.ListDefaultStreamPool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientListStreams(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "ListStreams")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListStreams is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "ListStreams", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "ListStreams")
	assert.NoError(t, err)

	type ListStreamsRequestInfo struct {
		ContainerId string
		Request     streaming.ListStreamsRequest
	}

	var requests []ListStreamsRequestInfo
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
				r := req.(*streaming.ListStreamsRequest)
				return c.ListStreams(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]streaming.ListStreamsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(streaming.ListStreamsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientStartArchiver(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "StartArchiver")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("StartArchiver is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "StartArchiver", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "StartArchiver")
	assert.NoError(t, err)

	type StartArchiverRequestInfo struct {
		ContainerId string
		Request     streaming.StartArchiverRequest
	}

	var requests []StartArchiverRequestInfo
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

			response, err := c.StartArchiver(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientStopArchiver(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "StopArchiver")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("StopArchiver is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "StopArchiver", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "StopArchiver")
	assert.NoError(t, err)

	type StopArchiverRequestInfo struct {
		ContainerId string
		Request     streaming.StopArchiverRequest
	}

	var requests []StopArchiverRequestInfo
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

			response, err := c.StopArchiver(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientUpdateArchiver(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "UpdateArchiver")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateArchiver is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "UpdateArchiver", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "UpdateArchiver")
	assert.NoError(t, err)

	type UpdateArchiverRequestInfo struct {
		ContainerId string
		Request     streaming.UpdateArchiverRequest
	}

	var requests []UpdateArchiverRequestInfo
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

			response, err := c.UpdateArchiver(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientUpdateDefaultStreamPool(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "UpdateDefaultStreamPool")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateDefaultStreamPool is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "UpdateDefaultStreamPool", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "UpdateDefaultStreamPool")
	assert.NoError(t, err)

	type UpdateDefaultStreamPoolRequestInfo struct {
		ContainerId string
		Request     streaming.UpdateDefaultStreamPoolRequest
	}

	var requests []UpdateDefaultStreamPoolRequestInfo
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

			response, err := c.UpdateDefaultStreamPool(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="opc_streaming_us_grp@oracle.com" jiraProject="STREAMSTR" opsJiraProject="STREAMOSS"
func TestStreamAdminClientUpdateStream(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("streaming", "UpdateStream")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateStream is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("streaming", "StreamAdmin", "UpdateStream", createStreamAdminClientWithProvider)
	assert.NoError(t, err)
	c := cc.(streaming.StreamAdminClient)

	body, err := testClient.getRequests("streaming", "UpdateStream")
	assert.NoError(t, err)

	type UpdateStreamRequestInfo struct {
		ContainerId string
		Request     streaming.UpdateStreamRequest
	}

	var requests []UpdateStreamRequestInfo
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

			response, err := c.UpdateStream(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
