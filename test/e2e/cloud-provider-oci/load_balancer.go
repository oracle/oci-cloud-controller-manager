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
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cloudprovider "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
	sharedfw "github.com/oracle/oci-cloud-controller-manager/test/e2e/framework"
	"github.com/oracle/oci-go-sdk/v65/containerengine"
	"github.com/oracle/oci-go-sdk/v65/core"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
)

var _ = Describe("Service [Slow]", func() {

	baseName := "service"
	f := sharedfw.NewDefaultFramework(baseName)

	testDefinedTags := map[string]map[string]interface{}{"oke-tag": {"oke-tagging": "ccm-test-integ"}}
	testDefinedTagsByteArray, _ := json.Marshal(testDefinedTags)

	basicTestArray := []struct {
		lbType              string
		CreationAnnotations map[string]string
	}{
		{
			"lb",
			map[string]string{},
		},
		{
			"nlb",
			map[string]string{
				cloudprovider.ServiceAnnotationLoadBalancerType:                              "nlb",
				cloudprovider.ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
			},
		},
		{
			"lb-wris",
			map[string]string{
				cloudprovider.ServiceAnnotationServiceAccountName:                     "sa",
				cloudprovider.ServiceAnnotationLoadBalancerInitialDefinedTagsOverride: string(testDefinedTagsByteArray),
			},
		},
		{
			"nlb-wris",
			map[string]string{
				cloudprovider.ServiceAnnotationServiceAccountName:                            "sa",
				cloudprovider.ServiceAnnotationLoadBalancerType:                              "nlb",
				cloudprovider.ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
				cloudprovider.ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride: string(testDefinedTagsByteArray),
			},
		},
	}
	Context("[cloudprovider][ccm][lb][SL][wris][system-tags]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer (change nodeport) [Canary]", func() {
			for _, test := range basicTestArray {
				if strings.HasSuffix(test.lbType, "-wris") && f.ClusterType != containerengine.ClusterTypeEnhancedCluster {
					sharedfw.Logf("Skipping Workload Identity Principal test for LB Type (%s) because the cluster is not an OKE ENHANCED_CLUSTER", test.lbType)
					continue
				}
				func() {
					By("Running test for: " + test.lbType)
					serviceName := "basic-" + test.lbType + "-test"
					ns := f.Namespace.Name

					jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

					nodeIP := sharedfw.PickNodeIP(jig.Client) // for later

					loadBalancerLagTimeout := sharedfw.LoadBalancerLagTimeoutDefault
					loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
					if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
						loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
					}
					var serviceAccount *v1.ServiceAccount
					if sa, exists := test.CreationAnnotations[cloudprovider.ServiceAnnotationServiceAccountName]; exists {
						// Create a service account in the same namespace as the service
						By("creating service account \"sa\" in namespace " + ns)
						serviceAccount = jig.CreateServiceAccountOrFail(ns, sa, nil)
					}

					// TODO(apryde): Test that LoadBalancers can receive static IP addresses
					// (in a provider agnostic manner?). OCI does not currently
					// support this.
					requestedIP := ""

					tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
						s.Spec.Type = v1.ServiceTypeLoadBalancer
						s.Spec.LoadBalancerIP = requestedIP // will be "" if not applicable
						s.ObjectMeta.Annotations = test.CreationAnnotations
					})

					if _, exists := test.CreationAnnotations[cloudprovider.ServiceAnnotationServiceAccountName]; exists {
						By("setting service account \"sa\" owner reference as the TCP service " + serviceName)
						// Set SA owner reference as the service to prevent deletion of service account before the service
						jig.SetServiceOwnerReferenceOnServiceAccountOrFail(ns, serviceAccount, tcpService)
						defer func() {
							dp := metav1.DeletePropagationForeground
							jig.Client.CoreV1().Services(f.Namespace.Name).Delete(context.Background(), tcpService.Name, metav1.DeleteOptions{
								PropagationPolicy: &dp,
							})
							time.Sleep(time.Second)
						}()
					}

					svcPort := int(tcpService.Spec.Ports[0].Port)

					By("creating a pod to be part of the TCP service " + serviceName)
					jig.RunOrFail(ns, nil)

					// TODO(apryde): Test UDP service. OCI does not currently support this.

					By("waiting for the TCP service to have a load balancer")
					// Wait for the load balancer to be created asynchronously
					tcpService = jig.WaitForLoadBalancerOrFail(ns, tcpService.Name, loadBalancerCreateTimeout)
					jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)

					if strings.HasSuffix(test.lbType, "-wris") {
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
						lbType := strings.TrimSuffix(test.lbType, "-wris")
						loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), lbType, "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
						sharedfw.ExpectNoError(err)

						if !reflect.DeepEqual(loadBalancer.DefinedTags, testDefinedTags) {
							sharedfw.Failf("Defined tag mismatch! Expected: %v, Got: %v", testDefinedTags, loadBalancer.DefinedTags)
						}
					}

					if strings.HasSuffix(test.lbType, "-wris") {
						sharedfw.Logf("skip evaluating system tag when the principal type is Workload identity")
					} else {
						By("validating system tags on the loadbalancer")
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
						lbType := test.lbType
						if strings.HasSuffix(test.lbType, "-wris") {
							lbType = strings.TrimSuffix(test.lbType, "-wris")
						}
						loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), lbType, "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
						sharedfw.ExpectNoError(err)
						sharedfw.Logf("Loadbalancer details %v:", loadBalancer)
						if setupF.AddOkeSystemTags && !sharedfw.HasOkeSystemTags(loadBalancer.SystemTags) {
							sharedfw.Failf("Loadbalancer is expected to have the system tags")
						}
					}

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
				}()
			}
		})
	})
})

