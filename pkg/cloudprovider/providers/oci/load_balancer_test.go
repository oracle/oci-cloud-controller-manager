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

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/core"
)

func newNodeObj(name string, labels map[string]string) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}

func Test_filterNodes(t *testing.T) {
	testCases := map[string]struct {
		nodes    []*v1.Node
		service  *v1.Service
		expected []*v1.Node
	}{
		"lb - no annotation": {
			nodes: []*v1.Node{
				newNodeObj("node1", nil),
				newNodeObj("node2", nil),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					Annotations: map[string]string{},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", nil),
				newNodeObj("node2", nil),
			},
		},
		"nlb - no annotation": {
			nodes: []*v1.Node{
				newNodeObj("node1", nil),
				newNodeObj("node2", nil),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", nil),
				newNodeObj("node2", nil),
			},
		},
		"lb - empty annotation": {
			nodes: []*v1.Node{
				newNodeObj("node1", nil),
				newNodeObj("node2", nil),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNodeFilter: "",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", nil),
				newNodeObj("node2", nil),
			},
		},
		"nlb - empty annotation": {
			nodes: []*v1.Node{
				newNodeObj("node1", nil),
				newNodeObj("node2", nil),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:              "nlb",
						ServiceAnnotationNetworkLoadBalancerNodeFilter: "",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", nil),
				newNodeObj("node2", nil),
			},
		},
		"lb - single selector select all": {
			nodes: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "bar"}),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNodeFilter: "foo=bar",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "bar"}),
			},
		},
		"nlb - single selector select all": {
			nodes: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "bar"}),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:              "nlb",
						ServiceAnnotationNetworkLoadBalancerNodeFilter: "foo=bar",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "bar"}),
			},
		},
		"lb - single selector select some": {
			nodes: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "baz"}),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNodeFilter: "foo=bar",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
			},
		},
		"nlb - single selector select some": {
			nodes: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "baz"}),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:              "nlb",
						ServiceAnnotationNetworkLoadBalancerNodeFilter: "foo=bar",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
			},
		},
		"lb - multi selector select all": {
			nodes: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "baz"}),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNodeFilter: "foo",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "baz"}),
			},
		},
		"nlb - multi selector select all": {
			nodes: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "baz"}),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:              "nlb",
						ServiceAnnotationNetworkLoadBalancerNodeFilter: "foo",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "baz"}),
			},
		},
		"lb - multi selector select some": {
			nodes: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "joe"}),
				newNodeObj("node2", map[string]string{"foo": "baz"}),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNodeFilter: "foo in (bar, baz)",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "baz"}),
			},
		},
		"nlb - multi selector select some": {
			nodes: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "joe"}),
				newNodeObj("node2", map[string]string{"foo": "baz"}),
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:              "nlb",
						ServiceAnnotationNetworkLoadBalancerNodeFilter: "foo in (bar, baz)",
					},
				},
			},
			expected: []*v1.Node{
				newNodeObj("node1", map[string]string{"foo": "bar"}),
				newNodeObj("node2", map[string]string{"foo": "baz"}),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			nodes, err := filterNodes(tc.service, tc.nodes)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(nodes, tc.expected) {
				t.Errorf("expected: %+v got %+v", tc.expected, nodes)
			}
		})
	}
}

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
						ServiceAnnotationLoadBalancerInternal: "true",
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
						ServiceAnnotationLoadBalancerInternal: "true",
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
		"defaults only no annotations nlb": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expected: []string{"one"},
		},
		"internal default subnet overridden with subnet annotation NLB": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:            "nlb",
						ServiceAnnotationNetworkLoadBalancerInternal: "true",
						ServiceAnnotationNetworkLoadBalancerSubnet:   "subnet",
					},
				},
			},
			expected: []string{"subnet"},
		},
		"internal no default subnet only subnet annotation nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:            "nlb",
						ServiceAnnotationNetworkLoadBalancerInternal: "true",
						ServiceAnnotationNetworkLoadBalancerSubnet:   "subnet",
					},
				},
			},
			expected: []string{"subnet"},
		},
		"override defaults with annotations nlb": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:          "nlb",
						ServiceAnnotationNetworkLoadBalancerSubnet: "annotation-one",
					},
				},
			},
			expected: []string{"annotation-one"},
		},
		"no default subnet defined override subnets via annotationsnlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:          "nlb",
						ServiceAnnotationNetworkLoadBalancerSubnet: "annotation-one",
					},
				},
			},
			expected: []string{"annotation-one"},
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

