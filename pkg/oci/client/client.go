// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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

package client

import (
	"context"
	"time"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/filestorage"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"github.com/oracle/oci-go-sdk/v65/networkloadbalancer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"
)

// defaultSynchronousAPIContextTimeout is the time we wait for synchronous APIs
// to respond before we timeout the request
const defaultSynchronousAPIContextTimeout = 10 * time.Second
const defaultSynchronousAPIPollContextTimeout = 10 * time.Minute

const Ipv6Stack = "IPv6"
const ClusterIpFamilyEnv = "CLUSTER_IP_FAMILY"

// Interface of consumed OCI API functionality.
type Interface interface {
	Compute() ComputeInterface
	LoadBalancer(*zap.SugaredLogger, string, string, *authv1.TokenRequest) GenericLoadBalancerInterface
	Networking(*OCIClientConfig) NetworkingInterface
	BlockStorage() BlockStorageInterface
	FSS(*OCIClientConfig) FileStorageInterface
	Identity(*OCIClientConfig) IdentityInterface
}

type OCIClientConfig struct {
	SaToken      *authv1.TokenRequest
	ParentRptURL string
	TenancyId    string
}

// RateLimiter reader and writer.
type RateLimiter struct {
	Reader flowcontrol.RateLimiter
	Writer flowcontrol.RateLimiter
}

type computeClient interface {
	GetInstance(ctx context.Context, request core.GetInstanceRequest) (response core.GetInstanceResponse, err error)
	ListInstances(ctx context.Context, request core.ListInstancesRequest) (response core.ListInstancesResponse, err error)
	ListVnicAttachments(ctx context.Context, request core.ListVnicAttachmentsRequest) (response core.ListVnicAttachmentsResponse, err error)

	GetVolumeAttachment(ctx context.Context, request core.GetVolumeAttachmentRequest) (response core.GetVolumeAttachmentResponse, err error)
	ListVolumeAttachments(ctx context.Context, request core.ListVolumeAttachmentsRequest) (response core.ListVolumeAttachmentsResponse, err error)
	AttachVolume(ctx context.Context, request core.AttachVolumeRequest) (response core.AttachVolumeResponse, err error)
	DetachVolume(ctx context.Context, request core.DetachVolumeRequest) (response core.DetachVolumeResponse, err error)
	ListInstanceDevices(ctx context.Context, request core.ListInstanceDevicesRequest) (response core.ListInstanceDevicesResponse, err error)
}

type virtualNetworkClient interface {
	GetVnic(ctx context.Context, request core.GetVnicRequest) (response core.GetVnicResponse, err error)
	GetSubnet(ctx context.Context, request core.GetSubnetRequest) (response core.GetSubnetResponse, err error)
	GetVcn(ctx context.Context, request core.GetVcnRequest) (response core.GetVcnResponse, err error)
	GetSecurityList(ctx context.Context, request core.GetSecurityListRequest) (response core.GetSecurityListResponse, err error)
	UpdateSecurityList(ctx context.Context, request core.UpdateSecurityListRequest) (response core.UpdateSecurityListResponse, err error)

	GetPrivateIp(ctx context.Context, request core.GetPrivateIpRequest) (response core.GetPrivateIpResponse, err error)
	ListPrivateIps(ctx context.Context, request core.ListPrivateIpsRequest) (response core.ListPrivateIpsResponse, err error)
	CreatePrivateIp(ctx context.Context, request core.CreatePrivateIpRequest) (response core.CreatePrivateIpResponse, err error)

	GetPublicIpByIpAddress(ctx context.Context, request core.GetPublicIpByIpAddressRequest) (response core.GetPublicIpByIpAddressResponse, err error)
	GetIpv6(ctx context.Context, request core.GetIpv6Request) (response core.GetIpv6Response, err error)

	CreateNetworkSecurityGroup(ctx context.Context, request core.CreateNetworkSecurityGroupRequest) (response core.CreateNetworkSecurityGroupResponse, err error)
	GetNetworkSecurityGroup(ctx context.Context, request core.GetNetworkSecurityGroupRequest) (response core.GetNetworkSecurityGroupResponse, err error)
	ListNetworkSecurityGroups(ctx context.Context, request core.ListNetworkSecurityGroupsRequest) (response core.ListNetworkSecurityGroupsResponse, err error)
	UpdateNetworkSecurityGroup(ctx context.Context, request core.UpdateNetworkSecurityGroupRequest) (response core.UpdateNetworkSecurityGroupResponse, err error)
	DeleteNetworkSecurityGroup(ctx context.Context, request core.DeleteNetworkSecurityGroupRequest) (response core.DeleteNetworkSecurityGroupResponse, err error)

	AddNetworkSecurityGroupSecurityRules(ctx context.Context, request core.AddNetworkSecurityGroupSecurityRulesRequest) (response core.AddNetworkSecurityGroupSecurityRulesResponse, err error)
	RemoveNetworkSecurityGroupSecurityRules(ctx context.Context, request core.RemoveNetworkSecurityGroupSecurityRulesRequest) (response core.RemoveNetworkSecurityGroupSecurityRulesResponse, err error)
	ListNetworkSecurityGroupSecurityRules(ctx context.Context, request core.ListNetworkSecurityGroupSecurityRulesRequest) (response core.ListNetworkSecurityGroupSecurityRulesResponse, err error)
	UpdateNetworkSecurityGroupSecurityRules(ctx context.Context, request core.UpdateNetworkSecurityGroupSecurityRulesRequest) (response core.UpdateNetworkSecurityGroupSecurityRulesResponse, err error)
}

