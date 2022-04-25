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
	"strings"

	"github.com/oracle/oci-go-sdk/v50/core"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cloud-provider"
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
func (cp *CloudProvider) GetZone(ctx context.Context) (cloudprovider.Zone, error) {
	return cloudprovider.Zone{}, cloudprovider.NotImplemented
}

// GetZoneByProviderID returns the Zone containing the current zone and
// locality region of the node specified by providerID This method is
// particularly used in the context of external cloud providers where node
// initialization must be down outside the kubelets.
func (cp *CloudProvider) GetZoneByProviderID(ctx context.Context, providerID string) (cloudprovider.Zone, error) {
	instanceID, err := MapProviderIDToInstanceID(providerID)
	if err != nil {
		return cloudprovider.Zone{}, err
	}
	item, exists, err := cp.instanceCache.GetByKey(instanceID)
	if err != nil {
		return cloudprovider.Zone{}, errors.Wrap(err, "Error fetching instance from instanceCache, will retry")
	}
	if exists {
		return cloudprovider.Zone{
			FailureDomain: mapAvailabilityDomainToFailureDomain(*item.(*core.Instance).AvailabilityDomain),
			Region:        *item.(*core.Instance).Region,
		}, nil
	}
	instance, err := cp.client.Compute().GetInstance(ctx, instanceID)
	if err != nil {
		return cloudprovider.Zone{}, err
	}
	if err := cp.instanceCache.Add(instance); err != nil {
		return cloudprovider.Zone{}, errors.Wrap(err, "Failed to add instance in instanceCache")
	}
	return cloudprovider.Zone{
		FailureDomain: mapAvailabilityDomainToFailureDomain(*instance.AvailabilityDomain),
		Region:        *instance.Region,
	}, nil
}

// GetZoneByNodeName returns the Zone containing the current zone and locality
// region of the node specified by node name This method is particularly used
// in the context of external cloud providers where node initialization must be
// down outside the kubelets.
func (cp *CloudProvider) GetZoneByNodeName(ctx context.Context, nodeName types.NodeName) (cloudprovider.Zone, error) {
	compartmentID, err := cp.getCompartmentIDByNodeName(mapNodeNameToInstanceName(nodeName))
	if err != nil {
		return cloudprovider.Zone{}, errors.Wrap(err, "Error getting CompartmentID from Node Name")
	}
	instance, err := cp.client.Compute().GetInstanceByNodeName(ctx, compartmentID, cp.config.VCNID, mapNodeNameToInstanceName(nodeName))
	if err != nil {
		return cloudprovider.Zone{}, err
	}
	return cloudprovider.Zone{
		FailureDomain: mapAvailabilityDomainToFailureDomain(*instance.AvailabilityDomain),
		Region:        *instance.Region,
	}, nil
}
