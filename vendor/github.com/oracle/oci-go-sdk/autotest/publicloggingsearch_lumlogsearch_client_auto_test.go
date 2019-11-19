package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/publicloggingsearch"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createLumLogSearchClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := publicloggingsearch.NewLumLogSearchClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="hydra_dev_us_grp@oracle.com" jiraProject="HYD" opsJiraProject="HYD"
func TestLumLogSearchClientSearchLogs(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("publicloggingsearch", "SearchLogs")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("SearchLogs is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("publicloggingsearch", "LumLogSearch", "SearchLogs", createLumLogSearchClientWithProvider)
	assert.NoError(t, err)
	c := cc.(publicloggingsearch.LumLogSearchClient)

	body, err := testClient.getRequests("publicloggingsearch", "SearchLogs")
	assert.NoError(t, err)

	type SearchLogsRequestInfo struct {
		ContainerId string
		Request     publicloggingsearch.SearchLogsRequest
	}

	var requests []SearchLogsRequestInfo
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
				r := req.(*publicloggingsearch.SearchLogsRequest)
				return c.SearchLogs(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]publicloggingsearch.SearchLogsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(publicloggingsearch.SearchLogsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
