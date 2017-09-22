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

package client

import (
	"errors"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "search-error-not-found",
			err:      &SearchError{NotFound: true},
			expected: true,
		},
		{
			name:     "search-error-found",
			err:      &SearchError{NotFound: false},
			expected: false,
		},
		{
			name:     "generic-error",
			err:      errors.New("something erroneous"),
			expected: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			isNotFoundErr := IsNotFound(tt.err)
			if isNotFoundErr != tt.expected {
				t.Errorf("IsNotFound(%+v) => %t, wanted %t", tt.err, isNotFoundErr, tt.expected)
			}
		})
	}
}
