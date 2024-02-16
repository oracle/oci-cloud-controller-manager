// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"net/http"
	"reflect"
	"sort"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	k8sports "k8s.io/kubernetes/pkg/cluster/ports"
)

var (
	addNetworkSecurityGroupSecurityRules = map[string]*core.AddNetworkSecurityGroupSecurityRulesResponse{
		"id": {
			RawResponse: &http.Response{Status: "200"},
			AddedNetworkSecurityGroupSecurityRules: core.AddedNetworkSecurityGroupSecurityRules{
				SecurityRules: []core.SecurityRule{
					{
						Id: common.String("1"),
					},
				},
			},
			OpcRequestId: common.String("opc-request-id"),
		},
	}

	removeNetworkSecurityGroupSecurityRules = map[string]*core.RemoveNetworkSecurityGroupSecurityRulesResponse{
		"id": {
			RawResponse:  &http.Response{Status: "200"},
			OpcRequestId: common.String("opc-request-id"),
		},
	}

	listNetworkSecurityGroupSecurityRules = map[string][]core.SecurityRule{
		"id": {
			{
				Direction:       "",
				Protocol:        nil,
				Description:     nil,
				Destination:     nil,
				DestinationType: "",
				IcmpOptions:     nil,
				Id:              common.String("1"),
				IsStateless:     nil,
				IsValid:         nil,
				Source:          nil,
				SourceType:      "",
				TcpOptions:      nil,
				TimeCreated:     nil,
				UdpOptions:      nil,
			},
			{Id: common.String("2")},
			{Id: common.String("3")},
			{Id: common.String("4")},
		},
	}
)

func (c *MockVirtualNetworkClient) AddNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.AddNetworkSecurityGroupSecurityRulesDetails) (*core.AddNetworkSecurityGroupSecurityRulesResponse, error) {
	return addNetworkSecurityGroupSecurityRules[id], nil
}

func (c *MockVirtualNetworkClient) RemoveNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.RemoveNetworkSecurityGroupSecurityRulesDetails) (*core.RemoveNetworkSecurityGroupSecurityRulesResponse, error) {
	return removeNetworkSecurityGroupSecurityRules[id], nil
}

func (c *MockVirtualNetworkClient) ListNetworkSecurityGroupSecurityRules(ctx context.Context, id string, direction core.ListNetworkSecurityGroupSecurityRulesDirectionEnum) ([]core.SecurityRule, error) {
	return listNetworkSecurityGroupSecurityRules[id], nil
}

func (c *MockVirtualNetworkClient) UpdateNetworkSecurityGroupSecurityRules(ctx context.Context, id string, details core.UpdateNetworkSecurityGroupSecurityRulesDetails) (*core.UpdateNetworkSecurityGroupSecurityRulesResponse, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) CreateNetworkSecurityGroup(ctx context.Context, compartmentId, vcnId, displayName, serviceUid string) (*core.NetworkSecurityGroup, error) {
	nsg := core.NetworkSecurityGroup{
		CompartmentId:  common.String(compartmentId),
		Id:             common.String("id"),
		LifecycleState: "ACTIVE",
		VcnId:          common.String(vcnId),
		DisplayName:    common.String(displayName),
		FreeformTags: map[string]string{
			"CreatedBy":  "CCM",
			"ServiceUid": serviceUid,
		},
	}
	return &nsg, nil
}

func (c *MockVirtualNetworkClient) UpdateNetworkSecurityGroup(ctx context.Context, id, etag string, freeformTags map[string]string) (*core.NetworkSecurityGroup, error) {
	return nil, nil
}

func (c *MockVirtualNetworkClient) GetNetworkSecurityGroup(ctx context.Context, id string) (*core.NetworkSecurityGroup, *string, error) {
	nsg := core.NetworkSecurityGroup{
		Id:             &id,
		LifecycleState: "ACTIVE",
	}
	return &nsg, common.String("etag"), nil
}

func (c *MockVirtualNetworkClient) ListNetworkSecurityGroups(ctx context.Context, displayName, compartmentId, vcnId string) ([]core.NetworkSecurityGroup, error) {
	return []core.NetworkSecurityGroup{}, nil
}

