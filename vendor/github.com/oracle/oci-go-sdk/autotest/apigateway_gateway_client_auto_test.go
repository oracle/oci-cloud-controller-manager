package autotest

import (
	"github.com/oracle/oci-go-sdk/apigateway"
	"github.com/oracle/oci-go-sdk/common"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createGatewayClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := apigateway.NewGatewayClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestGatewayClientChangeGatewayCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "ChangeGatewayCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeGatewayCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Gateway", "ChangeGatewayCompartment", createGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.GatewayClient)

	body, err := testClient.getRequests("apigateway", "ChangeGatewayCompartment")
	assert.NoError(t, err)

	type ChangeGatewayCompartmentRequestInfo struct {
		ContainerId string
		Request     apigateway.ChangeGatewayCompartmentRequest
	}

	var requests []ChangeGatewayCompartmentRequestInfo
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

			response, err := c.ChangeGatewayCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestGatewayClientCreateGateway(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "CreateGateway")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateGateway is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Gateway", "CreateGateway", createGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.GatewayClient)

	body, err := testClient.getRequests("apigateway", "CreateGateway")
	assert.NoError(t, err)

	type CreateGatewayRequestInfo struct {
		ContainerId string
		Request     apigateway.CreateGatewayRequest
	}

	var requests []CreateGatewayRequestInfo
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

			response, err := c.CreateGateway(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestGatewayClientDeleteGateway(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "DeleteGateway")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteGateway is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Gateway", "DeleteGateway", createGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.GatewayClient)

	body, err := testClient.getRequests("apigateway", "DeleteGateway")
	assert.NoError(t, err)

	type DeleteGatewayRequestInfo struct {
		ContainerId string
		Request     apigateway.DeleteGatewayRequest
	}

	var requests []DeleteGatewayRequestInfo
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

			response, err := c.DeleteGateway(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestGatewayClientGetGateway(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "GetGateway")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetGateway is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Gateway", "GetGateway", createGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.GatewayClient)

	body, err := testClient.getRequests("apigateway", "GetGateway")
	assert.NoError(t, err)

	type GetGatewayRequestInfo struct {
		ContainerId string
		Request     apigateway.GetGatewayRequest
	}

	var requests []GetGatewayRequestInfo
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

			response, err := c.GetGateway(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestGatewayClientListGateways(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "ListGateways")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListGateways is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Gateway", "ListGateways", createGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.GatewayClient)

	body, err := testClient.getRequests("apigateway", "ListGateways")
	assert.NoError(t, err)

	type ListGatewaysRequestInfo struct {
		ContainerId string
		Request     apigateway.ListGatewaysRequest
	}

	var requests []ListGatewaysRequestInfo
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
				r := req.(*apigateway.ListGatewaysRequest)
				return c.ListGateways(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]apigateway.ListGatewaysResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(apigateway.ListGatewaysResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestGatewayClientUpdateGateway(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "UpdateGateway")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateGateway is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Gateway", "UpdateGateway", createGatewayClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.GatewayClient)

	body, err := testClient.getRequests("apigateway", "UpdateGateway")
	assert.NoError(t, err)

	type UpdateGatewayRequestInfo struct {
		ContainerId string
		Request     apigateway.UpdateGatewayRequest
	}

	var requests []UpdateGatewayRequestInfo
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

			response, err := c.UpdateGateway(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
