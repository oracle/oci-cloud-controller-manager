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
	"fmt"

	baremetal "github.com/oracle/bmcs-go-sdk"
)

// SearchError represents an error searching the API.
type SearchError struct {
	Err      string
	NotFound bool
}

func (e *SearchError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.NotFound {
		return fmt.Sprintf("no matches: %s", e.Err)
	}
	return e.Err
}

// IsNotFound checks if the error is the not found error returned from BMC
func IsNotFound(err error) bool {
	// TODO(horwitz): This is temporary until we remove SearchError in
	// favor of just using the BMCS client errors directly.
	ociErr, ok := err.(*baremetal.Error)
	if ok {
		return ociErr.Status == "404"
	}

	se, ok := err.(*SearchError)
	return ok && se.NotFound
}
