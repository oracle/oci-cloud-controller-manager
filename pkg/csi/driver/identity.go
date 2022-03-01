package driver

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// GetPluginInfo returns metadata of the plugin
func (d *Driver) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	resp := &csi.GetPluginInfoResponse{
		Name:          d.name,
		VendorVersion: d.version,
	}

	return resp, nil
}

// GetPluginCapabilities returns available capabilities of the plugin
func (d *Driver) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	accessibilityConstraints := csi.PluginCapability_Service_{
		Service: &csi.PluginCapability_Service{
			Type: csi.PluginCapability_Service_VOLUME_ACCESSIBILITY_CONSTRAINTS,
		},
	}
	pluginCapabilityService := csi.PluginCapability_Service_{
		Service: &csi.PluginCapability_Service{
			Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
		},
	}

	var capabilities []*csi.PluginCapability
	if d.name == BlockVolumeDriverName {
		capabilities = []*csi.PluginCapability{
			{
				Type: &pluginCapabilityService,
			},
			{
				Type: &accessibilityConstraints,
			},
		}
	} else {
		capabilities = []*csi.PluginCapability{
			{
				Type: &pluginCapabilityService,
			},
		}
	}
	resp := &csi.GetPluginCapabilitiesResponse{
		Capabilities: capabilities,
	}

	return resp, nil
}

// Probe returns the health and readiness of the plugin
func (d *Driver) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {

	d.readyMu.Lock()
	defer d.readyMu.Unlock()

	return &csi.ProbeResponse{
		Ready: &wrappers.BoolValue{
			Value: d.ready,
		},
	}, nil
}
