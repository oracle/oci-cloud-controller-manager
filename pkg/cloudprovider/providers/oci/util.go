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
	"strings"

	"github.com/pkg/errors"
	api "k8s.io/api/core/v1"
)

// MapProviderIDToInstanceID parses the provider id and returns the instance ocid.
func MapProviderIDToInstanceID(providerID string) (string, error) {
	if providerID == "" {
		return providerID, errors.New("provider ID is empty")
	}
	if strings.HasPrefix(providerID, providerPrefix) {
		return strings.TrimPrefix(providerID, providerPrefix), nil
	}
	return providerID, nil
}

// NodeInternalIP returns the nodes internal ip
// A node managed by the CCM will always have an internal ip
// since it's not possible to deploy an instance without a private ip.
func NodeInternalIP(node *api.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == api.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}
