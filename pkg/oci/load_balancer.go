// Copyright 2017 Oracle and/or its affiliates. All rights reserved.
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
	"context"

	"github.com/golang/glog"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/loadbalancer"
	"github.com/pkg/errors"

	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	sets "k8s.io/apimachinery/pkg/util/sets"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"
	k8sports "k8s.io/kubernetes/pkg/master/ports"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/util"
)

const (
	// ServiceAnnotationLoadBalancerInternal is a service annotation for
	// specifying that a load balancer should be internal.
	ServiceAnnotationLoadBalancerInternal = "service.beta.kubernetes.io/oci-load-balancer-internal"

	// ServiceAnnotationLoadBalancerShape is a Service annotation for
	// specifying the Shape of a load balancer. The shape is a template that
	// determines the load balancer's total pre-provisioned maximum capacity
	// (bandwidth) for ingress plus egress traffic. Available shapes include
	// "100Mbps", "400Mbps", and "8000Mbps".
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

// DefaultLoadBalancerPolicy defines the default traffic policy for load
// balancers created by the CCM.
const DefaultLoadBalancerPolicy = "ROUND_ROBIN"

const (
	// Fallback value if annotation on service is not set
	lbDefaultShape = "100Mbps"

	lbNodesHealthCheckPath  = "/healthz"
	lbNodesHealthCheckPort  = k8sports.ProxyHealthzPort
	lbNodesHealthCheckProto = "HTTP"
)

// GetLoadBalancer returns whether the specified load balancer exists, and if
// so, what its status is.
func (cp *CloudProvider) GetLoadBalancer(clusterName string, service *api.Service) (*api.LoadBalancerStatus, bool, error) {
	name := GetLoadBalancerName(service)
	glog.V(4).Infof("Fetching load balancer with name %q", name)

	lb, err := cp.client.LoadBalancer().GetLoadBalancerByName(context.TODO(), name)
	if err != nil {
		if client.IsNotFound(err) {
			glog.V(2).Infof("Load balancer %q does not exist", name)
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

// getSubnets returns a list of Subnet objects for the corrosponding OCIDs.
func getSubnets(ctx context.Context, subnetIDs []string, n client.NetworkingInterface) ([]*core.Subnet, error) {
	subnets := make([]*core.Subnet, len(subnetIDs))
	for i, id := range subnetIDs {
		subnet, err := n.GetSubnet(context.TODO(), id)
		if err != nil {
			return nil, err
		}
		subnets[i] = subnet
	}
	return subnets, nil
}

// getSubnetsForNodes returns the de-duplicated subnets in which the given
// internal IP addresses reside.
func getSubnetsForNodes(ctx context.Context, nodes []*api.Node, client client.Interface) ([]*core.Subnet, error) {
	var (
		subnetOCIDs = sets.NewString()
		subnets     []*core.Subnet
		ipSet       = sets.NewString()
	)

	for _, node := range nodes {
		ipSet.Insert(util.NodeInternalIP(node))
	}

	for _, node := range nodes {
		// First see if the IP of the node belongs to a subnet in the cache.
		ip := util.NodeInternalIP(node)
		subnet, err := client.Networking().GetSubnetFromCacheByIP(ip)
		if err != nil {
			return nil, err
		}
		if subnet != nil {
			// cache hit
			if !subnetOCIDs.Has(*subnet.Id) {
				subnetOCIDs.Insert(*subnet.Id)
				subnets = append(subnets, subnet)
			}
			// Since we got a cache hit we don't need to do the expensive query to find the subnet.
			continue
		}

		id := util.MapProviderIDToInstanceID(node.Spec.ProviderID)
		vnic, err := client.Compute().GetPrimaryVNICForInstance(ctx, id)
		if err != nil {
			return nil, err
		}

		if vnic.PrivateIp != nil && ipSet.Has(*vnic.PrivateIp) &&
			!subnetOCIDs.Has(*vnic.SubnetId) {
			subnet, err := client.Networking().GetSubnet(ctx, *vnic.SubnetId)
			if err != nil {
				return nil, errors.Wrapf(err, "get subnet %q for instance %q", *vnic.SubnetId, id, err)
			}

			subnets = append(subnets, subnet)
			subnetOCIDs.Insert(*vnic.SubnetId)
		}
	}

	return subnets, nil
}

// readSSLSecret returns the certificate and private key from a Kubernetes TLS
// private key Secret.
func (cp *CloudProvider) readSSLSecret(svc *api.Service) (string, string, error) {
	secretString, ok := svc.Annotations[ServiceAnnotationLoadBalancerTLSSecret]
	if !ok {
		return "", "", errors.Errorf("no %q annotation found", ServiceAnnotationLoadBalancerTLSSecret)
	}

	ns, name := parseSecretString(secretString)
	if ns == "" {
		ns = svc.Namespace
	}
	secret, err := cp.kubeclient.CoreV1().Secrets(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", "", err
	}

	var cert, key []byte
	if cert, ok = secret.Data[sslCertificateFileName]; !ok {
		return "", "", errors.Errorf("%s not found in secret %s/%s", sslCertificateFileName, ns, name)
	}
	if key, ok = secret.Data[sslPrivateKeyFileName]; !ok {
		return "", "", errors.Errorf("%s not found in secret %s/%s", sslPrivateKeyFileName, ns, name)
	}

	return string(cert), string(key), nil
}

// ensureSSLCertificate creates a OCI SSL certificate to the given load
// balancer, if it doesn't already exist.
func (cp *CloudProvider) ensureSSLCertificate(ctx context.Context, spec LBSpec, lb *loadbalancer.LoadBalancer) error {
	name := spec.SSLConfig.Name
	_, err := cp.client.LoadBalancer().GetCertificateByName(ctx, *lb.Id, name)
	if err == nil {
		glog.V(4).Infof("Certificate %q already exists on load balancer %q. Nothing to do.", name, *lb.DisplayName)
		return nil
	}
	if !client.IsNotFound(err) {
		return err
	}

	// Although we iterate here only one certificate is supported at the
	// moment as specifying a
	certs, err := spec.GetCertificates()
	if err != nil {
		return err
	}
	for _, cert := range certs {
		wrID, err := cp.client.LoadBalancer().CreateCertificate(ctx, *lb.Id, *cert.PublicCertificate, *cert.PrivateKey)
		if err != nil {
			return err
		}
		_, err = cp.client.LoadBalancer().AwaitWorkRequest(ctx, wrID)
		if err != nil {
			return err
		}

		glog.V(2).Infof("Created certificate %q on load balancer %q", *cert.CertificateName, *lb.DisplayName)
	}
	return nil
}

// createLoadBalancer creates a new OCI load balancer based on the given spec.
func (cp *CloudProvider) createLoadBalancer(ctx context.Context, spec LBSpec, sourceCIDRs []string) (*api.LoadBalancerStatus, error) {
	glog.Infof("Attempting to create a new load balancer with name %q", spec.Name)

	// First update the security lists so that if it fails (due to
	// the etag bug or otherwise) we'll retry prior to LB creation.
	lbSubnets, err := getSubnets(ctx, spec.Subnets, cp.client.Networking())
	if err != nil {
		return nil, errors.Wrap(err, "getting subnets for load balancers")
	}
	nodeSubnets, err := getSubnetsForNodes(ctx, spec.Nodes, cp.client)
	if err != nil {
		return nil, errors.Wrap(err, "getting subnets for nodes")
	}

	for name, bs := range spec.GetBackendSets() {
		backendPort := *bs.Backends[0].Port
		healthCheckPort := *bs.HealthChecker.Port
		listenerPort := *spec.GetListeners()[name].Port
		if err = cp.securityListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, listenerPort, backendPort, healthCheckPort); err != nil {
			return nil, err
		}
	}

	// Then we create the load balancer and wait for it to be online.
	certs, err := spec.GetCertificates()
	if err != nil {
		return nil, errors.Wrap(err, "get certificates")
	}
	details := loadbalancer.CreateLoadBalancerDetails{
		CompartmentId: common.String(cp.config.Auth.CompartmentOCID),
		DisplayName:   common.String(spec.Name),
		ShapeName:     common.String(spec.Shape),
		IsPrivate:     common.Bool(spec.Internal),
		SubnetIds:     spec.Subnets,
		BackendSets:   spec.GetBackendSets(),
		Listeners:     spec.GetListeners(),
		Certificates:  certs,
	}
	glog.V(4).Infof("CreateLoadBalancerDetails: %#v", details.String())
	wrID, err := cp.client.LoadBalancer().CreateLoadBalancer(ctx, details)
	if err != nil {
		glog.Errorf("Failed to create load balancer: %+v", err)
		return nil, err
	}
	wr, err := cp.client.LoadBalancer().AwaitWorkRequest(ctx, wrID)
	if err != nil {
		return nil, errors.Wrap(err, "awaiting load balancer")
	}

	lb, err := cp.client.LoadBalancer().GetLoadBalancer(ctx, *wr.LoadBalancerId)
	if err != nil {
		return nil, errors.Wrapf(err, "get load balancer %q", *wr.LoadBalancerId)
	}
	glog.Infof("Created load balancer %q with OCID %q", *lb.DisplayName, *lb.Id)
	return loadBalancerToStatus(lb)
}

// EnsureLoadBalancer creates a new load balancer or updates the existing one.
// Returns the status of the balancer (i.e it's public IP address if one exists).
func (cp *CloudProvider) EnsureLoadBalancer(clusterName string, service *api.Service, nodes []*api.Node) (*api.LoadBalancerStatus, error) {
	lbName := GetLoadBalancerName(service)
	lb, err := cp.client.LoadBalancer().GetLoadBalancerByName(context.TODO(), lbName)
	if err != nil && !client.IsNotFound(err) {
		return nil, err
	}
	exists := !client.IsNotFound(err)

	var ssl *SSLConfig
	if needsCerts(service) {
		ports, err := getSSLEnabledPorts(service)
		if err != nil {
			return nil, err
		}
		ssl = NewSSLConfig(lbName, ports, cp)
	}
	subnets := []string{cp.config.LoadBalancer.Subnet1, cp.config.LoadBalancer.Subnet2}
	spec, err := NewLBSpec(service, nodes, subnets, ssl)
	if err != nil {
		glog.Errorf("Failed to create LBSpec: %v", err)
		return nil, err
	}

	sourceCIDRs, err := getLoadBalancerSourceRanges(service)
	if err != nil {
		return nil, err
	}

	glog.V(4).Infof("Ensure load balancer %q called for %q with %d nodes.", spec.Name, service.Name, len(nodes))

	if !exists {
		return cp.createLoadBalancer(context.TODO(), spec, sourceCIDRs)
	}

	// Existing load balancers cannot change subnets. This ensures that the spec matches
	// what the actual load balancer has listed as the subnet ids. If the load balancer
	// was just created then these values would be equal; however, if the load balancer
	// already existed and the default subnet ids changed, then this would ensure
	// we are setting the security rules on the correct subnets.
	spec.Subnets = lb.SubnetIds

	// If the load balancer needs an SSL cert ensure it is present.
	if needsCerts(service) {
		if err := cp.ensureSSLCertificate(context.TODO(), spec, lb); err != nil {
			return nil, err
		}
	}

	err = cp.updateLoadBalancer(context.TODO(), lb, spec, sourceCIDRs)
	if err != nil {
		return nil, err
	}

	status, err := loadBalancerToStatus(lb)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("Successfully ensured load balancer %q", *lb.DisplayName)
	return status, nil
}

func (cp *CloudProvider) updateLoadBalancer(ctx context.Context, lb *loadbalancer.LoadBalancer, spec LBSpec, sourceCIDRs []string) error {
	lbOCID := *lb.Id

	actualBackendSets := lb.BackendSets
	desiredBackendSets := spec.GetBackendSets()
	backendSetActions := getBackendSetChanges(actualBackendSets, desiredBackendSets)

	actualListeners := lb.Listeners
	desiredListeners := spec.GetListeners()
	listenerActions := getListenerChanges(actualListeners, desiredListeners)

	if len(backendSetActions) == 0 && len(listenerActions) == 0 {
		return nil // Nothing to do.
	}

	lbSubnets, err := getSubnets(ctx, spec.Subnets, cp.client.Networking())
	if err != nil {
		return errors.Wrapf(err, "getting load balancer subnets")
	}
	nodeSubnets, err := getSubnetsForNodes(ctx, spec.Nodes, cp.client)
	if err != nil {
		return errors.Wrap(err, "get subnets for nodes")
	}

	actions := sortAndCombineActions(backendSetActions, listenerActions)
	for _, action := range actions {
		switch a := action.(type) {
		case *BackendSetAction:
			err := cp.updateBackendSet(context.TODO(), lbOCID, a, lbSubnets, nodeSubnets)
			if err != nil {
				return errors.Wrap(err, "updating BackendSet")
			}
		case *ListenerAction:
			backendSetName := *a.Listener.DefaultBackendSetName
			var backendPort, healthCheckPort int
			if a.Type() == Delete {
				// If we need to delete the BackendSet then it'll no longer be
				// present in the spec since that's what is desired, so we need
				// to fetch it from the load balancer object.
				bs := lb.BackendSets[backendSetName]
				// FIXME(apryde): panics when no backends.
				backendPort = *bs.Backends[0].Port
				healthCheckPort = *bs.HealthChecker.Port
			} else {
				bs := spec.GetBackendSets()[*a.Listener.DefaultBackendSetName]
				// FIXME(apryde): panics when no backends.
				backendPort = *bs.Backends[0].Port
				healthCheckPort = *bs.HealthChecker.Port
			}

			err := cp.updateListener(ctx, lbOCID, a, backendPort, healthCheckPort, lbSubnets, nodeSubnets, sourceCIDRs)
			if err != nil {
				return errors.Wrap(err, "updating listener")
			}
		}
	}
	return nil
}

func (cp *CloudProvider) updateBackendSet(ctx context.Context, lbOCID string, action *BackendSetAction, lbSubnets, nodeSubnets []*core.Subnet) error {
	sourceCIDRs := []string{}
	listenerPort := 0

	var workRequestID string
	var err error

	bs := action.BackendSet

	glog.V(2).Infof("Applying %q action on backend set %q for lb %q", action.Type(), action.Name(), lbOCID)

	if len(bs.Backends) < 1 {
		return errors.New("no backends provided")
	}
	backendPort := *bs.Backends[0].Port
	healthCheckPort := *bs.HealthChecker.Port

	switch action.Type() {
	case Create:
		err = cp.securityListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, listenerPort, backendPort, healthCheckPort)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().CreateBackendSet(ctx, lbOCID, action.Name(), bs)
	case Update:
		err = cp.securityListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, listenerPort, backendPort, healthCheckPort)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().UpdateBackendSet(ctx, lbOCID, action.Name(), bs)
	case Delete:
		err = cp.securityListManager.Delete(ctx, lbSubnets, nodeSubnets, listenerPort, backendPort, healthCheckPort)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().DeleteBackendSet(ctx, lbOCID, action.Name())
	}

	if err != nil {
		return err
	}

	_, err = cp.client.LoadBalancer().AwaitWorkRequest(ctx, workRequestID)
	if err != nil {
		return err
	}

	return nil
}

func (cp *CloudProvider) updateListener(ctx context.Context, lbOCID string, action *ListenerAction, backendPort int, healthCheckPort int, lbSubnets, nodeSubnets []*core.Subnet, sourceCIDRs []string) error {
	var (
		workRequestID string
		err           error
		l             = action.Listener
		listenerPort  = *l.Port
	)

	glog.V(2).Infof("Applying %q action on listener %q for lb %q", action.Type(), action.Name(), lbOCID)

	switch action.Type() {
	case Create:
		err = cp.securityListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, listenerPort, backendPort, healthCheckPort)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().CreateListener(ctx, lbOCID, action.Name(), l)
	case Update:
		err = cp.securityListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, listenerPort, backendPort, healthCheckPort)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().UpdateListener(ctx, lbOCID, action.Name(), l)
	case Delete:
		err = cp.securityListManager.Delete(ctx, lbSubnets, nodeSubnets, listenerPort, backendPort, healthCheckPort)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().DeleteListener(ctx, lbOCID, action.Name())
	}

	if err != nil {
		return err
	}

	_, err = cp.client.LoadBalancer().AwaitWorkRequest(ctx, workRequestID)
	return err
}

