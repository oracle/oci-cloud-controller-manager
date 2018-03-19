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
	"context"
	"fmt"
	"reflect"
	"sort"

	"github.com/golang/glog"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/pkg/errors"

	api "k8s.io/api/core/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	sets "k8s.io/apimachinery/pkg/util/sets"
	listersv1 "k8s.io/client-go/listers/core/v1"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
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
	Update(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, sourceCIDRs []string, listenerPort, backendPort, healthCheckPort int) error
	// Delete the security list rules associated with the listener & backends.
	//
	// If the listener is nil, then only the egress rules from the LB's to the backends and the
	// ingress rules from the LB's to the backends will be cleaned up.
	// If the listener is not nil, then the ingress rules to the LB's will be cleaned up.
	Delete(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, listenerPort, backendPort, healthCheckPort int) error
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

func (s *securityListManagerImpl) Update(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, sourceCIDRs []string, listenerPort, backendPort, healthCheckPort int) error {
	if err := s.updateLoadBalancerRules(ctx, lbSubnets, backendSubnets, sourceCIDRs, listenerPort, backendPort, healthCheckPort); err != nil {
		return err
	}

	return s.updateBackendRules(ctx, lbSubnets, backendSubnets, backendPort, healthCheckPort)
}

func (s *securityListManagerImpl) Delete(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, listenerPort, backendPort, healthCheckPort int) error {
	noSubnets := []*core.Subnet{}
	noSourceCIDRs := []string{}

	err := s.updateLoadBalancerRules(ctx, lbSubnets, noSubnets, noSourceCIDRs, listenerPort, backendPort, healthCheckPort)
	if err != nil {
		return err
	}

	return s.updateBackendRules(ctx, noSubnets, backendSubnets, backendPort, healthCheckPort)
}

// updateBackendRules handles adding ingress rules to the backend subnets from the load balancer subnets.
func (s *securityListManagerImpl) updateBackendRules(ctx context.Context, lbSubnets []*core.Subnet, nodeSubnets []*core.Subnet, backendPort, healthCheckPort int) error {
	for _, subnet := range nodeSubnets {
		secList, etag, err := getDefaultSecurityList(ctx, s.client.Networking(), subnet.SecurityListIds)
		if err != nil {
			return errors.Wrapf(err, "get security list for subnet %q", *subnet.Id)
		}

		ingressRules := getNodeIngressRules(secList.IngressSecurityRules, lbSubnets, backendPort, s.serviceLister)
		ingressRules = getNodeIngressRules(ingressRules, lbSubnets, healthCheckPort, s.serviceLister)

		if !securityListRulesChanged(secList, ingressRules, secList.EgressSecurityRules) {
			glog.V(4).Infof("No changes for node subnet security list %q", *secList.Id)
			continue
		}

		err = s.updateSecurityListRules(ctx, *secList.Id, etag, ingressRules, secList.EgressSecurityRules)
		if err != nil {
			return errors.Wrapf(err, "update security list rules %q for subnet %q", *secList.Id, *subnet.Id)
		}
	}

	return nil
}

// updateLoadBalancerRules handles updating the ingress and egress rules for the load balance subnets.
// If the listener is nil, then only egress rules from the load balancer to the backend subnets will be checked.
func (s *securityListManagerImpl) updateLoadBalancerRules(ctx context.Context, lbSubnets []*core.Subnet, nodeSubnets []*core.Subnet, sourceCIDRs []string, listenerPort, backendPort, healthCheckPort int) error {
	for _, lbSubnet := range lbSubnets {
		lbSecurityList, etag, err := getDefaultSecurityList(ctx, s.client.Networking(), lbSubnet.SecurityListIds)
		if err != nil {
			return errors.Wrapf(err, "get lb security list for subnet %q", *lbSubnet.Id)
		}

		lbEgressRules := getLoadBalancerEgressRules(lbSecurityList.EgressSecurityRules, nodeSubnets, backendPort, s.serviceLister)
		lbEgressRules = getLoadBalancerEgressRules(lbEgressRules, nodeSubnets, healthCheckPort, s.serviceLister)

		lbIngressRules := lbSecurityList.IngressSecurityRules
		if listenerPort != 0 {
			lbIngressRules = getLoadBalancerIngressRules(lbIngressRules, sourceCIDRs, listenerPort, s.serviceLister)
		}

		if !securityListRulesChanged(lbSecurityList, lbIngressRules, lbEgressRules) {
			glog.V(4).Infof("No changes for lb subnet security list %q", *lbSecurityList.Id)
			continue
		}

		err = s.updateSecurityListRules(ctx, *lbSecurityList.Id, etag, lbIngressRules, lbEgressRules)
		if err != nil {
			return errors.Wrapf(err, "update lb security list rules %q for subnet %q", *lbSecurityList.Id, *lbSubnet.Id)
		}
	}

	return nil
}

