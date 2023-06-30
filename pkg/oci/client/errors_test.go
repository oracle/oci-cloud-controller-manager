// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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
	"net/http"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
)

type mockServiceError struct {
	StatusCode   int
	Code         string
	Message      string
	OpcRequestID string
}

func (m mockServiceError) GetHTTPStatusCode() int {
	return m.StatusCode
}

func (m mockServiceError) GetMessage() string {
	return m.Message
}

func (m mockServiceError) GetCode() string {
	return m.Code
}

func (m mockServiceError) GetOpcRequestID() string {
	return m.OpcRequestID
}

func TestIsRetryableServiceError(t *testing.T) {
	testCases := map[string]struct {
		error    common.ServiceError
		expected bool
	}{
		"HTTP400RelatedResourceNotAuthorizedOrNotFound": {
			error: mockServiceError{
				StatusCode: http.StatusBadRequest,
				Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
			},
			expected: true,
		},
		"HTTP401NotAuthenticated": {
			error: mockServiceError{
				StatusCode: http.StatusUnauthorized,
				Code:       HTTP401NotAuthenticatedCode,
			},
			expected: true,
		},
		"HTTP404NotAuthorizedOrNotFound": {
			error: mockServiceError{
				StatusCode: http.StatusNotFound,
				Code:       HTTP404NotAuthorizedOrNotFoundCode,
			},
			expected: true,
		},
		"HTTP409IncorrectState": {
			error: mockServiceError{
				StatusCode: http.StatusConflict,
				Code:       HTTP409IncorrectStateCode,
			},
			expected: true,
		},
		"HTTP409NotAuthorizedOrResourceAlreadyExists": {
			error: mockServiceError{
				StatusCode: http.StatusConflict,
				Code:       HTTP409NotAuthorizedOrResourceAlreadyExistsCode,
			},
			expected: true,
		},
		"HTTP429TooManyRequests": {
			error: mockServiceError{
				StatusCode: http.StatusTooManyRequests,
				Code:       HTTP429TooManyRequestsCode,
			},
			expected: true,
		},
		"HTTP500InternalServerError": {
			error: mockServiceError{
				StatusCode: http.StatusInternalServerError,
				Code:       HTTP500InternalServerErrorCode,
			},
			expected: true,
		},
		"NonRetryable": {
			error: mockServiceError{
				StatusCode: http.StatusConflict,
				Code:       HTTP500InternalServerErrorCode,
			},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := isRetryableServiceError(tc.error)
			if result != tc.expected {
				t.Errorf("isRetryableServiceError(%v) = %v ; wanted %v", tc.error, result, tc.expected)
			}
		})
	}

}
