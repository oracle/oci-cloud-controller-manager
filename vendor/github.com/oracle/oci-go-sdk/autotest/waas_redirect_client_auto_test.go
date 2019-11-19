package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/waas"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createRedirectClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := waas.NewRedirectClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="oci_waas_dev_us_grp@oracle.com" jiraProject="WAAS" opsJiraProject="WAF"
func TestRedirectClientChangeHttpRedirectCompartment(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("waas", "ChangeHttpRedirectCompartment")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ChangeHttpRedirectCompartment is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("waas", "Redirect", "ChangeHttpRedirectCompartment", createRedirectClientWithProvider)
	assert.NoError(t, err)
	c := cc.(waas.RedirectClient)

	body, err := testClient.getRequests("waas", "ChangeHttpRedirectCompartment")
	assert.NoError(t, err)

	type ChangeHttpRedirectCompartmentRequestInfo struct {
		ContainerId string
		Request     waas.ChangeHttpRedirectCompartmentRequest
	}

	var requests []ChangeHttpRedirectCompartmentRequestInfo
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

			response, err := c.ChangeHttpRedirectCompartment(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_waas_dev_us_grp@oracle.com" jiraProject="WAAS" opsJiraProject="WAF"
func TestRedirectClientCreateHttpRedirect(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("waas", "CreateHttpRedirect")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateHttpRedirect is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("waas", "Redirect", "CreateHttpRedirect", createRedirectClientWithProvider)
	assert.NoError(t, err)
	c := cc.(waas.RedirectClient)

	body, err := testClient.getRequests("waas", "CreateHttpRedirect")
	assert.NoError(t, err)

	type CreateHttpRedirectRequestInfo struct {
		ContainerId string
		Request     waas.CreateHttpRedirectRequest
	}

	var requests []CreateHttpRedirectRequestInfo
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

			response, err := c.CreateHttpRedirect(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_waas_dev_us_grp@oracle.com" jiraProject="WAAS" opsJiraProject="WAF"
func TestRedirectClientDeleteHttpRedirect(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("waas", "DeleteHttpRedirect")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteHttpRedirect is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("waas", "Redirect", "DeleteHttpRedirect", createRedirectClientWithProvider)
	assert.NoError(t, err)
	c := cc.(waas.RedirectClient)

	body, err := testClient.getRequests("waas", "DeleteHttpRedirect")
	assert.NoError(t, err)

	type DeleteHttpRedirectRequestInfo struct {
		ContainerId string
		Request     waas.DeleteHttpRedirectRequest
	}

	var requests []DeleteHttpRedirectRequestInfo
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

			response, err := c.DeleteHttpRedirect(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_waas_dev_us_grp@oracle.com" jiraProject="WAAS" opsJiraProject="WAF"
func TestRedirectClientGetHttpRedirect(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("waas", "GetHttpRedirect")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetHttpRedirect is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("waas", "Redirect", "GetHttpRedirect", createRedirectClientWithProvider)
	assert.NoError(t, err)
	c := cc.(waas.RedirectClient)

	body, err := testClient.getRequests("waas", "GetHttpRedirect")
	assert.NoError(t, err)

	type GetHttpRedirectRequestInfo struct {
		ContainerId string
		Request     waas.GetHttpRedirectRequest
	}

	var requests []GetHttpRedirectRequestInfo
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

			response, err := c.GetHttpRedirect(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_waas_dev_us_grp@oracle.com" jiraProject="WAAS" opsJiraProject="WAF"
func TestRedirectClientListHttpRedirects(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("waas", "ListHttpRedirects")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListHttpRedirects is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("waas", "Redirect", "ListHttpRedirects", createRedirectClientWithProvider)
	assert.NoError(t, err)
	c := cc.(waas.RedirectClient)

	body, err := testClient.getRequests("waas", "ListHttpRedirects")
	assert.NoError(t, err)

	type ListHttpRedirectsRequestInfo struct {
		ContainerId string
		Request     waas.ListHttpRedirectsRequest
	}

	var requests []ListHttpRedirectsRequestInfo
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
				r := req.(*waas.ListHttpRedirectsRequest)
				return c.ListHttpRedirects(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]waas.ListHttpRedirectsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(waas.ListHttpRedirectsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_waas_dev_us_grp@oracle.com" jiraProject="WAAS" opsJiraProject="WAF"
func TestRedirectClientUpdateHttpRedirect(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("waas", "UpdateHttpRedirect")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateHttpRedirect is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("waas", "Redirect", "UpdateHttpRedirect", createRedirectClientWithProvider)
	assert.NoError(t, err)
	c := cc.(waas.RedirectClient)

	body, err := testClient.getRequests("waas", "UpdateHttpRedirect")
	assert.NoError(t, err)

	type UpdateHttpRedirectRequestInfo struct {
		ContainerId string
		Request     waas.UpdateHttpRedirectRequest
	}

	var requests []UpdateHttpRedirectRequestInfo
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

			response, err := c.UpdateHttpRedirect(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