func getDefaultSecurityList(ctx context.Context, n client.NetworkingInterface, ids []string) (*core.SecurityList, string, error) {
	if len(ids) < 1 {
		return nil, "", errors.Errorf("no security lists")
	}

	responses := make([]core.GetSecurityListResponse, len(ids))
	for i, id := range ids {
		response, err := n.GetSecurityList(ctx, id)
		if err != nil {
			return nil, "", err
		}
		responses[i] = response
	}

	sort.Slice(responses, func(i, j int) bool {
		return responses[i].TimeCreated.Before(responses[j].TimeCreated.Time)
	})

	return &responses[0].SecurityList, *responses[0].Etag, nil
}

// securityListRulesChanged compares the ingress rules and egress rules against the rules in the security list. It checks that all the passed in egress & ingress rules
// exist or not in the security list rules. If a rule does not exist then the rules have changed, so an update is required.
func securityListRulesChanged(securityList *core.SecurityList, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) bool {
	if len(ingressRules) != len(securityList.IngressSecurityRules) || len(egressRules) != len(securityList.EgressSecurityRules) {
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
func (s *securityListManagerImpl) updateSecurityListRules(ctx context.Context, id string, etag string, ingressRules []core.IngressSecurityRule, egressRules []core.EgressSecurityRule) error {
	_, err := s.client.Networking().UpdateSecurityList(ctx, core.UpdateSecurityListRequest{
		SecurityListId: &id,
		IfMatch:        &etag,
		UpdateSecurityListDetails: core.UpdateSecurityListDetails{
			IngressSecurityRules: ingressRules,
			EgressSecurityRules:  egressRules,
		},
	})
	return err
}

func getNodeIngressRules(rules []core.IngressSecurityRule, lbSubnets []*core.Subnet, port int, serviceLister listersv1.ServiceLister) []core.IngressSecurityRule {
	desired := sets.NewString()
	for _, lbSubnet := range lbSubnets {
		desired.Insert(*lbSubnet.CidrBlock)
	}

	ingressRules := []core.IngressSecurityRule{}

	for _, rule := range rules {
		if rule.TcpOptions == nil || rule.TcpOptions.SourcePortRange != nil || rule.TcpOptions.DestinationPortRange == nil ||
			*rule.TcpOptions.DestinationPortRange.Min != port || *rule.TcpOptions.DestinationPortRange.Max != port {
			// this rule doesn't apply to this service so nothing to do but keep it
			ingressRules = append(ingressRules, rule)
			continue
		}

		if desired.Has(*rule.Source) {
			// This rule still exists so lets keep it
			ingressRules = append(ingressRules, rule)
			desired.Delete(*rule.Source)
			continue
		}

		inUse, err := healthCheckPortInUse(serviceLister, int32(port))
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

func getLoadBalancerIngressRules(rules []core.IngressSecurityRule, sourceCIDRs []string, port int, serviceLister listersv1.ServiceLister) []core.IngressSecurityRule {
	desired := sets.NewString(sourceCIDRs...)

	ingressRules := []core.IngressSecurityRule{}
	for _, rule := range rules {
		if rule.TcpOptions == nil || rule.TcpOptions.SourcePortRange != nil || rule.TcpOptions.DestinationPortRange == nil ||
			*rule.TcpOptions.DestinationPortRange.Min != port || *rule.TcpOptions.DestinationPortRange.Max != port {
			// this rule doesn't apply to this service so nothing to do but keep it
			ingressRules = append(ingressRules, rule)
			continue
		}

		if desired.Has(*rule.Source) {
			// This rule still exists so lets keep it
			ingressRules = append(ingressRules, rule)
			desired.Delete(*rule.Source)
			continue
		}

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

func getLoadBalancerEgressRules(rules []core.EgressSecurityRule, nodeSubnets []*core.Subnet, port int, serviceLister listersv1.ServiceLister) []core.EgressSecurityRule {
	nodeCIDRs := sets.NewString()
	for _, subnet := range nodeSubnets {
		nodeCIDRs.Insert(*subnet.CidrBlock)
	}

	egressRules := []core.EgressSecurityRule{}
	for _, rule := range rules {
		if rule.TcpOptions == nil || rule.TcpOptions.SourcePortRange != nil || rule.TcpOptions.DestinationPortRange == nil ||
			*rule.TcpOptions.DestinationPortRange.Min != port || *rule.TcpOptions.DestinationPortRange.Max != port {
			// this rule doesn't apply to this service so nothing to do but keep it
			egressRules = append(egressRules, rule)
			continue
		}

		if nodeCIDRs.Has(*rule.Destination) {
			// This rule still exists so lets keep it
			egressRules = append(egressRules, rule)
			nodeCIDRs.Delete(*rule.Destination)
			continue
		}

		inUse, err := healthCheckPortInUse(serviceLister, int32(port))
		if err != nil {
			// Unable to determine if this port is in use by another service, so I guess
			// we better err on the safe side and keep the rule.
			glog.Errorf("failed to determine if port: %d is still in use: %v", port, err)
			egressRules = append(egressRules, rule)
			continue
		}

		if inUse {
			// This rule is no longer needed for this service, but is still used
			// by another service, so we must still keep it.
			glog.V(4).Infof("Port %d still in use by another service.", port)
			egressRules = append(egressRules, rule)
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
func makeEgressSecurityRule(cidrBlock string, port int) core.EgressSecurityRule {
	return core.EgressSecurityRule{
		Destination: common.String(cidrBlock),
		Protocol:    common.String(fmt.Sprintf("%d", ProtocolTCP)),
		TcpOptions: &core.TcpOptions{
			DestinationPortRange: &core.PortRange{
				Min: common.Int(port),
				Max: common.Int(port),
			},
		},
		IsStateless: common.Bool(false),
	}
}

// TODO(apryde): UDP support.
func makeIngressSecurityRule(cidrBlock string, port int) core.IngressSecurityRule {
	return core.IngressSecurityRule{
		Source:   common.String(cidrBlock),
		Protocol: common.String(fmt.Sprintf("%d", ProtocolTCP)),
		TcpOptions: &core.TcpOptions{
			DestinationPortRange: &core.PortRange{
				Min: common.Int(port),
				Max: common.Int(port),
			},
		},
		IsStateless: common.Bool(false),
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

func healthCheckPortInUse(serviceLister listersv1.ServiceLister, port int32) (bool, error) {
	if port != lbNodesHealthCheckPort {
		// This service is using a custom healthcheck port (enabled through setting
		// extenalTrafficPolicy=Local on the service). As this port is unique
		// per service, we know no other service will be using this port too.
		return false, nil
	}

	// This service is using the default healthcheck port, so we must check if
	// any other service is also using this default healthcheck port.
	serviceList, err := serviceLister.List(labels.Everything())
	if err != nil {
		return false, err
	}
	for _, service := range serviceList {
		if service.Spec.Type == api.ServiceTypeLoadBalancer {
			healthCheckPath, _ := apiservice.GetServiceHealthCheckPathPort(service)
			if healthCheckPath == "" {
				// We have found another service using the default port.
				return true, nil
			}
		}
	}
	return false, nil
}

// securityListManagerNOOP implements the securityListManager interface but does
// no logic, so that it can be used to not handle security lists if the user doesn't wish
// to use that feature.
type securityListManagerNOOP struct{}

func (s *securityListManagerNOOP) Update(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, sourceCIDRs []string, listenerPort int, backendPort int, healthCheckPort int) error {
	return nil
}

func (s *securityListManagerNOOP) Delete(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, listenerPort int, backendPort int, healthCheckPort int) error {
	return nil
}

func newSecurityListManagerNOOP() securityListManager {
	return &securityListManagerNOOP{}
}
