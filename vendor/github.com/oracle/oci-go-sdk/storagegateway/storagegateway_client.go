// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Storage Gateway API
//
// API for the Storage Gateway service. Use this API to manage storage gateways and related items. For more
// information, see Overview of Storage Gateway (https://docs.cloud.oracle.com/iaas/Content/StorageGateway/Concepts/storagegatewayoverview.htm).
//

package storagegateway

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

//StorageGatewayClient a client for StorageGateway
type StorageGatewayClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewStorageGatewayClientWithConfigurationProvider Creates a new default StorageGateway client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewStorageGatewayClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client StorageGatewayClient, err error) {
	baseClient, err := common.NewClientWithConfig(configProvider)
	if err != nil {
		return
	}

	client = StorageGatewayClient{BaseClient: baseClient}
	client.BasePath = "20190101"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *StorageGatewayClient) SetRegion(region string) {
	client.Host = common.StringToRegion(region).EndpointForTemplate("storagegateway", "https://storage-gateway.{region}.oci.{secondLevelDomain}")
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *StorageGatewayClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
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
func (client *StorageGatewayClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// CancelCloudSync Cancels the specified cloud sync in the specified storage gateway.
func (client StorageGatewayClient) CancelCloudSync(ctx context.Context, request CancelCloudSyncRequest) (response CancelCloudSyncResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.cancelCloudSync, policy)
	if err != nil {
		if ociResponse != nil {
			response = CancelCloudSyncResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CancelCloudSyncResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CancelCloudSyncResponse")
	}
	return
}

// cancelCloudSync implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) cancelCloudSync(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/storageGateways/{storageGatewayId}/cloudSyncs/{cloudSyncName}/actions/cancel")
	if err != nil {
		return nil, err
	}

	var response CancelCloudSyncResponse
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

// ChangeStorageGatewayCompartment Moves a storage gateway into a different compartment within the same tenancy. For information about moving
// resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
func (client StorageGatewayClient) ChangeStorageGatewayCompartment(ctx context.Context, request ChangeStorageGatewayCompartmentRequest) (response ChangeStorageGatewayCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeStorageGatewayCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			response = ChangeStorageGatewayCompartmentResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeStorageGatewayCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeStorageGatewayCompartmentResponse")
	}
	return
}

// changeStorageGatewayCompartment implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) changeStorageGatewayCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/storageGateways/{storageGatewayId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeStorageGatewayCompartmentResponse
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

// ConnectFileSystem Connects the specified file system in the specified storage gateway to its object storage bucket.
func (client StorageGatewayClient) ConnectFileSystem(ctx context.Context, request ConnectFileSystemRequest) (response ConnectFileSystemResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.connectFileSystem, policy)
	if err != nil {
		if ociResponse != nil {
			response = ConnectFileSystemResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ConnectFileSystemResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ConnectFileSystemResponse")
	}
	return
}

// connectFileSystem implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) connectFileSystem(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/storageGateways/{storageGatewayId}/fileSystems/{fileSystemName}/actions/connect")
	if err != nil {
		return nil, err
	}

	var response ConnectFileSystemResponse
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

// CreateCloudSync Creates a cloud sync in the specified storage gateway and compartment. For general information about cloud syncs,
// see Using Storage Gateway Cloud Sync (https://docs.cloud.oracle.com/iaas/Content/StorageGateway/Reference/storagegatewaycloudsync.htm).
// After you send your request, the new object's state will temporarily be CREATING. Before using the
// the object, first make sure its state has changed to ACTIVE.
// For general information about Oracle Cloud Infrastructure API requests, see
// REST APIs (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm).
func (client StorageGatewayClient) CreateCloudSync(ctx context.Context, request CreateCloudSyncRequest) (response CreateCloudSyncResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createCloudSync, policy)
	if err != nil {
		if ociResponse != nil {
			response = CreateCloudSyncResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateCloudSyncResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateCloudSyncResponse")
	}
	return
}

// createCloudSync implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) createCloudSync(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/storageGateways/{storageGatewayId}/cloudSyncs")
	if err != nil {
		return nil, err
	}

	var response CreateCloudSyncResponse
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

// CreateFileSystem Creates a file system in the specified storage gateway. For more information about storage gateway file systems,
// see Creating Your First File System (https://docs.cloud.oracle.com/iaas/Content/StorageGateway/Tasks/creatingyourfirstfilesystem.htm).
// After you send your request, the new object's state will temporarily be CREATING. Before using the
// the object, first make sure its state has changed to ACTIVE.
// For general information about Oracle Cloud Infrastructure API requests, see
// REST APIs (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm).
func (client StorageGatewayClient) CreateFileSystem(ctx context.Context, request CreateFileSystemRequest) (response CreateFileSystemResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createFileSystem, policy)
	if err != nil {
		if ociResponse != nil {
			response = CreateFileSystemResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateFileSystemResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateFileSystemResponse")
	}
	return
}

// createFileSystem implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) createFileSystem(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/storageGateways/{storageGatewayId}/fileSystems")
	if err != nil {
		return nil, err
	}

	var response CreateFileSystemResponse
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

// CreateStorageGateway Creates a new storage gateway in the specified compartment. For general information about storage gateways, see
// Overview of Storage Gateway (https://docs.cloud.oracle.com/iaas/Content/StorageGateway/Concepts/storagegatewayoverview.htm).
// For the purposes of access control, you must provide the OCID of the compartment where you want the storage
// gateway to reside. For information about access control and compartments, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm).
// You must specify a name for the storage gateway. The name must be unique across all storage gateways in your
// compartment. The storage gateway name cannot be changed.
// After you send your request, the new object's state will temporarily be CREATING. Before using the object,
// ensure that its state is either INACTIVE or ACTIVE.
// For general information about Oracle Cloud Infrastructure API requests, see
// REST APIs (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm).
func (client StorageGatewayClient) CreateStorageGateway(ctx context.Context, request CreateStorageGatewayRequest) (response CreateStorageGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createStorageGateway, policy)
	if err != nil {
		if ociResponse != nil {
			response = CreateStorageGatewayResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateStorageGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateStorageGatewayResponse")
	}
	return
}

// createStorageGateway implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) createStorageGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/storageGateways")
	if err != nil {
		return nil, err
	}

	var response CreateStorageGatewayResponse
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

// DeleteCloudSync Deletes the specified cloud sync.
func (client StorageGatewayClient) DeleteCloudSync(ctx context.Context, request DeleteCloudSyncRequest) (response DeleteCloudSyncResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteCloudSync, policy)
	if err != nil {
		if ociResponse != nil {
			response = DeleteCloudSyncResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteCloudSyncResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteCloudSyncResponse")
	}
	return
}

// deleteCloudSync implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) deleteCloudSync(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/storageGateways/{storageGatewayId}/cloudSyncs/{cloudSyncName}")
	if err != nil {
		return nil, err
	}

	var response DeleteCloudSyncResponse
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

// DeleteFileSystem Deletes the specified storage gateway file system.
func (client StorageGatewayClient) DeleteFileSystem(ctx context.Context, request DeleteFileSystemRequest) (response DeleteFileSystemResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteFileSystem, policy)
	if err != nil {
		if ociResponse != nil {
			response = DeleteFileSystemResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteFileSystemResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteFileSystemResponse")
	}
	return
}

// deleteFileSystem implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) deleteFileSystem(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/storageGateways/{storageGatewayId}/fileSystems/{fileSystemName}")
	if err != nil {
		return nil, err
	}

	var response DeleteFileSystemResponse
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

// DeleteStorageGateway Deletes the specified storage gateway.
func (client StorageGatewayClient) DeleteStorageGateway(ctx context.Context, request DeleteStorageGatewayRequest) (response DeleteStorageGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteStorageGateway, policy)
	if err != nil {
		if ociResponse != nil {
			response = DeleteStorageGatewayResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteStorageGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteStorageGatewayResponse")
	}
	return
}

// deleteStorageGateway implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) deleteStorageGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/storageGateways/{storageGatewayId}")
	if err != nil {
		return nil, err
	}

	var response DeleteStorageGatewayResponse
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

// DisconnectFileSystem Disconnects the specified file system in the specified storage gateway from its object storage bucket.
func (client StorageGatewayClient) DisconnectFileSystem(ctx context.Context, request DisconnectFileSystemRequest) (response DisconnectFileSystemResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.disconnectFileSystem, policy)
	if err != nil {
		if ociResponse != nil {
			response = DisconnectFileSystemResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DisconnectFileSystemResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DisconnectFileSystemResponse")
	}
	return
}

// disconnectFileSystem implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) disconnectFileSystem(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/storageGateways/{storageGatewayId}/fileSystems/{fileSystemName}/actions/disconnect")
	if err != nil {
		return nil, err
	}

	var response DisconnectFileSystemResponse
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

// GetCloudSync Gets the specified cloud sync's configuration information.
func (client StorageGatewayClient) GetCloudSync(ctx context.Context, request GetCloudSyncRequest) (response GetCloudSyncResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCloudSync, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetCloudSyncResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetCloudSyncResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetCloudSyncResponse")
	}
	return
}

