// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// CloudEvents API
//
// API for the CloudEvents Service. Use this API to manage rules and actions that create automation
// in your tenancy. For more information, see Overview of Events (https://docs.cloud.oracle.com/iaas/Content/Events/Concepts/eventsoverview.htm).
//

package cloudevents

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

//CloudEventsClient a client for CloudEvents
type CloudEventsClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewCloudEventsClientWithConfigurationProvider Creates a new default CloudEvents client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewCloudEventsClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client CloudEventsClient, err error) {
	baseClient, err := common.NewClientWithConfig(configProvider)
	if err != nil {
		return
	}

	client = CloudEventsClient{BaseClient: baseClient}
	client.BasePath = "20181201"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *CloudEventsClient) SetRegion(region string) {
	client.Host = common.StringToRegion(region).Endpoint("cloudevents")
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *CloudEventsClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
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
func (client *CloudEventsClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// CreateRule Creates a new rule.
func (client CloudEventsClient) CreateRule(ctx context.Context, request CreateRuleRequest) (response CreateRuleResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createRule, policy)
	if err != nil {
		if ociResponse != nil {
			response = CreateRuleResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateRuleResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateRuleResponse")
	}
	return
}

// createRule implements the OCIOperation interface (enables retrying operations)
func (client CloudEventsClient) createRule(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/rules")
	if err != nil {
		return nil, err
	}

	var response CreateRuleResponse
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

// DeleteRule Deletes an event rule.
func (client CloudEventsClient) DeleteRule(ctx context.Context, request DeleteRuleRequest) (response DeleteRuleResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteRule, policy)
	if err != nil {
		if ociResponse != nil {
			response = DeleteRuleResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteRuleResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteRuleResponse")
	}
	return
}

// deleteRule implements the OCIOperation interface (enables retrying operations)
func (client CloudEventsClient) deleteRule(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/rules/{ruleId}")
	if err != nil {
		return nil, err
	}

	var response DeleteRuleResponse
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

// GetRule Retrieves an event rule.
func (client CloudEventsClient) GetRule(ctx context.Context, request GetRuleRequest) (response GetRuleResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getRule, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetRuleResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetRuleResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetRuleResponse")
	}
	return
}

// getRule implements the OCIOperation interface (enables retrying operations)
func (client CloudEventsClient) getRule(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/rules/{ruleId}")
	if err != nil {
		return nil, err
	}

	var response GetRuleResponse
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

// ListRules Lists rules for this compartment.
func (client CloudEventsClient) ListRules(ctx context.Context, request ListRulesRequest) (response ListRulesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listRules, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListRulesResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListRulesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListRulesResponse")
	}
	return
}

// listRules implements the OCIOperation interface (enables retrying operations)
func (client CloudEventsClient) listRules(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/rules")
	if err != nil {
		return nil, err
	}

	var response ListRulesResponse
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

// UpdateRule Updates an event rule.
func (client CloudEventsClient) UpdateRule(ctx context.Context, request UpdateRuleRequest) (response UpdateRuleResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateRule, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateRuleResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateRuleResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateRuleResponse")
	}
	return
}

// updateRule implements the OCIOperation interface (enables retrying operations)
func (client CloudEventsClient) updateRule(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/rules/{ruleId}")
	if err != nil {
		return nil, err
	}

	var response UpdateRuleResponse
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
