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

package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/golang/glog"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"

	baremetal "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/instancemeta"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/util"
)

const (
	// DefaultLoadBalancerPolicy is the default load balancing algorithm when
	// creating a BackendSet.
	DefaultLoadBalancerPolicy = "ROUND_ROBIN"
)

// Interface abstracts the OCI SDK and application specific convenience methods
// for interacting with the OCI API.
type Interface interface {
	BaremetalInterface

	// GetInstanceByNodeName retrieves the baremetal.Instance corresponding or
	// a SearchError if no instance matching the node name is found.
	GetInstanceByNodeName(name string) (*baremetal.Instance, error)
	// GetNodeAddressesForInstance gets the node addresses of an instance based
	// on the given instance OCID.
	GetNodeAddressesForInstance(id string) ([]api.NodeAddress, error)
	// GetAttachedVnicsForInstance returns a slice of AVAILABLE Vnics for a
	// given instance ocid.
	GetAttachedVnicsForInstance(id string) ([]*baremetal.Vnic, error)
	// CreateAndAwaitLoadBalancer creates a load balancer and blocks until data
	// is available or timeout is reached.
	CreateAndAwaitLoadBalancer(name, shape string, subnets []string, internal bool) (*baremetal.LoadBalancer, error)
	// GetLoadBalancerByName gets a load balancer by its DisplayName.
	GetLoadBalancerByName(name string) (*baremetal.LoadBalancer, error)
	// GetCertificateByName gets a certificate by its name.
	GetCertificateByName(loadBalancerID string, name string) (*baremetal.Certificate, error)
	// CreateAndAwaitBackendSet creates the given BackendSet for the given
	// LoadBalancer.
	CreateAndAwaitBackendSet(lb *baremetal.LoadBalancer, bs baremetal.BackendSet) (*baremetal.BackendSet, error)
	// CreateAndAwaitListener creates the given Listener for the given
	// LoadBalancer.
	CreateAndAwaitListener(lb *baremetal.LoadBalancer, listener baremetal.Listener) error
	// CreateAndAwaitCertificate creates a certificate for the given
	// LoadBalancer.
	CreateAndAwaitCertificate(lb *baremetal.LoadBalancer, name string, certificate string, key string) error
	// AwaitWorkRequest blocks until the work request succeeds, fails or if it
	// timesout after exponential backoff.
	AwaitWorkRequest(id string) (*baremetal.WorkRequest, error)
	// GetSubnetsForNodes returns the deduplicated subnets in which the
	// given internal IP addresses reside.
	GetSubnetsForNodes([]*api.Node) ([]*baremetal.Subnet, error)
	// GetSubnets returns the subnets for the given ocids.
	GetSubnets(ocids []string) ([]*baremetal.Subnet, error)
	// GetDefaultSecurityList gets the default SecurityList for the given Subnet
	// by assuming that the default SecurityList is always the oldest (as it is
	// created automatically when the Subnet is created and cannot be deleted).
	GetDefaultSecurityList(subnet *baremetal.Subnet) (*baremetal.SecurityList, error)
}

