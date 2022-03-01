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
	"net"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	cloudprovider "k8s.io/cloud-provider"
)

func assertNodeAddressesContainValidIPs(addrs []v1.NodeAddress) {
	for _, addr := range addrs {
		if addr.Type == v1.NodeExternalIP || addr.Type == v1.NodeInternalIP {
			ip := net.ParseIP(addr.Address)
			ExpectWithOffset(1, ip).NotTo(BeNil())
		}
	}
}

var _ = Describe("Instances", func() {
	f := sharedfw.NewFrameworkWithCloudProvider("instances")
	var (
		cs        clientset.Interface
		instances cloudprovider.Instances
		node      v1.Node
	)
	BeforeEach(func() {
		var enabled bool
		instances, enabled = f.CloudProvider.Instances()
		Expect(enabled).To(BeTrue())

		cs = f.ClientSet
		nodes := sharedfw.GetReadySchedulableNodesOrDie(cs)
		Expect(len(nodes.Items)).NotTo(BeZero())
		node = nodes.Items[0]
	})

	Context("[cloudprovider][ccm]", func() {
		It("should be possible to get node addresses", func() {
			nodeName := apitypes.NodeName(node.Name)
			Expect(nodeName).NotTo(BeEmpty())

			By("calling NodeAddresses()")
			addresses, err := instances.NodeAddresses(context.Background(), nodeName)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(addresses)).NotTo(BeZero())
			sharedfw.Logf("%q: addresses=%v", nodeName, addresses)
			assertNodeAddressesContainValidIPs(addresses)

			providerID := node.Spec.ProviderID
			Expect(providerID).NotTo(BeEmpty())

			By("calling NodeAddressesByProviderID()")
			addresses2, err := instances.NodeAddressesByProviderID(context.Background(), providerID)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(addresses2)).NotTo(BeZero())
			sharedfw.Logf("%q (%s): addresses=%v", node.Name, providerID, addresses2)
			assertNodeAddressesContainValidIPs(addresses2)

			Expect(addresses).To(ConsistOf(addresses2))
		})

		It("should be possible to get the provider ID of an instance", func() {
			nodeName := apitypes.NodeName(node.Name)
			Expect(nodeName).NotTo(BeEmpty())

			By("calling InstanceID()")
			providerID, err := instances.InstanceID(context.Background(), nodeName)
			Expect(err).NotTo(HaveOccurred())
			Expect(providerID).NotTo(BeEmpty())
			sharedfw.Logf("%q: providerID=%q", nodeName, providerID)
		})

		It("should be possible to get the type of an instance", func() {
			nodeName := apitypes.NodeName(node.Name)
			Expect(nodeName).NotTo(BeEmpty())

			By("calling InstanceType()")
			instanceType, err := instances.InstanceType(context.Background(), nodeName)
			Expect(err).NotTo(HaveOccurred())
			Expect(instanceType).NotTo(BeEmpty())
			sharedfw.Logf("%q: instanceType=%q", nodeName, instanceType)

			providerID := node.Spec.ProviderID
			Expect(providerID).NotTo(BeEmpty())

			By("calling InstanceTypeByProviderID()")
			instanceType2, err := instances.InstanceTypeByProviderID(context.Background(), providerID)
			Expect(err).NotTo(HaveOccurred())
			Expect(instanceType2).NotTo(BeEmpty())
			sharedfw.Logf("%q (%s): instanceType=%q", node.Name, providerID, instanceType2)

			Expect(instanceType).To(Equal(instanceType2))
		})

		It("should be possible to check an instance exists", func() {
			providerID := node.Spec.ProviderID
			Expect(providerID).NotTo(BeEmpty())
			By("calling InstanceExistsByProviderID()")
			exists, err := instances.InstanceExistsByProviderID(context.Background(), providerID)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
		It("should be possible to check required annotations and labels are added to node", func() {
			// OCI Labels and Annotations
			compartmentID := node.ObjectMeta.Annotations[oci.CompartmentIDAnnotation]
			Expect(compartmentID).NotTo(BeEmpty())
			faultDomain := node.ObjectMeta.Labels[oci.FaultDomainLabel]
			Expect(faultDomain).NotTo(BeEmpty())

			// Kubernetes Beta Labels
			fdZone := node.ObjectMeta.Labels[v1.LabelZoneFailureDomain]
			Expect(fdZone).NotTo(BeEmpty())
			region := node.ObjectMeta.Labels[v1.LabelZoneRegion]
			Expect(region).NotTo(BeEmpty())
			instanceType := node.ObjectMeta.Labels[v1.LabelInstanceType]
			Expect(instanceType).NotTo(BeEmpty())

			// Kubernetes Stable Labels
			fdZone = node.ObjectMeta.Labels[v1.LabelZoneFailureDomainStable]
			Expect(fdZone).NotTo(BeEmpty())
			region = node.ObjectMeta.Labels[v1.LabelZoneRegionStable]
			Expect(region).NotTo(BeEmpty())
			instanceType = node.ObjectMeta.Labels[v1.LabelInstanceTypeStable]
			Expect(instanceType).NotTo(BeEmpty())

		})
	})
})
