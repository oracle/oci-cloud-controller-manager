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
	"fmt"
	"reflect"

	"github.com/golang/glog"
	baremetal "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"

	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	listersv1 "k8s.io/client-go/listers/core/v1"
)

const (
	// ProtocolTCP is the IANA decimal protocol number for the Transmission
	// Control Protocol (TCP).
	ProtocolTCP = 6
	// ProtocolUDP is the IANA decimal protocol number for the User
	// Datagram Protocol (UDP).
	ProtocolUDP = 17
)

type securityListManager interface {
	// Update the security list rules associated with the listener and backends.
	//
	// Ingress rules added:
	// 		from source cidrs to lb subnets on the listener port
	// 		from LB subnets to backend subnets on the backend port
	// Egress rules added:
	// 		from LB subnets to backend subnets on the backend port
	Update(lbSubnets []*baremetal.Subnet, backendSubnets []*baremetal.Subnet, sourceCIDRs []string, listenerPort uint64, backendPort uint64) error
	// Delete the security list rules associated with the listener & backends.
	//
	// If the listener is nil, then only the egress rules from the LB's to the backends and the
	// ingress rules from the LB's to the backends will be cleaned up.
	// If the listener is not nil, then the ingress rules to the LB's will be cleaned up.
	Delete(lbSubnets []*baremetal.Subnet, backendSubnets []*baremetal.Subnet, listenerPort uint64, backendPort uint64) error
}

type securityListManagerImpl struct {
	client        client.Interface
	serviceLister listersv1.ServiceLister
}

func newSecurityListManager(client client.Interface, serviceLister listersv1.ServiceLister) securityListManager {
	return &securityListManagerImpl{
		client:        client,
		serviceLister: serviceLister,
	}
}

func (s *securityListManagerImpl) Update(
	lbSubnets []*baremetal.Subnet,
	backendSubnets []*baremetal.Subnet,
	sourceCIDRs []string,
	listenerPort uint64,
	backendPort uint64) error {

	err := s.updateLoadBalancerRules(lbSubnets, backendSubnets, sourceCIDRs, listenerPort, backendPort)
	if err != nil {
		return err
	}

	return s.updateBackendRules(lbSubnets, backendSubnets, backendPort)
}

func (s *securityListManagerImpl) Delete(
	lbSubnets []*baremetal.Subnet,
	backendSubnets []*baremetal.Subnet,
	listenerPort uint64,
	backendPort uint64,
) error {
	noSubnets := []*baremetal.Subnet{}
	noSourceCIDRs := []string{}

	err := s.updateLoadBalancerRules(lbSubnets, noSubnets, noSourceCIDRs, listenerPort, backendPort)
	if err != nil {
		return err
	}

	return s.updateBackendRules(noSubnets, backendSubnets, backendPort)
}

// updateBackendRules handles adding ingress rules to the backend subnets from the load balancer subnets.
func (s *securityListManagerImpl) updateBackendRules(lbSubnets []*baremetal.Subnet, nodeSubnets []*baremetal.Subnet, backendPort uint64) error {
	for _, subnet := range nodeSubnets {
		secList, err := s.client.GetDefaultSecurityList(subnet)
		if err != nil {
			return fmt.Errorf("get security list for subnet `%s`: %v", subnet.ID, err)
		}

		ingressRules := getNodeIngressRules(secList, lbSubnets, backendPort)

		if !securityListRulesChanged(secList, ingressRules, secList.EgressSecurityRules) {
			glog.V(4).Infof("No changes for node subnet security list `%s`", secList.ID)
			continue
		}

		err = s.updateSecurityListRules(secList.ID, secList.ETag, ingressRules, secList.EgressSecurityRules)
		if err != nil {
			return fmt.Errorf("update security list rules `%s` for subnet `%s: %v", secList.ID, subnet.ID, err)
		}
	}

	return nil
}

// updateLoadBalancerRules handles updating the ingress and egress rules for the load balance subnets.
// If the listener is nil, then only egress rules from the load balancer to the backend subnets will be checked.
func (s *securityListManagerImpl) updateLoadBalancerRules(lbSubnets []*baremetal.Subnet, nodeSubnets []*baremetal.Subnet, sourceCIDRs []string, listenerPort uint64, backendPort uint64) error {
	for _, lbSubnet := range lbSubnets {
		lbSecurityList, err := s.client.GetDefaultSecurityList(lbSubnet)
		if err != nil {
			return fmt.Errorf("get lb security list for subnet `%s`: %v", lbSubnet.ID, err)
		}

		lbEgressRules := getLoadBalancerEgressRules(lbSecurityList, nodeSubnets, backendPort)

		lbIngressRules := lbSecurityList.IngressSecurityRules
		if listenerPort != 0 {
			lbIngressRules = getLoadBalancerIngressRules(lbSecurityList, sourceCIDRs, listenerPort, s.serviceLister)
		}

		if !securityListRulesChanged(lbSecurityList, lbIngressRules, lbEgressRules) {
			glog.V(4).Infof("No changes for lb subnet security list `%s`", lbSecurityList.ID)
			continue
		}

		err = s.updateSecurityListRules(lbSecurityList.ID, lbSecurityList.ETag, lbIngressRules, lbEgressRules)
		if err != nil {
			return fmt.Errorf("update lb security list rules `%s` for subnet `%s: %v", lbSecurityList.ID, lbSubnet.ID, err)
		}
	}

	return nil
}