func (c *MockVirtualNetworkClient) DeleteNetworkSecurityGroup(ctx context.Context, id, etag string) (*string, error) {
	return nil, nil
}

func TestGenerateLbNsgIngressRules(t *testing.T) {
	testCases := []struct {
		name        string
		sourceCIDRs []string
		port        map[string]portSpec
		lbId        string
		expected    []core.SecurityRule
	}{
		{
			name: "source cidr's",
			sourceCIDRs: []string{
				"0.0.0.0/0",
				"1.1.1.1/1",
			},
			port: map[string]portSpec{"test": {
				ListenerPort:      80,
				BackendPort:       0,
				HealthCheckerPort: 0,
			},
			},
			lbId: "lbocid",
			expected: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
			},
		},
		{
			name: "source cidr's multiple ports",
			sourceCIDRs: []string{
				"0.0.0.0/0",
				"1.1.1.1/1",
			},
			port: map[string]portSpec{"TCP-80": {
				ListenerPort:      80,
				BackendPort:       0,
				HealthCheckerPort: 0,
			},
				"TCP-443": {
					ListenerPort:      443,
					BackendPort:       0,
					HealthCheckerPort: 0,
				},
			},
			lbId: "lbocid",
			expected: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 443, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 443, core.SecurityRuleSourceTypeCidrBlock),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules := generateNsgLoadBalancerIngressRules(zap.S(), tc.sourceCIDRs, tc.port, tc.lbId)
			sort.Slice(rules, func(i, j int) bool {
				return *rules[i].TcpOptions.DestinationPortRange.Min < *rules[j].TcpOptions.DestinationPortRange.Min
			})
			if !reflect.DeepEqual(rules, tc.expected) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expected, rules)
			}
		})
	}
}

func TestGenerateLbNsgEgressRules(t *testing.T) {
	testCases := []struct {
		name          string
		backendNsgIds []string
		desiredPort   map[string]portSpec
		lbId          string
		expected      []core.SecurityRule
	}{
		{
			name: "egress on multiple backend nsgIds",
			desiredPort: map[string]portSpec{"TCP-80": {
				ListenerPort:      80,
				BackendPort:       30001,
				HealthCheckerPort: 10257,
			},
			},
			lbId:          "lbocid",
			backendNsgIds: []string{"nsgId1", "nsgId2", "nsgId3"},
			expected: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "nsgId1", "lbocid", 10257, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "nsgId2", "lbocid", 10257, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "nsgId3", "lbocid", 10257, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "nsgId1", "lbocid", 30001, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "nsgId2", "lbocid", 30001, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "nsgId3", "lbocid", 30001, core.SecurityRuleSourceTypeNetworkSecurityGroup),
			},
		},
		{
			name: "new egress single backend nsg ocid",
			desiredPort: map[string]portSpec{"TCP-80": {
				ListenerPort:      0,
				BackendPort:       30001,
				HealthCheckerPort: 10257,
			},
			},
			backendNsgIds: []string{"backendNSGocid"},
			lbId:          "lbocid",
			expected: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "backendNSGocid", "lbocid", 10257, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "backendNSGocid", "lbocid", 30001, core.SecurityRuleSourceTypeNetworkSecurityGroup),
			},
		},
		{
			name: "new egress single backend nsg ocid",
			desiredPort: map[string]portSpec{"TCP-80": {
				ListenerPort:      0,
				BackendPort:       30001,
				HealthCheckerPort: 10257,
			},
				"TCP-443": {
					ListenerPort:      0,
					BackendPort:       30002,
					HealthCheckerPort: 10257,
				},
			},
			backendNsgIds: []string{"backendNSGocid"},
			lbId:          "lbocid",
			expected: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "backendNSGocid", "lbocid", 10257, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "backendNSGocid", "lbocid", 30001, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionEgress, "backendNSGocid", "lbocid", 30002, core.SecurityRuleSourceTypeNetworkSecurityGroup),
			},
		},
		{
			name: "empty backend nsg ocid",
			desiredPort: map[string]portSpec{"TCP-80": {
				ListenerPort:      0,
				BackendPort:       30001,
				HealthCheckerPort: 10257,
			},
			},
			backendNsgIds: []string{},
			expected:      []core.SecurityRule{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules := generateNsgLoadBalancerEgressRules(zap.S(), tc.desiredPort, tc.backendNsgIds, tc.lbId)
			sort.Slice(rules, func(i, j int) bool {
				return *rules[i].TcpOptions.DestinationPortRange.Min < *rules[j].TcpOptions.DestinationPortRange.Min
			})
			if !reflect.DeepEqual(rules, tc.expected) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expected, rules)
			}
		})
	}
}

