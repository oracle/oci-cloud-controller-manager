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

package bmcs

import (
	"errors"
	"fmt"

	"github.com/oracle/kubernetes-cloud-controller-manager/pkg/bmcs/client"

	"github.com/golang/glog"

	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/cloudprovider"
	k8sports "k8s.io/kubernetes/pkg/master/ports"

	baremetal "github.com/oracle/bmcs-go-sdk"
)

const (
	// ServiceAnnotationLoadBalancerInternal is a service annotation for
	// specifying that a load balancer should be internal.
	ServiceAnnotationLoadBalancerInternal = "service.beta.kubernetes.io/bmcs-load-balancer-internal"
	// ServiceAnnotationLoadBalancerShape is a Service annotation for
	// specifying the Shape of a load balancer.
	ServiceAnnotationLoadBalancerShape = "service.beta.kubernetes.io/bmcs-load-balancer-shape"
	// ServiceAnnotationLoadBalancerSubnet1 is a Service annotation for
	// specifying the first subnet of a load balancer.
	ServiceAnnotationLoadBalancerSubnet1 = "service.beta.kubernetes.io/bmcs-load-balancer-subnet1"
	// ServiceAnnotationLoadBalancerSubnet2 is a Service annotation for
	// specifying the second subnet of a load balancer.
	ServiceAnnotationLoadBalancerSubnet2 = "service.beta.kubernetes.io/bmcs-load-balancer-subnet2"
)

const (
	// Fallback value if annotation on service is not set
	lbDefaultShape = "100Mbps"

	lbNodesHealthCheckPath  = "/healthz"
	lbNodesHealthCheckPort  = k8sports.ProxyHealthzPort
	lbNodesHealthCheckProto = "HTTP"
)

// LBSpec holds the data required to build a BMCS load balancer from a
// kubernetes service.
type LBSpec struct {
	Name    string
	Shape   string
	Service *api.Service
	NodeIPs []string
	Subnets []string
}

// NewLBSpec creates a LB Spec from a kubernetes service and a slice of nodes.
func NewLBSpec(cp *CloudProvider, service *api.Service, nodeIPs []string) (LBSpec, error) {
	if err := validateProtocols(service.Spec.Ports); err != nil {
		return LBSpec{}, err
	}

	if service.Spec.SessionAffinity != api.ServiceAffinityNone {
		return LBSpec{}, errors.New("BMCS only supports SessionAffinity `None` currently")
	}

	if service.Spec.LoadBalancerIP != "" {
		return LBSpec{}, errors.New("BMCS does not support setting the LoadBalancerIP")
	}

	internalLB := false
	internalAnnotation := service.Annotations[ServiceAnnotationLoadBalancerInternal]
	if internalAnnotation != "" {
		internalLB = true
	}

	if internalLB {
		return LBSpec{}, fmt.Errorf("BMCS does not currently support internal load balancers")
	}

	// TODO (apryde): We should detect when this changes and WARN as we don't
	// support updating a load balancer's Shape.
	lbShape := service.Annotations[ServiceAnnotationLoadBalancerShape]
	if lbShape == "" {
		lbShape = lbDefaultShape
	}

	// TODO (apryde): What happens if this changes?
	subnet1, ok := service.Annotations[ServiceAnnotationLoadBalancerSubnet1]
	if !ok {
		subnet1 = cp.config.LoadBalancer.Subnet1
	}
	subnet2, ok := service.Annotations[ServiceAnnotationLoadBalancerSubnet2]
	if !ok {
		subnet2 = cp.config.LoadBalancer.Subnet2
	}

	return LBSpec{
		Name:    deriveLoadBalancerName(service),
		Shape:   lbShape,
		Service: service,
		NodeIPs: nodeIPs,
		Subnets: []string{subnet1, subnet2},
	}, nil
}