type loadBalancerClient interface {
	GetLoadBalancer(ctx context.Context, request loadbalancer.GetLoadBalancerRequest) (response loadbalancer.GetLoadBalancerResponse, err error)
	ListLoadBalancers(ctx context.Context, request loadbalancer.ListLoadBalancersRequest) (response loadbalancer.ListLoadBalancersResponse, err error)
	CreateLoadBalancer(ctx context.Context, request loadbalancer.CreateLoadBalancerRequest) (response loadbalancer.CreateLoadBalancerResponse, err error)
	DeleteLoadBalancer(ctx context.Context, request loadbalancer.DeleteLoadBalancerRequest) (response loadbalancer.DeleteLoadBalancerResponse, err error)
	ListCertificates(ctx context.Context, request loadbalancer.ListCertificatesRequest) (response loadbalancer.ListCertificatesResponse, err error)
	CreateCertificate(ctx context.Context, request loadbalancer.CreateCertificateRequest) (response loadbalancer.CreateCertificateResponse, err error)
	GetWorkRequest(ctx context.Context, request loadbalancer.GetWorkRequestRequest) (response loadbalancer.GetWorkRequestResponse, err error)
	ListWorkRequests(ctx context.Context, request loadbalancer.ListWorkRequestsRequest) (response loadbalancer.ListWorkRequestsResponse, err error)
	CreateBackendSet(ctx context.Context, request loadbalancer.CreateBackendSetRequest) (response loadbalancer.CreateBackendSetResponse, err error)
	UpdateBackendSet(ctx context.Context, request loadbalancer.UpdateBackendSetRequest) (response loadbalancer.UpdateBackendSetResponse, err error)
	DeleteBackendSet(ctx context.Context, request loadbalancer.DeleteBackendSetRequest) (response loadbalancer.DeleteBackendSetResponse, err error)
	CreateListener(ctx context.Context, request loadbalancer.CreateListenerRequest) (response loadbalancer.CreateListenerResponse, err error)
	UpdateListener(ctx context.Context, request loadbalancer.UpdateListenerRequest) (response loadbalancer.UpdateListenerResponse, err error)
	DeleteListener(ctx context.Context, request loadbalancer.DeleteListenerRequest) (response loadbalancer.DeleteListenerResponse, err error)
	UpdateLoadBalancerShape(ctx context.Context, request loadbalancer.UpdateLoadBalancerShapeRequest) (response loadbalancer.UpdateLoadBalancerShapeResponse, err error)
	UpdateNetworkSecurityGroups(ctx context.Context, request loadbalancer.UpdateNetworkSecurityGroupsRequest) (response loadbalancer.UpdateNetworkSecurityGroupsResponse, err error)
	UpdateLoadBalancer(ctx context.Context, request loadbalancer.UpdateLoadBalancerRequest) (response loadbalancer.UpdateLoadBalancerResponse, err error)
}

