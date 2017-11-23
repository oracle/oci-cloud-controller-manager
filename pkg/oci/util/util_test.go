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

package util

import "testing"

func TestMapProviderIDToInstanceID(t *testing.T) {
	testCases := map[string]struct {
		providerID string
		expected   string
	}{
		"no cloud prefix": {
			providerID: "testid",
			expected:   "testid",
		},
		"cloud prefix": {
			providerID: providerPrefix + "testid",
			expected:   "testid",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := MapProviderIDToInstanceID(tc.providerID)
			if result != tc.expected {
				t.Errorf("Expected instance id %q, but got %q", tc.expected, result)
			}
		})
	}
}