// GetBackendSets builds a map of BackendSets based on the LBSpec.
// TODO (apryde): Can/should we support SSL config here?
// NOTE (apryde): Currently adds a node health-check per service port as
// BackendSets and HealthCheckers are coupled.
func (s *LBSpec) GetBackendSets() map[string]baremetal.BackendSet {
	backendSets := make(map[string]baremetal.BackendSet)
	for _, servicePort := range s.Service.Spec.Ports {
		name := fmt.Sprintf("%v-%d", servicePort.Protocol, servicePort.Port)
		backendSet := baremetal.BackendSet{
			Name:     name,
			Policy:   client.DefaultLoadBalancerPolicy,
			Backends: []baremetal.Backend{},
			HealthChecker: &baremetal.HealthChecker{
				Protocol: lbNodesHealthCheckProto,
				URLPath:  lbNodesHealthCheckPath,
				Port:     lbNodesHealthCheckPort,
			},
		}
		for _, ip := range s.NodeIPs {
			backendSet.Backends = append(backendSet.Backends, baremetal.Backend{
				IPAddress: ip,
				Port:      int(servicePort.NodePort),
				Weight:    1,
			})
		}
		backendSets[name] = backendSet
	}
	return backendSets
}

// GetListeners builds a map of listeners based on the LBSpec.
func (s *LBSpec) GetListeners() map[string]baremetal.Listener {
	listeners := make(map[string]baremetal.Listener)
	for _, servicePort := range s.Service.Spec.Ports {
		name := fmt.Sprintf("%v-%d", servicePort.Protocol, servicePort.Port)
		listener := baremetal.Listener{
			Name: name,
			DefaultBackendSetName: name,
			Protocol:              string(servicePort.Protocol),
			Port:                  int(servicePort.Port),
		}
		listeners[name] = listener
	}
	return listeners
}

func deriveLoadBalancerName(service *api.Service) string {
	return fmt.Sprintf("%s-%s", service.Name, cloudprovider.GetLoadBalancerName(service))
}

// Extract a list of all the external IP addresses for the available Kubernetes nodes
// Each node IP address must be added to the backend set
func extractNodeIPs(nodes []*api.Node) []string {
	nodeIPs := []string{}
	for _, node := range nodes {
		for _, nodeAddr := range node.Status.Addresses {
			if nodeAddr.Type == api.NodeInternalIP {
				nodeIPs = append(nodeIPs, nodeAddr.Address)
			}
		}
	}
	return nodeIPs
}

