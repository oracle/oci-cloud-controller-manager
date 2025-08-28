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
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/loadbalancer"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	testclient "k8s.io/client-go/kubernetes/fake"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	errors1 "github.com/pkg/errors"
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
			provisionedNodes, err := filterNodes(tc.service, tc.nodes)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(provisionedNodes, tc.expected) {
				t.Errorf("expected: %+v got %+v", tc.expected, provisionedNodes)
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
						ProviderID: "ocid1.basic-complete",
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
						ProviderID: "ocid1.basic-complete",
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
		"Ipv6 subnets return subnet without any error GUA": {
			nodes: []*v1.Node{
				{
					Spec: v1.NodeSpec{
						ProviderID: "ocid1.ipv6-instance",
					},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							CompartmentIDAnnotation: "compID1",
						},
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{
								Type:    v1.NodeExternalIP,
								Address: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
							},
						},
					},
				},
			},
			subnets: []*core.Subnet{subnets["IPv6-subnet"]},
			err:     nil,
		},
		"Ipv6 subnets return subnet without any error ULA": {
			nodes: []*v1.Node{
				{
					Spec: v1.NodeSpec{
						ProviderID: "ocid1.ipv6-instance",
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
								Address: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
							},
						},
					},
				},
			},
			subnets: []*core.Subnet{subnets["IPv6-subnet"]},
			err:     nil,
		},
		"Private IPv4 and GUA IPv6": {
			nodes: []*v1.Node{
				{
					Spec: v1.NodeSpec{
						ProviderID: "ocid1.ipv6-gua-ipv4-instance",
					},
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							CompartmentIDAnnotation: "compID1",
						},
					},
					Status: v1.NodeStatus{
						Addresses: []v1.NodeAddress{
							{
								Type:    v1.NodeExternalIP,
								Address: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
							},
							{
								Type:    v1.NodeInternalIP,
								Address: "10.0.0.1",
							},
						},
					},
				},
			},
			subnets: []*core.Subnet{subnets["ipv6-gua-ipv4-instance"]},
			err:     nil,
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

