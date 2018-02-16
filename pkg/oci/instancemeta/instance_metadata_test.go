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

package instancemeta

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const exampleResponse = `{
  "availabilityDomain" : "NWuj:PHX-AD-1",
  "compartmentId" : "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
  "displayName" : "trjl-kb8s-master",
  "id" : "ocid1.instance.oc1.phx.abyhqljtj775udgtbu7nddt6j2hqgxdsgrnpweepogvvsmqfppefewile5zq",
  "image" : "ocid1.image.oc1.phx.aaaaaaaaamx6ta37uxltor6n5lxfgd5lkb3lwmoqurlpn2x4dz5ockekiuea",
  "metadata" : {
    "ssh_authorized_keys" : "ssh-rsa some-key-data tlangfor@tlangfor-mac\n"
  },
  "region" : "phx",
  "shape" : "VM.Standard1.1",
  "state" : "Provisioning",
  "timeCreated" : 1496415602152
}`

func TestGetMetadata(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, exampleResponse)
	}))
	defer ts.Close()
	getter := metadataGetter{client: ts.Client(), baseURL: ts.URL}
	meta, err := getter.Get()
	if err != nil {
		t.Fatalf("Uexpected error calling Get(): %v", err)
	}

	expected := &InstanceMetadata{
		CompartmentOCID: "ocid1.compartment.oc1..aaaaaaaa3um2atybwhder4qttfhgon4j3hcxgmsvnyvx4flfjyewkkwfzwnq",
		Region:          "phx",
	}
	if !reflect.DeepEqual(meta, expected) {
		t.Errorf("Get() => %+v, want %+v", meta, expected)
	}
}
