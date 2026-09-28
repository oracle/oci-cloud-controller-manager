package driver

import (
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidateFSSExportPath(t *testing.T) {
	tests := []struct {
		name             string
		exportPath       string
		encryptInTransit bool
		wantCode         codes.Code
	}{
		{
			name:             "encrypted ASCII space",
			exportPath:       "/small-filesystem -o ro",
			encryptInTransit: true,
			wantCode:         codes.InvalidArgument,
		},
		{
			name:             "encrypted tab",
			exportPath:       "/small-filesystem\t-o\tro",
			encryptInTransit: true,
			wantCode:         codes.InvalidArgument,
		},
		{
			name:             "encrypted newline",
			exportPath:       "/small-filesystem\n-o ro",
			encryptInTransit: true,
			wantCode:         codes.InvalidArgument,
		},
		{
			name:             "encrypted Unicode whitespace",
			exportPath:       "/small-filesystem\u00a0-o-ro",
			encryptInTransit: true,
			wantCode:         codes.InvalidArgument,
		},
		{
			name:             "encrypted path without whitespace",
			exportPath:       "/small-filesystem",
			encryptInTransit: true,
			wantCode:         codes.OK,
		},
		{
			name:             "unencrypted path with whitespace",
			exportPath:       "/small-filesystem -o ro",
			encryptInTransit: false,
			wantCode:         codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFSSExportPath(tt.exportPath, tt.encryptInTransit)
			if got := status.Code(err); got != tt.wantCode {
				t.Fatalf("expected gRPC code %s, got %s: %v", tt.wantCode, got, err)
			}
			if tt.wantCode == codes.OK {
				return
			}
			if got := status.Convert(err).Message(); got != "FSS export path must not contain whitespace when encryptInTransit is enabled" {
				t.Fatalf("unexpected error message: %q", got)
			}
		})
	}
}

func TestNodeStageVolumeRejectsWhitespaceInEncryptedExportPath(t *testing.T) {
	driver := FSSNodeDriver{
		NodeDriver: NodeDriver{
			logger: logging.Logger().Sugar(),
			nodeMetadata: &csi_util.NodeMetadata{
				IsNodeMetadataLoaded: true,
				Ipv4Enabled:          true,
			},
		},
	}
	req := &csi.NodeStageVolumeRequest{
		VolumeId:          "ocid1.filesystem.oc1.phx.test:10.0.10.147:/small-filesystem -o ro",
		StagingTargetPath: "/staging-path",
		VolumeCapability: &csi.VolumeCapability{
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{},
			},
		},
		VolumeContext: map[string]string{"encryptInTransit": "true"},
	}

	_, err := driver.NodeStageVolume(t.Context(), req)
	if got := status.Code(err); got != codes.InvalidArgument {
		t.Fatalf("expected gRPC code %s, got %s: %v", codes.InvalidArgument, got, err)
	}
	if got := status.Convert(err).Message(); got != "FSS export path must not contain whitespace when encryptInTransit is enabled" {
		t.Fatalf("unexpected error message: %q", got)
	}
}
