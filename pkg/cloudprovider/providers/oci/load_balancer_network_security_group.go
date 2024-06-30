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
	"strings"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
)

const (
	// RuleManagementModeNsg denotes the management of loadbalancer ingress via NSG
	RuleManagementModeNsg = "NSG"
	// RuleManagementModeSlAll denotes the management of security list rules for load
	// balancer ingress/egress, health checkers, and worker ingress/egress.
	RuleManagementModeSlAll = "SL-All"
	// RuleManagementModeSlFrontend denotes the management of security list rules for load
	// balancer ingress only.
	RuleManagementModeSlFrontend = "SL-Frontend"
)

const (
	batchSize = 25
)

type securityRuleComponents struct {
	frontendNsgOcid  string
	backendNsgOcids  []string
	ports            map[string]portSpec
	sourceCIDRs      []string
	isPreserveSource bool
	serviceUid       string
	lbSubnets        []*core.Subnet
	backendSubnets   []*core.Subnet
	actualPorts      *portSpec
	desiredPorts     portSpec
	ipFamilies       []string
}

// generateNsgBackendIngressRules is a helper method to generate the ingress rules for the backend NSG
func generateNsgBackendIngressRules(
	logger *zap.SugaredLogger,
	ports map[string]portSpec,
	sourceCIDRs []string,
	isPreserveSource bool,
	frontendNsgId string,
	serviceUid string,
) []core.SecurityRule {

	// Additional sourceCIDR rule for NLB only, for source IP preservation
	ingressRules := []core.SecurityRule{}
	if isPreserveSource {
		for _, port := range ports {
			if port.BackendPort != 0 {
				for _, sourceCIDR := range sourceCIDRs {
					nlbRule := makeNsgSecurityRule(core.SecurityRuleDirectionIngress, sourceCIDR, serviceUid, port.BackendPort, core.SecurityRuleSourceTypeCidrBlock)
					logger.With(
						"source", *nlbRule.Source,
						"destinationPortRangeMin", *nlbRule.TcpOptions.DestinationPortRange.Min,
						"destinationPortRangeMax", *nlbRule.TcpOptions.DestinationPortRange.Max,
					).Debug("Adding node port ingress security rule on backend nsg(s)")
					ingressRules = append(ingressRules, nlbRule)
				}
			}
		}
	}

	healthCheckPortFound := false
	for _, port := range ports {
		if port.BackendPort != 0 { // Can happen when there are no backends.
			rule := makeNsgSecurityRule(core.SecurityRuleDirectionIngress, frontendNsgId, serviceUid, port.BackendPort, core.SecurityRuleSourceTypeNetworkSecurityGroup)
			logger.With(
				"source", *rule.Source,
				"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
				"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
			).Debug("Adding node port ingress security rule on backend nsg(s)")
			ingressRules = append(ingressRules, rule)

		}
		if !healthCheckPortFound && port.HealthCheckerPort != 0 {
			healthCheckPortFound = true
			rule := makeNsgSecurityRule(core.SecurityRuleDirectionIngress, frontendNsgId, serviceUid, port.HealthCheckerPort, core.SecurityRuleSourceTypeNetworkSecurityGroup)
			logger.With(
				"source", *rule.Source,
				"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
				"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
			).Debug("Adding healthcheck node port ingress security rule on backend nsg(s)")
			ingressRules = append(ingressRules, rule)
		}
	}
	return ingressRules
}

// generateNsgLoadBalancerIngressRules is a helper method to generate the ingress rules for the frontend NSG
func generateNsgLoadBalancerIngressRules(
	logger *zap.SugaredLogger,
	sourceCIDRs []string,
	ports map[string]portSpec,
	serviceUid string,
) []core.SecurityRule {
	var ingressRules []core.SecurityRule

	if len(sourceCIDRs) == 0 {
		// actual is the same as desired so there is nothing to do
		return ingressRules
	}

	for _, port := range ports {
		if port.ListenerPort != 0 {
			for _, cidr := range sourceCIDRs {
				rule := makeNsgSecurityRule(core.SecurityRuleDirectionIngress, cidr, serviceUid, port.ListenerPort, core.SecurityRuleSourceTypeCidrBlock)
				logger.With(
					"source", *rule.Source,
					"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
					"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
				).Debug("Adding load balancer ingress security rule for frontend nsg")
				ingressRules = append(ingressRules, rule)
			}
		}
	}

	return ingressRules
}