// GetLoadBalancer returns whether the specified load balancer exists, and if
// so, what its status is.
func (cp *CloudProvider) GetLoadBalancer(clusterName string, service *api.Service) (status *api.LoadBalancerStatus, exists bool, retErr error) {
	name := deriveLoadBalancerName(service)
	glog.V(4).Infof("Fetching load balancer with name '%s'", name)

	lb, err := cp.client.GetLoadBalancerByName(name)
	if err != nil {
		if err, ok := err.(*client.SearchError); ok {
			if err.NotFound {
				glog.V(2).Infof("Load balancer '%s' does not exist", name)
				return nil, false, nil
			}
		}
		return nil, false, err
	}

	lbStatus, err := loadBalancerToStatus(lb)
	if err != nil {
		return nil, false, err
	}

	return lbStatus, true, nil
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
		if err, ok := err.(*client.SearchError); ok {
			if err.NotFound {
				glog.Infof("Attempting to create a load balancer with name '%s'", spec.Name)
				var cerr error
				lb, cerr = cp.client.CreateAndAwaitLoadBalancer(spec.Name, spec.Shape, spec.Subnets)
				if cerr != nil {
					glog.Errorf("Failed to create load balancer: %s", err)
					return nil, cerr
				}
				glog.Infof("Created load balancer '%s' with OCID '%s'", lb.DisplayName, lb.ID)
			}
		} else {
			return nil, err
		}
	}

	desiredBackendSets := spec.GetBackendSets()
	desiredListeners := spec.GetListeners()

	secListMngr, err := newSecurityListManagerFromLBSpec(cp.client, &spec)
	if err != nil {
		return nil, err
	}

	{
		// 1. Has the front end port changed?
		additions, removals, err := getListenerModifications(desiredListeners, lb.Listeners)
		if err != nil {
			return nil, err
		}

		if len(additions) > 0 {
			glog.V(4).Infof("Adding %d listeners", len(additions))
			for _, listener := range additions {
				// Create a backend set for this listener
				key := listener.Name
				bes, ok := desiredBackendSets[key]
				if !ok {
					return nil, fmt.Errorf("Cannot create backend set with name %s", key)
				}

				backendSet, err := cp.client.CreateAndAwaitBackendSet(lb, bes)
				if err != nil {
					return nil, err
				}

				err = secListMngr.EnsureRulesAdded(uint64(backendSet.Backends[0].Port))
				if err != nil {
					return nil, err
				}

				lb.BackendSets[backendSet.Name] = *backendSet

				err = cp.client.CreateAndAwaitListener(lb, listener)
				if err != nil {
					return nil, err
				}

				lb.Listeners[listener.Name] = listener
			}
		}
		if len(removals) > 0 {
			glog.V(4).Infof("Removing %d Listeners and BackendSets", len(removals))
			for _, listener := range removals {
				// TODO (apryde): We should probably at least spawn go routines to
				// poll these WorkRequests and log loudly if they fail.
				_, err := cp.client.DeleteListener(lb.ID, listener.Name, nil)
				if err != nil {
					return nil, err
				}
				delete(lb.Listeners, listener.Name)

				backendSet := lb.BackendSets[listener.Name]
				err = secListMngr.EnsureRulesRemoved(uint64(backendSet.Backends[0].Port))
				if err != nil {
					return nil, err
				}

				_, err = cp.client.DeleteBackendSet(lb.ID, backendSet.Name, nil)
				if err != nil {
					return nil, err
				}
				delete(lb.BackendSets, listener.Name)
			}
		}
	}

	// At this point we have a load balancer, and all listeners are correct
	// and a backend set exists for each listener. Now we just need to
	// update the backends if:
	//  1. A Node has been added or removed.
	//  2. A NodePort has been modified.

	{
		additions, removals := getAllBackendModifications(desiredBackendSets, lb.BackendSets)

		if len(additions) > 0 {
			for backendName, backends := range additions {
				err = secListMngr.EnsureRulesAdded(uint64(backends[0].Port))
				if err != nil {
					return nil, err
				}

				for _, backend := range backends {
					glog.V(4).Infof("Adding Backend '%s:%d' to '%s'", backend.IPAddress, backend.Port, backendName)
					_, err = cp.client.CreateBackend(lb.ID, backendName, backend.IPAddress, backend.Port, nil)
					if err != nil {
						return nil, err
					}
				}
			}
		}
		if len(removals) > 0 {
			for backendName, backends := range removals {
				err = secListMngr.EnsureRulesRemoved(uint64(backends[0].Port))
				if err != nil {
					return nil, err
				}

				for _, backend := range backends {
					target := fmt.Sprintf("%s:%d", backend.IPAddress, backend.Port)
					glog.V(4).Infof("Deleting Backend '%s' from '%s'", target, backendName)
					_, err = cp.client.DeleteBackend(lb.ID, backendName, target, nil)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	err = secListMngr.Save()
	if err != nil {
		return nil, err
	}

	return loadBalancerToStatus(lb)
}

// getAllBackendModifications returns the Backends that need to be
// added/removed for the actual state of a []BackendSets to converge with the
// desired state. Returns a map keyed by BackendSet.Name as BackendSet.Name is
// needed when adding/deleting Backends.
func getAllBackendModifications(desired, actual map[string]baremetal.BackendSet) (additions, removals map[string][]baremetal.Backend) {
	additions = make(map[string][]baremetal.Backend)
	removals = make(map[string][]baremetal.Backend)

	for _, backendSet := range desired {
		toAdd, toRemove := getBackendModifications(backendSet, actual[backendSet.Name])
		if len(toAdd) > 0 {
			additions[backendSet.Name] = toAdd
		}
		if len(toRemove) > 0 {
			removals[backendSet.Name] = toRemove
		}
	}
	return additions, removals
}

// getBackendModifications returns the load balancer Backends that need to be
// added/removed for the actual state of a BackendSet to converge with the
// desired state.
func getBackendModifications(desired, actual baremetal.BackendSet) (additions, removals []baremetal.Backend) {
	nameFormat := "%s-%d"
	lookup := make(map[string]baremetal.Backend)

	desiredSet := sets.NewString()
	for _, backend := range desired.Backends {
		name := fmt.Sprintf(nameFormat, backend.IPAddress, backend.Port)
		desiredSet.Insert(name)
		lookup[name] = backend
	}

	actualSet := sets.NewString()
	for _, backend := range actual.Backends {
		name := fmt.Sprintf(nameFormat, backend.IPAddress, backend.Port)
		actualSet.Insert(name)
		lookup[name] = backend
	}

	additionNames := desiredSet.Difference(actualSet)
	for _, name := range additionNames.List() {
		additions = append(additions, lookup[name])
	}
	removalNames := actualSet.Difference(desiredSet)
	for _, name := range removalNames.List() {
		removals = append(removals, lookup[name])
	}

	return additions, removals
}

// getListenerModifications returns the load balancer Listeners that need to be
// added/removed for the actual state to converge with the desired state.
func getListenerModifications(desired, actual map[string]baremetal.Listener) (additions, removals []baremetal.Listener, err error) {
	nameFormat := "%s-%d"

	desiredSet := sets.NewString()
	for _, listener := range desired {
		desiredSet.Insert(fmt.Sprintf(nameFormat, listener.Protocol, listener.Port))
	}

	actualSet := sets.NewString()
	for _, listener := range actual {
		actualSet.Insert(fmt.Sprintf(nameFormat, listener.Protocol, listener.Port))
	}

	additionNames := desiredSet.Difference(actualSet)
	removalNames := actualSet.Difference(desiredSet)

	for _, name := range additionNames.List() {
		listener, ok := desired[name]
		if !ok {
			return nil, nil, fmt.Errorf("could not find listener with name %q", name)
		}
		additions = append(additions, listener)
	}

	for _, name := range removalNames.List() {
		listener, ok := actual[name]
		if !ok {
			return nil, nil, fmt.Errorf("could not find listener with name %q", name)
		}
		removals = append(removals, listener)
	}

	return additions, removals, nil
}

// UpdateLoadBalancer : TODO find out where this is called
func (cp *CloudProvider) UpdateLoadBalancer(clusterName string, service *api.Service, nodes []*api.Node) error {
	name := deriveLoadBalancerName(service)
	glog.Infof("Attempting to update load balancer '%s'", name)

	_, err := cp.EnsureLoadBalancer(clusterName, service, nodes)
	return err
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it
// exists, returning nil if the load balancer specified either didn't exist or
// was successfully deleted.
func (cp *CloudProvider) EnsureLoadBalancerDeleted(clusterName string, service *api.Service) error {
	name := deriveLoadBalancerName(service)

	glog.Infof("Attempting to delete load balancer with name '%s'", name)
	lb, err := cp.client.GetLoadBalancerByName(name)
	if err != nil {
		if err, ok := err.(*client.SearchError); ok {
			if err.NotFound {
				glog.Infof("Could not find load balancer with name '%s'. Nothing to do.", name)
				return nil
			}
		}
		return err
	}

	nodeIPs := sets.NewString()
	for _, backendSet := range lb.BackendSets {
		for _, backend := range backendSet.Backends {
			nodeIPs.Insert(backend.IPAddress)
		}
	}
	spec, err := NewLBSpec(cp, service, nodeIPs.List())
	if err != nil {
		return err
	}

	// Remove related SecurityList rules.
	secListMngr, err := newSecurityListManagerFromLBSpec(cp.client, &spec)
	if err != nil {
		return err
	}
	for _, backendSet := range lb.BackendSets {
		err = secListMngr.EnsureRulesRemoved(uint64(backendSet.Backends[0].Port))
		if err != nil {
			return err
		}
	}
	err = secListMngr.Save()
	if err != nil {
		return err
	}

	glog.Infof("Deleting load balancer '%s' (OCID: '%s')", lb.DisplayName, lb.ID)

	workReqID, err := cp.client.DeleteLoadBalancer(lb.ID, &baremetal.ClientRequestOptions{})
	if err != nil {
		return err
	}

	_, err = cp.client.AwaitWorkRequest(workReqID)
	return err
}

// Given an BMCS load balancer, return a LoadBalancerStatus
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

// validateProtocols validates that BMCS supports the protocol of all
// ServicePorts defined by a service.
func validateProtocols(servicePorts []api.ServicePort) error {
	for _, servicePort := range servicePorts {
		if servicePort.Protocol == api.ProtocolUDP {
			return fmt.Errorf("BMCS load balancers do not support UDP")
		}
	}
	return nil
}
