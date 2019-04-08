package autotest

import (
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/marketplace"

	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createListingClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := marketplace.NewListingClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestListingClientGetAgreement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "GetAgreement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetAgreement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Listing", "GetAgreement", createListingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.ListingClient)

	body, err := testClient.getRequests("marketplace", "GetAgreement")
	assert.NoError(t, err)

	type GetAgreementRequestInfo struct {
		ContainerId string
		Request     marketplace.GetAgreementRequest
	}

	var requests []GetAgreementRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetAgreement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestListingClientGetListing(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "GetListing")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetListing is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Listing", "GetListing", createListingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.ListingClient)

	body, err := testClient.getRequests("marketplace", "GetListing")
	assert.NoError(t, err)

	type GetListingRequestInfo struct {
		ContainerId string
		Request     marketplace.GetListingRequest
	}

	var requests []GetListingRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetListing(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestListingClientGetPackage(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "GetPackage")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetPackage is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Listing", "GetPackage", createListingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.ListingClient)

	body, err := testClient.getRequests("marketplace", "GetPackage")
	assert.NoError(t, err)

	type GetPackageRequestInfo struct {
		ContainerId string
		Request     marketplace.GetPackageRequest
	}

	var requests []GetPackageRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, req := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetPackage(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestListingClientListAgreements(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "ListAgreements")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListAgreements is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Listing", "ListAgreements", createListingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.ListingClient)

	body, err := testClient.getRequests("marketplace", "ListAgreements")
	assert.NoError(t, err)

	type ListAgreementsRequestInfo struct {
		ContainerId string
		Request     marketplace.ListAgreementsRequest
	}

	var requests []ListAgreementsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*marketplace.ListAgreementsRequest)
				return c.ListAgreements(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]marketplace.ListAgreementsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(marketplace.ListAgreementsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestListingClientListListings(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "ListListings")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListListings is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Listing", "ListListings", createListingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.ListingClient)

	body, err := testClient.getRequests("marketplace", "ListListings")
	assert.NoError(t, err)

	type ListListingsRequestInfo struct {
		ContainerId string
		Request     marketplace.ListListingsRequest
	}

	var requests []ListListingsRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*marketplace.ListListingsRequest)
				return c.ListListings(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]marketplace.ListListingsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(marketplace.ListListingsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestListingClientListPackages(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "ListPackages")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListPackages is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Listing", "ListPackages", createListingClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.ListingClient)

	body, err := testClient.getRequests("marketplace", "ListPackages")
	assert.NoError(t, err)

	type ListPackagesRequestInfo struct {
		ContainerId string
		Request     marketplace.ListPackagesRequest
	}

	var requests []ListPackagesRequestInfo
	var dataHolder []map[string]interface{}
	err = json.Unmarshal([]byte(body), &dataHolder)
	assert.NoError(t, err)
	err = unmarshalRequestInfo(dataHolder, &requests, testClient.Log)
	assert.NoError(t, err)

	var retryPolicy *common.RetryPolicy
	for i, request := range requests {
		t.Run(fmt.Sprintf("request:%v", i), func(t *testing.T) {
			retryPolicy = retryPolicyForTests()
			request.Request.RequestMetadata.RetryPolicy = retryPolicy
			listFn := func(req common.OCIRequest) (common.OCIResponse, error) {
				r := req.(*marketplace.ListPackagesRequest)
				return c.ListPackages(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]marketplace.ListPackagesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(marketplace.ListPackagesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
