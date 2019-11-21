package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/events"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createEventsClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := events.NewEventsClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="oci_events_dev_grp@oracle.com" jiraProject="https://jira.oci.oraclecorp.com/projects/CLEV" opsJiraProject="https://jira-sd.mc1.oracleiaas.com/projects/CLEV"
func TestEventsClientChangeRuleCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("events", "ChangeRuleCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeRuleCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("events", "Events", "ChangeRuleCompartment", createEventsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(events.EventsClient)

	body, err := testClient.getRequests("events", "ChangeRuleCompartment")
	assert.NoError(t, err)

	type ChangeRuleCompartmentRequestInfo struct {
		ContainerId string
		Request     events.ChangeRuleCompartmentRequest
	}

	var requests []ChangeRuleCompartmentRequestInfo
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

			response, err := c.ChangeRuleCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_events_dev_grp@oracle.com" jiraProject="https://jira.oci.oraclecorp.com/projects/CLEV" opsJiraProject="https://jira-sd.mc1.oracleiaas.com/projects/CLEV"
func TestEventsClientCreateRule(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("events", "CreateRule")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateRule is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("events", "Events", "CreateRule", createEventsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(events.EventsClient)

	body, err := testClient.getRequests("events", "CreateRule")
	assert.NoError(t, err)

	type CreateRuleRequestInfo struct {
		ContainerId string
		Request     events.CreateRuleRequest
	}

	var requests []CreateRuleRequestInfo
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

			response, err := c.CreateRule(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_events_dev_grp@oracle.com" jiraProject="https://jira.oci.oraclecorp.com/projects/CLEV" opsJiraProject="https://jira-sd.mc1.oracleiaas.com/projects/CLEV"
func TestEventsClientDeleteRule(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("events", "DeleteRule")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteRule is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("events", "Events", "DeleteRule", createEventsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(events.EventsClient)

	body, err := testClient.getRequests("events", "DeleteRule")
	assert.NoError(t, err)

	type DeleteRuleRequestInfo struct {
		ContainerId string
		Request     events.DeleteRuleRequest
	}

	var requests []DeleteRuleRequestInfo
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

			response, err := c.DeleteRule(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_events_dev_grp@oracle.com" jiraProject="https://jira.oci.oraclecorp.com/projects/CLEV" opsJiraProject="https://jira-sd.mc1.oracleiaas.com/projects/CLEV"
func TestEventsClientGetRule(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("events", "GetRule")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetRule is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("events", "Events", "GetRule", createEventsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(events.EventsClient)

	body, err := testClient.getRequests("events", "GetRule")
	assert.NoError(t, err)

	type GetRuleRequestInfo struct {
		ContainerId string
		Request     events.GetRuleRequest
	}

	var requests []GetRuleRequestInfo
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

			response, err := c.GetRule(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_events_dev_grp@oracle.com" jiraProject="https://jira.oci.oraclecorp.com/projects/CLEV" opsJiraProject="https://jira-sd.mc1.oracleiaas.com/projects/CLEV"
func TestEventsClientListRules(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("events", "ListRules")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListRules is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("events", "Events", "ListRules", createEventsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(events.EventsClient)

	body, err := testClient.getRequests("events", "ListRules")
	assert.NoError(t, err)

	type ListRulesRequestInfo struct {
		ContainerId string
		Request     events.ListRulesRequest
	}

	var requests []ListRulesRequestInfo
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
				r := req.(*events.ListRulesRequest)
				return c.ListRules(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]events.ListRulesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(events.ListRulesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_events_dev_grp@oracle.com" jiraProject="https://jira.oci.oraclecorp.com/projects/CLEV" opsJiraProject="https://jira-sd.mc1.oracleiaas.com/projects/CLEV"
func TestEventsClientUpdateRule(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("events", "UpdateRule")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateRule is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("events", "Events", "UpdateRule", createEventsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(events.EventsClient)

	body, err := testClient.getRequests("events", "UpdateRule")
	assert.NoError(t, err)

	type UpdateRuleRequestInfo struct {
		ContainerId string
		Request     events.UpdateRuleRequest
	}

	var requests []UpdateRuleRequestInfo
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

			response, err := c.UpdateRule(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
