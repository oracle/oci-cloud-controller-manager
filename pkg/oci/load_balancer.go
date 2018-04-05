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
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/loadbalancer"
	"github.com/pkg/errors"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	sets "k8s.io/apimachinery/pkg/util/sets"
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
func (cp *CloudProvider) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (*v1.LoadBalancerStatus, bool, error) {
	name := GetLoadBalancerName(service)
	glog.V(4).Infof("Fetching load balancer with name %q", name)

	lb, err := cp.client.LoadBalancer().GetLoadBalancerByName(ctx, cp.config.Auth.CompartmentOCID, name)
	if err != nil {
		if client.IsNotFound(err) {
			glog.V(2).Infof("Load balancer %q does not exist", name)
			return nil, false, nil
		}

		return nil, false, err
	}

	lbStatus, err := loadBalancerToStatus(lb)
	return lbStatus, (err == nil), err
}

// getSubnets returns a list of Subnet objects for the corrosponding OCIDs.
func getSubnets(ctx context.Context, subnetIDs []string, n client.NetworkingInterface) ([]*core.Subnet, error) {
	subnets := make([]*core.Subnet, len(subnetIDs))
	for i, id := range subnetIDs {
		subnet, err := n.GetSubnet(ctx, id)
		if err != nil {
			return nil, err
		}
		subnets[i] = subnet
	}
	return subnets, nil
}

