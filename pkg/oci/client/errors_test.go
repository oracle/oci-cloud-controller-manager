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
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
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
func (m mockServiceError) Error() string {
	return m.Message
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

func TestIsSystemTagNotFoundOrNotAuthorisedError(t *testing.T) {
	systemTagError := mockServiceError{
		StatusCode: http.StatusBadRequest,
		Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
		Message:    "The following tag namespaces / keys are not authorized or not found: 'orcl-containerengine'",
	}
	systemTagError2 := mockServiceError{
		StatusCode: http.StatusBadRequest,
		Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
		Message:    "The following tag namespaces / keys are not authorized or not found: TagDefinition cluster_foobar in TagNamespace orcl-containerengine does not exists.\\n",
	}
	userDefinedTagError1 := mockServiceError{
		StatusCode: http.StatusBadRequest,
		Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
		Message:    "The following tag namespaces / keys are not authorized or not found: 'foobar-namespace'",
	}
	userDefinedTagError2 := mockServiceError{
		StatusCode: http.StatusBadRequest,
		Code:       HTTP400RelatedResourceNotAuthorizedOrNotFoundCode,
		Message:    "TagNamespace orcl-foobar does not exists.\\nTagNamespace orcl-foobar-name does not exists.\\n",
	}
	tests := map[string]struct {
		se               mockServiceError
		wrappedError     error
		expectIsTagError bool
	}{
		"base case": {
			wrappedError:     errors.WithMessage(systemTagError, "taggin failure"),
			expectIsTagError: true,
		},
		"three layer wrapping - resource tracking system tag error": {
			wrappedError:     errors.Wrap(errors.Wrap(errors.WithMessage(systemTagError, "taggin failure"), "first layer"), "second layer"),
			expectIsTagError: true,
		},
		"wrapping with stack trace - resource tracking system tag error": {
			wrappedError:     errors.WithStack(errors.Wrap(errors.WithMessage(systemTagError2, "taggin failure"), "first layer")),
			expectIsTagError: true,
		},
		"three layer wrapping - user defined tag error": {
			wrappedError:     errors.Wrap(errors.Wrap(errors.WithMessage(userDefinedTagError1, "taggin failure"), "first layer"), "second layer"),
			expectIsTagError: false,
		},
		"wrapping with stack trace - user defined tag error": {
			wrappedError:     errors.WithStack(errors.Wrap(errors.WithMessage(userDefinedTagError2, "taggin failure"), "first layer")),
			expectIsTagError: false,
		},
		"not a service error": {
			wrappedError:     errors.Wrap(fmt.Errorf("not a service error"), "precheck error"),
			expectIsTagError: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualResult := IsSystemTagNotFoundOrNotAuthorisedError(zap.S(), test.wrappedError)
			if actualResult != test.expectIsTagError {
				t.Errorf("expected %t but got %t", actualResult, test.expectIsTagError)
			}
		})
	}
}