// securityListRulesChanged compares the ingress rules and egress rules against the rules in the security list. It checks that all the passed in egress & ingress rules
// exist or not in the security list rules. If a rule does not exist then the rules have changed, so an update is required.
func securityListRulesChanged(securityList *baremetal.SecurityList, ingressRules []baremetal.IngressSecurityRule, egressRules []baremetal.EgressSecurityRule) bool {
	if len(ingressRules) != len(securityList.IngressSecurityRules) ||
		len(egressRules) != len(securityList.EgressSecurityRules) {
		return true
	}

	for _, rule := range ingressRules {
		found := false
		for _, existingRule := range securityList.IngressSecurityRules {
			if reflect.DeepEqual(existingRule, rule) {
				found = true
				break
			}
		}

		if !found {
			return true
		}
	}

	for _, rule := range egressRules {
		found := false
		for _, existingRule := range securityList.EgressSecurityRules {
			if reflect.DeepEqual(existingRule, rule) {
				found = true
				break
			}
		}

		if !found {
			return true
		}
	}

	return false
}

// updateSecurityListRules updates the security list rules and saves the security list in the cache upon successful update.
func (s *securityListManagerImpl) updateSecurityListRules(securityListID string, etag string, ingressRules []baremetal.IngressSecurityRule, egressRules []baremetal.EgressSecurityRule) error {
	_, err := s.client.UpdateSecurityList(securityListID, &baremetal.UpdateSecurityListOptions{
		IfMatchDisplayNameOptions: baremetal.IfMatchDisplayNameOptions{
			IfMatchOptions: baremetal.IfMatchOptions{
				IfMatch: etag,
			},
		},
		EgressRules:  egressRules,
		IngressRules: ingressRules,
	})
	return err
}

func getNodeIngressRules(securityList *baremetal.SecurityList, lbSubnets []*baremetal.Subnet, port uint64) []baremetal.IngressSecurityRule {
	desired := sets.NewString()
	for _, lbSubnet := range lbSubnets {
		desired.Insert(lbSubnet.CIDRBlock)
	}

	ingressRules := []baremetal.IngressSecurityRule{}

	for _, rule := range securityList.IngressSecurityRules {
		if rule.TCPOptions == nil || rule.TCPOptions.SourcePortRange != nil || rule.TCPOptions.DestinationPortRange == nil ||
			(rule.TCPOptions.DestinationPortRange.Min != port &&
				rule.TCPOptions.DestinationPortRange.Max != port) {
			// this rule doesn't apply to this service so nothing to do but keep it
			ingressRules = append(ingressRules, rule)
			continue
		}

		if desired.Has(rule.Source) {
			// This rule still exists so lets keep it
			ingressRules = append(ingressRules, rule)
			desired.Delete(rule.Source)
			continue
		}
		// else the actual cidr no longer exists so we don't need to do
		// anything but ignore / delete it.
	}

	if desired.Len() == 0 {
		// actual is the same as desired so there is nothing to do
		return ingressRules
	}

	// All the remaining node cidr's are new and don't have a corresponding rule
	// so we need to create one for each.
	for _, cidr := range desired.List() {
		ingressRules = append(ingressRules, makeIngressSecurityRule(cidr, port))
	}

	return ingressRules
}

