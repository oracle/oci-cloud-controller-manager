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

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/pkg/errors"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider"
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

func (cp *CloudProvider) extractNodeAddresses(ctx context.Context, instanceID string) ([]api.NodeAddress, error) {
	addresses := []api.NodeAddress{}
	vnic, err := cp.client.Compute().GetPrimaryVNICForInstance(ctx, cp.config.CompartmentID, instanceID)
	if err != nil {
		return nil, errors.Wrap(err, "GetPrimaryVNICForInstance")
	}

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

	// OKE does not support setting DNS since this changes the override hostname we setup to be the ip address.
	// Changing this can have wide reaching impact.
	//
	// if vnic.HostnameLabel != nil && *vnic.HostnameLabel != "" {
	// 	subnet, err := cp.client.Networking().GetSubnet(ctx, *vnic.SubnetId)
	// 	if err != nil {
	// 		return nil, errors.Wrap(err, "GetSubnetForInstance")
	// 	}
	// 	if subnet != nil && subnet.DnsLabel != nil && *subnet.DnsLabel != "" {
	// 		vcn, err := cp.client.Networking().GetVcn(ctx, *subnet.VcnId)
	// 		if err != nil {
	// 			return nil, errors.Wrap(err, "GetVcnForInstance")
	// 		}
	// 		if vcn != nil && vcn.DnsLabel != nil && *vcn.DnsLabel != "" {
	// 			fqdn := strings.Join([]string{*vnic.HostnameLabel, *subnet.DnsLabel, *vcn.DnsLabel, "oraclevcn.com"}, ".")
	// 			addresses = append(addresses, api.NodeAddress{Type: api.NodeHostName, Address: fqdn})
	// 			addresses = append(addresses, api.NodeAddress{Type: api.NodeInternalDNS, Address: fqdn})
	// 		}
	// 	}
	// }

	return addresses, nil
}

// NodeAddresses returns the addresses of the specified instance.
// TODO(roberthbailey): This currently is only used in such a way that it
// returns the address of the calling instance. We should do a rename to
// make this clearer.
func (cp *CloudProvider) NodeAddresses(ctx context.Context, name types.NodeName) ([]api.NodeAddress, error) {
	cp.logger.With("nodeName", name).Debug("Getting node addresses")

	inst, err := cp.client.Compute().GetInstanceByNodeName(ctx, cp.config.CompartmentID, cp.config.VCNID, mapNodeNameToInstanceName(name))
	if err != nil {
		return nil, errors.Wrap(err, "GetInstanceByNodeName")
	}
	return cp.extractNodeAddresses(ctx, *inst.Id)
}

// NodeAddressesByProviderID returns the addresses of the specified instance.
// The instance is specified using the providerID of the node. The ProviderID is
// a unique identifier of the node. This will not be called from the node whose
// nodeaddresses are being queried. i.e. local metadata services cannot be used
// in this method to obtain nodeaddresses.
func (cp *CloudProvider) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]api.NodeAddress, error) {
	cp.logger.With("instanceID", providerID).Debug("Getting node addresses by provider id")

	instanceID, err := MapProviderIDToInstanceID(providerID)
	if err != nil {
		return nil, errors.Wrap(err, "MapProviderIDToInstanceID")
	}
	return cp.extractNodeAddresses(ctx, instanceID)

}

// InstanceID returns the cloud provider ID of the node with the specified NodeName.
func (cp *CloudProvider) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	cp.logger.With("nodeName", nodeName).Debug("Getting instance id for node name")

	name := mapNodeNameToInstanceName(nodeName)
	inst, err := cp.client.Compute().GetInstanceByNodeName(ctx, cp.config.CompartmentID, cp.config.VCNID, name)
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
	cp.logger.With("nodeName", name).Debug("Getting instance type by node name")

	inst, err := cp.client.Compute().GetInstanceByNodeName(ctx, cp.config.CompartmentID, cp.config.VCNID, mapNodeNameToInstanceName(name))
	if err != nil {
		return "", errors.Wrap(err, "GetInstanceByNodeName")
	}
	return *inst.Shape, nil
}

// InstanceTypeByProviderID returns the type of the specified instance.
func (cp *CloudProvider) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	cp.logger.With("instanceID", providerID).Debug("Getting instance type by provider id")

	instanceID, err := MapProviderIDToInstanceID(providerID)
	if err != nil {
		return "", errors.Wrap(err, "MapProviderIDToInstanceID")
	}
	inst, err := cp.client.Compute().GetInstance(ctx, instanceID)
	if err != nil {
		return "", errors.Wrap(err, "GetInstance")
	}
	return *inst.Shape, nil
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (cp *CloudProvider) AddSSHKeyToAllInstances(ctx context.Context, user string, keyData []byte) error {
	return cloudprovider.NotImplemented
}

// CurrentNodeName returns the name of the node we are currently running on
// On most clouds (e.g. GCE) this is the hostname, so we provide the hostname
func (cp *CloudProvider) CurrentNodeName(ctx context.Context, hostname string) (types.NodeName, error) {
	return "", cloudprovider.NotImplemented

}

// InstanceExistsByProviderID returns true if the instance for the given
// provider id still is running. If false is returned with no error, the
// instance will be immediately deleted by the cloud controller manager.
func (cp *CloudProvider) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	cp.logger.With("instanceID", providerID).Debug("Checking instance exists by provider id")
	instanceID, err := MapProviderIDToInstanceID(providerID)
	if err != nil {
		return false, err
	}
	instance, err := cp.client.Compute().GetInstance(ctx, instanceID)
	if client.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !client.IsInstanceInTerminalState(instance), nil
}

// InstanceShutdownByProviderID returns true if the instance is shutdown in cloudprovider.
func (cp *CloudProvider) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {
	cp.logger.With("instanceID", providerID).Debug("Checking instance is stopped by provider id")
	instanceID, err := MapProviderIDToInstanceID(providerID)
	if err != nil {
		return false, err
	}

	instance, err := cp.client.Compute().GetInstance(ctx, instanceID)
	if err != nil {
		return false, err
	}

	return client.IsInstanceInStoppedState(instance), nil
}