// generateNsgLoadBalancerEgressRules is a helper method to generate the egress rules for the frontend NSG
func generateNsgLoadBalancerEgressRules(logger *zap.SugaredLogger, ports map[string]portSpec, backendNsgIds []string, serviceUid string) []core.SecurityRule {
	egressRules := []core.SecurityRule{}
	if len(backendNsgIds) == 0 {
		// actual is the same as desired so there is nothing to do
		return egressRules
	}

	rule := core.SecurityRule{}
	if len(backendNsgIds) != 0 {
		healthCheckPortFound := false
		for _, port := range ports {
			if port.BackendPort != 0 {
				for _, backendNsgId := range backendNsgIds {
					rule = makeNsgSecurityRule(core.SecurityRuleDirectionEgress, backendNsgId, serviceUid, port.BackendPort, core.SecurityRuleSourceTypeNetworkSecurityGroup)
					egressRules = append(egressRules, rule)
					logger.With(
						"destination", *rule.Destination,
						"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
						"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
					).Debug("Adding load balancer egress security rule with backend port on frontend nsg")
				}
			}
			if !healthCheckPortFound && port.HealthCheckerPort != 0 {
				healthCheckPortFound = true
				for _, backendNsgId := range backendNsgIds {
					rule = makeNsgSecurityRule(core.SecurityRuleDirectionEgress, backendNsgId, serviceUid, port.HealthCheckerPort, core.SecurityRuleSourceTypeNetworkSecurityGroup)
					egressRules = append(egressRules, rule)
					logger.With(
						"destination", *rule.Destination,
						"destinationPortRangeMin", *rule.TcpOptions.DestinationPortRange.Min,
						"destinationPortRangeMax", *rule.TcpOptions.DestinationPortRange.Max,
					).Debug("Adding load balancer egress security rule with healthcheck port on frontend nsg")
				}
			}

		}
	}
	return egressRules
}

// makeNsgSecurityRule is a helper method to build the Security Rule using direction, source and sourceType (cidr/nsg)
func makeNsgSecurityRule(direction core.SecurityRuleDirectionEnum, source string, serviceUid string, port int, sourceType core.SecurityRuleSourceTypeEnum) core.SecurityRule {
	rule := core.SecurityRule{
		Description: common.String(serviceUid),
		Protocol:    common.String(fmt.Sprintf("%d", ProtocolTCP)),
		TcpOptions: &core.TcpOptions{
			DestinationPortRange: &core.PortRange{
				Min: &port,
				Max: &port,
			},
		},
		IsStateless: common.Bool(false),
	}
	if direction == core.SecurityRuleDirectionEgress {
		rule.Direction = core.SecurityRuleDirectionEgress
		rule.Destination = common.String(source)
		rule.DestinationType = core.SecurityRuleDestinationTypeEnum(sourceType)
	} else {
		rule.Source = common.String(source)
		rule.SourceType = sourceType
		rule.Direction = core.SecurityRuleDirectionIngress
	}
	return rule
}

// getNsg implements the client method to get nsg
func (s *CloudProvider) getNsg(ctx context.Context, id string) (*core.NetworkSecurityGroup, error) {
	if id == "" {
		return nil, errors.New("invalid; empty nsg id provided") // should never happen
	}
	response, _, err := s.client.Networking(nil).GetNetworkSecurityGroup(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get nsg with id %s", id)
	}
	return response, nil
}

// listNsgRules implements the client method to list nsg rules based on direction
func (s *CloudProvider) listNsgRules(ctx context.Context, id string, direction core.ListNetworkSecurityGroupSecurityRulesDirectionEnum) ([]core.SecurityRule, error) {
	if id == "" {
		return nil, errors.New("invalid; empty nsg id provided") // should never happen
	}

	response, err := s.client.Networking(nil).ListNetworkSecurityGroupSecurityRules(ctx, id, direction)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list Security Rules for nsg: %s", id)
	}
	return response, nil
}

