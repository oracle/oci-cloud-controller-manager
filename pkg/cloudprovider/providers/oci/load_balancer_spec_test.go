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
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/pointer"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/common"
)

var (
	backendSecret  = "backendsecret"
	listenerSecret = "listenersecret"
	testNodeString = "ocid1.testNodeTargetID"
)

var (
	tenMbps    = 10
	eightyMbps = 80
)

type mockSSLSecretReader struct {
	returnError bool

	returnMap map[struct {
		namespaceArg string
		nameArg      string
	}]*certificateData
}

func (ssr mockSSLSecretReader) readSSLSecret(ns, name string) (sslSecret *certificateData, err error) {
	if ssr.returnError {
		return nil, errors.New("Oops, something went wrong")
	}
	for key, returnValue := range ssr.returnMap {
		if key.namespaceArg == ns && key.nameArg == name {
			return returnValue, nil
		}
	}
	return nil, nil
}

func TestNewLBSpecSuccess(t *testing.T) {
	enableOkeSystemTags = true
	testCases := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		nodes            []*v1.Node
		virtualPods      []*v1.Pod
		service          *v1.Service
		expected         *LBSpec
		sslConfig        *SSLConfig
		clusterTags      *providercfg.InitialTags
		IpVersions       *IpVersions
	}{
		"defaults": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"defaults-nlb-cluster-policy": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:    "kube-system/testservice/test-uid",
				Type:    "nlb",
				Shape:   "flexible",
				Subnets: []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
						IpVersion:             GenericIpVersion(client.GenericIPv4),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("FIVE_TUPLE"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"defaults-nlb-local-policy": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
					ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeLocal,
				},
			},
			expected: &LBSpec{
				Name:    "kube-system/testservice/test-uid",
				Type:    "nlb",
				Shape:   "flexible",
				Subnets: []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
						IpVersion:             GenericIpVersion(client.GenericIPv4),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(true),
						Policy:           common.String("FIVE_TUPLE"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(true),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"internal with default subnet": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "true",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: true,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"internal with overridden regional subnet1": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "true",
						ServiceAnnotationLoadBalancerSubnet1:  "regional-subnet",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: true,
				Subnets:  []string{"regional-subnet"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"internal with overridden regional subnet2": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "true",
						ServiceAnnotationLoadBalancerSubnet2:  "regional-subnet",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: true,
				Subnets:  []string{"regional-subnet"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"internal with no default subnets provide subnet1 via annotation": {
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "true",
						ServiceAnnotationLoadBalancerSubnet1:  "annotation-one",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: true,
				Subnets:  []string{"annotation-one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"use default subnet in case of no subnet overrides via annotation": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": client.GenericListener{
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": client.GenericBackendSetDetails{
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"no default subnets provide subnet1 via annotation as regional-subnet": {
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "regional-subnet",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"regional-subnet"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": client.GenericListener{
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": client.GenericBackendSetDetails{
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"no default subnets provide subnet2 via annotation": {
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet2: "regional-subnet",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"regional-subnet"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": client.GenericListener{
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": client.GenericBackendSetDetails{
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"override default subnet via subnet1 annotation as regional subnet": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "regional-subnet",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"regional-subnet"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": client.GenericListener{
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": client.GenericBackendSetDetails{
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"override default subnet via subnet2 annotation as regional subnet": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet2: "regional-subnet",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"regional-subnet"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": client.GenericListener{
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": client.GenericBackendSetDetails{
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"override default subnet via subnet1 and subnet2 annotation": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "annotation-one",
						ServiceAnnotationLoadBalancerSubnet2: "annotation-two",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						v1.ServicePort{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"annotation-one", "annotation-two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": client.GenericListener{
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": client.GenericBackendSetDetails{
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		//"security list manager annotation":
		"custom shape": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShape: "8000Mbps",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "8000Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"custom idle connection timeout": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerConnectionIdleTimeout: "404",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
						ConnectionConfiguration: &client.GenericConnectionConfiguration{
							IdleTimeout: common.Int64(404),
						},
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"custom proxy protocol version w/o timeout for multiple listeners": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerConnectionProxyProtocolVersion: "2",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
						{
							Protocol: "HTTP",
							Port:     int32(443),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
						ConnectionConfiguration: &client.GenericConnectionConfiguration{
							IdleTimeout:                    common.Int64(300), // fallback to default timeout for TCP
							BackendTcpProxyProtocolVersion: common.Int(2),
						},
					},
					"HTTP-443": {
						Name:                  common.String("HTTP-443"),
						DefaultBackendSetName: common.String("HTTP-443"),
						Port:                  common.Int(443),
						Protocol:              common.String("HTTP"),
						ConnectionConfiguration: &client.GenericConnectionConfiguration{
							IdleTimeout:                    common.Int64(60), // fallback to default timeout for HTTP
							BackendTcpProxyProtocolVersion: common.Int(2),
						},
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
					"HTTP-443": {
						Name:      common.String("HTTP-443"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
					"HTTP-443": {
						ListenerPort:      443,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"custom proxy protocol version and timeout": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerConnectionIdleTimeout:          "404",
						ServiceAnnotationLoadBalancerConnectionProxyProtocolVersion: "2",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
						ConnectionConfiguration: &client.GenericConnectionConfiguration{
							IdleTimeout:                    common.Int64(404),
							BackendTcpProxyProtocolVersion: common.Int(2),
						},
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"protocol annotation set to http": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerBEProtocol: "HTTP",
						ServiceAnnotationLoadBalancerSubnet1:    "annotation-one",
						ServiceAnnotationLoadBalancerSubnet2:    "annotation-two",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"annotation-one", "annotation-two"},
				Listeners: map[string]client.GenericListener{
					"HTTP-80": {
						Name:                  common.String("HTTP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("HTTP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"protocol annotation set to tcp": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerBEProtocol: "TCP",
						ServiceAnnotationLoadBalancerSubnet1:    "annotation-one",
						ServiceAnnotationLoadBalancerSubnet2:    "annotation-two",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"annotation-one", "annotation-two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"protocol annotation empty": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerBEProtocol: "",
						ServiceAnnotationLoadBalancerSubnet1:    "annotation-one",
						ServiceAnnotationLoadBalancerSubnet2:    "annotation-two",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"annotation-one", "annotation-two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"LBSpec returned with proper SSLConfiguration": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(443),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					fmt.Sprintf("TCP-443"): {
						Name:                  common.String("TCP-443"),
						DefaultBackendSetName: common.String("TCP-443"),
						Port:                  common.Int(443),
						Protocol:              common.String("TCP"),
						SslConfiguration: &client.GenericSslConfigurationDetails{
							CertificateName:       &listenerSecret,
							VerifyDepth:           common.Int(0),
							VerifyPeerCertificate: common.Bool(false),
						},
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-443": {
						Name:     common.String("TCP-443"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						SslConfiguration: &client.GenericSslConfigurationDetails{
							CertificateName:       &backendSecret,
							VerifyDepth:           common.Int(0),
							VerifyPeerCertificate: common.Bool(false),
						},
						IpVersion: GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-443": {
						ListenerPort:      443,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager: newSecurityListManagerNOOP(),
				SSLConfig: &SSLConfig{
					Ports:                   sets.NewInt(443),
					ListenerSSLSecretName:   listenerSecret,
					BackendSetSSLSecretName: backendSecret,
				},
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
			sslConfig: &SSLConfig{
				Ports:                   sets.NewInt(443),
				ListenerSSLSecretName:   listenerSecret,
				BackendSetSSLSecretName: backendSecret,
			},
		},
		"custom health check config": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerHealthCheckRetries:  "1",
						ServiceAnnotationLoadBalancerHealthCheckTimeout:  "1000",
						ServiceAnnotationLoadBalancerHealthCheckInterval: "3000",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(1),
							TimeoutInMillis:  common.Int(1000),
							IntervalInMillis: common.Int(3000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"flex shape": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShape:        "Flexible",
						ServiceAnnotationLoadBalancerShapeFlexMin: "10",
						ServiceAnnotationLoadBalancerShapeFlexMax: "80",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "flexible",
				FlexMin:  &tenMbps,
				FlexMax:  &eightyMbps,
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"valid loadbalancer policy": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShape:  "8000Mbps",
						ServiceAnnotationLoadBalancerPolicy: "IP_HASH",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "8000Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("IP_HASH"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"default loadbalancer policy": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShape: "8000Mbps",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "8000Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"load balancer with reserved ip": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShape: "8000Mbps",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					LoadBalancerIP:  "10.0.0.0",
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "8000Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				LoadBalancerIP:              "10.0.0.0",
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"defaults with tags": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialFreeformTagsOverride: `{"cluster":"resource", "unique":"tag"}`,
						ServiceAnnotationLoadBalancerInitialDefinedTagsOverride:  `{"namespace":{"key":"value", "owner":"team"}}`,
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					FreeformTags: map[string]string{"cluster": "cluster"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"cluster": "name", "owner": "cluster"}},
				},
			},

			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				FreeformTags: map[string]string{"cluster": "resource", "unique": "tag"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
			},
		},
		"merge default tags with common tags": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialFreeformTagsOverride: `{"cluster":"resource", "unique":"tag"}`,
						ServiceAnnotationLoadBalancerInitialDefinedTagsOverride:  `{"namespace":{"key":"value", "owner":"team"}, "namespace2": {"cost": "staging"}}`,
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					DefinedTags: map[string]map[string]interface{}{"namespace": {"cluster": "name", "owner": "cluster"}},
				},
				Common: &providercfg.TagConfig{
					DefinedTags: map[string]map[string]interface{}{"namespace": {"cluster": "CommonCluster", "owner": "CommonClusterOwner"}},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				FreeformTags: map[string]string{"cluster": "resource", "unique": "tag"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"cluster": "CommonCluster", "owner": "CommonClusterOwner"}, "namespace2": {"cost": "staging"}},
			},
		},
		"merge intial lb tags with common tags": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					FreeformTags: map[string]string{"cluster": "testname", "project": "pre-prod"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"cluster": "name", "owner": "cluster"}},
				},
				Common: &providercfg.TagConfig{
					FreeformTags: map[string]string{"access": "developers"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"cluster": "CommonCluster", "owner": "CommonClusterOwner"}, "cost": {"unit": "shared", "env": "pre-prod"}},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				FreeformTags: map[string]string{"cluster": "testname", "project": "pre-prod", "access": "developers"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"cluster": "CommonCluster", "owner": "CommonClusterOwner"}, "cost": {"unit": "shared", "env": "pre-prod"}},
			},
		},
		"SingleStack IPv6 - NLB": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv6},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Type:    v1.NodeInternalIP,
								Address: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv6)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "kube-system/testservice/test-uid",
				Type:     "nlb",
				Shape:    "flexible",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80-IPv6": {
						Name:                  common.String("TCP-80-IPv6"),
						DefaultBackendSetName: common.String("TCP-80-IPv6"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
						IpVersion:             GenericIpVersion(client.GenericIPv6),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80-IPv6": {
						Name:     common.String("TCP-80-IPv6"),
						Backends: []client.GenericBackend{{IpAddress: common.String("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), Port: common.Int(0), Weight: common.Int(1)}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("FIVE_TUPLE"),
						IpVersion:        GenericIpVersion(client.GenericIPv6),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"::/0"},
				Ports: map[string]portSpec{
					"TCP-80-IPv6": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv6},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv6),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv6},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Type:    v1.NodeInternalIP,
									Address: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"Prefer DualStack IPv4 and IPv6 LB": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4, IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyPreferDualStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4AndIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Type:    v1.NodeInternalIP,
								Address: "10.0.0.1",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "lb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy:  (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyPreferDualStack))),
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("10.0.0.1"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0", "::/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4, IPv6},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyPreferDualStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4AndIPv6),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Type:    v1.NodeInternalIP,
									Address: "10.0.0.1",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"PreferDualStack IPv4 and IPv6 NLB": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4, IPv6},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyPreferDualStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4AndIPv6),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4, client.GenericIPv6},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Type:    v1.NodeInternalIP,
								Address: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
							},
							{
								Type:    v1.NodeInternalIP,
								Address: "10.0.0.1",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy:  (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyPreferDualStack))),
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "kube-system/testservice/test-uid",
				Type:     "nlb",
				Shape:    "flexible",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
						IpVersion:             GenericIpVersion(client.GenericIPv4),
					},
					"TCP-80-IPv6": {
						Name:                  common.String("TCP-80-IPv6"),
						DefaultBackendSetName: common.String("TCP-80-IPv6"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
						IpVersion:             GenericIpVersion(client.GenericIPv6),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:     common.String("TCP-80"),
						Backends: []client.GenericBackend{{IpAddress: common.String("10.0.0.1"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("FIVE_TUPLE"),
						IpVersion:        GenericIpVersion(client.GenericIPv4),
					},
					"TCP-80-IPv6": {
						Name:     common.String("TCP-80-IPv6"),
						Backends: []client.GenericBackend{{IpAddress: common.String("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), Port: common.Int(0), Weight: common.Int(1)}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("FIVE_TUPLE"),
						IpVersion:        GenericIpVersion(client.GenericIPv6),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0", "::/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
					"TCP-80-IPv6": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4, IPv6},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicyPreferDualStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4AndIPv6),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4, client.GenericIPv6},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Type:    v1.NodeInternalIP,
									Address: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
								},
								{
									Type:    v1.NodeInternalIP,
									Address: "10.0.0.1",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
		"GRPC listeners": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerBEProtocol: "GRPC",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(443),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one", "two"},
				Listeners: map[string]client.GenericListener{
					fmt.Sprintf("GRPC-443"): {
						Name:                  common.String("GRPC-443"),
						DefaultBackendSetName: common.String("TCP-443"),
						Port:                  common.Int(443),
						Protocol:              common.String("GRPC"),
						SslConfiguration: &client.GenericSslConfigurationDetails{
							CertificateName:       &listenerSecret,
							VerifyDepth:           common.Int(0),
							VerifyPeerCertificate: common.Bool(false),
							CipherSuiteName:       common.String(DefaultCipherSuiteForGRPC),
						},
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-443": {
						Name:     common.String("TCP-443"),
						Backends: []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
						SslConfiguration: &client.GenericSslConfigurationDetails{
							CertificateName:       &backendSecret,
							VerifyDepth:           common.Int(0),
							VerifyPeerCertificate: common.Bool(false),
						},
						IpVersion: GenericIpVersion(client.GenericIPv4),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-443": {
						ListenerPort:      443,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager: newSecurityListManagerNOOP(),
				SSLConfig: &SSLConfig{
					Ports:                   sets.NewInt(443),
					ListenerSSLSecretName:   listenerSecret,
					BackendSetSSLSecretName: backendSecret,
				},
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
			sslConfig: &SSLConfig{
				Ports:                   sets.NewInt(443),
				ListenerSSLSecretName:   listenerSecret,
				BackendSetSSLSecretName: backendSecret,
			},
		},
	}

	cp := &CloudProvider{
		client: MockOCIClient{},
		config: &providercfg.Config{CompartmentID: "testCompartment"},
	}

	for name, tc := range testCases {
		logger := zap.L()
		t.Run(name, func(t *testing.T) {
			// we expect the service to be unchanged
			tc.expected.service = tc.service
			cp.config = &providercfg.Config{
				LoadBalancer: &providercfg.LoadBalancerConfig{
					Subnet1: tc.defaultSubnetOne,
					Subnet2: tc.defaultSubnetTwo,
				},
			}
			subnets, err := cp.getLoadBalancerSubnets(context.Background(), logger.Sugar(), tc.service)
			if err != nil {
				t.Error(err)
			}
			slManagerFactory := func(mode string) securityListManager {
				return newSecurityListManagerNOOP()
			}

			result, err := NewLBSpec(logger.Sugar(), tc.service, tc.nodes, subnets, tc.sslConfig, slManagerFactory, tc.IpVersions, tc.clusterTags, nil)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				results, _ := json.Marshal(result)
				expected, _ := json.Marshal(tc.expected)
				t.Errorf("Expected load balancer spec failed\nExpected: %s\nResults: %s\n", expected, results)
			}
		})
	}
}

func TestNewLBSpecForTags(t *testing.T) {
	enableOkeSystemTags = true
	tests := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		nodes            []*v1.Node
		virtualPods      []*v1.Pod
		service          *v1.Service
		sslConfig        *SSLConfig
		expected         *LBSpec
		clusterTags      *providercfg.InitialTags
		featureEnabled   bool
		IpVersions       *IpVersions
	}{
		"no resource & cluster level tags but common tags from config": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					DefinedTags: map[string]map[string]interface{}{"namespace": {"cluster": "name", "owner": "cluster"}},
				},
				Common: &providercfg.TagConfig{
					DefinedTags: map[string]map[string]interface{}{"namespace": {"cluster": "CommonCluster", "owner": "CommonClusterOwner"}},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				FreeformTags: map[string]string{},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"cluster": "CommonCluster", "owner": "CommonClusterOwner"}},
			},
			featureEnabled: true,
		},
		"no resource or cluster level tags and no common tags": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
			featureEnabled: true,
		},
		"resource level tags with common tags from config": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                               "nlb",
						ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride: `{"cluster":"resource", "unique":"tag"}`,
						ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride:  `{"namespace":{"key":"value", "owner":"team"}}`,
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				Common: &providercfg.TagConfig{
					FreeformTags: map[string]string{"name": "development_cluster"},
					DefinedTags:  map[string]map[string]interface{}{"namespace2": {"owner2": "team2", "key2": "value2"}},
				},
			},
			expected: &LBSpec{
				Name:     "kube-system/testservice/test-uid",
				Type:     "nlb",
				Shape:    "flexible",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
						IpVersion:             GenericIpVersion(client.GenericIPv4),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("FIVE_TUPLE"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				FreeformTags: map[string]string{"cluster": "resource", "unique": "tag", "name": "development_cluster"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}, "namespace2": {"owner2": "team2", "key2": "value2"}},
			},
			featureEnabled: true,
		},
		"resource level defined tags and common defined tags from config with same key": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialFreeformTagsOverride: `{"cluster":"resource", "unique":"tag"}`,
						ServiceAnnotationLoadBalancerInitialDefinedTagsOverride:  `{"namespace":{"key":"value", "owner":"team"}}`,
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				Common: &providercfg.TagConfig{
					FreeformTags: map[string]string{"name": "development_cluster"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner2": "team2", "key2": "value2"}},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				FreeformTags: map[string]string{"cluster": "resource", "unique": "tag", "name": "development_cluster"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner2": "team2", "key2": "value2"}},
			},
			featureEnabled: true,
		},
		"cluster level tags and common tags": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					FreeformTags: map[string]string{"lbname": "development_cluster_loadbalancer"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
				},
				Common: &providercfg.TagConfig{
					FreeformTags: map[string]string{"name": "development_cluster"},
					DefinedTags:  map[string]map[string]interface{}{"namespace2": {"owner2": "team2", "key2": "value2"}},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				FreeformTags: map[string]string{"lbname": "development_cluster_loadbalancer", "name": "development_cluster"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}, "namespace2": {"owner2": "team2", "key2": "value2"}},
			},
			featureEnabled: true,
		},
		"cluster level defined tags and common defined tags with same key": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					FreeformTags: map[string]string{"lbname": "development_cluster_loadbalancer"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
				},
				Common: &providercfg.TagConfig{
					FreeformTags: map[string]string{"name": "development_cluster"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner2": "team2", "key2": "value2"}},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				FreeformTags: map[string]string{"lbname": "development_cluster_loadbalancer", "name": "development_cluster"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner2": "team2", "key2": "value2"}},
			},
			featureEnabled: true,
		},
		"cluster level tags with no common tags": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					FreeformTags: map[string]string{"lbname": "development_cluster_loadbalancer"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				FreeformTags: map[string]string{"lbname": "development_cluster_loadbalancer"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
			},
			featureEnabled: true,
		},
		"no cluster or level tags but common tags from config": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				Common: &providercfg.TagConfig{
					FreeformTags: map[string]string{"lbname": "development_cluster_loadbalancer"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				FreeformTags: map[string]string{"lbname": "development_cluster_loadbalancer"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
			},
			featureEnabled: true,
		},
		"when the feature is disabled": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			clusterTags: &providercfg.InitialTags{
				Common: &providercfg.TagConfig{
					FreeformTags: map[string]string{"lbname": "development_cluster_loadbalancer"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
			featureEnabled: false,
		},
	}
	cp := &CloudProvider{
		client: MockOCIClient{},
		config: &providercfg.Config{CompartmentID: "testCompartment"},
	}

	for name, tc := range tests {
		logger := zap.L()
		enableOkeSystemTags = tc.featureEnabled
		t.Run(name, func(t *testing.T) {
			// we expect the service to be unchanged
			tc.expected.service = tc.service
			cp.config = &providercfg.Config{
				LoadBalancer: &providercfg.LoadBalancerConfig{
					Subnet1: tc.defaultSubnetOne,
					Subnet2: tc.defaultSubnetTwo,
				},
			}
			subnets, err := cp.getLoadBalancerSubnets(context.Background(), logger.Sugar(), tc.service)
			if err != nil {
				t.Error(err)
			}
			slManagerFactory := func(mode string) securityListManager {
				return newSecurityListManagerNOOP()
			}
			result, err := NewLBSpec(logger.Sugar(), tc.service, tc.nodes, subnets, tc.sslConfig, slManagerFactory, tc.IpVersions, tc.clusterTags, nil)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected load balancer spec\n%+v\nbut got\n%+v", tc.expected, result)
			}
		})
	}
}