// BaremetalInterface defines the subset of the baremetal API exposed by the
// client. It is composed into Interface.
type BaremetalInterface interface {
	Validate() error
	// LaunchInstance creates an instance
	// THIS SHOULD ONLY BE USED FOR INTEGRATION TESTING
	LaunchInstance(
		availabilityDomain,
		compartmentID,
		image,
		shape,
		subnetID string,
		opts *baremetal.LaunchInstanceOptions) (*baremetal.Instance, error)
	// TerminateInstance terminates the given instance.
	// THIS SHOULD ONLY BE USED FOR INTEGRATION TESTING
	TerminateInstance(id string, opts *baremetal.IfMatchOptions) error

	GetInstance(id string) (*baremetal.Instance, error)

	GetSubnet(id string) (*baremetal.Subnet, error)

	UpdateSecurityList(id string, opts *baremetal.UpdateSecurityListOptions) (*baremetal.SecurityList, error)

	CreateBackendSet(
		loadBalancerID string,
		name string,
		policy string,
		backends []baremetal.Backend,
		healthChecker *baremetal.HealthChecker,
		sslConfig *baremetal.SSLConfiguration,
		sessionPersistenceConfig *baremetal.SessionPersistenceConfiguration,
		opts *baremetal.LoadBalancerOptions,
	) (workRequestID string, e error)
	UpdateBackendSet(loadBalancerID string, backendSetName string, opts *baremetal.UpdateLoadBalancerBackendSetOptions) (workRequestID string, e error)
	DeleteBackendSet(loadBalancerID string, backendSetName string, opts *baremetal.ClientRequestOptions) (string, error)

	CreateListener(
		loadBalancerID string,
		name string,
		defaultBackendSetName string,
		protocol string,
		port int,
		sslConfig *baremetal.SSLConfiguration,
		opts *baremetal.LoadBalancerOptions,
	) (workRequestID string, e error)
	UpdateListener(loadBalancerID string, listenerName string, opts *baremetal.UpdateLoadBalancerListenerOptions) (workRequestID string, e error)
	DeleteListener(loadBalancerID string, listenerName string, opts *baremetal.ClientRequestOptions) (workRequestID string, e error)

	DeleteLoadBalancer(id string, opts *baremetal.ClientRequestOptions) (string, error)
}

