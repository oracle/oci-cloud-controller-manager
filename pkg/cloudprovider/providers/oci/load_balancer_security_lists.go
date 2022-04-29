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

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/core"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	api "k8s.io/api/core/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	sets "k8s.io/apimachinery/pkg/util/sets"
	informersv1 "k8s.io/client-go/informers/core/v1"
	listersv1 "k8s.io/client-go/listers/core/v1"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"
)

const (
	// ProtocolTCP is the IANA decimal protocol number for the Transmission
	// Control Protocol (TCP).
	ProtocolTCP = 6
	// ProtocolUDP is the IANA decimal protocol number for the User
	// Datagram Protocol (UDP).
	ProtocolUDP = 17
)

const (
	// ManagementModeAll denotes the management of security list rules for load
	// balancer ingress/egress, health checkers, and worker ingress/egress.
	ManagementModeAll = "All"
	// ManagementModeFrontend denotes the management of security list rules for load
	// balancer ingress only.
	ManagementModeFrontend = "Frontend"
	// ManagementModeNone denotes the management of no security list rules.
	ManagementModeNone = "None"
)

type portSpec struct {
	ListenerPort      int
	BackendPort       int
	HealthCheckerPort int
}

type securityListManager interface {
	Update(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, sourceCIDRs []string, actualPorts *portSpec, desiredPorts portSpec, isPreserveSource bool) error
	Delete(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, actualPorts portSpec, sourceCIDRs []string, isPreserveSource bool) error
}

type baseSecurityListManager struct {
	client        client.Interface
	serviceLister listersv1.ServiceLister
	securityLists map[string]string

	logger *zap.SugaredLogger
}

type securityListManagerFactory func(mode string) securityListManager

func newSecurityListManager(logger *zap.SugaredLogger, client client.Interface, serviceInformer informersv1.ServiceInformer, securityLists map[string]string, mode string) securityListManager {
	if securityLists == nil {
		securityLists = make(map[string]string)
	}
	baseMgr := baseSecurityListManager{
		client:        client,
		securityLists: securityLists,
		logger:        logger,
	}

	if mode != ManagementModeNone {
		baseMgr.serviceLister = serviceInformer.Lister()
	}

	switch mode {
	case ManagementModeFrontend:
		logger.Infof("Security list management mode: %q. Managing frontend security lists only.", ManagementModeFrontend)
		return &frontendSecurityListManager{baseSecurityListManager: baseMgr}
	case ManagementModeNone:
		logger.Infof("Security list management mode: %q. Not managing security lists.", ManagementModeNone)
		return &securityListManagerNOOP{}
	default:
		logger.Infof("Security list management mode: %q. Managing all security lists.", ManagementModeAll)
		return &defaultSecurityListManager{baseSecurityListManager: baseMgr}
	}
}

// updateBackendRules handles adding ingress rules to the backend subnets from the load balancer subnets.
// TODO: Pass parameters in a struct
func (s *baseSecurityListManager) updateBackendRules(ctx context.Context, lbSubnets []*core.Subnet, nodeSubnets []*core.Subnet, actualPorts *portSpec, desiredPorts portSpec, sourceCIDRs []string, isPreserveSource bool) error {
	for _, subnet := range nodeSubnets {
		secList, etag, err := s.getSecurityList(ctx, subnet)
		if err != nil {
			return errors.Wrapf(err, "get security list for subnet %q", *subnet.Id)
		}

		logger := s.logger.With("securityListID", *secList.Id)

		ingressRules := getNodeIngressRules(logger, secList.IngressSecurityRules, lbSubnets, actualPorts, desiredPorts, s.serviceLister, sourceCIDRs, isPreserveSource)

		if !securityListRulesChanged(secList, ingressRules, secList.EgressSecurityRules) {
			logger.Debug("No changes for node subnet security list")
			continue
		}

		logger.Info("Node subnet security list changed")

		_, err = s.client.Networking().UpdateSecurityList(ctx, *secList.Id, etag, ingressRules, secList.EgressSecurityRules)
		if err != nil {
			return errors.Wrapf(err, "update security list rules %q for subnet %q", *secList.Id, *subnet.Id)
		}
	}

	return nil
}