func TestGenerateBackendNsgIngressRules(t *testing.T) {
	testCases := []struct {
		name             string
		frontendNsgId    string
		desiredPorts     map[string]portSpec
		sourceCIDRs      []string
		isPreserveSource bool
		lbId             string
		expected         []core.SecurityRule
	}{
		{
			name:          "ingress backend rules",
			frontendNsgId: "frontendnsgocid",
			desiredPorts: map[string]portSpec{"TCP-80": {
				BackendPort:       80,
				HealthCheckerPort: k8sports.ProxyHealthzPort,
			},
			},
			isPreserveSource: false,
			sourceCIDRs:      []string{"0.0.0.0/0"},
			lbId:             "lbocid",
			expected: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "frontendnsgocid", "lbocid", 80, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "frontendnsgocid", "lbocid", k8sports.ProxyHealthzPort, core.SecurityRuleSourceTypeNetworkSecurityGroup),
			},
		},
		{
			name:          "ingress backend rules",
			frontendNsgId: "frontendnsgocid",
			desiredPorts: map[string]portSpec{"TCP-80": {
				BackendPort:       80,
				HealthCheckerPort: k8sports.ProxyHealthzPort,
			},
				"TCP-443": {
					BackendPort:       443,
					HealthCheckerPort: k8sports.ProxyHealthzPort,
				},
			},
			isPreserveSource: false,
			sourceCIDRs:      []string{"0.0.0.0/0"},
			lbId:             "lbocid",
			expected: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "frontendnsgocid", "lbocid", 80, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "frontendnsgocid", "lbocid", 443, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "frontendnsgocid", "lbocid", k8sports.ProxyHealthzPort, core.SecurityRuleSourceTypeNetworkSecurityGroup),
			},
		},
		{
			name: "ingress backend rules with isPreserveSourceIP set to true",
			desiredPorts: map[string]portSpec{"TCP-80": {
				BackendPort:       3000,
				HealthCheckerPort: k8sports.ProxyHealthzPort,
			},
			},
			isPreserveSource: true,
			sourceCIDRs:      []string{"0.0.0.0/0", "1.1.1.1/1"},
			frontendNsgId:    "frontendnsgId",
			lbId:             "lbocid",
			expected: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 3000, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 3000, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "frontendnsgId", "lbocid", 3000, core.SecurityRuleSourceTypeNetworkSecurityGroup),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "frontendnsgId", "lbocid", k8sports.ProxyHealthzPort, core.SecurityRuleSourceTypeNetworkSecurityGroup),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules := generateNsgBackendIngressRules(zap.S(), tc.desiredPorts, tc.sourceCIDRs, tc.isPreserveSource, tc.frontendNsgId, tc.lbId)
			sort.Slice(rules, func(i, j int) bool {
				return *rules[i].TcpOptions.DestinationPortRange.Min < *rules[j].TcpOptions.DestinationPortRange.Min
			})
			if !reflect.DeepEqual(rules, tc.expected) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expected, rules)
			}
		})
	}
}

