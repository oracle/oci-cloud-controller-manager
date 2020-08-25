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

package metadata

import (
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	baseURL            = "http://169.254.169.254"
	metadataEndpoint   = "/opc/v2/instance/"
	defaultHTTPTimeout = 5 * time.Second
)

// InstanceMetadata holds the subset of the instance metadata retrieved from the
// local OCI instance metadata API endpoint.
// https://docs.us-phoenix-1.oraclecloud.com/Content/Compute/Tasks/gettingmetadata.htm
type InstanceMetadata struct {
	CompartmentID       string `json:"compartmentId"`
	Region              string `json:"region"`
	CanonicalRegionName string `json:"canonicalRegionName"`
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
	var client = &http.Client{Timeout: defaultHTTPTimeout}
	return &metadataGetter{client: client, baseURL: baseURL}
}

// Get either returns the cached metadata for the current instance or queries
// the instance metadata API, populates the cache, and returns the result.
func (m *metadataGetter) Get() (*InstanceMetadata, error) {
	req, err := http.NewRequest("GET", m.baseURL+metadataEndpoint, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return m.executeRequest(req)
}

func (m *metadataGetter) executeRequest(req *http.Request) (*InstanceMetadata, error) {
	req.Header.Add("Authorization", "Bearer Oracle")
	resp, err := m.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		zap.S().With(zap.Error(err)).Warn("Failed to get instance metadata with endpoint v2. Falling back to v1.")
		if resp != nil {
			v2resp := resp
			defer v2resp.Body.Close()
		}
		v1Req := *req
		v1Path := strings.Replace(req.URL.Path, "/opc/v2", "/opc/v1", 1)
		v1Req.URL.Path = v1Path
		resp, err = m.client.Do(&v1Req)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to get instance metadata with v1 endpoint after falling back from v2 endpoint")
		}
	}

	zap.S().Infof("Metadata endpoint %s returned response successfully", req.URL.Path)

	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, errors.Errorf("metadata endpoint v1 returned status %d; expected 200 OK", resp.StatusCode)
		}
	}
	md := &InstanceMetadata{}
	err = json.NewDecoder(resp.Body).Decode(md)
	if err != nil {
		return nil, errors.Wrap(err, "decoding instance metadata response")
	}

	return md, nil
}
