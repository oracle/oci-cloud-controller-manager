// Copyright 2017 The OCI Cloud Controller Manager Authors
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
	baremetal "github.com/oracle/bmcs-go-sdk"
)

type statusCode string

const (
	// https://docs.us-phoenix-1.oraclecloud.com/Content/API/References/apierrors.htm
	notFoundStatus statusCode = "404"
)

// NewNotFoundError creates a new baremetal error with the correct
// status and code. The message of the error is set to the message passed in.
func NewNotFoundError(msg string) error {
	return &baremetal.Error{
		Status:  string(notFoundStatus),
		Code:    baremetal.NotAuthorizedOrNotFound,
		Message: msg,
	}
}

// IsNotFound checks if the error is the not found error returned from OCI.
func IsNotFound(err error) bool {
	return IsStatus(err, notFoundStatus)
}

// IsStatus is a helper function that ensures the error is an OCI
// client error and that the status is what is expected.
// https://docs.us-phoenix-1.oraclecloud.com/Content/API/References/apierrors.htm
func IsStatus(err error, status statusCode) bool {
	ociErr, ok := err.(*baremetal.Error)
	return ok && ociErr.Status == string(status)
}
