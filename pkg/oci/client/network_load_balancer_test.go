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
	"errors"
	"fmt"
	errors2 "github.com/pkg/errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/networkloadbalancer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/flowcontrol"
)

func TestNLB_AwaitWorkRequest(t *testing.T) {
	var tests = map[string]struct {
		skip         bool // set true to skip a test-case
		loadbalancer networkLoadbalancer
		wantErr      error
	}{
		"getWorkRequestTimedOut": {
			skip: true,
			loadbalancer: networkLoadbalancer{
				networkloadbalancer: &MockNetworkLoadBalancerClient{debug: true}, // set true to run test with debug logs
				requestMetadata:     common.RequestMetadata{},
				rateLimiter: RateLimiter{
					Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
					Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
				},
			},
			wantErr: wait.ErrWaitTimeout,
		},
		"getWorkRequestTimedOutOnce": {
			loadbalancer: networkLoadbalancer{
				networkloadbalancer: &MockNetworkLoadBalancerClient{debug: false},
				requestMetadata:     common.RequestMetadata{},
				rateLimiter: RateLimiter{
					Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
					Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
				},
			},
			wantErr: TestNonRetryableError,
		},
		"getWorkRequestTimedOutOnceWrappedError": {
			loadbalancer: networkLoadbalancer{
				networkloadbalancer: &MockNetworkLoadBalancerClient{debug: false},
				requestMetadata:     common.RequestMetadata{},
				rateLimiter: RateLimiter{
					Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
					Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
				},
			},
			wantErr: TestNonRetryableError,
		},
		"getWorkRequestSuccess": {
			loadbalancer: networkLoadbalancer{
				networkloadbalancer: &MockNetworkLoadBalancerClient{debug: false},
				requestMetadata:     common.RequestMetadata{},
				rateLimiter: RateLimiter{
					Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
					Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
				},
			},
			wantErr: nil,
		},
	}

	t.Parallel() // test will run in parallel with others
	for name, tt := range tests {
		if !tt.skip {
			tt, name := tt, name // new local variables necessary for proper parallel execution of test cases
			t.Run(name, func(t *testing.T) {
				t.Parallel() // test case will run in parallel with others
				_, err := tt.loadbalancer.AwaitWorkRequest(context.Background(), name)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Expected error = %v, but got %v", tt.wantErr, err)
					return
				}
			})
		}
	}
}

var (
	fakeNlbOcid1   = "ocid.nlb.fake1"
	fakeNlbName1   = "fake display name 1"
	fakeNlbOcid2   = "ocid.nlb.fake2"
	fakeNlbName2   = "fake display name 2"
	fakeSubnetOcid = "ocid.subnet.fake"

	NLBMap = map[string]networkloadbalancer.NetworkLoadBalancer{
		"ocid.nlb.fake1": networkloadbalancer.NetworkLoadBalancer{
			Id:          &fakeNlbOcid1,
			DisplayName: &fakeNlbName1,
			SubnetId:    &fakeSubnetOcid,
		},
		"ocid.nlb.fake2": networkloadbalancer.NetworkLoadBalancer{
			Id:          &fakeNlbOcid2,
			DisplayName: &fakeNlbName2,
			SubnetId:    &fakeSubnetOcid,
		},
	}
)