// addNetworkSecurityGroupSecurityRules implements the client method to add nsg rules given the NSG id and security rules
func (s *CloudProvider) addNetworkSecurityGroupSecurityRules(ctx context.Context, nsgId *string, rules []core.SecurityRule) (*core.AddNetworkSecurityGroupSecurityRulesResponse, error) {
	rulesInBatches := splitRulesIntoBatches(rules)
	var response *core.AddNetworkSecurityGroupSecurityRulesResponse
	var err error
	for i, _ := range rulesInBatches {
		response, err = s.client.Networking(nil).AddNetworkSecurityGroupSecurityRules(ctx,
			*nsgId,
			core.AddNetworkSecurityGroupSecurityRulesDetails{SecurityRules: securityRuleToAddSecurityRuleDetails(rulesInBatches[i])})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to add security rules for nsg: %s OpcRequestId: %s", *nsgId, pointer.StringDeref(response.OpcRequestId, ""))
		}
		s.logger.Infof("AddNetworkSecurityGroupSecurityRules OpcRequestId %s", pointer.StringDeref(response.OpcRequestId, ""))
	}
	return response, nil
}

// removeNetworkSecurityGroupSecurityRules implements the client method to remove nsg rules given the NSG id and security rule ids
func (s *CloudProvider) removeNetworkSecurityGroupSecurityRules(ctx context.Context, nsgId *string, ids []string) (*core.RemoveNetworkSecurityGroupSecurityRulesResponse, error) {
	rulesInBatches := splitRuleIdsIntoBatches(ids)
	var response *core.RemoveNetworkSecurityGroupSecurityRulesResponse
	var err error
	for i, _ := range rulesInBatches {
		response, err = s.client.Networking(nil).RemoveNetworkSecurityGroupSecurityRules(ctx, *nsgId,
			core.RemoveNetworkSecurityGroupSecurityRulesDetails{SecurityRuleIds: rulesInBatches[i]})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to remove security rules for nsg: %s OpcRequestId: %s", *nsgId, pointer.StringDeref(response.OpcRequestId, ""))
		}
		s.logger.Infof("RemoveNetworkSecurityGroupSecurityRules OpcRequestId %s", pointer.StringDeref(response.OpcRequestId, ""))
	}
	return response, nil
}

// securityRuleToAddSecurityRuleDetails is a helper method for type conversion from SecurityRules to AddSecurityRuleDetails
func securityRuleToAddSecurityRuleDetails(securityRules []core.SecurityRule) []core.AddSecurityRuleDetails {
	addSecurityRuleDetails := make([]core.AddSecurityRuleDetails, 0)

	for _, securityRule := range securityRules {
		addSecurityRuleDetails = append(addSecurityRuleDetails, core.AddSecurityRuleDetails{
			Direction:       core.AddSecurityRuleDetailsDirectionEnum(securityRule.Direction),
			Protocol:        securityRule.Protocol,
			Description:     securityRule.Description,
			Destination:     securityRule.Destination,
			DestinationType: core.AddSecurityRuleDetailsDestinationTypeEnum(securityRule.DestinationType),
			IcmpOptions:     securityRule.IcmpOptions,
			IsStateless:     securityRule.IsStateless,
			Source:          securityRule.Source,
			SourceType:      core.AddSecurityRuleDetailsSourceTypeEnum(securityRule.SourceType),
			TcpOptions:      securityRule.TcpOptions,
			UdpOptions:      securityRule.UdpOptions,
		})
	}
	return addSecurityRuleDetails
}

func splitRulesIntoBatches(rules []core.SecurityRule) [][]core.SecurityRule {
	securityRulesInBatches := make([][]core.SecurityRule, 0, (len(rules)+batchSize-1)/batchSize)

	for batchSize < len(rules) {
		rules, securityRulesInBatches = rules[batchSize:], append(securityRulesInBatches, rules[0:batchSize:batchSize])
	}

	securityRulesInBatches = append(securityRulesInBatches, rules)
	return securityRulesInBatches
}

func splitRuleIdsIntoBatches(rules []string) [][]string {
	securityRulesInBatches := make([][]string, 0, (len(rules)+batchSize-1)/batchSize)
	for batchSize < len(rules) {
		rules, securityRulesInBatches = rules[batchSize:], append(securityRulesInBatches, rules[0:batchSize:batchSize])
	}

	securityRulesInBatches = append(securityRulesInBatches, rules)
	return securityRulesInBatches
}