func TestNewLBSpecSingleAD(t *testing.T) {
	testCases := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		nodes            []*v1.Node
		virtualPods      []*v1.Pod
		service          *v1.Service
		expected         *LBSpec
		clusterTags      *providercfg.InitialTags
		IpVersions       *IpVersions
	}{
		"single subnet for single AD": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			nodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "0.0.0.0",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerBEProtocol: "",
						ServiceAnnotationLoadBalancerSubnet1:    "annotation-one",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
			},
			expected: &LBSpec{
				Name:     "test-uid",
				Type:     "lb",
				Shape:    "100Mbps",
				Internal: false,
				Subnets:  []string{"annotation-one"},
				Listeners: map[string]client.GenericListener{
					"TCP-80": {
						Name:                  common.String("TCP-80"),
						DefaultBackendSetName: common.String("TCP-80"),
						Port:                  common.Int(80),
						Protocol:              common.String("TCP"),
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Name:      common.String("TCP-80"),
						IpVersion: GenericIpVersion(client.GenericIPv4),
						Backends:  []client.GenericBackend{{IpAddress: common.String("0.0.0.0"), Port: common.Int(0), Weight: common.Int(1), TargetId: &testNodeString}},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							IsForcePlainText: common.Bool(false),
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(false),
						Policy:           common.String("ROUND_ROBIN"),
					},
				},
				IsPreserveSource:        common.Bool(false),
				NetworkSecurityGroupIds: []string{},
				SourceCIDRs:             []string{"0.0.0.0/0"},
				Ports: map[string]portSpec{
					"TCP-80": {
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager:         newSecurityListManagerNOOP(),
				ManagedNetworkSecurityGroup: &ManagedNetworkSecurityGroup{frontendNsgId: "", backendNsgId: []string{}, nsgRuleManagementMode: ManagementModeNone},
				IpVersions: &IpVersions{
					IpFamilies:               []string{IPv4},
					IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
					LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
					ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
				},
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
			},
		},
	}

	cp := &CloudProvider{
		client: MockOCIClient{},
		config: &providercfg.Config{CompartmentID: "testCompartment"},
	}

	for name, tc := range testCases {
		logger := zap.L()
		t.Run(name, func(t *testing.T) {
			// we expect the service to be unchanged
			tc.expected.service = tc.service
			cp.config = &providercfg.Config{
				LoadBalancer: &providercfg.LoadBalancerConfig{
					Subnet1: tc.defaultSubnetOne,
					Subnet2: tc.defaultSubnetTwo,
				},
			}
			subnets, err := cp.getLoadBalancerSubnets(context.Background(), logger.Sugar(), tc.service)
			if err != nil {
				t.Error(err)
			}
			slManagerFactory := func(mode string) securityListManager {
				return newSecurityListManagerNOOP()
			}

			result, err := NewLBSpec(logger.Sugar(), tc.service, tc.nodes, subnets, nil, slManagerFactory, tc.IpVersions, tc.clusterTags, nil)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected load balancer spec\n%+v\nbut got\n%+v", tc.expected, result)
			}
		})
	}
}

func TestNewLBSpecFailure(t *testing.T) {
	testCases := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		nodes            []*v1.Node
		virtualPods      []*v1.Pod
		service          *v1.Service
		//add cp or cp security list
		expectedErrMsg string
		clusterTags    *providercfg.InitialTags
		IpVersions     *IpVersions
	}{
		"unsupported udp protocol": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolUDP},
					},
				},
			},
			expectedErrMsg: "invalid service: OCI load balancers do not support UDP",
		},
		"unsupported session affinity": {
			defaultSubnetOne: "one",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityClientIP,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "invalid service: OCI only supports SessionAffinity \"None\" currently",
		},
		"invalid idle connection timeout": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerConnectionIdleTimeout: "whoops",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "error parsing service annotation: service.beta.kubernetes.io/oci-load-balancer-connection-idle-timeout=whoops",
		},
		"invalid connection proxy protocol version": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerConnectionProxyProtocolVersion: "bla",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "error parsing service annotation: service.beta.kubernetes.io/oci-load-balancer-connection-proxy-protocol-version=bla",
		},
		"internal lb missing subnet1": {
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "true",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports:           []v1.ServicePort{},
					//add security list mananger in spec
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"internal lb with empty subnet1 annotation": {
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "true",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports:           []v1.ServicePort{},
					//add security list mananger in spec
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"non boolean internal lb": {
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "yes",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports:           []v1.ServicePort{},
				},
			},
			expectedErrMsg: fmt.Sprintf("invalid value: yes provided for annotation: %s: strconv.ParseBool: parsing \"yes\": invalid syntax", ServiceAnnotationLoadBalancerInternal),
		},
		"invalid flex shape missing min": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShape:        "flexible",
						ServiceAnnotationLoadBalancerShapeFlexMax: "80",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "error parsing service annotation: service.beta.kubernetes.io/oci-load-balancer-shape=flexible requires service.beta.kubernetes.io/oci-load-balancer-shape-flex-min and service.beta.kubernetes.io/oci-load-balancer-shape-flex-max to be set",
		},
		"invalid flex shape missing max": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShape:        "flexible",
						ServiceAnnotationLoadBalancerShapeFlexMin: "10",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "error parsing service annotation: service.beta.kubernetes.io/oci-load-balancer-shape=flexible requires service.beta.kubernetes.io/oci-load-balancer-shape-flex-min and service.beta.kubernetes.io/oci-load-balancer-shape-flex-max to be set",
		},
		"invalid loadbalancer policy": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShapeFlexMin: "10Mbps",
						ServiceAnnotationLoadBalancerPolicy:       "not-valid-loadbalancer-policy",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: `loadbalancer policy "not-valid-loadbalancer-policy" is not valid`,
		},
		"invalid loadBalancerIP format": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					LoadBalancerIP:  "non-ip-format",
					SessionAffinity: v1.ServiceAffinityNone,
				},
			},
			expectedErrMsg: "invalid value \"non-ip-format\" provided for LoadBalancerIP",
		},
		"unsupported loadBalancerIP for internal load balancer": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "true",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					LoadBalancerIP:  "10.0.0.0",
					SessionAffinity: v1.ServiceAffinityNone,
					Ports:           []v1.ServicePort{},
				},
			},
			expectedErrMsg: `invalid service: cannot create a private load balancer with Reserved IP`,
		},
		"invalid defined tags": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialDefinedTagsOverride: "whoops",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "failed to parse defined tags annotation: invalid character 'w' looking for beginning of value",
		},
		"empty subnets": {
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"empty strings for subnets": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"empty string for subnet1 annotation": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "",
						ServiceAnnotationLoadBalancerSubnet2: "annotation-two",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"default string for cloud config subnet2": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "random",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "",
						ServiceAnnotationLoadBalancerSubnet2: "",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"regional string for subnet2 annotation": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "",
			IpVersions: &IpVersions{
				IpFamilies:               []string{IPv4},
				IpFamilyPolicy:           common.String(string(v1.IPFamilyPolicySingleStack)),
				LbEndpointIpVersion:      GenericIpVersion(client.GenericIPv4),
				ListenerBackendIpVersion: []client.GenericIpVersion{client.GenericIPv4},
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "",
						ServiceAnnotationLoadBalancerSubnet2: "",
					},
				},
				Spec: v1.ServiceSpec{
					IPFamilies:      []v1.IPFamily{v1.IPFamily(IPv4)},
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
	}

	cp := &CloudProvider{
		client: MockOCIClient{},
		config: &providercfg.Config{CompartmentID: "testCompartment"},
	}

	for name, tc := range testCases {
		logger := zap.L()
		t.Run(name, func(t *testing.T) {
			cp.config = &providercfg.Config{
				LoadBalancer: &providercfg.LoadBalancerConfig{
					Subnet1: tc.defaultSubnetOne,
					Subnet2: tc.defaultSubnetTwo,
				},
			}
			subnets, err := cp.getLoadBalancerSubnets(context.Background(), logger.Sugar(), tc.service)
			tc.service.Spec.IPFamilies = []v1.IPFamily{v1.IPFamily(IPv4)}
			if err == nil {
				slManagerFactory := func(mode string) securityListManager {
					return newSecurityListManagerNOOP()
				}
				_, err = NewLBSpec(logger.Sugar(), tc.service, tc.nodes, subnets, nil, slManagerFactory, tc.IpVersions, tc.clusterTags, nil)
			}
			if err == nil || err.Error() != tc.expectedErrMsg {
				t.Errorf("Expected error with message %q but got %q", tc.expectedErrMsg, err)
			}
		})
	}
}

func TestNewSSLConfig(t *testing.T) {
	testCases := map[string]struct {
		secretListenerString   string
		secretBackendSetString string
		service                *v1.Service
		ports                  []int
		ssr                    sslSecretReader

		expectedResult *SSLConfig
	}{
		"noopSSLSecretReader if ssr is nil and uses the default service namespace": {
			secretListenerString:   "listenerSecretName",
			secretBackendSetString: "backendSetSecretName",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
			},
			ports: []int{8080},
			ssr:   nil,

			expectedResult: &SSLConfig{
				Ports:                        sets.NewInt(8080),
				ListenerSSLSecretName:        "listenerSecretName",
				ListenerSSLSecretNamespace:   "default",
				BackendSetSSLSecretName:      "backendSetSecretName",
				BackendSetSSLSecretNamespace: "default",
				sslSecretReader:              noopSSLSecretReader{},
			},
		},
		"ssr is assigned if provided and uses the default service namespace": {
			secretListenerString:   "listenerSecretName",
			secretBackendSetString: "backendSetSecretName",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
			},
			ports: []int{8080},
			ssr:   &mockSSLSecretReader{},

			expectedResult: &SSLConfig{
				Ports:                        sets.NewInt(8080),
				ListenerSSLSecretName:        "listenerSecretName",
				ListenerSSLSecretNamespace:   "default",
				BackendSetSSLSecretName:      "backendSetSecretName",
				BackendSetSSLSecretNamespace: "default",
				sslSecretReader:              &mockSSLSecretReader{},
			},
		},
		"If namespace is specified in secret string, use it": {
			secretListenerString:   "namespaceone/listenerSecretName",
			secretBackendSetString: "namespacetwo/backendSetSecretName",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
				},
			},
			ports: []int{8080},
			ssr:   &mockSSLSecretReader{},

			expectedResult: &SSLConfig{
				Ports:                        sets.NewInt(8080),
				ListenerSSLSecretName:        "listenerSecretName",
				ListenerSSLSecretNamespace:   "namespaceone",
				BackendSetSSLSecretName:      "backendSetSecretName",
				BackendSetSSLSecretNamespace: "namespacetwo",
				sslSecretReader:              &mockSSLSecretReader{},
			},
		},
		"Empty secret string results in empty name and namespace": {
			ports: []int{8080},
			ssr:   &mockSSLSecretReader{},

			expectedResult: &SSLConfig{
				Ports:                        sets.NewInt(8080),
				ListenerSSLSecretName:        "",
				ListenerSSLSecretNamespace:   "",
				BackendSetSSLSecretName:      "",
				BackendSetSSLSecretNamespace: "",
				sslSecretReader:              &mockSSLSecretReader{},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := NewSSLConfig(tc.secretListenerString, tc.secretBackendSetString, tc.service, tc.ports, tc.ssr)
			if !reflect.DeepEqual(result, tc.expectedResult) {
				t.Errorf("Expected SSlConfig \n%+v\nbut got\n%+v", tc.expectedResult, result)
			}
		})
	}
}

