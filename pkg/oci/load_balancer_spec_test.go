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

	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewLBSpecSuccess(t *testing.T) {
	testCases := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		nodes            []*api.Node
		service          *api.Service
		expected         LBSpec
	}{
		"defaults": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &api.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: api.ServiceSpec{
					SessionAffinity: api.ServiceAffinityNone,
					Ports: []api.ServicePort{
						{Protocol: api.ProtocolTCP},
					},
				},
			},
			expected: LBSpec{
				Name:     "test-uid",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
			},
		},
		"internal": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &api.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "",
					},
				},
				Spec: api.ServiceSpec{
					SessionAffinity: api.ServiceAffinityNone,
					Ports: []api.ServicePort{
						{Protocol: api.ProtocolTCP},
					},
				},
			},
			expected: LBSpec{
				Name:     "test-uid",
				Shape:    "100Mbps",
				Internal: true,
				Subnets:  []string{"one"},
			},
		},
		"subnet annotations": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &api.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "annotation-one",
						ServiceAnnotationLoadBalancerSubnet2: "annotation-two",
					},
				},
				Spec: api.ServiceSpec{
					SessionAffinity: api.ServiceAffinityNone,
					Ports: []api.ServicePort{
						{Protocol: api.ProtocolTCP},
					},
				},
			},
			expected: LBSpec{
				Name:     "test-uid",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"annotation-one", "annotation-two"},
			},
		},
		"custom shape": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &api.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShape: "8000Mbps",
					},
				},
				Spec: api.ServiceSpec{
					SessionAffinity: api.ServiceAffinityNone,
					Ports: []api.ServicePort{
						{Protocol: api.ProtocolTCP},
					},
				},
			},
			expected: LBSpec{
				Name:     "test-uid",
				Shape:    "8000Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
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
		nodes            []*api.Node
		service          *api.Service
		expectedErrMsg   string
	}{
		"unsupported udp protocol": {
			service: &api.Service{
				Spec: api.ServiceSpec{
					Ports: []api.ServicePort{
						{Protocol: api.ProtocolUDP},
					},
				},
			},
			expectedErrMsg: "OCI load balancers do not support UDP",
		},
		"unsupported LB IP": {
			service: &api.Service{
				Spec: api.ServiceSpec{
					LoadBalancerIP:  "127.0.0.1",
					SessionAffinity: api.ServiceAffinityNone,
					Ports: []api.ServicePort{
						{Protocol: api.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "OCI does not support setting the LoadBalancerIP",
		},
		"unsupported session affinity": {
			service: &api.Service{
				Spec: api.ServiceSpec{
					SessionAffinity: api.ServiceAffinityClientIP,
					Ports: []api.ServicePort{
						{Protocol: api.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "OCI only supports SessionAffinity `None` currently",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			subnets := []string{tc.defaultSubnetOne, tc.defaultSubnetTwo}
			_, err := NewLBSpec(tc.service, tc.nodes, subnets, nil)
			if err == nil || err.Error() != tc.expectedErrMsg {
				t.Errorf("Expected error with message %q but got `%v`", tc.expectedErrMsg, err)
			}
		})
	}
}
