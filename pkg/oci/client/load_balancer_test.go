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

// MockLoadBalancerClient mocks LoadBalancer client implementation.
type MockLoadBalancerClient struct {
	counter int
	debug   bool // set true to run tests with debug logs
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
	return
}
func (c *MockLoadBalancerClient) ListWorkRequests(ctx context.Context, request loadbalancer.ListWorkRequestsRequest) (response loadbalancer.ListWorkRequestsResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) ListLoadBalancers(ctx context.Context, request loadbalancer.ListLoadBalancersRequest) (response loadbalancer.ListLoadBalancersResponse, err error) {
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
func (c *MockLoadBalancerClient) UpdateLoadBalancerShape(ctx context.Context, request loadbalancer.UpdateLoadBalancerShapeRequest) (response loadbalancer.UpdateLoadBalancerShapeResponse, err error) {
	return
}
func (c *MockLoadBalancerClient) UpdateNetworkSecurityGroups(ctx context.Context, request loadbalancer.UpdateNetworkSecurityGroupsRequest) (response loadbalancer.UpdateNetworkSecurityGroupsResponse, err error) {
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
