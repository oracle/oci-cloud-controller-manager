// Copyright 2017 The Oracle Kubernetes Cloud Controller Manager Authors
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

	baremetal "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
)

func TestGetBackendPort(t *testing.T) {
	backends := []baremetal.Backend{
		{Port: 80},
	}

	port := getBackendPort(backends)
	if port != 80 {
		t.Errorf("expected port 80 but got %d", port)
	}
}

func TestGetNodeIngressRules(t *testing.T) {
	testCases := []struct {
		name         string
		securityList *baremetal.SecurityList
		lbSubnets    []*baremetal.Subnet
		port         uint64
		expected     []baremetal.IngressSecurityRule
	}{
		{
			name: "new ingress",
			securityList: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
				},
			},
			lbSubnets: []*baremetal.Subnet{
				{CIDRBlock: "1"},
				{CIDRBlock: "2"},
			},
			port: 80,
			expected: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("2", 80),
			},
		}, {
			name: "no change",
			securityList: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
				},
			},
			lbSubnets: []*baremetal.Subnet{
				{CIDRBlock: "1"},
				{CIDRBlock: "2"},
			},
			port: 80,
			expected: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("2", 80),
			},
		}, {
			name: "change lb subnet",
			securityList: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
					makeIngressSecurityRule("existing", 9001),
				},
			},
			lbSubnets: []*baremetal.Subnet{
				{CIDRBlock: "1"},
				{CIDRBlock: "3"},
			},
			port: 80,
			expected: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("existing", 9001),
				makeIngressSecurityRule("3", 80),
			},
		}, {
			name: "remove lb subnets",
			securityList: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
					makeIngressSecurityRule("existing", 9001),
				},
			},
			lbSubnets: []*baremetal.Subnet{},
			port:      80,
			expected: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("existing", 9001),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules := getNodeIngressRules(tc.securityList, tc.lbSubnets, tc.port)
			if !reflect.DeepEqual(rules, tc.expected) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expected, rules)
			}
		})
	}
}

func TestGetLoadBalancerIngressRules(t *testing.T) {
	testCases := []struct {
		name         string
		securityList *baremetal.SecurityList
		sourceCIDRs  []string
		port         uint64
		expected     []baremetal.IngressSecurityRule
	}{
		{
			name: "new source cidrs",
			securityList: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
				},
			},
			sourceCIDRs: []string{
				"1",
				"2",
			},
			port: 80,
			expected: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("2", 80),
			},
		}, {
			name: "no change",
			securityList: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
				},
			},
			sourceCIDRs: []string{
				"1",
				"2",
			},
			port: 80,
			expected: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("2", 80),
			},
		}, {
			name: "change source cidr",
			securityList: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
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
			port: 80,
			expected: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("existing", 9001),
				makeIngressSecurityRule("3", 80),
			},
		}, {
			name: "remove source cidrs",
			securityList: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
					makeIngressSecurityRule("existing", 9000),
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 80),
					makeIngressSecurityRule("existing", 9001),
				},
			},
			sourceCIDRs: []string{},
			port:        80,
			expected: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("existing", 9000),
				makeIngressSecurityRule("existing", 9001),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules := getLoadBalancerIngressRules(tc.securityList, tc.sourceCIDRs, tc.port)
			if !reflect.DeepEqual(rules, tc.expected) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expected, rules)
			}
		})
	}
}

