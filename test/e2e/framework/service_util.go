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

package framework

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gerrors "github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	imageutils "k8s.io/kubernetes/test/utils/image"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	cloudprovider "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
)

const (
	// KubeProxyLagTimeout is the maximum time a kube-proxy daemon on a node is allowed
	// to not notice a Service update, such as type=NodePort.
	// TODO: This timeout should be O(10s), observed values are O(1m), 5m is very
	// liberal. Fix tracked in #20567.
	KubeProxyLagTimeout = 5 * time.Minute

	// KubeProxyEndpointLagTimeout is the maximum time a kube-proxy daemon on a node is allowed
	// to not notice an Endpoint update.
	KubeProxyEndpointLagTimeout = 30 * time.Second

	// LoadBalancerLagTimeoutDefault is the maximum time a load balancer is allowed to
	// not respond after creation.
	LoadBalancerLagTimeoutDefault = 5 * time.Minute

	// LoadBalancerLagTimeoutAWS is the delay between ELB creation and serving traffic
	// on AWS. A few minutes is typical, so use 10m.
	LoadBalancerLagTimeoutAWS = 10 * time.Minute

	// How long to wait for a load balancer to be created/modified.
	//TODO: once support ticket 21807001 is resolved, reduce this timeout back to something reasonable
	LoadBalancerCreateTimeoutDefault = 20 * time.Minute
	LoadBalancerCreateTimeoutLarge   = 2 * time.Hour

	// Time required by the loadbalancer to cleanup, proportional to numApps/Ing.
	// Bring the cleanup timeout back down to 5m once b/33588344 is resolved.
	LoadBalancerCleanupTimeout = 15 * time.Minute

	// On average it takes ~6 minutes for a single backend to come online in GCE.
	LoadBalancerPollTimeout  = 15 * time.Minute
	LoadBalancerPollInterval = 30 * time.Second

	LargeClusterMinNodesNumber = 100

	// Don't test with more than 3 nodes.
	// Many tests create an endpoint per node, in large clusters, this is
	// resource and time intensive.
	MaxNodesForEndpointsTests = 3

	// ServiceTestTimeout is used for most polling/waiting activities
	ServiceTestTimeout = 60 * time.Second

	// GCPMaxInstancesInInstanceGroup is the maximum number of instances supported in
	// one instance group on GCP.
	GCPMaxInstancesInInstanceGroup = 2000

	// OCI LB Shape Update Timeout
	OCILBShapeUpdateTimeout = 5 * time.Minute

	// OCI LB NSG Update Timeout
	OCILBNSGUpdateTimeout = 5 * time.Minute
)

// This should match whatever the default/configured range is
var ServiceNodePortRange = utilnet.PortRange{Base: 30000, Size: 2768}

// A test jig to help service testing.
type ServiceTestJig struct {
	ID     string
	Name   string
	Client clientset.Interface
	Labels map[string]string
}

// NewServiceTestJig allocates and inits a new ServiceTestJig.
func NewServiceTestJig(client clientset.Interface, name string) *ServiceTestJig {
	j := &ServiceTestJig{}
	j.Client = client
	j.Name = name
	j.ID = j.Name + "-" + string(uuid.NewUUID())
	j.Labels = map[string]string{"testid": j.ID}

	return j
}

// newServiceTemplate returns the default v1.Service template for this jig, but
// does not actually create the Service.  The default Service has the same name
// as the jig and exposes the given port.
func (j *ServiceTestJig) newServiceTemplate(namespace string, proto v1.Protocol, port int32) *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      j.Name,
			Labels:    j.Labels,
		},
		Spec: v1.ServiceSpec{
			Selector: j.Labels,
			Ports: []v1.ServicePort{
				{
					Protocol: proto,
					Port:     port,
				},
			},
		},
	}
	return service
}

// CreateTCPServiceWithPort creates a new TCP Service with given port based on the
// jig's defaults. Callers can provide a function to tweak the Service object before
// it is created.
func (j *ServiceTestJig) CreateTCPServiceWithPort(namespace string, tweak func(svc *v1.Service), port int32) *v1.Service {
	svc := j.newServiceTemplate(namespace, v1.ProtocolTCP, port)
	if tweak != nil {
		tweak(svc)
	}
	result, err := j.Client.CoreV1().Services(namespace).Create(context.Background(), svc, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create TCP Service %q: %v", svc.Name, err)
	}
	return result
}

// CreateTCPServiceOrFail creates a new TCP Service based on the jig's
// defaults.  Callers can provide a function to tweak the Service object before
// it is created.
func (j *ServiceTestJig) CreateTCPServiceOrFail(namespace string, tweak func(svc *v1.Service)) *v1.Service {
	svc := j.newServiceTemplate(namespace, v1.ProtocolTCP, 80)
	if tweak != nil {
		tweak(svc)
	}
	result, err := j.Client.CoreV1().Services(namespace).Create(context.Background(), svc, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create TCP Service %q: %v", svc.Name, err)
	}
	return result
}

// CreateUDPServiceOrFail creates a new UDP Service based on the jig's
// defaults.  Callers can provide a function to tweak the Service object before
// it is created.
func (j *ServiceTestJig) CreateUDPServiceOrFail(namespace string, tweak func(svc *v1.Service)) *v1.Service {
	svc := j.newServiceTemplate(namespace, v1.ProtocolUDP, 80)
	if tweak != nil {
		tweak(svc)
	}
	result, err := j.Client.CoreV1().Services(namespace).Create(context.Background(), svc, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create UDP Service %q: %v", svc.Name, err)
	}
	return result
}