type networkLoadBalancerClient interface {
	GetNetworkLoadBalancer(ctx context.Context, request networkloadbalancer.GetNetworkLoadBalancerRequest) (response networkloadbalancer.GetNetworkLoadBalancerResponse, err error)
	ListNetworkLoadBalancers(ctx context.Context, request networkloadbalancer.ListNetworkLoadBalancersRequest) (response networkloadbalancer.ListNetworkLoadBalancersResponse, err error)
	CreateNetworkLoadBalancer(ctx context.Context, request networkloadbalancer.CreateNetworkLoadBalancerRequest) (response networkloadbalancer.CreateNetworkLoadBalancerResponse, err error)
	DeleteNetworkLoadBalancer(ctx context.Context, request networkloadbalancer.DeleteNetworkLoadBalancerRequest) (response networkloadbalancer.DeleteNetworkLoadBalancerResponse, err error)
	GetWorkRequest(ctx context.Context, request networkloadbalancer.GetWorkRequestRequest) (response networkloadbalancer.GetWorkRequestResponse, err error)
	ListWorkRequests(ctx context.Context, request networkloadbalancer.ListWorkRequestsRequest) (response networkloadbalancer.ListWorkRequestsResponse, err error)
	CreateBackendSet(ctx context.Context, request networkloadbalancer.CreateBackendSetRequest) (response networkloadbalancer.CreateBackendSetResponse, err error)
	UpdateBackendSet(ctx context.Context, request networkloadbalancer.UpdateBackendSetRequest) (response networkloadbalancer.UpdateBackendSetResponse, err error)
	DeleteBackendSet(ctx context.Context, request networkloadbalancer.DeleteBackendSetRequest) (response networkloadbalancer.DeleteBackendSetResponse, err error)
	CreateListener(ctx context.Context, request networkloadbalancer.CreateListenerRequest) (response networkloadbalancer.CreateListenerResponse, err error)
	UpdateListener(ctx context.Context, request networkloadbalancer.UpdateListenerRequest) (response networkloadbalancer.UpdateListenerResponse, err error)
	DeleteListener(ctx context.Context, request networkloadbalancer.DeleteListenerRequest) (response networkloadbalancer.DeleteListenerResponse, err error)
	UpdateNetworkSecurityGroups(ctx context.Context, request networkloadbalancer.UpdateNetworkSecurityGroupsRequest) (response networkloadbalancer.UpdateNetworkSecurityGroupsResponse, err error)
	UpdateNetworkLoadBalancer(ctx context.Context, request networkloadbalancer.UpdateNetworkLoadBalancerRequest) (response networkloadbalancer.UpdateNetworkLoadBalancerResponse, err error)
}

type filestorageClient interface {
	CreateFileSystem(ctx context.Context, request filestorage.CreateFileSystemRequest) (response filestorage.CreateFileSystemResponse, err error)
	GetFileSystem(ctx context.Context, request filestorage.GetFileSystemRequest) (response filestorage.GetFileSystemResponse, err error)
	ListFileSystems(ctx context.Context, request filestorage.ListFileSystemsRequest) (response filestorage.ListFileSystemsResponse, err error)
	DeleteFileSystem(ctx context.Context, request filestorage.DeleteFileSystemRequest) (response filestorage.DeleteFileSystemResponse, err error)

	CreateExport(ctx context.Context, request filestorage.CreateExportRequest) (response filestorage.CreateExportResponse, err error)
	ListExports(ctx context.Context, request filestorage.ListExportsRequest) (response filestorage.ListExportsResponse, err error)
	GetExport(ctx context.Context, request filestorage.GetExportRequest) (response filestorage.GetExportResponse, err error)
	DeleteExport(ctx context.Context, request filestorage.DeleteExportRequest) (response filestorage.DeleteExportResponse, err error)

	GetMountTarget(ctx context.Context, request filestorage.GetMountTargetRequest) (response filestorage.GetMountTargetResponse, err error)
	CreateMountTarget(ctx context.Context, request filestorage.CreateMountTargetRequest) (response filestorage.CreateMountTargetResponse, err error)
	DeleteMountTarget(ctx context.Context, request filestorage.DeleteMountTargetRequest) (response filestorage.DeleteMountTargetResponse, err error)
	ListMountTargets(ctx context.Context, request filestorage.ListMountTargetsRequest) (response filestorage.ListMountTargetsResponse, err error)
}

