// Copyright 2017 The OCI Cloud Controller Manager Authors
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

	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	baremetal "github.com/oracle/bmcs-go-sdk"
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
					actionType: Create,
					BackendSet: baremetal.BackendSet{Name: "TCP-80"},
				},
				&BackendSetAction{
					actionType: Create,
					BackendSet: baremetal.BackendSet{Name: "TCP-443"},
				},
			},
			listenerActions: []Action{
				&ListenerAction{
					actionType: Create,
					Listener:   baremetal.Listener{Name: "TCP-443"},
				},
				&ListenerAction{
					actionType: Create,
					Listener:   baremetal.Listener{Name: "TCP-80"},
				},
			},
			expected: []Action{
				&BackendSetAction{
					actionType: Create,
					BackendSet: baremetal.BackendSet{Name: "TCP-443"},
				},
				&ListenerAction{
					actionType: Create,
					Listener:   baremetal.Listener{Name: "TCP-443"},
				},
				&BackendSetAction{
					actionType: Create,
					BackendSet: baremetal.BackendSet{Name: "TCP-80"},
				},
				&ListenerAction{
					actionType: Create,
					Listener:   baremetal.Listener{Name: "TCP-80"},
				},
			},
		},
		"update": {
			backendSetActions: []Action{
				&BackendSetAction{
					actionType: Update,
					BackendSet: baremetal.BackendSet{Name: "TCP-80"},
				},
				&BackendSetAction{
					actionType: Update,
					BackendSet: baremetal.BackendSet{Name: "TCP-443"},
				},
			},
			listenerActions: []Action{
				&ListenerAction{
					actionType: Update,
					Listener:   baremetal.Listener{Name: "TCP-443"},
				},
				&ListenerAction{
					actionType: Update,
					Listener:   baremetal.Listener{Name: "TCP-80"},
				},
			},
			expected: []Action{
				&ListenerAction{
					actionType: Update,
					Listener:   baremetal.Listener{Name: "TCP-443"},
				},
				&BackendSetAction{
					actionType: Update,
					BackendSet: baremetal.BackendSet{Name: "TCP-443"},
				},
				&ListenerAction{
					actionType: Update,
					Listener:   baremetal.Listener{Name: "TCP-80"},
				},
				&BackendSetAction{
					actionType: Update,
					BackendSet: baremetal.BackendSet{Name: "TCP-80"},
				},
			},
		},
		"delete": {
			backendSetActions: []Action{
				&BackendSetAction{
					actionType: Delete,
					BackendSet: baremetal.BackendSet{Name: "TCP-80"},
				},
				&BackendSetAction{
					actionType: Delete,
					BackendSet: baremetal.BackendSet{Name: "TCP-443"},
				},
			},
			listenerActions: []Action{
				&ListenerAction{
					actionType: Delete,
					Listener:   baremetal.Listener{Name: "TCP-443"},
				},
				&ListenerAction{
					actionType: Delete,
					Listener:   baremetal.Listener{Name: "TCP-80"},
				},
			},
			expected: []Action{
				&ListenerAction{
					actionType: Delete,
					Listener:   baremetal.Listener{Name: "TCP-443"},
				},
				&BackendSetAction{
					actionType: Delete,
					BackendSet: baremetal.BackendSet{Name: "TCP-443"},
				},
				&ListenerAction{
					actionType: Delete,
					Listener:   baremetal.Listener{Name: "TCP-80"},
				},
				&BackendSetAction{
					actionType: Delete,
					BackendSet: baremetal.BackendSet{Name: "TCP-80"},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := sortAndCombineActions(tc.backendSetActions, tc.listenerActions)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected\n%+v\nbut got\n%+v", tc.expected, result)
			}
		})
	}
}