func TestCloudProvider_getLoadBalancerProvider(t *testing.T) {

	kc := NewSimpleClientset(
		&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "sa", Namespace: "ns",
			},
		})

	factory := informers.NewSharedInformerFactoryWithOptions(kc, time.Second, informers.WithNamespace("ns"))
	serviceAccountInformer := factory.Core().V1().ServiceAccounts()
	go serviceAccountInformer.Informer().Run(wait.NeverStop)

	time.Sleep(time.Second)

	cp := &CloudProvider{
		client:               MockOCIClient{},
		kubeclient:           kc,
		ServiceAccountLister: serviceAccountInformer.Lister(),
		config:               &providercfg.Config{CompartmentID: "testCompartment"},
		logger:               zap.S(),
	}

	tests := map[string]struct {
		service    *v1.Service
		wantErr    bool
		wantLbType string
		cp         *CloudProvider
	}{
		"Get Load Balancer Provider type LB with Workload Identity RP": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "testservice-lb",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationServiceAccountName: `sa`,
					},
				},
			},
			wantErr:    false,
			wantLbType: "*oci.MockLoadBalancerClient",
			cp:         cp,
		},
		"Get Load Balancer Provider type NLB with Workload Identity RP": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "testservice-nlb",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:   "nlb",
						ServiceAnnotationServiceAccountName: `sa`,
					},
				},
			},
			wantErr:    false,
			wantLbType: "*oci.MockNetworkLoadBalancerClient",
			cp:         cp,
		},
		"Fail to Get Load Balancer Provider type LB with Workload Identity RP when SA does not exist": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "testservice-get-sa-error",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationServiceAccountName: `sa-does-not-exist`,
					},
				},
			},
			wantErr: true,
			cp:      cp,
		},
		"Get Load Balancer Provider type LB": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "testservice-lb-2",
					UID:       "test-uid",
				},
			},
			wantErr:    false,
			wantLbType: "*oci.MockLoadBalancerClient",
			cp:         cp,
		},
		"Get Load Balancer Provider type NLB": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns",
					Name:      "testservice-nlb-2",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			wantErr:    false,
			wantLbType: "*oci.MockNetworkLoadBalancerClient",
			cp:         cp,
		},
	}

	t.Parallel()
	for name, tt := range tests {
		name, tt := name, tt
		t.Run(name, func(t *testing.T) {
			got, _ := tt.cp.getLoadBalancerProvider(context.Background(), tt.service)

			if !tt.wantErr {
				if reflect.DeepEqual(got, CloudLoadBalancerProvider{}) {
					t.Errorf("GetLoadBalancerProvider() didn't expect an empty provider")
				}
				if fmt.Sprintf("%T", got.lbClient) != tt.wantLbType {
					t.Errorf("GetLoadBalancerProvider() got LB of type = %T, want %T", got.lbClient, tt.wantLbType)
				}
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
				Spec: v1.ServiceSpec{
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
				},
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
				Spec: v1.ServiceSpec{
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
				},
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
				Spec: v1.ServiceSpec{
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
					Selector:   map[string]string{"hello": "world"},
				},
			},
			err:     "",
			wantErr: false,
		},
		{
			name: "Security List Management mode 'All' - fetch node failure",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
				},
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
			name: "Security List Management mode 'NSG' - no err",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "NSG",
					},
				},
			},
			err:     "",
			wantErr: false,
		},
		{
			name: "no management mode provided in annotation - no err",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
				},
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
				Spec: v1.ServiceSpec{
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
				},
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
		kubeclient: testclient.NewSimpleClientset(
			&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testservice", Namespace: "kube-system",
				},
			}),
		lbLocks: NewLoadBalancerLocks(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := cp.EnsureLoadBalancerDeleted(context.Background(), "test", tt.service); (err != nil) != tt.wantErr {
				t.Errorf("EnsureLoadBalancerDeleted() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_addLoadBalancerOkeSystemTags(t *testing.T) {
	tests := map[string]struct {
		//config  *providercfg.Config
		lb      *client.GenericLoadBalancer
		spec    *LBSpec
		wantErr error
	}{
		"expect an error when spec system tag is nil": {
			lb: &client.GenericLoadBalancer{
				Id: common.String("ocid1.loadbalancer."),
			},
			spec: &LBSpec{
				service: &v1.Service{},
			},
			wantErr: errors.New("oke system tag is not found in LB spec. ignoring.."),
		},
		"expect an error when spec system tag is empty map": {
			lb: &client.GenericLoadBalancer{
				Id: common.String("ocid1.loadbalancer."),
			},
			spec: &LBSpec{
				SystemTags: map[string]map[string]interface{}{},
				service:    &v1.Service{},
			},
			wantErr: errors.New("oke system tag namespace is not found in LB spec"),
		},
		"expect an error when defined tags are limits are reached": {
			lb: &client.GenericLoadBalancer{
				Id:          common.String("defined tag limit of 64 reached"),
				DefinedTags: make(map[string]map[string]interface{}),
			},
			spec: &LBSpec{
				Type:       LB,
				SystemTags: map[string]map[string]interface{}{"orcl-containerengine": {"Cluster": "val"}},
				service:    &v1.Service{},
			},
			wantErr: errors.New("max limit of defined tags for lb is reached. skip adding tags. sending metric"),
		},
		"expect an error when updateLoadBalancer work request fails": {
			lb: &client.GenericLoadBalancer{
				Id:           common.String("work request fail"),
				FreeformTags: map[string]string{"key": "val"},
				DefinedTags:  map[string]map[string]interface{}{"ns1": {"key1": "val1"}},
			},
			spec: &LBSpec{
				SystemTags: map[string]map[string]interface{}{"orcl-containerengine": {"Cluster": "val"}},
				service:    &v1.Service{},
			},
			wantErr: errors.New("UpdateLoadBalancer request failed: internal server error"),
		},
		"expect an error when using workload identity": {
			lb: &client.GenericLoadBalancer{
				Id:           common.String("service using workload identity"),
				FreeformTags: map[string]string{"key": "val"},
				DefinedTags:  map[string]map[string]interface{}{"ns1": {"key1": "val1"}},
			},
			spec: &LBSpec{
				SystemTags: map[string]map[string]interface{}{"orcl-containerengine": {"Cluster": "val"}},
				service: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							ServiceAnnotationServiceAccountName: "test-service-account",
						},
					},
				},
			},
			wantErr: errors.New("principal type is workload identity. skip addition of oke system tags."),
		},
	}

	for name, testcase := range tests {
		clb := &CloudLoadBalancerProvider{
			lbClient: &MockLoadBalancerClient{},
			logger:   zap.S(),
			//config:   testcase.config,
		}
		t.Run(name, func(t *testing.T) {
			if strings.Contains(name, "limit") {
				// add 64 defined tags
				for i := 1; i <= 64; i++ {
					testcase.lb.DefinedTags["ns"+strconv.Itoa(i)] = map[string]interface{}{"key": strconv.Itoa(i)}
				}
			}
			err := clb.addLoadBalancerOkeSystemTags(context.Background(), testcase.lb, testcase.spec)
			t.Logf("%s", err.Error())
			if !assertError(err, testcase.wantErr) {
				t.Errorf("Expected error = %v, but got %v", testcase.wantErr, err)
				return
			}
		})
	}
}