// CreateExternalNameServiceOrFail creates a new ExternalName type Service based on the jig's defaults.
// Callers can provide a function to tweak the Service object before it is created.
func (j *ServiceTestJig) CreateExternalNameServiceOrFail(namespace string, tweak func(svc *v1.Service)) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      j.Name,
			Labels:    j.Labels,
		},
		Spec: v1.ServiceSpec{
			Selector:     j.Labels,
			ExternalName: "foo.example.com",
			Type:         v1.ServiceTypeExternalName,
		},
	}
	if tweak != nil {
		tweak(svc)
	}
	result, err := j.Client.CoreV1().Services(namespace).Create(context.Background(), svc, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create ExternalName Service %q: %v", svc.Name, err)
	}
	return result
}

func (j *ServiceTestJig) ChangeServiceType(namespace, name string, newType v1.ServiceType, timeout time.Duration) {
	ingressIP := ""
	svc := j.UpdateServiceOrFail(namespace, name, func(s *v1.Service) {
		for _, ing := range s.Status.LoadBalancer.Ingress {
			if ing.IP != "" {
				ingressIP = ing.IP
			}
		}
		s.Spec.Type = newType
		s.Spec.Ports[0].NodePort = 0
	})
	if ingressIP != "" {
		j.WaitForLoadBalancerDestroyOrFail(namespace, svc.Name, ingressIP, int(svc.Spec.Ports[0].Port), timeout)
	}
}

// CreateOnlyLocalNodePortService creates a NodePort service with
// ExternalTrafficPolicy set to Local and sanity checks its nodePort.
// If createPod is true, it also creates an RC with 1 replica of
// the standard agnhost container used everywhere in this test.
func (j *ServiceTestJig) CreateOnlyLocalNodePortService(namespace, serviceName string, createPod bool) *v1.Service {
	By("creating a service " + namespace + "/" + serviceName + " with type=NodePort and ExternalTrafficPolicy=Local")
	svc := j.CreateTCPServiceOrFail(namespace, func(svc *v1.Service) {
		svc.Spec.Type = v1.ServiceTypeNodePort
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
		svc.Spec.Ports = []v1.ServicePort{{Protocol: "TCP", Port: 80}}
	})

	if createPod {
		By("creating a pod to be part of the service " + serviceName)
		j.RunOrFail(namespace, nil)
	}
	j.SanityCheckService(svc, v1.ServiceTypeNodePort)
	return svc
}

// CreateOnlyLocalLoadBalancerService creates a loadbalancer service with
// ExternalTrafficPolicy set to Local and waits for it to acquire an ingress IP.
// If createPod is true, it also creates an RC with 1 replica of
// the standard agnhost container used everywhere in this test.
func (j *ServiceTestJig) CreateOnlyLocalLoadBalancerService(namespace, serviceName string, timeout time.Duration, createPod bool, creationAnnotations map[string]string,
	tweak func(svc *v1.Service)) *v1.Service {
	By("creating a service " + namespace + "/" + serviceName + " with type=LoadBalancer and ExternalTrafficPolicy=Local")
	svc := j.CreateTCPServiceOrFail(namespace, func(svc *v1.Service) {
		svc.Spec.Type = v1.ServiceTypeLoadBalancer
		// We need to turn affinity off for our LB distribution tests
		svc.Spec.SessionAffinity = v1.ServiceAffinityNone
		svc.Spec.ExternalTrafficPolicy = v1.ServiceExternalTrafficPolicyTypeLocal
		if tweak != nil {
			tweak(svc)
		}
		svc.ObjectMeta.Annotations = creationAnnotations
	})

	if createPod {
		By("creating a pod to be part of the service " + serviceName)
		j.RunOrFail(namespace, nil)
	}
	By("waiting for loadbalancer for service " + namespace + "/" + serviceName)
	svc = j.WaitForLoadBalancerOrFail(namespace, serviceName, timeout)
	j.SanityCheckService(svc, v1.ServiceTypeLoadBalancer)
	return svc
}

// CreateLoadBalancerService creates a loadbalancer service and waits
// for it to acquire an ingress IP.
func (j *ServiceTestJig) CreateLoadBalancerService(namespace, serviceName string, timeout time.Duration, tweak func(svc *v1.Service)) *v1.Service {
	By("creating a service " + namespace + "/" + serviceName + " with type=LoadBalancer")
	svc := j.CreateTCPServiceOrFail(namespace, func(svc *v1.Service) {
		svc.Spec.Type = v1.ServiceTypeLoadBalancer
		// We need to turn affinity off for our LB distribution tests
		svc.Spec.SessionAffinity = v1.ServiceAffinityNone
		if tweak != nil {
			tweak(svc)
		}
	})

	By("waiting for loadbalancer for service " + namespace + "/" + serviceName)
	svc = j.WaitForLoadBalancerOrFail(namespace, serviceName, timeout)
	j.SanityCheckService(svc, v1.ServiceTypeLoadBalancer)
	return svc
}

func GetNodeAddresses(node *v1.Node, addressType v1.NodeAddressType) (ips []string) {
	for j := range node.Status.Addresses {
		nodeAddress := &node.Status.Addresses[j]
		if nodeAddress.Type == addressType {
			ips = append(ips, nodeAddress.Address)
		}
	}
	return
}

func CollectAddresses(nodes *v1.NodeList, addressType v1.NodeAddressType) []string {
	ips := []string{}
	for i := range nodes.Items {
		ips = append(ips, GetNodeAddresses(&nodes.Items[i], addressType)...)
	}
	return ips
}

func GetNodePublicIps(c clientset.Interface) ([]string, error) {
	nodes := GetReadySchedulableNodesOrDie(c)

	ips := CollectAddresses(nodes, v1.NodeExternalIP)
	if len(ips) == 0 {
		// If ExternalIP isn't set, assume the test programs can reach the InternalIP
		ips = CollectAddresses(nodes, v1.NodeInternalIP)
	}
	return ips, nil
}

