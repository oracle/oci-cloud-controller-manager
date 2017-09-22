// Copyright 2017 The Oracle Kubernetes Cloud Controller Manager Authors
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
	"fmt"

	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"
	k8sports "k8s.io/kubernetes/pkg/master/ports"

	"github.com/golang/glog"
	baremetal "github.com/oracle/bmcs-go-sdk"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
)

const (
	// ServiceAnnotationLoadBalancerInternal is a service annotation for
	// specifying that a load balancer should be internal.
	ServiceAnnotationLoadBalancerInternal = "service.beta.kubernetes.io/oci-load-balancer-internal"

	// ServiceAnnotationLoadBalancerShape is a Service annotation for
	// specifying the Shape of a load balancer.
	ServiceAnnotationLoadBalancerShape = "service.beta.kubernetes.io/oci-load-balancer-shape"

	// ServiceAnnotationLoadBalancerSubnet1 is a Service annotation for
	// specifying the first subnet of a load balancer.
	ServiceAnnotationLoadBalancerSubnet1 = "service.beta.kubernetes.io/oci-load-balancer-subnet1"

	// ServiceAnnotationLoadBalancerSubnet2 is a Service annotation for
	// specifying the second subnet of a load balancer.
	ServiceAnnotationLoadBalancerSubnet2 = "service.beta.kubernetes.io/oci-load-balancer-subnet2"

	// ServiceAnnotationLoadBalancerSSLPorts is a Service annotation for
	// specifying the ports to enable SSL termination on the corresponding load
	// balancer listener.
	ServiceAnnotationLoadBalancerSSLPorts = "service.beta.kubernetes.io/oci-load-balancer-ssl-ports"

	// ServiceAnnotationLoadBalancerTLSSecret is a Service annotation for
	// specifying the TLS secret ti install on the load balancer listeners which
	// have SSL enabled.
	// See: https://kubernetes.io/docs/concepts/services-networking/ingress/#tls
	ServiceAnnotationLoadBalancerTLSSecret = "service.beta.kubernetes.io/oci-load-balancer-tls-secret"
)

const (
	// Fallback value if annotation on service is not set
	lbDefaultShape = "100Mbps"

	lbNodesHealthCheckPath  = "/healthz"
	lbNodesHealthCheckPort  = k8sports.ProxyHealthzPort
	lbNodesHealthCheckProto = "HTTP"
)

func (cp *CloudProvider) readTLSSecret(secretString, serviceNS string) (cert, key string, err error) {
	ns, name := parseSecretString(secretString)
	if ns == "" {
		ns = serviceNS
	}
	secret, err := cp.kubeclient.CoreV1().Secrets(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return cert, key, err
	}

	certBytes, ok := secret.Data[sslCertificateFileName]
	if !ok {
		err = fmt.Errorf("%s not found in secret %s/%s", sslCertificateFileName, ns, name)
		return
	}
	keyBytes, ok := secret.Data[sslPrivateKeyFileName]
	if !ok {
		err = fmt.Errorf("%s not found in secret %s/%s", sslCertificateFileName, ns, name)
		return
	}

	return string(certBytes), string(keyBytes), nil
}

// ensureSSLCertificate creates a OCI SSL certificate to the given load
// balancer, if it doesn't already exist.
func (cp *CloudProvider) ensureSSLCertificate(name string, svc *api.Service, lb *baremetal.LoadBalancer) error {
	_, err := cp.client.GetCertificateByName(lb.ID, name)
	if err == nil {
		glog.V(4).Infof("Certificate: %q already exists on load balancer: %q", name, lb.DisplayName)
		return nil
	}
	if !client.IsNotFound(err) {
		return err
	}

	secretString, ok := svc.Annotations[ServiceAnnotationLoadBalancerTLSSecret]
	if !ok {
		return fmt.Errorf("no %s annotation found", ServiceAnnotationLoadBalancerTLSSecret)
	}

	cert, key, err := cp.readTLSSecret(secretString, svc.Namespace)
	if err != nil {
		return err
	}

	err = cp.client.CreateAndAwaitCertificate(lb, name, cert, key)
	if err != nil {
		return err
	}

	glog.V(2).Infof("Created certificate %q on load balancer %q", name, lb.DisplayName)
	return nil
}

