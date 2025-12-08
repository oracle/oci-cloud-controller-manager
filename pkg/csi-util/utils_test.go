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

package csi_util

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
)

func TestUtil_getAvailableDomainInNodeLabel(t *testing.T) {
	type fields struct {
		logger *zap.SugaredLogger
	}
	type args struct {
		fullAD string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Get AD name from the tenancy specific AD name.",
			fields: fields{
				logger: zap.S(),
			},
			args: args{fullAD: "zkJl:US-ASHBURN-AD-1"},
			want: "US-ASHBURN-AD-1",
		},
		{
			name: "Get AD name from the tenancy specific AD name for empty string",
			fields: fields{
				logger: zap.S(),
			},
			args: args{fullAD: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &Util{
				Logger: tt.fields.logger,
			}
			if got := u.GetAvailableDomainInNodeLabel(tt.args.fullAD); got != tt.want {
				t.Errorf("Util.getAvailableDomainInNodeLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateFsType(t *testing.T) {
	type args struct {
		logger *zap.SugaredLogger
		fsType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Return ext4",
			args: args{
				logger: zap.S(),
				fsType: "ext4",
			},
			want: "ext4",
		},
		{
			name: "Return ext3",
			args: args{
				logger: zap.S(),
				fsType: "ext3",
			},
			want: "ext3",
		},
		{
			name: "Return xfs",
			args: args{
				logger: zap.S(),
				fsType: "xfs",
			},
			want: "xfs",
		},
		{
			name: "Return default ext4 for empty string",
			args: args{
				logger: zap.S(),
				fsType: "",
			},
			want: "ext4",
		},
		{
			name: "Return default ext4 for unsupported string",
			args: args{
				logger: zap.S(),
				fsType: "xxxxx",
			},
			want: "ext4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateFsType(tt.args.logger, tt.args.fsType); got != tt.want {
				t.Errorf("validateFsType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ValidateFssId(t *testing.T) {
	tests := []struct {
		name                 string
		volumeHandle         string
		wantFssVolumeHandler *FSSVolumeHandler
	}{
		{
			name:         "Filesystem exposed Ipv4 Mount Target",
			volumeHandle: "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa:10.0.2.44:/FileSystem-Test",
			wantFssVolumeHandler: &FSSVolumeHandler{
				FilesystemOcid:       "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa",
				MountTargetIPAddress: "10.0.2.44",
				FsExportPath:         "/FileSystem-Test",
			},
		},
		{
			name:         "Filesystem exposed by Mount Target having dns",
			volumeHandle: "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa:myhostname.subnet123.dnslabel.oraclevcn.com:/FileSystem-Test",
			wantFssVolumeHandler: &FSSVolumeHandler{
				FilesystemOcid:       "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa",
				MountTargetIPAddress: "myhostname.subnet123.dnslabel.oraclevcn.com",
				FsExportPath:         "/FileSystem-Test",
			},
		},
		{
			name:                 "Invalid Ipv4 provided in volume handle",
			volumeHandle:         "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa:10.0.2:/FileSystem-Test",
			wantFssVolumeHandler: &FSSVolumeHandler{},
		},
		{
			name:                 "Filesystem ocid not provided",
			volumeHandle:         "10.0.2:/FileSystem-Test",
			wantFssVolumeHandler: &FSSVolumeHandler{},
		},
		{
			name:                 "Mount target Ip Address not provided",
			volumeHandle:         "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa:10.0.2",
			wantFssVolumeHandler: &FSSVolumeHandler{},
		},
		{
			name:                 "Empty volume handle provided",
			volumeHandle:         "",
			wantFssVolumeHandler: &FSSVolumeHandler{},
		},
		{
			name:         "Filesystem exposed over Ipv6 Mount Target",
			volumeHandle: "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa:[fd00:c1::a9fe:504]:/FileSystem-Test",
			wantFssVolumeHandler: &FSSVolumeHandler{
				FilesystemOcid:       "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa",
				MountTargetIPAddress: "[fd00:c1::a9fe:504]",
				FsExportPath:         "/FileSystem-Test",
			},
		},
		{
			name:         "Filesystem exposed over Ipv6 Mount Target",
			volumeHandle: "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa:fd00:c1::a9fe:504:/FileSystem-Test",
			wantFssVolumeHandler: &FSSVolumeHandler{
				FilesystemOcid:       "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa",
				MountTargetIPAddress: "fd00:c1::a9fe:504",
				FsExportPath:         "/FileSystem-Test",
			},
		},
		{
			name:                 "Invalid volumeHandle ::",
			volumeHandle:         "::",
			wantFssVolumeHandler: &FSSVolumeHandler{},
		},
		{
			name:                 "Invalid input",
			volumeHandle:         ":",
			wantFssVolumeHandler: &FSSVolumeHandler{},
		},
		{
			name:                 "Just export provided",
			volumeHandle:         ":/FileSystem-Test",
			wantFssVolumeHandler: &FSSVolumeHandler{},
		},
		{
			name:                 "Export not provided",
			volumeHandle:         "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa:fd00:c1::a9fe:504:",
			wantFssVolumeHandler: &FSSVolumeHandler{},
		},
		{
			name:                 "Invalid dns name provided in volume handle",
			volumeHandle:         "ocid1.filesystem.oc1.phx.aaaaaaaaaahjpdudobuhqllqojxwiotqnb4c2ylefuzaaaaa:Invalid Dns:/FileSystem-Test",
			wantFssVolumeHandler: &FSSVolumeHandler{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFssVolumeHandler := ValidateFssId(tt.volumeHandle)
			if gotFssVolumeHandler.MountTargetIPAddress != tt.wantFssVolumeHandler.MountTargetIPAddress ||
				gotFssVolumeHandler.FsExportPath != tt.wantFssVolumeHandler.FsExportPath ||
				gotFssVolumeHandler.FilesystemOcid != tt.wantFssVolumeHandler.FilesystemOcid {
				t.Errorf("ValidateFssId() = %v, want %v", gotFssVolumeHandler, tt.wantFssVolumeHandler)
			}
		})
	}
}

func Test_ConvertIscsiIpFromIpv4ToIpv6(t *testing.T) {

	tests := []struct {
		name        string
		ipv4IscsiIp string
		want        string
		err         error
	}{
		{
			name:        "Return Icsci Ipv6 from valid Iscsi Ipv4",
			ipv4IscsiIp: "169.254.2.2",
			want:        "fd00:c1::a9fe:202",
		},
		{
			name:        "Return Icsci Ipv6 from valid Iscsi Ipv4",
			ipv4IscsiIp: "169.254.5.4",
			want:        "fd00:c1::a9fe:504",
		},
		{
			name:        "Invalid Iscsi Ipv4 should error",
			ipv4IscsiIp: "169.254.2",
			want:        "",
			err:         fmt.Errorf("invalid iSCSIIp identified %s", "169.254.2"),
		},
		{
			name:        "Invalid Iscsi Ipv4 should error",
			ipv4IscsiIp: "",
			want:        "",
			err:         fmt.Errorf("invalid iSCSIIp identified %s", ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertIscsiIpFromIpv4ToIpv6(tt.ipv4IscsiIp)
			if got != tt.want {
				t.Errorf("ConvertIscsiIpFromIpv4ToIpv6() = %v, want %v", got, tt.want)
			}

			if err != nil && !strings.EqualFold(err.Error(), tt.err.Error()) {
				t.Errorf("ConvertIscsiIpFromIpv4ToIpv6() = %v, want %v", err.Error(), tt.err.Error())
			}
		})
	}
}

func Test_DiskByPathPatternForIscsi(t *testing.T) {

	tests := []struct {
		name            string
		diskByPathValue string
		want            bool
		err             error
	}{
		{
			name:            "DiskByPathPatternForIscsi is Able to match ipv6 disk path",
			diskByPathValue: "/dev/disk/by-path/ip-fd00:00c1::a9fe:202:3260-iscsi-iqn.2015-12.com.oracleiaas:63a2e76c-5353-4a75-82d0-ee31a39471ca-lun-2",
			want:            true,
		},
		{
			name:            "DiskByPathPatternForIscsi is Able to match ipv4 disk path",
			diskByPathValue: "/dev/disk/by-path/ip-169.254.2.2:3260-iscsi-iqn.2015-12.com.oracleiaas:63a2e76c-5353-4a75-82d0-ee31a39471ca-lun-2",
			want:            true,
		},
		{
			name:            "DiskByPathPatternForIscsi does to match invalid ipv6 disk path",
			diskByPathValue: "/dev/disk/by-path/ip-@3#:3260-iscsi-iqn.2015-12.com.oracleiaas:63a2e76c-5353-4a75-82d0-ee31a39471ca-lun-2",
			want:            false,
		},
		{
			name:            "DiskByPathPatternForIscsi does to match invalid ipv4 disk path",
			diskByPathValue: "/dev/disk/by-path/ip-16@9.25#4.2.2:3260-iscsi-iqn.2015-12.com.oracleiaas:63a2e76c-5353-4a75-82d0-ee31a39471ca-lun-2",
			want:            false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := regexp.MatchString(DiskByPathPatternISCSI, tt.diskByPathValue)
			if got != tt.want {
				t.Errorf("DiskByPathPatternForIscsi() = %v, want %v", got, tt.want)
			}

		})
	}
}

func Test_DiskByPathPatternForPV(t *testing.T) {

	tests := []struct {
		name            string
		diskByPathValue string
		want            bool
		err             error
	}{
		{
			name:            "DiskByPathPatternForPV is Able to valid pv disk path",
			diskByPathValue: "/dev/disk/by-path/pci-0000:00:04.0-scsi-0:0:0:4",
			want:            true,
		},
		{
			name:            "DiskByPathPatternForPV does to match invalid pv disk path",
			diskByPathValue: "/dev/disk/by-path/pci-00@#0:00:04.0-scsi-0:0:4",
			want:            false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := regexp.MatchString(DiskByPathPatternPV, tt.diskByPathValue)
			if got != tt.want {
				t.Errorf("DiskByPathPatternPV() = %v, want %v", got, tt.want)
			}

		})
	}
}

func Test_LoadNodeMetadataFromApiServer(t *testing.T) {

	tests := []struct {
		name       string
		nodeName   string
		want       *NodeMetadata
		kubeclient kubernetes.Interface
		err        error
	}{
		{
			name:     "should return ipv6 for ipv6 preferred node",
			nodeName: "ipv6Preferred",
			want: &NodeMetadata{
				FullAvailabilityDomain: "xyz:PHX-AD-3",
				AvailabilityDomain:     "PHX-AD-3",
				PreferredNodeIpFamily:  Ipv6Stack,
				Ipv4Enabled:            true,
				Ipv6Enabled:            true,
			},
		},
		{
			name:     "should return ipv4 for ipv4 preferred node",
			nodeName: "ipv4Preferred",
			want: &NodeMetadata{
				PreferredNodeIpFamily: Ipv4Stack,
				AvailabilityDomain:    "PHX-AD-3",
				Ipv4Enabled:           true,
				Ipv6Enabled:           true,
			},
		},
		{
			name:     "should return default IPv4 family for no ip preference",
			nodeName: "noIpPreference",
			want: &NodeMetadata{
				AvailabilityDomain:    "PHX-AD-3",
				PreferredNodeIpFamily: Ipv4Stack,
				Ipv4Enabled:           true,
				Ipv6Enabled:           false,
			},
		},
		{
			name:     "should return error for invalid node",
			nodeName: "InvalidNode",
			want:     &NodeMetadata{},
			err:      fmt.Errorf("Failed to get node information from kube api server, please check if kube api server is accessible."),
		},
		{
			name:     "should return error for node with any ad labels",
			nodeName: "nodeWithMissingAdLabels",
			want: &NodeMetadata{
				PreferredNodeIpFamily: Ipv4Stack,
				Ipv4Enabled:           true,
				Ipv6Enabled:           false,
			},
			err: fmt.Errorf("Failed to get node information from kube api server, please check if kube api server is accessible."),
		},
		{
			name:     "Call to get node info is done  even if health check fails",
			nodeName: "ipv4Preferred",
			want: &NodeMetadata{
				PreferredNodeIpFamily: Ipv4Stack,
				AvailabilityDomain:    "PHX-AD-3",
				Ipv4Enabled:           true,
				Ipv6Enabled:           true,
			},
			kubeclient: &util.MockKubeClientWithFailingRestClient{
				CoreClient: &util.MockCoreClientWithFailingRestClient{},
			},
		},
		{
			name:     "should return error for invalid node and failing health check",
			nodeName: "InvalidNode",
			want:     &NodeMetadata{},
			err:      fmt.Errorf("Failed to get node information from kube api server, please check if kube api server is accessible."),
			kubeclient: &util.MockKubeClientWithFailingRestClient{
				CoreClient: &util.MockCoreClientWithFailingRestClient{},
			},
		},
	}

	logger, _ := zap.NewDevelopment()
	sugar := logger.Sugar()
	u := &Util{
		Logger: sugar,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			log.SetOutput(os.Stdout)
			nodeMetadata := &NodeMetadata{}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var k kubernetes.Interface
			if tt.kubeclient != nil {
				k = tt.kubeclient
			} else {
				k = &util.MockKubeClient{CoreClient: &util.MockCoreClient{}}
			}

			err := u.LoadNodeMetadataFromApiServer(ctx, k, tt.nodeName, nodeMetadata)
			if (tt.want != nodeMetadata) && (tt.want.PreferredNodeIpFamily != nodeMetadata.PreferredNodeIpFamily ||
				tt.want.Ipv6Enabled != nodeMetadata.Ipv6Enabled || tt.want.Ipv4Enabled != nodeMetadata.Ipv4Enabled) {
				t.Errorf("LoadNodeMetadataFromApiServer() = %v, want %v", nodeMetadata, tt.want)
			}
			if err != nil && !strings.EqualFold(tt.err.Error(), err.Error()) {
				t.Errorf("LoadNodeMetadataFromApiServer() = %v, want %v", err, tt.err)
			}

		})
	}

}

func Test_ExtractISCSIInformationFromMountPath(t *testing.T) {

	tests := []struct {
		name     string
		diskPath []string
		target   string
		iqn      string
		err      error
	}{
		{
			name:     "Valid Ipv6 Disk By Path returns valid Iscsi Target",
			diskPath: []string{"/dev/disk/by-path/ip-fd00:00c1::a9fe:202:3260-iscsi-iqn.2015-12.com.oracleiaas:63a2e76c-5353-4a75-82d0-ee31a39471ca-lun-2"},
			target:   "[fd00:00c1::a9fe:202]:3260",
			iqn:      "iqn.2015-12.com.oracleiaas:63a2e76c-5353-4a75-82d0-ee31a39471ca",
		},
		{
			name:     "Invalid Ipv6 Disk By Path returns error",
			diskPath: []string{"/dev/disk/by-path/ip-fd$%00:00c1::a9fe:202:3260-iscsi-iqn.2015-12.com.oracleiaas:63a2e76c-5353-4a75-82d0-ee31a39471ca-lun-2"},
			err:      fmt.Errorf("iSCSI information not found for mount point"),
		},
		{
			name:     "Valid Ipv4 Disk By Path returns valid Iscsi Target",
			diskPath: []string{"/dev/disk/by-path/ip-169.254.2.2:3260-iscsi-iqn.2015-12.com.oracleiaas:63a2e76c-5353-4a75-82d0-ee31a39471ca-lun-2"},
			target:   "169.254.2.2:3260",
			iqn:      "iqn.2015-12.com.oracleiaas:63a2e76c-5353-4a75-82d0-ee31a39471ca",
		},
		{
			name:     "Invalid Ipv4 Disk By Path returns error",
			diskPath: []string{"/dev/disk/by-path/ip-16#$9.254.2.2:326@#0-iscsi-iqn.2015-12.com.oracleiaas:63a2e76c-5353-4a75-82d0-ee31a39471ca-lun-2"},
			err:      fmt.Errorf("iSCSI information not found for mount point"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disk, err := ExtractISCSIInformationFromMountPath(zap.S(), tt.diskPath)

			if err != nil && !strings.EqualFold(tt.err.Error(), err.Error()) {
				t.Errorf("ExtractISCSIInformationFromMountPath() = %v, want %v", err, tt.err)
			}
			if err == nil {
				if disk.Target() != tt.target {
					t.Errorf("ExtractISCSIInformationFromMountPath() Target = %v, want %v", disk.Target(), tt.target)
				}
				if disk.IQN != tt.iqn {
					t.Errorf("ExtractISCSIInformationFromMountPath() IQN = %v, want %v", disk.Target(), tt.target)
				}
			}

		})
	}
}

func Test_Ip_Util_Methods(t *testing.T) {

	tests := []struct {
		name               string
		ipAddress          string
		formattedIpAddress string
		isIpv4             bool
		isIpv6             bool
	}{
		{
			name:               "Valid ipv4",
			ipAddress:          "10.0.0.10",
			formattedIpAddress: "10.0.0.10",
			isIpv4:             true,
			isIpv6:             false,
		},
		{
			name:               "Valid ipv6",
			ipAddress:          "fd00:00c1::a9fe:202",
			formattedIpAddress: "[fd00:00c1::a9fe:202]",
			isIpv4:             false,
			isIpv6:             true,
		},
		{
			name:               "Valid formatted ipv6",
			ipAddress:          "[fd00:00c1::a9fe:202]",
			formattedIpAddress: "[fd00:00c1::a9fe:202]",
			isIpv4:             false,
			isIpv6:             true,
		},
		{
			name:               "Invalid ipv4",
			ipAddress:          "10.0.0.",
			formattedIpAddress: "10.0.0.",
			isIpv4:             false,
			isIpv6:             false,
		},
		{
			name:               "Invalid ipv6",
			ipAddress:          "zxf0:00c1::a9fe",
			formattedIpAddress: "zxf0:00c1::a9fe",
			isIpv4:             false,
			isIpv6:             false,
		},
		{
			name:               "dns name",
			ipAddress:          "mtwithdns.subc7a90bc13.cluster1.oraclevcn.com",
			formattedIpAddress: "mtwithdns.subc7a90bc13.cluster1.oraclevcn.com",
			isIpv4:             false,
			isIpv6:             false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := IsIpv4(tt.ipAddress)
			if tt.isIpv4 != got {
				t.Errorf("IsIpv4() = %v, want %v", got, tt.isIpv4)
			}

			got = IsIpv6(tt.ipAddress)
			if tt.isIpv6 != got {
				t.Errorf("IsIpv6() = %v, want %v", got, tt.isIpv6)
			}

			gotStr := FormatValidIp(tt.ipAddress)
			if tt.formattedIpAddress != gotStr {
				t.Errorf("FormatValidIp() = %v, want %v", gotStr, tt.isIpv6)
			}
		})
	}

}

func Test_FormatValidIpStackInK8SConvention(t *testing.T) {

	tests := []struct {
		name    string
		ipStack string
		want    string
	}{
		{
			name:    "ipv4 stack name in non k8s convention i.e IPv4 ",
			ipStack: "ipv4",
			want:    "IPv4",
		},
		{
			name:    "ipv6 stack name in non k8s convention i.e IPv6 ",
			ipStack: "ipv6",
			want:    "IPv6",
		},
		{
			name:    "invalid ip stack ",
			ipStack: "invalid",
			want:    "invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValidIpStackInK8SConvention(tt.ipStack)
			if got != tt.want {
				t.Errorf("FormatValidIpStackInK8SConvention() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_IsValidIpFamilyPresentInClusterIpFamily(t *testing.T) {

	tests := []struct {
		name            string
		clusterIpFamily string
		isValid         bool
	}{
		{
			name:            "Single stack ipv4 clusters",
			clusterIpFamily: "IPv4",
			isValid:         true,
		},
		{
			name:            "Single stack ipv6 clusters",
			clusterIpFamily: "IPv6",
			isValid:         true,
		},
		{
			name:            "Dual stack IPv6 preferred cluster",
			clusterIpFamily: "IPv6,IPv4",
			isValid:         true,
		},
		{
			name:            "Dual stack IPv4 preferred cluster",
			clusterIpFamily: "IPv4,IPv6",
			isValid:         true,
		},
		{
			name:            "Invalid Ip Family",
			clusterIpFamily: "Invalid",
			isValid:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidIpFamilyPresentInClusterIpFamily(tt.clusterIpFamily)
			if got != tt.isValid {
				t.Errorf("IsValidIpFamilyPresentInClusterIpFamily() = %v, want %v", got, tt.isValid)
			}
		})
	}
}

func Test_SubnetStack(t *testing.T) {
	tests := []struct {
		name                    string
		subnet                  *core.Subnet
		isIpv4SingleStackSubnet bool
		isIpv6SingleStackSubnet bool
		IsDualStackSubnet       bool
	}{
		{
			name: "IPv4 Single stack subnet",
			subnet: &core.Subnet{
				CidrBlock: pointer.String("10.0.0.1/24"),
			},
			isIpv4SingleStackSubnet: true,
			isIpv6SingleStackSubnet: false,
			IsDualStackSubnet:       false,
		},
		{
			name: "IPv6 Single stack subnet",
			subnet: &core.Subnet{
				CidrBlock:      pointer.String("<null>"),
				Ipv6CidrBlock:  pointer.String("2603:c020:e:897e::/64"),
				Ipv6CidrBlocks: []string{"2603:c020:e:897e::/64"},
			},
			isIpv4SingleStackSubnet: false,
			isIpv6SingleStackSubnet: true,
			IsDualStackSubnet:       false,
		},
		{
			name: "IPv6 Single stack subnet with CidrBlock nil",
			subnet: &core.Subnet{
				CidrBlock:      nil,
				Ipv6CidrBlock:  pointer.String("2603:c020:e:897e::/64"),
				Ipv6CidrBlocks: []string{"2603:c020:e:897e::/64"},
			},
			isIpv4SingleStackSubnet: false,
			isIpv6SingleStackSubnet: true,
			IsDualStackSubnet:       false,
		},
		{
			name: "IPv6 Single stack subnet with ULA local address cidr",
			subnet: &core.Subnet{
				CidrBlock:      pointer.String("<null>"),
				Ipv6CidrBlocks: []string{"fc00:0000:0000:0000:0000:0000:0000:0000/64"},
			},
			isIpv4SingleStackSubnet: false,
			isIpv6SingleStackSubnet: true,
			IsDualStackSubnet:       false,
		},
		{
			name: "Dual stack subnet",
			subnet: &core.Subnet{
				CidrBlock:      pointer.String("10.0.2.0/24"),
				Ipv6CidrBlock:  pointer.String("2603:c020:e:897e::/64"),
				Ipv6CidrBlocks: []string{"2603:c020:e:897e::/64"},
			},
			isIpv4SingleStackSubnet: false,
			isIpv6SingleStackSubnet: false,
			IsDualStackSubnet:       true,
		},
		{
			name: "Dual  stack subnet with ULA local address cidr",
			subnet: &core.Subnet{
				CidrBlock:      pointer.String("10.0.2.0/24"),
				Ipv6CidrBlocks: []string{"fc00:0000:0000:0000:0000:0000:0000:0000/64"},
			},
			isIpv4SingleStackSubnet: false,
			isIpv6SingleStackSubnet: false,
			IsDualStackSubnet:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := IsIpv4SingleStackSubnet(tt.subnet)
			if tt.isIpv4SingleStackSubnet != got {
				t.Errorf("IsIpv4SingleStackSubnet() = %v, want %v", got, tt.isIpv4SingleStackSubnet)
			}

			got = IsIpv6SingleStackSubnet(tt.subnet)
			if tt.isIpv6SingleStackSubnet != got {
				t.Errorf("IsIpv6SingleStackSubnet() = %v, want %v", got, tt.isIpv6SingleStackSubnet)
			}

			gotStr := IsDualStackSubnet(tt.subnet)
			if tt.IsDualStackSubnet != gotStr {
				t.Errorf("IsDualStackSubnet() = %v, want %v", gotStr, tt.IsDualStackSubnet)
			}
		})
	}
}

func Test_ValidateDNSName(t *testing.T) {
	tests := []struct {
		name           string
		dnsName        string
		expectedResult bool
	}{
		{
			name:           "Valid DNS Name",
			dnsName:        "myhostname.subnet123.dnslabel.oraclevcn.com",
			expectedResult: true,
		},
		{
			name:           "Valid DNS Name",
			dnsName:        "mymounttarget.dev",
			expectedResult: true,
		},
		{
			name:           "Valid DNS Name",
			dnsName:        "all.chars-123ns.org",
			expectedResult: true,
		},
		{
			name:           "Invalid dns",
			dnsName:        "-myhostname.com",
			expectedResult: false,
		},
		{
			name:           "Invalid dns",
			dnsName:        "InvalidDns",
			expectedResult: false,
		},
		{
			name:           "Invalid dns",
			dnsName:        "10.10.0.0",
			expectedResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationResult := ValidateDNSName(tt.dnsName)
			if validationResult != tt.expectedResult {
				t.Errorf("ValidateDNSName() = %v, want %v", validationResult, tt.expectedResult)
			}
		})
	}
}
func Test_LoadCSIConfigFromConfigMap(t *testing.T) {

	tests := []struct {
		name          string
		configMapName string
		want          *util.CSIConfig
	}{
		{
			name:          "Parse Configs correctly when csi config map is present",
			configMapName: "oci-csi-config",
			want: &util.CSIConfig{
				Lustre: &util.DriverConfig{
					SkipNodeUnstage:      true,
					SkipLustreParameters: true,
				},
			},
		},
		{
			name:          "Return default config if config map is not present",
			configMapName: "invalid",
			want:          &util.CSIConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csiConfig := &util.CSIConfig{}
			LoadCSIConfigFromConfigMap(csiConfig, &util.MockKubeClient{
				CoreClient: &util.MockCoreClient{},
			}, tt.configMapName, zap.S(), context.Background())

			if !reflect.DeepEqual(tt.want, csiConfig) {
				t.Errorf("LoadCSIConfigFromConfigMap() => got : %v, want :  %v", csiConfig, tt.want)
			}
		})
	}

}

// Mock for common.ServiceErrorRichInfo
type mockServiceError struct {
	StatusCode              int
	Code                    string            `json:"code,omitempty"`
	Message                 string            `json:"message,omitempty"`
	OriginalMessage         string            `json:"originalMessage"`
	OriginalMessageTemplate string            `json:"originalMessageTemplate"`
	MessageArgument         map[string]string `json:"messageArguments"`
	OpcRequestID            string            `json:"opc-request-id"`
	// debugging information
	TargetService string         `json:"target-service"`
	OperationName string         `json:"operation-name"`
	Timestamp     common.SDKTime `json:"timestamp"`
	RequestTarget string         `json:"request-target"`
	ClientVersion string         `json:"client-version"`

	// troubleshooting guidance
	OperationReferenceLink   string `json:"operation-reference-link"`
	ErrorTroubleshootingLink string `json:"error-troubleshooting-link"`
}

func (m *mockServiceError) GetTargetService() string            { return m.TargetService }
func (m *mockServiceError) GetHTTPStatusCode() int              { return m.StatusCode }
func (m *mockServiceError) GetCode() string                     { return m.Code }
func (m *mockServiceError) GetOpcRequestID() string             { return m.OpcRequestID }
func (m *mockServiceError) GetMessage() string                  { return m.Message }
func (m *mockServiceError) GetOperationName() string            { return m.OperationName }
func (m *mockServiceError) GetTimestamp() common.SDKTime        { return m.Timestamp }
func (m *mockServiceError) GetRequestTarget() string            { return m.RequestTarget }
func (m *mockServiceError) GetClientVersion() string            { return m.ClientVersion }
func (m *mockServiceError) GetOperationReferenceLink() string   { return m.OperationReferenceLink }
func (m *mockServiceError) GetErrorTroubleshootingLink() string { return m.ErrorTroubleshootingLink }
func (m *mockServiceError) Error() string                       { return m.Message }

func TestGetOCIServiceError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error returns empty string",
			err:      nil,
			expected: "",
		},

		{
			name:     "non-OCI error returns err.Error()",
			err:      errors.New("Timeout of waiting for condition. Lot of big text."),
			expected: "Timeout of waiting for condition. Lot of big text.",
		},
		{
			name: "OCI service error returns formatted string",
			err: &mockServiceError{
				TargetService:            "Compute Service",
				StatusCode:               400,
				Code:                     "LimitExceeded",
				OpcRequestID:             "8239a6cbec18bed008924abe8b12416b/35C4EBDF5A7728341854B0685048586B/D66BF5A8491C409D1C61B18D343607A4",
				Message:                  "Cannot attach volume ocid1.volume.oc1.phx.abyhqljrwv2qo35rdomkhvmgtrwr6wpkezzh6gb355tn7phtp5szt5hleqja to instance ocid1.instance.oc1.phx.anyhqljrh4gjgpyc2ighn7uq32unt2e4grzkw3ta37kxbi6c3z3xhlnx66aq because the instance already has the maximum number 16 of attached PV volumes for the shape VM.Standard.E5.Flex. To attach this volume, first detach an existing PV volume from the instance, and then try again..",
				OperationName:            "AttachVolume",
				ClientVersion:            "Oracle-GoSDK/xVersion",
				RequestTarget:            "https://iaas.us-phoenix-1.oraclecloud.com/20160918/volumeAttachments",
				OperationReferenceLink:   "Test",
				ErrorTroubleshootingLink: "Test",
			},
			expected: "Error returned by Compute Service Service. Http Status Code: 400. Error Code: LimitExceeded. Message: Cannot attach volume ocid1.volume.oc1.phx.abyhqljrwv2qo35rdomkhvmgtrwr6wpkezzh6gb355tn7phtp5szt5hleqja to instance ocid1.instance.oc1.phx.anyhqljrh4gjgpyc2ighn7uq32unt2e4grzkw3ta37kxbi6c3z3xhlnx66aq because the instance already has the maximum number 16 of attached PV volumes for the shape VM.Standard.E5.Flex. To attach this volume, first detach an existing PV volume from the instance, and then try again... Operation Name: AttachVolume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetOCIServiceError(tt.err)
			if !strings.EqualFold(tt.expected, got) {
				t.Errorf("GetOCIServiceError() = %s, want %s", got, tt.expected)
			}
		})
	}
}
func Test_TruncateError(t *testing.T) {
	maxVolumeAttachDetachErrorMsgBytes := 1024

	tests := []struct {
		name        string
		err         error
		maxBytes    int
		expectedErr error
	}{
		{
			name:        "Test the truncation is working properly",
			err:         errors.New("Timeout of waiting for condition. Lot of big text."),
			maxBytes:    21,
			expectedErr: errors.New("Timeout of waiting..."),
		},
		{
			name:        "Test the truncation is working properly for errors created with status.Errorf",
			err:         status.Errorf(codes.Internal, "Timeout of waiting for condition. %s.", "Lot of big texttttttttttttttttt"),
			maxBytes:    55,
			expectedErr: errors.New("rpc error: code = Internal desc = Timeout of waiting..."),
		},
		{
			name:        "error with multi-byte truncate in middle",
			err:         errors.New("éééérror"),
			maxBytes:    6,
			expectedErr: errors.New("é..."),
		},
		{
			name:        "nil error",
			err:         nil,
			maxBytes:    6,
			expectedErr: errors.New(""),
		},
		{
			name:        "empty error",
			err:         errors.New(""),
			maxBytes:    6,
			expectedErr: errors.New(""),
		},
		{
			name:        "ensure suffix is not appended when maxBytes is less than 3",
			err:         errors.New("ab"),
			maxBytes:    1,
			expectedErr: errors.New("a"),
		},
		{
			name: "oci service error is truncated properly",
			err: errors.New(GetOCIServiceError(&mockServiceError{
				TargetService:            "Compute Service",
				StatusCode:               400,
				Code:                     "LimitExceeded",
				OpcRequestID:             "8239a6cbec18bed008924abe8b12416b/35C4EBDF5A7728341854B0685048586B/D66BF5A8491C409D1C61B18D343607A4",
				Message:                  "Cannot attach volume ocid1.volume.oc1.phx.abyhqljrwv2qo35rdomkhvmgtrwr6wpkezzh6gb355tn7phtp5szt5hleqja to instance ocid1.instance.oc1.phx.anyhqljrh4gjgpyc2ighn7uq32unt2e4grzkw3ta37kxbi6c3z3xhlnx66aq because the instance already has the maximum number 16 of attached PV volumes for the shape VM.Standard.E5.Flex. To attach this volume, first detach an existing PV volume from the instance, and then try again..",
				OperationName:            "AttachVolume",
				ClientVersion:            "Oracle-GoSDK/xVersion",
				RequestTarget:            "https://iaas.us-phoenix-1.oraclecloud.com/20160918/volumeAttachments",
				OperationReferenceLink:   "Test",
				ErrorTroubleshootingLink: "Test",
			})),
			maxBytes:    maxVolumeAttachDetachErrorMsgBytes,
			expectedErr: errors.New("Error returned by Compute Service Service. Http Status Code: 400. Error Code: LimitExceeded. Message: Cannot attach volume ocid1.volume.oc1.phx.abyhqljrwv2qo35rdomkhvmgtrwr6wpkezzh6gb355tn7phtp5szt5hleqja to instance ocid1.instance.oc1.phx.anyhqljrh4gjgpyc2ighn7uq32unt2e4grzkw3ta37kxbi6c3z3xhlnx66aq because the instance already has the maximum number 16 of attached PV volumes for the shape VM.Standard.E5.Flex. To attach this volume, first detach an existing PV volume from the instance, and then try again... Operation Name: AttachVolume"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := TruncateError(tt.err, tt.maxBytes)
			if !strings.EqualFold(tt.expectedErr.Error(), gotErr.Error()) {
				t.Errorf("TruncateError() = %v, want %v", gotErr, tt.expectedErr)
			}
		})
	}
}

