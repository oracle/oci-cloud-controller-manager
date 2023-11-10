// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
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

package framework

import (
	cloudprovider "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
)

const (
	okeSystemTagKey = "Cluster"
)

func HasOkeSystemTags(systemTags map[string]map[string]interface{}) bool {
	Logf("actual system tags on the resource: %v", systemTags)
	if systemTags != nil {
		if okeSystemTag, okeSystemTagNsExists := systemTags[cloudprovider.OkeSystemTagNamesapce]; okeSystemTagNsExists {
			if _, okeSystemTagKeyExists := okeSystemTag[okeSystemTagKey]; okeSystemTagKeyExists {
				return true
			}
		}
		return false
	}
	return false
}