func TestCertificates(t *testing.T) {

	backendSecretCaCert := "cacert1"
	backendSecretPublicCert := "publiccert1"
	backendSecretPrivateKey := "privatekey1"
	backendSecretPassphrase := "passphrase1"

	listenerSecretCaCert := "cacert2"
	listenerSecretPublicCert := "publiccert2"
	listenerSecretPrivateKey := "privatekey2"
	listenerSecretPassphrase := "passphrase2"

	testCases := map[string]struct {
		lbSpec         *LBSpec
		expectedResult map[string]client.GenericCertificate
		expectError    bool
	}{
		"No SSLConfig results in empty certificate details array": {
			expectError:    false,
			lbSpec:         &LBSpec{},
			expectedResult: make(map[string]client.GenericCertificate),
		},
		"Return backend SSL secret": {
			expectError: false,
			lbSpec: &LBSpec{
				service: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "testnamespace",
					},
				},
				SSLConfig: &SSLConfig{
					BackendSetSSLSecretName:      backendSecret,
					BackendSetSSLSecretNamespace: "backendnamespace",
					sslSecretReader: &mockSSLSecretReader{
						returnError: false,
						returnMap: map[struct {
							namespaceArg string
							nameArg      string
						}]*certificateData{
							{namespaceArg: "backendnamespace", nameArg: backendSecret}: {
								Name:       "certificatename",
								CACert:     []byte(backendSecretCaCert),
								PublicCert: []byte(backendSecretPublicCert),
								PrivateKey: []byte(backendSecretPrivateKey),
								Passphrase: []byte(backendSecretPassphrase),
							},
						},
					},
				},
			},
			expectedResult: map[string]client.GenericCertificate{
				backendSecret: {
					CertificateName:   &backendSecret,
					CaCertificate:     &backendSecretCaCert,
					Passphrase:        &backendSecretPassphrase,
					PrivateKey:        &backendSecretPrivateKey,
					PublicCertificate: &backendSecretPublicCert,
				},
			},
		},
		"Return both backend and listener SSL secret": {
			expectError: false,
			lbSpec: &LBSpec{
				service: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "testnamespace",
					},
				},
				SSLConfig: &SSLConfig{
					BackendSetSSLSecretName:      backendSecret,
					BackendSetSSLSecretNamespace: "backendnamespace",
					ListenerSSLSecretName:        listenerSecret,
					ListenerSSLSecretNamespace:   "listenernamespace",
					sslSecretReader: &mockSSLSecretReader{
						returnError: false,
						returnMap: map[struct {
							namespaceArg string
							nameArg      string
						}]*certificateData{
							{namespaceArg: "backendnamespace", nameArg: backendSecret}: {
								Name:       "backendcertificatename",
								CACert:     []byte(backendSecretCaCert),
								PublicCert: []byte(backendSecretPublicCert),
								PrivateKey: []byte(backendSecretPrivateKey),
								Passphrase: []byte(backendSecretPassphrase),
							},
							{namespaceArg: "listenernamespace", nameArg: listenerSecret}: {
								Name:       "listenercertificatename",
								CACert:     []byte(listenerSecretCaCert),
								PublicCert: []byte(listenerSecretPublicCert),
								PrivateKey: []byte(listenerSecretPrivateKey),
								Passphrase: []byte(listenerSecretPassphrase),
							},
						},
					},
				},
			},
			expectedResult: map[string]client.GenericCertificate{
				backendSecret: {
					CertificateName:   &backendSecret,
					CaCertificate:     &backendSecretCaCert,
					Passphrase:        &backendSecretPassphrase,
					PrivateKey:        &backendSecretPrivateKey,
					PublicCertificate: &backendSecretPublicCert,
				},
				listenerSecret: {
					CertificateName:   &listenerSecret,
					CaCertificate:     &listenerSecretCaCert,
					Passphrase:        &listenerSecretPassphrase,
					PrivateKey:        &listenerSecretPrivateKey,
					PublicCertificate: &listenerSecretPublicCert,
				},
			},
		},
		"Error returned from SSL secret reader is handled gracefully": {
			expectError: true,
			lbSpec: &LBSpec{
				service: &v1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "testnamespace",
					},
				},
				SSLConfig: &SSLConfig{
					BackendSetSSLSecretName: backendSecret,
					sslSecretReader: &mockSSLSecretReader{
						returnError: true,
					},
				},
			},
			expectedResult: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			certDetails, err := tc.lbSpec.Certificates()
			if err != nil && !tc.expectError {
				t.Errorf("Was not expected an error to be returned, but got one:\n%+v", err)
			}
			if !reflect.DeepEqual(certDetails, tc.expectedResult) {
				t.Errorf("Expected certificate details \n%+v\nbut got\n%+v", tc.expectedResult, certDetails)
			}
		})
	}
}

func TestRequiresCertificate(t *testing.T) {
	testCases := map[string]struct {
		expected    bool
		annotations map[string]string
	}{
		"Contains the Load Balancer SSL Ports Annotation": {
			expected: true,
			annotations: map[string]string{
				ServiceAnnotationLoadBalancerSSLPorts: "443",
			},
		},
		"Does not contain the Load Balancer SSL Ports Annotation": {
			expected:    false,
			annotations: make(map[string]string, 0),
		},
		"Always false for NLBs": {
			expected: false,
			annotations: map[string]string{
				ServiceAnnotationLoadBalancerSSLPorts: "443",
				ServiceAnnotationLoadBalancerType:     "nlb",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := requiresCertificate(&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tc.annotations,
				},
			})
			if result != tc.expected {
				t.Error("Did not get the correct result")
			}
		})
	}
}

func TestRequiresFrontendNsg(t *testing.T) {
	testCases := map[string]struct {
		expected    bool
		annotations map[string]string
	}{
		"Contains annotation for NSG Rule management": {
			expected: true,
			annotations: map[string]string{
				ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "NSG",
			},
		},
		"Does not contain annotation for NSG Rule management": {
			expected:    false,
			annotations: make(map[string]string, 0),
		},
		"Contains annotation (NLB) for NSG Rule management": {
			expected: true,
			annotations: map[string]string{
				ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "NSG",
				ServiceAnnotationLoadBalancerType:                       "nlb",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := requiresNsgManagement(&v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tc.annotations,
				},
			})
			if result != tc.expected {
				t.Error("Did not get the correct result")
			}
		})
	}
}

func Test_getBackends(t *testing.T) {
	type args struct {
		nodes       []*v1.Node
		virtualPods []*v1.Pod
		nodePort    int32
	}
	var tests = []struct {
		name     string
		args     args
		want     []client.GenericBackend
		wantIPv6 []client.GenericBackend
	}{
		{
			name:     "no nodes",
			args:     args{nodePort: 80},
			want:     []client.GenericBackend{},
			wantIPv6: []client.GenericBackend{},
		},
		{
			name: "single node with assigned IP",
			args: args{
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				nodePort: 80,
			},
			want: []client.GenericBackend{
				{IpAddress: common.String("0.0.0.0"), Port: common.Int(80), Weight: common.Int(1), TargetId: &testNodeString},
			},
			wantIPv6: []client.GenericBackend{},
		},
		{
			name: "single node with unassigned IP",
			args: args{
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:        nil,
							Allocatable:     nil,
							Phase:           "",
							Conditions:      nil,
							Addresses:       []v1.NodeAddress{},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				nodePort: 80,
			},
			want:     []client.GenericBackend{},
			wantIPv6: []client.GenericBackend{},
		},
		{
			name: "multiple nodes - all with assigned IP",
			args: args{
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.1",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				nodePort: 80,
			},
			want: []client.GenericBackend{
				{IpAddress: common.String("0.0.0.0"), Port: common.Int(80), Weight: common.Int(1), TargetId: &testNodeString},
				{IpAddress: common.String("0.0.0.1"), Port: common.Int(80), Weight: common.Int(1), TargetId: &testNodeString},
			},
			wantIPv6: []client.GenericBackend{},
		},
		{
			name: "multiple nodes - all with unassigned IP",
			args: args{
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:        nil,
							Allocatable:     nil,
							Phase:           "",
							Conditions:      nil,
							Addresses:       []v1.NodeAddress{},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec:       v1.NodeSpec{},
						Status: v1.NodeStatus{
							Capacity:        nil,
							Allocatable:     nil,
							Phase:           "",
							Conditions:      nil,
							Addresses:       []v1.NodeAddress{},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				nodePort: 80,
			},
			want:     []client.GenericBackend{},
			wantIPv6: []client.GenericBackend{},
		},
		{
			name: "multiple nodes - one with unassigned IP",
			args: args{
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.0",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:        nil,
							Allocatable:     nil,
							Phase:           "",
							Conditions:      nil,
							Addresses:       []v1.NodeAddress{},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "0.0.0.1",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				nodePort: 80,
			},
			want: []client.GenericBackend{
				{IpAddress: common.String("0.0.0.0"), Port: common.Int(80), Weight: common.Int(1), TargetId: &testNodeString},
				{IpAddress: common.String("0.0.0.1"), Port: common.Int(80), Weight: common.Int(1), TargetId: &testNodeString},
			},
			wantIPv6: []client.GenericBackend{},
		},
		{
			name: "multiple nodes - one with unassigned IP",
			args: args{
				nodes: []*v1.Node{
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "2001:0000:130F:0000:0000:09C0:876A:130B",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:        nil,
							Allocatable:     nil,
							Phase:           "",
							Conditions:      nil,
							Addresses:       []v1.NodeAddress{},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
					{
						TypeMeta:   metav1.TypeMeta{},
						ObjectMeta: metav1.ObjectMeta{},
						Spec: v1.NodeSpec{
							ProviderID: testNodeString,
						},
						Status: v1.NodeStatus{
							Capacity:    nil,
							Allocatable: nil,
							Phase:       "",
							Conditions:  nil,
							Addresses: []v1.NodeAddress{
								{
									Address: "2001:0000:130F:0000:0000:09C0:876A:1300",
									Type:    "InternalIP",
								},
							},
							DaemonEndpoints: v1.NodeDaemonEndpoints{},
							NodeInfo:        v1.NodeSystemInfo{},
							Images:          nil,
							VolumesInUse:    nil,
							VolumesAttached: nil,
							Config:          nil,
						},
					},
				},
				nodePort: 80,
			},
			want: []client.GenericBackend{},
			wantIPv6: []client.GenericBackend{
				{IpAddress: common.String("2001:0000:130F:0000:0000:09C0:876A:130B"), Port: common.Int(80), Weight: common.Int(1)},
				{IpAddress: common.String("2001:0000:130F:0000:0000:09C0:876A:1300"), Port: common.Int(80), Weight: common.Int(1)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.L()
			gotIpv4, gotIpv6 := getBackends(logger.Sugar(), tt.args.nodes, tt.args.nodePort)
			if !reflect.DeepEqual(gotIpv4, tt.want) {
				t.Errorf("getBackends() = %+v, want %+v", gotIpv4, tt.want)
			}
			if !reflect.DeepEqual(gotIpv6, tt.wantIPv6) {
				t.Errorf("getBackends() = %+v, want %+v", gotIpv6, tt.wantIPv6)
			}

		})
	}
}

func TestIsInternal(t *testing.T) {
	testCases := map[string]struct {
		service    *v1.Service
		isInternal bool
		err        error
	}{
		"no ServiceAnnotationLoadBalancerInternal annotation": {
			service:    &v1.Service{},
			isInternal: false,
			err:        nil,
		},
		"ServiceAnnotationLoadBalancerInternal is true": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "true",
					},
				},
			},
			isInternal: true,
			err:        nil,
		},
		"ServiceAnnotationLoadBalancerInternal is TRUE": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "TRUE",
					},
				},
			},
			isInternal: true,
			err:        nil,
		},
		"ServiceAnnotationLoadBalancerInternal is FALSE": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "FALSE",
					},
				},
			},
			isInternal: false,
			err:        nil,
		},
		"ServiceAnnotationLoadBalancerInternal is false": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "FALSE",
					},
				},
			},
			isInternal: false,
			err:        nil,
		},
		"ServiceAnnotationLoadBalancerInternal is non boolean": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "yes",
					},
				},
			},
			isInternal: false,
			err:        fmt.Errorf("invalid value: yes provided for annotation: %s: strconv.ParseBool: parsing \"yes\": invalid syntax", ServiceAnnotationLoadBalancerInternal),
		},
		"no ServiceAnnotationNetworkLoadBalancerInternal annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			isInternal: false,
			err:        nil,
		},
		"ServiceAnnotationNetworkLoadBalancerInternal is true": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:            "nlb",
						ServiceAnnotationNetworkLoadBalancerInternal: "true",
					},
				},
			},
			isInternal: true,
			err:        nil,
		},
		"ServiceAnnotationNetworkLoadBalancerInternal is TRUE": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:            "nlb",
						ServiceAnnotationNetworkLoadBalancerInternal: "TRUE",
					},
				},
			},
			isInternal: true,
			err:        nil,
		},
		"ServiceAnnotationNetworkLoadBalancerInternal is FALSE": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:            "nlb",
						ServiceAnnotationNetworkLoadBalancerInternal: "FALSE",
					},
				},
			},
			isInternal: false,
			err:        nil,
		},
		"ServiceAnnotationNetworkLoadBalancerInternal is false": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:            "nlb",
						ServiceAnnotationNetworkLoadBalancerInternal: "FALSE",
					},
				},
			},
			isInternal: false,
			err:        nil,
		},
		"ServiceAnnotationNetworkLoadBalancerInternal is non boolean": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:            "nlb",
						ServiceAnnotationNetworkLoadBalancerInternal: "yes",
					},
				},
			},
			isInternal: false,
			err:        fmt.Errorf("invalid value: yes provided for annotation: %s: strconv.ParseBool: parsing \"yes\": invalid syntax", ServiceAnnotationNetworkLoadBalancerInternal),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			internal, err := isInternalLB(tc.service)
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Expected internal LB error\n%+v\nbut got\n%+v", tc.err, err)
			}
			if internal != tc.isInternal {
				t.Errorf("Expected internal LB\n%+v\nbut got\n%+v", tc.isInternal, internal)
			}
		})
	}
}