// GetLoadBalancer returns whether the specified load balancer exists, and if
// so, what its status is.
func (cp *CloudProvider) GetLoadBalancer(clusterName string, service *api.Service) (status *api.LoadBalancerStatus, exists bool, retErr error) {
	name := GetLoadBalancerName(service)
	glog.V(4).Infof("Fetching load balancer with name '%s'", name)

	lb, err := cp.client.GetLoadBalancerByName(name)
	if err != nil {
		if client.IsNotFound(err) {
			glog.V(2).Infof("Load balancer '%s' does not exist", name)
			return nil, false, nil
		}

		return nil, false, err
	}

	lbStatus, err := loadBalancerToStatus(lb)
	if err != nil {
		return nil, false, err
	}

	return lbStatus, true, nil
}

func getCertificateName(lb *baremetal.LoadBalancer) string {
	return lb.DisplayName
}

// EnsureLoadBalancer creates a new load balancer or updates the existing one.
// Returns the status of the balancer (i.e it's public IP address if one exists).
func (cp *CloudProvider) EnsureLoadBalancer(clusterName string, service *api.Service, nodes []*api.Node) (*api.LoadBalancerStatus, error) {
	spec, err := NewLBSpec(cp, service, extractNodeIPs(nodes))
	if err != nil {
		glog.Errorf("Failed to create LBSpec: %v", err)
		return nil, err
	}

	glog.V(4).Infof("Ensure load balancer '%s' called for '%s' with %d nodes.", spec.Name, spec.Service.Name, len(nodes))

	var lb *baremetal.LoadBalancer
	lb, err = cp.client.GetLoadBalancerByName(spec.Name)
	if err != nil {
		if client.IsNotFound(err) {
			glog.Infof("Attempting to create a load balancer with name '%s'", spec.Name)
			var cerr error
			lb, cerr = cp.client.CreateAndAwaitLoadBalancer(spec.Name, spec.Shape, spec.Subnets)
			if cerr != nil {
				glog.Errorf("Failed to create load balancer: %s", err)
				return nil, cerr
			}
			glog.Infof("Created load balancer '%s' with OCID '%s'", lb.DisplayName, lb.ID)
		} else {
			return nil, err
		}
	}

	certificateName := getCertificateName(lb)

	sslConfigMap, err := spec.GetSSLConfig(certificateName)
	if sslEnabled(sslConfigMap) {
		if err = cp.ensureSSLCertificate(certificateName, spec.Service, lb); err != nil {
			return nil, err
		}
	}

	sourceCIDRs, err := getLoadBalancerSourceRanges(service)
	if err != nil {
		return nil, err
	}

	err = cp.updateBackendSets(lb, spec)
	if err != nil {
		return nil, fmt.Errorf("update backendsets: %v", err)
	}

	err = cp.updateListeners(lb, spec, sslConfigMap, sourceCIDRs)
	if err != nil {
		return nil, fmt.Errorf("udpate listeners: %v", err)
	}

	return loadBalancerToStatus(lb)
}

