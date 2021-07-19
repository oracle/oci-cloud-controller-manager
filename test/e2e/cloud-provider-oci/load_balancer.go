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
	"fmt"
	"net"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudprovider "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

var _ = Describe("Service [Slow]", func() {

	baseName := "service"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer (change nodeport) [Canary]", func() {
			serviceName := "basic-lb-test"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			nodeIP := sharedfw.PickNodeIP(jig.Client) // for later

			loadBalancerLagTimeout := sharedfw.LoadBalancerLagTimeoutDefault
			loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
			if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
				loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
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
			sharedfw.Logf("TCP node port: %d", tcpNodePort)

			if requestedIP != "" && sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != requestedIP {
				sharedfw.Failf("unexpected TCP Status.LoadBalancer.Ingress (expected %s, got %s)", requestedIP, sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
			}
			tcpIngressIP := sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
			sharedfw.Logf("TCP load balancer: %s", tcpIngressIP)

			if f.NodePortTest {
				By("hitting the TCP service's NodePort")
				jig.TestReachableHTTP(false, nodeIP, tcpNodePort, sharedfw.KubeProxyLagTimeout)
			}

			By("hitting the TCP service's LoadBalancer")
			jig.TestReachableHTTP(false, tcpIngressIP, svcPort, loadBalancerLagTimeout)

			// Change the services' node ports.

			By("changing the TCP service's NodePort")
			// Count the number of ingress/egress rules with the original port so
			// we can check the correct number are updated.
			numEgressRules, numIngressRules := sharedfw.CountSinglePortSecListRules(f.Client, f.CCMSecListID, f.K8SSecListID, tcpNodePort)
			tcpService = jig.ChangeServiceNodePortOrFail(ns, tcpService.Name, tcpNodePort)
			jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)

			tcpNodePortOld := tcpNodePort
			tcpNodePort = int(tcpService.Spec.Ports[0].NodePort)
			if tcpNodePort == tcpNodePortOld {
				sharedfw.Failf("TCP Spec.Ports[0].NodePort (%d) did not change", tcpNodePort)
			}
			if sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != tcpIngressIP {
				sharedfw.Failf("TCP Status.LoadBalancer.Ingress changed (%s -> %s) when not expected", tcpIngressIP, sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
			}

			// Check the correct number of rules are present.
			sharedfw.WaitForSinglePortEgressRulesAfterPortChangeOrFail(f.Client, f.CCMSecListID, numEgressRules, tcpNodePortOld, tcpNodePort)
			sharedfw.WaitForSinglePortIngressRulesAfterPortChangeOrFail(f.Client, f.K8SSecListID, numIngressRules, tcpNodePortOld, tcpNodePort)

			sharedfw.Logf("TCP node port: %d", tcpNodePort)

			if f.NodePortTest {
				By("hitting the TCP service's new NodePort")
				jig.TestReachableHTTP(false, nodeIP, tcpNodePort, sharedfw.KubeProxyLagTimeout)

				By("checking the old TCP NodePort is closed")
				jig.TestNotReachableHTTP(nodeIP, tcpNodePortOld, sharedfw.KubeProxyLagTimeout)
			}

			By("hitting the TCP service's LoadBalancer")
			jig.TestReachableHTTP(false, tcpIngressIP, svcPort, loadBalancerLagTimeout)

			// Change the services' main ports.

			By("changing the TCP service's port")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Ports[0].Port++
			})
			jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)
			svcPortOld := svcPort
			svcPort = int(tcpService.Spec.Ports[0].Port)
			if svcPort == svcPortOld {
				sharedfw.Failf("TCP Spec.Ports[0].Port (%d) did not change", svcPort)
			}
			if int(tcpService.Spec.Ports[0].NodePort) != tcpNodePort {
				sharedfw.Failf("TCP Spec.Ports[0].NodePort (%d) changed", tcpService.Spec.Ports[0].NodePort)
			}
			if sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != tcpIngressIP {
				sharedfw.Failf("TCP Status.LoadBalancer.Ingress changed (%s -> %s) when not expected", tcpIngressIP, sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
			}

			sharedfw.Logf("service port (TCP): %d", svcPort)
			if f.NodePortTest {
				By("hitting the TCP service's NodePort")
				jig.TestReachableHTTP(false, nodeIP, tcpNodePort, sharedfw.KubeProxyLagTimeout)
			}

			By("hitting the TCP service's LoadBalancer")
			jig.TestReachableHTTP(false, tcpIngressIP, svcPort, loadBalancerCreateTimeout) // this may actually recreate the LB

			// Change the services back to ClusterIP.

			By("changing TCP service back to type=ClusterIP")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.Ports[0].NodePort = 0
			})
			// Wait for the load balancer to be destroyed asynchronously
			tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)

			if f.NodePortTest {
				By("checking the TCP NodePort is closed")
				jig.TestNotReachableHTTP(nodeIP, tcpNodePort, sharedfw.KubeProxyLagTimeout)
			}

			By("checking the TCP LoadBalancer is closed")
			jig.TestNotReachableHTTP(tcpIngressIP, svcPort, loadBalancerLagTimeout)
		})
	})
})