func Test_getNetworkSecurityGroups(t *testing.T) {
	testCases := map[string]struct {
		service *v1.Service
		nsgList []string
		err     error
	}{
		"empty ServiceAnnotationLoadBalancerNetworkSecurityGroups annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNetworkSecurityGroups: "",
					},
				},
			},
			nsgList: []string{},
			err:     nil,
		},
		"no ServiceAnnotationLoadBalancerNetworkSecurityGroups annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			nsgList: []string{},
			err:     nil,
		},
		"ServiceAnnotationLoadBalancerNetworkSecurityGroups update annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNetworkSecurityGroups: "ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
				},
			},
			nsgList: []string{"ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			err:     nil,
		},
		"ServiceAnnotationLoadBalancerNetworkSecurityGroups Allow maximum NSG OCIDS (Max: 5)": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNetworkSecurityGroups: "ocid1,ocid2,ocid3,ocid4,ocid5",
					},
				},
			},
			nsgList: []string{"ocid1", "ocid2", "ocid3", "ocid4", "ocid5"},
			err:     fmt.Errorf("invalid number of Network Security Groups (Max: 5) provided for annotation: oci.oraclecloud.com/oci-network-security-groups"),
		},
		"ServiceAnnotationLoadBalancerNetworkSecurityGroups Exceed maximum NSG OCIDS": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNetworkSecurityGroups: "ocid1,ocid2,ocid3,ocid4,ocid5,ocid6",
					},
				},
			},
			nsgList: nil,
			err:     fmt.Errorf("invalid number of Network Security Groups (Max: 5) provided for annotation: oci.oraclecloud.com/oci-network-security-groups"),
		},
		"ServiceAnnotationLoadBalancerNetworkSecurityGroups Invalid NSG OCIDS": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNetworkSecurityGroups: "ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-;,ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbb",
					},
				},
			},
			nsgList: []string{"ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-;", "ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbb"},
			err:     nil,
		},
		"ServiceAnnotationLoadBalancerNetworkSecurityGroups duplicate NSG OCIDS": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerNetworkSecurityGroups: "ocid1,ocid2, ocid1",
					},
				},
			},
			nsgList: []string{"ocid1", "ocid2"},
			err:     nil,
		},
		"empty ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                         "nlb",
						ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups: "",
					},
				},
			},
			nsgList: []string{},
			err:     nil,
		},
		"no ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			nsgList: []string{},
			err:     nil,
		},
		"ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups update annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                         "nlb",
						ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups: "ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
				},
			},
			nsgList: []string{"ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			err:     nil,
		},
		"ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups Allow maximum NSG OCIDS (Max: 5)": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                         "nlb",
						ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups: "ocid1,ocid2,ocid3,ocid4,ocid5",
					},
				},
			},
			nsgList: []string{"ocid1", "ocid2", "ocid3", "ocid4", "ocid5"},
			err:     fmt.Errorf("invalid number of Network Security Groups (Max: 5) provided for annotation: oci-network-load-balancer.oraclecloud.com/oci-network-security-groups"),
		},
		"ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups Exceed maximum NSG OCIDS": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                         "nlb",
						ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups: "ocid1,ocid2,ocid3,ocid4,ocid5,ocid6",
					},
				},
			},
			nsgList: nil,
			err:     fmt.Errorf("invalid number of Network Security Groups (Max: 5) provided for annotation: oci-network-load-balancer.oraclecloud.com/oci-network-security-groups"),
		},
		"ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups Invalid NSG OCIDS": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                         "nlb",
						ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups: "ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-;,ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbb",
					},
				},
			},
			nsgList: []string{"ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-;", "ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaabbbbbbbbbbbbbbbbbbbbb"},
			err:     nil,
		},
		"ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups duplicate NSG OCIDS": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                         "nlb",
						ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups: "ocid1, ocid2, ocid1",
					},
				},
			},
			nsgList: []string{"ocid1", "ocid2"},
			err:     nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			nsgList, err := getNetworkSecurityGroupIds(tc.service)
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Expected  NSG List error\n%+v\nbut got\n%+v", tc.err, err)
			}
			if !reflect.DeepEqual(nsgList, tc.nsgList) {
				t.Errorf("Expected NSG List\n%+v\nbut got\n%+v", tc.nsgList, nsgList)
			}
		})
	}
}

func Test_getLoadBalancerTags(t *testing.T) {
	emptyInitialTags := providercfg.InitialTags{}
	emptyTags := providercfg.TagConfig{}
	testCases := map[string]struct {
		service       *v1.Service
		initialTags   *providercfg.InitialTags
		desiredLBTags *providercfg.TagConfig
		err           error
	}{
		"no tag annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: &emptyTags,
			err:           nil,
		},
		"empty ServiceAnnotationLoadBalancerInitialDefinedTagsOverride annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialDefinedTagsOverride: "",
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: &emptyTags,
			err:           nil,
		},
		"empty ServiceAnnotationLoadBalancerInitialFreeformTagsOverride annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialFreeformTagsOverride: "",
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: &emptyTags,
			err:           nil,
		},
		"invalid ServiceAnnotationLoadBalancerInitialFreeformTagsOverride annotation value": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialFreeformTagsOverride: "a",
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: nil,
			err:           errors.New("failed to parse free form tags annotation: invalid character 'a' looking for beginning of value"),
		},
		"invalid ServiceAnnotationLoadBalancerInitialDefinedTagsOverride annotation value": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialDefinedTagsOverride: "a",
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: nil,
			err:           errors.New("failed to parse defined tags annotation: invalid character 'a' looking for beginning of value"),
		},
		"invalid json in resource level freeform tags": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialFreeformTagsOverride: `{'test':'tag'}`,
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: nil,
			err:           errors.New(`failed to parse free form tags annotation: invalid character '\'' looking for beginning of object key string`),
		},
		"only resource level freeform tags": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialFreeformTagsOverride: `{"test":"tag"}`,
					},
				},
			},
			initialTags: &emptyInitialTags,
			desiredLBTags: &providercfg.TagConfig{
				FreeformTags: map[string]string{"test": "tag"},
				// Defined tags are always present as Oracle-Tags are added by default
			},
			err: nil,
		},
		"only resource level defined tags": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialDefinedTagsOverride: `{"namespace":{"key":"value"}}`,
					},
				},
			},
			initialTags: &emptyInitialTags,
			desiredLBTags: &providercfg.TagConfig{
				DefinedTags: map[string]map[string]interface{}{"namespace": {"key": "value"}},
			},
			err: nil,
		},
		"only cluster level defined tags": {
			service: &v1.Service{},
			initialTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					DefinedTags: map[string]map[string]interface{}{"namespace": {"key": "value"}},
				},
			},
			desiredLBTags: &providercfg.TagConfig{
				DefinedTags: map[string]map[string]interface{}{"namespace": {"key": "value"}},
			},
			err: nil,
		},
		"resource and cluster level tags, only resource level tags are added": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInitialFreeformTagsOverride: `{"cluster":"resource", "unique":"tag"}`,
						ServiceAnnotationLoadBalancerInitialDefinedTagsOverride:  `{"namespace":{"key":"value", "owner":"team"}}`,
					},
				},
			},
			initialTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					FreeformTags: map[string]string{"cluster": "cluster"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"cluster": "name", "owner": "cluster"}},
				},
			},
			desiredLBTags: &providercfg.TagConfig{
				FreeformTags: map[string]string{"cluster": "resource", "unique": "tag"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
			},
			err: nil,
		},
		"no tag annotation for nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: &emptyTags,
			err:           nil,
		},
		"empty ServiceAnnotationLoadBalancerInitialDefinedTagsOverride  NLB annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride: "",
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: &emptyTags,
			err:           nil,
		},
		"empty ServiceAnnotationLoadBalancerInitialFreeformTagsOverride NLB annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                               "nlb",
						ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride: "",
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: &emptyTags,
			err:           nil,
		},
		"invalid ServiceAnnotationLoadBalancerInitialFreeformTagsOverride NLB annotation value": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                               "nlb",
						ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride: "a",
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: nil,
			err:           errors.New("failed to parse free form tags annotation: invalid character 'a' looking for beginning of value"),
		},
		"invalid ServiceAnnotationLoadBalancerInitialDefinedTagsOverride NLB annotation value": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride: "a",
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: nil,
			err:           errors.New("failed to parse defined tags annotation: invalid character 'a' looking for beginning of value"),
		},
		"invalid json in resource level freeform tags for nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                               "nlb",
						ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride: `{'test':'tag'}`,
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: nil,
			err:           errors.New(`failed to parse free form tags annotation: invalid character '\'' looking for beginning of object key string`),
		},
		"should ignore tags if lb tag override annotation is used for nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                        "nlb",
						ServiceAnnotationLoadBalancerInitialFreeformTagsOverride: `{'test':'tag'}`,
						ServiceAnnotationLoadBalancerInitialDefinedTagsOverride:  `{"namespace":{"key":"value", "owner":"team"}}`,
					},
				},
			},
			initialTags:   &emptyInitialTags,
			desiredLBTags: &emptyTags,
			err:           nil,
		},
		"only resource level freeform tags for nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                               "nlb",
						ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride: `{"test":"tag"}`,
					},
				},
			},
			initialTags: &emptyInitialTags,
			desiredLBTags: &providercfg.TagConfig{
				FreeformTags: map[string]string{"test": "tag"},
				// Defined tags are always present as Oracle-Tags are added by default
			},
			err: nil,
		},
		"only resource level defined tags for nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride: `{"namespace":{"key":"value"}}`,
					},
				},
			},
			initialTags: &emptyInitialTags,
			desiredLBTags: &providercfg.TagConfig{
				DefinedTags: map[string]map[string]interface{}{"namespace": {"key": "value"}},
			},
			err: nil,
		},
		"resource and cluster level tags, only resource level tags are added for nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                               "nlb",
						ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride: `{"cluster":"resource", "unique":"tag"}`,
						ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride:  `{"namespace":{"key":"value", "owner":"team"}}`,
					},
				},
			},
			initialTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					FreeformTags: map[string]string{"cluster": "cluster"},
					DefinedTags:  map[string]map[string]interface{}{"namespace": {"cluster": "name", "owner": "cluster"}},
				},
			},
			desiredLBTags: &providercfg.TagConfig{
				FreeformTags: map[string]string{"cluster": "resource", "unique": "tag"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
			},
			err: nil,
		},
		"reverse compatibility tags test for nlb 1": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                               "nlb",
						ServiceAnnotationNetworkLoadBalancerFreeformTags:                `{"cluster":"resource1", "unique":"tag1"}`,
						ServiceAnnotationNetworkLoadBalancerDefinedTags:                 `{"namespace":{"key":"value1", "owner":"team1"}}`,
						ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride: `{"cluster":"resource", "unique":"tag"}`,
						ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride:  `{"namespace":{"key":"value", "owner":"team"}}`,
					},
				},
			},
			desiredLBTags: &providercfg.TagConfig{
				FreeformTags: map[string]string{"cluster": "resource", "unique": "tag"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
			},
			err: nil,
		},
		"reverse compatibility tags test for nlb 2": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                "nlb",
						ServiceAnnotationNetworkLoadBalancerFreeformTags: `{"cluster":"resource1", "unique":"tag1"}`,
						ServiceAnnotationNetworkLoadBalancerDefinedTags:  `{"namespace":{"key":"value1", "owner":"team1"}}`,
					},
				},
			},
			desiredLBTags: &providercfg.TagConfig{
				FreeformTags: map[string]string{"cluster": "resource1", "unique": "tag1"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team1", "key": "value1"}},
			},
			err: nil,
		},
		"reverse compatibility tags test for nlb 3": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerFreeformTags:               `{"cluster":"resource1", "unique":"tag1"}`,
						ServiceAnnotationNetworkLoadBalancerDefinedTags:                `{"namespace":{"key":"value1", "owner":"team1"}}`,
						ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride: `{"namespace":{"key":"value", "owner":"team"}}`,
					},
				},
			},
			desiredLBTags: &providercfg.TagConfig{
				FreeformTags: map[string]string{"cluster": "resource1", "unique": "tag1"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
			},
			err: nil,
		},
		"reverse compatibility tags test for nlb 4": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                               "nlb",
						ServiceAnnotationNetworkLoadBalancerFreeformTags:                `{"cluster":"resource1", "unique":"tag1"}`,
						ServiceAnnotationNetworkLoadBalancerDefinedTags:                 `{"namespace":{"key":"value1", "owner":"team1"}}`,
						ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride: `{"cluster":"resource", "unique":"tag"}`,
					},
				},
			},
			desiredLBTags: &providercfg.TagConfig{
				FreeformTags: map[string]string{"cluster": "resource", "unique": "tag"},
				DefinedTags:  map[string]map[string]interface{}{"namespace": {"owner": "team1", "key": "value1"}},
			},
			err: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actualTags, err := getLoadBalancerTags(tc.service, tc.initialTags)
			t.Log("Error:", err)
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Expected error\n%+v\nbut got\n%+v", tc.err, err)
			}
			if !reflect.DeepEqual(tc.desiredLBTags, actualTags) {
				t.Errorf("Expected LB Tags\n%+v\nbut got\n%+v", tc.desiredLBTags, actualTags)
			}
		})
	}
}

func Test_getHealthChecker(t *testing.T) {
	testCases := map[string]struct {
		service  *v1.Service
		expected *client.GenericHealthChecker
		err      error
	}{
		"defaults": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expected: &client.GenericHealthChecker{
				Protocol:         "HTTP",
				IsForcePlainText: common.Bool(false),
				Port:             common.Int(10256),
				UrlPath:          common.String("/healthz"),
				Retries:          common.Int(3),
				TimeoutInMillis:  common.Int(3000),
				IntervalInMillis: common.Int(10000),
				ReturnCode:       common.Int(http.StatusOK),
			},
			err: nil,
		},
		"retries timeout intervals annotations for lb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerHealthCheckTimeout:  "3500",
						ServiceAnnotationLoadBalancerHealthCheckRetries:  "4",
						ServiceAnnotationLoadBalancerHealthCheckInterval: "14500",
					},
				},
			},
			expected: &client.GenericHealthChecker{
				Protocol:         "HTTP",
				IsForcePlainText: common.Bool(false),
				Port:             common.Int(10256),
				UrlPath:          common.String("/healthz"),
				Retries:          common.Int(4),
				TimeoutInMillis:  common.Int(3500),
				IntervalInMillis: common.Int(14500),
				ReturnCode:       common.Int(http.StatusOK),
			},
			err: nil,
		},
		"defaults-nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expected: &client.GenericHealthChecker{
				Protocol:         "HTTP",
				IsForcePlainText: common.Bool(false),
				Port:             common.Int(10256),
				UrlPath:          common.String("/healthz"),
				Retries:          common.Int(3),
				TimeoutInMillis:  common.Int(3000),
				IntervalInMillis: common.Int(10000),
				ReturnCode:       common.Int(http.StatusOK),
			},
			err: nil,
		},
		"retries timeout intervals annotations for nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                       "nlb",
						ServiceAnnotationNetworkLoadBalancerHealthCheckTimeout:  "3500",
						ServiceAnnotationNetworkLoadBalancerHealthCheckRetries:  "4",
						ServiceAnnotationNetworkLoadBalancerHealthCheckInterval: "14500",
					},
				},
			},
			expected: &client.GenericHealthChecker{
				Protocol:         "HTTP",
				IsForcePlainText: common.Bool(false),
				Port:             common.Int(10256),
				UrlPath:          common.String("/healthz"),
				Retries:          common.Int(4),
				TimeoutInMillis:  common.Int(3500),
				IntervalInMillis: common.Int(14500),
				ReturnCode:       common.Int(http.StatusOK),
			},
			err: nil,
		},
		"lb wrong interval value - lesser than min": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerHealthCheckInterval: "300",
					},
				},
			},
			expected: nil,
			err:      fmt.Errorf("invalid value for health check interval, should be between %v and %v", LBHealthCheckIntervalMin, LBHealthCheckIntervalMax),
		},
		"lb wrong interval value - greater than max": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerHealthCheckInterval: "3000000",
					},
				},
			},
			expected: nil,
			err:      fmt.Errorf("invalid value for health check interval, should be between %v and %v", LBHealthCheckIntervalMin, LBHealthCheckIntervalMax),
		},
		"nlb wrong interval value - lesser than min": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                       "nlb",
						ServiceAnnotationNetworkLoadBalancerHealthCheckInterval: "3000",
					},
				},
			},
			expected: nil,
			err:      fmt.Errorf("invalid value for health check interval, should be between %v and %v", NLBHealthCheckIntervalMin, NLBHealthCheckIntervalMax),
		},
		"nlb wrong interval value - greater than max": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                       "nlb",
						ServiceAnnotationNetworkLoadBalancerHealthCheckInterval: "3000000",
					},
				},
			},
			expected: nil,
			err:      fmt.Errorf("invalid value for health check interval, should be between %v and %v", NLBHealthCheckIntervalMin, NLBHealthCheckIntervalMax),
		},
		"http healthcheck for https backends": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                "lb",
						ServiceAnnotationLoadBalancerTLSBackendSetSecret: "testSecret",
					},
				},
			},
			expected: &client.GenericHealthChecker{
				Protocol:         "HTTP",
				IsForcePlainText: common.Bool(true),
				Port:             common.Int(10256),
				UrlPath:          common.String("/healthz"),
				Retries:          common.Int(3),
				TimeoutInMillis:  common.Int(3000),
				IntervalInMillis: common.Int(10000),
				ReturnCode:       common.Int(http.StatusOK),
			},
			err: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := getHealthChecker(tc.service)

			if tc.err != nil && err == nil {
				t.Errorf("Error: expected\n%+v\nbut got\n%+v", tc.err, err)
			}
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Error: expected\n%+v\nbut got\n%+v", tc.err, err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.expected, result)
			}
		})
	}
}

