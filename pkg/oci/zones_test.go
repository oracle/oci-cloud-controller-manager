// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.
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
package oci

import (
	"testing"
)

var mapAvailabilityDomainToFailureDomainTestCases = []struct {
	ad string
	fd string
}{
	{ad: "NWuj:PHX-AD-1", fd: "PHX-AD-1"},
	{ad: "NWuj:PHX-AD-2", fd: "PHX-AD-2"},
	{ad: "NWuj:PHX-AD-3", fd: "PHX-AD-3"},
	{ad: "", fd: ""},
	{ad: "PHX-AD-3", fd: "PHX-AD-3"},
}

func TestMapAvailabilityDomainToFailureDomain(t *testing.T) {
	for _, tt := range mapAvailabilityDomainToFailureDomainTestCases {
		v := mapAvailabilityDomainToFailureDomain(tt.ad)
		if v != tt.fd {
			t.Errorf("mapAvailabilityDomainToFailureDomain(%q) => %q, want %q", tt.ad, v, tt.fd)
		}
	}
}