// NOTE: OCI LBaaS is not a passthrough load balancer so ESIPP (External Source IP
// Presevation) is not possible, however, this test covers support for node-local
// routing (i.e. avoidance of a second hop).
var _ = Describe("ESIPP [Slow]", func() {

	baseName := "esipp"
	f := sharedfw.NewDefaultFramework(baseName)

	loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
	serviceLBNames := []string{}

	var cs clientset.Interface
	BeforeEach(func() {
		cs = f.ClientSet
	})
	Context("[cloudprovider][ccm]", func() {
		It("should only target nodes with endpoints", func() {
			namespace := f.Namespace.Name
			serviceName := "external-local"
			jig := sharedfw.NewServiceTestJig(cs, serviceName)
			nodes := jig.GetNodes(sharedfw.MaxNodesForEndpointsTests)

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
				sharedfw.Failf("Service HealthCheck NodePort was not allocated")
			}

			ips := sharedfw.CollectAddresses(nodes, v1.NodeExternalIP)

			ingressIP := sharedfw.GetIngressPoint(&svc.Status.LoadBalancer.Ingress[0])
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
					jig.TestReachableHTTP(false, ingressIP, svcTCPPort, sharedfw.KubeProxyLagTimeout)
					expectedSuccess := nodes.Items[n].Name == endpointNodeName
					port := strconv.Itoa(healthCheckNodePort)
					ipPort := net.JoinHostPort(publicIP, port)
					sharedfw.Logf("Health checking %s, http://%s%s, expectedSuccess %v", nodes.Items[n].Name, ipPort, path, expectedSuccess)
					Expect(jig.TestHTTPHealthCheckNodePort(publicIP, healthCheckNodePort, path, sharedfw.KubeProxyEndpointLagTimeout, expectedSuccess, threshold)).NotTo(HaveOccurred())
				}
				sharedfw.ExpectNoError(sharedfw.DeleteRCAndWaitForGC(f.ClientSet, namespace, serviceName))
			}
		})

		It("should work from pods", func() {
			namespace := f.Namespace.Name
			serviceName := "external-local"
			jig := sharedfw.NewServiceTestJig(cs, serviceName)
			nodes := jig.GetNodes(sharedfw.MaxNodesForEndpointsTests)

			svc := jig.CreateOnlyLocalLoadBalancerService(namespace, serviceName, loadBalancerCreateTimeout, true, nil)
			serviceLBNames = append(serviceLBNames, cloudprovider.GetLoadBalancerName(svc))
			defer func() {
				jig.ChangeServiceType(svc.Namespace, svc.Name, v1.ServiceTypeClusterIP, loadBalancerCreateTimeout)
				Expect(cs.CoreV1().Services(svc.Namespace).Delete(svc.Name, nil)).NotTo(HaveOccurred())
			}()

			ingressIP := sharedfw.GetIngressPoint(&svc.Status.LoadBalancer.Ingress[0])
			port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
			ipPort := net.JoinHostPort(ingressIP, port)
			path := fmt.Sprintf("%s/clientip", ipPort)
			nodeName := nodes.Items[0].Name
			podName := "execpod-sourceip"

			By(fmt.Sprintf("Creating %v on node %v", podName, nodeName))
			execPodName := sharedfw.CreateExecPodOrFail(f.ClientSet, namespace, podName, func(pod *v1.Pod) {
				pod.Spec.NodeName = nodeName
			})
			defer func() {
				err := cs.CoreV1().Pods(namespace).Delete(execPodName, nil)
				Expect(err).NotTo(HaveOccurred())
			}()
			execPod, err := f.ClientSet.CoreV1().Pods(namespace).Get(execPodName, metav1.GetOptions{})
			sharedfw.ExpectNoError(err)

			sharedfw.Logf("Waiting up to %v wget %v", sharedfw.KubeProxyLagTimeout, path)
			cmd := fmt.Sprintf(`wget -T 30 -qO- %v`, path)

			var srcIP string
			By(fmt.Sprintf("Hitting external lb %v from pod %v on node %v", ingressIP, podName, nodeName))
			if pollErr := wait.PollImmediate(sharedfw.K8sResourcePoll, sharedfw.LoadBalancerCreateTimeoutDefault, func() (bool, error) {
				stdout, err := sharedfw.RunHostCmd(execPod.Namespace, execPod.Name, cmd)
				if err != nil {
					sharedfw.Logf("got err: %v, retry until timeout", err)
					return false, nil
				}
				srcIP = strings.TrimSpace(strings.Split(stdout, ":")[0])
				return srcIP == execPod.Status.PodIP, nil
			}); pollErr != nil {
				sharedfw.Failf("Source IP not preserved from %v, expected '%v' got '%v'", podName, execPod.Status.PodIP, srcIP)
			}
		})
	})
})

