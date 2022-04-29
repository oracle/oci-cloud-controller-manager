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
	"testing"

	"github.com/oracle/oci-go-sdk/v50/core"
	"k8s.io/client-go/util/flowcontrol"
)

func TestInstanceTerminalState(t *testing.T) {
	testCases := map[string]struct {
		state    core.InstanceLifecycleStateEnum
		expected bool
	}{
		"not terminal - running": {
			state:    core.InstanceLifecycleStateRunning,
			expected: false,
		},
		"not terminal - stopped": {
			state:    core.InstanceLifecycleStateStopped,
			expected: false,
		},
		"is terminal - terminating": {
			state:    core.InstanceLifecycleStateTerminating,
			expected: true,
		},
		"is terminal - terminated": {
			state:    core.InstanceLifecycleStateTerminated,
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := IsInstanceInTerminalState(&core.Instance{
				LifecycleState: tc.state,
			})
			if result != tc.expected {
				t.Errorf("IsInstanceInTerminalState(%q) = %v ; wanted %v", tc.state, result, tc.expected)
			}
		})
	}
}

func TestInstanceStoppedState(t *testing.T) {
	testCases := map[string]struct {
		state    core.InstanceLifecycleStateEnum
		expected bool
	}{
		"not stopped - running": {
			state:    core.InstanceLifecycleStateRunning,
			expected: false,
		},
		"not stopped - terminated": {
			state:    core.InstanceLifecycleStateTerminated,
			expected: false,
		},
		"is stopped - stopped": {
			state:    core.InstanceLifecycleStateStopped,
			expected: true,
		},
		"is stopped - stopping": {
			state:    core.InstanceLifecycleStateStopping,
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := IsInstanceInStoppedState(&core.Instance{
				LifecycleState: tc.state,
			})
			if result != tc.expected {
				t.Errorf("IsInstanceInStoppedState(%q) = %v ; wanted %v", tc.state, result, tc.expected)
			}
		})
	}
}

type mockComputeClient struct{}

type mockVirtualNetworkClient struct{}

func TestRateLimiting(t *testing.T) {
	var qpsRead float32 = 5
	bucketRead := 5
	var qpsWrite float32 = 10
	bucketWrite := 5

	rateLimiter := RateLimiter{
		Reader: flowcontrol.NewTokenBucketRateLimiter(
			qpsRead,
			bucketRead),
		Writer: flowcontrol.NewTokenBucketRateLimiter(
			qpsWrite,
			bucketWrite),
	}

	client := newClient(rateLimiter)

	// Read requests up to qpsRead should pass and the others fail
	for i := 0; i < int(qpsRead)*2; i++ {
		_, err := client.Compute().GetInstance(context.Background(), "123345")
		p := (err == nil)

		if (i < int(qpsRead) && !p) || (i >= int(qpsRead) && p) {
			t.Errorf("unexpected result from request %d: %v", i, err)
		}
	}

	// Write requests up to qpsWrite should pass and the others fail
	for i := 0; i < int(qpsWrite)*2; i++ {

		_, err := client.Networking().UpdateSecurityList(context.Background(), "1234", "", nil, nil)
		p := (err == nil)

		if (i < int(qpsRead) && !p) || (i >= int(qpsRead) && p) {
			t.Errorf("unexpected result from request %d: %v", i, err)
		}
	}
}

func newClient(rateLimiter RateLimiter) Interface {
	return &client{
		compute:     &mockComputeClient{},
		network:     &mockVirtualNetworkClient{},
		rateLimiter: rateLimiter,
	}
}

/* Mock ComputeClient interface implementations */
func (c *mockComputeClient) GetInstance(ctx context.Context, request core.GetInstanceRequest) (response core.GetInstanceResponse, err error) {
	return core.GetInstanceResponse{}, nil
}

func (c *mockComputeClient) ListInstances(ctx context.Context, request core.ListInstancesRequest) (response core.ListInstancesResponse, err error) {
	return core.ListInstancesResponse{}, nil
}

func (c *mockComputeClient) ListVnicAttachments(ctx context.Context, request core.ListVnicAttachmentsRequest) (response core.ListVnicAttachmentsResponse, err error) {
	return core.ListVnicAttachmentsResponse{}, nil
}

func (c *mockComputeClient) GetVolumeAttachment(ctx context.Context, request core.GetVolumeAttachmentRequest) (response core.GetVolumeAttachmentResponse, err error) {
	return core.GetVolumeAttachmentResponse{}, nil
}

func (c *mockComputeClient) ListVolumeAttachments(ctx context.Context, request core.ListVolumeAttachmentsRequest) (response core.ListVolumeAttachmentsResponse, err error) {
	return core.ListVolumeAttachmentsResponse{}, nil
}

func (c *mockComputeClient) AttachVolume(ctx context.Context, request core.AttachVolumeRequest) (response core.AttachVolumeResponse, err error) {
	return core.AttachVolumeResponse{}, nil
}

func (c *mockComputeClient) DetachVolume(ctx context.Context, request core.DetachVolumeRequest) (response core.DetachVolumeResponse, err error) {
	return core.DetachVolumeResponse{}, nil
}

func (c *mockComputeClient) ListInstanceDevices(ctx context.Context, request core.ListInstanceDevicesRequest) (response core.ListInstanceDevicesResponse, err error) {
	return core.ListInstanceDevicesResponse{}, nil
}

/* Mock NetworkClient interface implementations */
func (c *mockVirtualNetworkClient) GetVnic(ctx context.Context, request core.GetVnicRequest) (response core.GetVnicResponse, err error) {
	return core.GetVnicResponse{}, nil
}

func (c *mockVirtualNetworkClient) GetSubnet(ctx context.Context, request core.GetSubnetRequest) (response core.GetSubnetResponse, err error) {
	return core.GetSubnetResponse{}, nil
}

func (c *mockVirtualNetworkClient) GetVcn(ctx context.Context, request core.GetVcnRequest) (response core.GetVcnResponse, err error) {
	return core.GetVcnResponse{}, nil
}

func (c *mockVirtualNetworkClient) GetSecurityList(ctx context.Context, request core.GetSecurityListRequest) (response core.GetSecurityListResponse, err error) {
	return core.GetSecurityListResponse{}, nil
}

func (c *mockVirtualNetworkClient) UpdateSecurityList(ctx context.Context, request core.UpdateSecurityListRequest) (response core.UpdateSecurityListResponse, err error) {
	return core.UpdateSecurityListResponse{}, nil
}

func (c *mockVirtualNetworkClient) GetPrivateIp(ctx context.Context, request core.GetPrivateIpRequest) (response core.GetPrivateIpResponse, err error) {
	return core.GetPrivateIpResponse{}, nil
}

func (c *mockVirtualNetworkClient) GetPublicIpByIpAddress(ctx context.Context, request core.GetPublicIpByIpAddressRequest) (response core.GetPublicIpByIpAddressResponse, err error) {
	return core.GetPublicIpByIpAddressResponse{}, nil
}
