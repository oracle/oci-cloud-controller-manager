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

const (
	notFoundStatus = "404"
)

// NewNotFoundError creates a new baremetal error with the correct
// status and code. The message of the error is set to the message passed in.
func NewNotFoundError(msg string) error {
	return &baremetal.Error{
		Status:  notFoundStatus,
		Code:    baremetal.NotAuthorizedOrNotFound,
		Message: msg,
	}
}

// IsConflict checks if the error is due to a conflict caused by etag mismatch.
func IsConflict(err error) bool {
	// TODO(horwitz): This is supposed to be fixed soon. It's a bug in the OCI API that causes a 409 to
	// be returned instead of a 412.
	return IsError(err, "409", "Conflict") || IsError(err, "412", "NoEtagMatch")
}

// IsNotFound checks if the error is the not found error returned from OCI.
func IsNotFound(err error) bool {
	return IsStatus(err, notFoundStatus)
}

// IsError checks that the error is an OCI error and that the status & code match.
// https://docs.us-phoenix-1.oraclecloud.com/Content/API/References/apierrors.htm
func IsError(err error, status string, code string) bool {
	return IsStatus(err, status) && IsCode(err, code)
}

// IsStatus is a helper function that ensures the error is an OCI
// client error and that the status is what is expected.
// https://docs.us-phoenix-1.oraclecloud.com/Content/API/References/apierrors.htm
func IsStatus(err error, status string) bool {
	ociErr, ok := err.(*baremetal.Error)
	return ok && ociErr.Status == status
}

// IsCode ensures that the error is an OCI error and that the code matches.
// https://docs.us-phoenix-1.oraclecloud.com/Content/API/References/apierrors.htm
func IsCode(err error, code string) bool {
	ociErr, ok := err.(*baremetal.Error)
	return ok && ociErr.Code == code
}
