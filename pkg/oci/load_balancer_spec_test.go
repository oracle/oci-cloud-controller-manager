// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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
	"reflect"
	"testing"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/loadbalancer"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewLBSpecSuccess(t *testing.T) {
	testCases := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		nodes            []*v1.Node
		service          *v1.Service
		expected         *LBSpec
	}{
		"defaults": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]loadbalancer.ListenerDetails{
					"TCP-80": loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Port:     common.Int(80),
						Protocol: common.String("TCP"),
					},
				},
				BackendSets: map[string]loadbalancer.BackendSetDetails{
					"TCP-80": loadbalancer.BackendSetDetails{
						Backends: []loadbalancer.BackendDetails{},
						HealthChecker: &loadbalancer.HealthCheckerDetails{
							Protocol: common.String("HTTP"),
							Port:     common.Int(10256),
							UrlPath:  common.String("/healthz"),
						},
						Policy: common.String("ROUND_ROBIN"),
					},
				},
				SourceCIDRs: []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
			},
		},
		"internal": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "",
					},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Shape:    "100Mbps",
				Internal: true,
				Subnets:  []string{"one"},
				Listeners: map[string]loadbalancer.ListenerDetails{
					"TCP-80": loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Port:     common.Int(80),
						Protocol: common.String("TCP"),
					},
				},
				BackendSets: map[string]loadbalancer.BackendSetDetails{
					"TCP-80": loadbalancer.BackendSetDetails{
						Backends: []loadbalancer.BackendDetails{},
						HealthChecker: &loadbalancer.HealthCheckerDetails{
							Protocol: common.String("HTTP"),
							Port:     common.Int(10256),
							UrlPath:  common.String("/healthz"),
						},
						Policy: common.String("ROUND_ROBIN"),
					},
				},
				SourceCIDRs: []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
			},
		},
		"subnet annotations": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "annotation-one",
						ServiceAnnotationLoadBalancerSubnet2: "annotation-two",
					},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"annotation-one", "annotation-two"},
				Listeners: map[string]loadbalancer.ListenerDetails{
					"TCP-80": loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Port:     common.Int(80),
						Protocol: common.String("TCP"),
					},
				},
				BackendSets: map[string]loadbalancer.BackendSetDetails{
					"TCP-80": loadbalancer.BackendSetDetails{
						Backends: []loadbalancer.BackendDetails{},
						HealthChecker: &loadbalancer.HealthCheckerDetails{
							Protocol: common.String("HTTP"),
							Port:     common.Int(10256),
							UrlPath:  common.String("/healthz"),
						},
						Policy: common.String("ROUND_ROBIN"),
					},
				},
				SourceCIDRs: []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
			},
		},
		"custom shape": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShape: "8000Mbps",
					},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Shape:    "8000Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]loadbalancer.ListenerDetails{
					"TCP-80": loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Port:     common.Int(80),
						Protocol: common.String("TCP"),
					},
				},
				BackendSets: map[string]loadbalancer.BackendSetDetails{
					"TCP-80": loadbalancer.BackendSetDetails{
						Backends: []loadbalancer.BackendDetails{},
						HealthChecker: &loadbalancer.HealthCheckerDetails{
							Protocol: common.String("HTTP"),
							Port:     common.Int(10256),
							UrlPath:  common.String("/healthz"),
						},
						Policy: common.String("ROUND_ROBIN"),
					},
				},
				SourceCIDRs: []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
			},
		},
		"custom idle connection timeout": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerConnectionIdleTimeout: "404",
					},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]loadbalancer.ListenerDetails{
					"TCP-80": loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Port:     common.Int(80),
						Protocol: common.String("TCP"),
						ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
							IdleTimeout: common.Int64(404),
						},
					},
				},
				BackendSets: map[string]loadbalancer.BackendSetDetails{
					"TCP-80": loadbalancer.BackendSetDetails{
						Backends: []loadbalancer.BackendDetails{},
						HealthChecker: &loadbalancer.HealthCheckerDetails{
							Protocol: common.String("HTTP"),
							Port:     common.Int(10256),
							UrlPath:  common.String("/healthz"),
						},
						Policy: common.String("ROUND_ROBIN"),
					},
				},
				SourceCIDRs: []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// we expect the service to be unchanged
			tc.expected.service = tc.service
			subnets := []string{tc.defaultSubnetOne, tc.defaultSubnetTwo}
			result, err := NewLBSpec(tc.service, tc.nodes, subnets, nil)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected load balancer spec\n%+v\nbut got\n%+v", tc.expected, result)
			}
		})
	}
}

func TestNewLBSpecFailure(t *testing.T) {
	testCases := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		nodes            []*v1.Node
		service          *v1.Service
		expectedErrMsg   string
	}{
		"unsupported udp protocol": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolUDP},
					},
				},
			},
			expectedErrMsg: "invalid service: OCI load balancers do not support UDP",
		},
		"unsupported LB IP": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					LoadBalancerIP:  "127.0.0.1",
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "invalid service: OCI does not support setting LoadBalancerIP",
		},
		"unsupported session affinity": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityClientIP,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "invalid service: OCI only supports SessionAffinity \"None\" currently",
		},
		"invalid idle connection timeout": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerConnectionIdleTimeout: "whoops",
					},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "error parsing service annotation: service.beta.kubernetes.io/oci-load-balancer-connection-idle-timeout=whoops",
		},
		"missing subnet defaults and annotations": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports:           []v1.ServicePort{},
				},
			},
			expectedErrMsg: "a configuration for both subnets must be specified",
		},
		"internal lb missing subnet1": {
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "",
					},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports:           []v1.ServicePort{},
				},
			},
			expectedErrMsg: "a configuration for subnet1 must be specified for an internal load balancer",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			subnets := []string{tc.defaultSubnetOne, tc.defaultSubnetTwo}
			_, err := NewLBSpec(tc.service, tc.nodes, subnets, nil)
			if err == nil || err.Error() != tc.expectedErrMsg {
				t.Errorf("Expected error with message %q but got %q", tc.expectedErrMsg, err)
			}
		})
	}
}