func Test_getListeners(t *testing.T) {
	var tests = []struct {
		service                  *v1.Service
		listenerBackendIpVersion []string
		name                     string
		sslConfig                *SSLConfig
		want                     map[string]client.GenericListener
	}{
		{
			name: "default",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			sslConfig:                nil,
			want: map[string]client.GenericListener{
				"TCP-80": {
					Name:                  common.String("TCP-80"),
					Port:                  common.Int(80),
					Protocol:              common.String("TCP"),
					DefaultBackendSetName: common.String("TCP-80"),
				},
			},
		},
		{
			name: "default-nlb",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			sslConfig:                nil,
			want: map[string]client.GenericListener{
				"TCP-80": {
					Name:                  common.String("TCP-80"),
					Port:                  common.Int(80),
					Protocol:              common.String("TCP"),
					DefaultBackendSetName: common.String("TCP-80"),
					IpVersion:             GenericIpVersion(client.GenericIPv4),
				},
			},
		},
		{
			name: "ssl configuration and cipher suite",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.Protocol("TCP"),
							Port:     int32(443),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadbalancerListenerSSLConfig: `{"cipherSuiteName":"oci-default-http2-ssl-cipher-suite-v1", "protocols":["TLSv1.2"]}`,
						ServiceAnnotationLoadBalancerSSLPorts:          "443",
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			sslConfig: &SSLConfig{
				Ports:                   sets.NewInt(443),
				ListenerSSLSecretName:   listenerSecret,
				BackendSetSSLSecretName: backendSecret,
			},
			want: map[string]client.GenericListener{
				"TCP-443": {
					Name:                  common.String("TCP-443"),
					Port:                  common.Int(443),
					Protocol:              common.String("TCP"),
					DefaultBackendSetName: common.String("TCP-443"),
					SslConfiguration: &client.GenericSslConfigurationDetails{
						CertificateName:       &listenerSecret,
						VerifyDepth:           common.Int(0),
						VerifyPeerCertificate: common.Bool(false),
						CipherSuiteName:       common.String("oci-default-http2-ssl-cipher-suite-v1"),
						Protocols:             []string{"TLSv1.2"},
					},
				},
			},
		},
		{
			name: "Listeners with ssl configuration information",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadbalancerListenerSSLConfig: `{"cipherSuiteName":"oci-default-http2-ssl-cipher-suite-v1", "protocols":["TLSv1.2"]}`,
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			sslConfig:                nil,
			want: map[string]client.GenericListener{
				"TCP-80": {
					Name:                  common.String("TCP-80"),
					Port:                  common.Int(80),
					Protocol:              common.String("TCP"),
					DefaultBackendSetName: common.String("TCP-80"),
				},
			},
		},
		{
			name: "grpc protocol no ssl",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.Protocol("GRPC"),
							Port:     int32(80),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			sslConfig:                nil,
			listenerBackendIpVersion: []string{IPv4},
			want:                     nil,
		},
		{
			name: "grpc protocol with ssl configuration and smart default cipher suite",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.Protocol("TCP"),
							Port:     int32(443),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerBEProtocol: ProtocolGrpc,
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			sslConfig: &SSLConfig{
				Ports:                   sets.NewInt(443),
				ListenerSSLSecretName:   listenerSecret,
				BackendSetSSLSecretName: backendSecret,
			},
			want: map[string]client.GenericListener{
				"GRPC-443": {
					Name:                  common.String("GRPC-443"),
					Port:                  common.Int(443),
					Protocol:              common.String("GRPC"),
					DefaultBackendSetName: common.String("TCP-443"),
					SslConfiguration: &client.GenericSslConfigurationDetails{
						CertificateName:       &listenerSecret,
						VerifyDepth:           common.Int(0),
						VerifyPeerCertificate: common.Bool(false),
						CipherSuiteName:       common.String(DefaultCipherSuiteForGRPC),
					},
				},
			},
		},
		{
			name: "grpc protocol with ssl configuration and cipher suite",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.Protocol("TCP"),
							Port:     int32(443),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerBEProtocol:        ProtocolGrpc,
						ServiceAnnotationLoadbalancerListenerSSLConfig: `{"cipherSuiteName":"oci-default-http2-ssl-cipher-suite-v1", "protocols":["TLSv1.2"]}`,
						ServiceAnnotationLoadBalancerSSLPorts:          "443",
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			sslConfig: &SSLConfig{
				Ports:                   sets.NewInt(443),
				ListenerSSLSecretName:   listenerSecret,
				BackendSetSSLSecretName: backendSecret,
			},
			want: map[string]client.GenericListener{
				"GRPC-443": {
					Name:                  common.String("GRPC-443"),
					Port:                  common.Int(443),
					Protocol:              common.String("GRPC"),
					DefaultBackendSetName: common.String("TCP-443"),
					SslConfiguration: &client.GenericSslConfigurationDetails{
						CertificateName:       &listenerSecret,
						VerifyDepth:           common.Int(0),
						VerifyPeerCertificate: common.Bool(false),
						CipherSuiteName:       common.String("oci-default-http2-ssl-cipher-suite-v1"),
						Protocols:             []string{"TLSv1.2"},
					},
				},
			},
		},
		{
			name: "Listeners with cipher suites",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(80),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadbalancerListenerSSLConfig: `{"cipherSuiteName":"oci-default-http2-ssl-cipher-suite-v1", "protocols":["TLSv1.2"]}`,
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			sslConfig:                nil,
			want: map[string]client.GenericListener{
				"TCP-80": {
					Name:                  common.String("TCP-80"),
					Port:                  common.Int(80),
					Protocol:              common.String("TCP"),
					DefaultBackendSetName: common.String("TCP-80"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.service
			if got, err := getListeners(svc, tt.sslConfig, tt.listenerBackendIpVersion); !reflect.DeepEqual(got, tt.want) {
				if err != nil {
					t.Errorf("Err %v", err.Error())
				}
				got, _ := json.Marshal(got)
				want, _ := json.Marshal(tt.want)
				t.Errorf("getListeners() failed want: %s \n got: %s \n", want, got)
			}
		})
	}
}

func Test_getSecurityListManagementMode(t *testing.T) {
	testCases := map[string]struct {
		service  *v1.Service
		expected string
	}{
		"defaults - lb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expected: "All",
		},
		"defaults - nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expected: "None",
		},
		"lb mode None": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityListManagementMode: "None",
					},
				},
			},
			expected: "None",
		},
		"lb mode all": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityListManagementMode: "All",
					},
				},
			},
			expected: "All",
		},
		"lb mode frontend": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityListManagementMode: "Frontend",
					},
				},
			},
			expected: "Frontend",
		},
		"defaults-nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expected: "None",
		},
		"nlb mode None": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "None",
					},
				},
			},
			expected: "None",
		},
		"nlb mode all": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
					},
				},
			},
			expected: "All",
		},
		"nlb mode frontend": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "Frontend",
					},
				},
			},
			expected: "Frontend",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := getSecurityListManagementMode(tc.service)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected Security List Mode \n%+v\nbut got\n%+v", tc.expected, result)
			}
		})
	}
}

func Test_getRuleManagementMode(t *testing.T) {
	testCases := map[string]struct {
		service  *v1.Service
		expected string
		nsg      *ManagedNetworkSecurityGroup
		error    error
	}{
		"defaults": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expected: "All",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"defaults - nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expected: "None",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"lb mode None": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "None",
					},
				},
			},
			expected: "None",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"lb mode all": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "SL-All",
					},
				},
			},
			expected: "All",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"lb mode frontend": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "SL-Frontend",
					},
				},
			},
			expected: "Frontend",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"lb mode nsg frontend": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "NSG",
					},
				},
			},
			expected: "NSG",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: "NSG",
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"defaults-nlb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expected: "None",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"nlb mode None": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                       "nlb",
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "None",
					},
				},
			},
			expected: "None",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"nlb mode all": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                       "nlb",
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "SL-All",
					},
				},
			},
			expected: "All",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"nlb mode frontend": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                       "nlb",
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "SL-Frontend",
					},
				},
			},
			expected: "Frontend",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"nlb mode nsg": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                       "nlb",
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "NSG",
					},
				},
			},
			expected: "NSG",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: "NSG",
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"lb mode precedence": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSecurityListManagementMode: "All",
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "NSG",
					},
				},
			},
			expected: "NSG",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: "NSG",
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"nlb mode precedence": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode:        "NSG",
					},
				},
			},
			expected: "NSG",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: "NSG",
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"case does not matter nsg": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode:        "nsg",
					},
				},
			},
			expected: "NSG",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: "NSG",
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"case does not matter sl-all": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode:        "sl-all",
					},
				},
			},
			expected: "All",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"case does not matter sl-frontend": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode:        "sl-frontend",
					},
				},
			},
			expected: "Frontend",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
		},
		"invalid values should return none": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "random",
						ServiceAnnotationLoadBalancerSecurityRuleManagementMode:        "random",
					},
				},
			},
			expected: "None",
			nsg: &ManagedNetworkSecurityGroup{
				nsgRuleManagementMode: ManagementModeNone,
				frontendNsgId:         "",
				backendNsgId:          []string{},
			},
			error: fmt.Errorf("invalid value: %s provided for annotation: oci.oraclecloud.com/security-rule-management-mode",
				"random"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, nsg, err := getRuleManagementMode(tc.service)
			if err != nil {
				if !reflect.DeepEqual(err, tc.error) {
					t.Errorf("Expected Security List Mode \n%+v\nbut got\n%+v", tc.error, err)
				}
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected Security List Mode \n%+v\nbut got\n%+v", tc.expected, result)
			}
			if !reflect.DeepEqual(nsg, tc.nsg) {
				t.Errorf("Expected Nsg values \n%+v\nbut got\n%+v", tc.nsg, nsg)
			}
		})
	}
}

func Test_getBackendNetworkSecurityGroups(t *testing.T) {
	testCases := map[string]struct {
		service *v1.Service
		nsgList []string
		err     error
	}{
		"empty ServiceAnnotationLoadBalancerNetworkSecurityGroups annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationBackendSecurityRuleManagement: "",
					},
				},
			},
			nsgList: []string{},
			err:     nil,
		},
		"no ServiceAnnotationBackendSecurityRuleManagement annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			nsgList: []string{},
			err:     nil,
		},
		"ServiceAnnotationBackendSecurityRuleManagement update annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationBackendSecurityRuleManagement: "ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					},
				},
			},
			nsgList: []string{"ocid1.networksecuritygroup.oc1.iad.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			err:     nil,
		},
		"ServiceAnnotationBackendSecurityRuleManagement more than 5": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationBackendSecurityRuleManagement: "ocid1,ocid2,ocid3,ocid4,ocid5,ocid6",
					},
				},
			},
			nsgList: []string{"ocid1", "ocid2", "ocid3", "ocid4", "ocid5", "ocid6"},
			err:     nil,
		},
		"ServiceAnnotationBackendSecurityRuleManagement duplicate NSG OCIDS": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationBackendSecurityRuleManagement: "ocid1,ocid2, ocid1",
					},
				},
			},
			nsgList: []string{"ocid1", "ocid2"},
			err:     nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			nsgList, err := getManagedBackendNSG(tc.service)
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Expected  NSG List error\n%+v\nbut got\n%+v", tc.err, err)
			}
			if !reflect.DeepEqual(nsgList, tc.nsgList) {
				t.Errorf("Expected NSG List\n%+v\nbut got\n%+v", tc.nsgList, nsgList)
			}
		})
	}
}

func Test_validateService(t *testing.T) {
	testCases := map[string]struct {
		service *v1.Service
		err     error
	}{
		"defaults": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			err: nil,
		},
		"nlb invalid seclist mgmt mode": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "Neither",
					},
				},
			},
			err: fmt.Errorf("invalid value: Neither provided for annotation: oci-network-load-balancer.oraclecloud.com/security-list-management-mode"),
		},
		"lb with protocol udp": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolUDP,
							Port:     int32(67),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			err: fmt.Errorf("OCI load balancers do not support UDP"),
		},
		"nlb udp with seclist mgmt not None": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolUDP,
							Port:     int32(67),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
					},
				},
			},
			err: fmt.Errorf("Security list management mode can only be 'None' for UDP protocol"),
		},
		"session affinity not none": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityClientIP,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolUDP,
							Port:     int32(67),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                              "nlb",
						ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "None",
					},
				},
			},
			err: fmt.Errorf("OCI only supports SessionAffinity \"None\" currently"),
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := validateService(tc.service)
			if tc.err != nil && err == nil {
				t.Errorf("Expected  \n%+v\nbut got\n%+v", tc.err, err)
			}
			if err != nil && tc.err == nil {
				t.Errorf("Error: expected\n%+v\nbut got\n%+v", tc.err, err)
			}
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.err, err)
			}
		})
	}
}

