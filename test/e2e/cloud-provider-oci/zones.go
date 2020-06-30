// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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

package e2e

import (
	"context"

	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/cloud-provider"
)

var _ = Describe("Zones", func() {
	f := sharedfw.NewFrameworkWithCloudProvider("zones")
	var (
		cs    clientset.Interface
		zones cloudprovider.Zones
		node  v1.Node
	)
	BeforeEach(func() {
		var enabled bool
		zones, enabled = f.CloudProvider.Zones()
		Expect(enabled).To(BeTrue())

		cs = f.ClientSet
		nodes := sharedfw.GetReadySchedulableNodesOrDie(cs)
		Expect(len(nodes.Items)).NotTo(BeZero())
		node = nodes.Items[0]
	})

	Context("[cloudprovider][ccm]", func() {
		It("should be possible to get a non-empty zone by provider ID", func() {
			providerID := node.Spec.ProviderID
			Expect(providerID).NotTo(BeEmpty())

			By("calling GetZoneByProviderID()")
			zone, err := zones.GetZoneByProviderID(context.Background(), providerID)
			Expect(err).NotTo(HaveOccurred())
			Expect(zone.Region).NotTo(BeEmpty())
			sharedfw.Logf("%q: Region=%q", providerID, zone.Region)
			Expect(zone.FailureDomain).NotTo(BeEmpty())
			sharedfw.Logf("%q: FailureDomain=%q", providerID, zone.FailureDomain)
		})

		It("should be possible to get a non-empty zone by node name", func() {
			nodeName := apitypes.NodeName(node.Name)
			Expect(nodeName).NotTo(BeEmpty())

			By("calling GetZoneByNodeName()")
			zone, err := zones.GetZoneByNodeName(context.Background(), nodeName)
			Expect(err).NotTo(HaveOccurred())
			Expect(zone.Region).NotTo(BeEmpty())
			sharedfw.Logf("%q: Region=%q", nodeName, zone.Region)
			Expect(zone.FailureDomain).NotTo(BeEmpty())
			sharedfw.Logf("%q: FailureDomain=%q", nodeName, zone.FailureDomain)
		})
	})
})