func getLoadBalancerIngressRules(lbSecurityList *baremetal.SecurityList, sourceCIDRs []string, port uint64, serviceLister listersv1.ServiceLister) []baremetal.IngressSecurityRule {
	desired := sets.NewString(sourceCIDRs...)

	ingressRules := []baremetal.IngressSecurityRule{}
	for _, rule := range lbSecurityList.IngressSecurityRules {
		if rule.TCPOptions == nil || rule.TCPOptions.SourcePortRange != nil || rule.TCPOptions.DestinationPortRange == nil ||
			(rule.TCPOptions.DestinationPortRange.Min != port &&
				rule.TCPOptions.DestinationPortRange.Max != port) {
			// this rule doesn't apply to this service so nothing to do but keep it
			ingressRules = append(ingressRules, rule)
			continue
		}

		if desired.Has(rule.Source) {
			// This rule still exists so lets keep it
			ingressRules = append(ingressRules, rule)
			desired.Delete(rule.Source)
			continue
		}

		if rule.TCPOptions.DestinationPortRange.Min == port && rule.TCPOptions.DestinationPortRange.Max == port {
			inUse, err := portInUse(serviceLister, int32(port))
			if err != nil {
				// Unable to determine if this port is in use by another service, so I guess
				// we better err on the safe side and keep the rule.
				glog.Errorf("failed to determine if port: %d is still in use: %v", port, err)
				ingressRules = append(ingressRules, rule)
				continue
			}

			if inUse {
				// This rule is no longer needed for this service, but is still used
				// by another service, so we must still keep it.
				glog.V(4).Infof("Port %d still in use by another service.", port)
				ingressRules = append(ingressRules, rule)
				continue
			}
		}

		// else the actual cidr no longer exists so we don't need to do
		// anything but ignore / delete it.
	}

	if desired.Len() == 0 {
		// actual is the same as desired so there is nothing to do
		return ingressRules
	}

	// All the remaining node cidr's are new and don't have a corresponding rule
	// so we need to create one for each.
	for _, cidr := range desired.List() {
		ingressRules = append(ingressRules, makeIngressSecurityRule(cidr, port))
	}

	return ingressRules
}

func getLoadBalancerEgressRules(lbSecurityList *baremetal.SecurityList, nodeSubnets []*baremetal.Subnet, port uint64) []baremetal.EgressSecurityRule {
	nodeCIDRs := sets.NewString()
	for _, subnet := range nodeSubnets {
		nodeCIDRs.Insert(subnet.CIDRBlock)
	}

	egressRules := []baremetal.EgressSecurityRule{}
	for _, rule := range lbSecurityList.EgressSecurityRules {
		if rule.TCPOptions == nil || rule.TCPOptions.SourcePortRange != nil || rule.TCPOptions.DestinationPortRange == nil ||
			(rule.TCPOptions.DestinationPortRange.Min != port &&
				rule.TCPOptions.DestinationPortRange.Max != port) {
			// this rule doesn't apply to this service so nothing to do but keep it
			egressRules = append(egressRules, rule)
			continue
		}

		if nodeCIDRs.Has(rule.Destination) {
			// This rule still exists so lets keep it
			egressRules = append(egressRules, rule)
			nodeCIDRs.Delete(rule.Destination)
			continue
		}
		// else the actual cidr no longer exists so we don't need to do
		// anything but ignore / delete it.
	}

	if nodeCIDRs.Len() == 0 {
		// actual is the same as desired so there is nothing to do
		return egressRules
	}

	// All the remaining node cidr's are new and don't have a corresponding rule
	// so we need to create one for each.
	for _, desired := range nodeCIDRs.List() {
		egressRules = append(egressRules, makeEgressSecurityRule(desired, port))
	}

	return egressRules
}

// TODO(apryde): UDP support.
func makeEgressSecurityRule(cidrBlock string, port uint64) baremetal.EgressSecurityRule {
	return baremetal.EgressSecurityRule{
		Destination: cidrBlock,
		Protocol:    fmt.Sprintf("%d", ProtocolTCP),
		TCPOptions: &baremetal.TCPOptions{
			DestinationPortRange: &baremetal.PortRange{
				Min: port,
				Max: port,
			},
		},
		IsStateless: false,
	}
}

// TODO(apryde): UDP support.
func makeIngressSecurityRule(cidrBlock string, port uint64) baremetal.IngressSecurityRule {
	return baremetal.IngressSecurityRule{
		Source:   cidrBlock,
		Protocol: fmt.Sprintf("%d", ProtocolTCP),
		TCPOptions: &baremetal.TCPOptions{
			DestinationPortRange: &baremetal.PortRange{
				Min: port,
				Max: port,
			},
		},
		IsStateless: false,
	}
}

func portInUse(serviceLister listersv1.ServiceLister, port int32) (bool, error) {
	serviceList, err := serviceLister.List(labels.Everything())
	if err != nil {
		return false, err
	}
	for _, service := range serviceList {
		if service.Spec.Type == api.ServiceTypeLoadBalancer {
			for _, p := range service.Spec.Ports {
				if p.Port == port {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// securityListManagerNOOP implements the securityListManager interface but does
// no logic, so that it can be used to not handle security lists if the user doesn't wish
// to use that feature.
type securityListManagerNOOP struct{}

func (s *securityListManagerNOOP) Update(lbSubnets []*baremetal.Subnet, backendSubnets []*baremetal.Subnet, sourceCIDRs []string, listenerPort uint64, backendPort uint64) error {
	return nil
}

func (s *securityListManagerNOOP) Delete(lbSubnets []*baremetal.Subnet, backendSubnets []*baremetal.Subnet, listenerPort uint64, backendPort uint64) error {
	return nil
}

func newSecurityListManagerNOOP() securityListManager {
	return &securityListManagerNOOP{}
}