func Test_getSubnets(t *testing.T) {
	tests := map[string]struct {
		subnetIds []string
		want      []*core.Subnet
		wantErr   bool
	}{
		"Get Subnets": {
			subnetIds: []string{"regional-subnet"},
			want: []*core.Subnet{
				{
					Id:                 common.String("regional-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
				},
			},
			wantErr: false,
		},
		"Get Subnets Error": {
			subnetIds: []string{"regional-subnet-not-found"},
			want:      nil,
			wantErr:   true,
		},
	}

	n := &MockVirtualNetworkClient{}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := getSubnets(context.Background(), tt.subnetIds, n)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSubnets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSubnets() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCloudProvider_GetLoadBalancer(t *testing.T) {

	tests := map[string]struct {
		service *v1.Service
		want    *v1.LoadBalancerStatus
		exists  bool
		wantErr bool
	}{
		"Get Load Balancer from LB client": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					UID: "privateLB",
				},
			},
			want: &v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{
					{
						IP: "10.0.50.5",
					},
				},
			},
			exists:  true,
			wantErr: false,
		},
		"Load Balancer IP address does not exist": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					UID: "privateLB-no-IP",
				},
			},
			want:    nil,
			exists:  false,
			wantErr: true,
		},
	}
	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clusterName := "test"
			got, got1, err := cp.GetLoadBalancer(context.Background(), clusterName, tt.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLoadBalancer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLoadBalancer() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.exists {
				t.Errorf("GetLoadBalancer() got1 = %v, want %v", got1, tt.exists)
			}
		})
	}
}

func TestUpdateLoadBalancerNetworkSecurityGroups(t *testing.T) {
	var tests = map[string]struct {
		spec         *LBSpec
		loadbalancer *client.GenericLoadBalancer
		wantErr      error
	}{
		"lb id is missing": {
			spec: &LBSpec{
				Name:                    "test",
				NetworkSecurityGroupIds: []string{"ocid1"},
			},
			loadbalancer: &client.GenericLoadBalancer{
				Id:          common.String(""),
				DisplayName: common.String("privateLB"),
			},
			wantErr: errors.New("failed to create UpdateNetworkSecurityGroups request: provided LB ID is empty"),
		},
		"failed to create workrequest": {
			spec: &LBSpec{
				Name:                    "test",
				NetworkSecurityGroupIds: []string{"ocid1"},
			},
			loadbalancer: &client.GenericLoadBalancer{
				Id:          common.String("failedToCreateRequest"),
				DisplayName: common.String("privateLB"),
			},
			wantErr: errors.New("failed to create UpdateNetworkSecurityGroups request: internal server error"),
		},
		"failed to get workrequest": {
			spec: &LBSpec{
				Name:                    "test",
				NetworkSecurityGroupIds: []string{"ocid1"},
			},
			loadbalancer: &client.GenericLoadBalancer{
				Id:          common.String("failedToGetUpdateNetworkSecurityGroupsWorkRequest"),
				DisplayName: common.String("privateLB"),
			},
			wantErr: errors.New("failed to await UpdateNetworkSecurityGroups workrequest: internal server error for get workrequest call"),
		},
		"Update NSG to existing LB": {
			spec: &LBSpec{
				Name:                    "test",
				NetworkSecurityGroupIds: []string{"ocid1"},
			},
			loadbalancer: &client.GenericLoadBalancer{
				Id:          common.String("ocid1"),
				DisplayName: common.String("privateLB"),
			},
			wantErr: nil,
		},
	}
	cp := &CloudLoadBalancerProvider{
		lbClient: &MockLoadBalancerClient{},
		client:   MockOCIClient{},
		logger:   zap.S(),
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := cp.updateLoadBalancerNetworkSecurityGroups(context.Background(), tt.loadbalancer, tt.spec)
			if !assertError(err, tt.wantErr) {
				t.Errorf("Expected error = %v, but got %v", tt.wantErr, err)
				return
			}
		})
	}
}

func TestCloudProvider_EnsureLoadBalancerDeleted(t *testing.T) {
	tests := []struct {
		name    string
		service *v1.Service
		err     string
		wantErr bool
	}{
		{
			name: "Security List Management mode 'None' - no err",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityListManagementMode: "None",
					},
				},
			},
			err:     "",
			wantErr: false,
		},
		{
			name: "Security List Management mode 'None' - delete err",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid-delete-err",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityListManagementMode: "None",
					},
				},
			},
			err:     "delete load balancer \"test-uid-delete-err\"",
			wantErr: true,
		},
		{
			name: "Security List Management mode 'All' - no err",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityListManagementMode: "All",
					},
				},
			},
			err:     "",
			wantErr: false,
		},
		{
			name: "Security List Management mode 'All' - fetch node failure",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid-node-err",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityListManagementMode: "All",
					},
				},
			},
			err:     "fetching nodes by internal ips",
			wantErr: true,
		},
		{
			name: "no management mode provided in annotation - no err",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
			},
			err:     "",
			wantErr: false,
		},
		{
			name: "no management mode provided in annotation - delete err",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid-delete-err",
				},
			},
			err:     "delete load balancer \"test-uid-delete-err\"",
			wantErr: true,
		},
	}
	cp := &CloudProvider{
		NodeLister: &mockNodeLister{},
		client:     MockOCIClient{},
		securityListManagerFactory: func(mode string) securityListManager {
			return MockSecurityListManager{}
		},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
		metricPusher:  nil,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cp.EnsureLoadBalancerDeleted(context.Background(), "test", tt.service); (err != nil) != tt.wantErr {
				t.Errorf("EnsureLoadBalancerDeleted() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func assertError(actual, expected error) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	return actual.Error() == expected.Error()
}