func Test_getListenersNetworkLoadBalancer(t *testing.T) {
	testOneListenerName := "TCP_AND_UDP-67"
	testOneBackendSetName := "TCP_AND_UDP-67"
	testOneProtocol := "TCP_AND_UDP"
	testOnePort := 67

	testTwoListenerNameOne := "TCP-67"
	testTwoBackendSetNameOne := "TCP-67"
	testTwoProtocolOne := "TCP"
	testTwoPortOne := 67

	testTwoListenerNameTwo := "UDP-68"
	testTwoBackendSetNameTwo := "UDP-68"
	testTwoProtocolTwo := "UDP"
	testTwoPortTwo := 68

	testThreeListenerName := "TCP-67"
	testThreeBackendSetName := "TCP-67"
	testThreeProtocol := "TCP"
	testThreePort := 67

	testFourListenerName := "UDP-67"
	testFourBackendSetName := "UDP-67"
	testFourProtocol := "UDP"
	testFourPort := 67

	IPFamilyPolicyPreferDualStack := v1.IPFamilyPolicyPreferDualStack
	IPFamilyPolicySingleStack := v1.IPFamilyPolicySingleStack
	testThreeListenerNameIPv6 := "TCP-67-IPv6"
	testThreeBackendSetNameIPv6 := "TCP-67-IPv6"

	testCases := map[string]struct {
		service                  *v1.Service
		listenerBackendIpVersion []string
		wantListeners            map[string]client.GenericListener
		err                      error
	}{
		"NLB_with_mixed_protocol_on_same_port": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
						},
						{
							Protocol: v1.ProtocolUDP,
							Port:     int32(67),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			wantListeners: map[string]client.GenericListener{
				"TCP_AND_UDP-67": {
					Name:                  &testOneListenerName,
					DefaultBackendSetName: common.String(testOneBackendSetName),
					Protocol:              &testOneProtocol,
					Port:                  &testOnePort,
				},
			},
			err: nil,
		},
		"NLB_with_mixed_protocol_on_different_port": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
						},
						{
							Protocol: v1.ProtocolUDP,
							Port:     int32(68),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			wantListeners: map[string]client.GenericListener{
				"TCP-67": {
					Name:                  &testTwoListenerNameOne,
					DefaultBackendSetName: common.String(testTwoBackendSetNameOne),
					Protocol:              &testTwoProtocolOne,
					Port:                  &testTwoPortOne,
				},
				"UDP-68": {
					Name:                  &testTwoListenerNameTwo,
					DefaultBackendSetName: common.String(testTwoBackendSetNameTwo),
					Protocol:              &testTwoProtocolTwo,
					Port:                  &testTwoPortTwo,
				},
			},
			err: nil,
		},
		"NLB_with_only_TCP_protocol": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			wantListeners: map[string]client.GenericListener{
				"TCP-67": {
					Name:                  &testThreeListenerName,
					DefaultBackendSetName: common.String(testThreeBackendSetName),
					Protocol:              &testThreeProtocol,
					Port:                  &testThreePort,
				},
			},
			err: nil,
		},
		"NLB_with_only_UDP_protocol": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolUDP,
							Port:     int32(67),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			wantListeners: map[string]client.GenericListener{
				"UDP-67": {
					Name:                  &testFourListenerName,
					DefaultBackendSetName: common.String(testFourBackendSetName),
					Protocol:              &testFourProtocol,
					Port:                  &testFourPort,
				},
			},
			err: nil,
		},
		"NLB_with_only_TCP_protocol_IPv4_IPv6": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
						},
					},
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy: &IPFamilyPolicyPreferDualStack,
				},
			},
			listenerBackendIpVersion: []string{IPv4, IPv6},
			wantListeners: map[string]client.GenericListener{
				"TCP-67-IPv6": {
					Name:                  &testThreeListenerNameIPv6,
					DefaultBackendSetName: common.String(testThreeBackendSetNameIPv6),
					Protocol:              &testThreeProtocol,
					Port:                  &testThreePort,
				},
				"TCP-67": {
					Name:                  &testThreeListenerName,
					DefaultBackendSetName: common.String(testThreeBackendSetName),
					Protocol:              &testThreeProtocol,
					Port:                  &testThreePort,
				},
			},
			err: nil,
		},
		"NLB_with_only_TCP_protocol_IPv6": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
						},
					},
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv6)},
					IPFamilyPolicy: &IPFamilyPolicySingleStack,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			listenerBackendIpVersion: []string{IPv6},
			wantListeners: map[string]client.GenericListener{
				"TCP-67-IPv6": {
					Name:                  &testThreeListenerNameIPv6,
					DefaultBackendSetName: common.String(testThreeBackendSetNameIPv6),
					Protocol:              &testThreeProtocol,
					Port:                  &testThreePort,
				},
			},
			err: nil,
		},
		"NLB_with_Ppv2_Enabled": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
						},
						{
							Protocol: v1.ProtocolUDP,
							Port:     int32(68),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                 "nlb",
						ServiceAnnotationNetworkLoadBalancerIsPpv2Enabled: "true",
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			wantListeners: map[string]client.GenericListener{
				"TCP-67": {
					Name:                  &testTwoListenerNameOne,
					DefaultBackendSetName: common.String(testTwoBackendSetNameOne),
					Protocol:              &testTwoProtocolOne,
					Port:                  &testTwoPortOne,
					IsPpv2Enabled:         pointer.Bool(true),
				},
				"UDP-68": {
					Name:                  &testTwoListenerNameTwo,
					DefaultBackendSetName: common.String(testTwoBackendSetNameTwo),
					Protocol:              &testTwoProtocolTwo,
					Port:                  &testTwoPortTwo,
					IsPpv2Enabled:         pointer.Bool(true),
				},
			},
			err: nil,
		},

		"NLB_with_Ppv2_Disabled": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
						},
						{
							Protocol: v1.ProtocolUDP,
							Port:     int32(68),
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                 "nlb",
						ServiceAnnotationNetworkLoadBalancerIsPpv2Enabled: "xyz",
					},
				},
			},
			listenerBackendIpVersion: []string{IPv4},
			wantListeners: map[string]client.GenericListener{
				"TCP-67": {
					Name:                  &testTwoListenerNameOne,
					DefaultBackendSetName: common.String(testTwoBackendSetNameOne),
					Protocol:              &testTwoProtocolOne,
					Port:                  &testTwoPortOne,
					IsPpv2Enabled:         pointer.Bool(true),
				},
				"UDP-68": {
					Name:                  &testTwoListenerNameTwo,
					DefaultBackendSetName: common.String(testTwoBackendSetNameTwo),
					Protocol:              &testTwoProtocolTwo,
					Port:                  &testTwoPortTwo,
					IsPpv2Enabled:         pointer.Bool(false),
				},
			},
			err: nil,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			gotListeners, err := getListenersNetworkLoadBalancer(tc.service, tc.listenerBackendIpVersion)
			if tc.err != nil && err == nil {
				t.Errorf("Expected  \n%+v\nbut got\n%+v", tc.err, err)
			}
			if err != nil && tc.err == nil {
				t.Errorf("Error: expected\n%+v\nbut got\n%+v", tc.err, err)
			}
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.err, err)
			}
			if len(gotListeners) != len(tc.wantListeners) {
				t.Errorf("Number of excpected listeners \n%+v\nbut got\n%+v", len(tc.wantListeners), len(gotListeners))
			}
			if len(gotListeners) != 0 {
				for name, listener := range tc.wantListeners {
					gotListener, ok := gotListeners[name]
					if !ok {
						t.Errorf("Expected listener with name \n%+v\nbut listener not present", *listener.Name)
					}
					if *gotListener.Name != *listener.Name {
						t.Errorf("Expected listener name \n%+v\nbut got listener name \n%+v", *listener.Name, *gotListener.Name)
					}
					if *gotListener.DefaultBackendSetName != *listener.DefaultBackendSetName {
						t.Errorf("Expected default backend set name \n%+v\nbut got default backend set name \n%+v", *listener.DefaultBackendSetName, *gotListener.DefaultBackendSetName)
					}
					if *gotListener.Protocol != *listener.Protocol {
						t.Errorf("Expected protocol \n%+v\nbut got protocol \n%+v", *listener.Protocol, *gotListener.Protocol)
					}
					if *gotListener.Port != *listener.Port {
						t.Errorf("Expected port number \n%+v\nbut got port number \n%+v", *listener.Port, *gotListener.Port)
					}
				}
			}
		})
	}
}

func Test_getPreserveSourceDestination(t *testing.T) {
	testCases := map[string]struct {
		service      *v1.Service
		expectedBool bool
		err          error
	}{
		"oci LB default": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedBool: false,
			err:          nil,
		},
		"oci LB, externalTrafficPolicy Local": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity:       v1.ServiceAffinityNone,
					ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeLocal,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedBool: false,
			err:          nil,
		},
		"oci NLB default": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expectedBool: false,
			err:          nil,
		},
		"oci NLB, externalTrafficPolicy Local": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity:       v1.ServiceAffinityNone,
					ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeLocal,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expectedBool: true,
			err:          nil,
		},
		"oci NLB, externalTrafficPolicy Local, disabled by annotation": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity:       v1.ServiceAffinityNone,
					ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeLocal,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                    "nlb",
						ServiceAnnotationNetworkLoadBalancerIsPreserveSource: "false",
					},
				},
			},
			expectedBool: false,
			err:          nil,
		},
		"oci NLB, externalTrafficPolicy Local, enabled via annotation": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity:       v1.ServiceAffinityNone,
					ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeLocal,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                    "nlb",
						ServiceAnnotationNetworkLoadBalancerIsPreserveSource: "true",
					},
				},
			},
			expectedBool: true,
			err:          nil,
		},
		"oci NLB, externalTrafficPolicy Local, bad annotation value": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity:       v1.ServiceAffinityNone,
					ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeLocal,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                    "nlb",
						ServiceAnnotationNetworkLoadBalancerIsPreserveSource: "disable",
					},
				},
			},
			expectedBool: false,
			err:          fmt.Errorf("failed to to parse oci-network-load-balancer.oraclecloud.com/is-preserve-source annotation value - disable"),
		},
		"oci NLB, externalTrafficPolicy Cluster, enabled via annotation": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity:       v1.ServiceAffinityNone,
					ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeCluster,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                    "nlb",
						ServiceAnnotationNetworkLoadBalancerIsPreserveSource: "true",
					},
				},
			},
			expectedBool: false,
			err:          fmt.Errorf("oci-network-load-balancer.oraclecloud.com/is-preserve-source annotation cannot be set when externalTrafficPolicy is set to Cluster"),
		},
		"oci NLB, externalTrafficPolicy Cluster, disabled via annotation": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity:       v1.ServiceAffinityNone,
					ExternalTrafficPolicy: v1.ServiceExternalTrafficPolicyTypeCluster,
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:                    "nlb",
						ServiceAnnotationNetworkLoadBalancerIsPreserveSource: "false",
					},
				},
			},
			expectedBool: false,
			err:          fmt.Errorf("oci-network-load-balancer.oraclecloud.com/is-preserve-source annotation cannot be set when externalTrafficPolicy is set to Cluster"),
		},
	}
	for name, tc := range testCases {
		logger := zap.L()
		t.Run(name, func(t *testing.T) {
			enable, err := getPreserveSource(logger.Sugar(), tc.service)
			if tc.err != nil && err == nil {
				t.Errorf("Expected  \n%+v\nbut got\n%+v", tc.err, err)
			}
			if err != nil && tc.err == nil {
				t.Errorf("Error: expected\n%+v\nbut got\n%+v", tc.err, err)
			}
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.err, err)
			}
			if enable != tc.expectedBool {
				t.Errorf("Expected  \n%+v\nbut got\n%+v", tc.expectedBool, enable)
			}

		})
	}
}

var getLBShapeTestCases = []struct {
	name                 string
	existingLb           *client.GenericLoadBalancer
	service              *v1.Service
	expectedShape        string
	expectedMinBandwidth int
	expectedMaxBandwidth int
	expectedError        error
}{
	{
		"default spec, no existing LB",
		nil,
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   "kube-system",
				Name:        "testservice",
				UID:         "test-uid",
				Annotations: map[string]string{},
			},
		},
		"100Mbps",
		0,
		0,
		nil,
	},
	{
		"flexible spec, no existing LB",
		nil,
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "testservice",
				UID:       "test-uid",
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerShape:        "flexible",
					ServiceAnnotationLoadBalancerShapeFlexMin: "1",
					ServiceAnnotationLoadBalancerShapeFlexMax: "10000000",
				},
			},
		},
		"flexible",
		10,
		8192,
		nil,
	},
	{
		"default shape in spec, existing LB converted to flexible",
		&client.GenericLoadBalancer{
			ShapeName: common.String("flexible"),
			ShapeDetails: &client.GenericShapeDetails{
				MinimumBandwidthInMbps: common.Int(12),
				MaximumBandwidthInMbps: common.Int(13),
			},
		},
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   "kube-system",
				Name:        "testservice",
				UID:         "test-uid",
				Annotations: map[string]string{},
			},
		},
		"flexible",
		12,
		13,
		nil,
	},
	{
		"bad flexible spec",
		nil,
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "testservice",
				UID:       "test-uid",
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerShape:        "flexible",
					ServiceAnnotationLoadBalancerShapeFlexMin: "1AB",
					ServiceAnnotationLoadBalancerShapeFlexMax: "2AB",
				},
			},
		},
		"",
		10,
		8192,
		errors.New("invalid format for service.beta.kubernetes.io/oci-load-balancer-shape-flex-min annotation : 1AB"),
	},
	{
		"bad flexible max bandwidth",
		nil,
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "testservice",
				UID:       "test-uid",
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerShape:        "flexible",
					ServiceAnnotationLoadBalancerShapeFlexMin: "10",
					ServiceAnnotationLoadBalancerShapeFlexMax: "2AB",
				},
			},
		},
		"",
		10,
		0,
		errors.New("invalid format for service.beta.kubernetes.io/oci-load-balancer-shape-flex-max annotation : 2AB"),
	},
	{
		"flexible max bandwidth lower than min bandwidth",
		nil,
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "testservice",
				UID:       "test-uid",
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerShape:        "flexible",
					ServiceAnnotationLoadBalancerShapeFlexMin: "100",
					ServiceAnnotationLoadBalancerShapeFlexMax: "10",
				},
			},
		},
		"flexible",
		100,
		100,
		nil,
	},
	{
		"bad flexible min and max bandwidth",
		nil,
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "testservice",
				UID:       "test-uid",
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerShape:        "flexible",
					ServiceAnnotationLoadBalancerShapeFlexMin: "100000",
					ServiceAnnotationLoadBalancerShapeFlexMax: "1",
				},
			},
		},
		"flexible",
		8192,
		8192,
		nil,
	},
	{
		"existing LB converted to flex outside of OKE",
		&client.GenericLoadBalancer{
			ShapeName: common.String("flexible"),
			ShapeDetails: &client.GenericShapeDetails{
				MinimumBandwidthInMbps: common.Int(10),
				MaximumBandwidthInMbps: common.Int(100),
			},
		},
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   "kube-system",
				Name:        "testservice",
				UID:         "test-uid",
				Annotations: map[string]string{},
			},
		},
		"flexible",
		10,
		100,
		nil,
	},
	{
		"existing LB converted to flex outside of OKE, but dynamic shape annotation still present",
		&client.GenericLoadBalancer{
			ShapeName: common.String("flexible"),
			ShapeDetails: &client.GenericShapeDetails{
				MinimumBandwidthInMbps: common.Int(10),
				MaximumBandwidthInMbps: common.Int(100),
			},
		},
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "testservice",
				UID:       "test-uid",
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerShape: "100Mbps",
				},
			},
		},
		"100Mbps",
		0,
		0,
		nil,
	},
	{
		"existing LB converted to flex outside of OKE, but flexible annotations have different value",
		&client.GenericLoadBalancer{
			ShapeName: common.String("flexible"),
			ShapeDetails: &client.GenericShapeDetails{
				MinimumBandwidthInMbps: common.Int(10),
				MaximumBandwidthInMbps: common.Int(100),
			},
		},
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "testservice",
				UID:       "test-uid",
				Annotations: map[string]string{
					ServiceAnnotationLoadBalancerShape:        "flexible",
					ServiceAnnotationLoadBalancerShapeFlexMin: "100",
					ServiceAnnotationLoadBalancerShapeFlexMax: "200",
				},
			},
		},
		"flexible",
		100,
		200,
		nil,
	},
}

func Test_getLBShape(t *testing.T) {
	for _, tc := range getLBShapeTestCases {
		actualShapeName, minBandwidth, maxBandwidth, err := getLBShape(tc.service, tc.existingLb)
		if actualShapeName != tc.expectedShape {
			t.Errorf("Expected  \n%+v\nbut got\n%+v", tc.expectedShape, actualShapeName)
		}
		if minBandwidth != nil && *minBandwidth != tc.expectedMinBandwidth {
			t.Errorf("Expected  \n%+v\nbut got\n%+v", tc.expectedMinBandwidth, minBandwidth)
		}
		if maxBandwidth != nil && *maxBandwidth != tc.expectedMaxBandwidth {
			t.Errorf("Expected  \n%+v\nbut got\n%+v", tc.expectedMaxBandwidth, maxBandwidth)
		}
		if err != nil && err.Error() != tc.expectedError.Error() {
			t.Errorf("Expected \n%+v\nbut got\n%+v", tc.expectedError, err)
		}
	}
}

func Test_getBackendSetNamePortMap(t *testing.T) {
	testCases := map[string]struct {
		in  *v1.Service
		out map[string]v1.ServicePort
	}{
		"single port": {
			in: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
			},
		},
		"multiple ports": {
			in: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
						{
							Protocol: v1.ProtocolTCP,
							Port:     81,
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
				"TCP-81": {
					Protocol: v1.ProtocolTCP,
					Port:     81,
				},
			},
		},
		"multiple ports with different protocols": {
			in: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
						{
							Protocol: v1.ProtocolUDP,
							Port:     81,
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
				"UDP-81": {
					Protocol: v1.ProtocolUDP,
					Port:     81,
				},
			},
		},
		"multiple ports with mixed protocols": {
			in: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     80,
						},
						{
							Protocol: v1.ProtocolUDP,
							Port:     81,
						},
						{
							Protocol: v1.ProtocolTCP,
							Port:     82,
						},
						{
							Protocol: v1.ProtocolUDP,
							Port:     82,
						},
					},
				},
			},
			out: map[string]v1.ServicePort{
				"TCP-80": {
					Protocol: v1.ProtocolTCP,
					Port:     80,
				},
				"UDP-81": {
					Protocol: v1.ProtocolUDP,
					Port:     81,
				},
				"TCP_AND_UDP-82": {
					Protocol: v1.ProtocolTCP,
					Port:     82,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ipFamilies := []v1.IPFamily{v1.IPFamily(IPv4)}
			tc.in.Spec.IPFamilies = ipFamilies
			got := getBackendSetNamePortMap(tc.in)
			if !reflect.DeepEqual(got, tc.out) {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.out, got)
			}
		})
	}
}