func TestShortenContextBeforeDeadline(t *testing.T) {
	const (
		tolerance = 10 * time.Millisecond
		buffer    = 50 * time.Millisecond
	)

	tests := []struct {
		name                 string
		parentDeadlineOffset time.Duration
		buffer               time.Duration
		expectCancel         bool
		expectedTimeout      time.Duration
	}{
		{
			name:                 "NoParentDeadline",
			parentDeadlineOffset: 0,
			buffer:               buffer,
			expectCancel:         false,
			expectedTimeout:      0,
		},
		{
			name:                 "DeadlineFarEnough",
			parentDeadlineOffset: 200 * time.Millisecond,
			buffer:               buffer,
			expectCancel:         false,
			expectedTimeout:      150 * time.Millisecond,
		},
		{
			name:                 "DeadlineExactlyBuffer",
			parentDeadlineOffset: buffer,
			buffer:               buffer,
			expectCancel:         true,
			expectedTimeout:      0,
		},
		{
			name:                 "DeadlineLessThanBuffer",
			parentDeadlineOffset: 10 * time.Millisecond,
			buffer:               buffer,
			expectCancel:         true,
			expectedTimeout:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parentCtx context.Context
			var parentCancel context.CancelFunc = func() {}
			now := time.Now()

			if tt.parentDeadlineOffset > 0 {
				parentCtx, parentCancel = context.WithDeadline(context.Background(), now.Add(tt.parentDeadlineOffset))
			} else {
				parentCtx = context.Background()
			}
			defer parentCancel()

			newCtx, cancel := ShortenContextBeforeDeadline(parentCtx, tt.buffer)
			defer cancel()

			if tt.expectCancel {
				if err := newCtx.Err(); err != context.Canceled {
					t.Errorf("ShortenContextBeforeDeadline() returned context.Err() = %v, want %v", err, context.Canceled)
				}
				return
			}

			if newCtx.Err() != nil {
				t.Fatalf("ShortenContextBeforeDeadline() returned context.Err() = %v, want nil", newCtx.Err())
			}

			if tt.parentDeadlineOffset == 0 {
				_, hasDeadline := newCtx.Deadline()
				if hasDeadline {
					t.Errorf("ShortenContextBeforeDeadline() with no parent deadline returned a context with a deadline")
				}
			} else {
				deadline, hasDeadline := newCtx.Deadline()
				if !hasDeadline {
					t.Fatalf("ShortenContextBeforeDeadline() returned a context without a deadline, want deadline")
				}

				actualDuration := deadline.Sub(now)

				expectedMin := tt.expectedTimeout - tolerance
				expectedMax := tt.expectedTimeout + tolerance

				if actualDuration < expectedMin || actualDuration > expectedMax {
					t.Errorf("ShortenContextBeforeDeadline() deadline duration %v, want approx %v", actualDuration, tt.expectedTimeout)
				}

				cancel()
				if err := newCtx.Err(); err != context.Canceled {
					t.Errorf("Cancel() did not cancel context. Got: %v, Want: %v", err, context.Canceled)
				}
			}
		})
	}
}