func TestBatchProcessingRules(t *testing.T) {
	testCases := []struct {
		name                     string
		existingSecurityRules    []core.SecurityRule
		existingRuleIds          []string
		expectedRulesInBatches   [][]core.SecurityRule
		expectedRuleIdsInBatches [][]string
	}{
		{
			name: "Security Rules less than 25",
			existingSecurityRules: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "2.2.2.2/2", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
			},
			existingRuleIds: []string{"1", "2", "3", "4"},

			expectedRulesInBatches: [][]core.SecurityRule{
				{
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "2.2.2.2/2", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				},
			},
			expectedRuleIdsInBatches: [][]string{
				{"1", "2", "3", "4"},
			},
		},
		{
			name: "Security Rules more than 25",
			existingSecurityRules: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
			},
			existingRuleIds: []string{
				"1", "2", "3", "4", "5",
				"10", "20", "30", "40", "50",
				"100", "200", "300", "400", "500",
				"1001", "2001", "3001", "4001", "5001",
				"1002", "2002", "3002", "4002", "5002",
				"1003", "2003", "3003", "4003", "5003",
			},
			expectedRulesInBatches: [][]core.SecurityRule{
				{makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 81, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				},
				{
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
					makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 8080, core.SecurityRuleSourceTypeCidrBlock),
				},
			},
			expectedRuleIdsInBatches: [][]string{
				[]string{"1", "2", "3", "4", "5", "10", "20", "30", "40", "50", "100", "200", "300", "400", "500",
					"1001", "2001", "3001", "4001", "5001", "1002", "2002", "3002", "4002", "5002"},
				[]string{"1003", "2003", "3003", "4003", "5003"},
			},
		},
		{
			name:                     "Security Rules Empty",
			existingSecurityRules:    []core.SecurityRule{},
			existingRuleIds:          []string{},
			expectedRulesInBatches:   [][]core.SecurityRule{{}},
			expectedRuleIdsInBatches: [][]string{{}},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			rulesInBatches := splitRulesIntoBatches(tc.existingSecurityRules)
			if !reflect.DeepEqual(rulesInBatches, tc.expectedRulesInBatches) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expectedRulesInBatches, rulesInBatches)
			}
			ruleIdsInBatches := splitRuleIdsIntoBatches(tc.existingRuleIds)
			if !reflect.DeepEqual(ruleIdsInBatches, tc.expectedRuleIdsInBatches) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expectedRuleIdsInBatches, ruleIdsInBatches)
			}
		})
	}
}

