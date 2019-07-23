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
	"time"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/filestorage"
	"github.com/oracle/oci-go-sdk/identity"
	"github.com/oracle/oci-go-sdk/loadbalancer"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
)

var (
	instances = map[string]*core.Vnic{
		"basic-complete": &core.Vnic{
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("basic-complete"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"no-external-ip": &core.Vnic{
			PrivateIp:     common.String("10.0.0.1"),
			HostnameLabel: common.String("no-external-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"no-internal-ip": &core.Vnic{
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-internal-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"invalid-internal-ip": &core.Vnic{
			PrivateIp:     common.String("10.0.0."),
			HostnameLabel: common.String("no-internal-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"invalid-external-ip": &core.Vnic{
			PublicIp:      common.String("0.0.0."),
			HostnameLabel: common.String("invalid-external-ip"),
			SubnetId:      common.String("subnetwithdnslabel"),
		},
		"no-hostname-label": &core.Vnic{
			PrivateIp: common.String("10.0.0.1"),
			PublicIp:  common.String("0.0.0.1"),
			SubnetId:  common.String("subnetwithdnslabel"),
		},
		"no-subnet-dns-label": &core.Vnic{
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-subnet-dns-label"),
			SubnetId:      common.String("subnetwithoutdnslabel"),
		},
		"no-vcn-dns-label": &core.Vnic{
			PrivateIp:     common.String("10.0.0.1"),
			PublicIp:      common.String("0.0.0.1"),
			HostnameLabel: common.String("no-vcn-dns-label"),
			SubnetId:      common.String("subnetwithnovcndnslabel"),
		},
	}

	subnets = map[string]*core.Subnet{
		"subnetwithdnslabel": &core.Subnet{
			Id:       common.String("subnetwithdnslabel"),
			DnsLabel: common.String("subnetwithdnslabel"),
			VcnId:    common.String("vcnwithdnslabel"),
		},
		"subnetwithoutdnslabel": &core.Subnet{
			Id:    common.String("subnetwithoutdnslabel"),
			VcnId: common.String("vcnwithdnslabel"),
		},
		"subnetwithnovcndnslabel": &core.Subnet{
			Id:       common.String("subnetwithnovcndnslabel"),
			DnsLabel: common.String("subnetwithnovcndnslabel"),
			VcnId:    common.String("vcnwithoutdnslabel"),
		},
	}

	vcns = map[string]*core.Vcn{
		"vcnwithdnslabel": &core.Vcn{
			Id:       common.String("vcnwithdnslabel"),
			DnsLabel: common.String("vcnwithdnslabel"),
		},
		"vcnwithoutdnslabel": &core.Vcn{
			Id: common.String("vcnwithoutdnslabel"),
		},
	}
)

type MockOCIClient struct{}

func (MockOCIClient) Compute() client.ComputeInterface {
	return &MockComputeClient{}
}

func (MockOCIClient) LoadBalancer() client.LoadBalancerInterface {
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
	return nil, nil
}

func (MockComputeClient) GetInstanceByNodeName(ctx context.Context, compartmentID, vcnID, nodeName string) (*core.Instance, error) {
	return nil, nil
}

func (MockComputeClient) GetPrimaryVNICForInstance(ctx context.Context, compartmentID, instanceID string) (*core.Vnic, error) {
	return instances[instanceID], nil
}

func (MockComputeClient) FindVolumeAttachment(ctx context.Context, compartmentID, volumeID string) (core.VolumeAttachment, error) {
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

// MockVirtualNetworkClient mocks VirtualNetwork client implementation
type MockVirtualNetworkClient struct {
}

func (c *MockVirtualNetworkClient) GetPrivateIP(ctx context.Context, id string) (*core.PrivateIp, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetSubnet(ctx context.Context, id string) (*core.Subnet, error) {
	return subnets[id], nil
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

func (c *MockVirtualNetworkClient) UpdateSecurityList(ctx context.Context, request core.UpdateSecurityListRequest) (core.UpdateSecurityListResponse, error) {
	return core.UpdateSecurityListResponse{}, nil
}

//// MockFileStorageClient mocks FileStorage client implementation.
type MockLoadBalancerClient struct{}

func (c *MockLoadBalancerClient) CreateLoadBalancer(ctx context.Context, details loadbalancer.CreateLoadBalancerDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetLoadBalancer(ctx context.Context, id string) (*loadbalancer.LoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) GetLoadBalancerByName(ctx context.Context, compartmentID, name string) (*loadbalancer.LoadBalancer, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) GetCertificateByName(ctx context.Context, lbID, name string) (*loadbalancer.Certificate, error) {
	return nil, nil
}

func (c *MockLoadBalancerClient) CreateCertificate(ctx context.Context, lbID string, cert loadbalancer.CertificateDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateBackendSet(ctx context.Context, lbID, name string, details loadbalancer.BackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateBackendSet(ctx context.Context, lbID, name string, details loadbalancer.BackendSetDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteBackendSet(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateBackend(ctx context.Context, lbID, bsName string, details loadbalancer.BackendDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteBackend(ctx context.Context, lbID, bsName, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) UpdateListener(ctx context.Context, lbID, name string, details loadbalancer.ListenerDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) CreateListener(ctx context.Context, lbID, name string, details loadbalancer.ListenerDetails) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) DeleteListener(ctx context.Context, lbID, name string) (string, error) {
	return "", nil
}

func (c *MockLoadBalancerClient) AwaitWorkRequest(ctx context.Context, id string) (*loadbalancer.WorkRequest, error) {
	return nil, nil
}

// MockBlockStorageClient mocks BlockStoargae client implementation
type MockBlockStorageClient struct{}

func (MockBlockStorageClient) AwaitVolumeAvailable(ctx context.Context, id string) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) AwaitVolumeAvailableORTimeout(ctx context.Context, id string, timeout time.Duration) (*core.Volume, error) {
	return nil, nil
}

func (MockBlockStorageClient) CreateVolume(ctx context.Context, details core.CreateVolumeDetails) (*core.Volume, error) {
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
				v1.NodeAddress{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				v1.NodeAddress{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "basic-complete.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "basic-complete.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "no-external-ip",
			in:   "no-external-ip",
			out: []v1.NodeAddress{
				v1.NodeAddress{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				// v1.NodeAddress{Type: v1.NodeHostName, Address: "no-external-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
				// v1.NodeAddress{Type: v1.NodeInternalDNS, Address: "no-external-ip.subnetwithdnslabel.vcnwithdnslabel.oraclevcn.com"},
			},
			err: nil,
		},
		{
			name: "no-internal-ip",
			in:   "no-internal-ip",
			out: []v1.NodeAddress{
				v1.NodeAddress{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
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
				v1.NodeAddress{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				v1.NodeAddress{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-subnet-dns-label",
			in:   "no-subnet-dns-label",
			out: []v1.NodeAddress{
				v1.NodeAddress{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				v1.NodeAddress{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-vcn-dns-label",
			in:   "no-vcn-dns-label",
			out: []v1.NodeAddress{
				v1.NodeAddress{Type: v1.NodeInternalIP, Address: "10.0.0.1"},
				v1.NodeAddress{Type: v1.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
	}

	cp := &CloudProvider{
		client: MockOCIClient{},
		config: &providercfg.Config{CompartmentID: "testCompartment"},
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
