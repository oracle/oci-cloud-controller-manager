/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	. "github.com/onsi/ginkgo"

	"k8s.io/api/core/v1"

	cloudprovider "github.com/oracle/oci-cloud-controller-manager/pkg/oci"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var _ = Describe("Service type:LoadBalancer", func() {
	f := framework.NewDefaultFramework("service")

	It("should be possible to create and mutate a Service type:LoadBalancer", func() {
		serviceName := "basic-lb-test"
		ns := f.Namespace.Name

		jig := framework.NewServiceTestJig(f.ClientSet, serviceName)
		nodeIP := framework.PickNodeIP(jig.Client) // for later

		loadBalancerLagTimeout := framework.LoadBalancerLagTimeoutDefault
		loadBalancerCreateTimeout := framework.LoadBalancerCreateTimeoutDefault
		if nodes := framework.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > framework.LargeClusterMinNodesNumber {
			loadBalancerCreateTimeout = framework.LoadBalancerCreateTimeoutLarge
		}

		// TODO(apryde): Test that LoadBalancers can receive static IP addresses
		// (in a provider agnostic manner?). OCI does not currently
		// support this.
		requestedIP := ""

		tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
			s.Spec.Type = v1.ServiceTypeLoadBalancer
			s.Spec.LoadBalancerIP = requestedIP // will be "" if not applicable
			// TODO(apryde): Remove and use default shape (resources permitting)
			s.Annotations = map[string]string{cloudprovider.ServiceAnnotationLoadBalancerShape: "400Mbps"}
		})

		svcPort := int(tcpService.Spec.Ports[0].Port)

		By("creating a pod to be part of the TCP service " + serviceName)
		jig.RunOrFail(ns, nil)

		// TODO(apryde): Test UDP service. OCI does not currently support this.

		By("waiting for the TCP service to have a load balancer")
		// Wait for the load balancer to be created asynchronously
		tcpService = jig.WaitForLoadBalancerOrFail(ns, tcpService.Name, loadBalancerCreateTimeout)
		jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)

		tcpNodePort := int(tcpService.Spec.Ports[0].NodePort)
		framework.Logf("TCP node port: %d", tcpNodePort)

		if requestedIP != "" && framework.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != requestedIP {
			framework.Failf("unexpected TCP Status.LoadBalancer.Ingress (expected %s, got %s)", requestedIP, framework.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
		}
		tcpIngressIP := framework.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
		framework.Logf("TCP load balancer: %s", tcpIngressIP)

		By("hitting the TCP service's NodePort")
		jig.TestReachableHTTP(nodeIP, tcpNodePort, framework.KubeProxyLagTimeout)

		By("hitting the TCP service's LoadBalancer")
		jig.TestReachableHTTP(tcpIngressIP, svcPort, loadBalancerLagTimeout)

		// Change the services' node ports.

		By("changing the TCP service's NodePort")
		tcpService = jig.ChangeServiceNodePortOrFail(ns, tcpService.Name, tcpNodePort)
		jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)
		tcpNodePortOld := tcpNodePort
		tcpNodePort = int(tcpService.Spec.Ports[0].NodePort)
		if tcpNodePort == tcpNodePortOld {
			framework.Failf("TCP Spec.Ports[0].NodePort (%d) did not change", tcpNodePort)
		}
		if framework.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != tcpIngressIP {
			framework.Failf("TCP Status.LoadBalancer.Ingress changed (%s -> %s) when not expected", tcpIngressIP, framework.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
		}
		framework.Logf("TCP node port: %d", tcpNodePort)

		By("hitting the TCP service's new NodePort")
		jig.TestReachableHTTP(nodeIP, tcpNodePort, framework.KubeProxyLagTimeout)

		By("checking the old TCP NodePort is closed")
		jig.TestNotReachableHTTP(nodeIP, tcpNodePortOld, framework.KubeProxyLagTimeout)

		By("hitting the TCP service's LoadBalancer")
		jig.TestReachableHTTP(tcpIngressIP, svcPort, loadBalancerLagTimeout)

		// Change the services' main ports.

		By("changing the TCP service's port")
		tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
			s.Spec.Ports[0].Port++
		})
		jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)
		svcPortOld := svcPort
		svcPort = int(tcpService.Spec.Ports[0].Port)
		if svcPort == svcPortOld {
			framework.Failf("TCP Spec.Ports[0].Port (%d) did not change", svcPort)
		}
		if int(tcpService.Spec.Ports[0].NodePort) != tcpNodePort {
			framework.Failf("TCP Spec.Ports[0].NodePort (%d) changed", tcpService.Spec.Ports[0].NodePort)
		}
		if framework.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != tcpIngressIP {
			framework.Failf("TCP Status.LoadBalancer.Ingress changed (%s -> %s) when not expected", tcpIngressIP, framework.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
		}

		framework.Logf("service port (TCP): %d", svcPort)

		By("hitting the TCP service's NodePort")
		jig.TestReachableHTTP(nodeIP, tcpNodePort, framework.KubeProxyLagTimeout)

		By("hitting the TCP service's LoadBalancer")
		jig.TestReachableHTTP(tcpIngressIP, svcPort, loadBalancerCreateTimeout) // this may actually recreate the LB

		// Change the services back to ClusterIP.

		By("changing TCP service back to type=ClusterIP")
		tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
			s.Spec.Type = v1.ServiceTypeClusterIP
			s.Spec.Ports[0].NodePort = 0
		})
		// Wait for the load balancer to be destroyed asynchronously
		tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
		jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)

		By("checking the TCP NodePort is closed")
		jig.TestNotReachableHTTP(nodeIP, tcpNodePort, framework.KubeProxyLagTimeout)

		By("checking the TCP LoadBalancer is closed")
		jig.TestNotReachableHTTP(tcpIngressIP, svcPort, loadBalancerLagTimeout)
	})
})
