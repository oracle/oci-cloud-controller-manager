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
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var errNotFound = errors.New("not found")
var rateLimitRetryRNG = rand.New(rand.NewSource(time.Now().UnixNano()))
var rateLimitRetryRNGMu sync.Mutex

var rateLimitRetryMaxAttempts uint = 6

var rateLimitRetryNextDuration = func(attempt uint) time.Duration {
	base := math.Pow(2, float64(attempt-1))
	rateLimitRetryRNGMu.Lock()
	jitter := 1 + (rateLimitRetryRNG.Float64()-0.5)*0.2
	rateLimitRetryRNGMu.Unlock()
	return time.Duration(base * jitter * float64(time.Second))
}

/*
Addition of system tags can fail due to permission issue while API returns error code: RelatedResourceNotAuthorizedOrNotFound &
the error message "The following tag namespaces / keys are not authorized or not found: 'orcl-containerengine'"
*/
var regexSystemTagNotFoundNotAuthorised = regexp.MustCompile(".*tag namespace.*orcl-containerengine.*")

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

// IsRetryable returns true if the given error is retriable.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
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

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	serviceErr, ok := common.IsServiceError(errors.Cause(err))
	return ok && serviceErr.GetHTTPStatusCode() == http.StatusTooManyRequests
}

func runWithRateLimitRetry(ctx context.Context, logger *zap.SugaredLogger, operation string, fn func(context.Context) error) error {
	var lastErr error

	for attempt := uint(1); rateLimitRetryMaxAttempts == 0 || attempt <= rateLimitRetryMaxAttempts; attempt++ {
		opErr := fn(ctx)
		lastErr = opErr
		if !isRateLimitError(opErr) {
			return opErr
		}

		backoff := rateLimitRetryNextDuration(attempt)
		if logger != nil {
			logger.Warnf("%s hit rate limit on attempt %d, retrying in %s", operation, attempt, backoff)
		}
		if deadline, ok := ctx.Deadline(); ok && time.Now().Add(backoff).After(deadline) {
			return fmt.Errorf("%s retry exceeded context deadline: %w", operation, context.DeadlineExceeded)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("operation %s reached retry limit", operation)
	}
	return fmt.Errorf("%s retry exceeded maximum attempts: %w", operation, lastErr)
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

func IsSystemTagNotFoundOrNotAuthorisedError(logger *zap.SugaredLogger, err error) bool {

	var ociServiceError common.ServiceError

	// unwrap till ociServiceError is found
	if errors.As(err, &ociServiceError) {
		logger.Debugf("API error code: %s", ociServiceError.GetCode())
		logger.Debugf("service error message: %s", ociServiceError.GetMessage())

		if ociServiceError.GetHTTPStatusCode() == http.StatusBadRequest &&
			ociServiceError.GetCode() == HTTP400RelatedResourceNotAuthorizedOrNotFoundCode {
			return regexSystemTagNotFoundNotAuthorised.MatchString(ociServiceError.GetMessage())
		}
	}
	return false
}
