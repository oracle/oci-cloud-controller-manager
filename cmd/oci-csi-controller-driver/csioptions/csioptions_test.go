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