func PickNodeIP(c clientset.Interface) string {
	publicIps, err := GetNodePublicIps(c)
	Expect(err).NotTo(HaveOccurred())
	if len(publicIps) == 0 {
		Failf("got unexpected number (%d) of public IPs", len(publicIps))
	}
	ip := publicIps[0]
	return ip
}

// GetEndpointNodes returns a map of nodenames:external-ip on which the
// endpoints of the given Service are running.
func (j *ServiceTestJig) GetEndpointNodes(svc *v1.Service) map[string][]string {
	nodes := j.GetNodes(MaxNodesForEndpointsTests)
	endpoints, err := j.Client.CoreV1().Endpoints(svc.Namespace).Get(context.Background(), svc.Name, metav1.GetOptions{})
	if err != nil {
		Failf("Get endpoints for service %s/%s failed (%s)", svc.Namespace, svc.Name, err)
	}
	if len(endpoints.Subsets) == 0 {
		Failf("Endpoint has no subsets, cannot determine node addresses.")
	}
	epNodes := sets.NewString()
	for _, ss := range endpoints.Subsets {
		for _, e := range ss.Addresses {
			if e.NodeName != nil {
				epNodes.Insert(*e.NodeName)
			}
		}
	}
	nodeMap := map[string][]string{}
	for _, n := range nodes.Items {
		if epNodes.Has(n.Name) {
			nodeMap[n.Name] = GetNodeAddresses(&n, v1.NodeExternalIP)
		}
	}
	return nodeMap
}

// GetNodes returns the first maxNodesForTest nodes. Useful in large clusters
// where we don't eg: want to create an endpoint per node.
func (j *ServiceTestJig) GetNodes(maxNodesForTest int) (nodes *v1.NodeList) {
	nodes = GetReadySchedulableNodesOrDie(j.Client)
	if len(nodes.Items) <= maxNodesForTest {
		maxNodesForTest = len(nodes.Items)
	}
	nodes.Items = nodes.Items[:maxNodesForTest]
	return nodes
}

func (j *ServiceTestJig) WaitForEndpointOnNode(namespace, serviceName, nodeName string) {
	err := wait.PollImmediate(K8sResourcePoll, LoadBalancerCreateTimeoutDefault, func() (bool, error) {
		endpoints, err := j.Client.CoreV1().Endpoints(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
		if err != nil {
			Logf("Get endpoints for service %s/%s failed (%s)", namespace, serviceName, err)
			return false, nil
		}
		if len(endpoints.Subsets) == 0 {
			Logf("Expect endpoints with subsets, got none.")
			return false, nil
		}
		// TODO: Handle multiple endpoints
		if len(endpoints.Subsets[0].Addresses) == 0 {
			Logf("Expected Ready endpoints - found none")
			return false, nil
		}
		epHostName := *endpoints.Subsets[0].Addresses[0].NodeName
		Logf("Pod for service %s/%s is on node %s", namespace, serviceName, epHostName)
		if epHostName != nodeName {
			Logf("Found endpoint on wrong node, expected %v, got %v", nodeName, epHostName)
			return false, nil
		}
		return true, nil
	})
	ExpectNoError(err)
}

func (j *ServiceTestJig) SanityCheckService(svc *v1.Service, svcType v1.ServiceType) {
	if svc.Spec.Type != svcType {
		Failf("unexpected Spec.Type (%s) for service, expected %s", svc.Spec.Type, svcType)
	}

	if svcType != v1.ServiceTypeExternalName {
		if svc.Spec.ExternalName != "" {
			Failf("unexpected Spec.ExternalName (%s) for service, expected empty", svc.Spec.ExternalName)
		}
		if svc.Spec.ClusterIP != v1.ClusterIPNone && svc.Spec.ClusterIP == "" {
			Failf("didn't get ClusterIP for non-ExternamName service")
		}
	} else {
		if svc.Spec.ClusterIP != "" {
			Failf("unexpected Spec.ClusterIP (%s) for ExternamName service, expected empty", svc.Spec.ClusterIP)
		}
	}

	expectNodePorts := false
	if svcType != v1.ServiceTypeClusterIP && svcType != v1.ServiceTypeExternalName {
		expectNodePorts = true
	}
	for i, port := range svc.Spec.Ports {
		hasNodePort := (port.NodePort != 0)
		if hasNodePort != expectNodePorts {
			Failf("unexpected Spec.Ports[%d].NodePort (%d) for service", i, port.NodePort)
		}
		if hasNodePort {
			if !ServiceNodePortRange.Contains(int(port.NodePort)) {
				Failf("out-of-range nodePort (%d) for service", port.NodePort)
			}
		}
	}
	expectIngress := false
	if svcType == v1.ServiceTypeLoadBalancer {
		expectIngress = true
	}
	hasIngress := len(svc.Status.LoadBalancer.Ingress) != 0
	if hasIngress != expectIngress {
		Failf("unexpected number of Status.LoadBalancer.Ingress (%d) for service", len(svc.Status.LoadBalancer.Ingress))
	}
	if hasIngress {
		for i, ing := range svc.Status.LoadBalancer.Ingress {
			if ing.IP == "" && ing.Hostname == "" {
				Failf("unexpected Status.LoadBalancer.Ingress[%d] for service: %#v", i, ing)
			}
		}
	}
}

// UpdateService fetches a service, calls the update function on it, and
// then attempts to send the updated service. It tries up to 3 times in the
// face of timeouts and conflicts.
func (j *ServiceTestJig) UpdateService(namespace, name string, update func(*v1.Service)) (*v1.Service, error) {
	for i := 0; i < 3; i++ {
		service, err := j.Client.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to get Service %q: %v", name, err)
		}
		update(service)
		service, err = j.Client.CoreV1().Services(namespace).Update(context.Background(), service, metav1.UpdateOptions{})
		if err == nil {
			return service, nil
		}
		if !errors.IsConflict(err) && !errors.IsServerTimeout(err) {
			return nil, fmt.Errorf("Failed to update Service %q: %v", name, err)
		}
	}
	return nil, fmt.Errorf("Too many retries updating Service %q", name)
}

