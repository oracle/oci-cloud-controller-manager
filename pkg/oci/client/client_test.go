// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.
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

package client

import (
	"errors"
	"reflect"
	"testing"

	baremetal "github.com/oracle/bmcs-go-sdk"

	api "k8s.io/api/core/v1"
)

func TestInstanceTerminalState(t *testing.T) {
	testCases := map[string]struct {
		state    string
		expected bool
	}{
		"not terminal - running": {
			state:    baremetal.ResourceRunning,
			expected: false,
		},
		"not terminal - stopped": {
			state:    baremetal.ResourceStopped,
			expected: false,
		},
		"is terminal - terminating": {
			state:    baremetal.ResourceTerminating,
			expected: true,
		},
		"is terminal - terminated": {
			state:    baremetal.ResourceTerminated,
			expected: true,
		},
		"is terminal - unknown": {
			state:    "UNKNOWN",
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := IsInstanceInTerminalState(&baremetal.Instance{
				State: tc.state,
			})
			if result != tc.expected {
				t.Errorf("IsInstanceInTerminalState(%q) = %v ; wanted %v", tc.state, result, tc.expected)
			}
		})
	}
}

func TestExtractNodeAddressesFromVNIC(t *testing.T) {
	testCases := []struct {
		name string
		in   *baremetal.Vnic
		out  []api.NodeAddress
		err  error
	}{
		{
			name: "basic-complete",
			in: &baremetal.Vnic{
				PrivateIPAddress: "10.0.0.1",
				PublicIPAddress:  "0.0.0.1",
			},
			out: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeInternalIP, Address: "10.0.0.1"},
				api.NodeAddress{Type: api.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-external-ip",
			in: &baremetal.Vnic{
				PrivateIPAddress: "10.0.0.1",
			},
			out: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeInternalIP, Address: "10.0.0.1"},
			},
			err: nil,
		},
		{
			name: "no-internal-ip",
			in: &baremetal.Vnic{
				PublicIPAddress: "0.0.0.1",
			},
			out: []api.NodeAddress{
				api.NodeAddress{Type: api.NodeExternalIP, Address: "0.0.0.1"},
			},
			err: nil,
		},
		{
			name: "invalid-external-ip",
			in: &baremetal.Vnic{
				PublicIPAddress: "0.0.0.",
			},
			out: nil,
			err: errors.New(`instance has invalid public address: "0.0.0."`),
		},
		{
			name: "invalid-external-ip",
			in: &baremetal.Vnic{
				PrivateIPAddress: "10.0.0.",
			},
			out: nil,
			err: errors.New(`instance has invalid private address: "10.0.0."`),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractNodeAddressesFromVNIC(tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("extractNodeAddressesFromVNIC(%+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("extractNodeAddressesFromVNIC(%+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}
