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

package block

import (
	"os"
	"testing"

	"github.com/oracle/oci-cloud-controller-manager/pkg/flexvolume"
)

var volumeOCIDTests = []struct {
	regionKey  string
	volumeName string
	expected   string
}{
	{"phx", "aaaaaa", "ocid1.volume.oc1.phx.aaaaaa"},
	{"iad", "aaaaaa", "ocid1.volume.oc1.iad.aaaaaa"},
	{"eu-frankfurt-1", "aaaaaa", "ocid1.volume.oc1.eu-frankfurt-1.aaaaaa"},
	{"uk-london-1", "aaaaaa", "ocid1.volume.oc1.uk-london-1.aaaaaa"},
}

func TestDeriveVolumeOCID(t *testing.T) {
	for _, tt := range volumeOCIDTests {
		result := deriveVolumeOCID(tt.regionKey, tt.volumeName)
		if result != tt.expected {
			t.Errorf("Failed to derive OCID. Expected %s got %s", tt.expected, result)
		}
	}
}

func TestGetConfigPath(t *testing.T) {
	testCases := map[string]struct {
		envvar   string
		value    string
		expected string
	}{
		"default": {
			envvar:   "",
			value:    "",
			expected: "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci/config.yaml",
		},
		"custom config dir": {
			envvar:   "OCI_FLEXD_CONFIG_DIRECTORY",
			value:    "/foo/bar/",
			expected: "/foo/bar/config.yaml",
		},
		"custom driver dir": {
			envvar:   "OCI_FLEXD_DRIVER_DIRECTORY",
			value:    "/foo/baz",
			expected: "/foo/baz/config.yaml",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// idk if we need this but figure it can't hurt
			original := os.Getenv(tc.envvar)
			defer os.Setenv(tc.envvar, original)

			// set env var value for the test.
			os.Setenv(tc.envvar, tc.value)

			result := GetConfigPath()
			if result != tc.expected {
				t.Errorf("GetDriverDirectory() = %q ; wanted %q", result, tc.expected)
			}
		})

	}
}

func TestGetVolumeName(t *testing.T) {
	testCases := map[string]struct {
		opts     flexvolume.Options
		expected flexvolume.DriverStatus
	}{
		"real": {
			opts: flexvolume.Options{
				"kubernetes.io/fsType":         "ext4",
				"kubernetes.io/pvOrVolumeName": "ocid1.volume.oc1.iad.abuwcljsd4fjqgn43gsnkj536z5sbb2unwsp35545y4jqm4pbrhf7azqpdtq",
				"kubernetes.io/readwrite":      "rw"},
			expected: flexvolume.DriverStatus{
				Status:     flexvolume.StatusSuccess,
				VolumeName: "abuwcljsd4fjqgn43gsnkj536z5sbb2unwsp35545y4jqm4pbrhf7azqpdtq",
			},
		},
		"empty": {
			opts: flexvolume.Options{},
			expected: flexvolume.DriverStatus{
				Status: flexvolume.StatusFailure,
			},
		},
		"invalid": {
			opts: flexvolume.Options{
				"kubernetes.io/fsType":         "ext4",
				"kubernetes.io/pvOrVolumeName": "coid1.volume.oc1.iad.abuwcljsd4fjqgn43gsnkj536z5sbb2unwsp35545y4jqm4pbrhf7azqpdtq",
			},
			expected: flexvolume.DriverStatus{
				Status: flexvolume.StatusFailure,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := GetVolumeName(tc.opts)
			if result.Status != tc.expected.Status || result.VolumeName != tc.expected.VolumeName {
				t.Errorf("GetVolumeName()\nactual: %#v\nwanted: %#v", result, tc.expected)
			}
		})
	}

}

func TestGetKubeconfigPath(t *testing.T) {
	testCases := map[string]struct {
		envvar   string
		value    string
		expected string
	}{
		"default": {
			envvar:   "",
			value:    "",
			expected: "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci/kubeconfig",
		},
		"custom config dir": {
			envvar:   "OCI_FLEXD_CONFIG_DIRECTORY",
			value:    "/foo/bar",
			expected: "/foo/bar/kubeconfig",
		},
		"custom config dir with trailing path seperator": {
			envvar:   "OCI_FLEXD_CONFIG_DIRECTORY",
			value:    "/foo/bar/",
			expected: "/foo/bar/kubeconfig",
		},
		"override kubeconfig path": {
			envvar:   "OCI_FLEXD_KUBECONFIG_PATH",
			value:    "/etc/kubevar/kubeconfig",
			expected: "/etc/kubevar/kubeconfig",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// idk if we need this but figure it can't hurt
			original := os.Getenv(tc.envvar)
			defer os.Setenv(tc.envvar, original)

			// set env var value for the test.
			os.Setenv(tc.envvar, tc.value)

			result := GetKubeconfigPath()
			if result != tc.expected {
				t.Errorf("GetKubeconfigPath() = %q ; wanted %q", result, tc.expected)
			}
		})
	}
}