func Test_doesLbHaveResourceTrackingSystemTags(t *testing.T) {
	tests := map[string]struct {
		lb   *client.GenericLoadBalancer
		spec *LBSpec
		want bool
	}{
		"base case": {
			lb: &client.GenericLoadBalancer{
				DefinedTags: map[string]map[string]interface{}{"ns": {"key": "val"}},
				SystemTags:  map[string]map[string]interface{}{"orcl-containerengine": {"Cluster": "val"}},
			},
			spec: &LBSpec{
				SystemTags: map[string]map[string]interface{}{"orcl-containerengine": {"Cluster": "val"}},
			},
			want: true,
		},
		"system tag exists for different ns in lb": {
			lb: &client.GenericLoadBalancer{
				DefinedTags: map[string]map[string]interface{}{"ns": {"key": "val"}},
				SystemTags:  map[string]map[string]interface{}{"orcl-free-tier": {"Cluster": "val"}},
			},
			spec: &LBSpec{
				SystemTags: map[string]map[string]interface{}{"orcl-containerengine": {"Cluster": "val"}},
			},
			want: false,
		},
		"resource tracking system tag doesnt exists in lb": {
			lb: &client.GenericLoadBalancer{
				DefinedTags: map[string]map[string]interface{}{"ns": {"key": "val"}},
			},
			spec: &LBSpec{
				SystemTags: map[string]map[string]interface{}{"orcl-containerengine": {"Cluster": "val"}},
			},
			want: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual := doesLbHaveOkeSystemTags(test.lb, test.spec)
			t.Logf("expected %v but got %v", test.want, actual)
			if test.want != actual {
				t.Errorf("expected %v but got %v", test.want, actual)
			}
		})
	}
}