// UpdateLoadBalancer : TODO find out where this is called
func (cp *CloudProvider) UpdateLoadBalancer(clusterName string, service *api.Service, nodes []*api.Node) error {
	name := GetLoadBalancerName(service)
	glog.Infof("Attempting to update load balancer %q", name)

	_, err := cp.EnsureLoadBalancer(clusterName, service, nodes)
	return err
}

// getNodesByIPs returns a slice of Nodes corrosponding to the given IP addresses.
func (cp *CloudProvider) getNodesByIPs(backendIPs []string) ([]*api.Node, error) {
	nodeList, err := cp.NodeLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	ipToNodeLookup := make(map[string]*api.Node)
	for _, node := range nodeList {
		ip := util.NodeInternalIP(node)
		ipToNodeLookup[ip] = node
	}

	var nodes []*api.Node
	for _, ip := range backendIPs {
		node, ok := ipToNodeLookup[ip]
		if !ok {
			return nil, errors.Errorf("node %q was not found by IP %q", node.Name, ip)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it exists,
// returning nil if the load balancer specified either didn't exist or was
// successfully deleted.
func (cp *CloudProvider) EnsureLoadBalancerDeleted(clusterName string, service *api.Service) error {
	name := GetLoadBalancerName(service)
	glog.Infof("Attempting to delete load balancer %q", name)

	lb, err := cp.client.LoadBalancer().GetLoadBalancerByName(context.TODO(), name)
	if err != nil {
		if client.IsNotFound(err) {
			glog.Infof("Could not find load balancer with name %q. Nothing to do.", name)
			return nil
		}
		return errors.Wrapf(err, "get load balancer %q by name", name)
	}

	id := *lb.Id

	nodeIPs := sets.NewString()
	for _, backendSet := range lb.BackendSets {
		for _, backend := range backendSet.Backends {
			nodeIPs.Insert(*backend.IpAddress)
		}
	}
	nodes, err := cp.getNodesByIPs(nodeIPs.List())
	if err != nil {
		return errors.Wrap(err, "fetching nodes by internal ips")
	}

	spec, err := NewLBSpec(service, nodes, []string{cp.config.LoadBalancer.Subnet1, cp.config.LoadBalancer.Subnet2}, nil)
	if err != nil {
		return errors.Wrap(err, "new lb spec")
	}

	lbSubnets, err := getSubnets(context.TODO(), spec.Subnets, cp.client.Networking())
	if err != nil {
		return errors.Wrap(err, "getting subnets for load balancers")
	}
	nodeSubnets, err := getSubnetsForNodes(context.TODO(), nodes, cp.client)
	if err != nil {
		return errors.Wrap(err, "getting subnets for nodes")
	}

	for listenerName, listener := range spec.GetListeners() {
		glog.V(4).Infof("Deleting security rules for listener %q for load balancer %q", listenerName, id)

		backendSetName := *listener.DefaultBackendSetName
		bs, ok := spec.GetBackendSets()[backendSetName]
		if !ok {
			return errors.Errorf("no backend set %q in spec", backendSetName)
		}
		if len(bs.Backends) < 1 {
			return errors.Errorf("backend set %q has no backends", backendSetName)
		}
		backendPort := *bs.Backends[0].Port
		if bs.HealthChecker == nil {
			return errors.Errorf("backend set %q has no health checker", backendSetName)
		}
		healthCheckPort := *bs.HealthChecker.Port

		if err := cp.securityListManager.Delete(context.TODO(), lbSubnets, nodeSubnets, *listener.Port, backendPort, healthCheckPort); err != nil {
			return errors.Wrapf(err, "delete security rules for listener %q on load balancer %q", listenerName, name)
		}
	}

	glog.Infof("Deleting load balancer %q (OCID: %q)", name, id)
	workReqID, err := cp.client.LoadBalancer().DeleteLoadBalancer(context.TODO(), id)
	if err != nil {
		return errors.Wrapf(err, "delete load balancer %q", id)
	}
	_, err = cp.client.LoadBalancer().AwaitWorkRequest(context.TODO(), workReqID)
	if err != nil {
		return errors.Wrapf(err, "awaiting deletion of load balancer %q", name)
	}

	glog.Infof("Deleted load balancer %q (OCID: %q)", name, id)
	return nil
}

// Given an OCI load balancer, return a LoadBalancerStatus
func loadBalancerToStatus(lb *loadbalancer.LoadBalancer) (*api.LoadBalancerStatus, error) {
	if len(lb.IpAddresses) == 0 {
		return nil, errors.Errorf("no ip addresses found for load balancer %q", *lb.DisplayName)
	}

	ingress := []api.LoadBalancerIngress{}
	for _, ip := range lb.IpAddresses {
		ingress = append(ingress, api.LoadBalancerIngress{IP: *ip.IpAddress})
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