func TestGetBackendSetChanges(t *testing.T) {
	var testCases = []struct {
		name     string
		desired  map[string]baremetal.BackendSet
		actual   map[string]baremetal.BackendSet
		expected []Action
	}{
		{
			name: "create backendset",
			desired: map[string]baremetal.BackendSet{
				"one": baremetal.BackendSet{
					Backends: []baremetal.Backend{
						{IPAddress: "0.0.0.0", Port: 80},
					},
				},
				"two": baremetal.BackendSet{
					Name: "two",
					Backends: []baremetal.Backend{
						{IPAddress: "0.0.0.3", Port: 80},
						{IPAddress: "0.0.0.4", Port: 80},
					},
				},
			},
			actual: map[string]baremetal.BackendSet{
				"one": baremetal.BackendSet{
					Backends: []baremetal.Backend{
						{IPAddress: "0.0.0.0", Port: 80},
					},
				},
			},
			expected: []Action{
				&BackendSetAction{
					actionType: Create,
					BackendSet: baremetal.BackendSet{
						Name: "two",
						Backends: []baremetal.Backend{
							{IPAddress: "0.0.0.3", Port: 80},
							{IPAddress: "0.0.0.4", Port: 80},
						},
					},
				},
			},
		},
		{
			name: "update backendset - add backend",
			desired: map[string]baremetal.BackendSet{
				"one": baremetal.BackendSet{
					Backends: []baremetal.Backend{
						{IPAddress: "0.0.0.0", Port: 80},
						{IPAddress: "0.0.0.1", Port: 80},
					},
				},
			},
			actual: map[string]baremetal.BackendSet{
				"one": baremetal.BackendSet{
					Backends: []baremetal.Backend{
						{IPAddress: "0.0.0.0", Port: 80},
					},
				},
			},
			expected: []Action{
				&BackendSetAction{
					actionType: Update,
					BackendSet: baremetal.BackendSet{
						Backends: []baremetal.Backend{
							{IPAddress: "0.0.0.0", Port: 80},
							{IPAddress: "0.0.0.1", Port: 80},
						},
					},
				},
			},
		},
		{
			name: "update backendset - remove backend",
			desired: map[string]baremetal.BackendSet{
				"one": baremetal.BackendSet{
					Backends: []baremetal.Backend{
						{IPAddress: "0.0.0.0", Port: 80},
					},
				},
			},
			actual: map[string]baremetal.BackendSet{
				"one": baremetal.BackendSet{
					Backends: []baremetal.Backend{
						{IPAddress: "0.0.0.0", Port: 80},
						{IPAddress: "0.0.0.1", Port: 80},
					},
				},
			},
			expected: []Action{
				&BackendSetAction{
					actionType: Update,
					BackendSet: baremetal.BackendSet{
						Backends: []baremetal.Backend{
							{IPAddress: "0.0.0.0", Port: 80},
						},
					},
				},
			},
		},
		{
			name:    "remove backendset",
			desired: map[string]baremetal.BackendSet{},
			actual: map[string]baremetal.BackendSet{
				"one": baremetal.BackendSet{
					Backends: []baremetal.Backend{
						{IPAddress: "0.0.0.0", Port: 80},
					},
				},
			},
			expected: []Action{
				&BackendSetAction{
					actionType: Delete,
					BackendSet: baremetal.BackendSet{
						Backends: []baremetal.Backend{
							{IPAddress: "0.0.0.0", Port: 80},
						},
					},
				},
			},
		},
		{
			name: "no change",
			desired: map[string]baremetal.BackendSet{
				"one": baremetal.BackendSet{
					Backends: []baremetal.Backend{
						{IPAddress: "0.0.0.0", Port: 80},
					},
				},
			},
			actual: map[string]baremetal.BackendSet{
				"one": baremetal.BackendSet{
					Backends: []baremetal.Backend{
						{IPAddress: "0.0.0.0", Port: 80},
					},
				},
			},
			expected: []Action{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			changes := getBackendSetChanges(tt.actual, tt.desired)
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
		desired  map[string]baremetal.Listener
		actual   map[string]baremetal.Listener
		expected []Action
	}{
		{
			name: "create listener",
			desired: map[string]baremetal.Listener{"TCP-443": baremetal.Listener{
				Name: "TCP-443",
				DefaultBackendSetName: "TCP-443",
				Protocol:              "TCP",
				Port:                  443,
			}},
			actual: map[string]baremetal.Listener{},
			expected: []Action{
				&ListenerAction{
					actionType: Create,
					Listener: baremetal.Listener{
						Name: "TCP-443",
						DefaultBackendSetName: "TCP-443",
						Protocol:              "TCP",
						Port:                  443,
					},
				},
			},
		},
		{
			name: "add listener",
			desired: map[string]baremetal.Listener{
				"TCP-80": baremetal.Listener{
					Name: "TCP-80",
					DefaultBackendSetName: "TCP-80",
					Protocol:              "TCP",
					Port:                  80,
				},
				"TCP-443": baremetal.Listener{
					Name: "TCP-443",
					DefaultBackendSetName: "TCP-443",
					Protocol:              "TCP",
					Port:                  443,
				},
			},
			actual: map[string]baremetal.Listener{
				"TCP-80": baremetal.Listener{
					Name: "TCP-80",
					DefaultBackendSetName: "TCP-80",
					Protocol:              "TCP",
					Port:                  80,
				},
			},
			expected: []Action{
				&ListenerAction{
					actionType: Create,
					Listener: baremetal.Listener{
						Name: "TCP-443",
						DefaultBackendSetName: "TCP-443",
						Protocol:              "TCP",
						Port:                  443,
					},
				},
			},
		},
		{
			name: "remove listener",
			desired: map[string]baremetal.Listener{
				"TCP-80": baremetal.Listener{
					Name: "TCP-80",
					DefaultBackendSetName: "TCP-80",
					Protocol:              "TCP",
					Port:                  80,
				},
			},
			actual: map[string]baremetal.Listener{
				"TCP-443": baremetal.Listener{
					Name: "TCP-443",
					DefaultBackendSetName: "TCP-443",
					Protocol:              "TCP",
					Port:                  443,
				},
				"TCP-80": baremetal.Listener{
					Name: "TCP-80",
					DefaultBackendSetName: "TCP-80",
					Protocol:              "TCP",
					Port:                  80,
				},
			},
			expected: []Action{
				&ListenerAction{
					actionType: Delete,
					Listener: baremetal.Listener{
						Name: "TCP-443",
						DefaultBackendSetName: "TCP-443",
						Protocol:              "TCP",
						Port:                  443,
					},
				},
			},
		},
		{
			name: "no change",
			desired: map[string]baremetal.Listener{
				"TCP-80": baremetal.Listener{
					Name: "TCP-80",
					DefaultBackendSetName: "TCP-80",
					Protocol:              "TCP",
					Port:                  80,
				},
			},
			actual: map[string]baremetal.Listener{
				"TCP-80": baremetal.Listener{
					Name: "TCP-80",
					DefaultBackendSetName: "TCP-80",
					Protocol:              "TCP",
					Port:                  80,
				},
			},
			expected: []Action{},
		},
		{
			name: "ssl config change",
			desired: map[string]baremetal.Listener{
				"TCP-80": baremetal.Listener{
					Name: "TCP-80",
					DefaultBackendSetName: "TCP-80",
					Protocol:              "TCP",
					Port:                  80,
					SSLConfig: &baremetal.SSLConfiguration{
						CertificateName: "desired",
					},
				},
			},
			actual: map[string]baremetal.Listener{
				"TCP-80": baremetal.Listener{
					Name: "TCP-80",
					DefaultBackendSetName: "TCP-80",
					Protocol:              "TCP",
					Port:                  80,
					SSLConfig: &baremetal.SSLConfiguration{
						CertificateName: "actual",
					},
				},
			},
			expected: []Action{
				&ListenerAction{
					actionType: Update,
					Listener: baremetal.Listener{
						Name: "TCP-80",
						DefaultBackendSetName: "TCP-80",
						Protocol:              "TCP",
						Port:                  80,
						SSLConfig: &baremetal.SSLConfiguration{
							CertificateName: "desired",
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
		expected    map[int]bool
	}{
		{
			name:        "empty",
			annotations: map[string]string{},
			expected:    nil,
		}, {
			name:        "empty string",
			annotations: map[string]string{"service.beta.kubernetes.io/oci-load-balancer-ssl-ports": ""},
			expected:    nil,
		}, {
			name:        "443",
			annotations: map[string]string{"service.beta.kubernetes.io/oci-load-balancer-ssl-ports": "443"},
			expected:    map[int]bool{443: true},
		}, {
			name:        "1,2,3",
			annotations: map[string]string{"service.beta.kubernetes.io/oci-load-balancer-ssl-ports": "1,2,3"},
			expected:    map[int]bool{1: true, 2: true, 3: true},
		}, {
			name:        "1, 2, 3",
			annotations: map[string]string{"service.beta.kubernetes.io/oci-load-balancer-ssl-ports": "1, 2, 3"},
			expected:    map[int]bool{1: true, 2: true, 3: true},
		}, {
			name:        "not-an-integer",
			annotations: map[string]string{"service.beta.kubernetes.io/oci-load-balancer-ssl-ports": "not-an-integer"},
			expected:    nil,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			sslEnabledPorts, _ := getSSLEnabledPorts(tt.annotations)
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
