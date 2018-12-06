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

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/loadbalancer"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	sets "k8s.io/apimachinery/pkg/util/sets"
	k8sports "k8s.io/kubernetes/pkg/master/ports"
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
	// specifying the TLS secret to install on the load balancer listeners which
	// have SSL enabled.
	// See: https://kubernetes.io/docs/concepts/services-networking/ingress/#tls
	ServiceAnnotationLoadBalancerTLSSecret = "service.beta.kubernetes.io/oci-load-balancer-tls-secret"

	// ServiceAnnotationLoadBalancerTLSBackendSetSecret is a Service annotation for
	// specifying the generic secret to install on the load balancer listeners which
	// have SSL enabled.
	// See: https://kubernetes.io/docs/concepts/services-networking/ingress/#tls
	ServiceAnnotationLoadBalancerTLSBackendSetSecret = "service.beta.kubernetes.io/oci-load-balancer-tls-backendset-secret"

	// ServiceAnnotationLoadBalancerConnectionIdleTimeout is the annotation used
	// on the service to specify the idle connection timeout.
	ServiceAnnotationLoadBalancerConnectionIdleTimeout = "service.beta.kubernetes.io/oci-load-balancer-connection-idle-timeout"

	// ServiceAnnotaionLoadBalancerSecurityListManagementMode is a Service annotation for
	// specifying the security list managment mode ("All", "Frontend", "None") that configures how security lists are managed by the CCM
	ServiceAnnotaionLoadBalancerSecurityListManagementMode = "service.beta.kubernetes.io/oci-load-balancer-security-list-management-mode"

	// ServiceAnnotationLoadBalancerBEProtocol is a Service annotation for specifying the
	// load balancer listener backend protocol ("TCP", "HTTP").
	// See: https://docs.cloud.oracle.com/iaas/Content/Balance/Concepts/balanceoverview.htm#concepts
	ServiceAnnotationLoadBalancerBEProtocol = "service.beta.kubernetes.io/oci-load-balancer-backend-protocol"
)

// DefaultLoadBalancerPolicy defines the default traffic policy for load
// balancers created by the CCM.
const DefaultLoadBalancerPolicy = "ROUND_ROBIN"

// DefaultLoadBalancerBEProtocol defines the default protocol for load
// balancer listeners created by the CCM.
const DefaultLoadBalancerBEProtocol = "TCP"

const (
	// Fallback value if annotation on service is not set
	lbDefaultShape = "100Mbps"

	lbNodesHealthCheckPath      = "/healthz"
	lbNodesHealthCheckPort      = k8sports.ProxyHealthzPort
	lbNodesHealthCheckProtoHTTP = "HTTP"
	lbNodesHealthCheckProtoTCP  = "TCP"
)

// GetLoadBalancer returns whether the specified load balancer exists, and if
// so, what its status is.
func (cp *CloudProvider) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (*v1.LoadBalancerStatus, bool, error) {
	name := GetLoadBalancerName(service)
	logger := cp.logger.With("loadBalancerName", name)
	logger.Debug("Getting load balancer")

	lb, err := cp.client.LoadBalancer().GetLoadBalancerByName(ctx, cp.config.CompartmentID, name)
	if err != nil {
		if client.IsNotFound(err) {
			logger.Info("Load balancer does not exist")
			return nil, false, nil
		}

		return nil, false, err
	}

	lbStatus, err := loadBalancerToStatus(lb)
	return lbStatus, (err == nil), err
}

// getSubnets returns a list of Subnet objects for the corresponding OCIDs.
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
		ipSet.Insert(NodeInternalIP(node))
	}

	for _, node := range nodes {
		// First see if the IP of the node belongs to a subnet in the cache.
		ip := NodeInternalIP(node)
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

		if node.Spec.ProviderID == "" {
			return nil, errors.Errorf(".spec.providerID was not present on node %q", node.Name)
		}

		id, err := MapProviderIDToInstanceID(node.Spec.ProviderID)
		if err != nil {
			return nil, errors.Wrap(err, "MapProviderIDToInstanceID")
		}

		vnic, err := client.Compute().GetPrimaryVNICForInstance(ctx, compartmentID, id)
		if err != nil {
			return nil, err
		}

		if vnic.PrivateIp != nil && ipSet.Has(*vnic.PrivateIp) &&
			!subnetOCIDs.Has(*vnic.SubnetId) {
			subnet, err := client.Networking().GetSubnet(ctx, *vnic.SubnetId)
			if err != nil {
				return nil, errors.Wrapf(err, "get subnet %q for instance %q", *vnic.SubnetId, id)
			}

			subnets = append(subnets, subnet)
			subnetOCIDs.Insert(*vnic.SubnetId)
		}
	}

	return subnets, nil
}