func TestGetLoadBalancerByName(t *testing.T) {
	var totalListCalls int
	var loadbalancer = NewNLBClient(
		&MockNetworkLoadBalancerClient{debug: true, listCalls: &totalListCalls},
		common.RequestMetadata{},
		&RateLimiter{
			Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
			Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
		})

	var tests = []struct {
		skip                        bool // set true to skip a test-case
		compartment, name, testname string
		want                        string
		wantErr                     error
		wantListCalls               int
	}{
		{
			testname:      "getFirstNLBFirstTime",
			compartment:   "ocid.compartment.fake",
			name:          fakeNlbName1,
			want:          fakeNlbOcid1,
			wantErr:       nil,
			wantListCalls: 1,
		},
		{
			testname:      "getFirstNLBSecondTime",
			compartment:   "ocid.compartment.fake",
			name:          fakeNlbName1,
			want:          fakeNlbOcid1,
			wantErr:       nil,
			wantListCalls: 1, // totals, no new list should be performed
		},
		{
			testname:      "getSecondNLBTime",
			compartment:   "ocid.compartment.fake",
			name:          fakeNlbName2,
			want:          fakeNlbOcid2,
			wantErr:       nil,
			wantListCalls: 2,
		},
		{
			testname:      "getFirstNLBThirdTime",
			compartment:   "ocid.compartment.fake",
			name:          fakeNlbName1,
			want:          fakeNlbOcid1,
			wantErr:       nil,
			wantListCalls: 2,
		},
		{
			testname:      "getSecondNLBSecondTime",
			compartment:   "ocid.compartment.fake",
			name:          fakeNlbName2,
			want:          fakeNlbOcid2,
			wantErr:       nil,
			wantListCalls: 2,
		},
	}

	for _, tt := range tests {
		if tt.skip {
			continue
		}

		t.Run(tt.testname, func(t *testing.T) {
			log.Println("running test ", tt.testname)
			got, err := loadbalancer.GetLoadBalancerByName(context.Background(), tt.compartment, tt.name)
			if got == nil || *got.Id != tt.want {
				t.Errorf("Expected %v, but got %v", tt.want, got)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, but got %v", tt.wantErr, err)
			}
			if totalListCalls != tt.wantListCalls {
				t.Errorf("Expected the total number of NLB list calls %d, but got %d", tt.wantListCalls, totalListCalls)
			}
		})
	}
}

type MockNetworkLoadBalancerClient struct {
	// MockLoadBalancerClient mocks LoadBalancer client implementation.
	counter   int
	debug     bool // set true to run tests with debug logs
	listCalls *int // number of list operations performed
}

type getNetworkLoadBalancerWorkRequestResponse struct {
	response networkloadbalancer.GetWorkRequestResponse
	err      error
}

var getNetworkLoadbalancerWorkRequestMap = map[string]getNetworkLoadBalancerWorkRequestResponse{
	"getWorkRequestTimedOut": {
		err: context.DeadlineExceeded,
	},
	"getWorkRequestTimedOutOnce": {
		response: networkloadbalancer.GetWorkRequestResponse{
			WorkRequest: networkloadbalancer.WorkRequest{
				Status:        networkloadbalancer.OperationStatusInProgress,
				OperationType: networkloadbalancer.OperationTypeCreateNetworkLoadBalancer,
			},
		},
		err: context.DeadlineExceeded,
	},
	"getWorkRequestTimedOutOnceWrappedError": {
		response: networkloadbalancer.GetWorkRequestResponse{
			WorkRequest: networkloadbalancer.WorkRequest{
				Status:        networkloadbalancer.OperationStatusInProgress,
				OperationType: networkloadbalancer.OperationTypeCreateNetworkLoadBalancer,
			},
		},
		err: errors2.Wrap(errors2.Wrap(errors2.WithStack(context.DeadlineExceeded), "Bar"), "Foo"),
	},
	"getWorkRequestNonRetryable": {
		err: TestNonRetryableError,
	},
	"getWorkRequestSuccess": {
		response: networkloadbalancer.GetWorkRequestResponse{
			WorkRequest: networkloadbalancer.WorkRequest{
				Status:        networkloadbalancer.OperationStatusSucceeded,
				OperationType: networkloadbalancer.OperationTypeCreateNetworkLoadBalancer,
			},
		},
		err: nil,
	},
}

