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

package oci

import (
	"context"
	"errors"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/core"
	"github.com/oracle/oci-go-sdk/v50/filestorage"
	"github.com/oracle/oci-go-sdk/v50/identity"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	instanceVnics = map[string]*core.Vnic{
		"basic-complete": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("basic-complete"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"no-external-ip": {
			PrivateIp:     common.String("10.0.0.1"),
			HostnameLabel: common.String("no-external-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"no-internal-ip": {
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-internal-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"invalid-internal-ip": {
			PrivateIp:     common.String("10.0.0."),
			HostnameLabel: common.String("no-internal-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"invalid-external-ip": {
			PublicIp:      common.String("0.0.0."),
			HostnameLabel: common.String("invalid-external-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"no-hostname-label": {
			PrivateIp: common.String("10.0.0.1"),
			PublicIp:  common.String("0.0.0.1"),
			SubnetId:  common.String("subnetwithdnslabel"),
		},
		"no-subnet-dns-label": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-subnet-dns-label"),
			SubnetId:      common.String("subnetwithoutdnslabel"),
		},
		"no-vcn-dns-label": {
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("subnetwithnovcndnslabel"),
		},
	}

	instances = map[string]*core.Instance{
		"basic-complete": {
			CompartmentId: common.String("default"),
		},
		"no-external-ip": {
			CompartmentId: common.String("default"),
		},
		"no-internal-ip": {
			CompartmentId: common.String("default"),
		},
		"invalid-internal-ip": {
			CompartmentId: common.String("default"),
		},
		"invalid-external-ip": {
			CompartmentId: common.String("default"),
		},
		"no-hostname-label": {
			CompartmentId: common.String("default"),
		},
		"no-subnet-dns-label": {
			CompartmentId: common.String("default"),
		},
		"no-vcn-dns-label": {
			CompartmentId: common.String("default"),
		},
		"instance1": {
			CompartmentId: common.String("compartment1"),
			Id:            common.String("instance1"),
			Shape:         common.String("VM.Standard1.2"),
			DisplayName:   common.String("instance1"),
		},
		"instance_zone_test": {
			AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
			CompartmentId:      common.String("compartment1"),
			Id:                 common.String("instance_zone_test"),
			Region:             common.String("PHX"),
			Shape:              common.String("VM.Standard1.2"),
			DisplayName:        common.String("instance_zone_test"),
		},
	}
	subnets = map[string]*core.Subnet{
		"subnetwithdnslabel": {
			Id:       common.String("subnetwithdnslabel"),
			DnsLabel: common.String("subnetwithdnslabel"),
			VcnId:    common.String("vcnwithdnslabel"),
		},
		"subnetwithoutdnslabel": {
			Id:    common.String("subnetwithoutdnslabel"),
			VcnId: common.String("vcnwithdnslabel"),
		},
		"subnetwithnovcndnslabel": {
			Id:       common.String("subnetwithnovcndnslabel"),
			DnsLabel: common.String("subnetwithnovcndnslabel"),
			VcnId:    common.String("vcnwithoutdnslabel"),
		},
		"one": {
			Id:                 common.String("one"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD1"),
		},
		"two": {
			Id:                 common.String("two"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD2"),
		},
		"annotation-one": {
			Id:                 common.String("annotation-one"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD1"),
		},
		"annotation-two": {
			Id:                 common.String("annotation-two"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: common.String("AD2"),
		},
		"regional-subnet": {
			Id:                 common.String("regional-subnet"),
			DnsLabel:           common.String("subnetwithnovcndnslabel"),
			VcnId:              common.String("vcnwithoutdnslabel"),
			AvailabilityDomain: nil,
		},
	}

	vcns = map[string]*core.Vcn{
		"vcnwithdnslabel": {
			Id:       common.String("vcnwithdnslabel"),
			DnsLabel: common.String("vcnwithdnslabel"),
		},
		"vcnwithoutdnslabel": {
			Id: common.String("vcnwithoutdnslabel"),
		},
	}

	nodeList = map[string]*v1.Node{
		"default": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "default",
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "default",
			},
		},
		"instance1": {
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					CompartmentIDAnnotation: "compartment1",
				},
			},
			Spec: v1.NodeSpec{
				ProviderID: "instance1",
			},
		},
	}

	loadBalancers = map[string]*client.GenericLoadBalancer{
		"privateLB": {
			Id:          common.String("privateLB"),
			DisplayName: common.String("privateLB"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
		},
		"privateLB-no-IP": {
			Id:          common.String("privateLB-no-IP"),
			DisplayName: common.String("privateLB-no-IP"),
			IpAddresses: []client.GenericIpAddress{},
		},
		"test-uid": {
			Id:          common.String("test-uid"),
			DisplayName: common.String("test-uid"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
		},
		"test-uid-delete-err": {
			Id:          common.String("test-uid-delete-err"),
			DisplayName: common.String("test-uid-delete-err"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
		},
		"test-uid-node-err": {
			Id:          common.String("test-uid-delete-err"),
			DisplayName: common.String("test-uid-delete-err"),
			IpAddresses: []client.GenericIpAddress{
				{
					IpAddress: common.String("10.0.50.5"),
					IsPublic:  common.Bool(false),
				},
			},
			SubnetIds: []string{*subnets["one"].Id, *subnets["two"].Id},
			Listeners: map[string]client.GenericListener{
				"one": {
					Name:                  common.String("one"),
					DefaultBackendSetName: common.String("one"),
					Port:                  common.Int(5665),
				}},
			BackendSets: map[string]client.GenericBackendSetDetails{
				"one": {
					Backends: []client.GenericBackend{{
						Name:      common.String("one"),
						IpAddress: common.String("10.0.50.5"),
						Port:      common.Int(5665),
					}},
				}},
		},
	}
)

type MockSecurityListManager struct{}

func (MockSecurityListManager) Update(ctx context.Context, lbSubnets []*core.Subnet, _ []*core.Subnet, sourceCIDRs []string, actualPorts *portSpec, desiredPorts portSpec, isPreserveSource bool) error {
	return nil
}

func (MockSecurityListManager) Delete(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, ports portSpec, sourceCIDRs []string, isPreserveSource bool) error {
	return nil
}

type MockSecurityListManagerFactory func(mode string) MockSecurityListManager

type MockOCIClient struct{}

func (MockOCIClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

func (MockOCIClient) LoadBalancer(string) client.GenericLoadBalancerInterface {
	return &MockLoadBalancerClient{}
}

func (MockOCIClient) Networking() client.NetworkingInterface {
	return &MockVirtualNetworkClient{}
}

func (MockOCIClient) BlockStorage() client.BlockStorageInterface {
	return &MockBlockStorageClient{}
}

func (MockOCIClient) FSS() client.FileStorageInterface {
	return &MockFileStorageClient{}
}

func (MockOCIClient) Identity() client.IdentityInterface {
	return &MockIdentityClient{}
}

// MockComputeClient mocks Compute client implementation
type MockComputeClient struct{}

func (MockComputeClient) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	if instance, ok := instances[id]; ok {
		return instance, nil
	}
	return &core.Instance{
		AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
		CompartmentId:      common.String("default"),
		Id:                 &id,
		Region:             common.String("PHX"),
		Shape:              common.String("VM.Standard1.2"),
	}, nil
}

func (MockComputeClient) GetInstanceByNodeName(ctx context.Context, compartmentID, vcnID, nodeName string) (*core.Instance, error) {
	if instance, ok := instances[nodeName]; ok {
		return instance, nil
	}
	return &core.Instance{
		AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
		CompartmentId:      &compartmentID,
		Id:                 &nodeName,
		Region:             common.String("PHX"),
		Shape:              common.String("VM.Standard1.2"),
	}, nil
}

func (MockComputeClient) GetPrimaryVNICForInstance(ctx context.Context, compartmentID, instanceID string) (*core.Vnic, error) {
	return instanceVnics[instanceID], nil
}

func (MockComputeClient) FindVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) AttachParavirtualizedVolume(ctx context.Context, instanceID, volumeID string, isPvEncryptionInTransitEnabled bool) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) AttachVolume(ctx context.Context, instanceID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) WaitForVolumeAttached(ctx context.Context, attachmentID string) (core.VolumeAttachment, error) {
	return nil, nil
}

func (MockComputeClient) DetachVolume(ctx context.Context, id string) error {
	return nil
}

func (MockComputeClient) WaitForVolumeDetached(ctx context.Context, attachmentID string) error {
	return nil
}

func (c *MockComputeClient) FindActiveVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
	return nil, nil
}

// MockVirtualNetworkClient mocks VirtualNetwork client implementation
type MockVirtualNetworkClient struct {
}

func (c *MockVirtualNetworkClient) IsRegionalSubnet(ctx context.Context, id string) (bool, error) {
	return subnets[id].AvailabilityDomain == nil, nil
}

func (c *MockVirtualNetworkClient) GetPrivateIP(ctx context.Context, id string) (*core.PrivateIp, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	if subnet, ok := subnets[id]; ok {
		return subnet, nil
	}
	return nil, errors.New("Subnet not found")
}

func (c *MockVirtualNetworkClient) GetVcn(ctx context.Context, id string) (*core.Vcn, error) {
	return vcns[id], nil
}

func (c *MockVirtualNetworkClient) GetSubnetFromCacheByIP(ip string) (*core.Subnet, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSecurityList(ctx context.Context, id string) (core.GetSecurityListResponse, error) {
	return core.GetSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) UpdateSecurityList(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) (core.UpdateSecurityListResponse, error) {
	return core.UpdateSecurityListResponse{}, nil
}

func (c *MockVirtualNetworkClient) GetPublicIpByIpAddress(ctx context.Context, id string) (*core.PublicIp, error) {
	return nil, nil
}

//// MockFileStorageClient mocks FileStorage client implementation.
type MockLoadBalancerClient struct{}

func (c *MockLoadBalancerClient) CreateLoadBalancer(ctx context.Context, details *client.GenericCreateLoadBalancerDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetLoadBalancer(ctx context.Context, id string) (*client.GenericLoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) GetLoadBalancerByName(ctx context.Context, compartmentID string, name string) (*client.GenericLoadBalancer, error) {
	if lb, ok := loadBalancers[name]; ok {
		return lb, nil
	}
	return nil, nil
}

func (c *MockLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	if id == "test-uid-delete-err" {
		return "workReqId", errors.New("error")
	}
	return "", nil
}

func (c *MockLoadBalancerClient) GetCertificateByName(ctx context.Context, lbID string, name string) (*client.GenericCertificate, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) CreateCertificate(ctx context.Context, lbID string, cert *client.GenericCertificate) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateBackendSet(ctx context.Context, lbID string, name string, details *client.GenericBackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateBackendSet(ctx context.Context, lbID string, name string, details *client.GenericBackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteBackendSet(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateListener(ctx context.Context, lbID string, name string, details *client.GenericListener) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateListener(ctx context.Context, lbID string, name string, details *client.GenericListener) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteListener(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

var awaitLoadbalancerWorkrequestMap = map[string]error{
	"failedToGetUpdateNetworkSecurityGroupsWorkRequest": errors.New("internal server error for get workrequest call"),
}

func (c *MockLoadBalancerClient) AwaitWorkRequest(ctx context.Context, id string) (*client.GenericWorkRequest, error) {
	if err, ok := awaitLoadbalancerWorkrequestMap[id]; ok {
		return nil, err
	}
	return nil, nil
}

func (c *MockLoadBalancerClient) UpdateLoadBalancerShape(context.Context, string, *client.GenericUpdateLoadBalancerShapeDetails) (string, error) {
	return "", nil
}

var updateNetworkSecurityGroupsLBsFailures = map[string]error{
	"":                      errors.New("provided LB ID is empty"),
	"failedToCreateRequest": errors.New("internal server error"),
}
var updateNetworkSecurityGroupsLBsWorkRequests = map[string]string{
	"failedToGetUpdateNetworkSecurityGroupsWorkRequest": "failedToGetUpdateNetworkSecurityGroupsWorkRequest",
}

func (c *MockLoadBalancerClient) UpdateNetworkSecurityGroups(ctx context.Context, lbId string, nsgIds []string) (string, error) {
	if err, ok := updateNetworkSecurityGroupsLBsFailures[lbId]; ok {
		return "", err
	}
	if wrID, ok := updateNetworkSecurityGroupsLBsWorkRequests[lbId]; ok {
		return wrID, nil
	}
	return "", nil
}

// MockBlockStorageClient mocks BlockStorage client implementation
type MockBlockStorageClient struct{}

func (MockBlockStorageClient) AwaitVolumeAvailableORTimeout(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error) {
	return nil, nil
}

func (c MockBlockStorageClient) UpdateVolume(ctx context.Context, volumeId string, details core.UpdateVolumeDetails) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) GetVolume(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) GetVolumesByName(ctx context.Context, volumeName, compartmentID string) ([]core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) DeleteVolume(ctx context.Context, id string) error {
	return nil
}

// MockFileStorageClient mocks FileStorage client implementation.
type MockFileStorageClient struct{}

func (MockFileStorageClient) AwaitMountTargetActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.MountTarget, error) {
	return nil, nil
}

func (MockFileStorageClient) GetFileSystem(ctx context.Context, id string) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (MockFileStorageClient) GetFileSystemSummaryByDisplayName(ctx context.Context, compartmentID, ad, displayName string) (*filestorage.FileSystemSummary, error) {
	return nil, nil
}

func (MockFileStorageClient) AwaitFileSystemActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (MockFileStorageClient) CreateFileSystem(ctx context.Context, details filestorage.CreateFileSystemDetails) (*filestorage.FileSystem, error) {
	return nil, nil
}

func (MockFileStorageClient) DeleteFileSystem(ctx context.Context, id string) error {
	return nil
}

func (MockFileStorageClient) CreateExport(ctx context.Context, details filestorage.CreateExportDetails) (*filestorage.Export, error) {
	return nil, nil
}

func (MockFileStorageClient) FindExport(ctx context.Context, compartmentID, fsID, exportSetID string) (*filestorage.ExportSummary, error) {
	return nil, nil
}

func (MockFileStorageClient) AwaitExportActive(ctx context.Context, logger *zap.SugaredLogger, id string) (*filestorage.Export, error) {
	return nil, nil
}

func (MockFileStorageClient) DeleteExport(ctx context.Context, id string) error {
	return nil
}

// MockIdentityClient mocks Identity client implementaion
type MockIdentityClient struct{}

func (MockIdentityClient) GetAvailabilityDomainByName(ctx context.Context, compartmentID, name string) (*identity.AvailabilityDomain, error) {
	return nil, nil
}

func (MockIdentityClient) ListAvailabilityDomains(ctx context.Context, compartmentID string) ([]identity.AvailabilityDomain, error) {
	return nil, nil
}

type mockInstanceCache struct{}

func (m mockInstanceCache) Add(obj interface{}) error {
	return nil
}

func (m mockInstanceCache) Update(obj interface{}) error {
	return nil
}

func (m mockInstanceCache) Delete(obj interface{}) error {
	return nil
}

func (m mockInstanceCache) List() []interface{} {
	return nil
}

func (m mockInstanceCache) ListKeys() []string {
	return nil
}

func (m mockInstanceCache) Get(obj interface{}) (item interface{}, exists bool, err error) {
	return instances["default"], true, nil
}

func (m mockInstanceCache) GetByKey(key string) (item interface{}, exists bool, err error) {
	if instance, ok := instances[key]; ok {
		return instance, true, nil
	}
	return nil, false, nil
}

func (m mockInstanceCache) Replace(i []interface{}, s string) error {
	return nil
}

func (m mockInstanceCache) Resync() error {
	return nil
}

func TestExtractNodeAddresses(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  []v1.NodeAddress
		err  error
	}{
		{
			name: "basic-complete",
			in:   "basic-complete",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "basic-complete.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "basic-complete.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "no-external-ip",
			in:   "no-external-ip",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "no-external-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "no-external-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "no-internal-ip",
			in:   "no-internal-ip",
			out: []v1.NodeAddress{
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "no-internal-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "no-internal-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "invalid-external-ip",
			in:   "invalid-external-ip",
			out:  nil,
			err:  errors.New(`instance has invalid public address: "0.0.0."`),
		},
		{
			name: "invalid-internal-ip",
			in:   "invalid-internal-ip",
			out:  nil,
			err:  errors.New(`instance has invalid private address: "10.0.0."`),
		},
		{
			name: "no-hostname-label",
			in:   "no-hostname-label",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-subnet-dns-label",
			in:   "no-subnet-dns-label",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-vcn-dns-label",
			in:   "no-vcn-dns-label",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
	}

	cp := &CloudProvider{
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister:    &mockNodeLister{},
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.extractNodeAddresses(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("extractNodeAddresses(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("extractNodeAddresses(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceID(t *testing.T) {
	testCases := []struct {
		name string
		in   types.NodeName
		out  string
		err  error
	}{
		{
			name: "get instance id from instance in the cache",
			in:   "instance1",
			out:  "instance1",
			err:  nil,
		},
		{
			name: "get instance id from instance not in the cache",
			in:   "default",
			out:  "default",
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceType(t *testing.T) {
	testCases := []struct {
		name string
		in   types.NodeName
		out  string
		err  error
	}{
		{
			name: "check node shape of instance in cache",
			in:   "instance1",
			out:  "VM.Standard1.2",
			err:  nil,
		},
		{
			name: "check node shape of instance not in cache",
			in:   "default",
			out:  "VM.Standard1.2",
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceType(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceType(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceType(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceTypeByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  string
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "instance1",
			out:  "VM.Standard1.2",
			err:  nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "instance1",
			out:  "VM.Standard1.2",
			err:  nil,
		},
		{
			name: "provider id with provider prefix and instance not in cache",
			in:   providerPrefix + "noncacheinstance",
			out:  "VM.Standard1.2",
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceTypeByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceTypeByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceTypeByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestNodeAddressesByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  []v1.NodeAddress
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "basic-complete",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "basic-complete",
			out: []v1.NodeAddress{
				{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.NodeAddressesByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("NodeAddressesByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("NodeAddressesByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceExistsByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  bool
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "instance1",
			out:  true,
			err:  nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "instance1",
			out:  true,
			err:  nil,
		},
		{
			name: "provider id with provider prefix and instance not in cache",
			in:   providerPrefix + "noncacheinstance",
			out:  true,
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceExistsByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceExistsByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceExistsByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestInstanceShutdownByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  bool
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "instance1",
			out:  false,
			err:  nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "instance1",
			out:  false,
			err:  nil,
		},
		{
			name: "provider id with provider prefix and instance not in cache",
			in:   providerPrefix + "noncacheinstance",
			out:  false,
			err:  nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.InstanceShutdownByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("InstanceShutdownByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("InstanceShutdownByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestGetCompartmentIDByInstanceID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  string
		err  error
	}{
		{
			name: "instance found in cache",
			in:   "instance1",
			out:  "compartment1",
			err:  nil,
		},
		{
			name: "instance found in node lister",
			in:   "default",
			out:  "default",
			err:  nil,
		},
		{
			name: "instance neither found in cache nor node lister",
			in:   "instancex",
			out:  "",
			err:  errors.New("compartmentID annotation missing in the node. Would retry"),
		},
	}

	cp := &CloudProvider{
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister:    &mockNodeLister{},
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.getCompartmentIDByInstanceID(tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("getCompartmentIDByInstanceID(%s) got error %s, expected %s", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("getCompartmentIDByInstanceID(%s) => %s, want %s", tt.in, result, tt.out)
			}
		})
	}
}

type mockNodeLister struct{}

func (s *mockNodeLister) List(selector labels.Selector) (ret []*v1.Node, err error) {
	nodes := make([]*v1.Node, len(nodeList))
	nodes[0] = nodeList["default"]
	nodes[1] = nodeList["instance1"]
	return nodes, nil
}

func (s *mockNodeLister) Get(name string) (*v1.Node, error) {
	if node, ok := nodeList[name]; ok {
		return node, nil
	}
	return nil, nil
}

func (s *mockNodeLister) ListWithPredicate() ([]*v1.Node, error) {
	return nil, nil
}
