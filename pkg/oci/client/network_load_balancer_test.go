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

type MockNetworkLoadBalancerClient struct {
	// MockLoadBalancerClient mocks LoadBalancer client implementation.
	counter int
	debug   bool // set true to run tests with debug logs
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
	return
}
func (c *MockNetworkLoadBalancerClient) ListWorkRequests(ctx context.Context, request networkloadbalancer.ListWorkRequestsRequest) (response networkloadbalancer.ListWorkRequestsResponse, err error) {
	return
}
func (c *MockNetworkLoadBalancerClient) ListNetworkLoadBalancers(ctx context.Context, request networkloadbalancer.ListNetworkLoadBalancersRequest) (response networkloadbalancer.ListNetworkLoadBalancersResponse, err error) {
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