// New creates a new OCI API client.
func New(cfg *Config) (Interface, error) {
	var opts = []baremetal.NewClientOptionsFunc{baremetal.PrivateKeyBytes([]byte(cfg.Auth.PrivateKey)),
		baremetal.Region(cfg.Auth.Region),
		// Kubernetes will handle retries.
		// The current go client will retry requests that are not retryable.
		baremetal.DisableAutoRetries(true)}

	if cfg.Auth.PrivateKeyPassphrase != "" {
		opts = append(opts, baremetal.PrivateKeyPassword(cfg.Auth.PrivateKeyPassphrase))
	}

	// Handles the case where we want to talk to OCI via a proxy.
	ociProxy := os.Getenv("OCI_PROXY")
	trustedCACertPath := os.Getenv("TRUSTED_CA_CERT_PATH")
	if ociProxy != "" || trustedCACertPath != "" {
		transport := http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		if ociProxy != "" {
			glog.Infof("using OCI proxy server: %s", ociProxy)
			proxyURL, err := url.Parse(ociProxy)
			if err != nil {
				return nil, fmt.Errorf("failed to parse OCI proxy url: %s, err: %v", ociProxy, err)
			}
			transport.Proxy = func(req *http.Request) (*url.URL, error) {
				return proxyURL, nil
			}
		}

		if trustedCACertPath != "" {
			glog.Infof("configuring OCI client with a new trusted ca: %s", trustedCACertPath)
			trustedCACert, err := ioutil.ReadFile(trustedCACertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read root certificate: %s, err: %v", trustedCACertPath, err)
			}
			caCertPool := x509.NewCertPool()
			ok := caCertPool.AppendCertsFromPEM(trustedCACert)
			if !ok {
				return nil, fmt.Errorf("failed to parse root certificate: %s", trustedCACertPath)
			}
			transport.TLSClientConfig = &tls.Config{RootCAs: caCertPool}
		}

		opts = append(opts, func(o *baremetal.NewClientOptions) {
			o.Transport = &transport
		})
	}

	ociClient, err := baremetal.NewClient(
		cfg.Auth.UserOCID,
		cfg.Auth.TenancyOCID,
		cfg.Auth.Fingerprint,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	compartmentOCID := cfg.Auth.CompartmentOCID
	if compartmentOCID == "" {
		glog.Info("compartment not supplied in config: attempting to infer from instance metadata")
		metadata, err := instancemeta.New().Get()
		if err != nil {
			return nil, err
		}
		compartmentOCID = metadata.CompartmentOCID
	}

	vcnID := cfg.VCNID
	if vcnID == "" {
		glog.Infof("No vcn provided in cloud provider config. Falling back to looking up VCN via LB subnet.")
		subnet, err := ociClient.GetSubnet(cfg.LoadBalancer.Subnet1)
		if err != nil {
			return nil, fmt.Errorf("failed to get load balancer subnet 1: %v", err)
		}
		vcnID = subnet.VcnID
	}

	return &client{
		Client:            ociClient,
		compartmentID:     compartmentOCID,
		vcnID:             vcnID,
		subnetCache:       cache.NewTTLStore(subnetCacheKeyFn, time.Duration(24)*time.Hour),
		securityListCache: cache.NewTTLStore(securityListKeyFn, time.Duration(24)*time.Hour),
	}, nil
}

// client is a wrapped baremetal.Client with additional methods/props for
// convenience.
type client struct {
	*baremetal.Client

	compartmentID string // the OCID of the Compartment in which the cluster resides.
	vcnID         string // the OCID of the VCN in which the cluster resides.

	subnetCache       cache.Store
	securityListCache cache.Store
}

// Just check we can talk to baremetal before doing anything else (failfast)
// Maybe do some more things like check the compartment we are give is valid....
func (c *client) Validate() error {
	_, err := c.Client.ListAvailabilityDomains(c.compartmentID)
	return err
}

// GetInstanceByNodeName gets the OCID of instance with a display name equal to
// the given node name.
func (c *client) GetInstanceByNodeName(nodeName string) (*baremetal.Instance, error) {
	glog.V(4).Infof("GetInstanceByNodeName(%q) called", nodeName)
	if nodeName == "" {
		return nil, fmt.Errorf("blank nodeName passed to getInstanceByNodeName()")
	}

	opts := &baremetal.ListInstancesOptions{
		DisplayNameListOptions: baremetal.DisplayNameListOptions{
			DisplayName: nodeName,
		},
	}

	r, err := c.ListInstances(c.compartmentID, opts)
	if err != nil {
		return nil, err
	}

	instances := getNonTerminalInstances(r.Instances)
	count := len(instances)

	switch {
	case count == 0:
		// If we can't find an instance by display name fall back to the more
		// expensive search method.
		return c.findInstanceByNodeNameIsVnic(nodeName)
	case count == 1:
		glog.V(4).Infof("getInstanceByNodeName(%q): Got instance %s", nodeName, instances[0].ID)
		return &instances[0], nil
	default:
		return nil, fmt.Errorf("expected one instance with display name '%s' but got %d", nodeName, count)
	}
}

// IsInstanceInTerminalState returns true if the instance is in a terminal state, false otherwise.
func IsInstanceInTerminalState(instance *baremetal.Instance) bool {
	return instance.State == baremetal.ResourceTerminated ||
		instance.State == baremetal.ResourceTerminating ||
		instance.State == "UNKNOWN"
}

func getNonTerminalInstances(instances []baremetal.Instance) []baremetal.Instance {
	var result []baremetal.Instance
	for _, instance := range instances {
		if !IsInstanceInTerminalState(&instance) {
			result = append(result, instance)
		}
	}
	return result
}

// findInstanceByNodeNameIsVnic tries to find the OCI Instance for a given node
// name. It makes the assumption that he node name is resolvable.
// See: https://kubernetes.io/docs/concepts/architecture/nodes/#management
func (c *client) findInstanceByNodeNameIsVnic(nodeName string) (*baremetal.Instance, error) {
	var running []baremetal.Instance
	opts := &baremetal.ListVnicAttachmentsOptions{}
	for {
		vnicAttachments, err := c.ListVnicAttachments(c.compartmentID, opts)
		if err != nil {
			return nil, err
		}
		for _, attachment := range vnicAttachments.Attachments {
			if attachment.State != baremetal.ResourceAttached {
				glog.Warningf("VNIC attachment `%s` for instance `%s` has a state of `%s`", attachment.ID, nodeName, attachment.State)
				continue
			}
			vnic, err := c.GetVnic(attachment.VnicID)
			if err != nil {
				return nil, err
			}

			// Skip VNICs that aren't attached to the cluster's VCN.
			subnet, err := c.GetSubnet(vnic.SubnetID)
			if err != nil {
				return nil, err
			}
			if subnet.VcnID != c.vcnID {
				continue
			}

			if vnic.PublicIPAddress == nodeName ||
				(vnic.HostnameLabel != "" && strings.HasPrefix(nodeName, vnic.HostnameLabel)) {
				instance, err := c.GetInstance(attachment.InstanceID)
				if err != nil {
					return nil, err
				}

				if IsInstanceInTerminalState(instance) {
					glog.Warningf("Instance %q is in state %q which is a terminal state", instance.ID, instance.State)
					continue
				}

				running = append(running, *instance)
			}
		}
		if hasNextPage := SetNextPageOption(vnicAttachments.NextPage, &opts.ListOptions.PageListOptions); !hasNextPage {
			break
		}
	}

	count := len(running)
	switch {
	case count == 0:
		return nil, NewNotFoundError(fmt.Sprintf("could not find instance for node name %q", nodeName))
	case count > 1:
		return nil, fmt.Errorf("expected one instance vnic ip/hostname %q but got %d", nodeName, count)
	}

	return &running[0], nil
}

// GetNodeAddressesForInstance gets the NodeAddress's of a given instance by
// OCID.
func (c *client) GetNodeAddressesForInstance(id string) ([]api.NodeAddress, error) {
	glog.V(4).Infof("GetNodeAddressesForInstance(%q) called", id)
	if id == "" {
		return nil, fmt.Errorf("blank id passed to GetNodeAddressesForInstance()")
	}

	vnics, err := c.GetAttachedVnicsForInstance(id)
	if err != nil {
		return nil, fmt.Errorf("get attached vnics for instance `%s`: %v", id, err)
	}

	addresses := []api.NodeAddress{}
	for _, vnic := range vnics {
		addrs, err := extractNodeAddressesFromVNIC(vnic)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, addrs...)
	}

	return addresses, nil
}

