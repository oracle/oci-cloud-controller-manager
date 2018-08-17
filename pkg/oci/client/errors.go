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

	"github.com/oracle/oci-go-sdk/common"
	"github.com/pkg/errors"
)

var errNotFound = errors.New("not found")

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

	err = errors.Cause(err)
	serviceErr, ok := common.IsServiceError(err)
	return ok && serviceErr.GetHTTPStatusCode() == http.StatusTooManyRequests
}

// RateLimitError produces an Errorf for rate limiting.
func RateLimitError(isWrite bool, opName string) error {
	opType := "read"
	if isWrite {
		opType = "write"
	}
	return errors.Errorf("rate limited(%s) for operation: %s", opType, opName)
}
