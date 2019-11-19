// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Public Logging Search API
//
// A description of the Public Logging Search API
//

package publicloggingsearch

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

//LumLogSearchClient a client for LumLogSearch
type LumLogSearchClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewLumLogSearchClientWithConfigurationProvider Creates a new default LumLogSearch client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewLumLogSearchClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client LumLogSearchClient, err error) {
	baseClient, err := common.NewClientWithConfig(configProvider)
	if err != nil {
		return
	}

	client = LumLogSearchClient{BaseClient: baseClient}
	client.BasePath = "20190909"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *LumLogSearchClient) SetRegion(region string) {
	client.Host = common.StringToRegion(region).EndpointForTemplate("publicloggingsearch", "https://logging.{region}.oci.{secondLevelDomain}")
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *LumLogSearchClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
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
func (client *LumLogSearchClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// SearchLogs Submit a query to search logs.
func (client LumLogSearchClient) SearchLogs(ctx context.Context, request SearchLogsRequest) (response SearchLogsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.searchLogs, policy)
	if err != nil {
		if ociResponse != nil {
			response = SearchLogsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(SearchLogsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into SearchLogsResponse")
	}
	return
}

// searchLogs implements the OCIOperation interface (enables retrying operations)
func (client LumLogSearchClient) searchLogs(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/search")
	if err != nil {
		return nil, err
	}

	var response SearchLogsResponse
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
