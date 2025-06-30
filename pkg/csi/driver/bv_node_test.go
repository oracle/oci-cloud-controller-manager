// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
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

package driver

import (
	"fmt"
	"testing"
)

func Test_getDevicePathAndAttachmentType(t *testing.T) {
	type args struct {
		path []string
	}
	tests := []struct {
		name           string
		args           args
		attachmentType string
		diskByPath     string
		wantErr        bool
	}{
		{
			"Testing PV path with digits only",
			args{path: []string{"/dev/disk/by-path/pci-0000:18:00.0-scsi-0:0:0:5"}},
			"paravirtualized",
			"/dev/disk/by-path/pci-0000:18:00.0-scsi-0:0:0:5",
			false,
		},
		{
			"Testing PV path with hexadecimal controller",
			args{path: []string{"/dev/disk/by-path/pci-0000:1a:00.0-scsi-0:0:4:1"}},
			"paravirtualized",
			"/dev/disk/by-path/pci-0000:1a:00.0-scsi-0:0:4:1",
			false,
		},
		{
			"Testing PV path with hexadecimal Bus",
			args{path: []string{"/dev/disk/by-path/pci-0000:00:ff.0-scsi-0:0:0:1"}},
			"paravirtualized",
			"/dev/disk/by-path/pci-0000:00:ff.0-scsi-0:0:0:1",
			false,
		},
		{
			"Testing ISCSI path",
			args{path: []string{"/dev/disk/by-path/ip-169.254.2.19:3260-iscsi-iqn.2015-12.com.oracleiaas:d0ee92cb-5220-423e-b029-07ae2b2ff08f-lun-1"}},
			"iscsi",
			"/dev/disk/by-path/ip-169.254.2.19:3260-iscsi-iqn.2015-12.com.oracleiaas:d0ee92cb-5220-423e-b029-07ae2b2ff08f-lun-1",
			false,
		},
		{
			"Testing UHP volume multidevice path",
			args{path: []string{"/dev/mapper/mpathd"}},
			"iscsi",
			"/dev/mapper/mpathd",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getDevicePathAndAttachmentType(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDevicePathAndAttachmentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.attachmentType {
				t.Errorf("getDevicePathAndAttachmentType() got = %v, want attachmentType %v", got, tt.attachmentType)
			}
			if got1 != tt.diskByPath {
				t.Errorf("getDevicePathAndAttachmentType() got1 = %v, want diskByPath %v", got1, tt.diskByPath)
			}
		})
	}
}

func Test_alreadyDeletedPathCheck(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Error contains 'does not exist'",
			args{err: fmt.Errorf("path /some/path does not exist")},
			true,
		},
		{
			"Nil error",
			args{err: nil},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := alreadyDeletedPathCheck(tt.args.err)
			if got != tt.want {
				t.Errorf("alreadyDeletedPathCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