func (s *CloudProvider) reconcileSecurityGroup(ctx context.Context, lbservice securityRuleComponents) error {
	if len(lbservice.backendNsgOcids) > 0 {
		updateRulesMutex.Lock()
		defer updateRulesMutex.Unlock()
	}

	frontendNsg, err := s.getNsg(ctx, lbservice.frontendNsgOcid)
	if err != nil {
		return err
	}
	logger := s.logger.With("frontendNsgId", *frontendNsg.Id)

	// Frontend NSG Ingress rules
	existingLbIngressSecurityRules, err := s.listNsgRules(ctx, *frontendNsg.Id, core.ListNetworkSecurityGroupSecurityRulesDirectionIngress)
	if err != nil {
		return err
	}
	logger.Info("generating frontend nsg rules")
	generatedLbIngressRules := generateNsgLoadBalancerIngressRules(logger, lbservice.sourceCIDRs, lbservice.ports, lbservice.serviceUid)
	addLbIngressRules, removeLbIngressRules, err := reconcileSecurityRules(logger, generatedLbIngressRules, filterSecurityRulesForService(existingLbIngressSecurityRules, lbservice.serviceUid))

	// Frontend NSG Egress rules
	existingLbEgressSecurityRules, err := s.listNsgRules(ctx, *frontendNsg.Id, core.ListNetworkSecurityGroupSecurityRulesDirectionEgress)
	if err != nil {
		return err
	}
	generatedLbEgressSecurityRules := generateNsgLoadBalancerEgressRules(logger, lbservice.ports, lbservice.backendNsgOcids, lbservice.serviceUid)
	addLbEgressRules, removeLbEgressRules, err := reconcileSecurityRules(logger, generatedLbEgressSecurityRules, filterSecurityRulesForService(existingLbEgressSecurityRules, lbservice.serviceUid))

	addLbRules := append(addLbIngressRules, addLbEgressRules...)
	if len(addLbRules) > 0 {
		logger.Infof("adding frontend nsg rules to nsg %s", *frontendNsg.Id)
		_, err = s.addNetworkSecurityGroupSecurityRules(ctx, frontendNsg.Id, addLbRules)
		if err != nil {
			return err
		}
	}

	removeLbRules := append(removeLbIngressRules, removeLbEgressRules...)
	if len(removeLbRules) > 0 {
		logger.Infof("removing frontend nsg rules for nsg %s", *frontendNsg.Id)
		_, err = s.removeNetworkSecurityGroupSecurityRules(ctx, frontendNsg.Id, removeLbRules)
		if err != nil {
			return err
		}
	}

	for _, nsg := range lbservice.backendNsgOcids {
		_, err := s.getNsg(ctx, nsg)
		if err != nil {
			return err
		}

		logger := s.logger.With("backendNsgId", nsg)
		existingBackendIngressSecurityRules, err := s.listNsgRules(ctx, nsg, core.ListNetworkSecurityGroupSecurityRulesDirectionIngress)
		if err != nil {
			return err
		}
		logger.Info("generating backend nsg rules")
		// Backend NSG Ingress rules
		generatedBackendIngressRules := generateNsgBackendIngressRules(logger, lbservice.ports, lbservice.sourceCIDRs, lbservice.isPreserveSource, lbservice.frontendNsgOcid, lbservice.serviceUid)
		addBackendIngressRules, removeBackendIngressRules, err := reconcileSecurityRules(logger, generatedBackendIngressRules, filterSecurityRulesForService(existingBackendIngressSecurityRules, lbservice.serviceUid))

		if len(addBackendIngressRules) > 0 {
			logger.Infof("adding backend nsg rules to nsg %s", nsg)
			_, err = s.addNetworkSecurityGroupSecurityRules(ctx, &nsg, addBackendIngressRules)
			if err != nil {
				return err
			}
		}

		if len(removeBackendIngressRules) > 0 {
			logger.Infof("removing backend nsg rules from nsg %s", nsg)
			_, err = s.removeNetworkSecurityGroupSecurityRules(ctx, &nsg, removeBackendIngressRules)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *CloudProvider) removeBackendSecurityGroupRules(ctx context.Context, lbservice securityRuleComponents) error {

	for _, backendNsg := range lbservice.backendNsgOcids {
		nsg, err := s.getNsg(ctx, backendNsg)
		if err != nil {
			return err
		}

		logger := s.logger.With("backendNsgId", *nsg.Id)
		existingBackendIngressSecurityRules, err := s.listNsgRules(ctx, *nsg.Id, core.ListNetworkSecurityGroupSecurityRulesDirectionIngress)
		if err != nil {
			return err
		}

		logger.Infof("gather backend nsg rules for service cleanup %s", *nsg.Id)
		// Get existing rules on backend NSG, check if they map to the source and ports to be deleted and return list of rule id's to be deleted
		deleteNsgIngressBackendRules := []string{}
		deleteNsgIngressBackendRules = filterSecurityRulesIdsForService(existingBackendIngressSecurityRules, lbservice.serviceUid)

		if len(deleteNsgIngressBackendRules) > 0 {
			logger.Infof("remove backend nsg rules for service cleanup %s", *nsg.Id)
			_, err = s.removeNetworkSecurityGroupSecurityRules(ctx, nsg.Id, deleteNsgIngressBackendRules)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func reconcileSecurityRules(logger *zap.SugaredLogger, generatedSecurityRules []core.SecurityRule, existingSecurityRules []core.SecurityRule) ([]core.SecurityRule, []string, error) {
	addRules := []core.SecurityRule{}
	removeRules := []string{}

	for _, existingSecurityRule := range existingSecurityRules {
		foundRule := false
		if findSecurityRule(generatedSecurityRules, existingSecurityRule) {
			foundRule = true
		}

		if !foundRule {
			logger.Infof("reconcileSecurityRules: rule (%s) - removing", existingSecurityRule)
			removeRules = append(removeRules, *existingSecurityRule.Id)
		}
	}

	for _, expectedRule := range generatedSecurityRules {
		foundRule := false
		if findSecurityRule(existingSecurityRules, expectedRule) {
			foundRule = true
		}

		if !foundRule {
			logger.Infof("reconcileSecurityRules: rule (%s) - adding", expectedRule)
			addRules = append(addRules, expectedRule)

		}
	}
	return addRules, removeRules, nil
}

func findSecurityRule(rules []core.SecurityRule, rule core.SecurityRule) bool {
	for _, existingRule := range rules {
		if !strings.EqualFold(pointer.StringDeref(existingRule.Description, ""), pointer.StringDeref(rule.Description, "")) {
			continue
		}
		if !strings.EqualFold(pointer.StringDeref(existingRule.Protocol, ""), pointer.StringDeref(rule.Protocol, "")) {
			continue
		}
		if !strings.EqualFold(pointer.StringDeref(existingRule.Source, ""), pointer.StringDeref(rule.Source, "")) {
			continue
		}
		if !reflect.DeepEqual(existingRule.SourceType, rule.SourceType) {
			continue
		}
		if !strings.EqualFold(pointer.StringDeref(existingRule.Destination, ""), pointer.StringDeref(rule.Destination, "")) {
			continue
		}
		if !reflect.DeepEqual(existingRule.DestinationType, rule.DestinationType) {
			continue
		}
		if !reflect.DeepEqual(existingRule.TcpOptions, rule.TcpOptions) {
			continue
		}
		if !strings.EqualFold(string(existingRule.Direction), string(rule.Direction)) {
			continue
		}
		return true
	}
	return false
}

func filterSecurityRulesForService(rules []core.SecurityRule, serviceUid string) []core.SecurityRule {
	rulesPerService := []core.SecurityRule{}
	for _, rule := range rules {
		if rule.Description != nil && *rule.Description == serviceUid {
			rulesPerService = append(rulesPerService, rule)
		}
	}
	return rulesPerService
}

func filterSecurityRulesIdsForService(rules []core.SecurityRule, serviceUid string) []string {
	rulesPerService := []string{}
	for _, rule := range rules {
		if rule.Description != nil && *rule.Description == serviceUid {
			rulesPerService = append(rulesPerService, *rule.Id)
		}
	}
	return rulesPerService
}