// UpdateServiceOrFail fetches a service, calls the update function on it, and
// then attempts to send the updated service. It tries up to 3 times in the
// face of timeouts and conflicts.
func (j *ServiceTestJig) UpdateServiceOrFail(namespace, name string, update func(*v1.Service)) *v1.Service {
	svc, err := j.UpdateService(namespace, name, update)
	if err != nil {
		Failf(err.Error())
	}
	return svc
}

func (j *ServiceTestJig) WaitForNewIngressIPOrFail(namespace, name, existingIP string, timeout time.Duration) *v1.Service {
	Logf("Waiting up to %v for service %q to get a new ingress IP", timeout, name)
	service := j.waitForConditionOrFail(namespace, name, timeout, "have a new ingress IP", func(svc *v1.Service) bool {
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return false
		}
		ip := svc.Status.LoadBalancer.Ingress[0].IP
		if ip == "" || ip == existingIP {
			return false
		}
		return true
	})
	return service
}

func (j *ServiceTestJig) ChangeServiceNodePortOrFail(namespace, name string, initial int) *v1.Service {
	var err error
	var service *v1.Service
	for i := 1; i < ServiceNodePortRange.Size; i++ {
		offs1 := initial - ServiceNodePortRange.Base
		offs2 := (offs1 + i) % ServiceNodePortRange.Size
		newPort := ServiceNodePortRange.Base + offs2
		service, err = j.UpdateService(namespace, name, func(s *v1.Service) {
			s.Spec.Ports[0].NodePort = int32(newPort)
		})
		if err != nil && strings.Contains(err.Error(), "provided port is already allocated") {
			Logf("tried nodePort %d, but it is in use, will try another", newPort)
			continue
		}
		// Otherwise err was nil or err was a real error
		break
	}
	if err != nil {
		Failf("Could not change the nodePort: %v", err)
	}
	return service
}

func (j *ServiceTestJig) WaitForLoadBalancerOrFail(namespace, name string, timeout time.Duration) *v1.Service {
	Logf("Waiting up to %v for service %q to have a LoadBalancer", timeout, name)
	service := j.waitForConditionOrFail(namespace, name, timeout, "have a load balancer", func(svc *v1.Service) bool {
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			return true
		}
		return false
	})
	return service
}

// EnsureLoadBalancerResourcesDeleted ensures that cloud load balancer resources
// that were created are actually cleaned up.
func EnsureLoadBalancerResourcesDeleted(ip, portRange string) error {
	// TODO: Implement check for OCI.
	return nil
}

func (j *ServiceTestJig) WaitForLoadBalancerDestroyOrFail(namespace, name string, ip string, port int, timeout time.Duration) *v1.Service {
	// TODO: once support ticket 21807001 is resolved, reduce this timeout back to something reasonable
	defer func() {
		if err := EnsureLoadBalancerResourcesDeleted(ip, strconv.Itoa(port)); err != nil {
			Logf("Failed to delete cloud resources for service: %s %d (%v)", ip, port, err)
		}
	}()

	Logf("Waiting up to %v for service %q to have no LoadBalancer", timeout, name)
	service := j.waitForConditionOrFail(namespace, name, timeout, "have no load balancer", func(svc *v1.Service) bool {
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			return true
		}
		return false
	})
	return service
}

func (j *ServiceTestJig) waitForConditionOrFail(namespace, name string, timeout time.Duration, message string, conditionFn func(*v1.Service) bool) *v1.Service {
	var service *v1.Service
	pollFunc := func() (bool, error) {
		svc, err := j.Client.CoreV1().Services(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if conditionFn(svc) {
			service = svc
			return true, nil
		}
		return false, nil
	}
	if err := wait.PollImmediate(K8sResourcePoll, timeout, pollFunc); err != nil {
		Failf("Timed out waiting for service %q to %s", name, message)
	}
	return service
}

// newRCTemplate returns the default v1.ReplicationController object for
// this jig, but does not actually create the RC.  The default RC has the same
// name as the jig and runs the "agnhost" container.
func (j *ServiceTestJig) newRCTemplate(namespace string) *v1.ReplicationController {
	var replicas int32 = 1

	rc := &v1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      j.Name,
			Labels:    j.Labels,
		},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &replicas,
			Selector: j.Labels,
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: j.Labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "agnhost",
							Image: agnhost,
							Args:  []string{"netexec", "--http-port=80", "--udp-port=80"},
							ReadinessProbe: &v1.Probe{
								PeriodSeconds: 3,
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Port: intstr.FromInt(80),
										Path: "/hostName",
									},
								},
							},
						},
					},
					TerminationGracePeriodSeconds: new(int64),
				},
			},
		},
	}
	return rc
}

func (j *ServiceTestJig) AddRCAntiAffinity(rc *v1.ReplicationController) {
	var replicas int32 = 2

	rc.Spec.Replicas = &replicas
	if rc.Spec.Template.Spec.Affinity == nil {
		rc.Spec.Template.Spec.Affinity = &v1.Affinity{}
	}
	if rc.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
		rc.Spec.Template.Spec.Affinity.PodAntiAffinity = &v1.PodAntiAffinity{}
	}
	rc.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
		rc.Spec.Template.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
		v1.PodAffinityTerm{
			LabelSelector: &metav1.LabelSelector{MatchLabels: j.Labels},
			Namespaces:    nil,
			TopologyKey:   "kubernetes.io/hostname",
		})
}

