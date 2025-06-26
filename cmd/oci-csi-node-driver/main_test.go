// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
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

package main

import (
	"os"
	"testing"
)

func Test_IsLustreDriverEnabled(t *testing.T) {
	tests := []struct {
		envValue string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"TrUe", true},
		{"false", false},
		{"", false},
		{"random", false},
	}

	for _, tc := range tests {
		// Set or unset the environment variable based on the test case.
		if tc.envValue == "" {
			os.Unsetenv("LUSTRE_DRIVER_ENABLED")
		} else {
			os.Setenv("LUSTRE_DRIVER_ENABLED", tc.envValue)
		}

		// Our logic under test: compare the environment variable with "true" (case-insensitive).
		enableLustreDriver := IsLustreDriverEnabled()

		if enableLustreDriver != tc.expected {
			t.Errorf("For LUSTRE_DRIVER_ENABLED=%q, expected %v but got %v",
				tc.envValue, tc.expected, enableLustreDriver)
		}
	}
}
