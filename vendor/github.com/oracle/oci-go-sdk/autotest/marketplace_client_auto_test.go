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

func createMarketplaceClientWithProvider(p common.ConfigurationProvider, testConfig TestingConfig) (interface{}, error) {

	client, err := marketplace.NewMarketplaceClientWithConfigurationProvider(p)
	if testConfig.Endpoint != "" {
		client.Host = testConfig.Endpoint
	} else {
		client.SetRegion(testConfig.Region)
	}
	return client, err
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientCreateAcceptedAgreement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "CreateAcceptedAgreement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("CreateAcceptedAgreement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "CreateAcceptedAgreement", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

	body, err := testClient.getRequests("marketplace", "CreateAcceptedAgreement")
	assert.NoError(t, err)

	type CreateAcceptedAgreementRequestInfo struct {
		ContainerId string
		Request     marketplace.CreateAcceptedAgreementRequest
	}

	var requests []CreateAcceptedAgreementRequestInfo
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

			response, err := c.CreateAcceptedAgreement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientDeleteAcceptedAgreement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "DeleteAcceptedAgreement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("DeleteAcceptedAgreement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "DeleteAcceptedAgreement", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

	body, err := testClient.getRequests("marketplace", "DeleteAcceptedAgreement")
	assert.NoError(t, err)

	type DeleteAcceptedAgreementRequestInfo struct {
		ContainerId string
		Request     marketplace.DeleteAcceptedAgreementRequest
	}

	var requests []DeleteAcceptedAgreementRequestInfo
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

			response, err := c.DeleteAcceptedAgreement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientGetAcceptedAgreement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "GetAcceptedAgreement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetAcceptedAgreement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "GetAcceptedAgreement", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

	body, err := testClient.getRequests("marketplace", "GetAcceptedAgreement")
	assert.NoError(t, err)

	type GetAcceptedAgreementRequestInfo struct {
		ContainerId string
		Request     marketplace.GetAcceptedAgreementRequest
	}

	var requests []GetAcceptedAgreementRequestInfo
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

			response, err := c.GetAcceptedAgreement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientGetAgreement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "GetAgreement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetAgreement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "GetAgreement", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

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
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetAgreement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientGetListing(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "GetListing")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetListing is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "GetListing", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

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
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetListing(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientGetPackage(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "GetPackage")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("GetPackage is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "GetPackage", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

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
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
			req.Request.RequestMetadata.RetryPolicy = retryPolicy

			response, err := c.GetPackage(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientListAcceptedAgreements(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "ListAcceptedAgreements")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListAcceptedAgreements is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "ListAcceptedAgreements", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

	body, err := testClient.getRequests("marketplace", "ListAcceptedAgreements")
	assert.NoError(t, err)

	type ListAcceptedAgreementsRequestInfo struct {
		ContainerId string
		Request     marketplace.ListAcceptedAgreementsRequest
	}

	var requests []ListAcceptedAgreementsRequestInfo
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
				r := req.(*marketplace.ListAcceptedAgreementsRequest)
				return c.ListAcceptedAgreements(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]marketplace.ListAcceptedAgreementsResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(marketplace.ListAcceptedAgreementsResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientListAgreements(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "ListAgreements")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListAgreements is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "ListAgreements", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

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
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
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
func TestMarketplaceClientListCategories(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "ListCategories")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListCategories is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "ListCategories", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

	body, err := testClient.getRequests("marketplace", "ListCategories")
	assert.NoError(t, err)

	type ListCategoriesRequestInfo struct {
		ContainerId string
		Request     marketplace.ListCategoriesRequest
	}

	var requests []ListCategoriesRequestInfo
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
				r := req.(*marketplace.ListCategoriesRequest)
				return c.ListCategories(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]marketplace.ListCategoriesResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(marketplace.ListCategoriesResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientListListings(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "ListListings")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListListings is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "ListListings", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

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
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
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
func TestMarketplaceClientListPackages(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "ListPackages")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListPackages is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "ListPackages", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

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
			if withRetry == true {
				retryPolicy = retryPolicyForTests()
			}
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

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientListPublishers(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "ListPublishers")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("ListPublishers is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "ListPublishers", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

	body, err := testClient.getRequests("marketplace", "ListPublishers")
	assert.NoError(t, err)

	type ListPublishersRequestInfo struct {
		ContainerId string
		Request     marketplace.ListPublishersRequest
	}

	var requests []ListPublishersRequestInfo
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
				r := req.(*marketplace.ListPublishersRequest)
				return c.ListPublishers(context.Background(), *r)
			}

			listResponses, err := testClient.generateListResponses(&request.Request, listFn)
			typedListResponses := make([]marketplace.ListPublishersResponse, len(listResponses))
			for i, lr := range listResponses {
				typedListResponses[i] = lr.(marketplace.ListPublishersResponse)
			}

			message, err := testClient.validateResult(request.ContainerId, request.Request, typedListResponses, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}

// IssueRoutingInfo tag="default" email="oci_marketplace_seattle_us_grp@oracle.com" jiraProject="MAR" opsJiraProject="CMP"
func TestMarketplaceClientUpdateAcceptedAgreement(t *testing.T) {
	defer failTestOnPanic(t)

	enabled, err := testClient.isApiEnabled("marketplace", "UpdateAcceptedAgreement")
	assert.NoError(t, err)
	if !enabled {
		t.Skip("UpdateAcceptedAgreement is not enabled by the testing service")
	}

	cc, err := testClient.createClientForOperation("marketplace", "Marketplace", "UpdateAcceptedAgreement", createMarketplaceClientWithProvider)
	assert.NoError(t, err)
	c := cc.(marketplace.MarketplaceClient)

	body, err := testClient.getRequests("marketplace", "UpdateAcceptedAgreement")
	assert.NoError(t, err)

	type UpdateAcceptedAgreementRequestInfo struct {
		ContainerId string
		Request     marketplace.UpdateAcceptedAgreementRequest
	}

	var requests []UpdateAcceptedAgreementRequestInfo
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

			response, err := c.UpdateAcceptedAgreement(context.Background(), req.Request)
			message, err := testClient.validateResult(req.ContainerId, req.Request, response, err)
			assert.NoError(t, err)
			assert.Empty(t, message, message)
		})
	}
}