func (cp *CloudProvider) updateBackendSets(lb *baremetal.LoadBalancer, spec LBSpec) error {
	lbOCID := lb.ID

	actual := lb.BackendSets
	desired := spec.GetBackendSets()

	actions := getBackendSetChanges(actual, desired)

	for _, action := range actions {
		var workRequestID string
		var err error

		be := action.BackendSet

		glog.V(2).Infof("Applying `%s` action on backend set `%s` for lb `%s`", action.Type, be.Name, lbOCID)

		switch action.Type {
		case Create:
			workRequestID, err = cp.client.CreateBackendSet(
				lbOCID,
				be.Name,
				be.Policy,
				be.Backends,
				be.HealthChecker,
				nil, // ssl config
				nil, // session persistence
				nil, // create opts
			)
		case Update:
			workRequestID, err = cp.client.UpdateBackendSet(lbOCID, be.Name, &baremetal.UpdateLoadBalancerBackendSetOptions{
				Policy:        be.Policy,
				HealthChecker: be.HealthChecker,
				Backends:      be.Backends,
			})
		case Delete:
			workRequestID, err = cp.client.DeleteBackendSet(lbOCID, be.Name, nil)
		}

		if err != nil {
			return err
		}

		_, err = cp.client.AwaitWorkRequest(workRequestID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cp *CloudProvider) updateListeners(lb *baremetal.LoadBalancer, spec LBSpec, sslConfigMap map[int]*baremetal.SSLConfiguration, sourceCIDRs []string) error {
	lbOCID := lb.ID

	desired := spec.GetListeners(sslConfigMap)
	actions := getListenerChanges(lb.Listeners, desired)
	for _, action := range actions {
		var workRequestID string
		var err error
		l := action.Listener

		glog.V(2).Infof("Applying `%s` action on listener `%s` for lb `%s`", action.Type, l.Name, lbOCID)

		backends := spec.GetBackendSets()[l.DefaultBackendSetName].Backends

		switch action.Type {
		case Create:
			err = cp.securityListManager.Update(spec.Subnets, sourceCIDRs, &l, backends)
			if err != nil {
				return err
			}

			workRequestID, err = cp.client.CreateListener(
				lbOCID,
				l.Name,
				l.DefaultBackendSetName,
				l.Protocol,
				l.Port,
				l.SSLConfig,
				nil, // create opts
			)
		case Update:
			err = cp.securityListManager.Update(spec.Subnets, sourceCIDRs, &l, backends)
			if err != nil {
				return err
			}

			workRequestID, err = cp.client.UpdateListener(lbOCID, l.Name, &baremetal.UpdateLoadBalancerListenerOptions{
				DefaultBackendSetName: l.DefaultBackendSetName,
				Port:      l.Port,
				Protocol:  l.Protocol,
				SSLConfig: l.SSLConfig,
			})
		case Delete:
			err = cp.securityListManager.Delete(spec.Subnets, &l, backends)
			if err != nil {
				return err
			}

			workRequestID, err = cp.client.DeleteListener(lbOCID, l.Name, nil)
		}

		if err != nil {
			return err
		}

		_, err = cp.client.AwaitWorkRequest(workRequestID)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateLoadBalancer : TODO find out where this is called
func (cp *CloudProvider) UpdateLoadBalancer(clusterName string, service *api.Service, nodes []*api.Node) error {
	name := GetLoadBalancerName(service)
	glog.Infof("Attempting to update load balancer '%s'", name)

	_, err := cp.EnsureLoadBalancer(clusterName, service, nodes)
	return err
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
func (cp *CloudProvider) EnsureLoadBalancerDeleted(clusterName string, service *api.Service) error {
	name := GetLoadBalancerName(service)

	glog.Infof("Attempting to delete load balancer with name '%s'", name)
	lb, err := cp.client.GetLoadBalancerByName(name)
	if err != nil {
		if client.IsNotFound(err) {
			glog.Infof("Could not find load balancer with name '%s'. Nothing to do.", name)
			return nil
		}

		return fmt.Errorf("get load balancer by name %s: %v", name, err)
	}

	nodeIPs := sets.NewString()
	for _, backendSet := range lb.BackendSets {
		for _, backend := range backendSet.Backends {
			nodeIPs.Insert(backend.IPAddress)
		}
	}

	spec, err := NewLBSpec(cp, service, nodeIPs.List())
	if err != nil {
		return fmt.Errorf("new lb spec: %v", err)
	}

	sslConfigMap, err := spec.GetSSLConfig(getCertificateName(lb))
	if err != nil {
		return fmt.Errorf("get ssl config: %v", err)
	}

	for _, listener := range spec.GetListeners(sslConfigMap) {
		glog.V(4).Infof("Deleting security rules for listener `%s` for load balancer `%s`", listener.Name, lb.ID)

		backends := spec.GetBackendSets()[listener.DefaultBackendSetName].Backends

		err := cp.securityListManager.Delete(spec.Subnets, &listener, backends)

		if err != nil {
			return fmt.Errorf("delete security rules for listener %s: %v", listener.Name, err)
		}
	}

	glog.Infof("Deleting load balancer '%s' (OCID: '%s')", lb.DisplayName, lb.ID)

	workReqID, err := cp.client.DeleteLoadBalancer(lb.ID, &baremetal.ClientRequestOptions{})
	if err != nil {
		return fmt.Errorf("delete load balancer: %v", err)
	}

	_, err = cp.client.AwaitWorkRequest(workReqID)
	return err
}

// Given an OCI load balancer, return a LoadBalancerStatus
func loadBalancerToStatus(lb *baremetal.LoadBalancer) (*api.LoadBalancerStatus, error) {
	if len(lb.IPAddresses) == 0 {
		return nil, fmt.Errorf("no IPAddresses found for load balancer '%s'", lb.DisplayName)
	}

	ingress := []api.LoadBalancerIngress{}
	for _, ip := range lb.IPAddresses {
		ingress = append(ingress, api.LoadBalancerIngress{IP: ip.IPAddress})
	}
	return &api.LoadBalancerStatus{Ingress: ingress}, nil
}

func getLoadBalancerSourceRanges(service *api.Service) ([]string, error) {
	sourceRanges, err := apiservice.GetLoadBalancerSourceRanges(service)
	if err != nil {
		return []string{}, err
	}

	sourceCIDRs := make([]string, 0, len(sourceRanges))
	for _, sourceRange := range sourceRanges {
		sourceCIDRs = append(sourceCIDRs, sourceRange.String())
	}

	return sourceCIDRs, nil
}
