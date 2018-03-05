// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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

func TestMapAvailabilityDomainToFailureDomain(t *testing.T) {
	var testCases = map[string]string{
		"NWuj:PHX-AD-1": "PHX-AD-1",
		"NWuj:PHX-AD-2": "PHX-AD-2",
		"NWuj:PHX-AD-3": "PHX-AD-3",
		"":              "",
		"PHX-AD-3":      "PHX-AD-3",
	}
	for ad, fd := range testCases {
		t.Run(ad, func(t *testing.T) {
			v := mapAvailabilityDomainToFailureDomain(ad)
			if v != fd {
				t.Errorf("mapAvailabilityDomainToFailureDomain(%q) => %q, want %q", ad, v, fd)
			}
		})
	}
}
