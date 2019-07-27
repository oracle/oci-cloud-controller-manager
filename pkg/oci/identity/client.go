// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package identity

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// MetadataClient is a client for identity metadata service
type MetadataClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewMetadataClientWithConfigurationProvider creates a new metadata service client.
func NewMetadataClientWithConfigurationProvider(provider common.ConfigurationProvider) (client MetadataClient, err error) {
	baseClient, err := common.NewClientWithConfig(provider)
	if err != nil {
		return
	}
	client = MetadataClient{BaseClient: baseClient}
	client.BasePath = "v1"
	err = client.setConfigurationProvider(provider)
	return
}

// SetRegion overrides the region of this client.
func (client *MetadataClient) SetRegion(region string) {
	client.Host = common.StringToRegion(region).Endpoint("auth")
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *MetadataClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
	if ok, err := common.IsConfigurationProviderValid(configProvider); !ok {
		return err
	}

	// Error has been checked already
	region, _ := configProvider.Region()
	client.SetRegion(region)
	client.config = &configProvider
	return nil
}

// GetTenantByCompartment gets tenancy information for a given compartment
func (client *MetadataClient) GetTenantByCompartment(ctx context.Context, request GetTenantByCompartmentRequest) (response GetTenantByCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getTenantByCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetTenantByCompartmentResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetTenantByCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetTenantByCompartmentResponse")
	}
	return
}

func (client *MetadataClient) getTenantByCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/compartments/{compartmentId}/tenant")
	if err != nil {
		return nil, err
	}
	var response GetTenantByCompartmentResponse
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