// readSSLSecret returns the certificate and private key from a Kubernetes TLS
// private key Secret.
func (cp *CloudProvider) readSSLSecret(ns, name string) (*certificateData, error) {
	secret, err := cp.kubeclient.CoreV1().Secrets(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	var ok bool
	var cacert, cert, key, pass []byte
	cacert = secret.Data[SSLCAFileName]
	if cert, ok = secret.Data[SSLCertificateFileName]; !ok {
		return nil, errors.Errorf("%s not found in secret %s/%s", SSLCertificateFileName, ns, name)
	}
	if key, ok = secret.Data[SSLPrivateKeyFileName]; !ok {
		return nil, errors.Errorf("%s not found in secret %s/%s", SSLPrivateKeyFileName, ns, name)
	}
	pass = secret.Data[SSLPassphrase]
	return &certificateData{CACert: cacert, PublicCert: cert, PrivateKey: key, Passphrase: pass}, nil
}

// ensureSSLCertificate creates a OCI SSL certificate to the given load
// balancer, if it doesn't already exist.
func (cp *CloudProvider) ensureSSLCertificates(ctx context.Context, lb *loadbalancer.LoadBalancer, spec *LBSpec) error {
	logger := cp.logger.With("loadBalancerID", *lb.Id)
	// Get all required certificates
	certs, err := spec.Certificates()
	if err != nil {
		return err
	}

	for _, cert := range certs {
		if _, ok := lb.Certificates[*cert.CertificateName]; !ok {
			logger = cp.logger.With("certificateName", *cert.CertificateName)
			wrID, err := cp.client.LoadBalancer().CreateCertificate(ctx, *lb.Id, cert)
			if err != nil {
				return err
			}
			_, err = cp.client.LoadBalancer().AwaitWorkRequest(ctx, wrID)
			if err != nil {
				return err
			}

			logger.Info("Certificate created")
		}
	}
	return nil
}

// createLoadBalancer creates a new OCI load balancer based on the given spec.
func (cp *CloudProvider) createLoadBalancer(ctx context.Context, spec *LBSpec) (*v1.LoadBalancerStatus, error) {
	logger := cp.logger.With("loadBalancerName", spec.Name)
	logger.Info("Attempting to create a new load balancer")

	// First update the security lists so that if it fails (due to the etag
	// bug or otherwise) we'll retry prior to LB creation.
	lbSubnets, err := getSubnets(ctx, spec.Subnets, cp.client.Networking())
	if err != nil {
		return nil, errors.Wrap(err, "getting subnets for load balancers")
	}
	nodeSubnets, err := getSubnetsForNodes(ctx, spec.nodes, cp.client, cp.config.CompartmentID)
	if err != nil {
		return nil, errors.Wrap(err, "getting subnets for nodes")
	}

	// Then we create the load balancer and wait for it to be online.
	certs, err := spec.Certificates()
	if err != nil {
		return nil, errors.Wrap(err, "get certificates")
	}
	details := loadbalancer.CreateLoadBalancerDetails{
		CompartmentId: &cp.config.CompartmentID,
		DisplayName:   &spec.Name,
		ShapeName:     &spec.Shape,
		IsPrivate:     &spec.Internal,
		SubnetIds:     spec.Subnets,
		BackendSets:   spec.BackendSets,
		Listeners:     spec.Listeners,
		Certificates:  certs,
	}

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

	logger.With("loadBalancerID", *lb.Id).Info("Load balancer created")
	status, err := loadBalancerToStatus(lb)
	if status != nil && len(status.Ingress) > 0 {
		// If the LB is successfully provisioned then open lb/node subnet seclists egress/ingress.
		for _, ports := range spec.Ports {
			if err = spec.securityListManager.Update(ctx, lbSubnets, nodeSubnets, spec.SourceCIDRs, nil, ports); err != nil {
				return nil, err
			}
		}
	}

	return status, err

}

// EnsureLoadBalancer creates a new load balancer or updates the existing one.
// Returns the status of the balancer (i.e it's public IP address if one exists).
func (cp *CloudProvider) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	lbName := GetLoadBalancerName(service)
	logger := cp.logger.With("loadbalancerName", lbName, "serviceName", service.Name)
	logger.With("nodes", len(nodes)).Info("Ensuring load balancer")

	lb, err := cp.client.LoadBalancer().GetLoadBalancerByName(ctx, cp.config.CompartmentID, lbName)
	if err != nil && !client.IsNotFound(err) {
		return nil, err
	}
	exists := !client.IsNotFound(err)

	var sslConfig *SSLConfig
	if requiresCertificate(service) {
		ports, err := getSSLEnabledPorts(service)
		if err != nil {
			return nil, err
		}
		secretListenerString := service.Annotations[ServiceAnnotationLoadBalancerTLSSecret]
		secretBackendSetString := service.Annotations[ServiceAnnotationLoadBalancerTLSBackendSetSecret]
		sslConfig = NewSSLConfig(secretListenerString, secretBackendSetString, ports, cp)
	}
	var subnets []string
	if cp.config.LoadBalancer.Subnet2 != "" {
		subnets = []string{cp.config.LoadBalancer.Subnet1, cp.config.LoadBalancer.Subnet2}
	} else {
		subnets = []string{cp.config.LoadBalancer.Subnet1}
	}
	spec, err := NewLBSpec(service, nodes, subnets, sslConfig, cp.securityListManagerFactory)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to derive LBSpec")
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
		if err := cp.ensureSSLCertificates(ctx, lb, spec); err != nil {
			return nil, errors.Wrap(err, "ensuring ssl certificates")
		}
	}

	if err := cp.updateLoadBalancer(ctx, lb, spec); err != nil {
		return nil, err
	}

	return loadBalancerToStatus(lb)
}

