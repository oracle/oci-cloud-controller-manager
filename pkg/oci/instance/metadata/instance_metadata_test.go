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
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

const exampleResponse = `{
  "availabilityDomain" : "NWuj:PHX-AD-1",
  "compartmentId" : "ocid1.compartment.oc1..abc",
  "displayName" : "trjl-kb8s-master",
  "id" : "ocid1.instance.oc1.phx.xyz",
  "image" : "ocid1.image.oc1.phx.pqr",
  "metadata" : {
    "ssh_authorized_keys" : "ssh-rsa some-key-data tlangfor@tlangfor-mac\n"
  },
  "region" : "phx",
  "canonicalRegionName" : "us-phoenix-1",
  "shape" : "VM.Standard1.1",
  "state" : "Provisioning",
  "timeCreated" : 1496415602152
}`

func TestGetMetadata(t *testing.T) {

	type Result struct {
		metadata *InstanceMetadata
		err      string
	}
	tests := []struct {
		name        string
		endpoint    string
		expected    Result
		handlerFunc http.HandlerFunc
	}{
		{
			name:     "metadata v1 response returned successfully",
			endpoint: "opc/v1/instance",
			expected: Result{
				metadata: &InstanceMetadata{
					CompartmentID:       "ocid1.compartment.oc1..abc",
					Region:              "phx",
					CanonicalRegionName: "us-phoenix-1",
				},
				err: "",
			},
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, exampleResponse)
			}),
		},
		{
			name:     "metadata v2 response returned successfully",
			endpoint: "opc/v2/instance",
			expected: Result{
				metadata: &InstanceMetadata{
					CompartmentID:       "ocid1.compartment.oc1..abc",
					Region:              "phx",
					CanonicalRegionName: "us-phoenix-1",
				},
				err: "",
			},
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, exampleResponse)
			}),
		},
		{
			name:     "metadata v1 and v2 response returned error",
			endpoint: "opc/v2/instance",
			expected: Result{
				metadata: nil,
				err:      fmt.Sprintf("metadata endpoint v1 returned status %d; expected 200 OK", 404),
			},
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "opc/v2") {
					w.WriteHeader(404)
				} else if strings.Contains(r.URL.Path, "opc/v1") {
					w.WriteHeader(404)
				}
			}),
		},
		{
			name:     "metadata v2 response returned error but v1 success",
			endpoint: "opc/v2/instance",
			expected: Result{
				metadata: &InstanceMetadata{
					CompartmentID:       "ocid1.compartment.oc1..abc",
					Region:              "phx",
					CanonicalRegionName: "us-phoenix-1",
				},
				err: "",
			},
			handlerFunc: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "opc/v2") {
					w.WriteHeader(404)
				} else if strings.Contains(r.URL.Path, "opc/v1") {
					fmt.Fprint(w, exampleResponse)
				}
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(tt.handlerFunc)
			getter := metadataGetter{client: ts.Client(), baseURL: ts.URL}
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", getter.baseURL, tt.endpoint), nil)
			if err != nil {
				t.Error(err)
			}
			meta, err := getter.executeRequest(req)

			if tt.expected.err != "" {
				if !reflect.DeepEqual(err.Error(), tt.expected.err) {
					t.Errorf("Get() => %+v, want %+v", err, tt.expected.err)
				}
			} else if !reflect.DeepEqual(meta, tt.expected.metadata) {
				t.Errorf("Get() => %+v, want %+v", meta, tt.expected.metadata)
			}
		})
	}
}