// GetAttachedVnicsForInstance returns a slice of AVAILABLE VNICs for a given
// instance OCID.
func (c *client) GetAttachedVnicsForInstance(id string) ([]*baremetal.Vnic, error) {
	glog.V(4).Infof("GetAttachedVnicsForInstance(%q) called", id)
	if id == "" {
		return nil, fmt.Errorf("blank instance id passed to GetAttachedVnicsForInstance()")
	}

	opts := &baremetal.ListVnicAttachmentsOptions{
		InstanceIDListOptions: baremetal.InstanceIDListOptions{InstanceID: id},
	}
	var vnics []*baremetal.Vnic
	for {
		r, err := c.ListVnicAttachments(c.compartmentID, opts)
		if err != nil {
			return nil, fmt.Errorf("list vnic attachments: %v", err)
		}

		for _, att := range r.Attachments {
			if att.State != baremetal.ResourceAttached {
				glog.Warningf("instance `%s` vnic attachment `%s` is in state %s", id, att.ID, att.State)
				continue
			}

			v, err := c.GetVnic(att.VnicID)
			if err != nil {
				return nil, fmt.Errorf("get vnic %s: %v", att.VnicID, err)
			}

			if v.State != baremetal.ResourceAvailable {
				glog.Warningf("instance `%s` vnic `%s` is in state %s", id, att.VnicID, v.State)
				continue
			}

			vnics = append(vnics, v)
		}

		if hasNexPage := SetNextPageOption(r.NextPage, &opts.ListOptions.PageListOptions); !hasNexPage {
			break
		}
	}
	return vnics, nil
}

// f(n) = 1.25 * f(n - 1) | f(1) = 2
var backoff = wait.Backoff{
	Steps:    15,
	Duration: 2 * time.Second,
	Factor:   1.25,
	Jitter:   0.1,
}