func (j *ServiceTestJig) CreatePDBOrFail(namespace string, rc *v1.ReplicationController) *policyv1beta1.PodDisruptionBudget {
	pdb := j.newPDBTemplate(namespace, rc)
	newPdb, err := j.Client.PolicyV1beta1().PodDisruptionBudgets(namespace).Create(context.Background(), pdb, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create PDB %q %v", pdb.Name, err)
	}
	if err := j.waitForPdbReady(namespace); err != nil {
		Failf("Failed waiting for PDB to be ready: %v", err)
	}

	return newPdb
}

// newPDBTemplate returns the default policyv1beta1.PodDisruptionBudget object for
// this jig, but does not actually create the PDB.  The default PDB specifies a
// MinAvailable of N-1 and matches the pods created by the RC.
func (j *ServiceTestJig) newPDBTemplate(namespace string, rc *v1.ReplicationController) *policyv1beta1.PodDisruptionBudget {
	minAvailable := intstr.FromInt(int(*rc.Spec.Replicas) - 1)

	pdb := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      j.Name,
			Labels:    j.Labels,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector:     &metav1.LabelSelector{MatchLabels: j.Labels},
		},
	}

	return pdb
}

// RunOrFail creates a ReplicationController and Pod(s) and waits for the
// Pod(s) to be running. Callers can provide a function to tweak the RC object
// before it is created.
func (j *ServiceTestJig) RunOrFail(namespace string, tweak func(rc *v1.ReplicationController)) *v1.ReplicationController {
	rc := j.newRCTemplate(namespace)
	if tweak != nil {
		tweak(rc)
	}
	result, err := j.Client.CoreV1().ReplicationControllers(namespace).Create(context.Background(), rc, metav1.CreateOptions{})
	if err != nil {
		Failf("Failed to create RC %q: %v", rc.Name, err)
	}
	pods, err := j.waitForPodsCreated(namespace, int(*(rc.Spec.Replicas)))
	if err != nil {
		Failf("Failed to create pods: %v", err)
	}
	if err := j.waitForPodsReady(namespace, pods); err != nil {
		Failf("Failed waiting for pods to be running: %v", err)
	}
	return result
}