type blockstorageClient interface {
	GetVolume(ctx context.Context, request core.GetVolumeRequest) (response core.GetVolumeResponse, err error)
	CreateVolume(ctx context.Context, request core.CreateVolumeRequest) (response core.CreateVolumeResponse, err error)
	DeleteVolume(ctx context.Context, request core.DeleteVolumeRequest) (response core.DeleteVolumeResponse, err error)
	ListVolumes(ctx context.Context, request core.ListVolumesRequest) (response core.ListVolumesResponse, err error)
	UpdateVolume(ctx context.Context, request core.UpdateVolumeRequest) (response core.UpdateVolumeResponse, err error)

	GetVolumeBackup(ctx context.Context, request core.GetVolumeBackupRequest) (response core.GetVolumeBackupResponse, err error)
	CreateVolumeBackup(ctx context.Context, request core.CreateVolumeBackupRequest) (response core.CreateVolumeBackupResponse, err error)
	DeleteVolumeBackup(ctx context.Context, request core.DeleteVolumeBackupRequest) (response core.DeleteVolumeBackupResponse, err error)
	ListVolumeBackups(ctx context.Context, request core.ListVolumeBackupsRequest) (response core.ListVolumeBackupsResponse, err error)
}

type identityClient interface {
	ListAvailabilityDomains(ctx context.Context, request identity.ListAvailabilityDomainsRequest) (identity.ListAvailabilityDomainsResponse, error)
}

// TODO: Uncomment when compartments is available in OCI Go-SDK
//type compartmentClient interface {
//	ListAvailabilityDomains(ctx context.Context, request compartments.ListAvailabilityDomainsRequest) (compartments.ListAvailabilityDomainsResponse, error)
//}

type client struct {
	compute             computeClient
	network             virtualNetworkClient
	loadbalancer        GenericLoadBalancerInterface
	networkloadbalancer GenericLoadBalancerInterface
	filestorage         filestorageClient
	bs                  blockstorageClient
	identity            identityClient
	//compartment 		compartmentClient

	requestMetadata common.RequestMetadata
	rateLimiter     RateLimiter

	subnetCache cache.Store
	logger      *zap.SugaredLogger
}

// New constructs an OCI API client.
func New(logger *zap.SugaredLogger, cp common.ConfigurationProvider, opRateLimiter *RateLimiter, cloudProviderConfig *providercfg.Config) (Interface, error) {

	compute, err := core.NewComputeClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewComputeClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &compute.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring load balancer client custom transport")
	}

	network, err := core.NewVirtualNetworkClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewVirtualNetworkClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &network.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring load balancer client custom transport")
	}

	lb, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewLoadBalancerClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &lb.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring loadbalancer client custom transport")
	}

	nlb, err := networkloadbalancer.NewNetworkLoadBalancerClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewNetworkLoadBalancerClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &nlb.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring networkloadbalancer client custom transport")
	}

	identity, err := identity.NewIdentityClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewIdentityClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &identity.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring identity service client custom transport")
	}

	// TODO: Uncomment when compartments is available in OCI Go-SDK
	//compartment, err := compartments.NewCompartmentsClientWithConfigurationProvider(cp)
	//if err != nil {
	//	return nil, errors.Wrap(err, "NewCompartmentsClientWithConfigurationProvider")
	//}
	//
	//setupBaseClient(logger, &compartment.BaseClient, signer, interceptor, "")
	//
	//err = configureCustomTransport(logger, &compartment.BaseClient)
	//if err != nil {
	//	return nil, errors.Wrap(err, "configuring compartment service client custom transport")
	//}

	bs, err := core.NewBlockstorageClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewBlockstorageClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &bs.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring block storage service client custom transport")
	}

	fss, err := filestorage.NewFileStorageClientWithConfigurationProvider(cp)
	if err != nil {
		return nil, errors.Wrap(err, "NewFileStorageClientWithConfigurationProvider")
	}

	err = configureCustomTransport(logger, &fss.BaseClient)
	if err != nil {
		return nil, errors.Wrap(err, "configuring file storage service client custom transport")
	}

	requestMetadata := common.RequestMetadata{
		RetryPolicy: newRetryPolicy(),
	}

	loadbalancer := NewLBClient(lb, requestMetadata, opRateLimiter)
	networkloadbalancer := NewNLBClient(nlb, requestMetadata, opRateLimiter)

	c := &client{
		compute:             &compute,
		network:             &network,
		identity:            &identity,
		loadbalancer:        loadbalancer,
		networkloadbalancer: networkloadbalancer,
		bs:                  &bs,
		filestorage:         &fss,
		//compartment:     	 &compartment,

		rateLimiter:     *opRateLimiter,
		requestMetadata: requestMetadata,

		subnetCache: cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
		logger:      logger,
	}

	return c, nil
}