func Test_getGoadBalancerStatus(t *testing.T) {
	var proxy = v1.LoadBalancerIPModeProxy
	var vip = v1.LoadBalancerIPModeVIP
	var ipAddress = "10.0.0.0"
	var lbName = "test-lb"
	var tests = map[string]struct {
		lb             *client.GenericLoadBalancer
		setIpMode      *v1.LoadBalancerIPMode
		expectedIpMode *v1.LoadBalancerIPMode
		wantErr        error
	}{
		"ipMode is Proxy": {
			lb: &client.GenericLoadBalancer{
				IpAddresses: []client.GenericIpAddress{
					{
						IpAddress: &ipAddress,
					},
				},
			},
			setIpMode:      &proxy,
			expectedIpMode: &proxy,
			wantErr:        nil,
		},
		"ipMode is VIP": {
			lb: &client.GenericLoadBalancer{
				IpAddresses: []client.GenericIpAddress{
					{
						IpAddress: &ipAddress,
					},
				},
			},
			setIpMode:      &vip,
			expectedIpMode: &vip,
			wantErr:        nil,
		},
		"ipMode not set": {
			lb: &client.GenericLoadBalancer{
				IpAddresses: []client.GenericIpAddress{
					{
						IpAddress: &ipAddress,
					},
				},
			},
			setIpMode:      nil,
			expectedIpMode: nil,
			wantErr:        nil,
		},
		"zero ip addresses": {
			lb: &client.GenericLoadBalancer{
				DisplayName: &lbName,
				IpAddresses: []client.GenericIpAddress{},
			},
			setIpMode:      nil,
			expectedIpMode: nil,
			wantErr:        errors.New(fmt.Sprintf("no ip addresses found for load balancer %q", lbName)),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := loadBalancerToStatus(test.lb, test.setIpMode, false, zap.S())
			if !assertError(err, test.wantErr) {
				t.Errorf("Expected error = %v, but got %v", test.wantErr, err)
				return
			}
			if err == nil && !reflect.DeepEqual(actual.Ingress[0].IPMode, test.expectedIpMode) {
				t.Errorf("expected %v but got %v", test.expectedIpMode, actual.Ingress[0].IPMode)
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

func Test_checkIfSubnetIPv6Compatible(t *testing.T) {
	tests := map[string]struct {
		subnets     []*core.Subnet
		ipVersion   string
		expectedErr error
	}{
		"Subnet with IPv4 cidrs only": {
			subnets: []*core.Subnet{
				{
					Id:                 common.String("IPv4-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("10.0.0.0/16"),
				},
			},
			ipVersion:   IPv6,
			expectedErr: errors1.Errorf("subnet with id IPv4-subnet does not have an ipv6 cidr block"),
		},
		"Subnet with both IPv4 & IPv6 cidrs": {
			subnets: []*core.Subnet{
				{
					Id:                 common.String("IPv4-IPv6-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("10.0.0.0/16"),
					Ipv6CidrBlocks:     []string{},
					Ipv6CidrBlock:      common.String("IPv6Cidr"),
				},
			},
			ipVersion:   IPv6,
			expectedErr: nil,
		},
		"Subnet with IPv6 cidrs only": {
			subnets: []*core.Subnet{
				{
					Id:                 common.String("IPv6-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("<null>"),
					Ipv6CidrBlock:      common.String("IPv6Cidr"),
					Ipv6CidrBlocks:     []string{"IPv6Cidr"},
				},
			},
			ipVersion:   IPv6,
			expectedErr: nil,
		},
		"Subnet with IPv6 cidrs only IPv4 error": {
			subnets: []*core.Subnet{
				{
					Id:                 common.String("IPv6-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					CidrBlock:          common.String("<null>"),
					Ipv6CidrBlock:      common.String("IPv6Cidr"),
					Ipv6CidrBlocks:     []string{"IPv6Cidr"},
					AvailabilityDomain: nil,
				},
			},
			ipVersion:   IPv4,
			expectedErr: errors1.Errorf("subnet with id IPv6-subnet does not have an ipv4 cidr block"),
		},
		"multiple subnets with single IPv6 cidr": {
			subnets: []*core.Subnet{
				{
					Id:                 common.String("IPv6-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("<null>"),
				},
				{
					Id:                 common.String("IPv6-subnet-1"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("<null>"),
					Ipv6CidrBlock:      common.String("IPv6Cidr"),
					Ipv6CidrBlocks:     []string{"IPv6Cidr"},
				},
			},
			ipVersion:   IPv6,
			expectedErr: nil,
		},
		"multiple subnets with single IPv6 cidr check for IPv4": {
			subnets: []*core.Subnet{
				{
					Id:                 common.String("IPv6-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("<null>"),
				},
				{
					Id:                 common.String("IPv6-subnet-1"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("<null>"),
					Ipv6CidrBlock:      common.String("IPv6Cidr"),
					Ipv6CidrBlocks:     []string{"IPv6Cidr"},
				},
			},
			ipVersion:   IPv4,
			expectedErr: errors1.Errorf("subnet with id IPv6-subnet-1 does not have an ipv4 cidr block"),
		},
		"multiple subnets with single IPv6 cidr check for IPv6": {
			subnets: []*core.Subnet{
				{
					Id:                 common.String("IPv6-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("<null>"),
				},
				{
					Id:                 common.String("IPv6-subnet-1"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("<null>"),
					Ipv6CidrBlock:      common.String("IPv6Cidr"),
					Ipv6CidrBlocks:     []string{"IPv6Cidr"},
				},
			},
			ipVersion:   IPv6,
			expectedErr: nil,
		},
		"multiple subnets with single IPv4 check for IPv4": {
			subnets: []*core.Subnet{
				{
					Id:                 common.String("IPv4-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("10.0.0.0/16"),
				},
				{
					Id:                 common.String("IPv4-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
				},
				{
					Id:                 common.String("IPv4-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
				},
			},
			ipVersion:   IPv4,
			expectedErr: nil,
		},
		"multiple subnets with single IPv6 check for IPv6": {
			subnets: []*core.Subnet{
				{
					Id:                 common.String("IPv4-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("10.0.0.0/16"),
				},
				{
					Id:                 common.String("IPv6-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
					CidrBlock:          common.String("<null>"),
					Ipv6CidrBlock:      common.String("IPv6Cidr"),
					Ipv6CidrBlocks:     []string{"IPv6Cidr"}},
				{
					Id:                 common.String("IPv4-subnet"),
					DnsLabel:           common.String("subnetwithnovcndnslabel"),
					VcnId:              common.String("vcnwithoutdnslabel"),
					AvailabilityDomain: nil,
				},
			},
			ipVersion:   IPv6,
			expectedErr: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkSubnetIpFamilyCompatibility(tt.subnets, tt.ipVersion)
			if err != nil {
				if !reflect.DeepEqual(err.Error(), tt.expectedErr.Error()) {
					t.Errorf("checkIfSubnetIPv6Compatible() expected error = %v,\n but got %v", tt.expectedErr, err)
					return
				}
			}
		})
	}
}

func Test_getLbEndpointVersion(t *testing.T) {
	var tests = map[string]struct {
		ipFamilies        []string
		ipFamilyPolicy    string
		subnets           []*core.Subnet
		lbEndpointVersion string
		wantErr           error
	}{
		"SingleStack IPv4": {
			ipFamilies:     []string{IPv4},
			ipFamilyPolicy: string(v1.IPFamilyPolicySingleStack),
			subnets: []*core.Subnet{
				{CidrBlock: common.String("10.0.1.0/24")},
			},
			lbEndpointVersion: IPv4,
			wantErr:           nil,
		},
		"SingleStack IPv6": {
			ipFamilies:     []string{IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicySingleStack),
			subnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			lbEndpointVersion: IPv6,
			wantErr:           nil,
		},
		"SingleStack IPv6 - wrong subnet": {
			ipFamilies:     []string{IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicySingleStack),
			subnets: []*core.Subnet{
				{
					Id:        common.String("ocid1.subnet"),
					CidrBlock: common.String("10.0.1.0/24"),
				},
			},
			lbEndpointVersion: "",
			wantErr:           errors.New("subnet does not have IPv6 CIDR blocks: subnet with id ocid1.subnet does not have an ipv6 cidr block"),
		},
		"RequireDualStack IPv4/IPv6": {
			ipFamilies:     []string{IPv4, IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicyRequireDualStack),
			subnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			lbEndpointVersion: IPv4AndIPv6,
			wantErr:           nil,
		},
		"RequireDualStack IPv4/IPv6 - wrong subnet": {
			ipFamilies:     []string{IPv4, IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicyRequireDualStack),
			subnets: []*core.Subnet{
				{
					Id:        common.String("ocid1.subnet"),
					CidrBlock: common.String("10.0.1.0/24"),
				},
			},
			lbEndpointVersion: "",
			wantErr:           errors.New("subnet does not have IPv6 CIDR blocks: subnet with id ocid1.subnet does not have an ipv6 cidr block"),
		},
		"PreferDualStack IPv4/IPv6": {
			ipFamilies:     []string{IPv4, IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicyPreferDualStack),
			subnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			lbEndpointVersion: IPv4AndIPv6,
			wantErr:           nil,
		},
		"PreferDualStack IPv4/IPv6 - wrong subnet": {
			ipFamilies:     []string{IPv4, IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicyPreferDualStack),
			subnets: []*core.Subnet{
				{
					CidrBlock: common.String("10.0.1.0/24"),
				},
			},
			lbEndpointVersion: IPv4,
			wantErr:           nil,
		},
		"PreferDualStack IPv4": {
			ipFamilies:     []string{IPv4},
			ipFamilyPolicy: string(v1.IPFamilyPolicyPreferDualStack),
			subnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			lbEndpointVersion: IPv4AndIPv6,
			wantErr:           nil,
		},
	}
	cp := &CloudProvider{
		client:     MockOCIClient{},
		config:     &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister: &mockNodeLister{},
		kubeclient: testclient.NewSimpleClientset(),
		logger:     zap.L().Sugar(),
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			lbEndpointVersion, err := cp.getLbEndpointIpVersion(tt.ipFamilies, tt.ipFamilyPolicy, tt.subnets)
			if lbEndpointVersion != tt.lbEndpointVersion {
				t.Errorf("Expected lbEndpointVersion = %s, but got %s", tt.lbEndpointVersion, lbEndpointVersion)
			}
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("Expected error = %s,\n but got %s", tt.wantErr.Error(), err.Error())
				return
			}
		})
	}
}

func Test_getLbListenerBackendSetIpVersion(t *testing.T) {
	var tests = map[string]struct {
		ipFamilies                   []string
		ipFamilyPolicy               string
		nodeSubnets                  []*core.Subnet
		listenerBackendSetIpVersions []string
		wantErr                      error
	}{
		"SingleStack IPv4": {
			ipFamilies:     []string{IPv4},
			ipFamilyPolicy: string(v1.IPFamilyPolicySingleStack),
			nodeSubnets: []*core.Subnet{
				{CidrBlock: common.String("10.0.1.0/24")},
			},
			listenerBackendSetIpVersions: []string{IPv4},
			wantErr:                      nil,
		},
		"SingleStack IPv6 - wrong subnet": {
			ipFamilies:     []string{IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicySingleStack),
			nodeSubnets: []*core.Subnet{
				{
					Id:        common.String("ocid1.subnet"),
					CidrBlock: common.String("10.0.1.0/24"),
				},
			},
			listenerBackendSetIpVersions: []string{},
			wantErr:                      errors.New("subnet does not have IPv6 CIDR blocks: subnet with id ocid1.subnet does not have an ipv6 cidr block"),
		},
		"RequireDualStack IPv4/IPv6": {
			ipFamilies:     []string{IPv4, IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicyRequireDualStack),
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			listenerBackendSetIpVersions: []string{IPv4, IPv6},
			wantErr:                      nil,
		},
		"RequireDualStack IPv4/IPv6 - wrong subnet": {
			ipFamilies:     []string{IPv4, IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicyRequireDualStack),
			nodeSubnets: []*core.Subnet{
				{
					Id:        common.String("ocid1.subnet"),
					CidrBlock: common.String("10.0.1.0/24"),
				},
			},
			listenerBackendSetIpVersions: []string{},
			wantErr:                      errors.New("subnet does not have IPv6 CIDR blocks: subnet with id ocid1.subnet does not have an ipv6 cidr block"),
		},
		"PreferDualStack IPv4/IPv6": {
			ipFamilies:     []string{IPv4, IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicyPreferDualStack),
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			listenerBackendSetIpVersions: []string{IPv4, IPv6},
			wantErr:                      nil,
		},
		"PreferDualStack IPv4/IPv6 - wrong subnet": {
			ipFamilies:     []string{IPv4, IPv6},
			ipFamilyPolicy: string(v1.IPFamilyPolicyPreferDualStack),
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock: common.String("10.0.1.0/24"),
				},
			},
			listenerBackendSetIpVersions: []string{IPv4},
			wantErr:                      nil,
		},
		"PreferDualStack IPv4": {
			ipFamilies:     []string{IPv4},
			ipFamilyPolicy: string(v1.IPFamilyPolicyPreferDualStack),
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			listenerBackendSetIpVersions: []string{IPv4, IPv6},
			wantErr:                      nil,
		},
		"PreferDualStack IPv4 multiple subnets": {
			ipFamilies:     []string{IPv4},
			ipFamilyPolicy: string(v1.IPFamilyPolicyPreferDualStack),
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock: common.String("10.0.1.0/24"),
				},
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			listenerBackendSetIpVersions: []string{IPv4, IPv6},
			wantErr:                      nil,
		},
	}
	cp := &CloudProvider{
		client:     MockOCIClient{},
		config:     &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister: &mockNodeLister{},
		kubeclient: testclient.NewSimpleClientset(),
		logger:     zap.L().Sugar(),
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := cp.getLbListenerBackendSetIpVersion(tt.ipFamilies, tt.ipFamilyPolicy, tt.nodeSubnets)
			if !reflect.DeepEqual(result, tt.listenerBackendSetIpVersions) {
				t.Errorf("Expected listenerBackendSetIpVersions\n%+v\nbut got\n%+v", tt.listenerBackendSetIpVersions, result)
			}
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("Expected error = %s,\n but got %s", tt.wantErr.Error(), err.Error())
				return
			}
		})
	}
}

func Test_getOciIpVersions(t *testing.T) {
	var tests = map[string]struct {
		nodeSubnets []*core.Subnet
		lbSubnets   []*core.Subnet
		service     *v1.Service
		result      *IpVersions
		wantErr     error
	}{
		"SingleStack IPv4": {
			lbSubnets: []*core.Subnet{
				{
					CidrBlock: common.String("10.0.2.0/24"),
				},
			},
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicySingleStack))),
				},
			},
			nodeSubnets: []*core.Subnet{
				{CidrBlock: common.String("10.0.1.0/24")},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			wantErr: nil,
		},
		"SingleStack IPv6 for NLB GUA prefix": {
			lbSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("<null>"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicySingleStack))),
				},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("<null>"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv6},
			},
			wantErr: nil,
		},
		"SingleStack IPv6 for NLB ULA prefix": {
			lbSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("<null>"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicySingleStack))),
				},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("<null>"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv6},
			},
			wantErr: nil,
		},
		"SingleStack IPv6 LB": {
			lbSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("<null>"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "lb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicySingleStack))),
				},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("<null>"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			result:  nil,
			wantErr: errors.New("SingleStack IPv6 is not supported for OCI LBaaS"),
		},
		"RequireDualStack IPv4/IPv6 - LB": {
			lbSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.0.0/16"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String("RequireDualStack")),
				},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv4, IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyRequireDualStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4AndIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			wantErr: nil,
		},
		"RequireDualStack IPv4/IPv6 - NLB": {
			lbSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.0.0/16"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String("RequireDualStack")),
				},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv4, IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyRequireDualStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4AndIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4, client.GenericIPv6},
			},
			wantErr: nil,
		},
		"PreferDualStack IPv4 LB - IPv4 Subnet cidrs only": {
			lbSubnets: []*core.Subnet{
				{
					CidrBlock: common.String("10.0.0.0/16"),
				},
			},
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyPreferDualStack))),
				},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock: common.String("10.0.1.0/24"),
				},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyPreferDualStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			wantErr: nil,
		},
		"PreferDualStack IPv4 NLB - IPv4 Subnet cidrs only": {
			lbSubnets: []*core.Subnet{
				{
					CidrBlock: common.String("10.0.0.0/16"),
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyPreferDualStack))),
				},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock: common.String("10.0.1.0/24"),
				},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyPreferDualStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			wantErr: nil,
		},
		"PreferDualStack IPv6 LB - IPv6 Subnet cidrs only": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyPreferDualStack))),
				},
			},
			lbSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("<null>"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"}},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("<null>"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			result:  nil,
			wantErr: errors.New("SingleStack IPv6 is not supported for OCI LBaaS"),
		},
		"PreferDualStack IPv6 NLB - IPv6 Subnet cidrs only": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyPreferDualStack))),
				},
			},
			lbSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("<null>"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"}},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("<null>"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyPreferDualStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv6},
			},
			wantErr: nil,
		},
		"PreferDualStack IPv4/IPv6 LB": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyPreferDualStack))),
				},
			},
			lbSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.0.0/16"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv4, IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyPreferDualStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4AndIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			wantErr: nil,
		},
		"PreferDualStack IPv4/IPv6 NLB": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyPreferDualStack))),
				},
			},
			lbSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.0.0/16"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlock:  common.String("2001:0000:130F:0000:0000:09C0:876A:130B"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv4, IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyPreferDualStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4AndIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4, client.GenericIPv6},
			},
			wantErr: nil,
		},
		"PreferDualStack IPv4/IPv6 multiple subnets": {
			lbSubnets: []*core.Subnet{
				{
					CidrBlock: common.String("10.0.0.0/16"),
				},
				{
					CidrBlock:      common.String("10.0.0.0/16"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyPreferDualStack))),
				},
			},
			nodeSubnets: []*core.Subnet{
				{
					CidrBlock: common.String("10.0.1.0/24"),
				},
				{
					CidrBlock:      common.String("10.0.1.0/24"),
					Ipv6CidrBlocks: []string{"2001:0000:130F:0000:0000:09C0:876A:130B"},
				},
			},
			result: &IpVersions{
				IpFamilies:               []string{IPv4, IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyPreferDualStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4AndIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4, client.GenericIPv6},
			},
			wantErr: nil,
		},
	}
	cp := &CloudProvider{
		client:     MockOCIClient{},
		config:     &providercfg.Config{CompartmentID: "testCompartment"},
		NodeLister: &mockNodeLister{},
		kubeclient: testclient.NewSimpleClientset(),
		logger:     zap.L().Sugar(),
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := cp.getOciIpVersions(tt.lbSubnets, tt.nodeSubnets, tt.service)
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Expected IpVersions\n%+v\nbut got\n%+v", tt.result, result)
			}
			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("Expected error = %s,\n but got %s", tt.wantErr.Error(), err.Error())
				return
			}
		})
	}
}

