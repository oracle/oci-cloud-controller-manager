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
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/pkg/errors"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
)

const (
	VirtualNodePoolIdAnnotation = "oci.oraclecloud.com/virtual-node-pool-id"
	IPv4NodeIPFamilyLabel       = "oci.oraclecloud.com/ip-family-ipv4"
	IPv6NodeIPFamilyLabel       = "oci.oraclecloud.com/ip-family-ipv6"
)

var _ cloudprovider.Instances = &CloudProvider{}

// mapNodeNameToInstanceName maps a kube NodeName to a OCI instance display
// name.
func mapNodeNameToInstanceName(nodeName types.NodeName) string {
	return string(nodeName)
}

func (cp *CloudProvider) getCompartmentIDByInstanceID(instanceID string) (string, error) {
	item, exists, err := cp.instanceCache.GetByKey(instanceID)
	if err != nil {
		return "", errors.Wrap(err, "error fetching instance from instanceCache")
	}
	if exists {
		return *item.(*core.Instance).CompartmentId, nil
	}
	nodeList, err := cp.NodeLister.List(labels.Everything())
	if err != nil {
		return "", errors.Wrap(err, "error listing all the nodes using node informer")
	}
	for _, node := range nodeList {
		providerID, err := MapProviderIDToResourceID(node.Spec.ProviderID)
		if err != nil {
			return "", errors.New("Failed to map providerID to instanceID")
		}
		if providerID == instanceID {
			if compartmentID, ok := node.Annotations[CompartmentIDAnnotation]; ok {
				if compartmentID != "" {
					return compartmentID, nil
				}
			}
		}
	}
	return "", errors.New("compartmentID annotation missing in the node. Would retry")
}