var _ = Describe("Service NSG [Slow]", func() {

	baseName := "service"
	f := sharedfw.NewDefaultFramework(baseName)

	basicTestArray := []struct {
		lbType              string
		CreationAnnotations map[string]string
	}{
		{
			"lb",
			map[string]string{
				cloudprovider.ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "NSG",
				cloudprovider.ServiceAnnotationBackendSecurityRuleManagement:          f.BackendNsgOcids,
			},
		},
		{
			"nlb",
			map[string]string{
				cloudprovider.ServiceAnnotationLoadBalancerType:                       "nlb",
				cloudprovider.ServiceAnnotationLoadBalancerSecurityRuleManagementMode: "NSG",
				cloudprovider.ServiceAnnotationBackendSecurityRuleManagement:          f.BackendNsgOcids,
			},
		},
	}
	Context("[cloudprovider][ccm][lb][managedNsg]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer (change nodeport) [Canary]", func() {
			for _, test := range basicTestArray {
				if strings.HasSuffix(test.lbType, "-wris") && f.ClusterType != containerengine.ClusterTypeEnhancedCluster {
					sharedfw.Logf("Skipping Workload Identity Principal test for LB Type (%s) because the cluster is not an OKE ENHANCED_CLUSTER", test.lbType)
					continue
				}

				By("Running test for: " + test.lbType)
				serviceName := "basic-" + test.lbType + "-test"
				ns := f.Namespace.Name

				jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

				nodeIP := sharedfw.PickNodeIP(jig.Client) // for later

				loadBalancerLagTimeout := sharedfw.LoadBalancerLagTimeoutDefault
				loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
				if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
					loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
				}

				if sa, exists := test.CreationAnnotations[cloudprovider.ServiceAnnotationServiceAccountName]; exists {
					// Create a service account in the same namespace as the service
					jig.CreateServiceAccountOrFail(ns, sa, nil)
				}

				// TODO(apryde): Test that LoadBalancers can receive static IP addresses
				// (in a provider agnostic manner?). OCI does not currently
				// support this.
				requestedIP := ""

				tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
					s.Spec.Type = v1.ServiceTypeLoadBalancer
					s.Spec.LoadBalancerIP = requestedIP // will be "" if not applicable
					s.ObjectMeta.Annotations = test.CreationAnnotations
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

				if f.NodePortTest {
					By("hitting the TCP service's NodePort")
					jig.TestReachableHTTP(false, nodeIP, tcpNodePort, sharedfw.KubeProxyLagTimeout)
				}

				By("hitting the TCP service's LoadBalancer")
				jig.TestReachableHTTP(false, tcpIngressIP, svcPort, loadBalancerLagTimeout)

				// Change the services' node ports.

				By("changing the TCP service's NodePort")
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
				lbType := strings.TrimSuffix(test.lbType, "-wris")
				loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), lbType, "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
				sharedfw.ExpectNoError(err)

				frontendNsgId := loadBalancer.NetworkSecurityGroupIds[0]
				// Count the number of ingress/egress rules with the original port so
				// we can check the correct number are updated.
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

				// Check the correct number of rules are present on the NSG
				sharedfw.WaitForSinglePortRulesAfterPortChangeOrFailNSG(f.Client, frontendNsgId, tcpNodePortOld, tcpNodePort, core.SecurityRuleDirectionEgress)
				backendNsgList := strings.Split(strings.ReplaceAll(setupF.BackendNsgOcid, " ", ""), ",")

				for _, backendNsg := range backendNsgList {
					sharedfw.WaitForSinglePortRulesAfterPortChangeOrFailNSG(f.Client, backendNsg, tcpNodePortOld, tcpNodePort, core.SecurityRuleDirectionIngress)
				}

				// Check if rules are not modified on the security list.
				numEgressRules, numIngressRules := sharedfw.CountSinglePortSecListRules(f.Client, f.CCMSecListID, f.K8SSecListID, tcpNodePort)
				if numEgressRules != 0 || numIngressRules != 0 {
					sharedfw.Logf("Count of Egress Rules added to sec list %d", numEgressRules)
					sharedfw.Logf("Count of Ingress Rules added to sec list %d", numIngressRules)
					sharedfw.Failf("Security List rules modified while service should be using NSG on port %d", tcpNodePort)
				}

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
				sharedfw.WaitForSinglePortRulesAfterPortChangeOrFailNSG(f.Client, frontendNsgId, svcPortOld, svcPort, core.SecurityRuleDirectionIngress)

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
			}
		})
	})
})

