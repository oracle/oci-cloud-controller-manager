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
	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"
	"reflect"
	"testing"
)

func TestMapAvailabilityDomainToFailureDomain(t *testing.T) {
	var testCases = map[string]string{
		"NWuj:PHX-AD-1": "PHX-AD-1",
		"NWuj:PHX-AD-2": "PHX-AD-2",
		"NWuj:PHX-AD-3": "PHX-AD-3",
		"":              "",
		"PHX-AD-3":      "PHX-AD-3",
	}
	for ad, fd := range testCases {
		t.Run(ad, func(t *testing.T) {
			v := mapAvailabilityDomainToFailureDomain(ad)
			if v != fd {
				t.Errorf("mapAvailabilityDomainToFailureDomain(%q) => %q, want %q", ad, v, fd)
			}
		})
	}
}

func TestGetZoneByProviderID(t *testing.T) {
	testCases := []struct {
		name string
		in   string
		out  cloudprovider.Zone
		err  error
	}{
		{
			name: "provider id without provider prefix",
			in:   "instance_zone_test",
			out: cloudprovider.Zone{
				FailureDomain: "PHX-AD-1",
				Region:        "PHX",
			},
			err: nil,
		},
		{
			name: "provider id with provider prefix",
			in:   providerPrefix + "instance_zone_test",
			out: cloudprovider.Zone{
				FailureDomain: "PHX-AD-1",
				Region:        "PHX",
			},
			err: nil,
		},
		{
			name: "provider id with provider prefix and instance not in cache",
			in:   providerPrefix + "instance_zone_test_noncache",
			out: cloudprovider.Zone{
				FailureDomain: "PHX-AD-1",
				Region:        "PHX",
			},
			err: nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.GetZoneByProviderID(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("GetZoneByProviderID(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("GetZoneByProviderID(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}

func TestGetZoneByNodeName(t *testing.T) {
	testCases := []struct {
		name string
		in   types.NodeName
		out  cloudprovider.Zone
		err  error
	}{
		{
			name: "get zone by node name",
			in:   "default",
			out: cloudprovider.Zone{
				FailureDomain: "PHX-AD-1",
				Region:        "PHX",
			},
			err: nil,
		},
	}

	cp := &CloudProvider{
		NodeLister:    &mockNodeLister{},
		client:        MockOCIClient{},
		config:        &providercfg.Config{CompartmentID: "testCompartment"},
		logger:        zap.S(),
		instanceCache: &mockInstanceCache{},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cp.GetZoneByNodeName(context.Background(), tt.in)
			if err != nil && err.Error() != tt.err.Error() {
				t.Errorf("GetZoneByNodeName(context, %+v) got error %v, expected %v", tt.in, err, tt.err)
			}
			if !reflect.DeepEqual(result, tt.out) {
				t.Errorf("GetZoneByNodeName(context, %+v) => %+v, want %+v", tt.in, result, tt.out)
			}
		})
	}
}