func (cp *CloudProvider) updateLoadBalancer(ctx context.Context, lb *loadbalancer.LoadBalancer, spec *LBSpec) error {
	lbID := *lb.Id

	logger := cp.logger.With("loadBalancerID", lbID, "compartmentID", cp.config.CompartmentID)

	actualBackendSets := lb.BackendSets
	desiredBackendSets := spec.BackendSets
	backendSetActions := getBackendSetChanges(logger, actualBackendSets, desiredBackendSets)

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
	nodeSubnets, err := getSubnetsForNodes(ctx, spec.nodes, cp.client, cp.config.CompartmentID)
	if err != nil {
		return errors.Wrap(err, "get subnets for nodes")
	}

	actions := sortAndCombineActions(logger, backendSetActions, listenerActions)
	for _, action := range actions {
		switch a := action.(type) {
		case *BackendSetAction:
			err := cp.updateBackendSet(ctx, lbID, a, lbSubnets, nodeSubnets, spec.securityListManager)
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
				ports = portsFromBackendSet(logger, backendSetName, &bs)
			} else {
				ports = spec.Ports[backendSetName]
			}

			err := cp.updateListener(ctx, lbID, a, ports, lbSubnets, nodeSubnets, spec.SourceCIDRs, spec.securityListManager)
			if err != nil {
				return errors.Wrap(err, "updating listener")
			}
		}
	}
	return nil
}

func (cp *CloudProvider) updateBackendSet(ctx context.Context, lbID string, action *BackendSetAction, lbSubnets, nodeSubnets []*core.Subnet, secListManager securityListManager) error {
	var (
		sourceCIDRs   = []string{}
		workRequestID string
		err           error
		bs            = action.BackendSet
		ports         = action.Ports
	)

	cp.logger.With(
		"actionType", action.Type(),
		"backendSetName", action.Name(),
		"ports", ports,
		"loadBalancerID", lbID).Info("Applying action on backend set")

	switch action.Type() {
	case Create:
		err = secListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, nil, ports)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().CreateBackendSet(ctx, lbID, action.Name(), bs)
	case Update:
		if err = secListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, action.OldPorts, ports); err != nil {
			return err
		}
		workRequestID, err = cp.client.LoadBalancer().UpdateBackendSet(ctx, lbID, action.Name(), bs)
	case Delete:
		err = secListManager.Delete(ctx, lbSubnets, nodeSubnets, ports)
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

