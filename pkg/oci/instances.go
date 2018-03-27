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
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/golang/glog"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/pkg/errors"

	api "k8s.io/api/core/v1"
	types "k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/kubernetes/pkg/cloudprovider"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/util"
)

var _ cloudprovider.Instances = &CloudProvider{}

// mapNodeNameToInstanceName maps a kube NodeName to a OCI instance display
// name.
func mapNodeNameToInstanceName(nodeName types.NodeName) string {
	return string(nodeName)
}

// mapInstanceToNodeName maps a OCI instance display name to a kube NodeName.
func mapInstanceNameToNodeName(displayName string) types.NodeName {
	// Node names are always lowercase
	return types.NodeName(strings.ToLower(displayName))
}

func extractNodeAddressesFromVNIC(vnic *core.Vnic) ([]api.NodeAddress, error) {
	addresses := []api.NodeAddress{}
	if vnic == nil {
		return addresses, nil
	}

	if vnic.PrivateIp != nil && *vnic.PrivateIp != "" {
		ip := net.ParseIP(*vnic.PrivateIp)
		if ip == nil {
			return nil, fmt.Errorf("instance has invalid private address: %q", *vnic.PrivateIp)
		}
		addresses = append(addresses, api.NodeAddress{Type: api.NodeInternalIP, Address: ip.String()})
	}

	if vnic.PublicIp != nil && *vnic.PublicIp != "" {
		ip := net.ParseIP(*vnic.PublicIp)
		if ip == nil {
			return nil, errors.Errorf("instance has invalid public address: %q", *vnic.PublicIp)
		}
		addresses = append(addresses, api.NodeAddress{Type: api.NodeExternalIP, Address: ip.String()})
	}

	return addresses, nil
}

// NodeAddresses returns the addresses of the specified instance.
// TODO(roberthbailey): This currently is only used in such a way that it
// returns the address of the calling instance. We should do a rename to
// make this clearer.
func (cp *CloudProvider) NodeAddresses(ctx context.Context, name types.NodeName) ([]api.NodeAddress, error) {
	glog.V(4).Infof("NodeAddresses(%q) called", name)

	inst, err := cp.client.Compute().GetInstanceByNodeName(ctx, mapNodeNameToInstanceName(name))
	if err != nil {
		return nil, errors.Wrap(err, "GetInstanceByNodeName")
	}

	vnic, err := cp.client.Compute().GetPrimaryVNICForInstance(ctx, *inst.Id)
	if err != nil {
		return nil, errors.Wrap(err, "GetPrimaryVNICForInstance")
	}
	return extractNodeAddressesFromVNIC(vnic)
}

// NodeAddressesByProviderID returns the addresses of the specified instance.
// The instance is specified using the providerID of the node. The ProviderID is
// a unique identifier of the node. This will not be called from the node whose
// nodeaddresses are being queried. i.e. local metadata services cannot be used
// in this method to obtain nodeaddresses.
func (cp *CloudProvider) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]api.NodeAddress, error) {
	glog.V(4).Infof("NodeAddressesByProviderID(%q) called", providerID)
	instanceID := util.MapProviderIDToInstanceID(providerID)
	vnic, err := cp.client.Compute().GetPrimaryVNICForInstance(ctx, instanceID)
	if err != nil {
		return nil, errors.Wrap(err, "GetPrimaryVNICForInstance")
	}
	return extractNodeAddressesFromVNIC(vnic)
}

// ExternalID returns the cloud provider ID of the node with the specified NodeName.
// Note that if the instance does not exist or is no longer running, we must
// return ("", cloudprovider.InstanceNotFound).
func (cp *CloudProvider) ExternalID(ctx context.Context, nodeName types.NodeName) (string, error) {
	glog.V(4).Infof("ExternalID(%q) called", nodeName)

	instName := mapNodeNameToInstanceName(nodeName)
	inst, err := cp.client.Compute().GetInstanceByNodeName(ctx, instName)
	if client.IsNotFound(err) {
		glog.Infof("Instance %q was not found. Unable to get ExternalID: %v", instName, err)
		return "", cloudprovider.InstanceNotFound
	}
	if err != nil {
		glog.Errorf("Failed to get ExternalID of %s: %v", nodeName, err)
		return "", err
	}

	glog.V(4).Infof("Got ExternalID %s for %s", *inst.Id, nodeName)
	return *inst.Id, nil
}

// InstanceID returns the cloud provider ID of the node with the specified NodeName.
func (cp *CloudProvider) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	glog.V(4).Infof("InstanceID(%q) called", nodeName)

	name := mapNodeNameToInstanceName(nodeName)
	inst, err := cp.client.Compute().GetInstanceByNodeName(ctx, name)
	if err != nil {
		if client.IsNotFound(err) {
			return "", cloudprovider.InstanceNotFound
		}
		return "", errors.Wrap(err, "GetInstanceByNodeName")
	}
	return *inst.Id, nil
}

// InstanceType returns the type of the specified instance.
func (cp *CloudProvider) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	glog.V(4).Infof("InstanceType(%q) called", name)

	inst, err := cp.client.Compute().GetInstanceByNodeName(ctx, mapNodeNameToInstanceName(name))
	if err != nil {
		return "", errors.Wrap(err, "GetInstanceByNodeName")
	}
	return *inst.Shape, nil
}

// InstanceTypeByProviderID returns the type of the specified instance.
func (cp *CloudProvider) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	glog.V(4).Infof("InstanceTypeByProviderID(%q) called", providerID)

	instanceID := util.MapProviderIDToInstanceID(providerID)
	inst, err := cp.client.Compute().GetInstance(ctx, instanceID)
	if err != nil {
		return "", errors.Wrap(err, "GetInstance")
	}
	return *inst.Shape, nil
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (cp *CloudProvider) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return errors.New("unimplemented")
}

// CurrentNodeName returns the name of the node we are currently running on
// On most clouds (e.g. GCE) this is the hostname, so we provide the hostname
func (cp *CloudProvider) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	glog.V(4).Infof("CurrentNodeName(%q) called", hostname)
	return "", errors.New("unimplemented")
}

// InstanceExistsByProviderID returns true if the instance for the given
// provider id still is running. If false is returned with no error, the
// instance will be immediately deleted by the cloud controller manager.
func (cp *CloudProvider) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	glog.V(4).Infof("InstanceExistsByProviderID(%q) called", providerID)
	instanceID := util.MapProviderIDToInstanceID(providerID)
	instance, err := cp.client.Compute().GetInstance(ctx, instanceID)
	if client.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !client.IsInstanceInTerminalState(instance), nil
}
