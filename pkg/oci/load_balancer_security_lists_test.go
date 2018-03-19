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
	"github.com/oracle/oci-go-sdk/core"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

func TestGetNodeIngressRules(t *testing.T) {
	testCases := []struct {
		name         string
		securityList *core.SecurityList
		lbSubnets    []*core.Subnet
		port         int
		services     []*v1.Service
		expected     []core.IngressSecurityRule
	}{
		{
			name: "new ingress",
			securityList: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
				},
			},
			lbSubnets: []*core.Subnet{
				{CidrBlock: common.String("1")},
				{CidrBlock: common.String("2")},
			},
			port:     80,
			services: []*v1.Service{},
			expected: []core.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("2", 80),
			},
		}, {
			name: "no change",
			securityList: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
				},
			},
			lbSubnets: []*core.Subnet{
				{CidrBlock: common.String("1")},
				{CidrBlock: common.String("2")},
			},
			port:     80,
			services: []*v1.Service{},
			expected: []core.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("2", 80),
			},
		}, {
			name: "change lb subnet",
			securityList: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
					makeIngressSecurityRule("existing", 9001),
				},
			},
			lbSubnets: []*core.Subnet{
				{CidrBlock: common.String("1")},
				{CidrBlock: common.String("3")},
			},
			port:     80,
			services: []*v1.Service{},
			expected: []core.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("existing", 9001),
				makeIngressSecurityRule("3", 80),
			},
		}, {
			name: "remove lb subnets",
			securityList: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
					makeIngressSecurityRule("existing", 9001),
				},
			},
			lbSubnets: []*core.Subnet{},
			port:      80,
			services:  []*v1.Service{},
			expected: []core.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("existing", 9001),
			},
		}, {
			name: "do not delete a port rule which is used by another services (default) health check",
			securityList: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("0.0.0.0/0", lbNodesHealthCheckPort),
				},
			},
			lbSubnets: []*core.Subnet{},
			port:      lbNodesHealthCheckPort,
			services: []*v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "namespace", Name: "using-default-health-check-port"},
					Spec: v1.ServiceSpec{
						Type:  v1.ServiceTypeLoadBalancer,
						Ports: []v1.ServicePort{{Port: 80}},
					},
				},
			},
			expected: []core.IngressSecurityRule{
				makeIngressSecurityRule("0.0.0.0/0", lbNodesHealthCheckPort),
			},
		},
	}

	for _, tc := range testCases {
		serviceCache := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
		serviceLister := v1listers.NewServiceLister(serviceCache)
		for i := range tc.services {
			if err := serviceCache.Add(tc.services[i]); err != nil {
				t.Fatalf("%s unexpected service add error: %v", tc.name, err)
			}
		}
		t.Run(tc.name, func(t *testing.T) {
			rules := getNodeIngressRules(tc.securityList.IngressSecurityRules, tc.lbSubnets, tc.port, serviceLister)
			if !reflect.DeepEqual(rules, tc.expected) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expected, rules)
			}
		})
	}
}

func TestGetLoadBalancerIngressRules(t *testing.T) {
	testCases := []struct {
		name         string
		securityList *core.SecurityList
		sourceCIDRs  []string
		port         int
		services     []*v1.Service
		expected     []core.IngressSecurityRule
	}{
		{
			name: "new source cidrs",
			securityList: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
				},
			},
			sourceCIDRs: []string{
				"1",
				"2",
			},
			port:     80,
			services: []*v1.Service{},
			expected: []core.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("2", 80),
			},
		}, {
			name: "no change",
			securityList: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
				},
			},
			sourceCIDRs: []string{
				"1",
				"2",
			},
			port:     80,
			services: []*v1.Service{},
			expected: []core.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("2", 80),
			},
		}, {
			name: "change source cidr",
			securityList: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
					makeIngressSecurityRule("existing", 9001),
				},
			},
			sourceCIDRs: []string{
				"1",
				"3",
			},
			port:     80,
			services: []*v1.Service{},
			expected: []core.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("existing", 9001),
				makeIngressSecurityRule("3", 80),
			},
		}, {
			name: "remove source cidrs",
			securityList: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
					makeIngressSecurityRule("existing", 9001),
				},
			},
			sourceCIDRs: []string{},
			port:        80,
			services:    []*v1.Service{},
			expected: []core.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("existing", 9001),
			},
		}, {
			name: "do not delete a port rule which is in use by another service",
			securityList: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("0.0.0.0/0", 80),
				},
			},
			sourceCIDRs: []string{},
			port:        80,
			services: []*v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "namespace", Name: "using-port-80"},
					Spec: v1.ServiceSpec{
						Type:  v1.ServiceTypeLoadBalancer,
						Ports: []v1.ServicePort{{Port: 80}},
					},
				},
			},
			expected: []core.IngressSecurityRule{
				makeIngressSecurityRule("0.0.0.0/0", 80),
			},
		},
	}

	for _, tc := range testCases {
		serviceCache := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
		serviceLister := v1listers.NewServiceLister(serviceCache)
		for i := range tc.services {
			if err := serviceCache.Add(tc.services[i]); err != nil {
				t.Fatalf("%s unexpected service add error: %v", tc.name, err)
			}
		}
		t.Run(tc.name, func(t *testing.T) {
			rules := getLoadBalancerIngressRules(tc.securityList.IngressSecurityRules, tc.sourceCIDRs, tc.port, serviceLister)
			if !reflect.DeepEqual(rules, tc.expected) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expected, rules)
			}
		})
	}
}