var _ = Describe("End to end TLS", func() {
	baseName := "endtoendtls-service"
	f := sharedfw.NewDefaultFramework(baseName)
	Context("[cloudprovider][ccm]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer [Canary]", func() {
			serviceName := "e2e-tls-lb-test"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			sslSecretName := "ssl-certificate-secret"
			_, err := f.ClientSet.CoreV1().Secrets(ns).Create(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      sslSecretName,
				},
				Data: map[string][]byte{
					cloudprovider.SSLCAFileName:          []byte(sharedfw.SSLCAData),
					cloudprovider.SSLCertificateFileName: []byte(sharedfw.SSLCertificateData),
					cloudprovider.SSLPrivateKeyFileName:  []byte(sharedfw.SSLPrivateData),
					cloudprovider.SSLPassphrase:          []byte(sharedfw.SSLPassphrase),
				},
			})
			sharedfw.ExpectNoError(err)
			loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
			if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
				loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
			}

			requestedIP := ""

			tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeLoadBalancer
				s.Spec.LoadBalancerIP = requestedIP
				s.Spec.Ports = []v1.ServicePort{v1.ServicePort{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
					v1.ServicePort{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
				s.ObjectMeta.Annotations = map[string]string{cloudprovider.ServiceAnnotationLoadBalancerSSLPorts: "443",
					cloudprovider.ServiceAnnotationLoadBalancerTLSSecret:           sslSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerTLSBackendSetSecret: sslSecretName}

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
			sharedfw.Logf("TCP node port: %d", tcpNodePort)

			if requestedIP != "" && sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != requestedIP {
				sharedfw.Failf("unexpected TCP Status.LoadBalancer.Ingress (expected %s, got %s)", requestedIP, sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
			}
			tcpIngressIP := sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
			sharedfw.Logf("TCP load balancer: %s", tcpIngressIP)

			By("changing TCP service back to type=ClusterIP")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.Ports[0].NodePort = 0
				s.Spec.Ports[1].NodePort = 0
			})

			// Wait for the load balancer to be destroyed asynchronously
			tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)

			err = f.ClientSet.CoreV1().Secrets(ns).Delete(sslSecretName, nil)
			sharedfw.ExpectNoError(err)
		})
	})
})

