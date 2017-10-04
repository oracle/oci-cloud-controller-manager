// Copyright 2017 The OCI Cloud Controller Manager Authors
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
	"errors"
	"strings"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/util"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

var _ cloudprovider.Zones = &CloudProvider{}

// mapAvailabilityDomainToFailureDomain maps a given Availability Domain to a
// k8s label compat. string.
func mapAvailabilityDomainToFailureDomain(AD string) string {
	parts := strings.SplitN(AD, ":", 2)
	if parts == nil {
		return ""
	}
	return parts[len(parts)-1]
}

// GetZone returns the Zone containing the current failure zone and locality
// region that the program is running in.
func (cp *CloudProvider) GetZone() (zone cloudprovider.Zone, err error) {
	return cloudprovider.Zone{}, errors.New("unimplemented")
}

// GetZoneByProviderID returns the Zone containing the current zone and
// locality region of the node specified by providerID This method is
// particularly used in the context of external cloud providers where node
// initialization must be down outside the kubelets.
func (cp *CloudProvider) GetZoneByProviderID(providerID string) (cloudprovider.Zone, error) {
	instanceID := util.MapProviderIDToInstanceID(providerID)
	instance, err := cp.client.GetInstance(instanceID)
	if err != nil {
		return cloudprovider.Zone{}, err
	}
	return cloudprovider.Zone{
		FailureDomain: mapAvailabilityDomainToFailureDomain(instance.AvailabilityDomain),
		Region:        instance.Region,
	}, nil
}

// GetZoneByNodeName returns the Zone containing the current zone and locality
// region of the node specified by node name This method is particularly used
// in the context of external cloud providers where node initialization must be
// down outside the kubelets.
func (cp *CloudProvider) GetZoneByNodeName(nodeName types.NodeName) (cloudprovider.Zone, error) {
	instance, err := cp.client.GetInstanceByNodeName(mapNodeNameToInstanceName(nodeName))
	if err != nil {
		return cloudprovider.Zone{}, err
	}
	return cloudprovider.Zone{
		FailureDomain: mapAvailabilityDomainToFailureDomain(instance.AvailabilityDomain),
		Region:        instance.Region,
	}, nil
}