// updateLoadBalancerRules handles updating the ingress and egress rules for the load balance subnets.
// If the listener is nil, then only egress rules from the load balancer to the backend subnets will be checked.
func (s *baseSecurityListManager) updateLoadBalancerRules(ctx context.Context, lbSubnets []*core.Subnet, nodeSubnets []*core.Subnet, sourceCIDRs []string, actualPorts *portSpec, desiredPorts portSpec) error {
	for _, lbSubnet := range lbSubnets {
		secList, etag, err := s.getSecurityList(ctx, lbSubnet)
		if err != nil {
			return errors.Wrapf(err, "get lb security list for subnet %q", *lbSubnet.Id)
		}

		logger := s.logger.With("securityListID", *secList.Id)

		// 0 denotes nil ports.
		var currentBackEndPort = 0
		var currentHealthCheck = 0
		if actualPorts != nil {
			currentBackEndPort = actualPorts.BackendPort
			currentHealthCheck = actualPorts.HealthCheckerPort
		}

		lbEgressRules := getLoadBalancerEgressRules(logger, secList.EgressSecurityRules, nodeSubnets, currentBackEndPort, desiredPorts.BackendPort, s.serviceLister)
		lbEgressRules = getLoadBalancerEgressRules(logger, lbEgressRules, nodeSubnets, currentHealthCheck, desiredPorts.HealthCheckerPort, s.serviceLister)

		lbIngressRules := secList.IngressSecurityRules
		if desiredPorts.ListenerPort != 0 {
			lbIngressRules = getLoadBalancerIngressRules(logger, lbIngressRules, sourceCIDRs, desiredPorts.ListenerPort, s.serviceLister)
		}

		if !securityListRulesChanged(secList, lbIngressRules, lbEgressRules) {
			logger.Debug("No changes for load balancer subnet security list")
			continue
		}

		logger.Info("Load balancer subnet security list changed")

		_, err = s.client.Networking().UpdateSecurityList(ctx, *secList.Id, etag, lbIngressRules, lbEgressRules)
		if err != nil {
			return errors.Wrapf(err, "update lb security list rules %q for subnet %q", *secList.Id, *lbSubnet.Id)
		}
	}

	return nil
}