func TestCompareExistingGeneratedRules(t *testing.T) {
	testCases := []struct {
		name           string
		existingRules  []core.SecurityRule
		generatedRules []core.SecurityRule
		expectedAdd    []core.SecurityRule
		expectedDelete []string
	}{
		{
			name: "case remove rules",
			existingRules: []core.SecurityRule{
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("1"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsgId"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("2"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsgId"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(10256),
							Min: common.Int(10256),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("4"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("0.0.0.0/0"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30001),
							Min: common.Int(30001),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("5"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("0.0.0.0/0"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(10257),
							Min: common.Int(10257),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
			},
			generatedRules: generateNsgBackendIngressRules(zap.S(), map[string]portSpec{
				"TCP-443": {
					ListenerPort:      80,
					BackendPort:       30000,
					HealthCheckerPort: 10256,
				},
			}, []string{}, false, "frontendNsgId", "lbocid"),
			expectedDelete: []string{"4", "5"},
			expectedAdd:    []core.SecurityRule{},
		},
		{
			name:          "case add rules",
			existingRules: []core.SecurityRule{},
			generatedRules: generateNsgBackendIngressRules(zap.S(), map[string]portSpec{
				"TCP-443": {
					ListenerPort:      80,
					BackendPort:       30000,
					HealthCheckerPort: 10256,
				},
			}, []string{}, false, "frontendNsgId", "lbocid"),
			expectedDelete: []string{},
			expectedAdd: []core.SecurityRule{
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsgId"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					UdpOptions: nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsgId"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(10256),
							Min: common.Int(10256),
						},
						SourcePortRange: nil,
					},
					UdpOptions: nil,
				},
			},
		},
		{
			name: "base case - ispreservesource",
			existingRules: []core.SecurityRule{
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("1"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsgId"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("2"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsgId"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(10256),
							Min: common.Int(10256),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("4"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("0.0.0.0/0"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("5"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("0.0.0.0/0"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(10256),
							Min: common.Int(10256),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
			},
			generatedRules: generateNsgBackendIngressRules(zap.S(), map[string]portSpec{
				"TCP-443": {
					ListenerPort:      80,
					BackendPort:       30000,
					HealthCheckerPort: 10256,
				},
			}, []string{"0.0.0.0/0"}, true, "frontendNsgId", "lbocid"),
			expectedDelete: []string{"5"},
			expectedAdd:    []core.SecurityRule{},
		},
		{
			name: "egress rules test",
			existingRules: []core.SecurityRule{
				{
					Direction:       core.SecurityRuleDirectionEgress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     common.String("frontendNsgId"),
					DestinationType: core.SecurityRuleDestinationTypeNetworkSecurityGroup,
					IcmpOptions:     nil,
					Id:              common.String("1"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          nil,
					SourceType:      "",
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionEgress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Source:          nil,
					SourceType:      "",
					IcmpOptions:     nil,
					Id:              common.String("2"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Destination:     common.String("frontendNsgId"),
					DestinationType: core.SecurityRuleDestinationTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(10256),
							Min: common.Int(10256),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionEgress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Source:          nil,
					SourceType:      "",
					IcmpOptions:     nil,
					Id:              common.String("5"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Destination:     common.String("frontendNsgId"),
					DestinationType: core.SecurityRuleDestinationTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30001),
							Min: common.Int(30001),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
			},
			generatedRules: generateNsgLoadBalancerEgressRules(zap.S(), map[string]portSpec{
				"TCP-443": {
					ListenerPort:      80,
					BackendPort:       30000,
					HealthCheckerPort: 10256,
				},
			}, []string{"frontendNsgId"}, "lbocid"),
			expectedDelete: []string{"5"},
			expectedAdd:    []core.SecurityRule{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			addRules, removeRules, _ := reconcileSecurityRules(zap.S(), tc.generatedRules, tc.existingRules)
			if !reflect.DeepEqual(addRules, tc.expectedAdd) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expectedAdd, addRules)
			}
			if !reflect.DeepEqual(removeRules, tc.expectedDelete) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expectedDelete, removeRules)
			}
		})
	}
}

func TestFilterRuleIds(t *testing.T) {
	testCases := []struct {
		name          string
		existingRules []core.SecurityRule
		serviceUid    string
		expected      []string
	}{
		{
			name:       "filter service uid",
			serviceUid: "lbocid",
			existingRules: []core.SecurityRule{
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("1"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsgId"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("2"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsgId"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(10256),
							Min: common.Int(10256),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid_random"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("4"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("0.0.0.0/0"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30001),
							Min: common.Int(30001),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid_random"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("5"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("0.0.0.0/0"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(10257),
							Min: common.Int(10257),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
			},
			expected: []string{
				"1",
				"2",
			},
		},
		{
			name:       "similar rules but for other services",
			serviceUid: "lbocid",
			existingRules: []core.SecurityRule{
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("1"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsg"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("2"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsg"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(10256),
							Min: common.Int(10256),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("3"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("0.0.0.0/0"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("4"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("1.1.1.1/1"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("randomrule"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("5"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("2.2.2.2/2"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
			},
			expected: []string{
				"1",
				"2",
				"3",
				"4",
			},
		},
		{
			name:       "description nil",
			serviceUid: "lbocid",
			existingRules: []core.SecurityRule{
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("1"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsg"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("2"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("frontendNsg"),
					SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(10256),
							Min: common.Int(10256),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("3"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("0.0.0.0/0"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("4"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("1.1.1.1/1"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     common.String("lbocid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("5"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("2.2.2.2/2"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
				{
					Direction:       core.SecurityRuleDirectionIngress,
					Protocol:        common.String(fmt.Sprintf("%d", ProtocolTCP)),
					Description:     nil,
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("6"),
					IsStateless:     common.Bool(false),
					IsValid:         nil,
					Source:          common.String("2.2.2.2/2"),
					SourceType:      core.SecurityRuleSourceTypeCidrBlock,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30001),
							Min: common.Int(30001),
						},
						SourcePortRange: nil,
					},
					TimeCreated: nil,
					UdpOptions:  nil,
				},
			},
			expected: []string{
				"1",
				"2",
				"3",
				"4",
				"5",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules := filterSecurityRulesIdsForService(tc.existingRules, tc.serviceUid)
			sort.Slice(rules, func(i, j int) bool {
				return rules[i] < rules[j]
			})
			if !reflect.DeepEqual(rules, tc.expected) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expected, rules)
			}
		})
	}
}

func TestFilterRules(t *testing.T) {
	testCases := []struct {
		name          string
		rules         []core.SecurityRule
		expectedRules []core.SecurityRule
		serviceUid    string
	}{
		{
			name: "base case of filtering required rules",
			rules: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "2.2.2.2/2", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "2.2.2.2/2", "randomUID", 80, core.SecurityRuleSourceTypeCidrBlock),
			},
			expectedRules: []core.SecurityRule{
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "0.0.0.0/0", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "1.1.1.1/1", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
				makeNsgSecurityRule(core.SecurityRuleDirectionIngress, "2.2.2.2/2", "lbocid", 80, core.SecurityRuleSourceTypeCidrBlock),
			},
			serviceUid: "lbocid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules := filterSecurityRulesForService(tc.rules, tc.serviceUid)
			if !reflect.DeepEqual(rules, tc.expectedRules) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expectedRules, rules)
			}
		})
	}
}

func TestGetNsgHelper(t *testing.T) {
	cp := CloudProvider{
		client: MockOCIClient{},
	}
	ctx := context.Background()
	testCases := []struct {
		nsg  *core.NetworkSecurityGroup
		id   string
		err  error
		name string
	}{
		{
			name: "getNSG",
			nsg: &core.NetworkSecurityGroup{
				Id:             common.String("id"),
				LifecycleState: "ACTIVE",
			},
			err: nil,
			id:  "id",
		},
		{
			name: "getNSG err",
			nsg:  nil,
			err:  errors.New("invalid; empty nsg id provided"),
			id:   "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nsg, err := cp.getNsg(ctx, tc.id)
			if !reflect.DeepEqual(nsg, tc.nsg) {
				t.Errorf("expected nsg\n%+v\nbut got\n%+v", tc.nsg, nsg)
			}
			if err != nil {
				if !reflect.DeepEqual(err.Error(), tc.err.Error()) {
					t.Errorf("expected err\n%+v\nbut got\n%+v", tc.err, err)
				}
			}
		})
	}
}

func TestListNsgRulesHelper(t *testing.T) {
	cp := CloudProvider{
		client: MockOCIClient{},
	}
	ctx := context.Background()
	testCases := []struct {
		securityRules []core.SecurityRule
		id            string
		direction     core.ListNetworkSecurityGroupSecurityRulesDirectionEnum
		err           error
		name          string
	}{
		{
			name: "listNsgRules",
			securityRules: []core.SecurityRule{
				{
					Direction:       "",
					Protocol:        nil,
					Description:     nil,
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					Id:              common.String("1"),
					IsStateless:     nil,
					IsValid:         nil,
					Source:          nil,
					SourceType:      "",
					TcpOptions:      nil,
					TimeCreated:     nil,
					UdpOptions:      nil,
				},
				{Id: common.String("2")},
				{Id: common.String("3")},
				{Id: common.String("4")},
			},
			err: nil,
			id:  "id",
		},
		{
			name:          "listNsgRules err",
			securityRules: nil,
			err:           errors.New("invalid; empty nsg id provided"),
			id:            "",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			securityRuleList, err := cp.listNsgRules(ctx, tc.id, tc.direction)
			if !reflect.DeepEqual(securityRuleList, tc.securityRules) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.securityRules, securityRuleList)
			}
			if err != nil {
				if !reflect.DeepEqual(err.Error(), tc.err.Error()) {
					t.Errorf("expected err\n%+v\nbut got\n%+v", tc.err, err)
				}
			}
		})
	}
}

func TestAddNsgRulesHelper(t *testing.T) {
	cp := CloudProvider{
		client: MockOCIClient{},
		logger: zap.S(),
	}
	ctx := context.Background()
	testCases := []struct {
		securityRules []core.SecurityRule
		id            *string
		err           error
		name          string
		response      *core.AddNetworkSecurityGroupSecurityRulesResponse
	}{
		{
			name: "add rules",
			securityRules: []core.SecurityRule{
				{
					Id: common.String("1"),
				},
			},
			err: nil,
			id:  common.String("id"),
			response: &core.AddNetworkSecurityGroupSecurityRulesResponse{
				RawResponse: &http.Response{Status: "200"},
				AddedNetworkSecurityGroupSecurityRules: core.AddedNetworkSecurityGroupSecurityRules{
					SecurityRules: []core.SecurityRule{
						{
							Id: common.String("1"),
						},
					},
				},
				OpcRequestId: common.String("opc-request-id"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := cp.addNetworkSecurityGroupSecurityRules(ctx, tc.id, tc.securityRules)
			if !reflect.DeepEqual(response, tc.response) {
				t.Errorf("expected response\n%+v\nbut got\n%+v", tc.response, response)
			}
			if err != nil {
				if !reflect.DeepEqual(err.Error(), tc.err.Error()) {
					t.Errorf("expected err\n%+v\nbut got\n%+v", tc.err, err)
				}
			}
		})
	}
}

func TestRemoveNsgRulesHelper(t *testing.T) {
	cp := CloudProvider{
		client: MockOCIClient{},
		logger: zap.S(),
	}
	ctx := context.Background()
	testCases := []struct {
		ruleIds  []string
		id       *string
		err      error
		name     string
		response *core.RemoveNetworkSecurityGroupSecurityRulesResponse
	}{
		{
			name:    "remove rules",
			ruleIds: []string{},
			err:     nil,
			id:      common.String("id"),
			response: &core.RemoveNetworkSecurityGroupSecurityRulesResponse{
				RawResponse:  &http.Response{Status: "200"},
				OpcRequestId: common.String("opc-request-id"),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := cp.removeNetworkSecurityGroupSecurityRules(ctx, tc.id, tc.ruleIds)
			if !reflect.DeepEqual(response, tc.response) {
				t.Errorf("expected response\n%+v\nbut got\n%+v", tc.response, response)
			}
			if err != nil {
				if !reflect.DeepEqual(err.Error(), tc.err.Error()) {
					t.Errorf("expected err\n%+v\nbut got\n%+v", tc.err, err)
				}
			}
		})
	}
}

func TestSecurityRulesToAddSecurityRulesHelper(t *testing.T) {
	testCases := []struct {
		rules         []core.SecurityRule
		expectedRules []core.AddSecurityRuleDetails
		name          string
	}{
		{
			name: "base case",
			rules: []core.SecurityRule{{
				Direction:       "",
				Protocol:        common.String("TCP"),
				Description:     common.String("service-uid"),
				Destination:     nil,
				DestinationType: "",
				IcmpOptions:     nil,
				IsStateless:     common.Bool(false),
				Source:          common.String("nsgocid"),
				SourceType:      core.SecurityRuleSourceTypeNetworkSecurityGroup,
				TcpOptions: &core.TcpOptions{
					DestinationPortRange: &core.PortRange{
						Max: common.Int(30000),
						Min: common.Int(30000),
					},
				},
				UdpOptions: nil,
			}},
			expectedRules: []core.AddSecurityRuleDetails{
				{
					Direction:       "",
					Protocol:        common.String("TCP"),
					Description:     common.String("service-uid"),
					Destination:     nil,
					DestinationType: "",
					IcmpOptions:     nil,
					IsStateless:     common.Bool(false),
					Source:          common.String("nsgocid"),
					SourceType:      core.AddSecurityRuleDetailsSourceTypeNetworkSecurityGroup,
					TcpOptions: &core.TcpOptions{
						DestinationPortRange: &core.PortRange{
							Max: common.Int(30000),
							Min: common.Int(30000),
						},
					},
					UdpOptions: nil,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rules := securityRuleToAddSecurityRuleDetails(tc.rules)
			if !reflect.DeepEqual(rules, tc.expectedRules) {
				t.Errorf("expected rules\n%+v\nbut got\n%+v", tc.expectedRules, rules)
			}
		})
	}
}
