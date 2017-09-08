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

package bmcs

import (
	"testing"

	baremetal "github.com/oracle/bmcs-go-sdk"
)

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

func TestBasicLoadBalancerSecurityListRuleCreation(t *testing.T) {
	backendSubnetCDIR := "10.0.0.0/16"
	backendSubnets := []*baremetal.Subnet{{
		ID:        "ocid1.subnet.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		CIDRBlock: backendSubnetCDIR,
		SecurityListIDs: []string{
			"ocid1.securitylist.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}}

	lb1CDIR := "10.1.0.0/16"
	lb2CDIR := "10.2.0.0/16"
	lbSubnets := []*baremetal.Subnet{{
		ID:        "ocid1.subnet.oc1.phx.bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		CIDRBlock: lb1CDIR,
		SecurityListIDs: []string{
			"ocid1.securitylist.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}, {
		ID:        "ocid1.subnet.oc1.phx.cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		CIDRBlock: lb2CDIR,
		SecurityListIDs: []string{
			"ocid1.securitylist.oc1.phx.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}}

	securityList := &baremetal.SecurityList{}

	mngr := newSecurityListManager(nil, backendSubnets, lbSubnets)
	mngr.addSecurityListForSubnet(securityList, backendSubnets[0].ID)
	mngr.addSecurityListForSubnet(securityList, lbSubnets[0].ID)
	mngr.addSecurityListForSubnet(securityList, lbSubnets[1].ID)

	port := uint64(80)
	err := mngr.EnsureRulesAdded(port)
	if err != nil {
		t.Fatalf("EnsureRulesAdded(%d) => error: %v", port, err)
	}

	// Check we have 1 EgressSecurityRule
	if len(securityList.EgressSecurityRules) != 1 {
		t.Errorf("Got %d EgressSecurityRules, expected 1", len(securityList.EgressSecurityRules))
	}

	// Check that this rule allows traffic destined for our backend subnet
	if len(securityList.EgressSecurityRules) > 0 && securityList.EgressSecurityRules[0].Destination != backendSubnetCDIR {
		t.Errorf("Expected EgressSecurityRule with Destination: %q, got %q",
			backendSubnetCDIR, securityList.EgressSecurityRules[0].Destination)
	}

	// Check that we have two ingress rules
	if len(securityList.IngressSecurityRules) != 2 {
		t.Errorf("Got %d IngressSecurityRules, expected 2", len(securityList.IngressSecurityRules))
	}

	// Check the first of these allows traffic from the first load balancer
	// subnet
	if len(securityList.IngressSecurityRules) > 0 && securityList.IngressSecurityRules[0].Source != lb1CDIR {
		t.Errorf("Expected IngressSecurityRules with Source: %q, got %q",
			lb1CDIR, securityList.IngressSecurityRules[0].Source)
	}

	// Check the second of these allows traffic from the second load
	// balancer subnet
	if len(securityList.IngressSecurityRules) > 0 && securityList.IngressSecurityRules[1].Source != lb2CDIR {
		t.Errorf("Expected IngressSecurityRules with Source: %q, got %q",
			lb1CDIR, securityList.IngressSecurityRules[1].Source)
	}
}