// This test covers support for node-local routing (i.e. avoidance of a second hop).
var _ = Describe("Node Local Routing [Slow]", func() {

	baseName := "node-local-routing"
	f := sharedfw.NewDefaultFramework(baseName)

	loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
	serviceLBNames := []string{}

	var cs clientset.Interface
	BeforeEach(func() {
		cs = f.ClientSet
	})
	nodeLocalTestsArray := []struct {
		lbType              string
		CreationAnnotations map[string]string
	}{
		{
			"lb",
			map[string]string{},
		},
		{
			"nlb",
			map[string]string{
				cloudprovider.ServiceAnnotationLoadBalancerType:                              "nlb",
				cloudprovider.ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
			},
		},
	}
	Context("[cloudprovider][ccm][lb][node-local]", func() {
		It("should only target nodes with endpoints", func() {
			for _, test := range nodeLocalTestsArray {
				By("Running test for: " + test.lbType)
				namespace := f.Namespace.Name
				serviceName := "external-local-" + test.lbType
				jig := sharedfw.NewServiceTestJig(cs, serviceName)
				nodes := jig.GetNodes(sharedfw.MaxNodesForEndpointsTests)

				svc := jig.CreateOnlyLocalLoadBalancerService(namespace, serviceName, loadBalancerCreateTimeout, false, test.CreationAnnotations,
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
					Expect(cs.CoreV1().Services(svc.Namespace).Delete(context.Background(), svc.Name, metav1.DeleteOptions{})).NotTo(HaveOccurred())
				}()

				healthCheckNodePort := int(svc.Spec.HealthCheckNodePort)
				if healthCheckNodePort == 0 {
					sharedfw.Failf("Service HealthCheck NodePort was not allocated")
				}

				ips := sharedfw.CollectAddresses(nodes, v1.NodeInternalIP)

				ingressIP := sharedfw.GetIngressPoint(&svc.Status.LoadBalancer.Ingress[0])
				svcTCPPort := int(svc.Spec.Ports[0].Port)

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

					// Make sure the loadbalancer picked up the health check change.
					// Confirm traffic can reach backend through LB before checking healthcheck nodeport.
					jig.TestReachableHTTP(false, ingressIP, svcTCPPort, sharedfw.KubeProxyLagTimeout)

					By("Creating a job to check pods health")
					script := CreateHealthCheckScript(healthCheckNodePort, ips, path, i)
					jig.CreateJobRunningScript(namespace, script, 3, test.lbType+"-health-checker-"+strconv.Itoa(i))
					sharedfw.ExpectNoError(sharedfw.DeleteRCAndWaitForGC(f.ClientSet, namespace, serviceName))
				}
			}
		})
		It("should work from pods", func() {
			for _, test := range nodeLocalTestsArray {
				By("Running test for: " + test.lbType)
				namespace := f.Namespace.Name
				serviceName := "external-local-" + test.lbType
				jig := sharedfw.NewServiceTestJig(cs, serviceName)
				nodes := jig.GetNodes(sharedfw.MaxNodesForEndpointsTests)

				svc := jig.CreateOnlyLocalLoadBalancerService(namespace, serviceName, loadBalancerCreateTimeout, true, test.CreationAnnotations, func(s *v1.Service) {
					if test.lbType == "lb" {
						s.ObjectMeta.Annotations = map[string]string{
							cloudprovider.ServiceAnnotationLoadBalancerInternal:     "true",
							cloudprovider.ServiceAnnotationLoadBalancerShape:        "flexible",
							cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMin: "10",
							cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMax: "100",
						}
					}
					if test.lbType == "nlb" {
						s.ObjectMeta.Annotations = map[string]string{
							cloudprovider.ServiceAnnotationNetworkLoadBalancerInternal: "true",
						}
					}
				})
				serviceLBNames = append(serviceLBNames, cloudprovider.GetLoadBalancerName(svc))
				defer func() {
					jig.ChangeServiceType(svc.Namespace, svc.Name, v1.ServiceTypeClusterIP, loadBalancerCreateTimeout)
					Expect(cs.CoreV1().Services(svc.Namespace).Delete(context.Background(), svc.Name, metav1.DeleteOptions{})).NotTo(HaveOccurred())
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
					err := cs.CoreV1().Pods(namespace).Delete(context.Background(), execPodName, metav1.DeleteOptions{})
					Expect(err).NotTo(HaveOccurred())
				}()
				execPod, err := f.ClientSet.CoreV1().Pods(namespace).Get(context.Background(), execPodName, metav1.GetOptions{})
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
			}
		})
	})
})

var _ = Describe("IpMode [Slow]", func() {

	baseName := "ingress-ipmode"
	f := sharedfw.NewDefaultFramework(baseName)

	loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
	serviceLBNames := []string{}

	var cs clientset.Interface
	BeforeEach(func() {
		cs = f.ClientSet
	})
	esippTestsArray := []struct {
		lbType              string
		CreationAnnotations map[string]string
	}{
		{
			"nlb",
			map[string]string{
				cloudprovider.ServiceAnnotationLoadBalancerType:                              "nlb",
				cloudprovider.ServiceAnnotationNetworkLoadBalancerInternal:                   "true",
				cloudprovider.ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
				cloudprovider.ServiceAnnotationNetworkLoadBalancerIsPreserveSource:           "false",
				cloudprovider.ServiceAnnotationIngressIpMode:                                 "Proxy",
			},
		},
	}
	Context("[cloudprovider][ccm][lb][ipMode]", func() {
		It("traffic should work from pods via load balancer", func() {
			if sharedfw.CompareVersions(f.OkeClusterK8sVersion, "v1.30") < 0 {
				Skip("Cluster K8s Version " + f.OkeClusterK8sVersion + " is less than v1.30, skipping test for Load Balancer ingress ipMode=\"Proxy\"")
			}
			for _, test := range esippTestsArray {
				By("Running test for: " + test.lbType)
				namespace := f.Namespace.Name
				serviceName := "internal-local-" + test.lbType
				jig := sharedfw.NewServiceTestJig(cs, serviceName)
				nodes := jig.GetNodes(sharedfw.MaxNodesForEndpointsTests)

				svc := jig.CreateOnlyLocalLoadBalancerService(namespace, serviceName, loadBalancerCreateTimeout, true, test.CreationAnnotations, nil)
				serviceLBNames = append(serviceLBNames, cloudprovider.GetLoadBalancerName(svc))
				defer func() {
					jig.ChangeServiceType(svc.Namespace, svc.Name, v1.ServiceTypeClusterIP, loadBalancerCreateTimeout)
					Expect(cs.CoreV1().Services(svc.Namespace).Delete(context.Background(), svc.Name, metav1.DeleteOptions{})).NotTo(HaveOccurred())
				}()

				if *svc.Status.LoadBalancer.Ingress[0].IPMode != v1.LoadBalancerIPModeProxy {
					sharedfw.Failf("IpMode on the service is '%v', expected 'Proxy'", *svc.Status.LoadBalancer.Ingress[0].IPMode)
				}

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
					err := cs.CoreV1().Pods(namespace).Delete(context.Background(), execPodName, metav1.DeleteOptions{})
					Expect(err).NotTo(HaveOccurred())
				}()
				execPod, err := f.ClientSet.CoreV1().Pods(namespace).Get(context.Background(), execPodName, metav1.GetOptions{})
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
					return srcIP == ingressIP, nil
				}); pollErr != nil {
					sharedfw.Failf("Source IP not preserved from %v, expected '%v' got '%v'", podName, ingressIP, srcIP)
				}
			}
		})
	})
})