var _ = Describe("BackendSet only enabled TLS", func() {

	baseName := "backendset-service"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer [Canary]", func() {
			serviceName := "backendset-tls-lb-test"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			sslSecretName := "ssl-certificate-secret"
			_, err := f.ClientSet.CoreV1().Secrets(ns).Create(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      sslSecretName,
				},
				Data: map[string][]byte{
					cloudprovider.SSLCAFileName:          []byte(sharedfw.SSLCAData),
					cloudprovider.SSLCertificateFileName: []byte(sharedfw.SSLCertificateData),
					cloudprovider.SSLPrivateKeyFileName:  []byte(sharedfw.SSLPrivateData),
					cloudprovider.SSLPassphrase:          []byte(sharedfw.SSLPassphrase),
				},
			})
			sharedfw.ExpectNoError(err)
			loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
			if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
				loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
			}

			requestedIP := ""

			tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeLoadBalancer
				s.Spec.LoadBalancerIP = requestedIP
				s.Spec.Ports = []v1.ServicePort{v1.ServicePort{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
					v1.ServicePort{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
				s.ObjectMeta.Annotations = map[string]string{cloudprovider.ServiceAnnotationLoadBalancerSSLPorts: "443",
					cloudprovider.ServiceAnnotationLoadBalancerTLSBackendSetSecret: sslSecretName}

			})

			svcPort := int(tcpService.Spec.Ports[0].Port)

			By("creating a pod to be part of the TCP service " + serviceName)
			jig.RunOrFail(ns, nil)

			By("waiting for the TCP service to have a load balancer")
			// Wait for the load balancer to be created asynchronously
			tcpService = jig.WaitForLoadBalancerOrFail(ns, tcpService.Name, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)

			tcpNodePort := int(tcpService.Spec.Ports[0].NodePort)
			sharedfw.Logf("TCP node port: %d", tcpNodePort)

			if requestedIP != "" && sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != requestedIP {
				sharedfw.Failf("unexpected TCP Status.LoadBalancer.Ingress (expected %s, got %s)", requestedIP, sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
			}
			tcpIngressIP := sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
			sharedfw.Logf("TCP load balancer: %s", tcpIngressIP)

			By("changing TCP service back to type=ClusterIP")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.Ports[0].NodePort = 0
				s.Spec.Ports[1].NodePort = 0
			})

			// Wait for the load balancer to be destroyed asynchronously
			tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)

			err = f.ClientSet.CoreV1().Secrets(ns).Delete(sslSecretName, nil)
			sharedfw.ExpectNoError(err)
		})
	})
})