// AwaitWorkRequest keeps polling a OCI work request until it succeeds. If it
// does not succeeded after N retries then return an error.
func (c *client) AwaitWorkRequest(id string) (*baremetal.WorkRequest, error) {
	glog.V(4).Infof("Polling WorkRequest %q...", id)

	var wr *baremetal.WorkRequest
	opts := &baremetal.ClientRequestOptions{}
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		twr, reqErr := c.GetWorkRequest(id, opts)
		if reqErr != nil {
			return false, reqErr
		}

		glog.V(4).Infof("WorkRequest %q state: '%s'", id, twr.State)

		switch twr.State {
		case baremetal.WorkRequestSucceeded:
			wr = twr
			return true, nil
		case baremetal.WorkRequestFailed:
			return false, fmt.Errorf("WorkRequest %q failed: %s", id, twr.Message)
		default:
			return false, nil
		}
	})
	return wr, err
}

// CreateAndAwaitLoadBalancer creates a load balancer and blocks until data is
// available or timeout is reached.
func (c *client) CreateAndAwaitLoadBalancer(name, shape string, subnets []string, internal bool) (*baremetal.LoadBalancer, error) {
	opts := &baremetal.CreateLoadBalancerOptions{
		DisplayNameOptions: baremetal.DisplayNameOptions{
			DisplayName: name,
		},
		IsPrivate: internal,
	}

	req, err := c.CreateLoadBalancer(nil, nil, c.compartmentID, nil, shape, subnets, opts)
	if err != nil {
		return nil, err
	}

	result, err := c.AwaitWorkRequest(req)
	if err != nil {
		return nil, err
	}

	return c.GetLoadBalancer(result.LoadBalancerID, &baremetal.ClientRequestOptions{})
}

// GetLoadBalancerByName fetches a load balancer by its DisplayName.
func (c *client) GetLoadBalancerByName(name string) (*baremetal.LoadBalancer, error) {
	opts := &baremetal.ListOptions{}
	for {
		r, err := c.ListLoadBalancers(c.compartmentID, opts)
		if err != nil {
			return nil, err
		}
		for _, lb := range r.LoadBalancers {
			if lb.DisplayName == name {
				return &lb, nil
			}
		}
		if hasNexPage := SetNextPageOption(r.NextPage, &opts.PageListOptions); !hasNexPage {
			break
		}
	}

	return nil, NewNotFoundError(fmt.Sprintf("could not find load balancer with name '%s'", name))
}

// GetCertificateByName gets a certificate by its name.
func (c *client) GetCertificateByName(lbID, name string) (*baremetal.Certificate, error) {
	opts := &baremetal.ClientRequestOptions{}
	r, err := c.ListCertificates(lbID, opts)
	if err != nil {
		return nil, err
	}

	for _, cert := range r.Certificates {
		if cert.CertificateName == name {
			return &cert, nil
		}
	}

	return nil, NewNotFoundError(fmt.Sprintf("certificate with name %q for load balancer %q not found", name, lbID))
}

// CreateAndAwaitBackendSet creates the given BackendSet for the given
// LoadBalancer.
func (c *client) CreateAndAwaitBackendSet(lb *baremetal.LoadBalancer, bs baremetal.BackendSet) (*baremetal.BackendSet, error) {
	glog.V(2).Infof("Creating BackendSet `%s` for load balancer `%s`", bs.Name, lb.DisplayName)
	wr, err := c.CreateBackendSet(
		lb.ID,
		bs.Name,
		bs.Policy,
		bs.Backends,
		bs.HealthChecker,
		bs.SSLConfig,
		bs.SessionPersistenceConfig,
		nil)
	if err != nil {
		return nil, err
	}

	_, err = c.AwaitWorkRequest(wr)
	if err != nil {
		return nil, err
	}

	return c.GetBackendSet(lb.ID, bs.Name, &baremetal.ClientRequestOptions{})
}

// CreateAndAwaitListener creates the given Listener for the given LoadBalancer.
func (c *client) CreateAndAwaitListener(lb *baremetal.LoadBalancer, listener baremetal.Listener) error {
	glog.V(2).Infof("Creating Listener `%s` for load balancer `%s`", listener.Name, lb.DisplayName)
	wr, err := c.CreateListener(
		lb.ID,
		listener.Name,
		listener.DefaultBackendSetName,
		listener.Protocol,
		listener.Port,
		listener.SSLConfig,
		nil)
	if err != nil {
		return err
	}
	_, err = c.AwaitWorkRequest(wr)
	if err != nil {
		return err
	}
	return nil
}

