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
	errors2 "github.com/pkg/errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/flowcontrol"
)

var TestNonRetryableError = errors.New("some non-retryable error")

func TestLB_AwaitWorkRequest(t *testing.T) {
	var tests = map[string]struct {
		skip         bool // set true to skip a test-case
		loadbalancer loadbalancerClientStruct
		wantErr      error
	}{
		"getWorkRequestTimedOut": {
			skip: true,
			loadbalancer: loadbalancerClientStruct{
				loadbalancer:    &MockLoadBalancerClient{debug: true}, // set true to run test with debug logs
				requestMetadata: common.RequestMetadata{},
				rateLimiter: RateLimiter{
					Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
					Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
				},
			},
			wantErr: wait.ErrWaitTimeout,
		},
		"getWorkRequestTimedOutOnce": {
			loadbalancer: loadbalancerClientStruct{
				loadbalancer:    &MockLoadBalancerClient{debug: false},
				requestMetadata: common.RequestMetadata{},
				rateLimiter: RateLimiter{
					Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
					Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
				},
			},
			wantErr: TestNonRetryableError,
		},
		"getWorkRequestTimedOutOnceWrappedError": {
			loadbalancer: loadbalancerClientStruct{
				loadbalancer:    &MockLoadBalancerClient{debug: false},
				requestMetadata: common.RequestMetadata{},
				rateLimiter: RateLimiter{
					Reader: flowcontrol.NewFakeAlwaysRateLimiter(),
					Writer: flowcontrol.NewFakeAlwaysRateLimiter(),
				},
			},
			wantErr: TestNonRetryableError,
		},
		"getWorkRequestSuccess": {
			loadbalancer: loadbalancerClientStruct{
				loadbalancer:    &MockLoadBalancerClient{debug: false},
				requestMetadata: common.RequestMetadata{},
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
	fakeLbCompartment = "ocid1.compartment.totally.fake"
	fakeLbOcid1       = "ocid.lb.fake1"
	fakeLbName1       = "fake display name 1"
	fakeLbOcid2       = "ocid.lb.fake2"
	fakeLbName2       = "fake display name 2"
	fakeLbSubnetOcid  = "ocid.subnet.fake"

	LBMap = map[string]loadbalancer.LoadBalancer{
		fakeLbOcid1: loadbalancer.LoadBalancer{
			Id:            &fakeLbOcid1,
			CompartmentId: &fakeLbCompartment,
			DisplayName:   &fakeLbName1,
			SubnetIds:     []string{fakeLbSubnetOcid},
		},
		fakeLbOcid2: loadbalancer.LoadBalancer{
			Id:            &fakeLbOcid2,
			CompartmentId: &fakeLbCompartment,
			DisplayName:   &fakeLbName2,
			SubnetIds:     []string{fakeLbSubnetOcid},
		},
	}
)

func TestGetLoadBalancerByName(t *testing.T) {
	var totalListCalls int
	var lbClient = NewLBClient(
		&MockLoadBalancerClient{debug: true, listCalls: &totalListCalls},
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
			testname:      "getLBFirstTime",
			compartment:   fakeLbCompartment,
			name:          fakeLbName1,
			want:          fakeLbOcid1,
			wantErr:       nil,
			wantListCalls: 1,
		},
		{
			testname:      "getLBSecondTime",
			compartment:   fakeLbCompartment,
			name:          fakeLbName1,
			want:          fakeLbOcid1,
			wantErr:       nil,
			wantListCalls: 1, // totals, no new list should be performed
		},
		{
			testname:      "getLBDifferentCompartment",
			compartment:   "differentCompartment",
			name:          fakeLbName1,
			want:          fakeLbOcid1,
			wantErr:       nil,
			wantListCalls: 2,
		},
		{
			testname:      "getLBCompartmentUpdated",
			compartment:   "differentCompartment",
			name:          fakeLbName1,
			want:          fakeLbOcid1,
			wantErr:       nil,
			wantListCalls: 2,
		},
		{
			testname:      "getSecondLBFirstTime",
			compartment:   fakeLbCompartment,
			name:          fakeLbName2,
			want:          fakeLbOcid2,
			wantErr:       nil,
			wantListCalls: 3,
		},
		{
			testname:      "getSecondLBSecondTime",
			compartment:   fakeLbCompartment,
			name:          fakeLbName2,
			want:          fakeLbOcid2,
			wantErr:       nil,
			wantListCalls: 4,
		},
	}

	for _, tt := range tests {
		if tt.skip {
			continue
		}

		t.Run(tt.testname, func(t *testing.T) {
			log.Println("running test ", tt.testname)
			got, err := lbClient.GetLoadBalancerByName(context.Background(), tt.compartment, tt.name)
			if got == nil || *got.Id != tt.want {
				t.Errorf("Expected %v, but got %v", tt.want, got)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Expected error = %v, but got %v", tt.wantErr, err)
			}
			if totalListCalls != tt.wantListCalls {
				t.Errorf("Expected the total number of LB list calls %d, but got %d", tt.wantListCalls, totalListCalls)
			}
		})
	}
}

// MockLoadBalancerClient mocks LoadBalancer client implementation.
type MockLoadBalancerClient struct {
	counter   int
	debug     bool // set true to run tests with debug logs
	listCalls *int // number of list operations performed
}

type getLoadBalancerWorkRequestResponse struct {
	response loadbalancer.GetWorkRequestResponse
	err      error
}

var reqType = "LB"
var getLoadbalancerWorkrequestMap = map[string]getLoadBalancerWorkRequestResponse{
	"getWorkRequestTimedOut": {
		err: context.DeadlineExceeded,
	},
	"getWorkRequestTimedOutOnce": {
		response: loadbalancer.GetWorkRequestResponse{
			WorkRequest: loadbalancer.WorkRequest{
				LifecycleState: loadbalancer.WorkRequestLifecycleStateInProgress,
				Type:           &reqType,
			},
		},
		err: context.DeadlineExceeded,
	},
	"getWorkRequestTimedOutOnceWrappedError": {
		response: loadbalancer.GetWorkRequestResponse{
			WorkRequest: loadbalancer.WorkRequest{
				LifecycleState: loadbalancer.WorkRequestLifecycleStateInProgress,
				Type:           &reqType,
			},
		},
		err: errors2.Wrap(errors2.Wrap(errors2.WithStack(context.DeadlineExceeded), "Bar"), "Foo"),
	},
	"getWorkRequestNonRetryable": {
		err: TestNonRetryableError,
	},
	"getWorkRequestSuccess": {
		response: loadbalancer.GetWorkRequestResponse{
			WorkRequest: loadbalancer.WorkRequest{
				LifecycleState: loadbalancer.WorkRequestLifecycleStateSucceeded,
				Type:           &reqType,
			},
		},
		err: nil,
	},
}

func (c *MockLoadBalancerClient) GetWorkRequest(ctx context.Context, request loadbalancer.GetWorkRequestRequest) (response loadbalancer.GetWorkRequestResponse, err error) {
	if resp, ok := getLoadbalancerWorkrequestMap[*request.WorkRequestId]; ok {
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
				return resp.response, getLoadbalancerWorkrequestMap["getWorkRequestNonRetryable"].err
			}
		}
		return resp.response, resp.err
	}
	return loadbalancer.GetWorkRequestResponse{
		WorkRequest: loadbalancer.WorkRequest{
			LifecycleState: loadbalancer.WorkRequestLifecycleStateSucceeded,
			Type:           &reqType,
		},
	}, nil
}

