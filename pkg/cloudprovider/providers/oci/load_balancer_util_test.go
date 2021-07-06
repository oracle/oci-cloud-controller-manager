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
	"os"
	"reflect"
	"testing"

	"go.uber.org/zap"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oracle/oci-go-sdk/v31/common"
	"github.com/oracle/oci-go-sdk/v31/loadbalancer"
)

func TestSortAndCombineActions(t *testing.T) {
	testCases := map[string]struct {
		backendSetActions []Action
		listenerActions   []Action
		expected          []Action
	}{
		"create": {
			backendSetActions: []Action{
				&BackendSetAction{
					name:       "TCP-80",
					actionType: Create,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&BackendSetAction{
					name:       "TCP-443",
					actionType: Create,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
			},
			listenerActions: []Action{
				&ListenerAction{
					name:       "TCP-443",
					actionType: Create,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&ListenerAction{
					name:       "TCP-80",
					actionType: Create,
					Listener:   loadbalancer.ListenerDetails{},
				},
			},
			expected: []Action{
				&BackendSetAction{
					name:       "TCP-443",
					actionType: Create,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&ListenerAction{
					name:       "TCP-443",
					actionType: Create,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&BackendSetAction{
					name:       "TCP-80",
					actionType: Create,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&ListenerAction{
					name:       "TCP-80",
					actionType: Create,
					Listener:   loadbalancer.ListenerDetails{},
				},
			},
		},
		"update": {
			backendSetActions: []Action{
				&BackendSetAction{
					name:       "TCP-80",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&BackendSetAction{
					name:       "TCP-443",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&BackendSetAction{
					name:       "TCP-445",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&BackendSetAction{
					name:       "TCP-444",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&BackendSetAction{
					name:       "TCP-442",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
			},
			listenerActions: []Action{
				&ListenerAction{
					name:       "TCP-443",
					actionType: Update,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&ListenerAction{
					name:       "TCP-80",
					actionType: Update,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&ListenerAction{
					name:       "TCP-445",
					actionType: Update,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&ListenerAction{
					name:       "TCP-442",
					actionType: Update,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&ListenerAction{
					name:       "TCP-444",
					actionType: Update,
					Listener:   loadbalancer.ListenerDetails{},
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-442",
					actionType: Update,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&BackendSetAction{
					name:       "TCP-442",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&ListenerAction{
					name:       "TCP-443",
					actionType: Update,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&BackendSetAction{
					name:       "TCP-443",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&ListenerAction{
					name:       "TCP-444",
					actionType: Update,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&BackendSetAction{
					name:       "TCP-444",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&ListenerAction{
					name:       "TCP-445",
					actionType: Update,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&BackendSetAction{
					name:       "TCP-445",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&ListenerAction{
					name:       "TCP-80",
					actionType: Update,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&BackendSetAction{
					name:       "TCP-80",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
			},
		},
		"delete": {
			backendSetActions: []Action{
				&BackendSetAction{
					name:       "TCP-80",
					actionType: Delete,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&BackendSetAction{
					name:       "TCP-443",
					actionType: Delete,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
			},
			listenerActions: []Action{
				&ListenerAction{
					name:       "TCP-443-secret",
					actionType: Delete,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&ListenerAction{
					name:       "TCP-80",
					actionType: Delete,
					Listener:   loadbalancer.ListenerDetails{},
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-443-secret",
					actionType: Delete,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&BackendSetAction{
					name:       "TCP-443",
					actionType: Delete,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
				&ListenerAction{
					name:       "TCP-80",
					actionType: Delete,
					Listener:   loadbalancer.ListenerDetails{},
				},
				&BackendSetAction{
					name:       "TCP-80",
					actionType: Delete,
					BackendSet: loadbalancer.BackendSetDetails{},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := sortAndCombineActions(zap.S(), tc.backendSetActions, tc.listenerActions)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected\n%+v\nbut got\n%+v", tc.expected, result)
			}
		})
	}
}

func TestGetBackendSetChanges(t *testing.T) {
	var testCases = []struct {
		name     string
		desired  map[string]loadbalancer.BackendSetDetails
		actual   map[string]loadbalancer.BackendSet
		expected []Action
	}{
		{
			name: "create backendset",
			desired: map[string]loadbalancer.BackendSetDetails{
				"one": loadbalancer.BackendSetDetails{
					Backends: []loadbalancer.BackendDetails{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
					},
				},
				"two": loadbalancer.BackendSetDetails{
					Backends: []loadbalancer.BackendDetails{
						{IpAddress: common.String("0.0.0.3"), Port: common.Int(80)},
						{IpAddress: common.String("0.0.0.4"), Port: common.Int(80)},
					},
				},
			},
			actual: map[string]loadbalancer.BackendSet{
				"one": loadbalancer.BackendSet{
					Name: common.String("one"),
					Backends: []loadbalancer.Backend{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80), Name: common.String("0.0.0.0:80")},
					},
				},
			},
			expected: []Action{
				&BackendSetAction{
					name:       "two",
					actionType: Create,
					BackendSet: loadbalancer.BackendSetDetails{
						Backends: []loadbalancer.BackendDetails{
							{IpAddress: common.String("0.0.0.3"), Port: common.Int(80)},
							{IpAddress: common.String("0.0.0.4"), Port: common.Int(80)},
						},
					},
					Ports: portSpec{
						BackendPort: 80,
					},
				},
			},
		},
		{
			name: "update backendset - add backend",
			desired: map[string]loadbalancer.BackendSetDetails{
				"one": loadbalancer.BackendSetDetails{
					Backends: []loadbalancer.BackendDetails{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
						{IpAddress: common.String("0.0.0.1"), Port: common.Int(80)},
					},
				},
			},
			actual: map[string]loadbalancer.BackendSet{
				"one": loadbalancer.BackendSet{
					Name: common.String("one"),
					Backends: []loadbalancer.Backend{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80), Name: common.String("0.0.0.0:80")},
					},
				},
			},
			expected: []Action{
				&BackendAction{
					bsName:     "one",
					actionType: Create,
					name:       "0.0.0.1:80",
					Backend: loadbalancer.BackendDetails{
						IpAddress: common.String("0.0.0.1"),
						Port:      common.Int(80),
					},
				},
			},
		},
		{
			name: "update backendset - remove backend",
			desired: map[string]loadbalancer.BackendSetDetails{
				"one": loadbalancer.BackendSetDetails{
					Backends: []loadbalancer.BackendDetails{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
					},
				},
			},
			actual: map[string]loadbalancer.BackendSet{
				"one": loadbalancer.BackendSet{
					Name: common.String("one"),
					Backends: []loadbalancer.Backend{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80), Name: common.String("0.0.0.0:80")},
						{IpAddress: common.String("0.0.0.1"), Port: common.Int(80), Name: common.String("0.0.0.1:80")},
					},
				},
			},
			expected: []Action{
				&BackendAction{
					bsName:     "one",
					actionType: Delete,
					name:       "0.0.0.1:80",
				},
			},
		},
		{
			name: "update backendset - update policy",
			desired: map[string]loadbalancer.BackendSetDetails{
				"one": loadbalancer.BackendSetDetails{
					Backends: []loadbalancer.BackendDetails{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
					},
					Policy: common.String("ROUND_ROBIN"),
				},
			},
			actual: map[string]loadbalancer.BackendSet{
				"one": loadbalancer.BackendSet{
					Name: common.String("one"),
					Backends: []loadbalancer.Backend{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80), Name: common.String("0.0.0.0:80")},
					},
					Policy: common.String("IP HASH"),
				},
			},
			expected: []Action{
				&BackendSetAction{
					name:       "one",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{
						Backends: []loadbalancer.BackendDetails{
							{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
						},
						Policy: common.String("ROUND_ROBIN"),
					},
					Ports: portSpec{
						BackendPort: 80,
					},
					OldPorts: &portSpec{
						BackendPort: 80,
					},
				},
			},
		},

		{
			name:    "remove backendset",
			desired: map[string]loadbalancer.BackendSetDetails{},
			actual: map[string]loadbalancer.BackendSet{
				"one": loadbalancer.BackendSet{
					Name: common.String("one"),
					Backends: []loadbalancer.Backend{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80), Name: common.String("0.0.0.0:80")},
					},
				},
			},
			expected: []Action{
				&BackendSetAction{
					name:       "one",
					actionType: Delete,
					BackendSet: loadbalancer.BackendSetDetails{
						Backends: []loadbalancer.BackendDetails{
							{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
						},
					},
					Ports: portSpec{
						BackendPort: 80,
					},
				},
			},
		},
		{
			name: "no change",
			desired: map[string]loadbalancer.BackendSetDetails{
				"one": loadbalancer.BackendSetDetails{
					Backends: []loadbalancer.BackendDetails{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
					},
				},
			},
			actual: map[string]loadbalancer.BackendSet{
				"one": loadbalancer.BackendSet{
					Name: common.String("one"),
					Backends: []loadbalancer.Backend{
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80), Name: common.String("0.0.0.0:80")},
					},
				},
			},
			expected: []Action{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			changes := getBackendSetChanges(zap.S(), tt.actual, tt.desired)
			if len(changes) == 0 && len(tt.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(changes, tt.expected) {
				t.Errorf("expected BackendSetActions\n%+v\nbut got\n%+v", tt.expected, changes)
			}
		})
	}
}

func TestGetListenerChanges(t *testing.T) {
	var testCases = []struct {
		name     string
		desired  map[string]loadbalancer.ListenerDetails
		actual   map[string]loadbalancer.Listener
		expected []Action
	}{
		{
			name: "create listener",
			desired: map[string]loadbalancer.ListenerDetails{"TCP-443": loadbalancer.ListenerDetails{
				DefaultBackendSetName: common.String("TCP-443"),
				Protocol:              common.String("TCP"),
				Port:                  common.Int(443),
			}},
			actual: map[string]loadbalancer.Listener{},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-443",
					actionType: Create,
					Listener: loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-443"),
						Protocol:              common.String("TCP"),
						Port:                  common.Int(443),
					},
				},
			},
		},
		{
			name: "add listener",
			desired: map[string]loadbalancer.ListenerDetails{
				"TCP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
				"TCP-443": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-443"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(443),
				},
			},
			actual: map[string]loadbalancer.Listener{
				"TCP-80": loadbalancer.Listener{
					Name:                  common.String("TCP-80"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-443",
					actionType: Create,
					Listener: loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-443"),
						Protocol:              common.String("TCP"),
						Port:                  common.Int(443),
					},
				},
			},
		},
		{
			name: "remove listener",
			desired: map[string]loadbalancer.ListenerDetails{
				"TCP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			actual: map[string]loadbalancer.Listener{
				"TCP-443": loadbalancer.Listener{
					Name:                  common.String("TCP-443"),
					DefaultBackendSetName: common.String("TCP-443"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(443),
				},
				"TCP-80": loadbalancer.Listener{
					Name:                  common.String("TCP-80"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-443",
					actionType: Delete,
					Listener: loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-443"),
						Protocol:              common.String("TCP"),
						Port:                  common.Int(443),
					},
				},
			},
		},
		{
			name: "remove listener [legacy listeners]",
			desired: map[string]loadbalancer.ListenerDetails{
				"TCP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			actual: map[string]loadbalancer.Listener{
				"TCP-443-secret": loadbalancer.Listener{
					Name:                  common.String("TCP-443-secret"),
					DefaultBackendSetName: common.String("TCP-443"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(443),
				},
				"TCP-80": loadbalancer.Listener{
					Name:                  common.String("TCP-80"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-443-secret",
					actionType: Delete,
					Listener: loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-443"),
						Protocol:              common.String("TCP"),
						Port:                  common.Int(443),
					},
				},
			},
		},
		{
			name: "no change",
			desired: map[string]loadbalancer.ListenerDetails{
				"TCP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			actual: map[string]loadbalancer.Listener{
				"TCP-80": loadbalancer.Listener{
					Name:                  common.String("TCP-80"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			expected: []Action{},
		},
		{
			name: "no change [legacy listeners]",
			desired: map[string]loadbalancer.ListenerDetails{
				"TCP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			actual: map[string]loadbalancer.Listener{
				"TCP-80-secret": loadbalancer.Listener{
					Name:                  common.String("TCP-80-secret"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			expected: []Action{},
		},
		{
			name: "ssl config change",
			desired: map[string]loadbalancer.ListenerDetails{
				"TCP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
					SslConfiguration: &loadbalancer.SslConfigurationDetails{
						CertificateName: common.String("desired"),
					},
				},
			},
			actual: map[string]loadbalancer.Listener{
				"TCP-80": loadbalancer.Listener{
					Name:                  common.String("TCP-80"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
					SslConfiguration: &loadbalancer.SslConfiguration{
						CertificateName: common.String("actual"),
					},
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-80",
					actionType: Update,
					Listener: loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Protocol:              common.String("TCP"),
						Port:                  common.Int(80),
						SslConfiguration: &loadbalancer.SslConfigurationDetails{
							CertificateName: common.String("desired"),
						},
					},
				},
			},
		},
		{
			name: "idle timeout change",
			desired: map[string]loadbalancer.ListenerDetails{
				"TCP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
					ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
						IdleTimeout: common.Int64(100),
					},
				},
			},
			actual: map[string]loadbalancer.Listener{
				"TCP-80": loadbalancer.Listener{
					Name:                  common.String("TCP-80"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
					ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
						IdleTimeout: common.Int64(200),
					},
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-80",
					actionType: Update,
					Listener: loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Protocol:              common.String("TCP"),
						Port:                  common.Int(80),
						ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
							IdleTimeout: common.Int64(100),
						},
					},
				},
			},
		},
		{
			name: "proxy protocol version change",
			desired: map[string]loadbalancer.ListenerDetails{
				"TCP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
					ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
						IdleTimeout:                    common.Int64(100),
						BackendTcpProxyProtocolVersion: common.Int(2),
					},
				},
			},
			actual: map[string]loadbalancer.Listener{
				"TCP-80": loadbalancer.Listener{
					Name:                  common.String("TCP-80"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
					ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
						IdleTimeout:                    common.Int64(100),
						BackendTcpProxyProtocolVersion: nil,
					},
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-80",
					actionType: Update,
					Listener: loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Protocol:              common.String("TCP"),
						Port:                  common.Int(80),
						ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
							IdleTimeout:                    common.Int64(100),
							BackendTcpProxyProtocolVersion: common.Int(2),
						},
					},
				},
			},
		},
		{
			name: "ssl config change [legacy listeners]",
			desired: map[string]loadbalancer.ListenerDetails{
				"TCP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
					SslConfiguration: &loadbalancer.SslConfigurationDetails{
						CertificateName: common.String("desired"),
					},
				},
			},
			actual: map[string]loadbalancer.Listener{
				"TCP-80-secret": loadbalancer.Listener{
					Name:                  common.String("TCP-80-secret"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
					SslConfiguration: &loadbalancer.SslConfiguration{
						CertificateName: common.String("arg"),
					},
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-80-secret",
					actionType: Update,
					Listener: loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Protocol:              common.String("TCP"),
						Port:                  common.Int(80),
						SslConfiguration: &loadbalancer.SslConfigurationDetails{
							CertificateName: common.String("desired"),
						},
					},
				},
			},
		},
		{
			name: "protocol change TCP to HTTP",
			desired: map[string]loadbalancer.ListenerDetails{
				"HTTP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("HTTP"),
					Port:                  common.Int(80),
				},
			},
			actual: map[string]loadbalancer.Listener{
				"TCP-80": loadbalancer.Listener{
					Name:                  common.String("TCP-80"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "TCP-80",
					actionType: Update,
					Listener: loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Protocol:              common.String("HTTP"),
						Port:                  common.Int(80),
					},
				},
			},
		},
		{
			name: "protocol change HTTP to TCP",
			desired: map[string]loadbalancer.ListenerDetails{
				"TCP-80": loadbalancer.ListenerDetails{
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("TCP"),
					Port:                  common.Int(80),
				},
			},
			actual: map[string]loadbalancer.Listener{
				"HTTP-80": loadbalancer.Listener{
					Name:                  common.String("HTTP-80"),
					DefaultBackendSetName: common.String("TCP-80"),
					Protocol:              common.String("HTTP"),
					Port:                  common.Int(80),
				},
			},
			expected: []Action{
				&ListenerAction{
					name:       "HTTP-80",
					actionType: Update,
					Listener: loadbalancer.ListenerDetails{
						DefaultBackendSetName: common.String("TCP-80"),
						Protocol:              common.String("TCP"),
						Port:                  common.Int(80),
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			changes := getListenerChanges(zap.S(), tt.actual, tt.desired)
			if len(changes) == 0 && len(tt.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(changes, tt.expected) {
				t.Errorf("expected ListenerActions\n%+v\nbut got\n%+v", tt.expected, changes)
			}
		})
	}
}

func TestGetSSLEnabledPorts(t *testing.T) {
	testCases := []struct {
		name        string
		annotations map[string]string
		expected    []int
	}{
		{
			name:        "empty",
			annotations: map[string]string{},
			expected:    []int{},
		}, {
			name:        "empty string",
			annotations: map[string]string{"service.beta.kubernetes.io/oci-load-balancer-ssl-ports": ""},
			expected:    []int{},
		}, {
			name:        "443",
			annotations: map[string]string{"service.beta.kubernetes.io/oci-load-balancer-ssl-ports": "443"},
			expected:    []int{443},
		}, {
			name:        "1,2,3",
			annotations: map[string]string{"service.beta.kubernetes.io/oci-load-balancer-ssl-ports": "1,2,3"},
			expected:    []int{1, 2, 3},
		}, {
			name:        "1, 2, 3",
			annotations: map[string]string{"service.beta.kubernetes.io/oci-load-balancer-ssl-ports": "1, 2, 3"},
			expected:    []int{1, 2, 3},
		}, {
			name:        "not-an-integer",
			annotations: map[string]string{"service.beta.kubernetes.io/oci-load-balancer-ssl-ports": "not-an-integer"},
			expected:    nil, // becuase we error
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			sslEnabledPorts, _ := getSSLEnabledPorts(&api.Service{
				ObjectMeta: metav1.ObjectMeta{Annotations: tt.annotations},
			})
			if !reflect.DeepEqual(sslEnabledPorts, tt.expected) {
				t.Errorf("getSSLEnabledPorts(%#v) => (%#v), expected (%#v)",
					tt.annotations, sslEnabledPorts, tt.expected)
			}
		})
	}
}

func TestParseSeceretString(t *testing.T) {
	testCases := []struct {
		secretName        string
		servcieNamespace  string
		expectedName      string
		expectedNamespace string
	}{
		{
			secretName:        "secret-name",
			expectedName:      "secret-name",
			expectedNamespace: "",
		}, {
			secretName:        "secret-namespace/secret-name",
			expectedName:      "secret-name",
			expectedNamespace: "secret-namespace",
		}, {
			secretName:        "secret-namespace/secret-name/some-extra-stuff",
			expectedName:      "secret-name",
			expectedNamespace: "secret-namespace",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.secretName, func(t *testing.T) {
			secretNamespace, secretName := parseSecretString(tt.secretName)
			if secretNamespace != tt.expectedNamespace || secretName != tt.expectedName {
				t.Errorf("parseSecretString(%s, %s) => (%s, %s), expected (%s, %s)",
					tt.secretName, tt.servcieNamespace, secretNamespace, secretName, tt.expectedNamespace, tt.expectedName)
			}
		})
	}
}

func TestGetLoadBalancerName(t *testing.T) {
	testCases := map[string]struct {
		prefix   string
		service  *api.Service
		expected string
	}{
		"no prefix": {
			prefix: "",
			service: &api.Service{
				ObjectMeta: metav1.ObjectMeta{
					UID: "fakeuid",
				},
			},
			expected: "fakeuid",
		},
		"prefix without hyphen": {
			prefix: "testprefix",
			service: &api.Service{
				ObjectMeta: metav1.ObjectMeta{
					UID: "fakeuid",
				},
			},
			expected: "testprefix-fakeuid",
		},
		"prefix with hyphen": {
			prefix: "testprefix-",
			service: &api.Service{
				ObjectMeta: metav1.ObjectMeta{
					UID: "fakeuid",
				},
			},
			expected: "testprefix-fakeuid",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := os.Setenv(lbNamePrefixEnvVar, tc.prefix)
			if err != nil {
				t.Fatal(err)
			}

			result := GetLoadBalancerName(tc.service)
			if result != tc.expected {
				t.Errorf("Expected load balancer name `%s` but got `%s`", tc.expected, result)
			}
		})
	}
}

func TestGetSanitizedName(t *testing.T) {
	testCases := []struct {
		name     string
		arg      string
		expected string
	}{
		{
			"legacy name (with suffix secret name added)",
			"TCP-80-mysecret1",
			"TCP-80",
		},
		{
			"new name (suffix secret name omitted)",
			"TCP-80",
			"TCP-80",
		},
		{
			"Name has HTTP",
			"HTTP-80",
			"TCP-80",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := getSanitizedName(tc.arg); got != tc.expected {
				t.Errorf("Expected sanitized listener name '%s' but got '%s'", tc.expected, got)
			}
		})
	}

}

func TestGetListenerName(t *testing.T) {
	type args struct {
		protocol string
		port     int
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			"name",
			args{
				"TCP",
				80,
			},
			"TCP-80",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getListenerName(tt.args.protocol, tt.args.port); got != tt.expected {
				t.Errorf("Expected listener name = %v, but got %v", tt.expected, got)
			}
		})
	}
}

func TestHasListenerChanged(t *testing.T) {
	var testCases = []struct {
		name     string
		desired  loadbalancer.ListenerDetails
		actual   loadbalancer.Listener
		expected bool
	}{
		{
			name: "DefaultBackendSetName changes",
			desired: loadbalancer.ListenerDetails{
				DefaultBackendSetName: common.String("TCP-443"),
				Protocol:              common.String("TCP"),
				Port:                  common.Int(443),
			},
			actual: loadbalancer.Listener{
				DefaultBackendSetName: common.String("TCP-4431"),
				Protocol:              common.String("TCP"),
				Port:                  common.Int(443)},
			expected: true,
		},
		{
			name: "Port changes",
			desired: loadbalancer.ListenerDetails{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
			},
			actual: loadbalancer.Listener{
				Protocol: common.String("TCP"),
				Port:     common.Int(442)},
			expected: true,
		},
		{
			name: "Protocol changes",
			desired: loadbalancer.ListenerDetails{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
			},
			actual: loadbalancer.Listener{
				Protocol: common.String("HTTP"),
				Port:     common.Int(443)},
			expected: true,
		},
		{
			name: "SSLConfigurationChanges present in actual but not in desired",
			desired: loadbalancer.ListenerDetails{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
			},
			actual: loadbalancer.Listener{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
				SslConfiguration: &loadbalancer.SslConfiguration{
					VerifyDepth:           common.Int(1),
					VerifyPeerCertificate: common.Bool(true),
					CertificateName:       common.String("actual"),
				},
			},

			expected: true,
		},
		{
			name: "SSLConfigurationChanges present in desired but not in actual",
			desired: loadbalancer.ListenerDetails{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
				SslConfiguration: &loadbalancer.SslConfigurationDetails{
					CertificateName: common.String("desired"),
				},
			},
			actual: loadbalancer.Listener{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
			},

			expected: true,
		},
		{
			name: "SSLConfigurationChanges CertificateName changes",
			desired: loadbalancer.ListenerDetails{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
				SslConfiguration: &loadbalancer.SslConfigurationDetails{
					CertificateName:       common.String("desired"),
					VerifyDepth:           common.Int(1),
					VerifyPeerCertificate: common.Bool(true),
				},
			},
			actual: loadbalancer.Listener{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
				SslConfiguration: &loadbalancer.SslConfiguration{
					VerifyDepth:           common.Int(1),
					VerifyPeerCertificate: common.Bool(true),
					CertificateName:       common.String("actual"),
				},
			},

			expected: true,
		},
		{
			name: "ConnectionConfiguration present in actual but not in desired",
			desired: loadbalancer.ListenerDetails{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
			},
			actual: loadbalancer.Listener{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
				ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
					IdleTimeout: common.Int64(300),
				},
			},

			expected: false,
		},
		{
			name: "ConnectionConfiguration present in desired but not in actual",
			desired: loadbalancer.ListenerDetails{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
				ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
					IdleTimeout: common.Int64(300),
				},
			},
			actual: loadbalancer.Listener{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
			},

			expected: true,
		},
		{
			name: "ConnectionConfiguration IdleTimeout changed",
			desired: loadbalancer.ListenerDetails{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
				ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
					IdleTimeout: common.Int64(300),
				},
			},
			actual: loadbalancer.Listener{
				Protocol: common.String("TCP"),
				Port:     common.Int(443),
				ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
					IdleTimeout: common.Int64(400),
				},
			},

			expected: true,
		},
		{
			name: "no changes",
			desired: loadbalancer.ListenerDetails{
				DefaultBackendSetName: common.String("TCP-443"),
				Protocol:              common.String("TCP"),
				Port:                  common.Int(443),
				SslConfiguration: &loadbalancer.SslConfigurationDetails{
					CertificateName: common.String("cert"),
					VerifyDepth:     common.Int(1),
				},
				ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
					IdleTimeout:                    common.Int64(300),
					BackendTcpProxyProtocolVersion: common.Int(1),
				},
			},
			actual: loadbalancer.Listener{
				DefaultBackendSetName: common.String("TCP-443"),
				Protocol:              common.String("TCP"),
				Port:                  common.Int(443),
				SslConfiguration: &loadbalancer.SslConfiguration{
					CertificateName: common.String("cert"),
					VerifyDepth:     common.Int(1),
				},
				ConnectionConfiguration: &loadbalancer.ConnectionConfiguration{
					IdleTimeout:                    common.Int64(300),
					BackendTcpProxyProtocolVersion: common.Int(1),
				},
			},
			expected: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			isListenerChanged := hasListenerChanged(zap.S(), tt.actual, tt.desired)
			if isListenerChanged == tt.expected {
				return
			}
			t.Errorf("expected ListenerChange\n%+v\nbut got\n%+v", tt.expected, isListenerChanged)
		})
	}
}

var (
	testBackendPort    = int(30500)
	testNewBackendPort = int(30600)
)

func TestHasBackendSetChanged(t *testing.T) {
	var testCases = []struct {
		name     string
		desired  loadbalancer.BackendSetDetails
		actual   loadbalancer.BackendSet
		expected bool
	}{
		{
			name: "Policy changes",
			desired: loadbalancer.BackendSetDetails{
				Policy: common.String("desired"),
			},
			actual: loadbalancer.BackendSet{
				Policy: common.String("actual"),
			},
			expected: true,
		},
		{
			name: "HealthChecker present in actual but not in desired",
			desired: loadbalancer.BackendSetDetails{
				Policy: common.String("policy"),
			},
			actual: loadbalancer.BackendSet{
				Policy: common.String("policy"),
				HealthChecker: &loadbalancer.HealthChecker{
					Port: common.Int(20),
				},
			},
			expected: false,
		},
		{
			name: "HealthChecker present in desired but not in actual",
			desired: loadbalancer.BackendSetDetails{
				Policy: common.String("policy"),
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port: common.Int(20),
				},
			},
			actual: loadbalancer.BackendSet{
				Policy: common.String("policy"),
			},
			expected: true,
		},
		{
			name: "HealthChecker port different",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port: common.Int(20),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port: common.Int(30),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker port is present in actual but not present in desired",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					ResponseBodyRegex: common.String("regex"),
					Port:              common.Int(30),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					ResponseBodyRegex: common.String("regex"),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker ResponseBodyRegex present in actual but not present in desired",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port: common.Int(20),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:              common.Int(20),
					ResponseBodyRegex: common.String("actual"),
				},
			},
			expected: false,
		},
		{
			name: "HealthChecker ResponseBodyRegex present in desired but not present in actual",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:              common.Int(20),
					ResponseBodyRegex: common.String("desired"),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port: common.Int(20),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker ResponseBodyRegex changes",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:              common.Int(20),
					ResponseBodyRegex: common.String("desired"),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:              common.Int(20),
					ResponseBodyRegex: common.String("actual"),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker Retries changes",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:    common.Int(20),
					Retries: common.Int(1),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:    common.Int(20),
					Retries: common.Int(2),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker ReturnCode changes",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:       common.Int(20),
					ReturnCode: common.Int(1),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:       common.Int(20),
					ReturnCode: common.Int(2),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker TimeoutInMillis changes",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:            common.Int(20),
					TimeoutInMillis: common.Int(1),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:            common.Int(20),
					TimeoutInMillis: common.Int(2),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker retries changes",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:    common.Int(20),
					Retries: common.Int(2),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:    common.Int(20),
					Retries: common.Int(3),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker IntervalInMillis changes",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:             common.Int(20),
					IntervalInMillis: common.Int(1000),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:             common.Int(20),
					IntervalInMillis: common.Int(300),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker TimeoutInMillis present in desired and not in actual",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:            common.Int(20),
					TimeoutInMillis: common.Int(1),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port: common.Int(20),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker retries present in desired and not in actual",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:    common.Int(20),
					Retries: common.Int(2),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port: common.Int(20),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker IntervalInMillis present in desired and not in actual",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:             common.Int(20),
					IntervalInMillis: common.Int(1000),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port: common.Int(20),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker TimeoutInMillis present in actual and not in desired",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port: common.Int(20),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:            common.Int(20),
					TimeoutInMillis: common.Int(1),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker retries present in actual and not in desired",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port: common.Int(20),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:    common.Int(20),
					Retries: common.Int(2),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker IntervalInMillis present in actual and not in desired",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port: common.Int(20),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:             common.Int(20),
					IntervalInMillis: common.Int(1000),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker UrlPath changes",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:    common.Int(20),
					UrlPath: common.String("/desired"),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:    common.Int(20),
					UrlPath: common.String("/actual"),
				},
			},
			expected: true,
		},
		{
			name: "HealthChecker Protocol changes",
			desired: loadbalancer.BackendSetDetails{
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:     common.Int(20),
					Protocol: common.String("desired"),
				},
			},
			actual: loadbalancer.BackendSet{
				HealthChecker: &loadbalancer.HealthChecker{
					Port:     common.Int(20),
					Protocol: common.String("actual"),
				},
			},
			expected: true,
		},
		{
			name: "no changes",
			desired: loadbalancer.BackendSetDetails{
				Policy: common.String("policy"),
				HealthChecker: &loadbalancer.HealthCheckerDetails{
					Port:     common.Int(20),
					Protocol: common.String("Protocol"),
					Retries:  common.Int(2),
				},
			},
			actual: loadbalancer.BackendSet{
				Policy: common.String("policy"),
				HealthChecker: &loadbalancer.HealthChecker{
					Port:     common.Int(20),
					Protocol: common.String("Protocol"),
					Retries:  common.Int(2),
				},
			},
			expected: false,
		},
		{
			name: "no change - nodeport",
			desired: loadbalancer.BackendSetDetails{
				Policy: common.String("policy"),
				Backends: []loadbalancer.BackendDetails{
					{
						Port: &testBackendPort,
					},
				},
			},
			actual: loadbalancer.BackendSet{
				Policy: common.String("policy"),
				Backends: []loadbalancer.Backend{
					{
						Port: &testBackendPort,
					},
				},
			},
			expected: false,
		},
		{
			name: "nodeport change",
			desired: loadbalancer.BackendSetDetails{
				Policy: common.String("policy"),
				Backends: []loadbalancer.BackendDetails{
					{
						Port: &testBackendPort,
					},
				},
			},
			actual: loadbalancer.BackendSet{
				Policy: common.String("policy"),
				Backends: []loadbalancer.Backend{
					{
						Port: &testNewBackendPort,
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			isListenerChanged := hasBackendSetChanged(zap.S(), tt.actual, tt.desired)
			if isListenerChanged == tt.expected {
				return
			}
			t.Errorf("expected BackendSetChanges\n%+v\nbut got\n%+v", tt.expected, isListenerChanged)
		})
	}
}

func TestGetHealthCheckerChanges(t *testing.T) {
	var testCases = []struct {
		name     string
		desired  loadbalancer.HealthCheckerDetails
		actual   loadbalancer.HealthChecker
		expected []string
	}{
		{
			name: "All Changed",
			desired: loadbalancer.HealthCheckerDetails{
				Port:              common.Int(20),
				ResponseBodyRegex: common.String("desired"),
				Retries:           common.Int(3),
				ReturnCode:        common.Int(200),
				TimeoutInMillis:   common.Int(200),
				UrlPath:           common.String("/desired"),
				Protocol:          common.String("HTTP"),
			},
			actual: loadbalancer.HealthChecker{
				Port:              common.Int(25),
				ResponseBodyRegex: common.String("actual"),
				Retries:           common.Int(2),
				TimeoutInMillis:   common.Int(300),
				UrlPath:           common.String("/actual"),
				Protocol:          common.String("TCP"),
			},
			expected: []string{
				fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:Port", 25, 20),
				fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:ResponseBodyRegex", "actual", "desired"),
				fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:Retries", 2, 3),
				fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:ReturnCode", 0, 200),
				fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:TimeoutInMillis", 200, 300),
				fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:UrlPath", "actual", "desired"),
				fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:UrlPath", "TCP", "HTTP"),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			changes := getHealthCheckerChanges(&tt.actual, &tt.desired)
			if len(changes) == len(tt.expected) {
				return
			}
			if !reflect.DeepEqual(changes, tt.expected) {
				t.Errorf("expected HealthCheckerChanges\n%+v\nbut got\n%+v", tt.expected, changes)
			}
		})
	}
}

func TestGetSSLConfigurationChanges(t *testing.T) {
	var testCases = []struct {
		name     string
		desired  loadbalancer.SslConfigurationDetails
		actual   loadbalancer.SslConfiguration
		expected []string
	}{
		{
			name: "All Changed",
			desired: loadbalancer.SslConfigurationDetails{
				CertificateName:       common.String("desired"),
				VerifyDepth:           common.Int(1),
				VerifyPeerCertificate: common.Bool(true),
			},
			actual: loadbalancer.SslConfiguration{
				CertificateName:       common.String("actual"),
				VerifyPeerCertificate: common.Bool(false),
			},
			expected: []string{
				fmt.Sprintf(changeFmtStr, "Listener:SSLConfiguration:CertificateName", "actual", "desired"),
				fmt.Sprintf(changeFmtStr, "Listener:SSLConfiguration:VerifyDepth", 0, 1),
				fmt.Sprintf(changeFmtStr, "Listener:SSLConfiguration:CertificateName", false, true),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			changes := getSSLConfigurationChanges(&tt.actual, &tt.desired)
			if len(changes) == len(tt.expected) {
				return
			}
			if !reflect.DeepEqual(changes, tt.expected) {
				t.Errorf("expected SSLConfigurationChanges\n%+v\nbut got\n%+v", tt.expected, changes)
			}
		})
	}
}

func TestGetConnectionConfigurationChanges(t *testing.T) {
	var testCases = []struct {
		name     string
		desired  loadbalancer.ConnectionConfiguration
		actual   loadbalancer.ConnectionConfiguration
		expected []string
	}{
		{
			name: "All Changed",
			desired: loadbalancer.ConnectionConfiguration{
				IdleTimeout:                    common.Int64(300),
				BackendTcpProxyProtocolVersion: common.Int(2),
			},
			actual: loadbalancer.ConnectionConfiguration{
				IdleTimeout:                    common.Int64(400),
				BackendTcpProxyProtocolVersion: common.Int(3),
			},
			expected: []string{
				fmt.Sprintf(changeFmtStr, "Listner:ConnectionConfiguration:IdleTimeout", 400, 300),
				fmt.Sprintf(changeFmtStr, "Listner:ConnectionConfiguration:BackendTcpProxyProtocolVersion", 3, 2),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			changes := getConnectionConfigurationChanges(&tt.actual, &tt.desired)
			if len(changes) == len(tt.expected) {
				return
			}
			if !reflect.DeepEqual(changes, tt.expected) {
				t.Errorf("expected ConnectionConfigurationChanges\n%+v\nbut got\n%+v", tt.expected, changes)
			}
		})
	}
}

var (
	hundredMbps   = "100Mbps"
	flexibleShape = "flexible"
	flexShape10   = 10
	flexShape100  = 100
	flexShape1000 = 1000
)

func TestHasLoadbalancerShapeChanged(t *testing.T) {
	var testCases = []struct {
		name     string
		lb       loadbalancer.LoadBalancer
		lbSpec   LBSpec
		expected bool
	}{
		{
			name: "No Changes",
			lb: loadbalancer.LoadBalancer{
				ShapeName: &hundredMbps,
			},
			lbSpec: LBSpec{
				Shape: "100Mbps",
			},
			expected: false,
		},
		{
			name: "No Changes flex",
			lb: loadbalancer.LoadBalancer{
				ShapeName: &flexibleShape,
				ShapeDetails: &loadbalancer.ShapeDetails{
					MinimumBandwidthInMbps: &flexShape10,
					MaximumBandwidthInMbps: &flexShape100,
				},
			},
			lbSpec: LBSpec{
				Shape:   "flexible",
				FlexMin: &flexShape10,
				FlexMax: &flexShape100,
			},
			expected: false,
		},
		{
			name: "Change fixed shape",
			lb: loadbalancer.LoadBalancer{
				ShapeName: &hundredMbps,
			},
			lbSpec: LBSpec{
				Shape: "400Mbps",
			},
			expected: true,
		},
		{
			name: "Change flex shape",
			lb: loadbalancer.LoadBalancer{
				ShapeName: &flexibleShape,
				ShapeDetails: &loadbalancer.ShapeDetails{
					MinimumBandwidthInMbps: &flexShape10,
					MaximumBandwidthInMbps: &flexShape100,
				},
			},
			lbSpec: LBSpec{
				Shape:   "flexible",
				FlexMin: &flexShape100,
				FlexMax: &flexShape1000,
			},
			expected: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			changed := hasLoadbalancerShapeChanged(context.TODO(), &tt.lbSpec, &tt.lb)
			if changed != tt.expected {
				t.Errorf("expected hasLBShapeChanged to be %+v\nbut got\n%+v", tt.expected, changed)
			}
		})
	}
}