// CreateAndAwaitCertificate creates a certificate for the given LoadBalancer.
func (c *client) CreateAndAwaitCertificate(lb *baremetal.LoadBalancer, name string, certificate string, key string) error {
	glog.V(4).Infof("Creating Certificate '%s' for load balancer '%s'", name, lb.DisplayName)
	wr, err := c.CreateCertificate(lb.ID, name, "", key, "", certificate, nil)
	if err != nil {
		return err
	}
	_, err = c.AwaitWorkRequest(wr)
	if err != nil {
		return err
	}
	return nil
}

// getSubnetFromCacheByIP checks to see if the given IP is contained by any subnet CIDR block in the subnet cache
// If no hits were found then no subnet and no error will be returned (nil, nil)
func (c *client) getSubnetFromCacheByIP(ip string) (*baremetal.Subnet, error) {
	ipAddr := net.ParseIP(ip)
	for _, subnetItem := range c.subnetCache.List() {
		subnet := subnetItem.(*baremetal.Subnet)
		_, cidr, err := net.ParseCIDR(subnet.CIDRBlock)
		if err != nil {
			// This should never actually error but just in case
			return nil, fmt.Errorf("unable to parse CIDR block %q for subnet %q: %v", subnet.CIDRBlock, subnet.ID, err)
		}

		if cidr.Contains(ipAddr) {
			return subnet, nil
		}
	}
	return nil, nil
}

// GetSubnetsForNodes returns the deduplicated subnets in which the given
// internal IP addresses reside.
func (c *client) GetSubnetsForNodes(nodes []*api.Node) ([]*baremetal.Subnet, error) {
	subnetOCIDs := sets.NewString()
	var subnets []*baremetal.Subnet

	ipSet := sets.NewString()
	for _, node := range nodes {
		ipSet.Insert(util.NodeInternalIP(node))
	}

	for _, node := range nodes {
		// First see if the ip of the node belongs to a subnet in the cache
		ip := util.NodeInternalIP(node)
		subnet, err := c.getSubnetFromCacheByIP(ip)
		if err != nil {
			return nil, err
		}
		if subnet != nil {
			// cache hit
			if !subnetOCIDs.Has(subnet.ID) {
				subnetOCIDs.Insert(subnet.ID)
				subnets = append(subnets, subnet)
			}
			// Since we got a cache hit we don't need to do the expensive query to find the subnet
			continue
		}

		id := util.MapProviderIDToInstanceID(node.Spec.ProviderID)
		vnics, err := c.GetAttachedVnicsForInstance(id)
		if err != nil {
			glog.Errorf("Unable to get vnics for instance %q: %v", id, err)
			return nil, err
		}

		for _, vnic := range vnics {
			if vnic.PrivateIPAddress != "" && ipSet.Has(vnic.PrivateIPAddress) &&
				!subnetOCIDs.Has(vnic.SubnetID) {
				subnet, err := c.GetSubnet(vnic.SubnetID)
				if err != nil {
					glog.Errorf("Unable to get subnet %q for instance %q: %v", vnic.SubnetID, id, err)
					return nil, err
				}

				subnets = append(subnets, subnet)
				subnetOCIDs.Insert(vnic.SubnetID)
			}
		}
	}

	return subnets, nil
}

// GetSubnet returns the subnet for the given ocid from either the cache
// or the API. If there is a cache miss the subnet will be added to the cache.
func (c *client) GetSubnet(id string) (*baremetal.Subnet, error) {
	item, exists, err := c.subnetCache.GetByKey(id)
	if err != nil {
		return nil, err
	}
	if exists {
		return item.(*baremetal.Subnet), nil
	}

	subnet, err := c.Client.GetSubnet(id)
	if err != nil {
		return nil, err
	}

	c.subnetCache.Add(subnet)
	return subnet, err
}