func TestGetLoadBalancerEgressRules(t *testing.T) {
	testCases := []struct {
		name         string
		securityList *baremetal.SecurityList
		subnets      []*baremetal.Subnet
		port         uint64
		expected     []baremetal.EgressSecurityRule
	}{
		{
			name: "new egress",
			securityList: &baremetal.SecurityList{
				EgressSecurityRules: []baremetal.EgressSecurityRule{
					makeEgressSecurityRule("existing", 9000),
				},
			},
			subnets: []*baremetal.Subnet{
				{CIDRBlock: "1"},
				{CIDRBlock: "2"},
			},
			port: 80,
			expected: []baremetal.EgressSecurityRule{
				makeEgressSecurityRule("existing", 9000),
				makeEgressSecurityRule("1", 80),
				makeEgressSecurityRule("2", 80),
			},
		}, {
			name: "no change",
			securityList: &baremetal.SecurityList{
				EgressSecurityRules: []baremetal.EgressSecurityRule{
					makeEgressSecurityRule("existing", 9000),
					makeEgressSecurityRule("1", 80),
					makeEgressSecurityRule("2", 80),
				},
			},
			subnets: []*baremetal.Subnet{
				{CIDRBlock: "1"},
				{CIDRBlock: "2"},
			},
			port: 80,
			expected: []baremetal.EgressSecurityRule{
				makeEgressSecurityRule("existing", 9000),
				makeEgressSecurityRule("1", 80),
				makeEgressSecurityRule("2", 80),
			},
		}, {
			name: "change node subnet",
			securityList: &baremetal.SecurityList{
				EgressSecurityRules: []baremetal.EgressSecurityRule{
					makeEgressSecurityRule("existing", 9000),
					makeEgressSecurityRule("1", 80),
					makeEgressSecurityRule("2", 80),
					makeEgressSecurityRule("existing", 9001),
				},
			},
			subnets: []*baremetal.Subnet{
				{CIDRBlock: "1"},
				{CIDRBlock: "3"},
			},
			port: 80,
			expected: []baremetal.EgressSecurityRule{
				makeEgressSecurityRule("existing", 9000),
				makeEgressSecurityRule("1", 80),
				makeEgressSecurityRule("existing", 9001),
				makeEgressSecurityRule("3", 80),
			},
		}, {
			name: "remove node subnets",
			securityList: &baremetal.SecurityList{
				EgressSecurityRules: []baremetal.EgressSecurityRule{
					makeEgressSecurityRule("existing", 9000),
					makeEgressSecurityRule("1", 80),
					makeEgressSecurityRule("2", 80),
					makeEgressSecurityRule("existing", 9001),
				},
			},
			subnets: []*baremetal.Subnet{},
			port:    80,
			expected: []baremetal.EgressSecurityRule{
				makeEgressSecurityRule("existing", 9000),
				makeEgressSecurityRule("existing", 9001),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules := getLoadBalancerEgressRules(tc.securityList, tc.subnets, tc.port)
			if !reflect.DeepEqual(rules, tc.expected) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expected, rules)
			}
		})
	}
}

func TestMakeIngressSecurityRuleHasProtocolOptions(t *testing.T) {
	cdirRange := "10.0.0.0/16"
	port := uint64(80)
	rule := makeIngressSecurityRule(cdirRange, port)
	if rule.TCPOptions == nil && rule.UDPOptions == nil {
		t.Errorf("makeIngressSecurityRule(%q, %d) did not set protocol options",
			cdirRange, port)
	}
}

func TestMakeEgressSecurityRuleHasProtocolOptions(t *testing.T) {
	cdirRange := "10.0.0.0/16"
	port := uint64(80)
	rule := makeEgressSecurityRule(cdirRange, port)
	if rule.TCPOptions == nil && rule.UDPOptions == nil {
		t.Errorf("makeEgressSecurityRule(%q, %d) did not set protocol options",
			cdirRange, port)
	}
}

