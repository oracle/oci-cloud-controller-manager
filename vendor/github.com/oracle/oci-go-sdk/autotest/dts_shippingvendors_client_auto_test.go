package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/dts"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createShippingVendorsClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := dts.NewShippingVendorsClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="data_transfer_platform_dev_ww_grp@oracle.com" jiraProject="BDTS" opsJiraProject="DTS"
func TestShippingVendorsClientListShippingVendors(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("dts", "ListShippingVendors")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListShippingVendors is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("dts", "ShippingVendors", "ListShippingVendors", createShippingVendorsClientWithProvider)
	assert.NoError(t, err)
	c := cc.(dts.ShippingVendorsClient)

	body, err := testClient.getRequests("dts", "ListShippingVendors")
	assert.NoError(t, err)

	type ListShippingVendorsRequestInfo struct {
		ContainerId string
		Request     dts.ListShippingVendorsRequest
	}

	var requests []ListShippingVendorsRequestInfo
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

			response, err := c.ListShippingVendors(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
