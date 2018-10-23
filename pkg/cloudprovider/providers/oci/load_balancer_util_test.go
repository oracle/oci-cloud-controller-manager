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
	"os"
	"reflect"
	"testing"

	"go.uber.org/zap"
	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/loadbalancer"
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
			},
			expected: []Action{
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
					name:       "TCP-443",
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
					name:       "TCP-443",
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
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
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
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
					},
				},
			},
			expected: []Action{
				&BackendSetAction{
					name:       "one",
					actionType: Update,
					BackendSet: loadbalancer.BackendSetDetails{
						Backends: []loadbalancer.BackendDetails{
							{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
							{IpAddress: common.String("0.0.0.1"), Port: common.Int(80)},
						},
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
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
						{IpAddress: common.String("0.0.0.1"), Port: common.Int(80)},
					},
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
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
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
						{IpAddress: common.String("0.0.0.0"), Port: common.Int(80)},
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
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			changes := getListenerChanges(tt.actual, tt.desired)
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