// getSubnetsForNodes returns the de-duplicated subnets in which the given
// internal IP addresses reside.
func getSubnetsForNodes(ctx context.Context, nodes []*v1.Node, client client.Interface, compartmentID string) ([]*core.Subnet, error) {
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
		vnic, err := client.Compute().GetPrimaryVNICForInstance(ctx, compartmentID, id)
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
func (cp *CloudProvider) readSSLSecret(svc *v1.Service) (string, string, error) {
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
func (cp *CloudProvider) ensureSSLCertificate(ctx context.Context, lb *loadbalancer.LoadBalancer, spec *LBSpec) error {
	name := spec.SSLConfig.Name
	_, err := cp.client.LoadBalancer().GetCertificateByName(ctx, *lb.Id, name)
	if err == nil {
		glog.V(4).Infof("Certificate %q already exists on load balancer %q. Nothing to do.", name, *lb.DisplayName)
		return nil
	}
	if !client.IsNotFound(err) {
		return err
	}

	// Although we iterate here only one certificate is supported at the moment.
	certs, err := spec.Certificates()
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
func (cp *CloudProvider) createLoadBalancer(ctx context.Context, spec *LBSpec) (*v1.LoadBalancerStatus, error) {
	glog.Infof("Attempting to create a new load balancer with name %q", spec.Name)

	// First update the security lists so that if it fails (due to the etag
	// bug or otherwise) we'll retry prior to LB creation.
	lbSubnets, err := getSubnets(ctx, spec.Subnets, cp.client.Networking())
	if err != nil {
		return nil, errors.Wrap(err, "getting subnets for load balancers")
	}
	nodeSubnets, err := getSubnetsForNodes(ctx, spec.nodes, cp.client, cp.config.Auth.CompartmentOCID)
	if err != nil {
		return nil, errors.Wrap(err, "getting subnets for nodes")
	}

	for _, ports := range spec.Ports {
		if err = cp.securityListManager.Update(ctx, lbSubnets, nodeSubnets, spec.SourceCIDRs, nil, ports); err != nil {
			return nil, err
		}
	}

	// Then we create the load balancer and wait for it to be online.
	certs, err := spec.Certificates()
	if err != nil {
		return nil, errors.Wrap(err, "get certificates")
	}
	details := loadbalancer.CreateLoadBalancerDetails{
		CompartmentId: &cp.config.Auth.CompartmentOCID,
		DisplayName:   &spec.Name,
		ShapeName:     &spec.Shape,
		IsPrivate:     &spec.Internal,
		SubnetIds:     spec.Subnets,
		BackendSets:   spec.BackendSets,
		Listeners:     spec.Listeners,
		Certificates:  certs,
	}

	glog.V(4).Infof("CreateLoadBalancerDetails: %#v", details.String())

	wrID, err := cp.client.LoadBalancer().CreateLoadBalancer(ctx, details)
	if err != nil {
		return nil, errors.Wrap(err, "creating load balancer")
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
func (cp *CloudProvider) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	lbName := GetLoadBalancerName(service)

	glog.V(4).Infof("Ensure load balancer %q called for %q with %d nodes.", lbName, service.Name, len(nodes))

	lb, err := cp.client.LoadBalancer().GetLoadBalancerByName(ctx, cp.config.Auth.CompartmentOCID, lbName)
	if err != nil && !client.IsNotFound(err) {
		return nil, err
	}
	exists := !client.IsNotFound(err)

	var ssl *SSLConfig
	if requiresCertificate(service) {
		ports, err := getSSLEnabledPorts(service)
		if err != nil {
			return nil, err
		}
		ssl = NewSSLConfig(lbName, ports, cp)
	}
	subnets := []string{cp.config.LoadBalancer.Subnet1, cp.config.LoadBalancer.Subnet2}
	spec, err := NewLBSpec(service, nodes, subnets, ssl)
	if err != nil {
		glog.Errorf("Failed to derive LBSpec: %+v", err)
		return nil, err
	}

	if !exists {
		return cp.createLoadBalancer(ctx, spec)
	}

	// Existing load balancers cannot change subnets. This ensures that the spec matches
	// what the actual load balancer has listed as the subnet ids. If the load balancer
	// was just created then these values would be equal; however, if the load balancer
	// already existed and the default subnet ids changed, then this would ensure
	// we are setting the security rules on the correct subnets.
	spec.Subnets = lb.SubnetIds

	// If the load balancer needs an SSL cert ensure it is present.
	if requiresCertificate(service) {
		if err := cp.ensureSSLCertificate(ctx, lb, spec); err != nil {
			return nil, errors.Wrap(err, "ensuring ssl certificate")
		}
	}

	if err := cp.updateLoadBalancer(ctx, lb, spec); err != nil {
		return nil, err
	}

	return loadBalancerToStatus(lb)
}

func (cp *CloudProvider) updateLoadBalancer(ctx context.Context, lb *loadbalancer.LoadBalancer, spec *LBSpec) error {
	lbID := *lb.Id

	actualBackendSets := lb.BackendSets
	desiredBackendSets := spec.BackendSets
	backendSetActions := getBackendSetChanges(actualBackendSets, desiredBackendSets)

	actualListeners := lb.Listeners
	desiredListeners := spec.Listeners
	listenerActions := getListenerChanges(actualListeners, desiredListeners)

	if len(backendSetActions) == 0 && len(listenerActions) == 0 {
		return nil // Nothing to do.
	}

	lbSubnets, err := getSubnets(ctx, spec.Subnets, cp.client.Networking())
	if err != nil {
		return errors.Wrapf(err, "getting load balancer subnets")
	}
	nodeSubnets, err := getSubnetsForNodes(ctx, spec.nodes, cp.client, cp.config.Auth.CompartmentOCID)
	if err != nil {
		return errors.Wrap(err, "get subnets for nodes")
	}

	actions := sortAndCombineActions(backendSetActions, listenerActions)
	for _, action := range actions {
		switch a := action.(type) {
		case *BackendSetAction:
			err := cp.updateBackendSet(ctx, lbID, a, lbSubnets, nodeSubnets)
			if err != nil {
				return errors.Wrap(err, "updating BackendSet")
			}
		case *ListenerAction:
			backendSetName := *a.Listener.DefaultBackendSetName
			var ports portSpec
			if a.Type() == Delete {
				// If we need to delete the BackendSet then it'll no longer be
				// present in the spec since that's what is desired, so we need
				// to fetch it from the load balancer object.
				bs := lb.BackendSets[backendSetName]
				ports = portsFromBackendSet(backendSetName, &bs)
			} else {
				ports = spec.Ports[backendSetName]
			}

			err := cp.updateListener(ctx, lbID, a, ports, lbSubnets, nodeSubnets, spec.SourceCIDRs)
			if err != nil {
				return errors.Wrap(err, "updating listener")
			}
		}
	}
	return nil
}

func (cp *CloudProvider) updateBackendSet(ctx context.Context, lbID string, action *BackendSetAction, lbSubnets, nodeSubnets []*core.Subnet) error {
	var (
		sourceCIDRs   = []string{}
		workRequestID string
		err           error
		bs            = action.BackendSet
		ports         = action.Ports
	)

	glog.V(2).Infof("Applying %q action on backend set %q for lb %q (ports=%+v)", action.Type(), action.Name(), lbID, ports)

	switch action.Type() {
	case Create:
		err = cp.securityListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, nil, ports)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().CreateBackendSet(ctx, lbID, action.Name(), bs)
	case Update:
		if err = cp.securityListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, action.OldPorts, ports); err != nil {
			return err
		}
		workRequestID, err = cp.client.LoadBalancer().UpdateBackendSet(ctx, lbID, action.Name(), bs)
	case Delete:
		err = cp.securityListManager.Delete(ctx, lbSubnets, nodeSubnets, ports)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().DeleteBackendSet(ctx, lbID, action.Name())
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

func (cp *CloudProvider) updateListener(ctx context.Context, lbID string, action *ListenerAction, ports portSpec, lbSubnets, nodeSubnets []*core.Subnet, sourceCIDRs []string) error {
	var workRequestID string
	var err error
	listener := action.Listener
	ports.ListenerPort = *listener.Port

	glog.V(2).Infof("Applying %q action on listener %q for lb %q (ports=%+v)", action.Type(), action.Name(), lbID, ports)

	switch action.Type() {
	case Create:
		err = cp.securityListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, nil, ports)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().CreateListener(ctx, lbID, action.Name(), listener)
	case Update:
		err = cp.securityListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, nil, ports)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().UpdateListener(ctx, lbID, action.Name(), listener)
	case Delete:
		err = cp.securityListManager.Delete(ctx, lbSubnets, nodeSubnets, ports)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().DeleteListener(ctx, lbID, action.Name())
	}

	if err != nil {
		return err
	}

	_, err = cp.client.LoadBalancer().AwaitWorkRequest(ctx, workRequestID)
	return err
}

// UpdateLoadBalancer : TODO find out where this is called
func (cp *CloudProvider) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	name := GetLoadBalancerName(service)
	glog.Infof("Attempting to update load balancer %q", name)

	_, err := cp.EnsureLoadBalancer(ctx, clusterName, service, nodes)
	return err
}

// getNodesByIPs returns a slice of Nodes corrosponding to the given IP addresses.
func (cp *CloudProvider) getNodesByIPs(backendIPs []string) ([]*v1.Node, error) {
	nodeList, err := cp.NodeLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	ipToNodeLookup := make(map[string]*v1.Node)
	for _, node := range nodeList {
		ip := util.NodeInternalIP(node)
		ipToNodeLookup[ip] = node
	}

	var nodes []*v1.Node
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
func (cp *CloudProvider) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	name := GetLoadBalancerName(service)
	glog.Infof("Attempting to delete load balancer %q", name)

	lb, err := cp.client.LoadBalancer().GetLoadBalancerByName(ctx, cp.config.Auth.CompartmentOCID, name)
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
	nodeSubnets, err := getSubnetsForNodes(ctx, nodes, cp.client, cp.config.Auth.CompartmentOCID)
	if err != nil {
		return errors.Wrap(err, "getting subnets for nodes")
	}

	lbSubnets, err := getSubnets(ctx, lb.SubnetIds, cp.client.Networking())
	if err != nil {
		return errors.Wrap(err, "getting subnets for load balancers")
	}

	for listenerName, listener := range lb.Listeners {
		backendSetName := *listener.DefaultBackendSetName
		bs, ok := lb.BackendSets[backendSetName]
		if !ok {
			return errors.Errorf("backend set %q missing (loadbalancer=%q)", backendSetName, id) // Should never happen.
		}

		ports := portsFromBackendSet(backendSetName, &bs)
		ports.ListenerPort = *listener.Port

		glog.V(4).Infof("Deleting security rules for listener %q for load balancer %q ports=%+v", listenerName, id, ports)

		if err := cp.securityListManager.Delete(ctx, lbSubnets, nodeSubnets, ports); err != nil {
			return errors.Wrapf(err, "delete security rules for listener %q on load balancer %q", listenerName, name)
		}
	}

	glog.Infof("Deleting load balancer %q (OCID: %q)", name, id)
	workReqID, err := cp.client.LoadBalancer().DeleteLoadBalancer(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "delete load balancer %q", id)
	}
	_, err = cp.client.LoadBalancer().AwaitWorkRequest(ctx, workReqID)
	if err != nil {
		return errors.Wrapf(err, "awaiting deletion of load balancer %q", name)
	}
	glog.Infof("Deleted load balancer %q (OCID: %q)", name, id)

	return nil
}

// Given an OCI load balancer, return a LoadBalancerStatus
func loadBalancerToStatus(lb *loadbalancer.LoadBalancer) (*v1.LoadBalancerStatus, error) {
	if len(lb.IpAddresses) == 0 {
		return nil, errors.Errorf("no ip addresses found for load balancer %q", *lb.DisplayName)
	}

	ingress := []v1.LoadBalancerIngress{}
	for _, ip := range lb.IpAddresses {
		ingress = append(ingress, v1.LoadBalancerIngress{IP: *ip.IpAddress})
	}
	return &v1.LoadBalancerStatus{Ingress: ingress}, nil
}
