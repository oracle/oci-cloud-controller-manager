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

package client

import (
	"testing"

	baremetal "github.com/oracle/bmcs-go-sdk"
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