// getCloudSync implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) getCloudSync(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/storageGateways/{storageGatewayId}/cloudSyncs/{cloudSyncName}")
	if err != nil {
		return nil, err
	}

	var response GetCloudSyncResponse
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

// GetCloudSyncHealth Gets the health status of the specified cloud sync.
func (client StorageGatewayClient) GetCloudSyncHealth(ctx context.Context, request GetCloudSyncHealthRequest) (response GetCloudSyncHealthResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCloudSyncHealth, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetCloudSyncHealthResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetCloudSyncHealthResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetCloudSyncHealthResponse")
	}
	return
}

// getCloudSyncHealth implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) getCloudSyncHealth(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/storageGateways/{storageGatewayId}/cloudSyncs/{cloudSyncName}/health")
	if err != nil {
		return nil, err
	}

	var response GetCloudSyncHealthResponse
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

// GetFileSystem Gets information about the specified file system.
func (client StorageGatewayClient) GetFileSystem(ctx context.Context, request GetFileSystemRequest) (response GetFileSystemResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getFileSystem, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetFileSystemResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetFileSystemResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetFileSystemResponse")
	}
	return
}

// getFileSystem implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) getFileSystem(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/storageGateways/{storageGatewayId}/fileSystems/{fileSystemName}")
	if err != nil {
		return nil, err
	}

	var response GetFileSystemResponse
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

