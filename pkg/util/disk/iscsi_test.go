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

package disk

import (
	"reflect"
	"testing"

	"github.com/oracle/oci-cloud-controller-manager/pkg/util/mount"
)

type mockMountLister struct {
	mps []mount.MountPoint
}

func (ml *mockMountLister) List() ([]mount.MountPoint, error) {
	return ml.mps, nil
}

func TestGetMountPointForPath(t *testing.T) {
	testCases := []struct {
		name     string
		mps      []mount.MountPoint
		path     string
		err      error
		expected mount.MountPoint
	}{
		{
			name: "single",
			mps: []mount.MountPoint{
				{Path: "/tmp/my-mountpoint"},
			},
			path:     "/tmp/my-mountpoint",
			err:      nil,
			expected: mount.MountPoint{Path: "/tmp/my-mountpoint"},
		}, {
			name: "multiple",
			mps: []mount.MountPoint{
				{Path: "/tmp/my-other-mountpoint"},
				{Path: "/tmp/my-mountpoint"},
			},
			path:     "/tmp/my-mountpoint",
			err:      nil,
			expected: mount.MountPoint{Path: "/tmp/my-mountpoint"},
		}, {
			name: "missing",
			mps: []mount.MountPoint{
				{Path: "/tmp/my-other-mountpoint"},
			},
			path:     "/tmp/my-mountpoint",
			err:      ErrMountPointNotFound,
			expected: mount.MountPoint{},
		},
	}

	mock := &mockMountLister{}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mock.mps = tt.mps
			result, err := getMountPointForPath(mock, tt.path)
			if err != tt.err {
				t.Fatalf("getMountPointForPath(mockLister, %q) => error: %v; expected %v", tt.path, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("getMountPointForPath(mockLister, %q) =>\n%+v\nExpected: %+v", tt.path, result, tt.expected)
			}
		})
	}
}