func (cp *CloudProvider) extractNodeAddresses(ctx context.Context, instanceID string) ([]api.NodeAddress, error) {
	var addresses []api.NodeAddress
	compartmentID, err := cp.getCompartmentIDByInstanceID(instanceID)
	if err != nil {
		return nil, err
	}

	vnic, err := cp.client.Compute().GetPrimaryVNICForInstance(ctx, compartmentID, instanceID)
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
	nodeIpFamily, err := cp.getNodeIpFamily(instanceID)
	if err != nil {
		return nil, err
	}
	if contains(nodeIpFamily, IPv6) {
		if vnic.Ipv6Addresses != nil {
			for _, ipv6Addresses := range vnic.Ipv6Addresses {
				if ipv6Addresses != "" {
					ip := net.ParseIP(ipv6Addresses)
					if ip == nil {
						return nil, errors.Errorf("instance has invalid ipv6 address: %q", vnic.Ipv6Addresses[0])
					}
					if ip.IsPrivate() {
						addresses = append(addresses, api.NodeAddress{Type: api.NodeInternalIP, Address: ip.String()})
					} else {
						addresses = append(addresses, api.NodeAddress{Type: api.NodeExternalIP, Address: ip.String()})
					}
				}
			}
		}
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

// getNodeIpFamily checks if label exists in the Node
// oci.oraclecloud.com/ip-family-ipv4
// oci.oraclecloud.com/ip-family-ipv6
func (cp *CloudProvider) getNodeIpFamily(instanceId string) ([]string, error) {
	nodeIpFamily := []string{}
	nodeList, err := cp.NodeLister.List(labels.Everything())
	if err != nil {
		return nodeIpFamily, errors.Wrap(err, "error listing nodes using node informer")
	}

	// TODO: @prasrira Add a cache to determine nodes that already have label https://github.com/kubernetes/client-go/blob/master/tools/cache/expiration_cache.go
	for _, node := range nodeList {
		providerID, err := MapProviderIDToResourceID(node.Spec.ProviderID)
		if err != nil {
			return nodeIpFamily, errors.New("Failed to map providerID to instanceID")
		}
		if providerID == instanceId {
			if _, ok := node.Labels[IPv4NodeIPFamilyLabel]; ok {
				nodeIpFamily = append(nodeIpFamily, IPv4)
			}
			if _, ok := node.Labels[IPv6NodeIPFamilyLabel]; ok {
				nodeIpFamily = append(nodeIpFamily, IPv6)
			}
		}
	}
	if len(nodeIpFamily) != 0 {
		cp.logger.Debugf("NodeIpFamily is %s for instance id %s", strings.Join(nodeIpFamily, ","), instanceId)
	}
	return nodeIpFamily, nil
}

// NodeAddresses returns the addresses of the specified instance.
// TODO(roberthbailey): This currently is only used in such a way that it
// returns the address of the calling instance. We should do a rename to
// make this clearer.
func (cp *CloudProvider) NodeAddresses(ctx context.Context, name types.NodeName) ([]api.NodeAddress, error) {
	cp.logger.With("nodeName", name).Debug("Getting node addresses")

	nodeName := mapNodeNameToInstanceName(name)
	compartmentID, err := cp.getCompartmentIDByNodeName(nodeName)
	if err != nil {
		return nil, errors.Wrap(err, "error getting CompartmentID from Node Name")
	}
	inst, err := cp.client.Compute().GetInstanceByNodeName(ctx, compartmentID, cp.config.VCNID, mapNodeNameToInstanceName(name))
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
	cp.logger.With("resourceID", providerID).Debug("Getting node addresses by provider id")

	resourceID, err := MapProviderIDToResourceID(providerID)
	if err != nil {
		return nil, errors.Wrap(err, "MapProviderIDToResourceOCID")
	}

	return cp.extractNodeAddresses(ctx, resourceID)
}

// InstanceID returns the cloud provider ID of the node with the specified NodeName.
func (cp *CloudProvider) InstanceID(ctx context.Context, nodeName types.NodeName) (string, error) {
	cp.logger.With("nodeName", nodeName).Debug("Getting instance id for node name")

	name := mapNodeNameToInstanceName(nodeName)
	compartmentID, err := cp.getCompartmentIDByNodeName(name)
	if err != nil {
		if cp.config.CompartmentID != "" {
			compartmentID = cp.config.CompartmentID
		} else {
			return "", errors.Wrap(err, "error getting CompartmentID from Node Name")
		}
	}
	inst, err := cp.client.Compute().GetInstanceByNodeName(ctx, compartmentID, cp.config.VCNID, name)
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
	compartmentID, err := cp.getCompartmentIDByNodeName(mapNodeNameToInstanceName(name))
	if err != nil {
		return "", errors.Wrap(err, "error getting CompartmentID from Node Name")
	}
	inst, err := cp.client.Compute().GetInstanceByNodeName(ctx, compartmentID, cp.config.VCNID, mapNodeNameToInstanceName(name))
	if err != nil {
		return "", errors.Wrap(err, "GetInstanceByNodeName")
	}
	return *inst.Shape, nil
}

// InstanceTypeByProviderID returns the type of the specified instance.
func (cp *CloudProvider) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	cp.logger.With("resourceID", providerID).Debug("Getting instance type by provider id")

	resourceID, err := MapProviderIDToResourceID(providerID)
	if err != nil {
		return "", errors.Wrap(err, "MapProviderIDToResourceOCID")
	}

	item, exists, err := cp.instanceCache.GetByKey(resourceID)
	if err != nil {
		return "", errors.Wrap(err, "error fetching instance from instanceCache, will retry")
	}
	if exists {
		return *item.(*core.Instance).Shape, nil
	}
	cp.logger.Debug("Unable to find the instance information from instanceCache. Calling OCI API")
	inst, err := cp.client.Compute().GetInstance(ctx, resourceID)
	if err != nil {
		return "", errors.Wrap(err, "GetInstance")
	}
	if err := cp.instanceCache.Add(inst); err != nil {
		return "", errors.Wrap(err, "failed to add instance in instanceCache")
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
	//Please do not try to optimise it by using Cache because we prefer correctness over efficiency here
	cp.logger.With("resourceID", providerID).Debug("Checking instance exists by provider id")
	resourceID, err := MapProviderIDToResourceID(providerID)
	if err != nil {
		return false, err
	}
	instance, err := cp.client.Compute().GetInstance(ctx, resourceID)
	if client.IsNotFound(err) {
		return cp.checkForAuthorizationError(ctx, providerID)
	}
	if err != nil {
		return false, err
	}

	return !client.IsInstanceInTerminalState(instance), nil
}

func (cp *CloudProvider) checkForAuthorizationError(ctx context.Context, instanceId string) (bool, error) {
	cp.logger.With("instanceId", instanceId).Info("Received 404 for an instance, listing instances to check for authorization errors")
	compartmentId, err := cp.getCompartmentIDByInstanceID(instanceId)
	if err != nil {
		return false, err
	}
	// to eliminate AD specific issues, list all ADs and make AD specific requests
	availabilityDomains, err := cp.client.Identity().ListAvailabilityDomains(ctx, compartmentId)
	for _, availabilityDomain := range availabilityDomains {
		instances, err := cp.client.Compute().ListInstancesByCompartmentAndAD(ctx, compartmentId, *availabilityDomain.Name)
		// if we are getting errors for ListInstances the issue can be authorization or other issues
		// so to be safe we return the error back as we can't list instances in the compartment
		if err != nil {
			cp.logger.With("instanceId", instanceId).Info("Received error when listing instances to check for authorization errors")
			return false, err
		}

		for _, instance := range instances {
			if *instance.Id == instanceId {
				// Can only happen if changes are done in policy in-between requests
				return true, nil
			}
		}
	}

	return false, nil
}

// InstanceShutdownByProviderID returns true if the instance is shutdown in cloudprovider.
func (cp *CloudProvider) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {
	//Please do not try to optimise it by using InstanceCache because we prefer correctness over efficiency here
	cp.logger.With("resourceID", providerID).Debug("Checking instance is stopped by provider id")
	resourceID, err := MapProviderIDToResourceID(providerID)
	if err != nil {
		return false, err
	}

	instance, err := cp.client.Compute().GetInstance(ctx, resourceID)
	if err != nil {
		return false, err
	}

	return client.IsInstanceInStoppedState(instance), nil
}

func (cp *CloudProvider) getCompartmentIDByNodeName(nodeName string) (string, error) {
	node, err := cp.NodeLister.Get(nodeName)
	if err != nil {
		cp.logger.Errorf("Error getting node using node informer %v", err)
		return "", errors.Wrap(err, "error getting node using node informer")
	}
	if compartmentID, present := node.ObjectMeta.Annotations[CompartmentIDAnnotation]; present {
		if compartmentID != "" {
			return compartmentID, nil
		}
	}
	cp.logger.Debug("CompartmentID annotation is not present")
	return "", errors.New("compartmentID annotation missing in the node. Would retry")
}