// GetFileSystemHealth Gets the health about the file system.
func (client StorageGatewayClient) GetFileSystemHealth(ctx context.Context, request GetFileSystemHealthRequest) (response GetFileSystemHealthResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getFileSystemHealth, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetFileSystemHealthResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetFileSystemHealthResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetFileSystemHealthResponse")
	}
	return
}

// getFileSystemHealth implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) getFileSystemHealth(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/storageGateways/{storageGatewayId}/fileSystems/{fileSystemName}/health")
	if err != nil {
		return nil, err
	}

	var response GetFileSystemHealthResponse
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

// GetStorageGateway Gets configuration and status information for the specified storage gateway.
func (client StorageGatewayClient) GetStorageGateway(ctx context.Context, request GetStorageGatewayRequest) (response GetStorageGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getStorageGateway, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetStorageGatewayResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetStorageGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetStorageGatewayResponse")
	}
	return
}

// getStorageGateway implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) getStorageGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/storageGateways/{storageGatewayId}")
	if err != nil {
		return nil, err
	}

	var response GetStorageGatewayResponse
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

// GetStorageGatewayHealth Gets health information for the specified storage gateway.
func (client StorageGatewayClient) GetStorageGatewayHealth(ctx context.Context, request GetStorageGatewayHealthRequest) (response GetStorageGatewayHealthResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getStorageGatewayHealth, policy)
	if err != nil {
		if ociResponse != nil {
			response = GetStorageGatewayHealthResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetStorageGatewayHealthResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetStorageGatewayHealthResponse")
	}
	return
}

// getStorageGatewayHealth implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) getStorageGatewayHealth(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/storageGateways/{storageGatewayId}/health")
	if err != nil {
		return nil, err
	}

	var response GetStorageGatewayHealthResponse
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

// ListCloudSyncs Lists all cloud syncs in the specified storage gateway and compartment.
func (client StorageGatewayClient) ListCloudSyncs(ctx context.Context, request ListCloudSyncsRequest) (response ListCloudSyncsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listCloudSyncs, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListCloudSyncsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListCloudSyncsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListCloudSyncsResponse")
	}
	return
}

// listCloudSyncs implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) listCloudSyncs(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/storageGateways/{storageGatewayId}/cloudSyncs")
	if err != nil {
		return nil, err
	}

	var response ListCloudSyncsResponse
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

// ListFileSystems Lists all file systems in the specified storage gateway.
func (client StorageGatewayClient) ListFileSystems(ctx context.Context, request ListFileSystemsRequest) (response ListFileSystemsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listFileSystems, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListFileSystemsResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListFileSystemsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListFileSystemsResponse")
	}
	return
}

// listFileSystems implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) listFileSystems(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/storageGateways/{storageGatewayId}/fileSystems")
	if err != nil {
		return nil, err
	}

	var response ListFileSystemsResponse
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