func TestGetLoadBalancerEgressRules(t *testing.T) {
	testCases := []struct {
		name         string
		securityList *core.SecurityList
		subnets      []*core.Subnet
		port         int
		services     []*v1.Service
		expected     []core.EgressSecurityRule
	}{
		{
			name: "new egress",
			securityList: &core.SecurityList{
				EgressSecurityRules: []core.EgressSecurityRule{
					makeEgressSecurityRule("existing", 9000),
				},
			},
			subnets: []*core.Subnet{
				{CidrBlock: common.String("1")},
				{CidrBlock: common.String("2")},
			},
			port:     80,
			services: []*v1.Service{},
			expected: []core.EgressSecurityRule{
				makeEgressSecurityRule("existing", 9000),
				makeEgressSecurityRule("1", 80),
				makeEgressSecurityRule("2", 80),
			},
		}, {
			name: "no change",
			securityList: &core.SecurityList{
				EgressSecurityRules: []core.EgressSecurityRule{
					makeEgressSecurityRule("existing", 9000),
					makeEgressSecurityRule("1", 80),
					makeEgressSecurityRule("2", 80),
				},
			},
			subnets: []*core.Subnet{
				{CidrBlock: common.String("1")},
				{CidrBlock: common.String("2")},
			},
			port:     80,
			services: []*v1.Service{},
			expected: []core.EgressSecurityRule{
				makeEgressSecurityRule("existing", 9000),
				makeEgressSecurityRule("1", 80),
				makeEgressSecurityRule("2", 80),
			},
		}, {
			name: "change node subnet",
			securityList: &core.SecurityList{
				EgressSecurityRules: []core.EgressSecurityRule{
					makeEgressSecurityRule("existing", 9000),
					makeEgressSecurityRule("1", 80),
					makeEgressSecurityRule("2", 80),
					makeEgressSecurityRule("existing", 9001),
				},
			},
			subnets: []*core.Subnet{
				{CidrBlock: common.String("1")},
				{CidrBlock: common.String("3")},
			},
			port:     80,
			services: []*v1.Service{},
			expected: []core.EgressSecurityRule{
				makeEgressSecurityRule("existing", 9000),
				makeEgressSecurityRule("1", 80),
				makeEgressSecurityRule("existing", 9001),
				makeEgressSecurityRule("3", 80),
			},
		}, {
			name: "remove node subnets",
			securityList: &core.SecurityList{
				EgressSecurityRules: []core.EgressSecurityRule{
					makeEgressSecurityRule("existing", 9000),
					makeEgressSecurityRule("1", 80),
					makeEgressSecurityRule("2", 80),
					makeEgressSecurityRule("existing", 9001),
				},
			},
			subnets:  []*core.Subnet{},
			port:     80,
			services: []*v1.Service{},
			expected: []core.EgressSecurityRule{
				makeEgressSecurityRule("existing", 9000),
				makeEgressSecurityRule("existing", 9001),
			},
		}, {
			name: "do not delete a port rule which is used by another services (default) health check",
			securityList: &core.SecurityList{
				EgressSecurityRules: []core.EgressSecurityRule{
					makeEgressSecurityRule("0.0.0.0/0", lbNodesHealthCheckPort),
				},
			},
			subnets: []*core.Subnet{},
			port:    lbNodesHealthCheckPort,
			services: []*v1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "namespace", Name: "using-default-health-check-port"},
					Spec: v1.ServiceSpec{
						Type:  v1.ServiceTypeLoadBalancer,
						Ports: []v1.ServicePort{{Port: 80}},
					},
				},
			},
			expected: []core.EgressSecurityRule{
				makeEgressSecurityRule("0.0.0.0/0", lbNodesHealthCheckPort),
			},
		},
	}

	for _, tc := range testCases {
		serviceCache := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
		serviceLister := v1listers.NewServiceLister(serviceCache)
		for i := range tc.services {
			if err := serviceCache.Add(tc.services[i]); err != nil {
				t.Fatalf("%s unexpected service add error: %v", tc.name, err)
			}
		}
		t.Run(tc.name, func(t *testing.T) {
			rules := getLoadBalancerEgressRules(tc.securityList.EgressSecurityRules, tc.subnets, tc.port, serviceLister)
			if !reflect.DeepEqual(rules, tc.expected) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expected, rules)
			}
		})
	}
}

