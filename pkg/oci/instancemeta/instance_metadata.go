// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.
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

package instancemeta

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	baseURL          = "http://169.254.169.254"
	metadataEndpoint = "/opc/v1/instance/"
)

// InstanceMetadata holds the subset of the instance metadata retrieved from the
// local OCI instance metadata API endpoint.
// https://docs.us-phoenix-1.oraclecloud.com/Content/Compute/Tasks/gettingmetadata.htm
type InstanceMetadata struct {
	CompartmentOCID string `json:"compartmentId"`
	Region          string `json:"region"`
}

// Interface defines how consumers access OCI instance metadata.
type Interface interface {
	Get() (*InstanceMetadata, error)
}

type metadataGetter struct {
	baseURL string
	client  *http.Client
}

// New returns the instance metadata for the host on which the code is being
// executed.
func New() Interface {
	return &metadataGetter{client: http.DefaultClient, baseURL: baseURL}
}

// Get either returns the cached metadata for the current instance or queries
// the instance metadata API, populates the cache, and returns the result.
func (m *metadataGetter) Get() (*InstanceMetadata, error) {
	req, err := http.NewRequest("GET", m.baseURL+metadataEndpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance metadata: %v", err)

	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metadata endpoint returned status %d; expected 200 OK", resp.StatusCode)
	}

	md := &InstanceMetadata{}
	err = json.NewDecoder(resp.Body).Decode(md)
	if err != nil {
		return nil, err
	}

	return md, nil
}
