package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/functions"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createFunctionsInvokeClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {
	client, err := functions.NewFunctionsInvokeClientWithConfigurationProvider(p, testConfig.Endpoint)
	return client, err
}

// IssueRoutingInfo tag="default" email="serverless_grp@oracle.com" jiraProject="FAAS" opsJiraProject="FAAS"
func TestFunctionsInvokeClientInvokeFunction(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("functions", "InvokeFunction")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("InvokeFunction is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("functions", "FunctionsInvoke", "InvokeFunction", createFunctionsInvokeClientWithProvider)
	assert.NoError(t, err)
	c := cc.(functions.FunctionsInvokeClient)

	body, err := testClient.getRequests("functions", "InvokeFunction")
	assert.NoError(t, err)

	type InvokeFunctionRequestInfo struct {
		ContainerId string
		Request     functions.InvokeFunctionRequest
	}

	var requests []InvokeFunctionRequestInfo
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

			response, err := c.InvokeFunction(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
