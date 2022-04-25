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
	"math"
	"net/http"
	"time"

	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/pkg/errors"
)

var errNotFound = errors.New("not found")

// HTTP Error Types
const (
	HTTP400RelatedResourceNotAuthorizedOrNotFoundCode = "RelatedResourceNotAuthorizedOrNotFound"
	HTTP401NotAuthenticatedCode                       = "NotAuthenticated"
	HTTP404NotAuthorizedOrNotFoundCode                = "NotAuthorizedOrNotFound"
	HTTP409IncorrectStateCode                         = "IncorrectState"
	HTTP409NotAuthorizedOrResourceAlreadyExistsCode   = "NotAuthorizedOrResourceAlreadyExists"
	HTTP429TooManyRequestsCode                        = "TooManyRequests"
	HTTP500InternalServerErrorCode                    = "InternalServerError"
)

// IsNotFound returns true if the given error indicates that a resource could
// not be found.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	err = errors.Cause(err)
	if err == errNotFound {
		return true
	}

	serviceErr, ok := common.IsServiceError(err)
	return ok && serviceErr.GetHTTPStatusCode() == http.StatusNotFound
}

//IsRetryable returns true if the given error is retriable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	err = errors.Cause(err)
	serviceErr, ok := common.IsServiceError(err)
	if !ok {
		return false
	}

	return isRetryableServiceError(serviceErr)
}

func isRetryableServiceError(serviceErr common.ServiceError) bool {
	return ((serviceErr.GetHTTPStatusCode() == http.StatusBadRequest) && (serviceErr.GetCode() == HTTP400RelatedResourceNotAuthorizedOrNotFoundCode)) ||
		((serviceErr.GetHTTPStatusCode() == http.StatusUnauthorized) && (serviceErr.GetCode() == HTTP401NotAuthenticatedCode)) ||
		((serviceErr.GetHTTPStatusCode() == http.StatusNotFound) && (serviceErr.GetCode() == HTTP404NotAuthorizedOrNotFoundCode)) ||
		((serviceErr.GetHTTPStatusCode() == http.StatusConflict) && (serviceErr.GetCode() == HTTP409IncorrectStateCode)) ||
		((serviceErr.GetHTTPStatusCode() == http.StatusConflict) && (serviceErr.GetCode() == HTTP409NotAuthorizedOrResourceAlreadyExistsCode)) ||
		((serviceErr.GetHTTPStatusCode() == http.StatusTooManyRequests) && (serviceErr.GetCode() == HTTP429TooManyRequestsCode)) ||
		((serviceErr.GetHTTPStatusCode() == http.StatusInternalServerError) && (serviceErr.GetCode() == HTTP500InternalServerErrorCode))
}

// RateLimitError produces an Errorf for rate limiting.
func RateLimitError(isWrite bool, opName string) error {
	opType := "read"
	if isWrite {
		opType = "write"
	}
	return errors.Errorf("rate limited(%s) for operation: %s", opType, opName)
}

func newRetryPolicy() *common.RetryPolicy {
	return NewRetryPolicyWithMaxAttempts(uint(2))
}

// NewRetryPolicyWithMaxAttempts returns a RetryPolicy with the specified max retryAttempts
func NewRetryPolicyWithMaxAttempts(retryAttempts uint) *common.RetryPolicy {
	isRetryableOperation := func(r common.OCIOperationResponse) bool {
		return IsRetryable(r.Error)
	}

	nextDuration := func(r common.OCIOperationResponse) time.Duration {
		// you might want wait longer for next retry when your previous one failed
		// this function will return the duration as:
		// 1s, 2s, 4s, 8s, 16s, 32s, 64s etc...
		return time.Duration(math.Pow(float64(2), float64(r.AttemptNumber-1))) * time.Second
	}

	policy := common.NewRetryPolicy(
		retryAttempts, isRetryableOperation, nextDuration,
	)
	return &policy
}
