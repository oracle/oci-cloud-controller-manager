// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Streaming Service API
//
// The API for the Streaming Service.
//

package streaming

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

//ConnectHarnessAdminClient a client for ConnectHarnessAdmin
type ConnectHarnessAdminClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewConnectHarnessAdminClientWithConfigurationProvider Creates a new default ConnectHarnessAdmin client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewConnectHarnessAdminClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client ConnectHarnessAdminClient, err error) {
	baseClient, err := common.NewClientWithConfig(configProvider)
	if err != nil {
		return
	}

	client = ConnectHarnessAdminClient{BaseClient: baseClient}
	client.BasePath = "20180418"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *ConnectHarnessAdminClient) SetRegion(region string) {
	client.Host = common.StringToRegion(region).EndpointForTemplate("streams", "https://streaming.{region}.oci.{secondLevelDomain}")
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *ConnectHarnessAdminClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
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
func (client *ConnectHarnessAdminClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// ChangeConnectHarnessCompartment Moves a resource into a different compartment. When provided, If-Match is checked against ETag values of the resource.
func (client ConnectHarnessAdminClient) ChangeConnectHarnessCompartment(ctx context.Context, request ChangeConnectHarnessCompartmentRequest) (response ChangeConnectHarnessCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.changeConnectHarnessCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			response = ChangeConnectHarnessCompartmentResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeConnectHarnessCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeConnectHarnessCompartmentResponse")
	}
	return
}

// changeConnectHarnessCompartment implements the OCIOperation interface (enables retrying operations)
func (client ConnectHarnessAdminClient) changeConnectHarnessCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/connectharnesses/{connectHarnessId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeConnectHarnessCompartmentResponse
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

// CreateConnectHarness Starts the provisioning of a new connect harness.
// To track the progress of the provisioning, you can periodically call ConnectHarness object tells you its current state.
func (client ConnectHarnessAdminClient) CreateConnectHarness(ctx context.Context, request CreateConnectHarnessRequest) (response CreateConnectHarnessResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createConnectHarness, policy)
	if err != nil {
		if ociResponse != nil {
			response = CreateConnectHarnessResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateConnectHarnessResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateConnectHarnessResponse")
	}
	return
}

// createConnectHarness implements the OCIOperation interface (enables retrying operations)
func (client ConnectHarnessAdminClient) createConnectHarness(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/connectharnesses")
	if err != nil {
		return nil, err
	}

	var response CreateConnectHarnessResponse
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

// DeleteConnectHarness Deletes a connect harness and its content. Connect harness contents are deleted immediately. The service retains records of the connect harness itself for 90 days after deletion.
// The `lifecycleState` parameter of the `ConnectHarness` object changes to `DELETING` and the connect harness becomes inaccessible for read or write operations.
// To verify that a connect harness has been deleted, make a GetConnectHarness request. If the call returns the connect harness's
// lifecycle state as `DELETED`, then the connect harness has been deleted. If the call returns a "404 Not Found" error, that means all records of the
// connect harness have been deleted.
func (client ConnectHarnessAdminClient) DeleteConnectHarness(ctx context.Context, request DeleteConnectHarnessRequest) (response DeleteConnectHarnessResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteConnectHarness, policy)
	if err != nil {
		if ociResponse != nil {
			response = DeleteConnectHarnessResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteConnectHarnessResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteConnectHarnessResponse")
	}
	return
}

// deleteConnectHarness implements the OCIOperation interface (enables retrying operations)
func (client ConnectHarnessAdminClient) deleteConnectHarness(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/connectharnesses/{connectHarnessId}")
	if err != nil {
		return nil, err
	}

	var response DeleteConnectHarnessResponse
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

// GetConnectHarness Gets detailed information about a connect harness.
func (client ConnectHarnessAdminClient) GetConnectHarness(ctx context.Context, request GetConnectHarnessRequest) (response GetConnectHarnessResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getConnectHarness, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetConnectHarnessResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetConnectHarnessResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetConnectHarnessResponse")
	}
	return
}

// getConnectHarness implements the OCIOperation interface (enables retrying operations)
func (client ConnectHarnessAdminClient) getConnectHarness(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/connectharnesses/{connectHarnessId}")
	if err != nil {
		return nil, err
	}

	var response GetConnectHarnessResponse
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

// ListConnectHarnesses Lists the connectharness.
func (client ConnectHarnessAdminClient) ListConnectHarnesses(ctx context.Context, request ListConnectHarnessesRequest) (response ListConnectHarnessesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listConnectHarnesses, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListConnectHarnessesResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListConnectHarnessesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListConnectHarnessesResponse")
	}
	return
}

// listConnectHarnesses implements the OCIOperation interface (enables retrying operations)
func (client ConnectHarnessAdminClient) listConnectHarnesses(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/connectharnesses")
	if err != nil {
		return nil, err
	}

	var response ListConnectHarnessesResponse
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

// UpdateConnectHarness Updates the tags applied to the connect harness.
func (client ConnectHarnessAdminClient) UpdateConnectHarness(ctx context.Context, request UpdateConnectHarnessRequest) (response UpdateConnectHarnessResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateConnectHarness, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateConnectHarnessResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateConnectHarnessResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateConnectHarnessResponse")
	}
	return
}

// updateConnectHarness implements the OCIOperation interface (enables retrying operations)
func (client ConnectHarnessAdminClient) updateConnectHarness(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/connectharnesses/{connectHarnessId}")
	if err != nil {
		return nil, err
	}

	var response UpdateConnectHarnessResponse
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