func TestGetSecurityList(t *testing.T) {

	testCases := []struct {
		name     string
		calls    []string
		subnet   *baremetal.Subnet
		cache    *baremetal.SecurityList
		client   *baremetal.SecurityList
		expected *baremetal.SecurityList
	}{
		{
			name:  "cache hit",
			calls: []string{},
			subnet: &baremetal.Subnet{
				SecurityListIDs: []string{"list"},
			},
			cache: &baremetal.SecurityList{
				ID:          "list",
				DisplayName: "cache",
			},
			client: nil,
			expected: &baremetal.SecurityList{
				ID:          "list",
				DisplayName: "cache",
			},
		}, {
			name:  "cache miss",
			calls: []string{"get-default-security-list"},
			subnet: &baremetal.Subnet{
				SecurityListIDs: []string{"list"},
			},
			cache: nil,
			client: &baremetal.SecurityList{
				ID:          "list",
				DisplayName: "client",
			},
			expected: &baremetal.SecurityList{
				ID:          "list",
				DisplayName: "client",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := client.NewFakeClient()
			mgr := newSecurityListManager(fakeClient).(*securityListManagerImpl)
			if tc.cache != nil {
				mgr.securityListCache.Add(tc.cache)
			}

			if tc.client != nil {
				fakeClient.DefaultSecurityLists[tc.client.ID] = tc.client
			}

			result, err := mgr.getSecurityList(tc.subnet)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(tc.calls, fakeClient.Calls) {
				t.Errorf("expected fake client calls\n%+v\nbut got\n%+v", tc.calls, fakeClient.Calls)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected security list\n%+v\nbut got\n%+v", tc.expected, result)
			}
		})
	}
}

func TestSecurityListRulesChanged(t *testing.T) {
	testCases := map[string]struct {
		list     *baremetal.SecurityList
		ingress  []baremetal.IngressSecurityRule
		egress   []baremetal.EgressSecurityRule
		expected bool
	}{
		"no change": {
			list: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
					makeIngressSecurityRule("1", 80),
				},
				EgressSecurityRules: []baremetal.EgressSecurityRule{
					makeEgressSecurityRule("1", 80),
				},
			},
			ingress: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("1", 80),
			},
			egress: []baremetal.EgressSecurityRule{
				makeEgressSecurityRule("1", 80),
			},
			expected: false,
		},
		"change ingress - add": {
			list: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
					makeIngressSecurityRule("1", 80),
				},
			},
			ingress: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("1", 80),
				makeIngressSecurityRule("2", 81),
			},
			expected: true,
		},
		"change ingress - remove": {
			list: &baremetal.SecurityList{
				IngressSecurityRules: []baremetal.IngressSecurityRule{
					makeIngressSecurityRule("1", 80),
					makeIngressSecurityRule("2", 81),
				},
			},
			ingress: []baremetal.IngressSecurityRule{
				makeIngressSecurityRule("1", 80),
			},
			expected: true,
		},
		"change egress - add": {
			list: &baremetal.SecurityList{
				EgressSecurityRules: []baremetal.EgressSecurityRule{
					makeEgressSecurityRule("1", 80),
				},
			},
			egress: []baremetal.EgressSecurityRule{
				makeEgressSecurityRule("1", 80),
				makeEgressSecurityRule("2", 81),
			},
			expected: true,
		},
		"change egress - remove": {
			list: &baremetal.SecurityList{
				EgressSecurityRules: []baremetal.EgressSecurityRule{
					makeEgressSecurityRule("1", 80),
					makeEgressSecurityRule("2", 81),
				},
			},
			egress: []baremetal.EgressSecurityRule{
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
	// lbSubnets := []*baremetal.Subnet{
	// 	{
	// 		ID:        "lb-subnet-1",
	// 		CIDRBlock: "lb-subnet-1",
	// 	},
	// 	{
	// 		ID:        "lb-subnet-2",
	// 		CIDRBlock: "lb-subnet-2",
	// 	},
	// }
	// lbSecurityLists := []*baremetal.SecurityList{
	// 	{
	// 		ID:                   "lb-subnet-1",
	// 		IngressSecurityRules: []baremetal.IngressSecurityRule{},
	// 		EgressSecurityRules:  []baremetal.EgressSecurityRule{},
	// 	},
	// 	{
	// 		ID:                   "lb-subnet-2",
	// 		IngressSecurityRules: []baremetal.IngressSecurityRule{},
	// 		EgressSecurityRules:  []baremetal.EgressSecurityRule{},
	// 	},
	// }

	// for _, s := range lbSubnets {
	// 	mgr.subnetCache.Add(s)
	// }

	// for _, s := range lbSecurityLists {
	// 	mgr.securityListCache.Add(s)
	// }

}