func (j *ServiceTestJig) waitForPdbReady(namespace string) error {
	timeout := 2 * time.Minute
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(2 * time.Second) {
		pdb, err := j.Client.PolicyV1beta1().PodDisruptionBudgets(namespace).Get(context.Background(), j.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if pdb.Status.DisruptionsAllowed > 0 {
			return nil
		}
	}

	return fmt.Errorf("Timeout waiting for PDB %q to be ready", j.Name)
}

func (j *ServiceTestJig) waitForPodsCreated(namespace string, replicas int) ([]string, error) {
	timeout := 5 * time.Minute
	// List the pods, making sure we observe all the replicas.
	label := labels.SelectorFromSet(labels.Set(j.Labels))
	Logf("Waiting up to %v for %d pods to be created", timeout, replicas)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(2 * time.Second) {
		options := metav1.ListOptions{LabelSelector: label.String()}
		pods, err := j.Client.CoreV1().Pods(namespace).List(context.Background(), options)
		if err != nil {
			return nil, err
		}

		found := []string{}
		for _, pod := range pods.Items {
			if pod.DeletionTimestamp != nil {
				continue
			}
			found = append(found, pod.Name)
		}
		if len(found) == replicas {
			Logf("Found all %d pods", replicas)
			return found, nil
		}
		Logf("Found %d/%d pods - will retry", len(found), replicas)
	}
	return nil, fmt.Errorf("Timeout waiting for %d pods to be created", replicas)
}

func (j *ServiceTestJig) waitForPodsReady(namespace string, pods []string) error {
	timeout := 2 * time.Minute
	if !CheckPodsRunningReady(j.Client, namespace, pods, timeout) {
		return fmt.Errorf("Timeout waiting for %d pods to be ready", len(pods))
	}
	return nil
}

func (j *ServiceTestJig) TestReachableHTTP(secure bool, host string, port int, timeout time.Duration) {
	j.TestReachableHTTPWithRetriableErrorCodes(secure, host, port, []int{502}, timeout)
}

func (j *ServiceTestJig) TestReachableHTTPWithRetriableErrorCodes(secure bool, host string, port int, retriableErrCodes []int, timeout time.Duration) {
	if err := wait.PollImmediate(K8sResourcePoll, timeout, func() (bool, error) {
		return TestReachableHTTPWithRetriableErrorCodes(secure, host, port, "/echo?msg=hello", "hello", retriableErrCodes)
	}); err != nil {
		if err == wait.ErrWaitTimeout {
			Failf("Could not reach HTTP service through %v:%v after %v", host, port, timeout)
		} else {
			Failf("Failed to reach HTTP service through %v:%v: %v", host, port, err)
		}
	}
}

func (j *ServiceTestJig) TestNotReachableHTTP(host string, port int, timeout time.Duration) {
	if err := wait.PollImmediate(K8sResourcePoll, timeout, func() (bool, error) { return TestNotReachableHTTP(host, port) }); err != nil {
		Failf("Could still reach HTTP service through %v:%v after %v: %v", host, port, timeout, err)
	}
}

func (j *ServiceTestJig) TestReachableUDP(host string, port int, timeout time.Duration) {
	if err := wait.PollImmediate(K8sResourcePoll, timeout, func() (bool, error) { return TestReachableUDP(host, port, "echo hello", "hello") }); err != nil {
		Failf("Could not reach UDP service through %v:%v after %v: %v", host, port, timeout, err)
	}
}

func (j *ServiceTestJig) TestNotReachableUDP(host string, port int, timeout time.Duration) {
	if err := wait.PollImmediate(K8sResourcePoll, timeout, func() (bool, error) { return TestNotReachableUDP(host, port, "echo hello") }); err != nil {
		Failf("Could still reach UDP service through %v:%v after %v: %v", host, port, timeout, err)
	}
}

func (j *ServiceTestJig) GetHTTPContent(secure bool, host string, port int, timeout time.Duration, url string) bytes.Buffer {
	var body bytes.Buffer
	var err error
	if pollErr := wait.PollImmediate(K8sResourcePoll, timeout, func() (bool, error) {
		var result bool
		result, err = TestReachableHTTPWithContent(secure, host, port, url, "", &body)
		if err != nil {
			Logf("Error hitting %v:%v%v, retrying: %v", host, port, url, err)
			return false, nil
		}
		return result, nil
	}); pollErr != nil {
		Failf("Could not reach HTTP service through %v:%v%v after %v: %v", host, port, url, timeout, err)
	}
	return body
}

func testHTTPHealthCheckNodePort(ip string, port int, request string) (bool, error) {
	ipPort := net.JoinHostPort(ip, strconv.Itoa(port))
	url := fmt.Sprintf("http://%s%s", ipPort, request)
	if ip == "" || port == 0 {
		Failf("Got empty IP for reachability check (%s)", url)
		return false, fmt.Errorf("Invalid input ip or port")
	}
	Logf("Testing HTTP health check on %v", url)
	resp, err := httpGetNoConnectionPool(url)
	if err != nil {
		Logf("Got error testing for reachability of %s: %v", url, err)
		return false, err
	}
	defer resp.Body.Close()
	if err != nil {
		Logf("Got error reading response from %s: %v", url, err)
		return false, err
	}
	// HealthCheck responder returns 503 for no local endpoints
	if resp.StatusCode == 503 {
		return false, nil
	}
	// HealthCheck responder returns 200 for non-zero local endpoints
	if resp.StatusCode == 200 {
		return true, nil
	}
	return false, fmt.Errorf("Unexpected HTTP response code %s from health check responder at %s", resp.Status, url)
}

func (j *ServiceTestJig) TestHTTPHealthCheckNodePort(host string, port int, request string, timeout time.Duration, expectSucceed bool, threshold int) error {
	count := 0
	condition := func() (bool, error) {
		success, _ := testHTTPHealthCheckNodePort(host, port, request)
		if success && expectSucceed ||
			!success && !expectSucceed {
			count++
		}
		if count >= threshold {
			return true, nil
		}
		return false, nil
	}

	if err := wait.PollImmediate(time.Second, timeout, condition); err != nil {
		return fmt.Errorf("error waiting for healthCheckNodePort: expected at least %d succeed=%v on %v%v, got %d", threshold, expectSucceed, host, port, count)
	}
	return nil
}

func (f *CloudProviderFramework) VerifyHealthCheckConfig(loadBalancerId string, retries, timeout, interval int, lbtype string) error {
	for start := time.Now(); time.Since(start) < 5*time.Minute; time.Sleep(5 * time.Second) {
		loadBalancer, err := f.Client.LoadBalancer(lbtype).GetLoadBalancer(context.TODO(), loadBalancerId)
		if err != nil {
			return err
		}
		success, err := testHealthCheckConfig(loadBalancer, retries, timeout, interval)
		if err != nil {
			return err
		}
		if success {
			Logf("Health Check config matches expected config.")
			return nil
		}
		Logf("Health Check config did not match expected - will retry")
	}
	return gerrors.Errorf("Timeout waiting for Health check config to be as expected.")
}

// WaitForLoadBalancerNSGChange polls for validating the associated NSGs
// to be the same as the spec
func (f *CloudProviderFramework) WaitForLoadBalancerNSGChange(lb *client.GenericLoadBalancer, nsgIds []string, lbtype string) error {
	condition := func() (bool, error) {
		updatedLB, err := f.Client.LoadBalancer(lbtype).GetLoadBalancer(context.TODO(), *lb.Id)
		if err != nil {
			return false, err
		}
		if !cloudprovider.DeepEqualLists(updatedLB.NetworkSecurityGroupIds, nsgIds) {
			return false, nil
		}
		return true, nil
	}
	if err := wait.Poll(15*time.Second, OCILBNSGUpdateTimeout, condition); err != nil {
		return fmt.Errorf("Failed to update LB NSGs, error: %s", err.Error())
	}
	return nil
}

// WaitForLoadBalancerShapeChange polls for the shape of the LB
// to be the same as the spec
func (f *CloudProviderFramework) WaitForLoadBalancerShapeChange(lb *client.GenericLoadBalancer, shape, fMin, fMax string) error {
	condition := func() (bool, error) {

		updatedLB, err := f.Client.LoadBalancer("lb").GetLoadBalancer(context.TODO(), *lb.Id)
		if err != nil {
			return false, err
		}
		if *updatedLB.ShapeName != shape {
			return false, nil
		}
		// in case of non-flex shapes ShapeDetails will be nil
		if updatedLB.ShapeDetails.MinimumBandwidthInMbps == nil &&
			updatedLB.ShapeDetails.MaximumBandwidthInMbps == nil {
			return true, nil
		}
		if strconv.Itoa(*updatedLB.ShapeDetails.MinimumBandwidthInMbps) != fMin ||
			strconv.Itoa(*updatedLB.ShapeDetails.MaximumBandwidthInMbps) != fMax {
			return false, nil
		}
		return true, nil
	}
	if err := wait.Poll(15*time.Second, OCILBShapeUpdateTimeout, condition); err != nil {
		return fmt.Errorf("Failed to update LB Shape from %s to %s. ShapeDetails (min, max) (%s, %s)", *lb.ShapeName, shape, fMin, fMax)
	}
	return nil
}
func testHealthCheckConfig(loadBalancer *client.GenericLoadBalancer, retries int, timeout int, interval int) (bool, error) {
	if loadBalancer != nil && len(loadBalancer.BackendSets) != 0 {
		for _, backendSet := range loadBalancer.BackendSets {
			if *backendSet.HealthChecker.Retries != retries {
				return false, nil
			}
			if *backendSet.HealthChecker.TimeoutInMillis != timeout {
				return false, nil
			}
			if *backendSet.HealthChecker.IntervalInMillis != interval {
				return false, nil
			}
		}
		return true, nil
	}
	return false, gerrors.Errorf("Could not find Health check config.")
}

func (f *CloudProviderFramework) VerifyLoadBalancerConnectionIdleTimeout(loadBalancerId string, connectionIdleTimeout int) error {
	for start := time.Now(); time.Since(start) < 5*time.Minute; time.Sleep(2 * time.Second) {
		loadBalancer, err := f.Client.LoadBalancer("lb").GetLoadBalancer(context.TODO(), loadBalancerId)
		if err != nil {
			return err
		}
		success, err := testConnectionIdleTimeout(loadBalancer, connectionIdleTimeout)
		if err != nil {
			return err
		}
		if success {
			Logf("Connection Idle Timeout matches expected value")
			return nil
		}
		Logf("Connection Idle Timeout did not match expected - will retry")
	}
	return gerrors.Errorf("Timeout waiting for Connection Idle Timeout to be as expected.")
}

func testConnectionIdleTimeout(loadBalancer *client.GenericLoadBalancer, connectionIdleTimeout int) (bool, error) {
	if loadBalancer != nil && len(loadBalancer.Listeners) != 0 {
		for _, listener := range loadBalancer.Listeners {
			if listener.ConnectionConfiguration != nil &&
				*listener.ConnectionConfiguration.IdleTimeout != int64(connectionIdleTimeout) {
				return false, nil
			}
		}
		return true, nil
	}
	return false, gerrors.Errorf("Could not find connection configuration.")
}

// Simple helper class to avoid too much boilerplate in tests
type ServiceTestFixture struct {
	ServiceName string
	Namespace   string
	Client      clientset.Interface

	TestId string
	Labels map[string]string

	rcs      map[string]bool
	services map[string]bool
	Name     string
	Image    string
}

func NewServerTest(client clientset.Interface, namespace string, serviceName string) *ServiceTestFixture {
	t := &ServiceTestFixture{}
	t.Client = client
	t.Namespace = namespace
	t.ServiceName = serviceName
	t.TestId = t.ServiceName + "-" + string(uuid.NewUUID())
	t.Labels = map[string]string{
		"testid": t.TestId,
	}

	t.rcs = make(map[string]bool)
	t.services = make(map[string]bool)

	t.Name = "webserver"
	t.Image = imageutils.GetE2EImage(imageutils.Agnhost)

	return t
}

// Build default config for a service (which can then be changed)
func (t *ServiceTestFixture) BuildServiceSpec() *v1.Service {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.ServiceName,
			Namespace: t.Namespace,
		},
		Spec: v1.ServiceSpec{
			Selector: t.Labels,
			Ports: []v1.ServicePort{{
				Port:       80,
				TargetPort: intstr.FromInt(80),
			}},
		},
	}
	return service
}