func (c *MockNetworkLoadBalancerClient) GetWorkRequest(ctx context.Context, request networkloadbalancer.GetWorkRequestRequest) (response networkloadbalancer.GetWorkRequestResponse, err error) {
	if resp, ok := getNetworkLoadbalancerWorkRequestMap[*request.WorkRequestId]; ok {
		if c.debug {
			log.Println(resp.err)
		}
		if *request.WorkRequestId == "getWorkRequestTimedOut" {
			time.Sleep(defaultSynchronousAPIContextTimeout)
		}
		if strings.Contains(*request.WorkRequestId, "getWorkRequestTimedOutOnce") {
			c.counter += 1
			if c.counter == 1 {
				time.Sleep(defaultSynchronousAPIContextTimeout)
				return resp.response, resp.err
			} else {
				return resp.response, getNetworkLoadbalancerWorkRequestMap["getWorkRequestNonRetryable"].err
			}
		}
		return resp.response, resp.err
	}
	return networkloadbalancer.GetWorkRequestResponse{
		WorkRequest: networkloadbalancer.WorkRequest{
			Status:        networkloadbalancer.OperationStatusSucceeded,
			OperationType: networkloadbalancer.OperationTypeCreateNetworkLoadBalancer,
		},
	}, nil
}

func (c *MockNetworkLoadBalancerClient) GetNetworkLoadBalancer(ctx context.Context, request networkloadbalancer.GetNetworkLoadBalancerRequest) (response networkloadbalancer.GetNetworkLoadBalancerResponse, err error) {
	if c.debug {
		log.Println(fmt.Sprintf("Getting NLB %v", *request.NetworkLoadBalancerId))
	}

	response = networkloadbalancer.GetNetworkLoadBalancerResponse{
		NetworkLoadBalancer: NLBMap[*request.NetworkLoadBalancerId],
	}
	return
}
func (c *MockNetworkLoadBalancerClient) ListWorkRequests(ctx context.Context, request networkloadbalancer.ListWorkRequestsRequest) (response networkloadbalancer.ListWorkRequestsResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) ListNetworkLoadBalancers(ctx context.Context, request networkloadbalancer.ListNetworkLoadBalancersRequest) (response networkloadbalancer.ListNetworkLoadBalancersResponse, err error) {
	if c.debug {
		log.Println(fmt.Sprintf("Lising NLBs in compartment %v", *request.CompartmentId))
	}

	for _, nlb := range NLBMap {
		response.NetworkLoadBalancerCollection.Items = append(response.NetworkLoadBalancerCollection.Items, networkloadbalancer.NetworkLoadBalancerSummary(nlb))
	}
	*c.listCalls += 1
	return
}
func (c *MockNetworkLoadBalancerClient) CreateNetworkLoadBalancer(ctx context.Context, request networkloadbalancer.CreateNetworkLoadBalancerRequest) (response networkloadbalancer.CreateNetworkLoadBalancerResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) DeleteNetworkLoadBalancer(ctx context.Context, request networkloadbalancer.DeleteNetworkLoadBalancerRequest) (response networkloadbalancer.DeleteNetworkLoadBalancerResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) CreateBackendSet(ctx context.Context, request networkloadbalancer.CreateBackendSetRequest) (response networkloadbalancer.CreateBackendSetResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) UpdateBackendSet(ctx context.Context, request networkloadbalancer.UpdateBackendSetRequest) (response networkloadbalancer.UpdateBackendSetResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) DeleteBackendSet(ctx context.Context, request networkloadbalancer.DeleteBackendSetRequest) (response networkloadbalancer.DeleteBackendSetResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) CreateListener(ctx context.Context, request networkloadbalancer.CreateListenerRequest) (response networkloadbalancer.CreateListenerResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) UpdateListener(ctx context.Context, request networkloadbalancer.UpdateListenerRequest) (response networkloadbalancer.UpdateListenerResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) DeleteListener(ctx context.Context, request networkloadbalancer.DeleteListenerRequest) (response networkloadbalancer.DeleteListenerResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) UpdateNetworkSecurityGroups(ctx context.Context, request networkloadbalancer.UpdateNetworkSecurityGroupsRequest) (response networkloadbalancer.UpdateNetworkSecurityGroupsResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) UpdateNetworkLoadBalancer(ctx context.Context, request networkloadbalancer.UpdateNetworkLoadBalancerRequest) (response networkloadbalancer.UpdateNetworkLoadBalancerResponse, err error) {
	return
}