// Test for ESIPP (External Source IP Preservation) via NLB using new ipMode="Proxy".
// Source and destination pods need to be on different nodes for this test so the test will be skipped id there are less than 2 nodes.
var _ = Describe("ESIPP - IpMode Proxy [Slow]", func() {

	baseName := "esipp-internal"
	f := sharedfw.NewDefaultFramework(baseName)

	loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
	serviceLBNames := []string{}

	var cs clientset.Interface
	BeforeEach(func() {
		cs = f.ClientSet
	})
	esippTestsArray := []struct {
		lbType              string
		CreationAnnotations map[string]string
	}{
		{
			"nlb",
			map[string]string{
				cloudprovider.ServiceAnnotationLoadBalancerType:                              "nlb",
				cloudprovider.ServiceAnnotationNetworkLoadBalancerInternal:                   "true",
				cloudprovider.ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode: "All",
				cloudprovider.ServiceAnnotationNetworkLoadBalancerIsPreserveSource:           "true",
				cloudprovider.ServiceAnnotationIngressIpMode:                                 "Proxy",
			},
		},
	}
	Context("[cloudprovider][ccm][lb][esipp]", func() {
		It("should preserve source IP of pod with ipMode Proxy", func() {
			if sharedfw.CompareVersions(f.OkeClusterK8sVersion, "v1.30") < 0 {
				Skip("Cluster K8s Version " + f.OkeClusterK8sVersion + " is less than v1.30, skipping test for Load Balancer ingress ipMode=\"Proxy\"")
			}
			for _, test := range esippTestsArray {
				By("Running test for: " + test.lbType)
				namespace := f.Namespace.Name
				serviceName := "internal-local-" + test.lbType
				jig := sharedfw.NewServiceTestJig(cs, serviceName)
				nodes := jig.GetNodes(sharedfw.MaxNodesForEndpointsTests)
				// Can not run the test if the cluster has less than 2 nodes
				if len(nodes.Items) < 2 {
					// We can decide to scale the nodepool as well. We already do so for [node-local] test.
					Skip("Skipping test since cluster has less than 2 nodes")
				}

				By("creating a pod to be part of the service " + serviceName)
				jig.RunOrFail(namespace, func(s *v1.ReplicationController) {
					nodeName := nodes.Items[0].Name
					s.Spec.Template.Spec.NodeName = nodeName
				})
				svc := jig.CreateOnlyLocalLoadBalancerService(namespace, serviceName, loadBalancerCreateTimeout, false, test.CreationAnnotations, func(s *v1.Service) {
					s.Spec.Ports = []v1.ServicePort{v1.ServicePort{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
						v1.ServicePort{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
				})
				serviceLBNames = append(serviceLBNames, cloudprovider.GetLoadBalancerName(svc))
				defer func() {
					jig.ChangeServiceType(svc.Namespace, svc.Name, v1.ServiceTypeClusterIP, loadBalancerCreateTimeout)
					Expect(cs.CoreV1().Services(svc.Namespace).Delete(context.Background(), svc.Name, metav1.DeleteOptions{})).NotTo(HaveOccurred())
				}()

				if *svc.Status.LoadBalancer.Ingress[0].IPMode != v1.LoadBalancerIPModeProxy {
					sharedfw.Failf("IpMode on the service is '%v', expected 'Proxy'", *svc.Status.LoadBalancer.Ingress[0].IPMode)
				}

				ingressIP := sharedfw.GetIngressPoint(&svc.Status.LoadBalancer.Ingress[0])
				port := strconv.Itoa(int(svc.Spec.Ports[0].Port))
				ipPort := net.JoinHostPort(ingressIP, port)
				path := fmt.Sprintf("%s/clientip", ipPort)
				nodeName := nodes.Items[1].Name
				podName := "execpod-sourceip"

				By(fmt.Sprintf("Creating %v on node %v", podName, nodeName))
				execPodName := sharedfw.CreateExecPodOrFail(f.ClientSet, namespace, podName, func(pod *v1.Pod) {
					pod.Spec.NodeName = nodeName
				})
				defer func() {
					err := cs.CoreV1().Pods(namespace).Delete(context.Background(), execPodName, metav1.DeleteOptions{})
					Expect(err).NotTo(HaveOccurred())
				}()
				execPod, err := f.ClientSet.CoreV1().Pods(namespace).Get(context.Background(), execPodName, metav1.GetOptions{})
				sharedfw.ExpectNoError(err)

				sharedfw.Logf("Waiting up to %v wget %v", sharedfw.KubeProxyLagTimeout, path)
				cmd := fmt.Sprintf(`wget -T 30 -qO- %v`, path)

				var srcIP, expectedIP string
				By(fmt.Sprintf("Hitting external lb %v from pod %v (%v) on node %v", ingressIP, podName, execPod.Status.PodIP, nodeName))
				if pollErr := wait.PollImmediate(sharedfw.K8sResourcePoll, 5*time.Minute, func() (bool, error) {
					expectedIP = execPod.Spec.NodeName // Node IP

					stdout, err := sharedfw.RunHostCmd(execPod.Namespace, execPod.Name, cmd)
					if err != nil {
						sharedfw.Logf("got err: %v, retry until timeout", err)
						return false, nil
					}
					srcIP = strings.TrimSpace(strings.Split(stdout, ":")[0])
					return srcIP == expectedIP, nil
				}); pollErr != nil {
					sharedfw.Failf("Source IP not preserved from %v, expected '%v' got '%v'", podName, expectedIP, srcIP)
				}
			}
		})
	})
})

var _ = Describe("End to end TLS", func() {
	baseName := "endtoendtls-service"
	f := sharedfw.NewDefaultFramework(baseName)
	Context("[cloudprovider][ccm][lb]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer [Canary]", func() {
			serviceName := "e2e-tls-lb-test"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			sslSecretName := "ssl-certificate-secret"
			_, err := f.ClientSet.CoreV1().Secrets(ns).Create(context.Background(), &v1.Secret{
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
			}, metav1.CreateOptions{})
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
					cloudprovider.ServiceAnnotationLoadBalancerTLSBackendSetSecret: sslSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerInternal:            "true",
				}
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

			err = f.ClientSet.CoreV1().Secrets(ns).Delete(context.Background(), sslSecretName, metav1.DeleteOptions{})
			sharedfw.ExpectNoError(err)
		})
	})
})