func Test_getOciLoadBalancerSubnets(t *testing.T) {
	testCases := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		service          *v1.Service
		expectedErrMsg   string
		subnets          []string
	}{
		"empty subnets": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"empty strings for subnets": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"empty string for subnet1 annotation": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "",
						ServiceAnnotationLoadBalancerSubnet2: "annotation-two",
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"default string for cloud config subnet2": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "random",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "",
						ServiceAnnotationLoadBalancerSubnet2: "",
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"regional string for subnet2 annotation": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "",
						ServiceAnnotationLoadBalancerSubnet2: "",
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for creating a load balancer",
		},
		"subnets passed via cloud config": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			subnets: []string{"one", "two"},
		},
		"subnets passed via annotation": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "annotation-one",
						ServiceAnnotationLoadBalancerSubnet2: "annotation-two",
					},
				},
			},
			subnets: []string{"annotation-one", "annotation-two"},
		},
		"regional subnet passed via subnet1 annotation": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "regional-subnet",
						ServiceAnnotationLoadBalancerSubnet2: "annotation-two",
					},
				},
			},
			subnets: []string{"regional-subnet"},
		},
		"regional subnet passed via subnet2 annotation": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet1: "annotation-one",
						ServiceAnnotationLoadBalancerSubnet2: "regional-subnet",
					},
				},
			},
			subnets: []string{"regional-subnet"},
		},
	}
	cp := &CloudProvider{
		client: MockOCIClient{},
		config: &providercfg.Config{CompartmentID: "testCompartment"},
	}

	for name, tc := range testCases {
		logger := zap.L()
		t.Run(name, func(t *testing.T) {
			cp.config = &providercfg.Config{
				LoadBalancer: &providercfg.LoadBalancerConfig{
					Subnet1: tc.defaultSubnetOne,
					Subnet2: tc.defaultSubnetTwo,
				},
			}
			subnets, err := cp.getOciLoadBalancerSubnets(context.Background(), logger.Sugar(), tc.service)
			if !reflect.DeepEqual(subnets, tc.subnets) {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.subnets, subnets)
			}
			if err != nil && err.Error() != tc.expectedErrMsg {
				t.Errorf("Expected error with message %q but got %q", tc.expectedErrMsg, err)
			}
		})
	}
}

func Test_getNetworkLoadbalancerSubnets(t *testing.T) {
	testCases := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		service          *v1.Service
		expectedErrMsg   string
		subnets          []string
	}{
		"empty subnets": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for a network load balancer",
		},
		"empty strings for subnets": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for a network load balancer",
		},
		"empty string for nlb subnet annotation": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:          "nlb",
						ServiceAnnotationNetworkLoadBalancerSubnet: "",
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for a network load balancer",
		},
		"default string for cloud config subnet2": {
			defaultSubnetOne: "",
			defaultSubnetTwo: "random",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			expectedErrMsg: "a subnet must be specified for a network load balancer",
		},
		"subnet for nlb annotation": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType:          "nlb",
						ServiceAnnotationNetworkLoadBalancerSubnet: "annotation-one",
					},
				},
			},
			subnets: []string{"annotation-one"},
		},
		"subnets passed via cloud config": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			subnets: []string{"one"},
		},
		"subnets passed via annotation": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			subnets: []string{"one"},
		},
	}
	cp := &CloudProvider{
		client: MockOCIClient{},
		config: &providercfg.Config{CompartmentID: "testCompartment"},
	}

	for name, tc := range testCases {
		logger := zap.L()
		t.Run(name, func(t *testing.T) {
			cp.config = &providercfg.Config{
				LoadBalancer: &providercfg.LoadBalancerConfig{
					Subnet1: tc.defaultSubnetOne,
					Subnet2: tc.defaultSubnetTwo,
				},
			}
			subnets, err := cp.getNetworkLoadbalancerSubnets(context.Background(), logger.Sugar(), tc.service)
			if !reflect.DeepEqual(subnets, tc.subnets) {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.subnets, subnets)
			}
			if err != nil && err.Error() != tc.expectedErrMsg {
				t.Errorf("Expected error with message %q but got %q", tc.expectedErrMsg, err)
			}
		})
	}
}

func Test_getResourceTrackingSysTagsFromConfig(t *testing.T) {
	tests := map[string]struct {
		initialTags *providercfg.InitialTags
		wantTag     map[string]map[string]interface{}
	}{
		"expect an empty system tag when has no common tags": {
			initialTags: &providercfg.InitialTags{},
			wantTag:     nil,
		},
		"expect an empty system tag when resource tracking tags are not in common tags": {
			initialTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					DefinedTags: map[string]map[string]interface{}{"ns": {"key": "val"}},
				},
				Common: &providercfg.TagConfig{
					DefinedTags: map[string]map[string]interface{}{"orcl-not-a-tracking-tag": {"Cluster": "ocid1.cluster.aa..."}},
				},
			},
			wantTag: nil,
		},
		"extract tracking system tag from config": {
			initialTags: &providercfg.InitialTags{
				LoadBalancer: &providercfg.TagConfig{
					DefinedTags: map[string]map[string]interface{}{"ns": {"key": "val"}},
				},
				Common: &providercfg.TagConfig{
					FreeformTags: map[string]string{"Cluster": "ocid1.cluster.aa..."},
					DefinedTags:  map[string]map[string]interface{}{"orcl-containerengine": {"Cluster": "ocid1.cluster.aa..."}},
				},
			},
			wantTag: map[string]map[string]interface{}{"orcl-containerengine": {"Cluster": "ocid1.cluster.aa..."}},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tag := getResourceTrackingSystemTagsFromConfig(zap.S(), test.initialTags)
			t.Logf("%#v", tag)
			if !reflect.DeepEqual(test.wantTag, tag) {
				t.Errorf("wanted %v but got %v", test.wantTag, tag)
			}
		})
	}
}

func Test_getIngressIpMode(t *testing.T) {
	var proxy = v1.LoadBalancerIPModeProxy
	var vip = v1.LoadBalancerIPModeVIP
	var tests = map[string]struct {
		service        *v1.Service
		expectedIpMode *v1.LoadBalancerIPMode
		wantErr        error
	}{
		"ipMode is Proxy": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationIngressIpMode: "Proxy",
					},
				},
			},
			expectedIpMode: &proxy,
			wantErr:        nil,
		},
		"ipMode is VIP": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationIngressIpMode: "VIP",
					},
				},
			},
			expectedIpMode: &vip,
			wantErr:        nil,
		},
		"ipMode not set": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
				},
			},
			expectedIpMode: nil,
			wantErr:        nil,
		},
		"ipMode is invalid": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationIngressIpMode: "tcp",
					},
				},
			},
			expectedIpMode: nil,
			wantErr:        errors.New("IpMode can only be set as Proxy or VIP"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := getIngressIpMode(test.service)
			if !assertError(err, test.wantErr) {
				t.Errorf("Expected error = %v, but got %v", test.wantErr, err)
				return
			}
			if err == nil && !reflect.DeepEqual(actual, test.expectedIpMode) {
				t.Errorf("expected %v but got %v", test.expectedIpMode, actual)
			}
		})
	}
}

func Test_getRequireIpVersions(t *testing.T) {

	testCases := map[string]struct {
		listenerBackendSetIpVersion []string
		requireIPv6                 bool
		requireIPv4                 bool
	}{
		"IPv4": {
			listenerBackendSetIpVersion: []string{IPv4},
			requireIPv4:                 true,
			requireIPv6:                 false,
		},
		"IPv6": {
			listenerBackendSetIpVersion: []string{IPv6},
			requireIPv4:                 false,
			requireIPv6:                 true,
		},
		"IPv4 and IPv6": {
			listenerBackendSetIpVersion: []string{IPv4, IPv6},
			requireIPv4:                 true,
			requireIPv6:                 true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			requireIPv4, requireIPv6 := getRequireIpVersions(tc.listenerBackendSetIpVersion)
			if requireIPv6 != tc.requireIPv6 {
				t.Errorf("Expected requireIPv6:%+v\nbut requireIPv6:%+v", tc.requireIPv6, requireIPv6)
			}
			if requireIPv4 != tc.requireIPv4 {
				t.Errorf("Expected requireIPv4:%+v\nbut got requireIPv4:%+v", tc.requireIPv4, requireIPv4)
			}
		})
	}
}