// GetSubnets returns the Subnets corresponding to the given OCIDs.
func (c *client) GetSubnets(ocids []string) ([]*baremetal.Subnet, error) {
	var subnets []*baremetal.Subnet
	for _, ocid := range ocids {
		subnet, err := c.GetSubnet(ocid)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, nil
}

// GetSecurityList returns the security list for the given ocid from either the cache
// or the API. If there is a cache miss the security list will be added to the cache.
func (c *client) GetSecurityList(id string) (*baremetal.SecurityList, error) {
	// Since we do leader election, the CCM leader should be the only CCM updating the security list.
	item, exists, err := c.securityListCache.GetByKey(id)
	if err != nil {
		return nil, err
	}
	if exists {
		secList := item.(*baremetal.SecurityList)
		return secList, nil
	}

	secList, err := c.Client.GetSecurityList(id)
	if err != nil {
		return nil, err
	}

	c.securityListCache.Add(secList)
	return secList, err
}

func (c *client) UpdateSecurityList(id string, opts *baremetal.UpdateSecurityListOptions) (*baremetal.SecurityList, error) {
	sl, err := c.Client.UpdateSecurityList(id, opts)
	if IsConflict(err) {
		secList, getErr := c.Client.GetSecurityList(id)
		if getErr != nil {
			return nil, fmt.Errorf("unable to get sec list %q after conflict: %v", id, getErr)
		}
		c.securityListCache.Add(secList)
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	c.securityListCache.Add(sl)
	return sl, nil
}

// GetDefaultSecurityList gets the default SecurityList for the given Subnet by
// assuming that the default SecurityList is always the oldest (as it is created
// automatically when the Subnet is created and cannot be deleted).
func (c *client) GetDefaultSecurityList(subnet *baremetal.Subnet) (*baremetal.SecurityList, error) {
	var lists []*baremetal.SecurityList
	for _, id := range subnet.SecurityListIDs {
		list, err := c.GetSecurityList(id)
		if err != nil {
			return nil, err
		}
		lists = append(lists, list)
	}

	if len(lists) < 1 {
		return nil, NewNotFoundError(fmt.Sprintf("no SecurityLists found for Subnet '%s'", subnet.ID))
	}

	sort.Slice(lists, func(i, j int) bool {
		return lists[i].TimeCreated.Before(lists[j].TimeCreated.Time)
	})
	return lists[0], nil
}

// extractNodeAddressesFromVNIC extracts Kuberenetes NodeAddresses from a given
// Vnic.
func extractNodeAddressesFromVNIC(vnic *baremetal.Vnic) ([]api.NodeAddress, error) {
	glog.V(4).Infof("extractNodeAddressesFromVNIC() called for %q", vnic.ID)
	if vnic == nil {
		return nil, fmt.Errorf("nil VNIC passed to extractNodeAddressesFromVNIC()")
	}

	addresses := []api.NodeAddress{}

	ip := net.ParseIP(vnic.PrivateIPAddress)
	if vnic.PrivateIPAddress != "" {
		if ip == nil {
			return nil, fmt.Errorf("instance has invalid private address: %q", vnic.PrivateIPAddress)
		}
		addresses = append(addresses, api.NodeAddress{Type: api.NodeInternalIP, Address: ip.String()})
	}

	if vnic.PublicIPAddress != "" {
		ip = net.ParseIP(vnic.PublicIPAddress)
		if ip == nil {
			return nil, fmt.Errorf("instance has invalid public address: %q", vnic.PublicIPAddress)
		}
		addresses = append(addresses, api.NodeAddress{Type: api.NodeExternalIP, Address: ip.String()})
	}

	glog.V(4).Infof("NodeAddresses: %+v ", addresses)

	return addresses, nil
}

func subnetCacheKeyFn(obj interface{}) (string, error) {
	return obj.(*baremetal.Subnet).ID, nil
}

func securityListKeyFn(obj interface{}) (string, error) {
	return obj.(*baremetal.SecurityList).ID, nil
}
