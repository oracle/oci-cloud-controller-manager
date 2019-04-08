// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Marketplace Service API
//
// Manage applications in Oracle Cloud Infrastructure Marketplace.
//

package marketplace

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

//ListingClient a client for Listing
type ListingClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewListingClientWithConfigurationProvider Creates a new default Listing client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewListingClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client ListingClient, err error) {
	baseClient, err := common.NewClientWithConfig(configProvider)
	if err != nil {
		return
	}

	client = ListingClient{BaseClient: baseClient}
	client.BasePath = "20181001"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *ListingClient) SetRegion(region string) {
	client.Host = common.StringToRegion(region).EndpointForTemplate("marketplace", "https://marketplace.{region}.oci.{secondLevelDomain}")
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *ListingClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
	if ok, err := common.IsConfigurationProviderValid(configProvider); !ok {
		return err
	}

	// Error has been checked already
	region, _ := configProvider.Region()
	client.SetRegion(region)
	client.config = &configProvider
	return nil
}

// ConfigurationProvider the ConfigurationProvider used in this client, or null if none set
func (client *ListingClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// GetAgreement Returns an agreement for a listing package with a time based signature which can be used to
// accept the agreement.
func (client ListingClient) GetAgreement(ctx context.Context, request GetAgreementRequest) (response GetAgreementResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getAgreement, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetAgreementResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetAgreementResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetAgreementResponse")
	}
	return
}

// getAgreement implements the OCIOperation interface (enables retrying operations)
func (client ListingClient) getAgreement(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/listings/{listingId}/packages/{packageVersion}/agreements/{agreementId}")
	if err != nil {
		return nil, err
	}

	var response GetAgreementResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetListing Gets detailed information about a listing, including the listing's name, version, description, and
// resources.
func (client ListingClient) GetListing(ctx context.Context, request GetListingRequest) (response GetListingResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getListing, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetListingResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetListingResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetListingResponse")
	}
	return
}

// getListing implements the OCIOperation interface (enables retrying operations)
func (client ListingClient) getListing(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/listings/{listingId}")
	if err != nil {
		return nil, err
	}

	var response GetListingResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetPackage Get the details of a package. This includes the fields needed to launch the package.
func (client ListingClient) GetPackage(ctx context.Context, request GetPackageRequest) (response GetPackageResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getPackage, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetPackageResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetPackageResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetPackageResponse")
	}
	return
}

// getPackage implements the OCIOperation interface (enables retrying operations)
func (client ListingClient) getPackage(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/listings/{listingId}/packages/{packageVersion}")
	if err != nil {
		return nil, err
	}

	var response GetPackageResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponseWithPolymorphicBody(httpResponse, &response, &listingpackage{})
	return response, err
}

// ListAgreements Returns the agreements that must be accepted to deploy this version of the listing package.
func (client ListingClient) ListAgreements(ctx context.Context, request ListAgreementsRequest) (response ListAgreementsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listAgreements, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListAgreementsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListAgreementsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListAgreementsResponse")
	}
	return
}

// listAgreements implements the OCIOperation interface (enables retrying operations)
func (client ListingClient) listAgreements(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/listings/{listingId}/packages/{packageVersion}/agreements")
	if err != nil {
		return nil, err
	}

	var response ListAgreementsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListListings Gets a list of listings from Oracle Cloud Infrastructure Marketplace by searching keywords and
// filtering according to listing attributes.
func (client ListingClient) ListListings(ctx context.Context, request ListListingsRequest) (response ListListingsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listListings, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListListingsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListListingsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListListingsResponse")
	}
	return
}

// listListings implements the OCIOperation interface (enables retrying operations)
func (client ListingClient) listListings(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/listings")
	if err != nil {
		return nil, err
	}

	var response ListListingsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListPackages Gets the list of packages for a listing.
func (client ListingClient) ListPackages(ctx context.Context, request ListPackagesRequest) (response ListPackagesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listPackages, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListPackagesResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListPackagesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListPackagesResponse")
	}
	return
}

// listPackages implements the OCIOperation interface (enables retrying operations)
func (client ListingClient) listPackages(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/listings/{listingId}/packages")
	if err != nil {
		return nil, err
	}

	var response ListPackagesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}