func Test_getBackendSets(t *testing.T) {
	testThreeBackendSetNameIPv6 := "TCP-67-IPv6"
	testThreeBackendSetNameIPv4 := "TCP-67"

	testCases := map[string]struct {
		service                  *v1.Service
		provisionedNodes         []*v1.Node
		virtualPods              []*v1.Pod
		sslCfg                   *SSLConfig
		isPreserveSource         bool
		listenerBackendIpVersion []string
		wantBackendSets          map[string]client.GenericBackendSetDetails
		err                      error
	}{
		"IpFamilies IPv4 ListenerBackendSetIpVersion IPv4": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			provisionedNodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:130B",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:1300",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.1",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.2",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			virtualPods:              []*v1.Pod{},
			sslCfg:                   nil,
			listenerBackendIpVersion: []string{IPv4},
			wantBackendSets: map[string]client.GenericBackendSetDetails{
				"TCP-67": {
					Name:   &testThreeBackendSetNameIPv4,
					Policy: common.String("FIVE_TUPLE"),
					HealthChecker: &client.GenericHealthChecker{
						Protocol:         "HTTP",
						IsForcePlainText: common.Bool(false),
						Port:             common.Int(10256),
						UrlPath:          common.String("/healthz"),
						Retries:          common.Int(3),
						TimeoutInMillis:  common.Int(3000),
						IntervalInMillis: common.Int(10000),
						ReturnCode:       common.Int(http.StatusOK),
					},
					Backends: []client.GenericBackend{
						{IpAddress: common.String("10.0.0.1"), Port: common.Int(36667), Weight: common.Int(1), TargetId: &testNodeString},
						{IpAddress: common.String("10.0.0.2"), Port: common.Int(36667), Weight: common.Int(1), TargetId: &testNodeString},
					},
					SessionPersistenceConfiguration: nil,
					SslConfiguration:                nil,
					IpVersion:                       GenericIpVersion(client.GenericIPv4),
					IsPreserveSource:                common.Bool(false),
				},
			},
			err: nil,
		},
		"IpFamilies IPv4IPv6 ListenerBackendSetIpVersion IPv4": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			provisionedNodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:130B",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:1300",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.1",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.2",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			virtualPods:              []*v1.Pod{},
			sslCfg:                   nil,
			listenerBackendIpVersion: []string{IPv4},
			wantBackendSets: map[string]client.GenericBackendSetDetails{
				"TCP-67": {
					Name:   &testThreeBackendSetNameIPv4,
					Policy: common.String("FIVE_TUPLE"),
					HealthChecker: &client.GenericHealthChecker{
						Protocol:         "HTTP",
						IsForcePlainText: common.Bool(false),
						Port:             common.Int(10256),
						UrlPath:          common.String("/healthz"),
						Retries:          common.Int(3),
						TimeoutInMillis:  common.Int(3000),
						IntervalInMillis: common.Int(10000),
						ReturnCode:       common.Int(http.StatusOK),
					},
					Backends: []client.GenericBackend{
						{IpAddress: common.String("10.0.0.1"), Port: common.Int(36667), Weight: common.Int(1), TargetId: &testNodeString},
						{IpAddress: common.String("10.0.0.2"), Port: common.Int(36667), Weight: common.Int(1), TargetId: &testNodeString},
					},
					SessionPersistenceConfiguration: nil,
					SslConfiguration:                nil,
					IpVersion:                       GenericIpVersion(client.GenericIPv4),
					IsPreserveSource:                common.Bool(false),
				},
			},
			err: nil,
		},
		"IpFamilies IPv4IPv6 ListenerBackendSetIpVersion IPv6": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv6)},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			provisionedNodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:130B",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:1300",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.1",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.2",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			virtualPods:              []*v1.Pod{},
			sslCfg:                   nil,
			listenerBackendIpVersion: []string{IPv6},
			wantBackendSets: map[string]client.GenericBackendSetDetails{
				"TCP-67-IPv6": {
					Name:   &testThreeBackendSetNameIPv6,
					Policy: common.String("FIVE_TUPLE"),
					HealthChecker: &client.GenericHealthChecker{
						Protocol:         "HTTP",
						IsForcePlainText: common.Bool(false),
						Port:             common.Int(10256),
						UrlPath:          common.String("/healthz"),
						Retries:          common.Int(3),
						TimeoutInMillis:  common.Int(3000),
						IntervalInMillis: common.Int(10000),
						ReturnCode:       common.Int(http.StatusOK),
					},
					Backends: []client.GenericBackend{
						{IpAddress: common.String("2001:0000:130F:0000:0000:09C0:876A:130B"), Port: common.Int(36667), Weight: common.Int(1)},
						{IpAddress: common.String("2001:0000:130F:0000:0000:09C0:876A:1300"), Port: common.Int(36667), Weight: common.Int(1)},
					},
					SessionPersistenceConfiguration: nil,
					SslConfiguration:                nil,
					IpVersion:                       GenericIpVersion(client.GenericIPv6),
					IsPreserveSource:                common.Bool(false),
				},
			},
			err: nil,
		},
		"IpFamilies IPv4IPv6 ListenerBackendSetIpVersion IPv4IPv6": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			provisionedNodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:130B",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:1300",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.1",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.2",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			virtualPods:              []*v1.Pod{},
			sslCfg:                   nil,
			listenerBackendIpVersion: []string{IPv4, IPv6},
			wantBackendSets: map[string]client.GenericBackendSetDetails{
				"TCP-67": {
					Name:   &testThreeBackendSetNameIPv4,
					Policy: common.String("FIVE_TUPLE"),
					HealthChecker: &client.GenericHealthChecker{
						Protocol:         "HTTP",
						IsForcePlainText: common.Bool(false),
						Port:             common.Int(10256),
						UrlPath:          common.String("/healthz"),
						Retries:          common.Int(3),
						TimeoutInMillis:  common.Int(3000),
						IntervalInMillis: common.Int(10000),
						ReturnCode:       common.Int(http.StatusOK),
					},
					Backends: []client.GenericBackend{
						{IpAddress: common.String("10.0.0.1"), Port: common.Int(36667), Weight: common.Int(1), TargetId: &testNodeString},
						{IpAddress: common.String("10.0.0.2"), Port: common.Int(36667), Weight: common.Int(1), TargetId: &testNodeString},
					},
					SessionPersistenceConfiguration: nil,
					SslConfiguration:                nil,
					IpVersion:                       GenericIpVersion(client.GenericIPv4),
					IsPreserveSource:                common.Bool(false),
				},
				"TCP-67-IPv6": {
					Name:   &testThreeBackendSetNameIPv6,
					Policy: common.String("FIVE_TUPLE"),
					HealthChecker: &client.GenericHealthChecker{
						Protocol:         "HTTP",
						IsForcePlainText: common.Bool(false),
						Port:             common.Int(10256),
						UrlPath:          common.String("/healthz"),
						Retries:          common.Int(3),
						TimeoutInMillis:  common.Int(3000),
						IntervalInMillis: common.Int(10000),
						ReturnCode:       common.Int(http.StatusOK),
					},
					Backends: []client.GenericBackend{
						{IpAddress: common.String("2001:0000:130F:0000:0000:09C0:876A:130B"), Port: common.Int(36667), Weight: common.Int(1)},
						{IpAddress: common.String("2001:0000:130F:0000:0000:09C0:876A:1300"), Port: common.Int(36667), Weight: common.Int(1)},
					},
					SessionPersistenceConfiguration: nil,
					SslConfiguration:                nil,
					IpVersion:                       GenericIpVersion(client.GenericIPv6),
					IsPreserveSource:                common.Bool(false),
				},
			},
			err: nil,
		},
		"IpFamilies IPv6 ListenerBackendSetIpVersion IPv6": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv6)},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerType: "nlb",
					},
				},
			},
			provisionedNodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:130B",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:1300",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.1",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.2",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			virtualPods:              []*v1.Pod{},
			sslCfg:                   nil,
			listenerBackendIpVersion: []string{IPv6},
			wantBackendSets: map[string]client.GenericBackendSetDetails{
				"TCP-67-IPv6": {
					Name:   &testThreeBackendSetNameIPv6,
					Policy: common.String("FIVE_TUPLE"),
					HealthChecker: &client.GenericHealthChecker{
						Protocol:         "HTTP",
						IsForcePlainText: common.Bool(false),
						Port:             common.Int(10256),
						UrlPath:          common.String("/healthz"),
						Retries:          common.Int(3),
						TimeoutInMillis:  common.Int(3000),
						IntervalInMillis: common.Int(10000),
						ReturnCode:       common.Int(http.StatusOK),
					},
					Backends: []client.GenericBackend{
						{IpAddress: common.String("2001:0000:130F:0000:0000:09C0:876A:130B"), Port: common.Int(36667), Weight: common.Int(1)},
						{IpAddress: common.String("2001:0000:130F:0000:0000:09C0:876A:1300"), Port: common.Int(36667), Weight: common.Int(1)},
					},
					SessionPersistenceConfiguration: nil,
					SslConfiguration:                nil,
					IpVersion:                       GenericIpVersion(client.GenericIPv6),
					IsPreserveSource:                common.Bool(false),
				},
			},
			err: nil,
		},
		"IpFamilies IPv4 BackendSet cipher suite configuration": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadbalancerBackendSetSSLConfig: `{"CipherSuiteName":"oci-default-http2-ssl-cipher-suite-v1", "Protocols": ["TLSv1.2"]}`,
						ServiceAnnotationLoadBalancerTLSBackendSetSecret: "example",
					},
				},
			},
			provisionedNodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:130B",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:1300",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.1",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.2",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			virtualPods: []*v1.Pod{},
			sslCfg: &SSLConfig{
				Ports:                   sets.NewInt(67),
				ListenerSSLSecretName:   listenerSecret,
				BackendSetSSLSecretName: backendSecret,
			},
			listenerBackendIpVersion: []string{IPv4},
			wantBackendSets: map[string]client.GenericBackendSetDetails{
				"TCP-67": {
					Name:   &testThreeBackendSetNameIPv4,
					Policy: common.String("FIVE_TUPLE"),
					HealthChecker: &client.GenericHealthChecker{
						Protocol:         "HTTP",
						IsForcePlainText: common.Bool(true),
						Port:             common.Int(10256),
						UrlPath:          common.String("/healthz"),
						Retries:          common.Int(3),
						TimeoutInMillis:  common.Int(3000),
						IntervalInMillis: common.Int(10000),
						ReturnCode:       common.Int(http.StatusOK),
					},
					Backends: []client.GenericBackend{
						{IpAddress: common.String("10.0.0.1"), Port: common.Int(36667), Weight: common.Int(1), TargetId: &testNodeString},
						{IpAddress: common.String("10.0.0.2"), Port: common.Int(36667), Weight: common.Int(1), TargetId: &testNodeString},
					},
					SessionPersistenceConfiguration: nil,
					SslConfiguration: &client.GenericSslConfigurationDetails{
						VerifyDepth:           common.Int(0),
						VerifyPeerCertificate: common.Bool(false),
						CertificateName:       common.String(backendSecret),
						CipherSuiteName:       common.String("oci-default-http2-ssl-cipher-suite-v1"),
						Protocols:             []string{"TLSv1.2"},
					},
					IpVersion:        GenericIpVersion(client.GenericIPv4),
					IsPreserveSource: common.Bool(false),
				},
			},
			err: nil,
		},
		"IpFamilies IPv4 BackendSet protocols is null ": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadbalancerBackendSetSSLConfig: `{"cipherSuiteName":"oci-default-http2-ssl-cipher-suite-v1", "protocols": ["TLSv1.2"]}`,
						ServiceAnnotationLoadBalancerTLSBackendSetSecret: "example",
					},
				},
			},
			provisionedNodes: []*v1.Node{
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:130B",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "2001:0000:130F:0000:0000:09C0:876A:1300",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.1",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
				{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: v1.NodeSpec{
						ProviderID: testNodeString,
					},
					Status: v1.NodeStatus{
						Capacity:    nil,
						Allocatable: nil,
						Phase:       "",
						Conditions:  nil,
						Addresses: []v1.NodeAddress{
							{
								Address: "10.0.0.2",
								Type:    "InternalIP",
							},
						},
						DaemonEndpoints: v1.NodeDaemonEndpoints{},
						NodeInfo:        v1.NodeSystemInfo{},
						Images:          nil,
						VolumesInUse:    nil,
						VolumesAttached: nil,
						Config:          nil,
					},
				},
			},
			virtualPods: []*v1.Pod{},
			sslCfg: &SSLConfig{
				Ports:                   sets.NewInt(67),
				ListenerSSLSecretName:   listenerSecret,
				BackendSetSSLSecretName: backendSecret,
			},
			listenerBackendIpVersion: []string{IPv4},
			wantBackendSets: map[string]client.GenericBackendSetDetails{
				"TCP-67": {
					Name:   &testThreeBackendSetNameIPv4,
					Policy: common.String("FIVE_TUPLE"),
					HealthChecker: &client.GenericHealthChecker{
						Protocol:         "HTTP",
						IsForcePlainText: common.Bool(true),
						Port:             common.Int(10256),
						UrlPath:          common.String("/healthz"),
						Retries:          common.Int(3),
						TimeoutInMillis:  common.Int(3000),
						IntervalInMillis: common.Int(10000),
						ReturnCode:       common.Int(http.StatusOK),
					},
					Backends: []client.GenericBackend{
						{IpAddress: common.String("10.0.0.1"), Port: common.Int(36667), Weight: common.Int(1), TargetId: &testNodeString},
						{IpAddress: common.String("10.0.0.2"), Port: common.Int(36667), Weight: common.Int(1), TargetId: &testNodeString},
					},
					SessionPersistenceConfiguration: nil,
					SslConfiguration: &client.GenericSslConfigurationDetails{
						VerifyDepth:           common.Int(0),
						VerifyPeerCertificate: common.Bool(false),
						CertificateName:       common.String(backendSecret),
						CipherSuiteName:       common.String("oci-default-http2-ssl-cipher-suite-v1"),
						Protocols:             []string{"TLSv1.2"},
					},
					IpVersion:        GenericIpVersion(client.GenericIPv4),
					IsPreserveSource: common.Bool(false),
				},
			},
			err: nil,
		},
	}
	for name, tc := range testCases {
		logger := zap.L()
		t.Run(name, func(t *testing.T) {
			gotBackendSets, err := getBackendSets(logger.Sugar(), tc.service, tc.provisionedNodes, tc.sslCfg, tc.isPreserveSource, tc.listenerBackendIpVersion)
			if tc.err != nil && err == nil {
				t.Errorf("Expected  \n%+v\nbut got\n%+v", tc.err, err)
			}
			if err != nil && tc.err == nil {
				t.Errorf("Error: expected\n%+v\nbut got\n%+v", tc.err, err)
			}
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("Expected \n%+v\nbut got\n%+v", tc.err, err)
			}
			if len(gotBackendSets) != len(tc.wantBackendSets) {
				t.Errorf("Number of excpected listeners \n%+v\nbut got\n%+v", len(tc.wantBackendSets), len(gotBackendSets))
			}
			if len(gotBackendSets) != 0 {
				for name, backendSetDetails := range tc.wantBackendSets {
					gotBackendSet, ok := gotBackendSets[name]
					if !ok {
						t.Errorf("Expected backendSetDetails with name \n%+v\nbut backendSetDetails not present", *backendSetDetails.Name)
					}
					if *gotBackendSet.Name != *backendSetDetails.Name {
						t.Errorf("Expected backendSetDetails name \n%+v\nbut got backendSetDetails name \n%+v", *backendSetDetails.Name, *gotBackendSet.Name)
					}
					if *gotBackendSet.IpVersion != *backendSetDetails.IpVersion {
						t.Errorf("Expected backendSetDetails IpVersion \n%+v\nbut got backendSetDetails IpVersion \n%+v", *backendSetDetails.IpVersion, *gotBackendSet.IpVersion)
					}
					if !reflect.DeepEqual(backendSetDetails.Backends, gotBackendSet.Backends) {
						t.Errorf("Expected backendSetDetails backends \n%+v\nbut got backendSetDetails backends \n%+v", backendSetDetails.Backends, gotBackendSet.Backends)
					}
					if !reflect.DeepEqual(backendSetDetails.HealthChecker, gotBackendSet.HealthChecker) {
						want, _ := json.Marshal(backendSetDetails.HealthChecker)
						got, _ := json.Marshal(gotBackendSet.HealthChecker)
						t.Errorf("backendSetDetails HealthChecker failed want: %s \n got: %s \n", want, got)
					}
					if !reflect.DeepEqual(backendSetDetails.SslConfiguration, gotBackendSet.SslConfiguration) {
						want, _ := json.Marshal(backendSetDetails.SslConfiguration)
						got, _ := json.Marshal(gotBackendSet.SslConfiguration)
						t.Errorf("backendSetDetails SslConfiguration failed want: %s \n got: %s \n", want, got)
					}
				}
			}
		})
	}
}

func Test_getPorts(t *testing.T) {
	var tests = []struct {
		name       string
		service    *v1.Service
		ipVersions []string
		err        error
		ports      map[string]portSpec
	}{
		{
			name: "IpFamilies IPv4 ListenerBackendSetIpVersion IPv4",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4)},
				},
			},
			err:        nil,
			ipVersions: []string{IPv4},
			ports: map[string]portSpec{
				"TCP-67": {
					ListenerPort:      67,
					BackendPort:       36667,
					HealthCheckerPort: 10256,
				},
			},
		},
		{
			name: "IpFamilies IPv6 ListenerBackendSetIpVersion IPv6",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv6)},
				},
			},
			err:        nil,
			ipVersions: []string{IPv6},
			ports: map[string]portSpec{
				"TCP-67-IPv6": {
					ListenerPort:      67,
					BackendPort:       36667,
					HealthCheckerPort: 10256,
				},
			},
		},
		{
			name: "IpFamilies IPv4, IPv6 ListenerBackendSetIpVersion IPv4",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
				},
			},
			err:        nil,
			ipVersions: []string{IPv4},
			ports: map[string]portSpec{
				"TCP-67": {
					ListenerPort:      67,
					BackendPort:       36667,
					HealthCheckerPort: 10256,
				},
			},
		},
		{
			name: "IpFamilies IPv4, IPv6 ListenerBackendSetIpVersion IPv6",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv6)},
				},
			},
			err:        nil,
			ipVersions: []string{IPv6},
			ports: map[string]portSpec{
				"TCP-67-IPv6": {
					ListenerPort:      67,
					BackendPort:       36667,
					HealthCheckerPort: 10256,
				},
			},
		},
		{
			name: "IpFamilies IPv4 IPv6 ListenerBackendSetIpVersion IPv4 IPv6",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{
							Protocol: v1.ProtocolTCP,
							Port:     int32(67),
							NodePort: 36667,
						},
					},
					IPFamilies: []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
				},
			},
			err:        nil,
			ipVersions: []string{IPv4, IPv6},
			ports: map[string]portSpec{
				"TCP-67": {
					ListenerPort:      67,
					BackendPort:       36667,
					HealthCheckerPort: 10256,
				},
				"TCP-67-IPv6": {
					ListenerPort:      67,
					BackendPort:       36667,
					HealthCheckerPort: 10256,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getPorts(tt.service, tt.ipVersions)
			if !reflect.DeepEqual(result, tt.ports) {
				t.Errorf("getPorts() = %+v, want %+v", result, tt.ports)
			}
			if err != nil {
				if !reflect.DeepEqual(err, tt.err) {
					t.Errorf("getPorts() = %+v, want %+v", err, tt.err)
				}
			}
		})
	}
}

func Test_getLoadBalancerSourceRanges(t *testing.T) {
	var tests = []struct {
		name        string
		service     *v1.Service
		sourceCIDRs []string
	}{
		{
			name: "IpFamilies IPv4 SingleStack",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicySingleStack))),
				},
			},
			sourceCIDRs: []string{"0.0.0.0/0"},
		},
		{
			name: "IpFamilies IPv6 SingleStack",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicySingleStack))),
				},
			},
			sourceCIDRs: []string{"::/0"},
		},
		{
			name: "IpFamilies IPv4, IPv6 PreferDualStack",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyPreferDualStack))),
				},
			},
			sourceCIDRs: []string{"0.0.0.0/0", "::/0"},
		},
		{
			name: "IpFamilies IPv4, IPv6 RequireDualStack",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					IPFamilies:     []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy: (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicyRequireDualStack))),
				},
			},
			sourceCIDRs: []string{"0.0.0.0/0", "::/0"},
		},
		{
			name: "IpFamilies IPv4 IPv6 Custom Cidr provided in spec",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					LoadBalancerSourceRanges: []string{"2.2.2.0/24"},
					IPFamilies:               []v1.IPFamily{v1.IPFamily(IPv4), v1.IPFamily(IPv6)},
					IPFamilyPolicy:           (*v1.IPFamilyPolicy)(common.String("PreferDualStack")),
				},
			},
			sourceCIDRs: []string{"2.2.2.0/24"},
		},
		{
			name: "IpFamilies IPv6 Custom Cidr provided in spec",
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					LoadBalancerSourceRanges: []string{"1.1.1.0/24", "2603:c120:000f:8b99:0000:0000:0000:0000/64"},
					IPFamilies:               []v1.IPFamily{v1.IPFamily(IPv6)},
					IPFamilyPolicy:           (*v1.IPFamilyPolicy)(common.String(string(v1.IPFamilyPolicySingleStack))),
				},
			},
			sourceCIDRs: []string{"1.1.1.0/24", "2603:c120:f:8b99::/64"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := getLoadBalancerSourceRanges(tt.service)
			for _, cidr := range result {
				if !contains(tt.sourceCIDRs, cidr) {
					t.Errorf("getLoadBalancerSourceRanges() = %+v, want %+v", result, tt.sourceCIDRs)
				}
			}
		})
	}
}

func TestIsSkipPrivateIP_NLB(t *testing.T) {
	tests := []struct {
		name           string
		svcAnnotations map[string]string
		expected       bool
		wantErr        bool
	}{
		{
			name: "skip-private-ip-enabled",
			svcAnnotations: map[string]string{
				ServiceAnnotationLoadBalancerType:                  NLB,
				ServiceAnnotationNetworkLoadBalancerExternalIpOnly: "true",
			},
			expected: true,
			wantErr:  false,
		},
		{
			name: "skip-private-ip-disabled",
			svcAnnotations: map[string]string{
				ServiceAnnotationLoadBalancerType:                  NLB,
				ServiceAnnotationNetworkLoadBalancerExternalIpOnly: "false",
			},
			expected: false,
			wantErr:  false,
		},
		{
			name: "skip-private-ip-invalid-value",
			svcAnnotations: map[string]string{
				ServiceAnnotationLoadBalancerType:                  NLB,
				ServiceAnnotationNetworkLoadBalancerExternalIpOnly: "invalid",
			},
			expected: false,
			wantErr:  true,
		},
		{
			name: "skip-private-ip with internal loadbalancer",
			svcAnnotations: map[string]string{
				ServiceAnnotationLoadBalancerType:                  NLB,
				ServiceAnnotationNetworkLoadBalancerInternal:       "true",
				ServiceAnnotationNetworkLoadBalancerExternalIpOnly: "true",
			},
			expected: false,
			wantErr:  false,
		},
		{
			name:           "no-skip-private-ip-annotation",
			svcAnnotations: map[string]string{},
			expected:       false,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: tt.svcAnnotations,
				},
			}
			got, err := isSkipPrivateIP(svc)
			if (err != nil) != tt.wantErr {
				t.Errorf("isSkipPrivateIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("isSkipPrivateIP() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