func (cp *CloudProvider) updateListener(ctx context.Context, lbID string, action *ListenerAction, ports portSpec, lbSubnets, nodeSubnets []*core.Subnet, sourceCIDRs []string, secListManager securityListManager) error {
	var workRequestID string
	var err error
	listener := action.Listener
	ports.ListenerPort = *listener.Port

	cp.logger.With(
		"actionType", action.Type(),
		"backendSetName", action.Name(),
		"ports", ports,
		"loadBalancerID", lbID).Info("Applying action on listener")

	switch action.Type() {
	case Create:
		err = secListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, nil, ports)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().CreateListener(ctx, lbID, action.Name(), listener)
	case Update:
		err = secListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, nil, ports)
		if err != nil {
			return err
		}

		workRequestID, err = cp.client.LoadBalancer().UpdateListener(ctx, lbID, action.Name(), listener)
	case Delete:
		err = secListManager.Delete(ctx, lbSubnets, nodeSubnets, ports)
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
	cp.logger.With("loadbalancerName", name).Info("Updating load balancer")

	_, err := cp.EnsureLoadBalancer(ctx, clusterName, service, nodes)
	return err
}

// getNodesByIPs returns a slice of Nodes corresponding to the given IP addresses.
func (cp *CloudProvider) getNodesByIPs(backendIPs []string) ([]*v1.Node, error) {
	nodeList, err := cp.NodeLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	ipToNodeLookup := make(map[string]*v1.Node)
	for _, node := range nodeList {
		ip := NodeInternalIP(node)
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
	logger := cp.logger.With("loadbalancerName", name)
	logger.Debug("Attempting to delete load balancer")

	lb, err := cp.client.LoadBalancer().GetLoadBalancerByName(ctx, cp.config.CompartmentID, name)
	if err != nil {
		if client.IsNotFound(err) {
			logger.Info("Could not find load balancer. Nothing to do.")
			return nil
		}
		return errors.Wrapf(err, "get load balancer %q by name", name)
	}

	id := *lb.Id
	logger = logger.With("loadBalancerID", id)

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
	nodeSubnets, err := getSubnetsForNodes(ctx, nodes, cp.client, cp.config.CompartmentID)
	if err != nil {
		return errors.Wrap(err, "getting subnets for nodes")
	}

	lbSubnets, err := getSubnets(ctx, lb.SubnetIds, cp.client.Networking())
	if err != nil {
		return errors.Wrap(err, "getting subnets for load balancers")
	}

	securityListManager := cp.securityListManagerFactory(
		service.Annotations[ServiceAnnotaionLoadBalancerSecurityListManagementMode])

	for listenerName, listener := range lb.Listeners {
		backendSetName := *listener.DefaultBackendSetName
		bs, ok := lb.BackendSets[backendSetName]
		if !ok {
			return errors.Errorf("backend set %q missing (loadbalancer=%q)", backendSetName, id) // Should never happen.
		}

		ports := portsFromBackendSet(cp.logger, backendSetName, &bs)
		ports.ListenerPort = *listener.Port

		logger.With("listenerName", listenerName, "ports", ports).Debug("Deleting security rules for listener")

		if err := securityListManager.Delete(ctx, lbSubnets, nodeSubnets, ports); err != nil {
			return errors.Wrapf(err, "delete security rules for listener %q on load balancer %q", listenerName, name)
		}
	}

	logger.Info("Deleting load balancer")

	workReqID, err := cp.client.LoadBalancer().DeleteLoadBalancer(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "delete load balancer %q", id)
	}
	_, err = cp.client.LoadBalancer().AwaitWorkRequest(ctx, workReqID)
	if err != nil {
		return errors.Wrapf(err, "awaiting deletion of load balancer %q", name)
	}

	logger.Info("Deleted load balancer")

	return nil
}

// Given an OCI load balancer, return a LoadBalancerStatus
func loadBalancerToStatus(lb *loadbalancer.LoadBalancer) (*v1.LoadBalancerStatus, error) {
	if len(lb.IpAddresses) == 0 {
		return nil, errors.Errorf("no ip addresses found for load balancer %q", *lb.DisplayName)
	}

	ingress := []v1.LoadBalancerIngress{}
	for _, ip := range lb.IpAddresses {
		if ip.IpAddress == nil {
			continue // should never happen but appears to when EnsureLoadBalancer is called with 0 nodes.
		}
		ingress = append(ingress, v1.LoadBalancerIngress{IP: *ip.IpAddress})
	}
	return &v1.LoadBalancerStatus{Ingress: ingress}, nil
}