var _ = Describe("CipherSuite tests Loadbalancer TLS", func() {
	baseName := "endtoendtls-service"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm][lb][sslconfig]", func() {
		It("should be possible to create and mutate a CipherSuite for Service type:LoadBalancer [Canary]", func() {
			serviceName := "e2e-tls-lb-test"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			sslSecretName := "ssl-certificate-secret"
			_, err := f.ClientSet.CoreV1().Secrets(ns).Create(context.Background(), &v1.Secret{
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
			}, metav1.CreateOptions{})
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
				s.ObjectMeta.Annotations = map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerSSLPorts:            "443",
					cloudprovider.ServiceAnnotationLoadBalancerTLSSecret:           sslSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerTLSBackendSetSecret: sslSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerInternal:            "true",
					cloudprovider.ServiceAnnotationLoadbalancerBackendSetSSLConfig: `{"CipherSuiteName":"oci-default-http2-ssl-cipher-suite-v1", "Protocols":["TLSv1.2"]}`,
					cloudprovider.ServiceAnnotationLoadbalancerListenerSSLConfig:   `{"CipherSuiteName":"oci-default-http2-ssl-cipher-suite-v1", "Protocols":["TLSv1.2"]}`}
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
			if requestedIP != "" && sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]) != requestedIP {
				sharedfw.Failf("unexpected TCP Status.LoadBalancer.Ingress (expected %s, got %s)", requestedIP, sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0]))
			}
			tcpIngressIP := sharedfw.GetIngressPoint(&tcpService.Status.LoadBalancer.Ingress[0])
			sharedfw.Logf("TCP load balancer: %s", tcpIngressIP)

			loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), "lb", "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
			sharedfw.ExpectNoError(err)

			err = f.WaitForLoadBalancerSSLConfigurationChange(loadBalancer, tcpService)
			sharedfw.ExpectNoError(err)

			By("changing the TCP service's CipherSuite and Protocols")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.ObjectMeta.Annotations = map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerSSLPorts:            "443",
					cloudprovider.ServiceAnnotationLoadBalancerTLSSecret:           sslSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerTLSBackendSetSecret: sslSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerInternal:            "true",
					cloudprovider.ServiceAnnotationLoadbalancerListenerSSLConfig:   `{"CipherSuiteName":"oci-tls-13-recommended-ssl-cipher-suite-v1", "Protocols":["TLSv1.3"]}`,
					cloudprovider.ServiceAnnotationLoadbalancerBackendSetSSLConfig: `{"CipherSuiteName":"oci-tls-13-recommended-ssl-cipher-suite-v1", "Protocols":["TLSv1.3"]}`,
				}
			})
			jig.SanityCheckService(tcpService, v1.ServiceTypeLoadBalancer)

			err = f.WaitForLoadBalancerSSLConfigurationChange(loadBalancer, tcpService)
			sharedfw.ExpectNoError(err)

			By("changing TCP service back to type=ClusterIP")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.Ports[0].NodePort = 0
				s.Spec.Ports[1].NodePort = 0
			})

			// Wait for the load balancer to be destroyed asynchronously
			tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)

			err = f.ClientSet.CoreV1().Secrets(ns).Delete(context.Background(), sslSecretName, metav1.DeleteOptions{})
			sharedfw.ExpectNoError(err)
		})
	})
})

var _ = Describe("BackendSet only enabled TLS", func() {

	baseName := "backendset-service"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm][lb]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer [Canary]", func() {
			serviceName := "backendset-tls-lb-test"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			sslSecretName := "ssl-certificate-secret"
			_, err := f.ClientSet.CoreV1().Secrets(ns).Create(context.Background(), &v1.Secret{
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
			}, metav1.CreateOptions{})
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
					cloudprovider.ServiceAnnotationLoadBalancerTLSBackendSetSecret: sslSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerInternal:            "true",
				}
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

			err = f.ClientSet.CoreV1().Secrets(ns).Delete(context.Background(), sslSecretName, metav1.DeleteOptions{})
			sharedfw.ExpectNoError(err)
		})
	})
})

var _ = Describe("Listener only enabled TLS", func() {

	baseName := "listener-service"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm][lb]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer [Canary]", func() {
			serviceName := "listener-tls-lb-test"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			sslSecretName := "ssl-certificate-secret"
			_, err := f.ClientSet.CoreV1().Secrets(ns).Create(context.Background(), &v1.Secret{
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
			}, metav1.CreateOptions{})
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
				s.ObjectMeta.Annotations = map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerSSLPorts:  "443",
					cloudprovider.ServiceAnnotationLoadBalancerTLSSecret: sslSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerInternal:  "true",
				}
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

			err = f.ClientSet.CoreV1().Secrets(ns).Delete(context.Background(), sslSecretName, metav1.DeleteOptions{})
			sharedfw.ExpectNoError(err)
		})
	})
})

var _ = Describe("GRPC Listeners only enabled TLS", func() {

	baseName := "listener-service"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm][lb][grpc]", func() {
		It("should be possible to create a GRPC listener for LB [Canary]", func() {
			serviceName := "listener-grpc-lb-test"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			sslSecretName := "ssl-certificate-secret"
			_, err := f.ClientSet.CoreV1().Secrets(ns).Create(context.Background(), &v1.Secret{
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
			}, metav1.CreateOptions{})
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
				s.ObjectMeta.Annotations = map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerSSLPorts:   "80,443",
					cloudprovider.ServiceAnnotationLoadBalancerTLSSecret:  sslSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerInternal:   "true",
					cloudprovider.ServiceAnnotationLoadBalancerBEProtocol: "GRPC",
				}
			})

			svcPort := int(tcpService.Spec.Ports[0].Port)

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

			By("Validate that GRPC Listener is created")
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
			loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), cloudprovider.LB, "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
			sharedfw.ExpectNoError(err)
			protocol := *loadBalancer.Listeners["GRPC-443"].Protocol
			Expect(protocol == cloudprovider.ProtocolGrpc).To(BeTrue())

			By("changing TCP service back to type=ClusterIP")
			tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
				s.Spec.Type = v1.ServiceTypeClusterIP
				s.Spec.Ports[0].NodePort = 0
				s.Spec.Ports[1].NodePort = 0
			})

			// Wait for the load balancer to be destroyed asynchronously
			tcpService = jig.WaitForLoadBalancerDestroyOrFail(ns, tcpService.Name, tcpIngressIP, svcPort, loadBalancerCreateTimeout)
			jig.SanityCheckService(tcpService, v1.ServiceTypeClusterIP)

			err = f.ClientSet.CoreV1().Secrets(ns).Delete(context.Background(), sslSecretName, metav1.DeleteOptions{})
			sharedfw.ExpectNoError(err)
		})
	})
})

