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

package oci

import (
	"reflect"
	"testing"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
)

var (
	instanceCompID = "instanceCompID"
	instanceFD     = "instanceFD"
	instanceID     = "ocid1.instanceID"
)

func TestGetNodePatchBytes(t *testing.T) {
	testCases := map[string]struct {
		node               *v1.Node
		instance           *core.Instance
		expectedPatchBytes []byte
	}{
		"FD label and CompartmentID annotation already present": {
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						CompartmentIDAnnotation: "compID",
					},
					Labels: map[string]string{
						FaultDomainLabel: "FD",
					},
				},
			},
			instance: &core.Instance{
				CompartmentId: &instanceCompID,
				FaultDomain:   &instanceFD,
			},
			expectedPatchBytes: nil,
		},
		"Only FD label present": {
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						FaultDomainLabel: "FD",
					},
				},
			},
			instance: &core.Instance{
				CompartmentId: &instanceCompID,
				FaultDomain:   &instanceFD,
			},
			expectedPatchBytes: []byte("{\"metadata\": {\"annotations\": {\"oci.oraclecloud.com/compartment-id\":\"instanceCompID\"}}}"),
		},
		"Only CompartmentID annotation present": {
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						CompartmentIDAnnotation: "compID",
					},
				},
			},
			instance: &core.Instance{
				CompartmentId: &instanceCompID,
				FaultDomain:   &instanceFD,
			},
			expectedPatchBytes: []byte("{\"metadata\": {\"labels\": {\"oci.oraclecloud.com/fault-domain\":\"instanceFD\"}}}"),
		},
		"none present": {
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{},
			},
			instance: &core.Instance{
				CompartmentId: &instanceCompID,
				FaultDomain:   &instanceFD,
			},
			expectedPatchBytes: []byte("{\"metadata\": {\"labels\": {\"oci.oraclecloud.com/fault-domain\":\"instanceFD\"},\"annotations\": {\"oci.oraclecloud.com/compartment-id\":\"instanceCompID\"}}}"),
		},
	}
	logger := zap.L()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			patchedBytes := getNodePatchBytes(tc.node, tc.instance, logger.Sugar())
			if !reflect.DeepEqual(patchedBytes, tc.expectedPatchBytes) {
				t.Errorf("Expected PatchBytes \n%+v\nbut got\n%+v", tc.expectedPatchBytes, patchedBytes)
			}
		})
	}
}

func TestGetInstanceByNode(t *testing.T) {
	testCases := map[string]struct {
		node             *v1.Node
		nic              *NodeInfoController
		expectedInstance *core.Instance
	}{
		"Get Instance": {
			node: &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: instanceID,
				},
			},
			nic: &NodeInfoController{
				ociClient: MockOCIClient{},
			},
			expectedInstance: &core.Instance{
				AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
				CompartmentId:      common.String("default"),
				Id:                 &instanceID,
				Region:             common.String("PHX"),
				Shape:              common.String("VM.Standard1.2"),
			},
		},
		"Get Instance when providerID is prefixed with providerName": {
			node: &v1.Node{
				Spec: v1.NodeSpec{
					ProviderID: providerPrefix + instanceID,
				},
			},
			nic: &NodeInfoController{
				ociClient: MockOCIClient{},
			},
			expectedInstance: &core.Instance{
				AvailabilityDomain: common.String("NWuj:PHX-AD-1"),
				CompartmentId:      common.String("default"),
				Id:                 &instanceID,
				Region:             common.String("PHX"),
				Shape:              common.String("VM.Standard1.2"),
			},
		},
	}

	logger := zap.L()
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			instance, err := getInstanceByNode(tc.node, tc.nic, logger.Sugar())
			if err != nil {
				t.Fatalf("%s unexpected service add error: %v", name, err)
			}
			if !reflect.DeepEqual(instance, tc.expectedInstance) {
				t.Errorf("Expected instance \n%+v\nbut got\n%+v", tc.expectedInstance, instanceID)
			}
		})
	}
}
