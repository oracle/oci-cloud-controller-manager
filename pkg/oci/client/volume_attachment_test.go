package client

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func Test_getDevicePath(t *testing.T) {
	var tests = map[string]struct {
		instanceID string
		want       string
		wantErr    error
	}{
		"getDevicePathNoDeviceAvailable": {
			instanceID: "ocid1.device-path-not-available",
			wantErr:    fmt.Errorf("Max number of volumes are already attached to instance %s. Please schedule workload on different node.", "ocid1.device-path-not-available"),
		},
		"getDevicePathOneDeviceAvailable": {
			instanceID: "ocid1.one-device-path-available",
			want:       "/dev/oracleoci/oraclevdac",
		},
		"getDevicePathReturnsError": {
			instanceID: "ocid1.device-path-returns-error",
			wantErr:    errNotFound,
		},
	}

	vaClient := &client{
		compute: &mockComputeClient{},
		logger:  zap.S(),
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := vaClient.getDevicePath(context.Background(), tc.instanceID)
			if tc.wantErr != nil && !strings.EqualFold(tc.wantErr.Error(), err.Error()) {
				t.Errorf("getDevicePath() = %v, want %v", err.Error(), tc.wantErr.Error())
			}
			if tc.want != "" && !strings.EqualFold(tc.want, *result) {
				t.Errorf("getDevicePath() = %v, want %v", *result, tc.want)
			}
		})
	}
}