func (s *baseSecurityListManager) getSecurityList(ctx context.Context, subnet *core.Subnet) (*core.SecurityList, string, error) {
	if len(subnet.SecurityListIds) < 1 {
		return nil, "", errors.Errorf("no security lists") // should never happen
	}

	// Use the security list from cloud-provider config if provided.
	if id, ok := s.securityLists[*subnet.Id]; ok && sets.NewString(subnet.SecurityListIds...).Has(id) {
		response, err := s.client.Networking().GetSecurityList(ctx, id)
		if err != nil {
			return nil, "", err
		}
		return &response.SecurityList, *response.Etag, nil
	}

	// Otherwise use the oldest security list.
	// NOTE(apryde): This is rather arbitrary but we're probably stuck with it at this point.
	responses := make([]core.GetSecurityListResponse, len(subnet.SecurityListIds))
	for i, id := range subnet.SecurityListIds {
		response, err := s.client.Networking().GetSecurityList(ctx, id)
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

// defaultSecurityListManager manages all security list rules required for
// a Service type=LoadBalancer.
type defaultSecurityListManager struct {
	baseSecurityListManager
}

// Update the security list rules associated with the listener and backends.
//
// Ingress rules added:
// 		from source cidrs to lb subnets on the listener port
// 		from LB subnets to backend subnets on the backend port
// Egress rules added:
// 		from LB subnets to backend subnets on the backend port
func (s *defaultSecurityListManager) Update(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, sourceCIDRs []string, actualPorts *portSpec, desiredPorts portSpec, isPreserveSource bool) error {
	if err := s.updateLoadBalancerRules(ctx, lbSubnets, backendSubnets, sourceCIDRs, actualPorts, desiredPorts); err != nil {
		return err
	}

	return s.updateBackendRules(ctx, lbSubnets, backendSubnets, actualPorts, desiredPorts, sourceCIDRs, isPreserveSource)
}

// Delete the security list rules associated with the listener and backends.
//
// If the listener is nil, then only the egress rules from the LB's to the backends and the
// ingress rules from the LB's to the backends will be cleaned up.
// If the listener is not nil, then the ingress rules to the LB's will be cleaned up.
func (s *defaultSecurityListManager) Delete(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, ports portSpec, sourceCIDRs []string, isPreserveSource bool) error {
	noSubnets := []*core.Subnet{}
	noSourceCIDRs := []string{}

	err := s.updateLoadBalancerRules(ctx, lbSubnets, noSubnets, noSourceCIDRs, &ports, ports)
	if err != nil {
		return err
	}

	return s.updateBackendRules(ctx, noSubnets, backendSubnets, &ports, ports, noSourceCIDRs, isPreserveSource)
}

// frontendSecurityListManager manages only the ingress security list rules required for
// a Service type=LoadBalancer.
type frontendSecurityListManager struct {
	baseSecurityListManager
}

// Update the ingress security list rules associated with the listener.
//
// Ingress rules added:
// 		from source cidrs to lb subnets on the listener port
func (s *frontendSecurityListManager) Update(ctx context.Context, lbSubnets []*core.Subnet, _ []*core.Subnet, sourceCIDRs []string, actualPorts *portSpec, desiredPorts portSpec, isPreserveSource bool) error {
	noSubnets := []*core.Subnet{}
	return s.updateLoadBalancerRules(ctx, lbSubnets, noSubnets, sourceCIDRs, actualPorts, desiredPorts)
}

// Delete the ingress security list rules associated with the listener.
func (s *frontendSecurityListManager) Delete(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, ports portSpec, sourceCIDRs []string, isPreserveSource bool) error {
	noSubnets := []*core.Subnet{}
	noSourceCIDRs := []string{}
	return s.updateLoadBalancerRules(ctx, lbSubnets, noSubnets, noSourceCIDRs, &ports, ports)
}

// securityListManagerNOOP implements the securityListManager interface but does
// no logic, so that it can be used to not handle security lists if the user doesn't wish
// to use that feature.
type securityListManagerNOOP struct{}

func (s *securityListManagerNOOP) Update(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, sourceCIDRs []string, actualPorts *portSpec, ports portSpec, isPreserveSource bool) error {
	return nil
}

func (s *securityListManagerNOOP) Delete(ctx context.Context, lbSubnets []*core.Subnet, backendSubnets []*core.Subnet, ports portSpec, sourceCIDRs []string, isPreserveSource bool) error {
	return nil
}

func newSecurityListManagerNOOP() securityListManager {
	return &securityListManagerNOOP{}
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

func portRangeMatchesSpec(r core.PortRange, ports *portSpec) bool {
	if ports == nil {
		return false
	}
	return (*r.Min == ports.BackendPort && *r.Max == ports.BackendPort) ||
		(*r.Min == ports.HealthCheckerPort && *r.Max == ports.HealthCheckerPort)
}

func getNodeIngressRules(
	logger *zap.SugaredLogger,
	rules []core.IngressSecurityRule,
	lbSubnets []*core.Subnet,
	actualPorts *portSpec,
	desiredPorts portSpec,
	serviceLister listersv1.ServiceLister,
	sourceCIDRs []string,
	isPreserveSource bool,
) []core.IngressSecurityRule {
	// 0 denotes nil ports.
	var currentBackEndPort = 0
	var currentHealthCheckPort = 0
	if actualPorts != nil {
		currentBackEndPort = actualPorts.BackendPort
		currentHealthCheckPort = actualPorts.HealthCheckerPort
	}

	desiredBackend := sets.NewString()
	desiredHealthChecker := sets.NewString()
	for _, lbSubnet := range lbSubnets {
		desiredBackend.Insert(*lbSubnet.CidrBlock)
		desiredHealthChecker.Insert(*lbSubnet.CidrBlock)
	}

	// Additional sourceCIDR rule for NLB only, for source IP preservation
	if isPreserveSource {
		for _, sourceCIDR := range sourceCIDRs {
			desiredBackend.Insert(sourceCIDR)
		}
	}

	ingressRules := []core.IngressSecurityRule{}

	for _, rule := range rules {
		// Remove (do not re-add) any rule that represents the old case when
		// mutating a single ranged backend port or health check port.
		if rule.TcpOptions != nil && rule.TcpOptions.DestinationPortRange != nil &&
			*rule.TcpOptions.DestinationPortRange.Min == *rule.TcpOptions.DestinationPortRange.Max &&
			*rule.TcpOptions.DestinationPortRange.Min != desiredPorts.BackendPort && *rule.TcpOptions.DestinationPortRange.Max != desiredPorts.BackendPort &&
			*rule.TcpOptions.DestinationPortRange.Min != desiredPorts.HealthCheckerPort && *rule.TcpOptions.DestinationPortRange.Max != desiredPorts.HealthCheckerPort {
			var rulePort = *rule.TcpOptions.DestinationPortRange.Min
			if rulePort == currentBackEndPort || rulePort == currentHealthCheckPort {
				logger.With(
					"source", *rule.Source,
					"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
					"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
				).Debug("Deleting load balancer ingres security rule")
				continue
			}
		}

		if rule.TcpOptions == nil || rule.TcpOptions.SourcePortRange != nil || rule.TcpOptions.DestinationPortRange == nil {
			// this rule doesn't apply to this service so nothing to do but keep it
			ingressRules = append(ingressRules, rule)
			continue
		}

		r := *rule.TcpOptions.DestinationPortRange
		if !(portRangeMatchesSpec(r, &desiredPorts) || portRangeMatchesSpec(r, actualPorts)) {
			// this rule doesn't apply to this service so nothing to do but keep it
			ingressRules = append(ingressRules, rule)
			continue
		}

		if *r.Max == desiredPorts.BackendPort && desiredBackend.Has(*rule.Source) {
			// This rule still exists so lets keep it
			ingressRules = append(ingressRules, rule)
			desiredBackend.Delete(*rule.Source)
			continue
		}

		if *r.Max == desiredPorts.HealthCheckerPort && desiredHealthChecker.Has(*rule.Source) {
			// This rule still exists so lets keep it
			ingressRules = append(ingressRules, rule)
			desiredHealthChecker.Delete(*rule.Source)
			continue
		} else if *r.Max == desiredPorts.HealthCheckerPort {
			inUse, err := healthCheckPortInUse(serviceLister, int32(desiredPorts.HealthCheckerPort))
			if err != nil {
				logger.Errorf("failed to determine if port: %d is still in use: %v", desiredPorts.HealthCheckerPort, err)
				ingressRules = append(ingressRules, rule)
			} else if inUse {
				logger.Infof("Port %d still in use by another service.", desiredPorts.HealthCheckerPort)
				ingressRules = append(ingressRules, rule)
			}
		}

		// else the actual cidr no longer exists so we don't need to do
		// anything but ignore / delete it.
		logger.With(
			"source", *rule.Source,
			"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
			"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
		).Debug("Deleting node ingres security rule")
	}

	if desiredBackend.Len() == 0 && desiredHealthChecker.Len() == 0 {
		// actual is the same as desired so there is nothing to do
		return ingressRules
	}

	// All the remaining node cidr's are new and don't have a corresponding rule
	// so we need to create one for each.
	if desiredPorts.BackendPort != 0 { // Can happen when there are no backends.
		for _, cidr := range desiredBackend.List() {
			rule := makeIngressSecurityRule(cidr, desiredPorts.BackendPort)
			logger.With(
				"source", *rule.Source,
				"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
				"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
			).Debug("Adding node port ingress security rule")
			ingressRules = append(ingressRules, rule)
		}
	}
	if desiredPorts.HealthCheckerPort != 0 {
		for _, cidr := range desiredHealthChecker.List() {
			rule := makeIngressSecurityRule(cidr, desiredPorts.HealthCheckerPort)
			logger.With(
				"source", *rule.Source,
				"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
				"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
			).Debug("Adding node port ingress security rule")
			ingressRules = append(ingressRules, rule)
		}
	}

	return ingressRules
}

func getLoadBalancerIngressRules(
	logger *zap.SugaredLogger,
	rules []core.IngressSecurityRule,
	sourceCIDRs []string, port int,
	serviceLister listersv1.ServiceLister,
) []core.IngressSecurityRule {
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
			logger.With(zap.Error(err), "port", port).Error("Failed to determine if port still in use")
			ingressRules = append(ingressRules, rule)
			continue
		}

		if inUse {
			// This rule is no longer needed for this service, but is still used
			// by another service, so we must still keep it.
			logger.With("port", port).Debug("Port still in use by another service.")
			ingressRules = append(ingressRules, rule)
			continue
		}

		// else the actual cidr no longer exists so we don't need to do
		// anything but ignore / delete it.
		logger.With(
			"source", *rule.Source,
			"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
			"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
		).Debug("Deleting load balancer ingres security rule")
	}

	if desired.Len() == 0 {
		// actual is the same as desired so there is nothing to do
		return ingressRules
	}

	// All the remaining node cidr's are new and don't have a corresponding rule
	// so we need to create one for each.
	for _, cidr := range desired.List() {
		rule := makeIngressSecurityRule(cidr, port)
		logger.With(
			"source", *rule.Source,
			"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
			"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
		).Debug("Adding load balancer ingress security rule")
		ingressRules = append(ingressRules, rule)
	}

	return ingressRules
}

func getLoadBalancerEgressRules(
	logger *zap.SugaredLogger,
	rules []core.EgressSecurityRule,
	nodeSubnets []*core.Subnet,
	actualPort, desiredPort int,
	serviceLister listersv1.ServiceLister,
) []core.EgressSecurityRule {
	nodeCIDRs := sets.NewString()
	for _, subnet := range nodeSubnets {
		nodeCIDRs.Insert(*subnet.CidrBlock)
	}

	egressRules := []core.EgressSecurityRule{}
	for _, rule := range rules {
		// Remove (do not re-add) any rule that represents the old case when mutating a single ranged port.
		if rule.TcpOptions != nil && rule.TcpOptions.DestinationPortRange != nil &&
			*rule.TcpOptions.DestinationPortRange.Min == *rule.TcpOptions.DestinationPortRange.Max &&
			*rule.TcpOptions.DestinationPortRange.Min != desiredPort && *rule.TcpOptions.DestinationPortRange.Max != desiredPort &&
			*rule.TcpOptions.DestinationPortRange.Min == actualPort && *rule.TcpOptions.DestinationPortRange.Max == actualPort {
			logger.With(
				"destination", *rule.Destination,
				"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
				"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
			).Debug("Deleting load balancer ingres security rule")
			continue
		}

		if rule.TcpOptions == nil || rule.TcpOptions.SourcePortRange != nil || rule.TcpOptions.DestinationPortRange == nil ||
			*rule.TcpOptions.DestinationPortRange.Min != desiredPort || *rule.TcpOptions.DestinationPortRange.Max != desiredPort {
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

		inUse, err := healthCheckPortInUse(serviceLister, int32(desiredPort))
		if err != nil {
			// Unable to determine if this port is in use by another service, so I guess
			// we better err on the safe side and keep the rule.
			logger.With(zap.Error(err), "port", desiredPort).Error("Failed to determine if port is still in use")
			egressRules = append(egressRules, rule)
			continue
		}

		if inUse {
			// This rule is no longer needed for this service, but is still used
			// by another service, so we must still keep it.
			logger.With("port", desiredPort).Debug("Port still in use by another service.")
			egressRules = append(egressRules, rule)
			continue
		}

		// else the actual cidr no longer exists so we don't need to do
		// anything but ignore / delete it.
		logger.With(
			"destination", *rule.Destination,
			"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
			"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
		).Debug("Deleting load balancer egress security rule")
	}

	if nodeCIDRs.Len() == 0 {
		// actual is the same as desired so there is nothing to do
		return egressRules
	}

	// All the remaining node cidr's are new and don't have a corresponding rule
	// so we need to create one for each.
	for _, desired := range nodeCIDRs.List() {
		rule := makeEgressSecurityRule(desired, desiredPort)
		logger.With(
			"destination", *rule.Destination,
			"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
			"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
		).Debug("Deleting load balancer egress security rule")
		egressRules = append(egressRules, rule)
	}

	return egressRules
}

// TODO(apryde): UDP support.
func makeEgressSecurityRule(cidrBlock string, port int) core.EgressSecurityRule {
	return core.EgressSecurityRule{
		Destination: &cidrBlock,
		Protocol:    common.String(fmt.Sprintf("%d", ProtocolTCP)),
		TcpOptions: &core.TcpOptions{
			DestinationPortRange: &core.PortRange{
				Min: &port,
				Max: &port,
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
				Min: &port,
				Max: &port,
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