// CreateRC creates a replication controller and records it for cleanup.
func (t *ServiceTestFixture) CreateRC(rc *v1.ReplicationController) (*v1.ReplicationController, error) {
	rc, err := t.Client.CoreV1().ReplicationControllers(t.Namespace).Create(context.Background(), rc, metav1.CreateOptions{})
	if err == nil {
		t.rcs[rc.Name] = true
	}
	return rc, err
}

// Create a service, and record it for cleanup
func (t *ServiceTestFixture) CreateService(service *v1.Service) (*v1.Service, error) {
	result, err := t.Client.CoreV1().Services(t.Namespace).Create(context.Background(), service, metav1.CreateOptions{})
	if err == nil {
		t.services[service.Name] = true
	}
	return result, err
}

// Delete a service, and remove it from the cleanup list
func (t *ServiceTestFixture) DeleteService(serviceName string) error {
	err := t.Client.CoreV1().Services(t.Namespace).Delete(context.Background(), serviceName, metav1.DeleteOptions{})
	if err == nil {
		delete(t.services, serviceName)
	}
	return err
}

func (t *ServiceTestFixture) Cleanup() []error {
	var errs []error
	for rcName := range t.rcs {
		By("stopping RC " + rcName + " in namespace " + t.Namespace)
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// First, resize the RC to 0.
			old, err := t.Client.CoreV1().ReplicationControllers(t.Namespace).Get(context.Background(), rcName, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) {
					return nil
				}
				return err
			}
			x := int32(0)
			old.Spec.Replicas = &x
			if _, err := t.Client.CoreV1().ReplicationControllers(t.Namespace).Update(context.Background(), old, metav1.UpdateOptions{}); err != nil {
				if errors.IsNotFound(err) {
					return nil
				}
				return err
			}
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
		// TODO(mikedanese): Wait.
		// Then, delete the RC altogether.
		if err := t.Client.CoreV1().ReplicationControllers(t.Namespace).Delete(context.Background(), rcName, metav1.DeleteOptions{}); err != nil {
			if !errors.IsNotFound(err) {
				errs = append(errs, err)
			}
		}
	}

	for serviceName := range t.services {
		By("deleting service " + serviceName + " in namespace " + t.Namespace)
		err := t.Client.CoreV1().Services(t.Namespace).Delete(context.Background(), serviceName, metav1.DeleteOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func GetIngressPoint(ing *v1.LoadBalancerIngress) string {
	host := ing.IP
	if host == "" {
		host = ing.Hostname
	}
	return host
}

// UpdateService fetches a service, calls the update function on it,
// and then attempts to send the updated service. It retries up to 2
// times in the face of timeouts and conflicts.
func UpdateService(c clientset.Interface, namespace, serviceName string, update func(*v1.Service)) (*v1.Service, error) {
	var service *v1.Service
	var err error
	for i := 0; i < 3; i++ {
		service, err = c.CoreV1().Services(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
		if err != nil {
			return service, err
		}

		update(service)

		service, err = c.CoreV1().Services(namespace).Update(context.Background(), service, metav1.UpdateOptions{})

		if !errors.IsConflict(err) && !errors.IsServerTimeout(err) {
			return service, err
		}
	}
	return service, err
}

func GetContainerPortsByPodUID(endpoints *v1.Endpoints) PortsByPodUID {
	m := PortsByPodUID{}
	for _, ss := range endpoints.Subsets {
		for _, port := range ss.Ports {
			for _, addr := range ss.Addresses {
				containerPort := port.Port
				if _, ok := m[addr.TargetRef.UID]; !ok {
					m[addr.TargetRef.UID] = make([]int, 0)
				}
				m[addr.TargetRef.UID] = append(m[addr.TargetRef.UID], int(containerPort))
			}
		}
	}
	return m
}

type PortsByPodName map[string][]int
type PortsByPodUID map[types.UID][]int

func translatePodNameToUIDOrFail(c clientset.Interface, ns string, expectedEndpoints PortsByPodName) PortsByPodUID {
	portsByUID := make(PortsByPodUID)

	for name, portList := range expectedEndpoints {
		pod, err := c.CoreV1().Pods(ns).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			Failf("failed to get pod %s, that's pretty weird. validation failed: %s", name, err)
		}
		portsByUID[pod.ObjectMeta.UID] = portList
	}
	// Logf("successfully translated pod names to UIDs: %v -> %v on namespace %s", expectedEndpoints, portsByUID, ns)
	return portsByUID
}

func validatePortsOrFail(endpoints PortsByPodUID, expectedEndpoints PortsByPodUID) {
	if len(endpoints) != len(expectedEndpoints) {
		// should not happen because we check this condition before
		Failf("invalid number of endpoints got %v, expected %v", endpoints, expectedEndpoints)
	}
	for podUID := range expectedEndpoints {
		if _, ok := endpoints[podUID]; !ok {
			Failf("endpoint %v not found", podUID)
		}
		if len(endpoints[podUID]) != len(expectedEndpoints[podUID]) {
			Failf("invalid list of ports for uid %v. Got %v, expected %v", podUID, endpoints[podUID], expectedEndpoints[podUID])
		}
		sort.Ints(endpoints[podUID])
		sort.Ints(expectedEndpoints[podUID])
		for index := range endpoints[podUID] {
			if endpoints[podUID][index] != expectedEndpoints[podUID][index] {
				Failf("invalid list of ports for uid %v. Got %v, expected %v", podUID, endpoints[podUID], expectedEndpoints[podUID])
			}
		}
	}
}

func CreateServiceSpec(serviceName, externalName string, isHeadless bool, selector map[string]string) *v1.Service {
	headlessService := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: v1.ServiceSpec{
			Selector: selector,
		},
	}
	if externalName != "" {
		headlessService.Spec.Type = v1.ServiceTypeExternalName
		headlessService.Spec.ExternalName = externalName
	} else {
		headlessService.Spec.Ports = []v1.ServicePort{
			{Port: 80, Name: "http", Protocol: "TCP"},
		}
	}
	if isHeadless {
		headlessService.Spec.ClusterIP = "None"
	}
	return headlessService
}

// EnableAndDisableInternalLB returns two functions for enabling and disabling the internal load balancer
// setting for the supported cloud providers: GCE/GKE and Azure
func EnableAndDisableInternalLB() (enable func(svc *v1.Service), disable func(svc *v1.Service)) {
	enable = func(svc *v1.Service) {}
	disable = func(svc *v1.Service) {}

	// TODO: this provider agnostic (switch TestContext.Provider)?
	enable = func(svc *v1.Service) {
		svc.ObjectMeta.Annotations = map[string]string{cloudprovider.ServiceAnnotationLoadBalancerInternal: ""}
	}
	disable = func(svc *v1.Service) {
		delete(svc.ObjectMeta.Annotations, cloudprovider.ServiceAnnotationLoadBalancerInternal)
	}

	return
}

func (f *CloudProviderFramework) VerifyLoadBalancerPolicy(loadBalancerId string, loadbalancerPolicy string, lbtype string) error {
	pollFunc := func() (done bool, err error) {
		loadBalancer, err := f.Client.LoadBalancer(lbtype).GetLoadBalancer(context.TODO(), loadBalancerId)
		if err != nil {
			return false, err
		}
		success, err := testLoadBalancerPolicy(loadBalancer, loadbalancerPolicy)
		if err != nil {
			return false, err
		}
		if success {
			Logf("loadbalancer policy matches expected policy.")
			return true, nil
		}
		Logf("loadbalancer policy did not match expected - will retry")
		return false, nil
	}
	return wait.PollImmediate(5*time.Second, 5*time.Minute, pollFunc)
}

func testLoadBalancerPolicy(loadBalancer *client.GenericLoadBalancer, loadbalancerPolicy string) (bool, error) {
	if loadBalancer == nil || len(loadBalancer.BackendSets) == 0 {
		return false, gerrors.Errorf("Could not find load balancer policy.")
	}

	for _, backendSet := range loadBalancer.BackendSets {
		if *backendSet.Policy != loadbalancerPolicy {
			return false, nil
		}
	}
	return true, nil
}
