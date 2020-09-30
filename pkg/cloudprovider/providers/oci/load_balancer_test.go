// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
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

	"github.com/oracle/oci-go-sdk/core"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_getDefaultLBSubnets(t *testing.T) {
	type args struct {
		subnet1 string
		subnet2 string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "no default subnets provided",
			args: args{},
			want: []string{""},
		},
		{
			name: "1st subnet provided",
			args: args{"subnet1", ""},
			want: []string{"subnet1"},
		},
		{
			name: "2nd subnet provided",
			args: args{"", "subnet2"},
			want: []string{"", "subnet2"},
		},
		{
			name: "both default subnets provided",
			args: args{"subnet1", "subnet2"},
			want: []string{"subnet1", "subnet2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDefaultLBSubnets(tt.args.subnet1, tt.args.subnet2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDefaultLBSubnets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLoadBalancerSubnets(t *testing.T) {
	testCases := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		nodes            []*v1.Node
		service          *v1.Service
		expected         []string
		sslConfig        *SSLConfig
	}{
		"defaults only no annotations": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
			},
			expected: []string{"one", "two"},
		},
		"internal default subnet overridden with subnet1 annotation": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "",
						ServiceAnnotationLoadBalancerSubnet1:  "regional-subnet",
					},
				},
			},
			expected: []string{"regional-subnet"},
		},
		"internal no default subnet only subnet1 annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "",
						ServiceAnnotationLoadBalancerSubnet1:  "regional-subnet",
					},
				},
			},
			expected: []string{"regional-subnet"},
		},
		"override defaults with annotations": {
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
			},
			expected: []string{"annotation-one", "annotation-two"},
		},
		"no default subnet defined override subnets via annotations": {
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
			},
			expected: []string{"annotation-one", "annotation-two"},
		},
		"no default subnet defined override subnet1 via annotations as regional subnet": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "regional-subnet",
					},
				},
			},
			expected: []string{"regional-subnet"},
		},
		"no default subnet defined override subnet2 via annotations as regional subnet": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet2: "regional-subnet",
					},
				},
			},
			expected: []string{"regional-subnet"},
		},
	}
	cp := &CloudProvider{
		client: MockOCIClient{},
		config: &providercfg.Config{CompartmentID: "testCompartment"},
	}

	for name, tc := range testCases {
		logger := zap.L()
		t.Run(name, func(t *testing.T) {

			cp.config = &providercfg.Config{
				LoadBalancer: &providercfg.LoadBalancerConfig{
					Subnet1: tc.defaultSubnetOne,
					Subnet2: tc.defaultSubnetTwo,
				},
			}
			subnets, err := cp.getLoadBalancerSubnets(context.Background(), logger.Sugar(), tc.service)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(subnets, tc.expected) {
				t.Errorf("Expected load balancer subnets\n%+v\nbut got\n%+v", tc.expected, subnets)
			}
		})
	}
}

func TestGetSubnetsForNodes(t *testing.T) {
	testCases := map[string]struct {
		nodes   []*v1.Node
		subnets []*core.Subnet
		err     error
	}{
		"Should return subnet without any error ": {
			nodes: []*v1.Node{
				{
					Spec: v1.NodeSpec{
						ProviderID: "basic-complete",
					},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							CompartmentIDAnnotation: "compID1",
						},
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{
								Type:    v1.NodeInternalIP,
								Address: "10.0.0.1",
							},
						},
					},
				},
			},
			subnets: []*core.Subnet{subnets["subnetwithdnslabel"]},
			err:     nil,
		},
		"Should return error for missing compartmentId annotation": {
			nodes: []*v1.Node{
				{
					Spec: v1.NodeSpec{
						ProviderID: "basic-complete",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "testnode",
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{
								Type:    v1.NodeInternalIP,
								Address: "10.0.0.1",
							},
						},
					},
				},
			},
			subnets: nil,
			err:     errors.New(`"oci.oraclecloud.com/compartment-id" annotation not present on node "testnode"`),
		},
		"Should return error for missing providerID": {
			nodes: []*v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "testnode",
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{
								Type:    v1.NodeInternalIP,
								Address: "10.0.0.1",
							},
						},
					},
				},
			},
			subnets: nil,
			err:     errors.New(`.spec.providerID was not present on node "testnode"`),
		},
	}
	client := MockOCIClient{}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			subnets, err := getSubnetsForNodes(context.Background(), tc.nodes, client)
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Expected node subnets error\n%+v\nbut got\n%+v", tc.err, err)
			}
			if !reflect.DeepEqual(subnets, tc.subnets) {
				t.Errorf("Expected node subnets\n%+v\nbut got\n%+v", tc.subnets, subnets)
			}
		})

	}
}
