// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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

package framework

import (
	"context"
	"time"

	client "github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	core "github.com/oracle/oci-go-sdk/v50/core"
)

// CountSinglePortSecListRules counts the number of 'single port'
// (non-ranged) egress/ingress rules for the specified list and port.
func CountSinglePortSecListRules(oci client.Interface, egressSecListID, ingressSecListID string, port int) (int, int) {
	numEgressRules := CountEgressSinglePortRules(oci, egressSecListID, port)
	numIngressRules := CountIngressSinglePortRules(oci, ingressSecListID, port)
	return numEgressRules, numIngressRules
}

// CountEgressSinglePortRules counts the number of 'single port' (non-ranged)
// egress rules for the specified seclist and port.
// If no client or seclist is provided, then 0 is returned.
func CountEgressSinglePortRules(oci client.Interface, seclistID string, port int) int {
	count := 0
	if oci != nil && seclistID != "" {
		secList, err := oci.Networking().GetSecurityList(context.Background(), seclistID)
		if err != nil {
			Failf("Could not obtain security list: %v", err)
		}
		filteredRules := []core.EgressSecurityRule{}
		for _, rule := range secList.EgressSecurityRules {
			if rule.TcpOptions != nil && rule.TcpOptions.DestinationPortRange != nil &&
				*rule.TcpOptions.DestinationPortRange.Max == port &&
				*rule.TcpOptions.DestinationPortRange.Min == port {
				filteredRules = append(filteredRules, rule)
			}
		}
		count = len(filteredRules)
	}
	return count
}

// HasValidSinglePortEgressRulesAfterPortChange checks the counts of 'single port'
// (non-ranged) egress rules in the provided seclist after a service port change.
func HasValidSinglePortEgressRulesAfterPortChange(oci client.Interface, seclistID string, expectedRuleCount, oldPort, newPort int) bool {
	// SecList checks are optional and only performed if a checlist is specified.
	if seclistID != "" {
		numOldPortRules := CountEgressSinglePortRules(oci, seclistID, oldPort)
		numNewPortRules := CountEgressSinglePortRules(oci, seclistID, newPort)
		if expectedRuleCount != numNewPortRules || numOldPortRules != 0 {
			// If this check is not configured all values will be 0.
			// If the original number of rules does not match the new number of rules, then there is an inconsistency.
			// If the number of rules for original port is not 0 after the change, there is an egress rules leak.
			return false
		}
	}
	return true
}

// WaitForSinglePortEgressRulesAfterPortChangeOrFail waits for the expected
// number of 'single port' (non-ranged) egress rules to be present in the
// specified seclist or fails.
func WaitForSinglePortEgressRulesAfterPortChangeOrFail(oci client.Interface, seclistID string, expectedRuleCount, oldPort, newPort int) {
	for start := time.Now(); time.Since(start) < 10*time.Second; {
		valid := HasValidSinglePortEgressRulesAfterPortChange(oci, seclistID, expectedRuleCount, oldPort, newPort)
		if !valid {
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}
	Failf("Failed: ValidSinglePortEgressRulesAfterPortChangeOrDie : (expectedRuleCount: %d, oldPort: %d, newPort: %d)", expectedRuleCount, oldPort, newPort)
}

// CountIngressSinglePortRules counts the number of 'single port' (non-ranged)
// ingress rules for the specified seclist and port.
// If no client or seclist is provided, then 0 is returned.
func CountIngressSinglePortRules(oci client.Interface, seclistID string, port int) int {
	count := 0
	if oci != nil && seclistID != "" {
		secList, err := oci.Networking().GetSecurityList(context.Background(), seclistID)
		if err != nil {
			Failf("Could not obtain security list: %v", err)
		}
		filteredRules := []core.IngressSecurityRule{}
		for _, rule := range secList.IngressSecurityRules {
			if rule.TcpOptions != nil && rule.TcpOptions.DestinationPortRange != nil &&
				*rule.TcpOptions.DestinationPortRange.Max == port &&
				*rule.TcpOptions.DestinationPortRange.Min == port {
				filteredRules = append(filteredRules, rule)
			}
		}
		count = len(filteredRules)
	}
	return count
}

// HasValidSinglePortIngressRulesAfterPortChange checks the counts of 'single port'
// (non-ranged) egress rules in the provided seclist after a service port change.
func HasValidSinglePortIngressRulesAfterPortChange(oci client.Interface, seclistID string, expectedRuleCount, oldPort, newPort int) bool {
	// SecList checks are optional and only performed if a checlist is specified.
	if seclistID != "" {
		numOldPortRules := CountIngressSinglePortRules(oci, seclistID, oldPort)
		numNewPortRules := CountIngressSinglePortRules(oci, seclistID, newPort)
		if expectedRuleCount != numNewPortRules || numOldPortRules != 0 {
			// If this check is not configured all values will be 0.
			// If the original number of rules does not match the new number of rules, then there is an inconsistency.
			// If the number of rules for original port is not 0 after the change, there is an egress rules leak.
			return false
		}
	}
	return true
}

// WaitForSinglePortIngressRulesAfterPortChangeOrFail waits for the expected
// number of 'single port' (non-ranged) ingress rules to be present in the
// specified seclist or fails.
func WaitForSinglePortIngressRulesAfterPortChangeOrFail(oci client.Interface, seclistID string, expectedRuleCount, oldPort, newPort int) {
	for start := time.Now(); time.Since(start) < 10*time.Second; {
		valid := HasValidSinglePortIngressRulesAfterPortChange(oci, seclistID, expectedRuleCount, oldPort, newPort)
		if !valid {
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}
	Failf("Failed: ValidSinglePortIngressRulesAfterPortChangeOrFail : (expectedRuleCount: %d, oldPort: %d, newPort: %d)", expectedRuleCount, oldPort, newPort)
}
