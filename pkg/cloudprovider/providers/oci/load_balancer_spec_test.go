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
	"net/http"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v50/common"
)

var (
	backendSecret  = "backendsecret"
	listenerSecret = "listenersecret"
	testNodeString = "testNodeTargetID"
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
	testCases := map[string]struct {
		defaultSubnetOne string
		defaultSubnetTwo string
		nodes            []*v1.Node
		service          *v1.Service
		expected         *LBSpec
		sslConfig        *SSLConfig
		clusterTags      *providercfg.InitialTags
	}{
		"defaults": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"defaults-nlb-cluster-policy": {
			defaultSubnetOne: "one",
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
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"defaults-nlb-local-policy": {
			defaultSubnetOne: "one",
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
					},
				},
				BackendSets: map[string]client.GenericBackendSetDetails{
					"TCP-80": {
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
							Port:             common.Int(10256),
							UrlPath:          common.String("/healthz"),
							Retries:          common.Int(3),
							TimeoutInMillis:  common.Int(3000),
							IntervalInMillis: common.Int(10000),
							ReturnCode:       common.Int(http.StatusOK),
						},
						IsPreserveSource: common.Bool(true),
						Policy:           common.String("FIVE_TUPLE"),
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"internal with default subnet": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"internal with overridden regional subnet1": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"internal with overridden regional subnet2": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"internal with no default subnets provide subnet1 via annotation": {
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"use default subnet in case of no subnet overrides via annotation": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"no default subnets provide subnet1 via annotation as regional-subnet": {
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"no default subnets provide subnet2 via annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerSubnet2: "annotation-two",
					},
				},
				Spec: v1.ServiceSpec{
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
				Subnets:  []string{"", "annotation-two"},
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"override default subnet via subnet1 annotation as regional subnet": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"override default subnet via subnet2 annotation as regional subnet": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"override default subnet via subnet1 and subnet2 annotation": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
					"TCP-80": portSpec{
						ListenerPort:      80,
						HealthCheckerPort: 10256,
					},
				},
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		//"security list manager annotation":
		"custom shape": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"custom idle connection timeout": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"custom proxy protocol version w/o timeout for multiple listeners": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"custom proxy protocol version and timeout": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"protocol annotation set to http": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"protocol annotation set to tcp": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"protocol annotation empty": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"LBSpec returned with proper SSLConfiguration": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:   "kube-system",
					Name:        "testservice",
					UID:         "test-uid",
					Annotations: map[string]string{},
				},
				Spec: v1.ServiceSpec{
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"flex shape": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"valid loadbalancer policy": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"default loadbalancer policy": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
			},
		},
		"load balancer with reserved ip": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
				LoadBalancerIP:      "10.0.0.0",
			},
		},
		"defaults with tags": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
				FreeformTags:        map[string]string{"cluster": "resource", "unique": "tag"},
				DefinedTags:         map[string]map[string]interface{}{"namespace": {"owner": "team", "key": "value"}},
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
			result, err := NewLBSpec(logger.Sugar(), tc.service, tc.nodes, subnets, tc.sslConfig, slManagerFactory, tc.clusterTags)
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
		service          *v1.Service
		expected         *LBSpec
		clusterTags      *providercfg.InitialTags
	}{
		"single subnet for single AD": {
			defaultSubnetOne: "one",
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
						Backends: []client.GenericBackend{},
						HealthChecker: &client.GenericHealthChecker{
							Protocol:         "HTTP",
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
				securityListManager: newSecurityListManagerNOOP(),
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
			result, err := NewLBSpec(logger.Sugar(), tc.service, tc.nodes, subnets, nil, slManagerFactory, tc.clusterTags)
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
		service          *v1.Service
		//add cp or cp security list
		expectedErrMsg string
		clusterTags    *providercfg.InitialTags
	}{
		"unsupported udp protocol": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolUDP},
					},
				},
			},
			expectedErrMsg: "invalid service: OCI load balancers do not support UDP",
		},
		"unsupported session affinity": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
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
					SessionAffinity: v1.ServiceAffinityNone,
					Ports:           []v1.ServicePort{},
					//add security list mananger in spec
				},
			},
			expectedErrMsg: "a configuration for subnet1 must be specified for an internal load balancer",
		},
		"internal lb with empty subnet1 annotation": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "true",
						ServiceAnnotationLoadBalancerSubnet1:  "",
					},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports:           []v1.ServicePort{},
					//add security list mananger in spec
				},
			},
			expectedErrMsg: "a configuration for subnet1 must be specified for an internal load balancer",
		},
		"non boolean internal lb": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerInternal: "yes",
					},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports:           []v1.ServicePort{},
				},
			},
			expectedErrMsg: fmt.Sprintf("invalid value: yes provided for annotation: %s: strconv.ParseBool: parsing \"yes\": invalid syntax", ServiceAnnotationLoadBalancerInternal),
		},
		"invalid flex shape missing min": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "error parsing service annotation: service.beta.kubernetes.io/oci-load-balancer-shape=flexible requires service.beta.kubernetes.io/oci-load-balancer-shape-flex-min and service.beta.kubernetes.io/oci-load-balancer-shape-flex-max to be set",
		},
		"invalid flex shape non int min/max": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kube-system",
					Name:      "testservice",
					UID:       "test-uid",
					Annotations: map[string]string{
						ServiceAnnotationLoadBalancerShape:        "flexible",
						ServiceAnnotationLoadBalancerShapeFlexMin: "10Mbps",
						ServiceAnnotationLoadBalancerShapeFlexMax: "100Mbps",
					},
				},
				Spec: v1.ServiceSpec{
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: `The annotation service.beta.kubernetes.io/oci-load-balancer-shape-flex-min should contain only integer value: strconv.Atoi: parsing "10Mbps": invalid syntax`,
		},
		"invalid loadbalancer policy": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: `loadbalancer policy "not-valid-loadbalancer-policy" is not valid`,
		},
		"invalid loadBalancerIP format": {
			service: &v1.Service{
				Spec: v1.ServiceSpec{
					LoadBalancerIP:  "non-ip-format",
					SessionAffinity: v1.ServiceAffinityNone,
				},
			},
			expectedErrMsg: "invalid value \"non-ip-format\" provided for LoadBalancerIP",
		},
		"unsupported loadBalancerIP for internal load balancer": {
			defaultSubnetOne: "one",
			defaultSubnetTwo: "two",
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
					SessionAffinity: v1.ServiceAffinityNone,
					Ports: []v1.ServicePort{
						{Protocol: v1.ProtocolTCP},
					},
				},
			},
			expectedErrMsg: "failed to parse defined tags annotation: invalid character 'w' looking for beginning of value",
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
			if err == nil {
				slManagerFactory := func(mode string) securityListManager {
					return newSecurityListManagerNOOP()
				}
				_, err = NewLBSpec(logger.Sugar(), tc.service, tc.nodes, subnets, nil, slManagerFactory, tc.clusterTags)
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
		"Does not container the Load Balancer SSL Ports Annotation": {
			expected:    false,
			annotations: make(map[string]string, 0),
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

func Test_getBackends(t *testing.T) {
	type args struct {
		nodes    []*v1.Node
		nodePort int32
	}
	var tests = []struct {
		name string
		args args
		want []client.GenericBackend
	}{
		{
			name: "no nodes",
			args: args{nodePort: 80},
			want: []client.GenericBackend{},
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
			want: []client.GenericBackend{},
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
			want: []client.GenericBackend{},
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.L()
			if got := getBackends(logger.Sugar(), tt.args.nodes, tt.args.nodePort); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBackends() = %+v, want %+v", got, tt.want)
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
		service *v1.Service
		name    string
		want    map[string]client.GenericListener
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
			if got, _ := getListeners(svc, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getListeners() = %+v, \n want %+v", got, tt.want)

			}
		})
	}
}

func Test_getSecurityListManagementMode(t *testing.T) {
	testCases := map[string]struct {
		service  *v1.Service
		expected string
	}{
		"defaults": {
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expected: "",
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

	testCases := map[string]struct {
		service       *v1.Service
		wantListeners map[string]client.GenericListener
		err           error
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
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			gotListeners, err := getListenersNetworkLoadBalancer(tc.service)
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