func TestMakeIngressSecurityRuleHasProtocolOptions(t *testing.T) {
	cdirRange := "10.0.0.0/16"
	port := 80
	rule := makeIngressSecurityRule(cdirRange, port)
	if rule.TcpOptions == nil && rule.UdpOptions == nil {
		t.Errorf("makeIngressSecurityRule(%q, %d) did not set protocol options",
			cdirRange, port)
	}
}

func TestMakeEgressSecurityRuleHasProtocolOptions(t *testing.T) {
	cdirRange := "10.0.0.0/16"
	port := 80
	rule := makeEgressSecurityRule(cdirRange, port)
	if rule.TcpOptions == nil && rule.UdpOptions == nil {
		t.Errorf("makeEgressSecurityRule(%q, %d) did not set protocol options",
			cdirRange, port)
	}
}

func TestSecurityListRulesChanged(t *testing.T) {
	testCases := map[string]struct {
		list     *core.SecurityList
		ingress  []core.IngressSecurityRule
		egress   []core.EgressSecurityRule
		expected bool
	}{
		"no change": {
			list: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("1", 80),
				},
				EgressSecurityRules: []core.EgressSecurityRule{
					makeEgressSecurityRule("1", 80),
				},
			},
			ingress: []core.IngressSecurityRule{
				makeIngressSecurityRule("1", 80),
			},
			egress: []core.EgressSecurityRule{
				makeEgressSecurityRule("1", 80),
			},
			expected: false,
		},
		"change ingress - add": {
			list: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("1", 80),
				},
			},
			ingress: []core.IngressSecurityRule{
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("2", 81),
			},
			expected: true,
		},
		"change ingress - remove": {
			list: &core.SecurityList{
				IngressSecurityRules: []core.IngressSecurityRule{
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 81),
				},
			},
			ingress: []core.IngressSecurityRule{
				makeIngressSecurityRule("1", 80),
			},
			expected: true,
		},
		"change egress - add": {
			list: &core.SecurityList{
				EgressSecurityRules: []core.EgressSecurityRule{
					makeEgressSecurityRule("1", 80),
				},
			},
			egress: []core.EgressSecurityRule{
				makeEgressSecurityRule("1", 80),
				makeEgressSecurityRule("2", 81),
			},
			expected: true,
		},
		"change egress - remove": {
			list: &core.SecurityList{
				EgressSecurityRules: []core.EgressSecurityRule{
					makeEgressSecurityRule("1", 80),
					makeEgressSecurityRule("2", 81),
				},
			},
			egress: []core.EgressSecurityRule{
				makeEgressSecurityRule("1", 80),
			},
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := securityListRulesChanged(tc.list, tc.ingress, tc.egress)
			if result != tc.expected {
				t.Errorf("Expected security rules changed to be `%t` but got `%t`", tc.expected, result)
			}
		})
	}
}

func TestUpdate(t *testing.T) {

}
func TestDelete(t *testing.T) {
	// TODO: add more tests instead of a basic acceptance test

	// fakeClient := client.NewFakeClient()
	// mgr := newSecurityListManager(fakeClient).(*securityListManagerImpl)

	// lbSubnetIDs := []string{
	// 	"lb-subnet-1",
	// 	"lb-subnet-2",
	// }
	// lbSubnets := []*core.Subnet{
	// 	{
	// 		ID:        "lb-subnet-1",
	// 		CidrBlock: "lb-subnet-1",
	// 	},
	// 	{
	// 		ID:        "lb-subnet-2",
	// 		CidrBlock: "lb-subnet-2",
	// 	},
	// }
	// lbSecurityLists := []*core.SecurityList{
	// 	{
	// 		ID:                   "lb-subnet-1",
	// 		IngressSecurityRules: []core.IngressSecurityRule{},
	// 		EgressSecurityRules:  []core.EgressSecurityRule{},
	// 	},
	// 	{
	// 		ID:                   "lb-subnet-2",
	// 		IngressSecurityRules: []core.IngressSecurityRule{},
	// 		EgressSecurityRules:  []core.EgressSecurityRule{},
	// 	},
	// }

	// for _, s := range lbSubnets {
	// 	mgr.subnetCache.Add(s)
	// }

	// for _, s := range lbSecurityLists {
	// 	mgr.securityListCache.Add(s)
	// }

}