var _ = Describe("Listener only enabled TLS", func() {

	baseName := "listener-service"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer [Canary]", func() {
			serviceName := "listener-tls-lb-test"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			sslSecretName := "ssl-certificate-secret"
			_, err := f.ClientSet.CoreV1().Secrets(ns).Create(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      sslSecretName,
				},
				Data: map[string][]byte{
					cloudprovider.SSLCAFileName:          []byte(sharedfw.SSLCAData),
					cloudprovider.SSLCertificateFileName: []byte(sharedfw.SSLCertificateData),
					cloudprovider.SSLPrivateKeyFileName:  []byte(sharedfw.SSLPrivateData),
					cloudprovider.SSLPassphrase:          []byte(sharedfw.SSLPassphrase),
				},
			})
			sharedfw.ExpectNoError(err)
			loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
			if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
				loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
			}

			requestedIP := ""

			tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeLoadBalancer
				s.Spec.LoadBalancerIP = requestedIP
				s.Spec.Ports = []v1.ServicePort{v1.ServicePort{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
					v1.ServicePort{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
				s.ObjectMeta.Annotations = map[string]string{cloudprovider.ServiceAnnotationLoadBalancerSSLPorts: "443",
					cloudprovider.ServiceAnnotationLoadBalancerTLSSecret: sslSecretName}

			})

			svcPort := int(tcpService.Spec.Ports[0].Port)

			By("creating a pod to be part of the TCP service " + serviceName)
			jig.RunOrFail(ns, nil)

			By("waiting for the TCP service to have a load balancer")
			// Wait for the load balancer to be created asynchronously
			tcpService = jig.WaitForLoadBalancerOrFail(ns, tcpService.Name, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)

			tcpNodePort := int(tcpService.Spec.Ports[0].NodePort)
			sharedfw.Logf("TCP node port: %d", tcpNodePort)

			if requestedIP != "" && sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != requestedIP {
				sharedfw.Failf("unexpected TCP Status.LoadBalancer.Ingress (expected %s, got %s)", requestedIP, sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
			}
			tcpIngressIP := sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
			sharedfw.Logf("TCP load balancer: %s", tcpIngressIP)

			By("changing TCP service back to type=ClusterIP")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.Ports[0].NodePort = 0
				s.Spec.Ports[1].NodePort = 0
			})

			// Wait for the load balancer to be destroyed asynchronously
			tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)

			err = f.ClientSet.CoreV1().Secrets(ns).Delete(sslSecretName, nil)
			sharedfw.ExpectNoError(err)
		})
	})
})

var _ = Describe("End to end enabled TLS - different certificates", func() {
	baseName := "e2e-diff-certs"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer [Canary]", func() {
			serviceName := "e2e-diff-certs-service"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			sslListenerSecretName := "ssl-certificate-secret-lis"
			sslBackendSetSecretName := "ssl-certificate-secret-backendset"
			_, err := f.ClientSet.CoreV1().Secrets(ns).Create(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      sslListenerSecretName,
				},
				Data: map[string][]byte{
					cloudprovider.SSLCAFileName:          []byte(sharedfw.SSLCAData),
					cloudprovider.SSLCertificateFileName: []byte(sharedfw.SSLCertificateData),
					cloudprovider.SSLPrivateKeyFileName:  []byte(sharedfw.SSLPrivateData),
					cloudprovider.SSLPassphrase:          []byte(sharedfw.SSLPassphrase),
				},
			})
			sharedfw.ExpectNoError(err)
			_, err = f.ClientSet.CoreV1().Secrets(ns).Create(&v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      sslBackendSetSecretName,
				},
				Data: map[string][]byte{
					cloudprovider.SSLCAFileName:          []byte(sharedfw.SSLCAData),
					cloudprovider.SSLCertificateFileName: []byte(sharedfw.SSLCertificateData),
					cloudprovider.SSLPrivateKeyFileName:  []byte(sharedfw.SSLPrivateData),
					cloudprovider.SSLPassphrase:          []byte(sharedfw.SSLPassphrase),
				},
			})
			sharedfw.ExpectNoError(err)
			loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
			if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
				loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
			}

			requestedIP := ""

			tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeLoadBalancer
				s.Spec.LoadBalancerIP = requestedIP
				s.Spec.Ports = []v1.ServicePort{v1.ServicePort{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
					v1.ServicePort{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
				s.ObjectMeta.Annotations = map[string]string{cloudprovider.ServiceAnnotationLoadBalancerSSLPorts: "443",
					cloudprovider.ServiceAnnotationLoadBalancerTLSSecret:           sslListenerSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerTLSBackendSetSecret: sslBackendSetSecretName}

			})

			svcPort := int(tcpService.Spec.Ports[0].Port)

			By("creating a pod to be part of the TCP service " + serviceName)
			jig.RunOrFail(ns, nil)

			By("waiting for the TCP service to have a load balancer")
			// Wait for the load balancer to be created asynchronously
			tcpService = jig.WaitForLoadBalancerOrFail(ns, tcpService.Name, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)

			tcpNodePort := int(tcpService.Spec.Ports[0].NodePort)
			sharedfw.Logf("TCP node port: %d", tcpNodePort)

			if requestedIP != "" && sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != requestedIP {
				sharedfw.Failf("unexpected TCP Status.LoadBalancer.Ingress (expected %s, got %s)", requestedIP, sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
			}
			tcpIngressIP := sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
			sharedfw.Logf("TCP load balancer: %s", tcpIngressIP)

			By("changing TCP service back to type=ClusterIP")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.Ports[0].NodePort = 0
				s.Spec.Ports[1].NodePort = 0
			})

			// Wait for the load balancer to be destroyed asynchronously
			tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)

			err = f.ClientSet.CoreV1().Secrets(ns).Delete(sslListenerSecretName, nil)
			sharedfw.ExpectNoError(err)
			err = f.ClientSet.CoreV1().Secrets(ns).Delete(sslBackendSetSecretName, nil)
			sharedfw.ExpectNoError(err)
		})
	})
})