// ListStorageGateways Lists all storage gateways in the specified compartment.
func (client StorageGatewayClient) ListStorageGateways(ctx context.Context, request ListStorageGatewaysRequest) (response ListStorageGatewaysResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listStorageGateways, policy)
	if err != nil {
		if ociResponse != nil {
			response = ListStorageGatewaysResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListStorageGatewaysResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListStorageGatewaysResponse")
	}
	return
}

// listStorageGateways implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) listStorageGateways(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/storageGateways")
	if err != nil {
		return nil, err
	}

	var response ListStorageGatewaysResponse
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

// ReclaimFileSystem Reclaims the specified file system in the specified storage gateway from its object storage bucket.
func (client StorageGatewayClient) ReclaimFileSystem(ctx context.Context, request ReclaimFileSystemRequest) (response ReclaimFileSystemResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.reclaimFileSystem, policy)
	if err != nil {
		if ociResponse != nil {
			response = ReclaimFileSystemResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ReclaimFileSystemResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ReclaimFileSystemResponse")
	}
	return
}

// reclaimFileSystem implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) reclaimFileSystem(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/storageGateways/{storageGatewayId}/fileSystems/{fileSystemName}/actions/reclaim")
	if err != nil {
		return nil, err
	}

	var response ReclaimFileSystemResponse
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

// RefreshFileSystem Refreshes the specified file system in the specified storage gateway with the contents in the object storage
// bucket. A file system can handle other API calls, such as `Disconnect` and `DeleteFileSystem`, while it is in
// refresh mode.
func (client StorageGatewayClient) RefreshFileSystem(ctx context.Context, request RefreshFileSystemRequest) (response RefreshFileSystemResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.refreshFileSystem, policy)
	if err != nil {
		if ociResponse != nil {
			response = RefreshFileSystemResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(RefreshFileSystemResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into RefreshFileSystemResponse")
	}
	return
}

// refreshFileSystem implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) refreshFileSystem(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/storageGateways/{storageGatewayId}/fileSystems/{fileSystemName}/actions/refresh")
	if err != nil {
		return nil, err
	}

	var response RefreshFileSystemResponse
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

// RunCloudSync Runs the specified cloud sync in the specified storage gateway.
func (client StorageGatewayClient) RunCloudSync(ctx context.Context, request RunCloudSyncRequest) (response RunCloudSyncResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.runCloudSync, policy)
	if err != nil {
		if ociResponse != nil {
			response = RunCloudSyncResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(RunCloudSyncResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into RunCloudSyncResponse")
	}
	return
}

// runCloudSync implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) runCloudSync(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/storageGateways/{storageGatewayId}/cloudSyncs/{cloudSyncName}/actions/run")
	if err != nil {
		return nil, err
	}

	var response RunCloudSyncResponse
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

// UpdateCloudSync Updates the configuration information of the specified cloud sync.
func (client StorageGatewayClient) UpdateCloudSync(ctx context.Context, request UpdateCloudSyncRequest) (response UpdateCloudSyncResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateCloudSync, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateCloudSyncResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateCloudSyncResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateCloudSyncResponse")
	}
	return
}

// updateCloudSync implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) updateCloudSync(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/storageGateways/{storageGatewayId}/cloudSyncs/{cloudSyncName}")
	if err != nil {
		return nil, err
	}

	var response UpdateCloudSyncResponse
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

// UpdateFileSystem Updates the configuration of the specified file system.
func (client StorageGatewayClient) UpdateFileSystem(ctx context.Context, request UpdateFileSystemRequest) (response UpdateFileSystemResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateFileSystem, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateFileSystemResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateFileSystemResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateFileSystemResponse")
	}
	return
}

// updateFileSystem implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) updateFileSystem(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/storageGateways/{storageGatewayId}/fileSystems/{fileSystemName}")
	if err != nil {
		return nil, err
	}

	var response UpdateFileSystemResponse
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

// UpdateStorageGateway Updates the specified storage gateway's configuration.
func (client StorageGatewayClient) UpdateStorageGateway(ctx context.Context, request UpdateStorageGatewayRequest) (response UpdateStorageGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateStorageGateway, policy)
	if err != nil {
		if ociResponse != nil {
			response = UpdateStorageGatewayResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateStorageGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateStorageGatewayResponse")
	}
	return
}

// updateStorageGateway implements the OCIOperation interface (enables retrying operations)
func (client StorageGatewayClient) updateStorageGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/storageGateways/{storageGatewayId}")
	if err != nil {
		return nil, err
	}

	var response UpdateStorageGatewayResponse
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
