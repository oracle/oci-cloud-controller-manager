// Copyright 2019 Oracle and/or its affiliates. All rights reserved.
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

package csioptions

import (
	"testing"
)

func Test_GetFssAddress(t *testing.T) {
	testCases := map[string]struct {
		csiAddress         string
		expectedFssAddress string
		defaultAddress     string
	}{
		"Valid csi address": {
			csiAddress:         "/var/run/shared-tmpfs/csi.sock",
			expectedFssAddress: "/var/run/shared-tmpfs/csi-fss.sock",
			defaultAddress:     "/var/run/shared-tmpfs/csi-fss.sock",
		},
		"Invalid csi address": {
			csiAddress:         "/var/run/shared-tmpfs/csi.sock.sock",
			expectedFssAddress: "/var/run/shared-tmpfs/csi-fss.sock",
			defaultAddress:     "/var/run/shared-tmpfs/csi-fss.sock",
		},
		"Valid csi endpoint": {
			csiAddress:         "unix:///var/run/shared-tmpfs/csi.sock",
			expectedFssAddress: "unix:///var/run/shared-tmpfs/csi-fss.sock",
			defaultAddress:     "unix:///var/run/shared-tmpfs/csi-fss.sock",
		},
		"Invalid csi endpoint": {
			csiAddress:         "unix:///var/run/shared-tmpfs/csi-fss.sock.sock",
			expectedFssAddress: "unix:///var/run/shared-tmpfs/csi-fss.sock",
			defaultAddress:     "unix:///var/run/shared-tmpfs/csi-fss.sock",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			fssAddress := GetFssAddress(tc.csiAddress, tc.defaultAddress)
			if tc.expectedFssAddress != fssAddress {
				t.Errorf("Expected \n%+v\n but got \n%+v", tc.expectedFssAddress, fssAddress)
			}
		})
	}
}

func Test_GetFssVolumeNamePrefix(t *testing.T) {
	testCases := map[string]struct {
		csiPrefix      string
		expectedPrefix string
	}{
		"Valid csi address": {
			csiPrefix:      "csi",
			expectedPrefix: "csi-fss",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			fssVolumeNamePrefix := GetFssVolumeNamePrefix(tc.csiPrefix)
			if tc.expectedPrefix != fssVolumeNamePrefix {
				t.Errorf("Expected \n%+v\n but got \n%+v", tc.expectedPrefix, fssVolumeNamePrefix)
			}
		})
	}
}