var _ = Describe("LB Properties", func() {
	baseName := "lb-properties"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm][lb]", func() {

		It("should be possible to create Service type:LoadBalancer and mutate health-check config", func() {
			serviceName := "e2e-healthcheck-config"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
			if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
				loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
			}

			requestedIP := ""

			tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeLoadBalancer
				s.Spec.LoadBalancerIP = requestedIP
				s.Spec.Ports = []v1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
					{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
				s.ObjectMeta.Annotations = map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckRetries:  "1",
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckTimeout:  "1000",
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckInterval: "4000",
				}

			})

			svcPort := int(tcpService.Spec.Ports[0].Port)

			By("creating a pod to be part of the TCP service " + serviceName)
			jig.RunOrFail(ns, nil)

			By("waiting for the TCP service to have a load balancer")
			// Wait for the load balancer to be created asynchronously
			tcpService = jig.WaitForLoadBalancerOrFail(ns, tcpService.Name, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)

			tcpIngressIP := sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
			sharedfw.Logf("TCP load balancer: %s", tcpIngressIP)

			By("waiting upto 5m0s to verify initial health check config")
			lbName := cloudprovider.GetLoadBalancerName(tcpService)
			sharedfw.Logf("LB Name is %s", lbName)
			ctx := context.TODO()
			compartmentId := ""
			if setupF.Compartment1 != "" {
				compartmentId = setupF.Compartment1
			} else if f.CloudProviderConfig.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.CompartmentID
			} else if f.CloudProviderConfig.Auth.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.Auth.CompartmentID
			} else {
				sharedfw.Failf("Compartment Id undefined.")
			}
			loadBalancer, err := f.Client.LoadBalancer().GetLoadBalancerByName(ctx, compartmentId, lbName)
			sharedfw.ExpectNoError(err)
			err = f.VerifyHealthCheckConfig(*loadBalancer.Id, 1, 1000, 4000)
			sharedfw.ExpectNoError(err)

			By("changing TCP service health check config")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.ObjectMeta.Annotations = map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckRetries:  "2",
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckTimeout:  "2000",
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckInterval: "6000",
				}
			})

			By("waiting upto 5m0s to verify health check config after modification to initial")
			err = f.VerifyHealthCheckConfig(*loadBalancer.Id, 2, 2000, 6000)
			sharedfw.ExpectNoError(err)

			By("changing TCP service health check config - remove annotations")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.ObjectMeta.Annotations = map[string]string{}
			})

			By("waiting upto 5m0s to verify health check config should fall back to default after removing annotations")
			err = f.VerifyHealthCheckConfig(*loadBalancer.Id, 3, 3000, 10000)
			sharedfw.ExpectNoError(err)

			By("changing TCP service to type=ClusterIP")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.Ports[0].NodePort = 0
				s.Spec.Ports[1].NodePort = 0
			})

			// Wait for the load balancer to be destroyed asynchronously
			tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)

		})

		lbShapeTestArray := []struct {
			testName     string
			initialShape string
			tests        []struct {
				shape   string
				flexMin string
				flexMax string
			}
		}{
			{
				"Creating a fixed shape LB and convert it to a flexible LB shape",
				"400Mbps",
				[]struct {
					shape   string
					flexMin string
					flexMax string
				}{
					{
						"100Mbps",
						"",
						"",
					},
					{
						"flexible",
						"10",
						"100",
					},
				},
			},
			{
				"Create and update flexible LB",
				"flexible",
				[]struct {
					shape   string
					flexMin string
					flexMax string
				}{
					{
						"flexible",
						"50",
						"150",
					},
					// Note: We can't go back to fixed shape after converting to flexible shape.
					// Use Min and Max values to be the same value to get fixed shape LB
				},
			},
		}

		It("should be possible to update shape of Service of type:LoadBalancer ", func() {
			serviceName := "e2e-lb-shape"
			ns := f.Namespace.Name

			// TODO: Implement a config validator and stop supporting
			// different config versions
			compartmentId := ""
			if setupF.Compartment1 != "" {
				compartmentId = setupF.Compartment1
			} else if f.CloudProviderConfig.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.CompartmentID
			} else if f.CloudProviderConfig.Auth.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.Auth.CompartmentID
			} else {
				sharedfw.Failf("Compartment Id undefined.")
			}

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
			if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
				loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
			}

			requestedIP := ""

			// Create a service of type:ClusterIP and mutate that to create a LB
			tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.LoadBalancerIP = requestedIP
				s.Spec.Ports = []v1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
					{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}

			})
			By("creating a pod to be part of the TCP service " + serviceName)
			jig.RunOrFail(ns, nil)
			for _, lbShapeTest := range lbShapeTestArray {
				By(lbShapeTest.testName)
				tcpService = jig.UpdateServiceOrFail(ns, jig.Name, func(s *v1.Service) {
					s.Spec.Type = v1.ServiceTypeLoadBalancer
					s.Spec.LoadBalancerIP = requestedIP
					s.Spec.Ports = []v1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
						{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
					s.ObjectMeta.Annotations = map[string]string{
						cloudprovider.ServiceAnnotationLoadBalancerShape: lbShapeTest.initialShape,
						// Setting default values for Min and Max (Does not matter for fixed shape test)
						cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMin: "10",
						cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMax: "100",
					}

				})

				svcPort := int(tcpService.Spec.Ports[0].Port)

				By("waiting for the TCP service to have a load balancer")
				// Wait for the load balancer to be created asynchronously
				tcpService = jig.WaitForLoadBalancerOrFail(ns, tcpService.Name, loadBalancerCreateTimeout)
				jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)

				tcpIngressIP := sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
				sharedfw.Logf("TCP load balancer: %s", tcpIngressIP)

				By("Verifying Load Balancer shape")
				lbName := cloudprovider.GetLoadBalancerName(tcpService)
				ctx := context.TODO()

				loadBalancer, err := f.Client.LoadBalancer().GetLoadBalancerByName(ctx, compartmentId, lbName)
				sharedfw.ExpectNoError(err)
				sharedfw.Logf("Actual Load Balancer Shape: %s, Expected shape: %s", *loadBalancer.ShapeName, lbShapeTest.initialShape)
				Expect(strings.Compare(*loadBalancer.ShapeName, lbShapeTest.initialShape) == 0).To(BeTrue())
				if lbShapeTest.initialShape == "flexible" {
					sharedfw.Logf("Actual Load Balancer Flex Min: %d, Expected Flex Min: %d", *loadBalancer.ShapeDetails.MinimumBandwidthInMbps, 10)
					Expect(*loadBalancer.ShapeDetails.MinimumBandwidthInMbps == 10).To(BeTrue())
					sharedfw.Logf("Actual Load Balancer Flex Max: %d, Expected Flex Max: %d", *loadBalancer.ShapeDetails.MaximumBandwidthInMbps, 100)
					Expect(*loadBalancer.ShapeDetails.MaximumBandwidthInMbps == 100).To(BeTrue())
				}

				// Change shape and wait for LB to update
				for _, lbShape := range lbShapeTest.tests {
					By("changing LB shape to " + lbShape.shape + " flexMin:" + lbShape.flexMin + " flexMax:" + lbShape.flexMax)
					tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
						s.Annotations[cloudprovider.ServiceAnnotationLoadBalancerShape] = lbShape.shape
						s.Annotations[cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMin] = lbShape.flexMin
						s.Annotations[cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMax] = lbShape.flexMax
					})
					err = f.WaitForLoadBalancerShapeChange(loadBalancer, lbShape.shape, lbShape.flexMin, lbShape.flexMax)
					sharedfw.ExpectNoError(err)
				}

				By("changing TCP service to type=ClusterIP")
				tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
					s.Spec.Type = v1.ServiceTypeClusterIP
					s.Spec.Ports[0].NodePort = 0
					s.Spec.Ports[1].NodePort = 0
				})

				// Wait for the load balancer to be destroyed asynchronously
				tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
				jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)
			}
		})

		It("should be possible to create Service type:LoadBalancer and mutate connection idle timeout", func() {
			serviceName := "e2e-idle-timeout"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
			if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
				loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
			}

			requestedIP := ""

			tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeLoadBalancer
				s.Spec.LoadBalancerIP = requestedIP
				s.Spec.Ports = []v1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
					{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
				s.ObjectMeta.Annotations = map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerConnectionIdleTimeout: "500",
				}
			})

			svcPort := int(tcpService.Spec.Ports[0].Port)

			By("creating a pod to be part of the TCP service " + serviceName)
			jig.RunOrFail(ns, nil)

			By("waiting for the TCP service to have a load balancer")
			// Wait for the load balancer to be created asynchronously
			tcpService = jig.WaitForLoadBalancerOrFail(ns, tcpService.Name, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)

			tcpIngressIP := sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
			sharedfw.Logf("TCP load balancer: %s", tcpIngressIP)

			By("waiting upto 5m0s to verify default connection idle timeout")
			lbName := cloudprovider.GetLoadBalancerName(tcpService)
			ctx := context.TODO()
			compartmentId := ""
			if setupF.Compartment1 != "" {
				compartmentId = setupF.Compartment1
			} else if f.CloudProviderConfig.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.CompartmentID
			} else if f.CloudProviderConfig.Auth.CompartmentID != "" {
				compartmentId = f.CloudProviderConfig.Auth.CompartmentID
			} else {
				sharedfw.Failf("Compartment Id undefined.")
			}
			loadBalancer, err := f.Client.LoadBalancer().GetLoadBalancerByName(ctx, compartmentId, lbName)
			sharedfw.ExpectNoError(err)
			err = f.VerifyLoadBalancerConnectionIdleTimeout(*loadBalancer.Id, 500)
			sharedfw.ExpectNoError(err)

			By("changing TCP service health check config")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.ObjectMeta.Annotations = map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerConnectionIdleTimeout: "800",
				}
			})

			By("waiting upto 5m0s to verify health check config after modification to initial")
			err = f.VerifyLoadBalancerConnectionIdleTimeout(*loadBalancer.Id, 800)
			sharedfw.ExpectNoError(err)

			By("changing TCP service to type=ClusterIP")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.Ports[0].NodePort = 0
				s.Spec.Ports[1].NodePort = 0
			})

			// Wait for the load balancer to be destroyed asynchronously
			tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)

		})
	})
})
