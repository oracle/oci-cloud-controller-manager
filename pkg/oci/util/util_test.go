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

package util

import "testing"

func TestMapProviderIDToInstanceID(t *testing.T) {
	testCases := map[string]struct {
		providerID string
		instanceID string
		error      bool
	}{
		"no cloud prefix": {
			providerID: "testid",
			instanceID: "testid",
			error:      false,
		},
		"cloud prefix": {
			providerID: providerPrefix + "testid",
			instanceID: "testid",
			error:      false,
		},
		"empty string": {
			providerID: "",
			instanceID: "",
			error:      true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := MapProviderIDToInstanceID(tc.providerID)
			if result != tc.instanceID {
				t.Errorf("Expected instance id %q, but got %q", tc.instanceID, result)
			}
			if (err == nil && tc.error) || (!tc.error && err != nil) {
				t.Errorf("Expected an error condition for input %q, but did no receive one; or received one, when not expecting", tc.providerID)
			}
		})
	}
}
