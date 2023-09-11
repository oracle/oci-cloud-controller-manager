/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"context"
	"time"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/core"
)

// CountSinglePortRules counts the number of 'single port' (non-ranged)
func CountSinglePortRules(oci client.Interface, nsgId string, port int, direction core.SecurityRuleDirectionEnum) int {
	count := 0
	if oci != nil && nsgId != "" {
		_, _, err := oci.Networking().GetNetworkSecurityGroup(context.Background(), nsgId)
		if err != nil {
			Failf("Could not obtain nsg: %v", err)
		}
		response, err := oci.Networking().ListNetworkSecurityGroupSecurityRules(context.Background(), nsgId,
			core.ListNetworkSecurityGroupSecurityRulesDirectionEnum(direction))
		filteredRules := []core.SecurityRule{}
		for _, rule := range response {
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

// HasValidSinglePortRulesAfterPortChangeNSG checks the counts of 'single port'
func HasValidSinglePortRulesAfterPortChangeNSG(oci client.Interface, nsgId string, oldPort, newPort int, direction core.SecurityRuleDirectionEnum) bool {
	if nsgId != "" {
		numOldPortRules := CountSinglePortRules(oci, nsgId, oldPort, direction)
		numNewPortRules := CountSinglePortRules(oci, nsgId, newPort, direction)
		if numOldPortRules != 0 {
			return false
		}
		if numNewPortRules != 1 {
			return false
		}
	}
	return true
}

// WaitForSinglePortRulesAfterPortChangeOrFailNSG waits for the expected rules to be added and validates
// that the rule on the old port is removed and the rule on the new port is added
func WaitForSinglePortRulesAfterPortChangeOrFailNSG(oci client.Interface, nsgId string, oldPort, newPort int, direction core.SecurityRuleDirectionEnum) {
	for start := time.Now(); time.Since(start) < 70*time.Second; {
		valid := HasValidSinglePortRulesAfterPortChangeNSG(oci, nsgId, oldPort, newPort, direction)
		if !valid {
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}
	Failf("Failed: ValidSinglePortRulesAfterPortChangeOrDie Rule %s on NSG for old port still present: oldPort: %d, newPort: %d)", string(direction), oldPort, newPort)
}
