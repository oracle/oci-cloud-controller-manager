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

package plugin

import (
	"github.com/oracle/oci-go-sdk/v50/identity"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v8/controller"
)

const (
	// OCIProvisionerName is the name of the provisioner defined in the storage class definitions
	OCIProvisionerName = "oracle/oci"
	// LabelZoneFailureDomain the availability domain in which the PD resides.
	LabelZoneFailureDomain = "failure-domain.beta.kubernetes.io/zone"
	// LabelZoneRegion the region in which the PD resides.
	LabelZoneRegion = "failure-domain.beta.kubernetes.io/region"
)

// ProvisionerPlugin implements the controller plugin plus some extras that are common
type ProvisionerPlugin interface {
	// Provision creates a volume i.e. the storage asset and returns a PV object
	// for the volume
	Provision(controller.ProvisionOptions, *identity.AvailabilityDomain) (*v1.PersistentVolume, error)
	// Delete removes the storage asset that was created by Provision backing the
	// given PV. Does not delete the PV object itself.
	//
	// May return IgnoredError to indicate that the call has been ignored and no
	// action taken.
	Delete(*v1.PersistentVolume) error
}