func (c *MockLoadBalancerClient) GetLoadBalancer(ctx context.Context, request loadbalancer.GetLoadBalancerRequest) (response loadbalancer.GetLoadBalancerResponse, err error) {
	if *request.LoadBalancerId == fakeLbOcid2 {
		return loadbalancer.GetLoadBalancerResponse{}, errors.New("not found")
	}

	return loadbalancer.GetLoadBalancerResponse{LoadBalancer: LBMap[*request.LoadBalancerId]}, nil
}
func (c *MockLoadBalancerClient) ListWorkRequests(ctx context.Context, request loadbalancer.ListWorkRequestsRequest) (response loadbalancer.ListWorkRequestsResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) ListLoadBalancers(ctx context.Context, request loadbalancer.ListLoadBalancersRequest) (response loadbalancer.ListLoadBalancersResponse, err error) {
	for _, lb := range LBMap {
		response.Items = append(response.Items, lb)
	}
	*c.listCalls += 1

	if *c.listCalls == 2 {
		lb := LBMap[fakeLbOcid1]
		lb.CompartmentId = request.CompartmentId
		LBMap[fakeLbOcid1] = lb
	}
	return
}
func (c *MockLoadBalancerClient) CreateLoadBalancer(ctx context.Context, request loadbalancer.CreateLoadBalancerRequest) (response loadbalancer.CreateLoadBalancerResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, request loadbalancer.DeleteLoadBalancerRequest) (response loadbalancer.DeleteLoadBalancerResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) ListCertificates(ctx context.Context, request loadbalancer.ListCertificatesRequest) (response loadbalancer.ListCertificatesResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) CreateCertificate(ctx context.Context, request loadbalancer.CreateCertificateRequest) (response loadbalancer.CreateCertificateResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) CreateBackendSet(ctx context.Context, request loadbalancer.CreateBackendSetRequest) (response loadbalancer.CreateBackendSetResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) UpdateBackendSet(ctx context.Context, request loadbalancer.UpdateBackendSetRequest) (response loadbalancer.UpdateBackendSetResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) DeleteBackendSet(ctx context.Context, request loadbalancer.DeleteBackendSetRequest) (response loadbalancer.DeleteBackendSetResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) CreateListener(ctx context.Context, request loadbalancer.CreateListenerRequest) (response loadbalancer.CreateListenerResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) UpdateListener(ctx context.Context, request loadbalancer.UpdateListenerRequest) (response loadbalancer.UpdateListenerResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) DeleteListener(ctx context.Context, request loadbalancer.DeleteListenerRequest) (response loadbalancer.DeleteListenerResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) CreateRuleSet(ctx context.Context, request loadbalancer.CreateRuleSetRequest) (response loadbalancer.CreateRuleSetResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) UpdateRuleSet(ctx context.Context, request loadbalancer.UpdateRuleSetRequest) (response loadbalancer.UpdateRuleSetResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) DeleteRuleSet(ctx context.Context, request loadbalancer.DeleteRuleSetRequest) (response loadbalancer.DeleteRuleSetResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) UpdateLoadBalancerShape(ctx context.Context, request loadbalancer.UpdateLoadBalancerShapeRequest) (response loadbalancer.UpdateLoadBalancerShapeResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) UpdateNetworkSecurityGroups(ctx context.Context, request loadbalancer.UpdateNetworkSecurityGroupsRequest) (response loadbalancer.UpdateNetworkSecurityGroupsResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) UpdateLoadBalancer(ctx context.Context, request loadbalancer.UpdateLoadBalancerRequest) (response loadbalancer.UpdateLoadBalancerResponse, err error) {
	return
}

func assertError(actual, expected error) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	return actual.Error() == expected.Error()
}

func assertApproxError(actual, expected error) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	return strings.Contains(actual.Error(), expected.Error())
}