// LoadBalancer constructs an OCI LB/NLB API client using workload identity token if service account provided
// or else returns the default cluster level client
func (c *client) LoadBalancer(logger *zap.SugaredLogger, lbType string, targetTenancyID string, tokenRequest *authv1.TokenRequest) (genericLoadBalancer GenericLoadBalancerInterface) {

	// tokenRequest is nil if Workload Identity LB/NLB client is not requested
	if tokenRequest == nil {
		if lbType == "nlb" {
			return c.networkloadbalancer
		}
		if lbType == "lb" {
			return c.loadbalancer
		}
		logger.Error("Failed to get Client since load-balancer-type is neither lb or nlb!")
		return nil
	}

	// If tokenRequest is present then the requested LB/NLB client is WRIS / Workload Identity RP based
	tokenProvider := auth.NewSuppliedServiceAccountTokenProvider(tokenRequest.Status.Token)
	configProvider, err := auth.OkeWorkloadIdentityConfigurationProviderWithServiceAccountTokenProvider(tokenProvider)
	if err != nil {
		logger.Error("Failed to get oke workload identity configuration provider! " + err.Error())
		return nil
	}

	if lbType == "lb" {
		lb, err := loadbalancer.NewLoadBalancerClientWithConfigurationProvider(configProvider)
		if err != nil {
			logger.Error("Failed to get new LB client with oke workload identity configuration provider! Error:" + err.Error())
			return nil
		}

		err = configureCustomTransport(logger, &lb.BaseClient)
		if err != nil {
			logger.Error("Failed configure custom transport for LB Client! Error:" + err.Error())
			return nil
		}

		return &loadbalancerClientStruct{
			loadbalancer:    lb,
			requestMetadata: c.requestMetadata,
			rateLimiter:     c.rateLimiter,
		}
	}
	if lbType == "nlb" {
		nlb, err := networkloadbalancer.NewNetworkLoadBalancerClientWithConfigurationProvider(configProvider)
		if err != nil {
			logger.Error("Failed to get new NLB client with oke workload identity configuration provider! Error:" + err.Error())
			return nil
		}

		err = configureCustomTransport(logger, &nlb.BaseClient)
		if err != nil {
			logger.Error("Failed configure custom transport for NLB Client! Error:" + err.Error())
			return nil
		}

		return &networkLoadbalancer{
			networkloadbalancer: nlb,
			requestMetadata:     c.requestMetadata,
			rateLimiter:         c.rateLimiter,
		}
	}
	logger.Error("Failed to get Client since load-balancer-type is neither lb or nlb!")
	return nil
}

func (c *client) Networking(ociClientConfig *OCIClientConfig) NetworkingInterface {
	if ociClientConfig == nil {
		return c
	}
	if ociClientConfig.SaToken != nil {
		configProvider, err := getConfigurationProvider(c.logger, ociClientConfig.SaToken, ociClientConfig.ParentRptURL)

		network, err := core.NewVirtualNetworkClientWithConfigurationProvider(configProvider)
		if err != nil {
			c.logger.Errorf("Failed to create Network workload identity client %v", err)
			return nil
		}

		err = configureCustomTransport(c.logger, &network.BaseClient)
		if err != nil {
			c.logger.Error("Failed configure custom transport for Network Client %v", err)
			return nil
		}

		return &client{
			network:         &network,
			requestMetadata: c.requestMetadata,
			rateLimiter:     c.rateLimiter,
			subnetCache:     cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
			logger:          c.logger,
		}
	}
	return c
}

