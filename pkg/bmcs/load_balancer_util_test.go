// Copyright 2017 The Oracle Kubernetes Cloud Controller Manager Authors
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
package bmcs

import (
	"reflect"
	"testing"

	baremetal "github.com/oracle/bmcs-go-sdk"
)

func TestGetBackendSetChanges(t *testing.T) {
	var testCases = []struct {
		name     string
		desired  map[string]baremetal.BackendSet
		actual   map[string]baremetal.BackendSet
		expected []BackendSetAction
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
			expected: []BackendSetAction{
				{
					Type: Create,
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
			expected: []BackendSetAction{
				{
					Type: Update,
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
			expected: []BackendSetAction{
				{
					Type: Update,
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
			expected: []BackendSetAction{
				{
					Type: Delete,
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
			expected: []BackendSetAction{},
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
		expected []ListenerAction
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
			expected: []ListenerAction{
				{
					Type: Create,
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
			expected: []ListenerAction{
				{
					Type: Create,
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
			expected: []ListenerAction{
				{
					Type: Delete,
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
			expected: []ListenerAction{},
		},
		{
			name: "ssl config chnage",
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
			expected: []ListenerAction{},
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
