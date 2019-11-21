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

func createDeploymentClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := apigateway.NewDeploymentClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestDeploymentClientChangeDeploymentCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "ChangeDeploymentCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeDeploymentCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Deployment", "ChangeDeploymentCompartment", createDeploymentClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.DeploymentClient)

	body, err := testClient.getRequests("apigateway", "ChangeDeploymentCompartment")
	assert.NoError(t, err)

	type ChangeDeploymentCompartmentRequestInfo struct {
		ContainerId string
		Request     apigateway.ChangeDeploymentCompartmentRequest
	}

	var requests []ChangeDeploymentCompartmentRequestInfo
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

			response, err := c.ChangeDeploymentCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestDeploymentClientCreateDeployment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "CreateDeployment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateDeployment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Deployment", "CreateDeployment", createDeploymentClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.DeploymentClient)

	body, err := testClient.getRequests("apigateway", "CreateDeployment")
	assert.NoError(t, err)

	type CreateDeploymentRequestInfo struct {
		ContainerId string
		Request     apigateway.CreateDeploymentRequest
	}

	var requests []CreateDeploymentRequestInfo
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

			response, err := c.CreateDeployment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestDeploymentClientDeleteDeployment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "DeleteDeployment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteDeployment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Deployment", "DeleteDeployment", createDeploymentClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.DeploymentClient)

	body, err := testClient.getRequests("apigateway", "DeleteDeployment")
	assert.NoError(t, err)

	type DeleteDeploymentRequestInfo struct {
		ContainerId string
		Request     apigateway.DeleteDeploymentRequest
	}

	var requests []DeleteDeploymentRequestInfo
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

			response, err := c.DeleteDeployment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestDeploymentClientGetDeployment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "GetDeployment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetDeployment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Deployment", "GetDeployment", createDeploymentClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.DeploymentClient)

	body, err := testClient.getRequests("apigateway", "GetDeployment")
	assert.NoError(t, err)

	type GetDeploymentRequestInfo struct {
		ContainerId string
		Request     apigateway.GetDeploymentRequest
	}

	var requests []GetDeploymentRequestInfo
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

			response, err := c.GetDeployment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestDeploymentClientListDeployments(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "ListDeployments")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListDeployments is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Deployment", "ListDeployments", createDeploymentClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.DeploymentClient)

	body, err := testClient.getRequests("apigateway", "ListDeployments")
	assert.NoError(t, err)

	type ListDeploymentsRequestInfo struct {
		ContainerId string
		Request     apigateway.ListDeploymentsRequest
	}

	var requests []ListDeploymentsRequestInfo
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
				r := req.(*apigateway.ListDeploymentsRequest)
				return c.ListDeployments(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]apigateway.ListDeploymentsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(apigateway.ListDeploymentsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_apigw_ww_grp@oracle.com" jiraProject="APIGW" opsJiraProject="APIGW"
func TestDeploymentClientUpdateDeployment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("apigateway", "UpdateDeployment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateDeployment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("apigateway", "Deployment", "UpdateDeployment", createDeploymentClientWithProvider)
	assert.NoError(t, err)
	c := cc.(apigateway.DeploymentClient)

	body, err := testClient.getRequests("apigateway", "UpdateDeployment")
	assert.NoError(t, err)

	type UpdateDeploymentRequestInfo struct {
		ContainerId string
		Request     apigateway.UpdateDeploymentRequest
	}

	var requests []UpdateDeploymentRequestInfo
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

			response, err := c.UpdateDeployment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