func (c *client) Compute() ComputeInterface {
	return c
}

func (c *client) Identity(ociClientConfig *OCIClientConfig) IdentityInterface {

	if ociClientConfig == nil {
		return c
	}
	if ociClientConfig.SaToken != nil {

		configProvider, err := getConfigurationProvider(c.logger, ociClientConfig.SaToken, ociClientConfig.ParentRptURL)

		identity, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
		if err != nil {
			c.logger.Errorf("Failed to create Identity workload identity  %v", err)
			return nil
		}

		err = configureCustomTransport(c.logger, &identity.BaseClient)
		if err != nil {
			c.logger.Error("Failed configure custom transport for Identity Client %v", err)
			return nil
		}

		// TODO: Uncomment when compartments is available in OCI Go-SDK
		//compartment, err := compartments.NewCompartmentsClientWithConfigurationProvider(configProvider)
		//if err != nil {
		//	c.logger.Errorf("Failed to create Compartments workload identity client  %v", err)
		//	return nil
		//}
		//
		//setupBaseClient(c.logger, &compartment.BaseClient, signer, interceptor, "")
		//
		//err = configureCustomTransport(c.logger, &compartment.BaseClient)
		//if err != nil {
		//	c.logger.Error("Failed configure custom transport for Compartments Client %v", err)
		//	return nil
		//}

		return &client{
			//compartment: 	     &compartment,
			identity:        &identity,
			requestMetadata: c.requestMetadata,
			rateLimiter:     c.rateLimiter,
			subnetCache:     cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
			logger:          c.logger,
		}
	}
	return c
}

func (c *client) BlockStorage() BlockStorageInterface {
	return c
}

func (c *client) FSS(ociClientConfig *OCIClientConfig) FileStorageInterface {

	if ociClientConfig == nil {
		return c
	}
	if ociClientConfig.SaToken != nil {

		configProvider, err := getConfigurationProvider(c.logger, ociClientConfig.SaToken, ociClientConfig.ParentRptURL)
		fc, err := filestorage.NewFileStorageClientWithConfigurationProvider(configProvider)
		if err != nil {
			c.logger.Errorf("Failed to create FSS workload identity client %v", err)
			return nil
		}

		err = configureCustomTransport(c.logger, &fc.BaseClient)
		if err != nil {
			c.logger.Errorf("Failed configure custom transport for FSS Client %v", err.Error())
			return nil
		}

		return &client{
			filestorage:     &fc,
			requestMetadata: c.requestMetadata,
			rateLimiter:     c.rateLimiter,
			subnetCache:     cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
			logger:          c.logger,
		}
	}
	return c
}

func configureCustomTransport(logger *zap.SugaredLogger, baseClient *common.BaseClient) error {
	return nil
}

func getDefaultRequestMetadata(existingRequestMetadata common.RequestMetadata) common.RequestMetadata {
	if existingRequestMetadata.RetryPolicy != nil {
		return existingRequestMetadata
	}
	requestMetadata := common.RequestMetadata{
		RetryPolicy: newRetryPolicy(),
	}
	return requestMetadata
}

func getConfigurationProvider(logger *zap.SugaredLogger, tokenRequest *authv1.TokenRequest, rptURL string) (common.ConfigurationProvider, error) {

	tokenProvider := auth.NewSuppliedServiceAccountTokenProvider(tokenRequest.Status.Token)
	configProvider, err := auth.OkeWorkloadIdentityConfigurationProviderWithServiceAccountTokenProvider(tokenProvider)
	if err != nil {
		logger.Errorf("failed to get workload identity configuration provider %v", err.Error())
		return nil, err
	}

	if rptURL != "" {
		configProvider, err = auth.ResourcePrincipalV3ConfiguratorBuilder(configProvider).WithParentRPSTURL("").WithParentRPTURL(rptURL).Build()
		if err != nil {
			logger.Errorf("failed to get resource Principal configuration provider %v", err.Error())
			return nil, err
		}
	}
	return configProvider, nil
}
