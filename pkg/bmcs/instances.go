// Copyright 2017 The Oracle Kubernetes Cloud Controller Manager Authors
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

package bmcs

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang/glog"

	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

var _ cloudprovider.Instances = &CloudProvider{}

// mapNodeNameToInstanceName maps a kube NodeName to a BMCS instance display
// name.
func mapNodeNameToInstanceName(nodeName types.NodeName) string {
	return string(nodeName)
}

// mapInstanceToNodeName maps a BMCS instance display name to a kube NodeName.
func mapInstanceNameToNodeName(displayName string) types.NodeName {
	// Node names are always lowercase
	return types.NodeName(strings.ToLower(displayName))
}

// NodeAddresses returns the addresses of the specified instance.
// TODO(roberthbailey): This currently is only used in such a way that it
// returns the address of the calling instance. We should do a rename to
// make this clearer.
func (cp *CloudProvider) NodeAddresses(name types.NodeName) ([]api.NodeAddress, error) {
	glog.V(4).Infof("NodeAddresses(%q) called", name)

	inst, err := cp.client.GetInstanceByNodeName(mapNodeNameToInstanceName(name))
	if err != nil {
		return nil, err
	}
	return cp.client.GetNodeAddressesForInstance(inst.ID)
}

// NodeAddressesByProviderID returns the addresses of the specified instance.
// The instance is specified using the providerID of the node. The ProviderID is
// a unique identifier of the node. This will not be called from the node whose
// nodeaddresses are being queried. i.e. local metadata services cannot be used
// in this method to obtain nodeaddresses.
func (cp *CloudProvider) NodeAddressesByProviderID(providerID string) ([]api.NodeAddress, error) {
	glog.V(4).Infof("NodeAddressesByProviderID(%q) called", providerID)
	return cp.client.GetNodeAddressesForInstance(providerID)
}

// ExternalID returns the cloud provider ID of the node with the specified NodeName.
// Note that if the instance does not exist or is no longer running, we must
// return ("", cloudprovider.InstanceNotFound).
func (cp *CloudProvider) ExternalID(nodeName types.NodeName) (string, error) {
	glog.V(4).Infof("ExternalID(%q) called", nodeName)

	instName := mapNodeNameToInstanceName(nodeName)
	inst, err := cp.client.GetInstanceByNodeName(instName)
	if err != nil {
		glog.Errorf("Failed to get External id of %s: %v", nodeName, err)
		return "", cloudprovider.InstanceNotFound
	}

	glog.V(4).Infof("Got id %s for %s", inst.ID, nodeName)
	return inst.ID, nil
}

// InstanceID returns the cloud provider ID of the node with the specified NodeName.
// TODO (apryde): AWS and GCE use format /<zone>/<instanceid> - should we?
func (cp *CloudProvider) InstanceID(nodeName types.NodeName) (string, error) {
	glog.V(4).Infof("InstanceID(%q) called", nodeName)

	inst, err := cp.client.GetInstanceByNodeName(mapNodeNameToInstanceName(nodeName))
	if err != nil {

	}
	return inst.ID, nil
}

// InstanceType returns the type of the specified instance.
func (cp *CloudProvider) InstanceType(name types.NodeName) (string, error) {
	glog.V(4).Infof("InstanceType(%q) called", name)

	inst, err := cp.client.GetInstanceByNodeName(mapNodeNameToInstanceName(name))
	if err != nil {
		return "", fmt.Errorf("getInstanceByNodeName failed for %q with %v", name, err)
	}
	return inst.Shape, nil
}

// InstanceTypeByProviderID returns the type of the specified instance.
func (cp *CloudProvider) InstanceTypeByProviderID(providerID string) (string, error) {
	glog.V(4).Infof("InstanceTypeByProviderID(%q) called", providerID)

	inst, err := cp.client.GetInstance(providerID)
	if err != nil {
		return "", err
	}
	return inst.Shape, nil
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (cp *CloudProvider) AddSSHKeyToAllInstances(user string, keyData []byte) error {
	return errors.New("unimplemented")
}

// CurrentNodeName returns the name of the node we are currently running on
// On most clouds (e.g. GCE) this is the hostname, so we provide the hostname
func (cp *CloudProvider) CurrentNodeName(hostname string) (types.NodeName, error) {
	glog.V(4).Infof("CurrentNodeName(%q) called", hostname)
	return "", errors.New("unimplemented")
}

// InstanceExistsByProviderID returns true if the instance for the given
// provider id still is running. If false is returned with no error, the
// instance will be immediately deleted by the cloud controller manager.
func (cp *CloudProvider) InstanceExistsByProviderID(providerID string) (bool, error) {
	glog.V(4).Infof("InstanceExistsByProviderID(%q) called", providerID)
	r, err := cp.client.GetInstance(providerID)
	return (r != nil), err
}
