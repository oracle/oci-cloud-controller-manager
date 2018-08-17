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
	"fmt"
	"net"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"

	cloudprovider "github.com/oracle/oci-cloud-controller-manager/pkg/oci"
	"github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
)

var _ = Describe("Service [Slow]", func() {
	f := framework.NewDefaultFramework("service")

	It("should be possible to create and mutate a Service type:LoadBalancer [Canary]", func() {
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

// NOTE: OCI LBaaS is not a passthrough load balancer so ESIPP (External Source IP
// Presevation) is not possible, however, this test covers support for node-local
// routing (i.e. avoidance of a second hop).
var _ = Describe("ESIPP [Slow]", func() {
	f := framework.NewDefaultFramework("esipp")

	loadBalancerCreateTimeout := framework.LoadBalancerCreateTimeoutDefault
	serviceLBNames := []string{}

	var cs clientset.Interface
	BeforeEach(func() {
		cs = f.ClientSet
	})

	It("should only target nodes with endpoints", func() {
		namespace := f.Namespace.Name
		serviceName := "external-local"
		jig := framework.NewServiceTestJig(cs, serviceName)
		nodes := jig.GetNodes(framework.MaxNodesForEndpointsTests)

		svc := jig.CreateOnlyLocalLoadBalancerService(namespace, serviceName, loadBalancerCreateTimeout, false,
			func(svc *v1.Service) {
				// Change service port to avoid collision with opened hostPorts
				// in other tests that run in parallel.
				if len(svc.Spec.Ports) != 0 {
					svc.Spec.Ports[0].TargetPort = intstr.FromInt(int(svc.Spec.Ports[0].Port))
					svc.Spec.Ports[0].Port = 8081
				}

			})
		serviceLBNames = append(serviceLBNames, cloudprovider.GetLoadBalancerName(svc))
		defer func() {
			jig.ChangeServiceType(svc.Namespace, svc.Name, v1.ServiceTypeClusterIP, loadBalancerCreateTimeout)
			Expect(cs.CoreV1().Services(svc.Namespace).Delete(svc.Name, nil)).NotTo(HaveOccurred())
		}()

		healthCheckNodePort := int(svc.Spec.HealthCheckNodePort)
		if healthCheckNodePort == 0 {
			framework.Failf("Service HealthCheck NodePort was not allocated")
		}

		ips := framework.CollectAddresses(nodes, v1.NodeExternalIP)

		ingressIP := framework.GetIngressPoint(&svc.Status.LoadBalancer.Ingress[0])
		svcTCPPort := int(svc.Spec.Ports[0].Port)

		threshold := 2
		path := "/healthz"
		for i := 0; i < len(nodes.Items); i++ {
			endpointNodeName := nodes.Items[i].Name

			By("creating a pod to be part of the service " + serviceName + " on node " + endpointNodeName)
			jig.RunOrFail(namespace, func(rc *v1.ReplicationController) {
				rc.Name = serviceName
				if endpointNodeName != "" {
					rc.Spec.Template.Spec.NodeName = endpointNodeName
				}
			})

			By(fmt.Sprintf("waiting for service endpoint on node %v", endpointNodeName))
			jig.WaitForEndpointOnNode(namespace, serviceName, endpointNodeName)

			// HealthCheck should pass only on the node where num(endpoints) > 0
			// All other nodes should fail the healthcheck on the service healthCheckNodePort
			for n, publicIP := range ips {
				// Make sure the loadbalancer picked up the health check change.
				// Confirm traffic can reach backend through LB before checking healthcheck nodeport.
				jig.TestReachableHTTP(ingressIP, svcTCPPort, framework.KubeProxyLagTimeout)
				expectedSuccess := nodes.Items[n].Name == endpointNodeName
				port := strconv.Itoa(healthCheckNodePort)
				ipPort := net.JoinHostPort(publicIP, port)
				framework.Logf("Health checking %s, http://%s%s, expectedSuccess %v", nodes.Items[n].Name, ipPort, path, expectedSuccess)
				Expect(jig.TestHTTPHealthCheckNodePort(publicIP, healthCheckNodePort, path, framework.KubeProxyEndpointLagTimeout, expectedSuccess, threshold)).NotTo(HaveOccurred())
			}
			framework.ExpectNoError(framework.DeleteRCAndWaitForGC(f.ClientSet, namespace, serviceName))
		}
	})

	It("should work from pods", func() {
		namespace := f.Namespace.Name
		serviceName := "external-local"
		jig := framework.NewServiceTestJig(cs, serviceName)
		nodes := jig.GetNodes(framework.MaxNodesForEndpointsTests)

		svc := jig.CreateOnlyLocalLoadBalancerService(namespace, serviceName, loadBalancerCreateTimeout, true, nil)
		serviceLBNames = append(serviceLBNames, cloudprovider.GetLoadBalancerName(svc))
		defer func() {
			jig.ChangeServiceType(svc.Namespace, svc.Name, v1.ServiceTypeClusterIP, loadBalancerCreateTimeout)
			Expect(cs.CoreV1().Services(svc.Namespace).Delete(svc.Name, nil)).NotTo(HaveOccurred())
		}()

		ingressIP := framework.GetIngressPoint(&svc.Status.LoadBalancer.Ingress[0])
		port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
		ipPort := net.JoinHostPort(ingressIP, port)
		path := fmt.Sprintf("%s/clientip", ipPort)
		nodeName := nodes.Items[0].Name
		podName := "execpod-sourceip"

		By(fmt.Sprintf("Creating %v on node %v", podName, nodeName))
		execPodName := framework.CreateExecPodOrFail(f.ClientSet, namespace, podName, func(pod *v1.Pod) {
			pod.Spec.NodeName = nodeName
		})
		defer func() {
			err := cs.CoreV1().Pods(namespace).Delete(execPodName, nil)
			Expect(err).NotTo(HaveOccurred())
		}()
		execPod, err := f.ClientSet.CoreV1().Pods(namespace).Get(execPodName, metav1.GetOptions{})
		framework.ExpectNoError(err)

		framework.Logf("Waiting up to %v wget %v", framework.KubeProxyLagTimeout, path)
		cmd := fmt.Sprintf(`wget -T 30 -qO- %v`, path)

		var srcIP string
		By(fmt.Sprintf("Hitting external lb %v from pod %v on node %v", ingressIP, podName, nodeName))
		if pollErr := wait.PollImmediate(framework.Poll, framework.LoadBalancerCreateTimeoutDefault, func() (bool, error) {
			stdout, err := framework.RunHostCmd(execPod.Namespace, execPod.Name, cmd)
			if err != nil {
				framework.Logf("got err: %v, retry until timeout", err)
				return false, nil
			}
			srcIP = strings.TrimSpace(strings.Split(stdout, ":")[0])
			return srcIP == execPod.Status.PodIP, nil
		}); pollErr != nil {
			framework.Failf("Source IP not preserved from %v, expected '%v' got '%v'", podName, execPod.Status.PodIP, srcIP)
		}
	})
})