func TestLoadBalancerToStatus(t *testing.T) {
	type testCase struct {
		name           string
		lb             *client.GenericLoadBalancer
		ipMode         v1.LoadBalancerIPMode
		skipPrivateIp  bool
		expectedOutput *v1.LoadBalancerStatus
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "No IP Addresses",
			lb: &client.GenericLoadBalancer{
				DisplayName: common.String("test-lb"),
				IpAddresses: []client.GenericIpAddress{},
			},
			ipMode:         v1.LoadBalancerIPModeVIP,
			skipPrivateIp:  false,
			expectedOutput: nil,
			expectedError:  errors1.Errorf("no ip addresses found for load balancer \"test-lb\""),
		},
		{
			name: "No IP Addresses - LB in Failed state",
			lb: &client.GenericLoadBalancer{
				DisplayName:    common.String("test-lb"),
				Id:             common.String("test-id"),
				IpAddresses:    []client.GenericIpAddress{},
				LifecycleState: common.String(string(loadbalancer.LoadBalancerLifecycleStateFailed)),
			},
			ipMode:         v1.LoadBalancerIPModeVIP,
			skipPrivateIp:  false,
			expectedOutput: &v1.LoadBalancerStatus{},
			expectedError:  nil,
		},
		{
			name: "Single Public IP Address",
			lb: &client.GenericLoadBalancer{
				DisplayName: common.String("test-lb"),
				IpAddresses: []client.GenericIpAddress{
					{
						IpAddress: common.String("192.168.1.100"),
						IsPublic:  common.Bool(true),
					},
				},
			},
			ipMode:        v1.LoadBalancerIPModeVIP,
			skipPrivateIp: false,
			expectedOutput: &v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{
					{
						IP: "192.168.1.100",
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Multiple IP Addresses",
			lb: &client.GenericLoadBalancer{
				DisplayName: common.String("test-lb"),
				IpAddresses: []client.GenericIpAddress{
					{
						IpAddress: common.String("192.168.1.100"),
						IsPublic:  common.Bool(true),
					},
					{
						IpAddress: common.String("10.0.0.100"),
						IsPublic:  common.Bool(false),
					},
				},
			},
			ipMode:        v1.LoadBalancerIPModeVIP,
			skipPrivateIp: false,
			expectedOutput: &v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{
					{
						IP: "192.168.1.100",
					},
					{
						IP: "10.0.0.100",
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Skip Private IP",
			lb: &client.GenericLoadBalancer{
				DisplayName: common.String("test-lb"),
				IpAddresses: []client.GenericIpAddress{
					{
						IpAddress: common.String("192.168.1.100"),
						IsPublic:  common.Bool(true),
					},
					{
						IpAddress: common.String("10.0.0.100"),
						IsPublic:  common.Bool(false),
					},
				},
			},
			ipMode:        v1.LoadBalancerIPModeVIP,
			skipPrivateIp: true,
			expectedOutput: &v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{
					{
						IP: "192.168.1.100",
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Nil IpAddress",
			lb: &client.GenericLoadBalancer{
				DisplayName: common.String("test-lb"),
				IpAddresses: []client.GenericIpAddress{
					{
						IpAddress: nil,
					},
				},
			},
			ipMode:        v1.LoadBalancerIPModeVIP,
			skipPrivateIp: false,
			expectedOutput: &v1.LoadBalancerStatus{
				Ingress: []v1.LoadBalancerIngress{},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOutput, actualError := loadBalancerToStatus(tc.lb, &tc.ipMode, tc.skipPrivateIp, zap.S())
			reflect.DeepEqual(tc.expectedOutput, actualOutput)
			if tc.expectedError != nil {
				reflect.DeepEqual(tc.expectedError.Error(), actualError.Error())
			}
		})
	}
}