var _ = Describe("End to end enabled TLS - different certificates", func() {
	baseName := "e2e-diff-certs"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm][lb]", func() {
		It("should be possible to create and mutate a Service type:LoadBalancer [Canary]", func() {
			serviceName := "e2e-diff-certs-service"
			ns := f.Namespace.Name

			jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

			sslListenerSecretName := "ssl-certificate-secret-lis"
			sslBackendSetSecretName := "ssl-certificate-secret-backendset"
			_, err := f.ClientSet.CoreV1().Secrets(ns).Create(context.Background(), &v1.Secret{
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
			}, metav1.CreateOptions{})
			sharedfw.ExpectNoError(err)
			_, err = f.ClientSet.CoreV1().Secrets(ns).Create(context.Background(), &v1.Secret{
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
			}, metav1.CreateOptions{})
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
				s.ObjectMeta.Annotations = map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerSSLPorts:            "443",
					cloudprovider.ServiceAnnotationLoadBalancerTLSSecret:           sslListenerSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerTLSBackendSetSecret: sslBackendSetSecretName,
					cloudprovider.ServiceAnnotationLoadBalancerInternal:            "true",
				}
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

			err = f.ClientSet.CoreV1().Secrets(ns).Delete(context.Background(), sslListenerSecretName, metav1.DeleteOptions{})
			sharedfw.ExpectNoError(err)
			err = f.ClientSet.CoreV1().Secrets(ns).Delete(context.Background(), sslBackendSetSecretName, metav1.DeleteOptions{})
			sharedfw.ExpectNoError(err)
		})
	})
})

var _ = Describe("Configure preservation of source IP in NLB", func() {

	baseName := "preserve-source"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm][lb]", func() {
		preserveSourceTestArray := []struct {
			lbType           string
			configuration    string
			annotations      map[string]string
			isPreserveSource bool
		}{
			{
				"nlb",
				"preserve-ip-true",
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerType:                    "nlb",
					cloudprovider.ServiceAnnotationNetworkLoadBalancerIsPreserveSource: "true",
					cloudprovider.ServiceAnnotationLoadBalancerInternal:                "true",
				},
				true,
			},
			{
				"nlb",
				"preserve-ip-false",
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerType:                    "nlb",
					cloudprovider.ServiceAnnotationNetworkLoadBalancerIsPreserveSource: "false",
					cloudprovider.ServiceAnnotationLoadBalancerInternal:                "true",
				},
				false,
			},
		}
		It("should be possible configure preservation of source IP in NLB", func() {
			for _, test := range preserveSourceTestArray {
				By("Running test for: " + test.configuration)
				serviceName := "e2e-" + test.configuration
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
					s.ObjectMeta.Annotations = test.annotations
					s.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
					s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationLoadBalancerInternal] = "true"
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
				loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), test.lbType, "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
				sharedfw.ExpectNoError(err)

				By("Validate isPreserveSource in the backend set is as expected")
				isPreserve := *loadBalancer.BackendSets["TCP-80"].IsPreserveSource
				Expect(isPreserve == test.isPreserveSource).To(BeTrue())

				isPreserve = *loadBalancer.BackendSets["TCP-443"].IsPreserveSource
				Expect(isPreserve == test.isPreserveSource).To(BeTrue())

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
	})
})

