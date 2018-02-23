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

package framework

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	uuid "k8s.io/apimachinery/pkg/util/uuid"
	wait "k8s.io/apimachinery/pkg/util/wait"
	kubernetes "k8s.io/client-go/kubernetes"
	api "k8s.io/kubernetes/pkg/api"
	imageutils "k8s.io/kubernetes/test/utils/image"
)

// ServiceNodePortRange should match whatever the default/configured range is
var ServiceNodePortRange = utilnet.PortRange{Base: 30000, Size: 2768}

// ServiceTestJig provides helper methods for testing services.
type ServiceTestJig struct {
	ID     string
	Name   string
	Client kubernetes.Interface
	Labels map[string]string
}

// NewServiceTestJig allocates and inits a new ServiceTestJig.
func NewServiceTestJig(client kubernetes.Interface, name string) *ServiceTestJig {
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

// CreateTCPServiceOrFail creates a new TCP Service based on the jig's
// defaults.  Callers can provide a function to tweak the Service object before
// it is created.
func (j *ServiceTestJig) CreateTCPServiceOrFail(namespace string, tweak func(svc *v1.Service)) *v1.Service {
	svc := j.newServiceTemplate(namespace, v1.ProtocolTCP, 80)
	if tweak != nil {
		tweak(svc)
	}
	result, err := j.Client.CoreV1().Services(namespace).Create(svc)
	if err != nil {
		Failf("Failed to create TCP Service %q: %v", svc.Name, err)
	}
	return result
}

// CreateLoadBalancerService creates a loadbalancer service and waits
// for it to acquire an ingress IP.
func (j *ServiceTestJig) CreateLoadBalancerService(namespace, serviceName string, timeout time.Duration, tweak func(svc *v1.Service)) *v1.Service {
	Logf("Creating a service %s/%s with type=LoadBalancer", namespace, serviceName)
	svc := j.CreateTCPServiceOrFail(namespace, func(svc *v1.Service) {
		svc.Spec.Type = v1.ServiceTypeLoadBalancer
		if tweak != nil {
			tweak(svc)
		}
	})

	Logf("Waiting for loadbalancer for service %s/%s", namespace, serviceName)
	svc = j.WaitForLoadBalancerOrFail(namespace, serviceName, timeout)
	return svc
}

func (j *ServiceTestJig) waitForConditionOrFail(namespace, name string, timeout time.Duration, message string, conditionFn func(*v1.Service) bool) *v1.Service {
	var service *v1.Service
	pollFunc := func() (bool, error) {
		svc, err := j.Client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if conditionFn(svc) {
			service = svc
			return true, nil
		}
		return false, nil
	}
	if err := wait.PollImmediate(Poll, timeout, pollFunc); err != nil {
		Failf("Timed out waiting for service %q to %s", name, message)
	}
	return service
}

// WaitForLoadBalancerOrFail waits for a Service type=LoadBalancer to be
// assigned an ingress point by the cloudprovider.
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

// newRCTemplate returns the default v1.ReplicationController object for
// this jig, but does not actually create the RC.  The default RC has the same
// name as the jig and runs the "netexec" container.
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
							Name:  "netexec",
							Image: imageutils.GetE2EImage(imageutils.Netexec),
							Args:  []string{"--http-port=80", "--udp-port=80"},
							ReadinessProbe: &v1.Probe{
								PeriodSeconds: 3,
								Handler: v1.Handler{
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

// RunOrFail creates a ReplicationController and Pod(s) and waits for the
// Pod(s) to be running. Callers can provide a function to tweak the RC object
// before it is created.
func (j *ServiceTestJig) RunOrFail(namespace string, tweak func(rc *v1.ReplicationController)) *v1.ReplicationController {
	rc := j.newRCTemplate(namespace)
	if tweak != nil {
		tweak(rc)
	}
	result, err := j.Client.CoreV1().ReplicationControllers(namespace).Create(rc)
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

func (j *ServiceTestJig) waitForPodsCreated(namespace string, replicas int) ([]string, error) {
	timeout := 2 * time.Minute
	// List the pods, making sure we observe all the replicas.
	label := labels.SelectorFromSet(labels.Set(j.Labels))
	Logf("Waiting up to %v for %d pods to be created", timeout, replicas)
	for start := time.Now(); time.Since(start) < timeout; time.Sleep(2 * time.Second) {
		options := metav1.ListOptions{LabelSelector: label.String()}
		pods, err := j.Client.CoreV1().Pods(namespace).List(options)
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

// SanityCheckService sanity checks some basic properties of a given Service.
func (j *ServiceTestJig) SanityCheckService(svc *v1.Service, svcType v1.ServiceType) {
	if svc.Spec.Type != svcType {
		Failf("Unexpected Spec.Type (%s) for service, expected %s", svc.Spec.Type, svcType)
	}

	if svcType != v1.ServiceTypeExternalName {
		if svc.Spec.ExternalName != "" {
			Failf("Unexpected Spec.ExternalName (%s) for service, expected empty", svc.Spec.ExternalName)
		}
		if svc.Spec.ClusterIP != api.ClusterIPNone && svc.Spec.ClusterIP == "" {
			Failf("Didn't get ClusterIP for non-ExternamName service")
		}
	} else {
		if svc.Spec.ClusterIP != "" {
			Failf("Unexpected Spec.ClusterIP (%s) for ExternamName service, expected empty", svc.Spec.ClusterIP)
		}
	}

	expectNodePorts := false
	if svcType != v1.ServiceTypeClusterIP && svcType != v1.ServiceTypeExternalName {
		expectNodePorts = true
	}
	for i, port := range svc.Spec.Ports {
		hasNodePort := (port.NodePort != 0)
		if hasNodePort != expectNodePorts {
			Failf("Unexpected Spec.Ports[%d].NodePort (%d) for service", i, port.NodePort)
		}
		if hasNodePort {
			if !ServiceNodePortRange.Contains(int(port.NodePort)) {
				Failf("Out-of-range nodePort (%d) for service", port.NodePort)
			}
		}
	}
	expectIngress := false
	if svcType == v1.ServiceTypeLoadBalancer {
		expectIngress = true
	}
	hasIngress := len(svc.Status.LoadBalancer.Ingress) != 0
	if hasIngress != expectIngress {
		Failf("Unexpected number of Status.LoadBalancer.Ingress (%d) for service", len(svc.Status.LoadBalancer.Ingress))
	}
	if hasIngress {
		for i, ing := range svc.Status.LoadBalancer.Ingress {
			if ing.IP == "" && ing.Hostname == "" {
				Failf("Unexpected Status.LoadBalancer.Ingress[%d] for service: %#v", i, ing)
			}
		}
	}
}

func (j *ServiceTestJig) TestReachableHTTP(host string, port int, timeout time.Duration) {
	j.TestReachableHTTPWithRetriableErrorCodes(host, port, []int{}, timeout)
}

func (j *ServiceTestJig) TestReachableHTTPWithRetriableErrorCodes(host string, port int, retriableErrCodes []int, timeout time.Duration) {
	if err := wait.PollImmediate(Poll, timeout, func() (bool, error) {
		return TestReachableHTTPWithRetriableErrorCodes(host, port, "/echo?msg=hello", "hello", retriableErrCodes)
	}); err != nil {
		if err == wait.ErrWaitTimeout {
			Failf("Could not reach HTTP service through %v:%v after %v", host, port, timeout)
		} else {
			Failf("Failed to reach HTTP service through %v:%v: %v", host, port, err)
		}
	}
}
