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

package core

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/identity"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	metav1 "k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/pkg/volume/util"
)

// chooseAvailabilityDomain selects the availability zone using the ZoneFailureDomain labels
// on the nodes. This only works if the nodes have been labeled by either the CCM or some other method.
func (p *OCIProvisioner) chooseAvailabilityDomain(ctx context.Context, pvc *v1.PersistentVolumeClaim) (string, *identity.AvailabilityDomain, error) {
	var (
		availabilityDomainName string
		ok                     bool
	)

	if pvc.Spec.Selector != nil {
		// Try the standard Kube label
		availabilityDomainName, ok = pvc.Spec.Selector.MatchLabels[metav1.LabelZoneFailureDomain]
		if !ok {
			// If not try backwards compat label
			availabilityDomainName, ok = pvc.Spec.Selector.MatchLabels["oci-availability-domain"]
		}
	}

	if !ok {
		nodes, err := p.nodeLister.List(labels.Everything())
		if err != nil {
			return "", nil, fmt.Errorf("failed to list nodes when choosing availability domain: %v", err)
		}
		validADs := sets.NewString()
		for _, node := range nodes {
			zone, ok := node.Labels[metav1.LabelZoneFailureDomain]
			if ok {
				validADs.Insert(zone)
			}
		}
		if validADs.Len() == 0 {
			return "", nil, fmt.Errorf("failed to choose availability domain; no zone labels (%q) on nodes", metav1.LabelZoneFailureDomain)
		}
		availabilityDomainName = util.ChooseZoneForVolume(validADs, pvc.Name)
		p.logger.With("availabilityDomain", availabilityDomainName).Info("No availability domain provided. Selecting one automatically.")
	}

	availabilityDomain, err := p.client.Identity().GetAvailabilityDomainByName(ctx, p.compartmentID, availabilityDomainName)
	if err != nil {
		return "", nil, err
	}

	return availabilityDomainName, availabilityDomain, nil
}