var _ = Describe("LB Properties", func() {
	baseName := "lb-properties"
	f := sharedfw.NewDefaultFramework(baseName)

	Context("[cloudprovider][ccm][lb][properties]", func() {

		healthCheckTestArray := []struct {
			lbType              string
			CreationAnnotations map[string]string
			UpdatedAnnotations  map[string]string
			RemovedAnnotations  map[string]string
			CreateInterval      int
			UpdateInterval      int
		}{
			{
				"lb",
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckRetries:  "1",
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckTimeout:  "1000",
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckInterval: "4000",
				},
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckRetries:  "2",
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckTimeout:  "2000",
					cloudprovider.ServiceAnnotationLoadBalancerHealthCheckInterval: "6000",
				},
				map[string]string{},
				4000,
				6000,
			},
			{
				"nlb",
				map[string]string{
					cloudprovider.ServiceAnnotationNetworkLoadBalancerHealthCheckRetries:  "1",
					cloudprovider.ServiceAnnotationNetworkLoadBalancerHealthCheckTimeout:  "1000",
					cloudprovider.ServiceAnnotationNetworkLoadBalancerHealthCheckInterval: "10000",
					cloudprovider.ServiceAnnotationLoadBalancerType:                       "nlb",
				},
				map[string]string{
					cloudprovider.ServiceAnnotationNetworkLoadBalancerHealthCheckRetries:  "2",
					cloudprovider.ServiceAnnotationNetworkLoadBalancerHealthCheckTimeout:  "2000",
					cloudprovider.ServiceAnnotationNetworkLoadBalancerHealthCheckInterval: "15000",
					cloudprovider.ServiceAnnotationLoadBalancerType:                       "nlb",
				},
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerType: "nlb",
				},
				10000,
				15000,
			},
		}
		It("should be possible to create Service type:LoadBalancer and mutate health-check config", func() {
			for _, test := range healthCheckTestArray {
				By("Running test for: " + test.lbType)
				serviceName := "e2e-" + test.lbType + "-healthcheck-config"
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
					s.ObjectMeta.Annotations = test.CreationAnnotations
					if test.lbType == "lb" {
						s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationLoadBalancerInternal] = "true"
					}
					if test.lbType == "nlb" {
						s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationNetworkLoadBalancerInternal] = "true"
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
				loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), test.lbType, "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
				sharedfw.ExpectNoError(err)
				err = f.VerifyHealthCheckConfig(*loadBalancer.Id, 1, 1000, test.CreateInterval, test.lbType)
				sharedfw.ExpectNoError(err)

				By("changing TCP service health check config")
				tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
					s.ObjectMeta.Annotations = test.UpdatedAnnotations
					if test.lbType == "lb" {
						s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationLoadBalancerInternal] = "true"
					}
					if test.lbType == "nlb" {
						s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationNetworkLoadBalancerInternal] = "true"
					}
				})

				By("waiting upto 5m0s to verify health check config after modification to initial")
				err = f.VerifyHealthCheckConfig(*loadBalancer.Id, 2, 2000, test.UpdateInterval, test.lbType)
				sharedfw.ExpectNoError(err)

				By("changing TCP service health check config - remove annotations")
				tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
					s.ObjectMeta.Annotations = test.RemovedAnnotations
					if test.lbType == "lb" {
						s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationLoadBalancerInternal] = "true"
					}
					if test.lbType == "nlb" {
						s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationNetworkLoadBalancerInternal] = "true"
					}
				})

				By("waiting upto 5m0s to verify health check config should fall back to default after removing annotations")
				err = f.VerifyHealthCheckConfig(*loadBalancer.Id, 3, 3000, 10000, test.lbType)
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
			}
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
					{
						"flexible",
						"50",
						"150",
					},
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
						cloudprovider.ServiceAnnotationLoadBalancerInternal:     "true",
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

				loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), "lb", "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
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
					cloudprovider.ServiceAnnotationLoadBalancerInternal:              "true",
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
			loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), "lb", "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
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

		nsgTestArray := []struct {
			lbtype        string
			Annotations   map[string]string
			nsgAnnotation string
		}{
			{
				"lb",
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerInternal:     "true",
					cloudprovider.ServiceAnnotationLoadBalancerShape:        "flexible",
					cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMin: "10",
					cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMax: "10",
				},
				cloudprovider.ServiceAnnotationLoadBalancerNetworkSecurityGroups,
			},
			{
				"nlb",
				map[string]string{
					cloudprovider.ServiceAnnotationNetworkLoadBalancerInternal: "true",
					cloudprovider.ServiceAnnotationLoadBalancerType:            "nlb",
				},
				cloudprovider.ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups,
			},
		}
		// Test NSG feature
		It("should be possible to create/update/delete Service type:LoadBalancer with NSGs config", func() {
			for _, test := range nsgTestArray {
				By("Running test for: " + test.lbtype)
				serviceName := "e2e-" + test.lbtype + "-nsg"
				ns := f.Namespace.Name

				jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

				loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
				if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
					loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
				}

				requestedIP := ""
				nsgList := strings.Split(strings.ReplaceAll(setupF.NsgOCIDS, " ", ""), ",")
				lbNSGTestArray := []struct {
					testName        string
					nsgIds          []string
					resultantNsgIds []string
				}{
					{
						"Update LB with new NsgIds provided in svc config",
						nsgList,
						nsgList,
					},
					{
						"Update LB with empty NsgIds provided in svc config",
						[]string{},
						[]string{},
					},
					{
						"Update LB when there are duplicate NSG OCIDS provided in svc config",
						[]string{nsgList[0], nsgList[1], nsgList[0]},
						[]string{nsgList[0], nsgList[1]},
					},
				}
				nsgIds := nsgList[0]
				sharedfw.Logf(nsgIds)
				test.Annotations[test.nsgAnnotation] = nsgIds
				tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
					s.Spec.Type = v1.ServiceTypeLoadBalancer
					s.Spec.LoadBalancerIP = requestedIP
					s.Spec.Ports = []v1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
						{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
					s.ObjectMeta.Annotations = test.Annotations
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

				By("waiting upto 5m0s to verify initial LB config")
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

				loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), test.lbtype, "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
				sharedfw.ExpectNoError(err)
				By("waiting upto 5m0s to verify whether LB has been created with provided initial NSGs through config")
				err = f.WaitForLoadBalancerNSGChange(loadBalancer, []string{nsgIds}, test.lbtype)
				sharedfw.ExpectNoError(err)

				for _, t := range lbNSGTestArray {
					By(t.testName)
					nsgIds = strings.Join(t.nsgIds, ",")
					test.Annotations[test.nsgAnnotation] = nsgIds
					tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
						s.ObjectMeta.Annotations = test.Annotations
						if test.lbtype == "lb" {
							s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationLoadBalancerInternal] = "true"
						}
						if test.lbtype == "nlb" {
							s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationNetworkLoadBalancerInternal] = "true"
						}
					})
					err = f.WaitForLoadBalancerNSGChange(loadBalancer, t.resultantNsgIds, test.lbtype)
					sharedfw.ExpectNoError(err)
				}

				By("changing TCP service back to type=ClusterIP")
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

		lbPolicyTestArray := []struct {
			lbType              string
			CreationAnnotations map[string]string
			UpdatedAnnotations  map[string]string
			PolicyAnnotation    string
		}{
			{
				"lb",
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerShape:        "flexible",
					cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMin: "10",
					cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMax: "10",
					cloudprovider.ServiceAnnotationLoadBalancerPolicy:       cloudprovider.IPHashLoadBalancerPolicy,
				},
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerPolicy: cloudprovider.LeastConnectionsLoadBalancerPolicy,
				},
				cloudprovider.ServiceAnnotationLoadBalancerPolicy,
			},
			{
				"nlb",
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerType:                 "nlb",
					cloudprovider.ServiceAnnotationNetworkLoadBalancerBackendPolicy: cloudprovider.NetworkLoadBalancingPolicyTwoTuple,
				},
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerType:                 "nlb",
					cloudprovider.ServiceAnnotationNetworkLoadBalancerBackendPolicy: cloudprovider.NetworkLoadBalancingPolicyThreeTuple,
				},
				cloudprovider.ServiceAnnotationNetworkLoadBalancerBackendPolicy,
			},
		}

		// Test creating loadBalancer with custom loadbalancer policy and updating the policy in existing loadbalancer
		It("should be possible to create a service type:LoadBalancer with custom loadbalancer policy and update the policy", func() {

			for _, test := range lbPolicyTestArray {
				By("Running test for: " + test.lbType)

				serviceName := "e2e-" + test.lbType + "-policy"
				ns := f.Namespace.Name

				jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

				loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
				if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
					loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
				}

				reservedIP := ""
				sharedfw.Logf(reservedIP)
				tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
					s.Spec.Type = v1.ServiceTypeLoadBalancer
					s.Spec.LoadBalancerIP = reservedIP
					s.Spec.Ports = []v1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
						{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
					s.ObjectMeta.Annotations = test.CreationAnnotations
					if test.lbType == "lb" {
						s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationLoadBalancerInternal] = "true"
					}
					if test.lbType == "nlb" {
						s.ObjectMeta.Annotations[cloudprovider.ServiceAnnotationNetworkLoadBalancerInternal] = "true"
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

				By("waiting upto 5m0s to verify initial LB config")
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

				loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), test.lbType, "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
				sharedfw.ExpectNoError(err)

				err = f.VerifyLoadBalancerPolicy(*loadBalancer.Id, test.CreationAnnotations[test.PolicyAnnotation], test.lbType)
				sharedfw.ExpectNoError(err)

				By("changing TCP service loadbalancer policy")
				tcpService = jig.UpdateServiceOrFail(ns, tcpService.Name, func(s *v1.Service) {
					s.ObjectMeta.Annotations = test.UpdatedAnnotations
				})

				By("waiting upto 5m0s to verify loadbalancer policy after modification")
				err = f.VerifyLoadBalancerPolicy(*loadBalancer.Id, test.UpdatedAnnotations[test.PolicyAnnotation], test.lbType)
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
			}
		})

		reservedIpTestArray := []struct {
			lbType              string
			CreationAnnotations map[string]string
		}{
			{
				"lb",
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerShape:        "flexible",
					cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMin: "10",
					cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMax: "10",
				},
			},
			{
				"nlb",
				map[string]string{
					cloudprovider.ServiceAnnotationLoadBalancerType: "nlb",
				},
			},
			{
				"lb-wris",
				map[string]string{
					cloudprovider.ServiceAnnotationServiceAccountName:       "sa",
					cloudprovider.ServiceAnnotationLoadBalancerShape:        "flexible",
					cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMin: "10",
					cloudprovider.ServiceAnnotationLoadBalancerShapeFlexMax: "10",
				},
			},
			{
				"nlb-wris",
				map[string]string{
					cloudprovider.ServiceAnnotationServiceAccountName: "sa",
					cloudprovider.ServiceAnnotationLoadBalancerType:   "nlb",
				},
			},
		}

		// Test Reserved IP feature
		It("should be possible to create Service type:LoadbBalancer with public reservedIP", func() {
			for _, test := range reservedIpTestArray {
				if strings.HasSuffix(test.lbType, "-wris") && f.ClusterType != containerengine.ClusterTypeEnhancedCluster {
					sharedfw.Logf("Skipping Workload Identity Principal test for LB Type (%s) because the cluster is not an OKE ENHANCED_CLUSTER", test.lbType)
					continue
				}
				By("Running test for: " + test.lbType)
				serviceName := "e2e-" + test.lbType + "-reserved-ip"
				ns := f.Namespace.Name

				jig := sharedfw.NewServiceTestJig(f.ClientSet, serviceName)

				loadBalancerCreateTimeout := sharedfw.LoadBalancerCreateTimeoutDefault
				if nodes := sharedfw.GetReadySchedulableNodesOrDie(f.ClientSet); len(nodes.Items) > sharedfw.LargeClusterMinNodesNumber {
					loadBalancerCreateTimeout = sharedfw.LoadBalancerCreateTimeoutLarge
				}

				reservedIP := setupF.ReservedIP
				sharedfw.Logf(reservedIP)

				tcpService := jig.CreateTCPServiceOrFail(ns, func(s *v1.Service) {
					s.Spec.Type = v1.ServiceTypeLoadBalancer
					s.Spec.LoadBalancerIP = reservedIP
					s.Spec.Ports = []v1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt(80)},
						{Name: "https", Port: 443, TargetPort: intstr.FromInt(80)}}
					s.ObjectMeta.Annotations = test.CreationAnnotations
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

				By("waiting upto 5m0s to verify initial LB config")
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

				loadBalancer, err := f.Client.LoadBalancer(zap.L().Sugar(), test.lbType, "", nil).GetLoadBalancerByName(ctx, compartmentId, lbName)
				sharedfw.ExpectNoError(err)
				By("waiting upto 5m0s to verify whether LB has been created with public reservedIP")

				reservedIPOCID, err := f.Client.Networking(nil).GetPublicIpByIpAddress(ctx, reservedIP)
				sharedfw.Logf("Loadbalancer reserved IP OCID is: %s  Expected reserved IP OCID: %s", *loadBalancer.IpAddresses[0].ReservedIp.Id, *reservedIPOCID.Id)
				Expect(strings.Compare(*loadBalancer.IpAddresses[0].ReservedIp.Id, *reservedIPOCID.Id) == 0).To(BeTrue())

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
	})
})

// ips is the list of private IPs of the nodes, the path is the endpoint at which health is checked,
// and nodeIndex is the node which has the current pod
func CreateHealthCheckScript(healthCheckNodePort int, ips []string, path string, nodeIndex int) string {
	script := ""

	for n, privateIP := range ips {
		port := strconv.Itoa(healthCheckNodePort)
		ipPort := net.JoinHostPort(privateIP, port)
		//command to get health status of the pod on the node
		script += "healthCheckPassed=$(curl -s http://" + ipPort + path + " | grep -i localEndpoints* | cut -d ':' -f2);"
		if n == nodeIndex {
			script += "if ((\"$healthCheckPassed\"==\"0\")); then exit 1; fi;"
		} else {
			script += "if ((\"$healthCheckPassed\"==\"1\")); then exit 1; fi;"
		}
	}

	sharedfw.Logf("Script used: %v", script)

	return script
}
