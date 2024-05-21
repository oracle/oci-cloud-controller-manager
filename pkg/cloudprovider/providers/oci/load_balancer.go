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
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	k8sports "k8s.io/kubernetes/pkg/cluster/ports"
	"k8s.io/utils/pointer"

	providercfg "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"github.com/oracle/oci-go-sdk/v65/networkloadbalancer"
	"github.com/pkg/errors"
)

// Defines the traffic policy for load balancers created by the CCM.
const (
	DefaultLoadBalancerPolicy            = "ROUND_ROBIN"
	RoundRobinLoadBalancerPolicy         = "ROUND_ROBIN"
	LeastConnectionsLoadBalancerPolicy   = "LEAST_CONNECTIONS"
	IPHashLoadBalancerPolicy             = "IP_HASH"
	DefaultNetworkLoadBalancerPolicy     = "FIVE_TUPLE"
	NetworkLoadBalancingPolicyTwoTuple   = "TWO_TUPLE"
	NetworkLoadBalancingPolicyThreeTuple = "THREE_TUPLE"
	NetworkLoadBalancingPolicyFiveTuple  = "FIVE_TUPLE"
)

var LbOperationAlreadyExists = errors.New("An operation for the service is already in progress.")

// DefaultLoadBalancerBEProtocol defines the default protocol for load
// balancer listeners created by the CCM.
const DefaultLoadBalancerBEProtocol = "TCP"

// DefaultNetworkLoadBalancerListenerProtocol defines the default protocol for network load
// balancer listeners created by the CCM.
const DefaultNetworkLoadBalancerListenerProtocol = "TCP"

// MaxNsgPerVnic is the maximum number of NSGs that can be attached to a vnic
// https://docs.oracle.com/en-us/iaas/Content/General/Concepts/servicelimits.htm#nsg_limits
const MaxNsgPerVnic = 5

const (
	// Fallback value if annotation on service is not set
	lbDefaultShape = "100Mbps"

	lbNodesHealthCheckPath  = "/healthz"
	lbNodesHealthCheckPort  = k8sports.ProxyHealthzPort
	lbNodesHealthCheckProto = "HTTP"

	// default connection idle timeout per protocol
	// https://docs.cloud.oracle.com/en-us/iaas/Content/Balance/Reference/connectionreuse.htm#ConnectionConfiguration
	lbConnectionIdleTimeoutTCP       = 300
	lbConnectionIdleTimeoutHTTP      = 60
	flexible                         = "flexible"
	lbLifecycleStateActive           = "ACTIVE"
	lbMaximumNetworkSecurityGroupIds = 5
	excludeBackendFromLBLabel        = "node.kubernetes.io/exclude-from-external-load-balancers"

	// Service Account Token expiration in seconds
	serviceAccountTokenExpiry = 21600 // 6 Hours
)

// Protects security rule addition against update by multiple LBs in parallel
var updateRulesMutex sync.Mutex

// CloudLoadBalancerProvider is an implementation of the cloud-provider struct
type CloudLoadBalancerProvider struct {
	client       client.Interface
	lbClient     client.GenericLoadBalancerInterface
	logger       *zap.SugaredLogger
	metricPusher *metrics.MetricPusher
	config       *providercfg.Config
}

func (cp *CloudProvider) getLoadBalancerProvider(ctx context.Context, svc *v1.Service) (CloudLoadBalancerProvider, error) {
	lbType := getLoadBalancerType(svc)
	name := GetLoadBalancerName(svc)
	var serviceAccountToken *authv1.TokenRequest
	var err error

	logger := cp.logger.With("loadBalancerName", name, "loadBalancerType", lbType)
	logger.Debug("Getting load balancer provider")

	if sa, useWI := svc.Annotations[ServiceAnnotationServiceAccountName]; useWI && sa == "" { // When using Workload Identity
		return CloudLoadBalancerProvider{}, errors.New("Error fetching service account, empty string provided via " + ServiceAnnotationServiceAccountName)
	} else if useWI {
		serviceAccountToken, err = cp.getServiceAccountTokenIfSet(ctx, svc)
		if err != nil {
			return CloudLoadBalancerProvider{}, errors.New("Unable to get service account token. Error:" + err.Error())
		}
	}

	lbClient := cp.client.LoadBalancer(logger, lbType, cp.config.Auth.TenancyID, serviceAccountToken)
	if lbClient == nil {
		return CloudLoadBalancerProvider{}, errors.New(fmt.Sprintf("Error creating Workload Identity based %s Client. Perhaps you are using an OKE BASIC_CLUSTER?", lbType))
	}
	return CloudLoadBalancerProvider{
		client:       cp.client,
		lbClient:     lbClient,
		logger:       cp.logger,
		metricPusher: cp.metricPusher,
		config:       cp.config,
	}, nil
}

// serviceNotExistsOrDeleted returns true if service has stopped existing or has been marked as Deleted
func (cp *CloudProvider) serviceDeletedOrDoesNotExist(ctx context.Context, svc *v1.Service) (bool, error) {
	service, err := cp.kubeclient.CoreV1().Services(svc.Namespace).Get(ctx, svc.Name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {
		return true, nil
	}
	if err != nil {
		return true, errors.New("Unable to check if service still exists. Error:" + err.Error())
	}
	if service.DeletionTimestamp != nil {
		return true, nil
	}
	return false, nil
}

var ServiceAccountTokenExpiry = int64(serviceAccountTokenExpiry)

// Use Worker Identity RP based Client based on annotation: "oke.oci.oraclecloud.com/use-service-account"
// if found.
func (cp *CloudProvider) getServiceAccountTokenIfSet(ctx context.Context, svc *v1.Service) (*authv1.TokenRequest, error) {
	_, err := cp.ServiceAccountLister.ServiceAccounts(svc.Namespace).Get(svc.Annotations[ServiceAnnotationServiceAccountName])
	if err != nil {
		return nil, err
	}

	tokenRequest := authv1.TokenRequest{Spec: authv1.TokenRequestSpec{ExpirationSeconds: &ServiceAccountTokenExpiry}}

	serviceAccountTokenRequest, err := cp.kubeclient.CoreV1().ServiceAccounts(svc.Namespace).CreateToken(ctx, svc.Annotations[ServiceAnnotationServiceAccountName], &tokenRequest, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return serviceAccountTokenRequest, nil
}

// GetLoadBalancerName returns the name of the loadbalancer
func (cp *CloudProvider) GetLoadBalancerName(ctx context.Context, clusterName string, service *v1.Service) string {
	return GetLoadBalancerName(service)
}

// GetLoadBalancer returns whether the specified load balancer exists, and if
// so, what its status is.
func (cp *CloudProvider) GetLoadBalancer(ctx context.Context, clusterName string, service *v1.Service) (*v1.LoadBalancerStatus, bool, error) {
	name := cp.GetLoadBalancerName(ctx, clusterName, service)
	logger := cp.logger.With("loadBalancerName", name, "loadBalancerType", getLoadBalancerType(service))
	if sa, useWI := service.Annotations[ServiceAnnotationServiceAccountName]; useWI { // When using Workload Identity
		logger = logger.With("serviceAccount", sa, "nameSpace", service.Namespace)
	}
	logger.Debug("Getting load balancer")

	lbProvider, err := cp.getLoadBalancerProvider(ctx, service)
	if err != nil {
		return nil, false, errors.Wrap(err, "Unable to get Load Balancer Client.")
	}
	lb, err := lbProvider.lbClient.GetLoadBalancerByName(ctx, cp.config.CompartmentID, name)
	if err != nil {
		if client.IsNotFound(err) {
			logger.Info("Load balancer does not exist")
			return nil, false, nil
		}

		return nil, false, err
	}

	lbStatus, err := loadBalancerToStatus(lb)
	return lbStatus, err == nil, err
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

// getReservedIpOcidByIpAddress returns the OCID of public reserved IP if it is in Available state.
func getReservedIpOcidByIpAddress(ctx context.Context, ipAddress string, n client.NetworkingInterface) (*string, error) {
	publicIp, err := n.GetPublicIpByIpAddress(ctx, ipAddress)
	if err != nil {
		return nil, err
	}
	if publicIp.LifecycleState != core.PublicIpLifecycleStateAvailable {
		return nil, errors.Errorf("The IP address provided is not available for use.")
	}
	return publicIp.Id, nil
}

// getSubnetsForNodes returns the de-duplicated subnets in which the given
// internal IP addresses reside.
func getSubnetsForNodes(ctx context.Context, nodes []*v1.Node, client client.Interface) ([]*core.Subnet, error) {
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

		id, err := MapProviderIDToResourceID(node.Spec.ProviderID)
		if err != nil {
			return nil, errors.Wrap(err, "MapProviderIDToResourceID")
		}

		compartmentID, ok := node.Annotations[CompartmentIDAnnotation]
		if !ok {
			return nil, errors.Errorf("%q annotation not present on node %q", CompartmentIDAnnotation, node.Name)
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
	secret, err := cp.kubeclient.CoreV1().Secrets(ns).Get(context.Background(), name, metav1.GetOptions{})
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
func (clb *CloudLoadBalancerProvider) ensureSSLCertificates(ctx context.Context, lb *client.GenericLoadBalancer, spec *LBSpec) error {
	logger := clb.logger.With("loadBalancerID", *lb.Id)
	// Get all required certificates
	certs, err := spec.Certificates()
	if err != nil {
		return err
	}

	for _, cert := range certs {
		if _, ok := lb.Certificates[*cert.CertificateName]; !ok {
			logger = clb.logger.With("certificateName", *cert.CertificateName)
			wrID, err := clb.lbClient.CreateCertificate(ctx, *lb.Id, &cert)
			if err != nil {
				return err
			}
			logger.With("workRequestID", wrID).Info("Await workrequest for create certificate")
			_, err = clb.lbClient.AwaitWorkRequest(ctx, wrID)
			if err != nil {
				return err
			}

			logger.Info("Workrequest for certificate create succeeded")
		}
	}
	return nil
}

// createLoadBalancer creates a new OCI load balancer based on the given spec.
func (clb *CloudLoadBalancerProvider) createLoadBalancer(ctx context.Context, spec *LBSpec) (lbStatus *v1.LoadBalancerStatus, lbOCID string, err error) {
	logger := clb.logger.With("loadBalancerName", spec.Name, "loadBalancerType", getLoadBalancerType(spec.service))
	logger.Info("Attempting to create a new load balancer")

	// First update the security lists so that if it fails (due to the etag
	// bug or otherwise) we'll retry prior to LB creation.
	lbSubnets, err := getSubnets(ctx, spec.Subnets, clb.client.Networking())
	if err != nil {
		return nil, "", errors.Wrap(err, "getting subnets for load balancers")
	}
	nodeSubnets, err := getSubnetsForNodes(ctx, spec.nodes, clb.client)
	if err != nil {
		return nil, "", errors.Wrap(err, "getting subnets for nodes")
	}

	// Then we create the load balancer and wait for it to be online.
	certs, err := spec.Certificates()
	if err != nil {
		return nil, "", errors.Wrap(err, "get certificates")
	}

	details := client.GenericCreateLoadBalancerDetails{
		CompartmentId:           &clb.config.CompartmentID,
		DisplayName:             &spec.Name,
		ShapeName:               &spec.Shape,
		IsPrivate:               &spec.Internal,
		SubnetIds:               spec.Subnets,
		BackendSets:             spec.BackendSets,
		Listeners:               spec.Listeners,
		Certificates:            certs,
		NetworkSecurityGroupIds: spec.NetworkSecurityGroupIds,
		FreeformTags:            spec.FreeformTags,
		DefinedTags:             spec.DefinedTags,
	}

	if spec.Shape == flexible {
		details.ShapeDetails = &client.GenericShapeDetails{
			MinimumBandwidthInMbps: spec.FlexMin,
			MaximumBandwidthInMbps: spec.FlexMax,
		}
	}

	if spec.LoadBalancerIP != "" {
		reservedIpOCID, err := getReservedIpOcidByIpAddress(ctx, spec.LoadBalancerIP, clb.client.Networking())
		if err != nil {
			return nil, "", err
		}

		details.ReservedIps = []client.GenericReservedIp{
			client.GenericReservedIp{
				Id: reservedIpOCID,
			},
		}
	}

	serviceUid := fmt.Sprintf("%s", spec.service.UID)
	wrID, err := clb.lbClient.CreateLoadBalancer(ctx, &details, &serviceUid)
	if err != nil {
		return nil, "", errors.Wrap(err, "creating load balancer")
	}
	logger.With("workRequestID", wrID).Info("Await workrequest for create loadbalancer")
	wr, err := clb.lbClient.AwaitWorkRequest(ctx, wrID)
	if err != nil {
		return nil, "", errors.Wrap(err, "awaiting load balancer")
	}
	logger.With("workRequestID", wrID).Info("Workrequest for create loadbalancer succeeded")

	if wr.LoadBalancerId == nil {
		return nil, "", errors.New("Could not get LoadBalancerId from workrequest")
	}

	lb, err := clb.lbClient.GetLoadBalancer(ctx, *wr.LoadBalancerId)
	if err != nil {
		return nil, "", errors.Wrapf(err, "get load balancer %q", *wr.LoadBalancerId)
	}

	logger.With("loadBalancerID", *lb.Id).Info("Load balancer created")
	status, err := loadBalancerToStatus(lb)

	if status != nil && len(status.Ingress) > 0 {
		// If the LB is successfully provisioned then open lb/node subnet seclists egress/ingress.
		// Security List Updates take place in a Global Critical Section
		if err = updateSecurityListsInCriticalSection(ctx, spec, lbSubnets, nodeSubnets); err != nil {
			return nil, "", err
		}
	}
	if lb.Id != nil {
		lbOCID = *lb.Id
	}
	return status, lbOCID, err
}

// getNodeFilter extracts the node filter based on load balancer type.
// if no selector is defined then an all label selector object is returned to match everything.
func getNodeFilter(svc *v1.Service) (labels.Selector, error) {
	lbType := getLoadBalancerType(svc)

	var labelSelector string

	switch lbType {
	case NLB:
		labelSelector = svc.Annotations[ServiceAnnotationNetworkLoadBalancerNodeFilter]
	default:
		labelSelector = svc.Annotations[ServiceAnnotationLoadBalancerNodeFilter]
	}

	if labelSelector == "" {
		return labels.Everything(), nil
	}

	return labels.Parse(labelSelector)
}

// filterNodes based on the label selector, if present, and returns the set of nodes
// that should be backends in the load balancer.
func filterNodes(svc *v1.Service, nodes []*v1.Node) ([]*v1.Node, error) {

	selector, err := getNodeFilter(svc)
	if err != nil {
		return nil, err
	}

	var filteredNodes []*v1.Node
	for _, n := range nodes {
		if selector.Matches(labels.Set(n.GetLabels())) {
			filteredNodes = append(filteredNodes, n)
		}
	}

	return filteredNodes, nil
}

// EnsureLoadBalancer creates a new load balancer or updates the existing one.
// Returns the status of the balancer (i.e it's public IP address if one exists).
func (cp *CloudProvider) EnsureLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, clusterNodes []*v1.Node) (*v1.LoadBalancerStatus, error) {
	startTime := time.Now()
	lbName := GetLoadBalancerName(service)
	loadBalancerType := getLoadBalancerType(service)
	logger := cp.logger.With("loadBalancerName", lbName, "serviceName", service.Name, "loadBalancerType", loadBalancerType, "serviceUid", service.UID)
	if sa, useWI := service.Annotations[ServiceAnnotationServiceAccountName]; useWI { // When using Workload Identity
		logger = logger.With("serviceAccount", sa, "namespace", service.Namespace)
	}

	if deleted, err := cp.serviceDeletedOrDoesNotExist(ctx, service); deleted {
		if err != nil {
			logger.With(zap.Error(err)).Error("Failed to check if service exists")
			return nil, errors.Wrap(err, "Failed to check service status")
		}
		logger.Info("Service already deleted or no more exists")
		return nil, errors.New("Service already deleted or no more exists")
	}
	loadBalancerService := fmt.Sprintf("%s/%s", service.Namespace, service.Name)
	if acquired := cp.lbLocks.TryAcquire(loadBalancerService); !acquired {
		logger.Error("Could not acquire lock for Ensuring Load Balancer")
		return nil, LbOperationAlreadyExists
	}
	defer cp.lbLocks.Release(loadBalancerService)

	nodes, err := filterNodes(service, clusterNodes)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to filter nodes with label selector")
		return nil, err
	}

	logger.With("nodes", len(nodes)).Info("Ensuring load balancer")

	dimensionsMap := make(map[string]string)

	var errorType string
	var lbMetricDimension string
	var nsgMetricDimension string

	lbProvider, err := cp.getLoadBalancerProvider(ctx, service)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to get Load Balancer Client.")
	}
	lb, err := lbProvider.lbClient.GetLoadBalancerByName(ctx, cp.config.CompartmentID, lbName)
	if err != nil && !client.IsNotFound(err) {
		logger.With(zap.Error(err)).Error("Failed to get loadbalancer by name")
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		dimensionsMap[metrics.ResourceOCIDDimension] = lbName
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), time.Since(startTime).Seconds(), dimensionsMap)
	}
	lbExists := !client.IsNotFound(err)
	lbOCID := ""
	if lb != nil && lb.Id != nil {
		lbOCID = *lb.Id
	} else {
		// if the LB does not exist already use the k8s service UID for reference
		// in logs and metrics
		lbOCID = lbName
	}

	logger = logger.With("loadBalancerID", lbOCID, "loadBalancerType", getLoadBalancerType(service))
	dimensionsMap[metrics.ResourceOCIDDimension] = lbOCID

	// Checks if we have pending work requests before processing the LoadBalancer further
	// Will error out if any in-progress work request are present for the LB
	if lb != nil && lb.Id != nil {
		err = cp.checkPendingLBWorkRequests(ctx, logger, lbProvider, lb, service, startTime)
		if err != nil {
			return nil, err
		}
	}

	var sslConfig *SSLConfig
	if requiresCertificate(service) {
		ports, err := getSSLEnabledPorts(service)
		if err != nil {
			logger.With(zap.Error(err)).Error("Failed to parse SSL port.")
			errorType = util.GetError(err)
			lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
			dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
			metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), time.Since(startTime).Seconds(), dimensionsMap)
			return nil, err
		}
		secretListenerString := service.Annotations[ServiceAnnotationLoadBalancerTLSSecret]
		secretBackendSetString := service.Annotations[ServiceAnnotationLoadBalancerTLSBackendSetSecret]
		sslConfig = NewSSLConfig(secretListenerString, secretBackendSetString, service, ports, cp)
	}
	subnets, err := cp.getLoadBalancerSubnets(ctx, logger, service)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to get Load balancer Subnets.")
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), time.Since(startTime).Seconds(), dimensionsMap)
		return nil, err
	}

	spec, err := NewLBSpec(logger, service, nodes, subnets, sslConfig, cp.securityListManagerFactory, cp.config.Tags, lb)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to derive LBSpec")
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), time.Since(startTime).Seconds(), dimensionsMap)

		return nil, err
	}

	if requiresNsgManagement(service) {
		// Fetch existing frontend NSG and use it to manage rules
		frontendNsgId := ""
		backendNsgs := spec.ManagedNetworkSecurityGroup.backendNsgId

		// Check if there are any NSGs which are created by CCM (and use that), but didn't get attached to LB because the LB creation failed.
		if !lbExists {
			frontendNsgId, _, err = cp.getFrontendNsgByName(ctx, logger, generateNsgName(service), cp.config.CompartmentID, cp.config.VCNID, fmt.Sprintf("%s", service.UID))
			if err != nil {
				return nil, err
			}
			logger.Infof("found managed NSG %s", frontendNsgId)
			if frontendNsgId != "" {
				spec, err = addFrontendNsgToSpec(spec, frontendNsgId)
				if err != nil {
					return nil, err
				}
			}
		}
		if lb != nil && lb.Id != nil && lb.NetworkSecurityGroupIds != nil {
			nsgs := lb.NetworkSecurityGroupIds
			for _, id := range nsgs {
				frontendNsgId, _, err = cp.getFrontendNsg(ctx, logger, id, fmt.Sprintf("%s", service.UID))
				if err != nil {
					errorType = util.GetError(err)
					nsgMetricDimension = util.GetMetricDimensionForComponent(errorType, util.NSGType)
					dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
					metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Get), time.Since(startTime).Seconds(), dimensionsMap)
				}
				if frontendNsgId != "" {
					spec, err = addFrontendNsgToSpec(spec, frontendNsgId)
					if err != nil {
						return nil, err
					}
					logger.With("loadBalancerID", *lb.Id).Infof("using existing frontendNsg %s", frontendNsgId)
					break
				}
			}

			if frontendNsgId == "" {
				// Check if there are any CCM created NSGs which might be manually removed by customer causing a dirty LB
				logger.Info("Check if managed NSGs present in VCN")
				frontendNsgId, _, err = cp.getFrontendNsgByName(ctx, logger, generateNsgName(service), cp.config.CompartmentID, cp.config.VCNID, fmt.Sprintf("%s", service.UID))
				if err != nil {
					return nil, err
				}
				logger.Infof("found managed NSG %s", frontendNsgId)
				if frontendNsgId != "" {
					spec, err = addFrontendNsgToSpec(spec, frontendNsgId)
					if err != nil {
						return nil, err
					}
				}
			}
		}

		// Create the NSG and add it to the LbSpec
		if frontendNsgId == "" {
			if len(spec.NetworkSecurityGroupIds) >= MaxNsgPerVnic {
				return nil, fmt.Errorf("invalid number of Network Security Groups (Max: 5) including managed nsg")
			}
			resp, err := cp.client.Networking().CreateNetworkSecurityGroup(ctx, cp.config.CompartmentID, cp.config.VCNID, generateNsgName(service), fmt.Sprintf("%s", service.UID))
			if err != nil {
				logger.With(zap.Error(err)).Error("Failed to create nsg")
				errorType = util.GetError(err)
				nsgMetricDimension = util.GetMetricDimensionForComponent(errorType, util.NSGType)
				dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
				metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Create), time.Since(startTime).Seconds(), dimensionsMap)
				return nil, err
			}
			frontendNsgId = *resp.Id
			spec, err = addFrontendNsgToSpec(spec, frontendNsgId)
			if err != nil {
				return nil, err
			}
			logger.With("frontendNsgId", *resp.Id).
				Info("Successfully created nsg")
			nsgMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.NSGType)
			dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
			dimensionsMap[metrics.ResourceOCIDDimension] = *resp.Id
			metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Create), time.Since(startTime).Seconds(), dimensionsMap)

		}
		if len(backendNsgs) > 0 {
			for _, nsg := range backendNsgs {
				resp, etag, err := cp.client.Networking().GetNetworkSecurityGroup(ctx, nsg)
				if err != nil {
					logger.With(zap.Error(err)).Error("Failed to get nsg")
					errorType = util.GetError(err)
					nsgMetricDimension = util.GetMetricDimensionForComponent(errorType, util.NSGType)
					dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
					metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Get), time.Since(startTime).Seconds(), dimensionsMap)
					return nil, err
				}
				freeformTags := resp.FreeformTags
				if _, ok := freeformTags["ManagedBy"]; !ok {
					if etag != nil {
						freeformTags["ManagedBy"] = "CCM"
						response, err := cp.client.Networking().UpdateNetworkSecurityGroup(ctx, nsg, *etag, freeformTags)
						if err != nil {
							logger.With(zap.Error(err)).Errorf("Failed to update nsg %s", nsg)
							errorType = util.GetError(err)
							nsgMetricDimension = util.GetMetricDimensionForComponent(errorType, util.NSGType)
							dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
							dimensionsMap[metrics.ResourceOCIDDimension] = nsg
							metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Update), time.Since(startTime).Seconds(), dimensionsMap)
							return nil, err
						}
						nsgMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.NSGType)
						dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
						dimensionsMap[metrics.ResourceOCIDDimension] = *response.Id
						metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Update), time.Since(startTime).Seconds(), dimensionsMap)
					}
				}
			}
		}
		serviceComponents := serviceComponents{
			frontendNsgOcid:  frontendNsgId,
			backendNsgOcids:  backendNsgs,
			ports:            spec.Ports,
			sourceCIDRs:      spec.SourceCIDRs,
			isPreserveSource: *spec.IsPreserveSource,
			serviceUid:       fmt.Sprintf("service-uid-%s", service.UID),
		}
		logger.Infof("(requiresNSGmanagement) Service Components %#v", serviceComponents)
		if err = cp.reconcileSecurityGroup(ctx, serviceComponents); err != nil {
			return nil, err
		}
	}

	if !lbExists {
		lbStatus, newLBOCID, err := lbProvider.createLoadBalancer(ctx, spec)
		if err != nil {
			logger.With(zap.Error(err)).Error("Failed to provision LoadBalancer")
			errorType = util.GetError(err)
			lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
			dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
			metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Create), time.Since(startTime).Seconds(), dimensionsMap)

		} else {
			logger.With("loadBalancerID", newLBOCID).
				Info("Successfully provisioned loadbalancer")
			lbMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.LoadBalancerType)
			dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
			dimensionsMap[metrics.ResourceOCIDDimension] = newLBOCID
			metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Create), time.Since(startTime).Seconds(), dimensionsMap)
		}
		return lbStatus, err
	}

	if lb.LifecycleState == nil || *lb.LifecycleState != lbLifecycleStateActive {
		logger := logger.With("lifecycleState", lb.LifecycleState)
		switch loadBalancerType {
		case NLB:
			// This check is added here since NLBs are marked as failed in case nlb work-requests fail NLB-26239
			if *lb.LifecycleState == string(networkloadbalancer.LifecycleStateFailed) {
				logger.Infof("NLB is in %s state, process the Loadbalancer", *lb.LifecycleState)
			} else {
				return nil, errors.Errorf("NLB is in %s state, wait for NLB to move to %s", *lb.LifecycleState, lbLifecycleStateActive)
			}
			break
		default:
			logger.Infof("LB is not in %s state, will retry EnsureLoadBalancer", lbLifecycleStateActive)
			return nil, errors.Errorf("rejecting request to update LB which is not in %s state", lbLifecycleStateActive)
		}
	}

	// Existing load balancers cannot change subnets. This ensures that the spec matches
	// what the actual load balancer has listed as the subnet ids. If the load balancer
	// was just created then these values would be equal; however, if the load balancer
	// already existed and the default subnet ids changed, then this would ensure
	// we are setting the security rules on the correct subnets.
	spec, err = updateSpecWithLbSubnets(spec, lb.SubnetIds)
	if err != nil {
		return nil, err
	}

	// If the load balancer needs an SSL cert ensure it is present.
	if requiresCertificate(service) {
		if err := lbProvider.ensureSSLCertificates(ctx, lb, spec); err != nil {
			logger.With(zap.Error(err)).Error("Failed to ensure ssl certificates")
			errorType = util.GetError(err)
			lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
			dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
			metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), time.Since(startTime).Seconds(), dimensionsMap)

			return nil, errors.Wrap(err, "ensuring ssl certificates")
		}
	}

	// If network partition, do not proceed
	isNetworkPartition, err := cp.checkForNetworkPartition(logger, clusterNodes)
	if err != nil {
		return nil, err
	} else if isNetworkPartition {
		return nil, nil
	}

	if err := lbProvider.updateLoadBalancer(ctx, lb, spec); err != nil {
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		logger.With(zap.Error(err)).Error("Failed to update LoadBalancer")
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), time.Since(startTime).Seconds(), dimensionsMap)
		return nil, err
	}

	syncTime := time.Since(startTime).Seconds()
	logger.Info("Successfully updated loadbalancer")
	lbMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.LoadBalancerType)
	dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
	dimensionsMap[metrics.BackendSetsCountDimension] = strconv.Itoa(len(lb.BackendSets))
	metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), syncTime, dimensionsMap)
	return loadBalancerToStatus(lb)
}

func getDefaultLBSubnets(subnet1, subnet2 string) []string {
	var subnets []string
	if subnet2 != "" {
		subnets = []string{subnet1, subnet2}
	} else {
		subnets = []string{subnet1}
	}
	return subnets
}

func (cp *CloudProvider) getNetworkLoadbalancerSubnets(ctx context.Context, logger *zap.SugaredLogger, svc *v1.Service) ([]string, error) {
	subnets := getDefaultLBSubnets(cp.config.LoadBalancer.Subnet1, cp.config.LoadBalancer.Subnet2)
	if s, ok := svc.Annotations[ServiceAnnotationNetworkLoadBalancerSubnet]; ok && len(s) != 0 {
		return []string{s}, nil
	}
	if len(subnets) == 0 {
		return nil, errors.Errorf("a subnet must be specified for a network load balancer to get created")
	}
	if len(subnets) > 0 && subnets[0] != "" {
		return []string{subnets[0]}, nil
	}
	return nil, errors.Errorf("a subnet must be specified for a network load balancer")
}

func (cp *CloudProvider) getOciLoadBalancerSubnets(ctx context.Context, logger *zap.SugaredLogger, svc *v1.Service) ([]string, error) {
	internal, err := isInternalLB(svc)
	if err != nil {
		return nil, err
	}

	// NOTE: These will be overridden for existing load balancers as load
	// balancer subnets cannot be modified.
	subnets := getDefaultLBSubnets(cp.config.LoadBalancer.Subnet1, cp.config.LoadBalancer.Subnet2)

	if s, ok := svc.Annotations[ServiceAnnotationLoadBalancerSubnet1]; ok && len(s) != 0 {
		subnets[0] = s
		r, err := cp.client.Networking().IsRegionalSubnet(ctx, s)
		if err != nil {
			return nil, err
		}
		if r {
			return subnets[:1], nil
		}
	}

	if s, ok := svc.Annotations[ServiceAnnotationLoadBalancerSubnet2]; ok && len(s) != 0 {
		r, err := cp.client.Networking().IsRegionalSubnet(ctx, s)
		if err != nil {
			return nil, err
		}
		if r {
			subnets[0] = s
			logger.Debugf("Considering annotation %s: %s for LB as it is the only regional subnet in annotations provided.", ServiceAnnotationLoadBalancerSubnet2, s)
			return subnets[:1], nil
		} else if len(subnets) > 1 {
			subnets[1] = s
		} else {
			subnets = append(subnets, s)
		}
	}

	if subnets[0] == "" || (len(subnets) == 2 && subnets[1] == "") {
		return nil, errors.Errorf("a subnet must be specified for creating a load balancer")
	}
	if internal {
		// Public load balancers need two subnets if they are AD specific and only first subnet is used if regional. Internal load
		// balancers will always use the first subnet.
		return subnets[:1], nil
	}

	return subnets, nil
}

func (cp *CloudProvider) getLoadBalancerSubnets(ctx context.Context, logger *zap.SugaredLogger, svc *v1.Service) ([]string, error) {
	lbType := getLoadBalancerType(svc)

	switch lbType {
	case NLB:
		return cp.getNetworkLoadbalancerSubnets(ctx, logger, svc)
	default:
		return cp.getOciLoadBalancerSubnets(ctx, logger, svc)
	}
}

func (clb *CloudLoadBalancerProvider) updateLoadBalancer(ctx context.Context, lb *client.GenericLoadBalancer, spec *LBSpec) error {
	lbID := *lb.Id

	logger := clb.logger.With("loadBalancerID", lbID, "compartmentID", clb.config.CompartmentID, "loadBalancerType", getLoadBalancerType(spec.service), "serviceName", spec.service.Name)

	var actualPublicReservedIP *string

	//identify the public reserved IP in IP addresses list
	for _, ip := range lb.IpAddresses {
		if ip.IpAddress == nil {
			continue // should never happen but appears to when EnsureLoadBalancer is called with 0 nodes.
		}
		if ip.ReservedIp != nil && *ip.IsPublic {
			actualPublicReservedIP = ip.IpAddress
			break
		}
	}

	actualBackendSets := lb.BackendSets
	desiredBackendSets := spec.BackendSets
	backendSetActions := getBackendSetChanges(logger, actualBackendSets, desiredBackendSets)

	actualListeners := lb.Listeners
	desiredListeners := spec.Listeners
	listenerActions := getListenerChanges(logger, actualListeners, desiredListeners)

	lbSubnets, err := getSubnets(ctx, spec.Subnets, clb.client.Networking())
	if err != nil {
		return errors.Wrapf(err, "getting load balancer subnets")
	}
	nodeSubnets, err := getSubnetsForNodes(ctx, spec.nodes, clb.client)
	if err != nil {
		return errors.Wrap(err, "get subnets for nodes")
	}

	if len(backendSetActions) == 0 && len(listenerActions) == 0 {
		// If there are no backendSetActions or Listener actions
		// this function must have been called because of a failed
		// seclist update when the load balancer was created
		// We try to update the seclist this way to prevent replication
		// of seclist reconciliation logic
		// Security List Updates happen in a Global Critical Section
		if err = updateSecurityListsInCriticalSection(ctx, spec, lbSubnets, nodeSubnets); err != nil {
			return err
		}
	}
	actions := sortAndCombineActions(logger, backendSetActions, listenerActions)
	for _, action := range actions {
		switch a := action.(type) {
		case *BackendSetAction:
			err := clb.updateBackendSet(ctx, lbID, a, lbSubnets, nodeSubnets, spec.securityListManager, spec)
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

			err := clb.updateListener(ctx, lbID, a, ports, lbSubnets, nodeSubnets, spec.SourceCIDRs, spec.securityListManager, spec)
			if err != nil {
				return errors.Wrap(err, "updating listener")
			}
		}
	}

	// Check if the customer managed LB NSGs have changed
	nsgChanged := hasLoadBalancerNetworkSecurityGroupsChanged(ctx, lb.NetworkSecurityGroupIds, spec.NetworkSecurityGroupIds)
	if nsgChanged {
		err = clb.updateLoadBalancerNetworkSecurityGroups(ctx, lb, spec)
		if err != nil {
			return err
		}
	}

	// Only LB supports fixed/flexible shapes which can be changed
	if spec.Type == LB {
		shapeChanged := hasLoadbalancerShapeChanged(ctx, spec, lb)

		if shapeChanged {
			err = clb.updateLoadbalancerShape(ctx, lb, spec)
			if err != nil {
				return err
			}
		}
	}

	// Check if the reservedIP has changed in spec
	if spec.LoadBalancerIP != "" || actualPublicReservedIP != nil {
		if actualPublicReservedIP == nil || *actualPublicReservedIP != spec.LoadBalancerIP {
			return errors.Errorf("The Load Balancer service reserved IP cannot be updated after the Load Balancer is created.")
		}
	}
	return nil
}

func (clb *CloudLoadBalancerProvider) updateLoadBalancerBackends(ctx context.Context, lb *client.GenericLoadBalancer, spec *LBSpec) error {
	lbID := *lb.Id

	logger := clb.logger.With("loadBalancerID", lbID, "compartmentID", clb.config.CompartmentID, "loadBalancerType", getLoadBalancerType(spec.service), "serviceName", spec.service.Name)

	actualBackendSets := lb.BackendSets
	desiredBackendSets := spec.BackendSets
	backendSetActions := getBackendSetChanges(logger, actualBackendSets, desiredBackendSets)

	lbSubnets, err := getSubnets(ctx, spec.Subnets, clb.client.Networking())
	if err != nil {
		return errors.Wrapf(err, "getting load balancer subnets")
	}
	nodeSubnets, err := getSubnetsForNodes(ctx, spec.nodes, clb.client)
	if err != nil {
		return errors.Wrap(err, "get subnets for nodes")
	}

	for _, action := range backendSetActions {
		switch a := action.(type) {
		case *BackendSetAction:
			switch a.Type() {
			case Update:
				err := clb.updateBackendSet(ctx, lbID, a, lbSubnets, nodeSubnets, spec.securityListManager, spec)
				if err != nil {
					return errors.Wrap(err, "updating BackendSet")
				}
			}
		}
	}
	return nil
}

func updateSecurityListsInCriticalSection(ctx context.Context, spec *LBSpec, lbSubnets, nodeSubnets []*core.Subnet) (err error) {
	updateRulesMutex.Lock()
	defer updateRulesMutex.Unlock()
	for _, ports := range spec.Ports {
		if err = spec.securityListManager.Update(ctx, lbSubnets, nodeSubnets, spec.SourceCIDRs, nil, ports, *spec.IsPreserveSource); err != nil {
			return err
		}
	}
	return nil
}

func (clb *CloudLoadBalancerProvider) updateBackendSet(ctx context.Context, lbID string, action *BackendSetAction, lbSubnets, nodeSubnets []*core.Subnet, secListManager securityListManager, spec *LBSpec) error {
	var (
		sourceCIDRs   = []string{}
		workRequestID string
		err           error
		bs            = action.BackendSet
		ports         = action.Ports
	)

	logger := clb.logger.With(
		"actionType", action.Type(),
		"backendSetName", action.Name(),
		"ports", ports,
		"loadBalancerID", lbID,
		"loadBalancerType", getLoadBalancerType(spec.service))
	logger.Info("Applying action on backend set")
	switch action.Type() {
	case Create:
		err = secListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, nil, ports, *spec.IsPreserveSource)
		if err != nil {
			return err
		}
		workRequestID, err = clb.lbClient.CreateBackendSet(ctx, lbID, action.Name(), &bs)
	case Update:
		// For NLB, due to source IP preservation we need to ensure ingress rules from sourceCIDRs are added to
		// the backends subnet's seclist as well

		if err = secListManager.Update(ctx, lbSubnets, nodeSubnets, spec.SourceCIDRs, action.OldPorts, ports, *spec.IsPreserveSource); err != nil {
			return err
		}
		workRequestID, err = clb.lbClient.UpdateBackendSet(ctx, lbID, action.Name(), &bs)
	case Delete:
		err = secListManager.Delete(ctx, lbSubnets, nodeSubnets, ports, sourceCIDRs, *spec.IsPreserveSource)
		if err != nil {
			return err
		}
		workRequestID, err = clb.lbClient.DeleteBackendSet(ctx, lbID, action.Name())
	}

	if err != nil {
		return err
	}
	logger = logger.With("workRequestID", workRequestID)
	logger.Info("Await workrequest for loadbalancer backendset")
	_, err = clb.lbClient.AwaitWorkRequest(ctx, workRequestID)
	if err != nil {
		return err
	}
	logger.Info("Workrequest for loadbalancer backendset completed successfully")

	return nil
}

func (clb *CloudLoadBalancerProvider) updateListener(ctx context.Context, lbID string, action *ListenerAction, ports portSpec, lbSubnets, nodeSubnets []*core.Subnet, sourceCIDRs []string, secListManager securityListManager, spec *LBSpec) error {
	var workRequestID string
	var err error
	listener := action.Listener
	ports.ListenerPort = *listener.Port

	logger := clb.logger.With(
		"actionType", action.Type(),
		"listenerName", action.Name(),
		"ports", ports,
		"loadBalancerID", lbID,
		"loadBalancerType", getLoadBalancerType(spec.service))
	logger.Info("Applying action on listener")
	switch action.Type() {
	case Create:
		err = secListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, nil, ports, *spec.IsPreserveSource)
		if err != nil {
			return err
		}
		workRequestID, err = clb.lbClient.CreateListener(ctx, lbID, action.Name(), &listener)
	case Update:
		err = secListManager.Update(ctx, lbSubnets, nodeSubnets, sourceCIDRs, nil, ports, *spec.IsPreserveSource)
		if err != nil {
			return err
		}
		workRequestID, err = clb.lbClient.UpdateListener(ctx, lbID, action.Name(), &listener)
	case Delete:
		err = secListManager.Delete(ctx, lbSubnets, nodeSubnets, ports, sourceCIDRs, *spec.IsPreserveSource)
		if err != nil {
			return err
		}
		workRequestID, err = clb.lbClient.DeleteListener(ctx, lbID, action.Name())
	}

	if err != nil {
		return err
	}
	logger = logger.With("workRequestID", workRequestID)
	logger.Info("Await workrequest for loadbalancer listener")
	_, err = clb.lbClient.AwaitWorkRequest(ctx, workRequestID)
	if err != nil {
		return err
	}
	logger.Info("Workrequest for loadbalancer listener completed successfully")
	return nil
}

// UpdateLoadBalancer updates an existing loadbalancer
func (cp *CloudProvider) UpdateLoadBalancer(ctx context.Context, clusterName string, service *v1.Service, nodes []*v1.Node) error {
	startTime := time.Now()
	lbName := GetLoadBalancerName(service)
	loadBalancerType := getLoadBalancerType(service)
	logger := cp.logger.With("loadBalancerName", lbName, "serviceName", service.Name, "loadBalancerType", loadBalancerType, "serviceUid", service.UID)
	if sa, useWI := service.Annotations[ServiceAnnotationServiceAccountName]; useWI { // When using Workload Identity
		logger = logger.With("serviceAccount", sa, "namespace", service.Namespace)
	}

	if deleted, err := cp.serviceDeletedOrDoesNotExist(ctx, service); deleted {
		if err != nil {
			logger.With(zap.Error(err)).Error("Failed to check if service exists")
			return errors.Wrap(err, "Failed to check service status")
		}
		logger.Info("Service already deleted or no more exists")
		return errors.New("Service already deleted or no more exists")
	}
	loadBalancerService := fmt.Sprintf("%s/%s", service.Namespace, service.Name)
	if acquired := cp.lbLocks.TryAcquire(loadBalancerService); !acquired {
		logger.Error("Could not acquire lock for Updating Load Balancer")
		return LbOperationAlreadyExists
	}
	defer cp.lbLocks.Release(loadBalancerService)

	nodes, err := filterNodes(service, nodes)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to filter nodes with label selector")
		return err
	}

	logger.With("nodes", len(nodes)).Info("Ensuring load balancer")

	// If network partition, do not proceed
	isNetworkPartition, err := cp.checkForNetworkPartition(logger, nodes)
	if err != nil {
		return err
	} else if isNetworkPartition {
		return nil
	}

	dimensionsMap := make(map[string]string)

	var errorType string
	var lbMetricDimension string

	lbProvider, err := cp.getLoadBalancerProvider(ctx, service)
	if err != nil {
		return errors.Wrap(err, "Unable to get Load Balancer Client.")
	}
	lb, err := lbProvider.lbClient.GetLoadBalancerByName(ctx, cp.config.CompartmentID, lbName)
	if err != nil && !client.IsNotFound(err) {
		logger.With(zap.Error(err)).Error("Failed to get loadbalancer by name")
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		dimensionsMap[metrics.ResourceOCIDDimension] = lbName
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), time.Since(startTime).Seconds(), dimensionsMap)
		return err
	} else if client.IsNotFound(err) {
		logger.Infof("Could not find load balancer, will not retry UpdateLoadBalancer.")
		return nil
	}

	if lb.LifecycleState == nil || *lb.LifecycleState != lbLifecycleStateActive {
		logger := logger.With("lifecycleState", lb.LifecycleState)
		switch loadBalancerType {
		case NLB:
			// This check is added here since NLBs are marked as failed in case nlb work-requests fail NLB-26239
			if *lb.LifecycleState == string(networkloadbalancer.LifecycleStateFailed) {
				logger.Infof("NLB is in %s state, process the Loadbalancer", *lb.LifecycleState)
			} else {
				return errors.Errorf("NLB is in %s state, wait for NLB to move to %s", *lb.LifecycleState, lbLifecycleStateActive)
			}
			break
		default:
			logger.Infof("LB is not in %s state, will retry UpdateLoadBalancer", lbLifecycleStateActive)
			return errors.Errorf("rejecting request to update LB which is not in %s state", lbLifecycleStateActive)
		}
	}

	lbOCID := ""
	if lb != nil && lb.Id != nil {
		lbOCID = *lb.Id
	} else {
		// if the LB does not exist already use the k8s service UID for reference
		// in logs and metrics
		logger.Error("Load Balancer Id is empty, will retry UpdateLoadBalancer.")
		return errors.New("Load Balancer service returned empty Id, will wait and retry")
	}

	logger = logger.With("loadBalancerID", lbOCID)
	dimensionsMap[metrics.ResourceOCIDDimension] = lbOCID

	err = cp.checkPendingLBWorkRequests(ctx, logger, lbProvider, lb, service, startTime)
	if err != nil {
		return err
	}

	subnets, err := cp.getLoadBalancerSubnets(ctx, logger, service)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to get Load balancer Subnets.")
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), time.Since(startTime).Seconds(), dimensionsMap)
		return err
	}

	spec, err := NewLBSpec(logger, service, nodes, subnets, nil, cp.securityListManagerFactory, cp.config.Tags, lb)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to derive LBSpec")
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), time.Since(startTime).Seconds(), dimensionsMap)

		return err
	}

	// Existing load balancers cannot change subnets. This ensures that the spec matches
	// what the actual load balancer has listed as the subnet ids. If the load balancer
	// was just created then these values would be equal; however, if the load balancer
	// already existed and the default subnet ids changed, then this would ensure
	// we are setting the security rules on the correct subnets.
	spec, err = updateSpecWithLbSubnets(spec, lb.SubnetIds)
	if err != nil {
		return err
	}

	if err := lbProvider.updateLoadBalancerBackends(ctx, lb, spec); err != nil {
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		logger.With(zap.Error(err)).Error("Failed to update LoadBalancer backends")
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), time.Since(startTime).Seconds(), dimensionsMap)
		return err
	}

	syncTime := time.Since(startTime).Seconds()
	logger.Info("Successfully updated loadbalancer backends")
	lbMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.LoadBalancerType)
	dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
	dimensionsMap[metrics.BackendSetsCountDimension] = strconv.Itoa(len(lb.BackendSets))
	metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Update), syncTime, dimensionsMap)
	return nil
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
			return nil, errors.Errorf("node was not found by IP %q", ip)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// EnsureLoadBalancerDeleted deletes the specified load balancer if it exists,
// returning nil if the load balancer specified either didn't exist or was
// successfully deleted.
func (cp *CloudProvider) EnsureLoadBalancerDeleted(ctx context.Context, clusterName string, service *v1.Service) error {
	startTime := time.Now()
	name := cp.GetLoadBalancerName(ctx, clusterName, service)
	loadBalancerType := getLoadBalancerType(service)
	logger := cp.logger.With("loadBalancerName", name, "loadBalancerType", loadBalancerType)
	if sa, useWI := service.Annotations[ServiceAnnotationServiceAccountName]; useWI { // When using Workload Identity
		logger = logger.With("serviceAccount", sa, "nameSpace", service.Namespace)
	}
	logger.Debug("Attempting to delete load balancer")
	loadBalancerService := fmt.Sprintf("%s/%s", service.Namespace, service.Name)
	if acquired := cp.lbLocks.TryAcquire(loadBalancerService); !acquired {
		logger.Error("Could not acquire lock for Deleting Load Balancer")
		return LbOperationAlreadyExists
	}
	defer cp.lbLocks.Release(loadBalancerService)

	var errorType string
	var lbMetricDimension string
	var nsgMetricDimension string

	dimensionsMap := make(map[string]string)
	var frontendNsgId = ""
	uid := fmt.Sprintf("%s", service.UID)
	var etag *string

	securityRuleManagementMode, nsg, err := getRuleManagementMode(service)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get rule management mode")
		return errors.Wrap(err, "failed to get rule management mode")
	}

	lbProvider, err := cp.getLoadBalancerProvider(ctx, service)
	if err != nil {
		return errors.Wrap(err, "Unable to get Load Balancer Client.")
	}
	lb, err := lbProvider.lbClient.GetLoadBalancerByName(ctx, cp.config.CompartmentID, name)
	if err != nil {
		if client.IsNotFound(err) {
			logger.Info("Could not find load balancer. Nothing to do.")
			if securityRuleManagementMode == NSG {
				displayName := generateNsgName(service)
				nsg.frontendNsgId, etag, err = cp.getFrontendNsgByName(ctx, logger, displayName, cp.config.CompartmentID, cp.config.VCNID, uid)
				if err != nil {
					return errors.Wrap(err, "failed to get frontend NSG")
				}
				// Delete of NSG happens if NSG was created but LB creation fails
				if nsg != nil && nsg.nsgRuleManagementMode == RuleManagementModeNsg && nsg.frontendNsgId != "" {
					if etag != nil {
						logger = logger.With("frontendNsgId", nsg.frontendNsgId)
						logger.Infof("deleting frontend nsg %s", nsg.frontendNsgId)
						nsgDeleted, err := cp.deleteNsg(ctx, logger, nsg.frontendNsgId, *etag)
						if !nsgDeleted || err != nil {
							logger.With(zap.Error(err)).Error("failed to delete nsg")
							errorType = util.GetError(err)
							nsgMetricDimension = util.GetMetricDimensionForComponent(errorType, util.NSGType)
							dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
							metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Delete), time.Since(startTime).Seconds(), dimensionsMap)
							return err
						}
						nsgMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.NSGType)
						dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
						metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Delete), time.Since(startTime).Seconds(), dimensionsMap)
						logger.Infof("Managed nsg with id %s deleted", nsg.frontendNsgId)
					}
				}
			}
			return nil
		}
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		logger.With(zap.Error(err)).Error("Failed to get loadbalancer by name")
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		dimensionsMap[metrics.ResourceOCIDDimension] = name
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Delete), time.Since(startTime).Seconds(), dimensionsMap)
		return errors.Wrapf(err, "get load balancer %q by name", name)
	}

	id := *lb.Id
	dimensionsMap[metrics.ResourceOCIDDimension] = id
	logger = logger.With("loadBalancerID", id, "loadBalancerType", getLoadBalancerType(service))

	if securityRuleManagementMode == NSG {
		// List network security groups
		nsgs := lb.NetworkSecurityGroupIds
		for _, nsgId := range nsgs {
			frontendNsgId, etag, err = cp.getFrontendNsg(ctx, logger, nsgId, uid)
			if err != nil {
				errorType = util.GetError(err)
				nsgMetricDimension = util.GetMetricDimensionForComponent(errorType, util.NSGType)
				dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
				metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Get), time.Since(startTime).Seconds(), dimensionsMap)
			}
			if frontendNsgId != "" {
				logger = logger.With("frontendNsgId", frontendNsgId)
				nsg.frontendNsgId = frontendNsgId
				break
			}
		}
	}

	// get annotation from load balancer spec and compare to ManagementModeNone
	if securityRuleManagementMode != ManagementModeNone {
		err := cp.cleanupSecurityRulesForLoadBalancerDelete(lb, logger, ctx, service, name, frontendNsgId)
		if err != nil {
			errorType = util.GetError(err)
			lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
			dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
			metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Delete), time.Since(startTime).Seconds(), dimensionsMap)

			return err
		}
	}

	logger.Info("Deleting load balancer")
	workReqID, err := lbProvider.lbClient.DeleteLoadBalancer(ctx, id)
	if err != nil {
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		logger.With(zap.Error(err)).Error("Failed to delete loadbalancer")
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Delete), time.Since(startTime).Seconds(), dimensionsMap)

		return errors.Wrapf(err, "delete load balancer %q", id)
	}
	logger.With("workRequestID", workReqID).Info("Await workrequest for delete loadbalancer")
	_, err = lbProvider.lbClient.AwaitWorkRequest(ctx, workReqID)
	if err != nil {
		logger.With(zap.Error(err)).Error("Timeout waiting for loadbalancer delete")
		errorType = util.GetError(err)
		lbMetricDimension = util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Delete), time.Since(startTime).Seconds(), dimensionsMap)
		return errors.Wrapf(err, "awaiting deletion of load balancer %q", name)
	}
	logger.With("workRequestID", workReqID).Info("Workrequest for delete loadbalancer succeeded")
	logger.Info("Loadbalancer deleted")
	lbMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.LoadBalancerType)
	dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
	metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, Delete), time.Since(startTime).Seconds(), dimensionsMap)

	// Delete of NSG happens after delete of the Loadbalancer
	if nsg != nil && nsg.nsgRuleManagementMode == RuleManagementModeNsg && nsg.frontendNsgId != "" {
		if etag != nil {
			logger = logger.With("frontendNsgId", nsg.frontendNsgId)
			logger.Infof("deleting frontend nsg %s", nsg.frontendNsgId)
			nsgDeleted, err := cp.deleteNsg(ctx, logger, nsg.frontendNsgId, *etag)
			if !nsgDeleted || err != nil {
				logger.With(zap.Error(err)).Error("failed to delete nsg")
				errorType = util.GetError(err)
				nsgMetricDimension = util.GetMetricDimensionForComponent(errorType, util.NSGType)
				dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
				metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Delete), time.Since(startTime).Seconds(), dimensionsMap)
				return err
			}
			nsgMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.NSGType)
			dimensionsMap[metrics.ComponentDimension] = nsgMetricDimension
			metrics.SendMetricData(cp.metricPusher, getMetric(util.NSGType, Delete), time.Since(startTime).Seconds(), dimensionsMap)
			logger.Infof("managed nsg with id %s deleted", nsg.frontendNsgId)
		}
	}

	return nil
}

// Critical Section for Security List Updates
func (cp *CloudProvider) cleanupSecurityRulesForLoadBalancerDelete(lb *client.GenericLoadBalancer, logger *zap.SugaredLogger, ctx context.Context, service *v1.Service, name string, frontendNsgOcid string) error {
	updateRulesMutex.Lock()
	defer updateRulesMutex.Unlock()

	id := *lb.Id
	ipSet := sets.NewString()
	for _, backendSet := range lb.BackendSets {
		for _, backend := range backendSet.Backends {
			ipSet.Insert(*backend.IpAddress)
		}
	}
	nodes, err := cp.getNodesByIPs(ipSet.List())
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to fetch nodes by internal ips")
		return errors.Wrap(err, "fetching nodes by internal ips")
	}
	nodeSubnets, err := getSubnetsForNodes(ctx, nodes, cp.client)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to get subnets for nodes")
		return errors.Wrap(err, "getting subnets for nodes")
	}

	lbSubnets, err := getSubnets(ctx, lb.SubnetIds, cp.client.Networking())
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to get subnets for load balancers")
		return errors.Wrap(err, "getting subnets for load balancers")
	}

	securityRuleManagerMode, managedNsg, err := getRuleManagementMode(service)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get Security Rule management mode")
		return errors.Wrap(err, "failed to get Security Rule management mode")
	}

	backendNsgIds, err := getManagedBackendNSG(service)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get backend Nsgs from spec")
		return errors.Wrap(err, "failed to get backend Nsgs from spec")
	}
	var securityListManager securityListManager
	if securityRuleManagerMode == NSG {
		if managedNsg != nil {
			if frontendNsgOcid != "" {
				managedNsg.frontendNsgId = frontendNsgOcid
			}
			if len(backendNsgIds) > 0 {
				managedNsg.backendNsgId = backendNsgIds
			}
		}
	} else {
		securityListManager = cp.securityListManagerFactory(
			securityRuleManagerMode)
	}

	isPreserveSource, err := getPreserveSource(logger, service)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to determine value for is-preserve-source")
		return errors.Wrap(err, "failed to determine value for is-preserve-source")
	}
	portsNsg, err := getPorts(service)
	if err != nil {
		return errors.Wrapf(err, "failed to get ports from spec")
	}
	sourceCIDRs, err := getLoadBalancerSourceRanges(service)
	if securityRuleManagerMode == NSG && len(managedNsg.backendNsgId) > 0 {
		serviceComponents := serviceComponents{
			frontendNsgOcid:  managedNsg.frontendNsgId,
			backendNsgOcids:  managedNsg.backendNsgId,
			sourceCIDRs:      sourceCIDRs,
			ports:            portsNsg,
			isPreserveSource: isPreserveSource,
			serviceUid:       fmt.Sprintf("service-uid-%s", service.UID),
		}
		logger.Infof("(ensureloadbalancer deleted) Service Components %#v", serviceComponents)
		err = cp.removeBackendSecurityGroupRules(ctx, serviceComponents)
		if err != nil {
			return err
		}
	}

	for listenerName, listener := range lb.Listeners {
		backendSetName := *listener.DefaultBackendSetName
		bs, ok := lb.BackendSets[backendSetName]
		if !ok {
			logger.With(zap.Error(err)).Errorf("Failed to delete loadbalencer as backend set %q missing (loadbalancer=%q)", backendSetName, id)
			return errors.Errorf("backend set %q missing (loadbalancer=%q)", backendSetName, id) // Should never happen.
		}

		ports := portsFromBackendSet(cp.logger, backendSetName, &bs)
		ports.ListenerPort = *listener.Port

		logger.With("listenerName", listenerName, "ports", ports).Debug("Deleting security rules for listener")

		sourceCIDRs, err := getLoadBalancerSourceRanges(service)
		if err != nil {
			logger.With(zap.Error(err)).Errorf("Failed to get security rules for listener %q on load balancer %q", listenerName, name)
			return errors.Wrapf(err, "delete security rules for listener %q on load balancer %q", listenerName, name)
		}

		logger.Infof("Security rule management mode %s", securityRuleManagerMode)
		if securityRuleManagerMode == ManagementModeAll || securityRuleManagerMode == ManagementModeFrontend {
			if err = securityListManager.Delete(ctx, lbSubnets, nodeSubnets, ports, sourceCIDRs, isPreserveSource); err != nil {
				logger.With(zap.Error(err)).Errorf("Failed to delete security rules for listener %q on load balancer %q", listenerName, name)
				return errors.Wrapf(err, "delete security rules for listener %q on load balancer %q", listenerName, name)
			}
		}

	}
	return nil
}

// only supported by LBaaS
func (clb *CloudLoadBalancerProvider) updateLoadbalancerShape(ctx context.Context, lb *client.GenericLoadBalancer, spec *LBSpec) error {
	shapeDetails := client.GenericUpdateLoadBalancerShapeDetails{
		ShapeName:    &spec.Shape,
		ShapeDetails: nil,
	}
	if *lb.ShapeName == flexible && spec.Shape != flexible {
		// LBaaS does not support converting from flexible to fixed shapes
		// as that can easily be achieved by setting the min and max bandwith to
		// whatever fixed shape that is needed
		return errors.New("cannot convert LB shape from flexible to fixed shape " + spec.Shape)
	}
	if spec.Shape == flexible {
		shapeDetails.ShapeDetails = &client.GenericShapeDetails{
			MinimumBandwidthInMbps: spec.FlexMin,
			MaximumBandwidthInMbps: spec.FlexMax,
		}
	}
	wrID, err := clb.lbClient.UpdateLoadBalancerShape(ctx, *lb.Id, &shapeDetails)
	if err != nil {
		return errors.Wrap(err, "failed to create UpdateLoadBalancerShape request")
	}
	logger := clb.logger.With("old-shape", *lb.ShapeName, "new-shape", spec.Shape,
		"flexMinimumMbps", spec.FlexMin, "flexMaximumMbps", spec.FlexMax,
		"opc-workrequest-id", wrID, "loadBalancerType", getLoadBalancerType(spec.service))
	logger.Info("Awaiting UpdateLoadBalancerShape workrequest")
	_, err = clb.lbClient.AwaitWorkRequest(ctx, wrID)
	if err != nil {
		return err
	}
	logger.Info("UpdateLoadBalancerShape request completed successfully")
	return nil
}

func (clb *CloudLoadBalancerProvider) updateLoadBalancerNetworkSecurityGroups(ctx context.Context, lb *client.GenericLoadBalancer, spec *LBSpec) error {
	wrID, err := clb.lbClient.UpdateNetworkSecurityGroups(ctx, *lb.Id, spec.NetworkSecurityGroupIds)
	if err != nil {
		return errors.Wrap(err, "failed to create UpdateNetworkSecurityGroups request")
	}
	logger := clb.logger.With("existingNSGIds", lb.NetworkSecurityGroupIds, "newNSGIds", spec.NetworkSecurityGroupIds,
		"opc-workrequest-id", wrID)
	logger.Info("Awaiting UpdateNetworkSecurityGroups workrequest")
	_, err = clb.lbClient.AwaitWorkRequest(ctx, wrID)
	if err != nil {
		return errors.Wrap(err, "failed to await UpdateNetworkSecurityGroups workrequest")
	}
	logger.Info("Loadbalancer UpdateNetworkSecurityGroups workrequest completed successfully")
	return nil
}

// Given an OCI load balancer, return a LoadBalancerStatus
func loadBalancerToStatus(lb *client.GenericLoadBalancer) (*v1.LoadBalancerStatus, error) {
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

func (cp *CloudProvider) checkAllBackendNodesNotReady(nodeList []*v1.Node) bool {
	for _, node := range nodeList {
		if _, hasExcludeBalancerLabel := node.Labels[excludeBackendFromLBLabel]; hasExcludeBalancerLabel {
			continue
		}
		for _, cond := range node.Status.Conditions {
			if cond.Type == v1.NodeReady {
				if cond.Status == v1.ConditionTrue {
					return false
				}
				break
			}
		}
	}
	return true
}

// If CCM manages the NSG for the service, CCM to delete the NSG when the LB/NLB service is deleted
func (cp *CloudProvider) deleteNsg(ctx context.Context, logger *zap.SugaredLogger, id, etag string) (bool, error) {
	opcRequestId, err := cp.client.Networking().DeleteNetworkSecurityGroup(ctx, id, etag)
	if err != nil {
		logger.Errorf("failed to delete nsg %s OpcRequestId %s", id, pointer.StringDeref(opcRequestId, ""))
		return false, err
	}
	logger.Infof("delete nsg OpcRequestId %s", pointer.StringDeref(opcRequestId, ""))
	return true, nil
}

func (cp *CloudProvider) getFrontendNsg(ctx context.Context, logger *zap.SugaredLogger, id, uid string) (frontendNsgId string, etag *string, err error) {
	nsg, etag, err := cp.client.Networking().GetNetworkSecurityGroup(ctx, id)
	if err != nil || nsg == nil || etag == nil {
		logger.Errorf("failed to get nsg %s", id)
		return "", nil, err
	}
	freeFormTags := map[string]string{"CreatedBy": "CCM", "ServiceUid": uid}
	if reflect.DeepEqual(nsg.FreeformTags, freeFormTags) {
		nsgId := pointer.StringDeref(nsg.Id, "")
		logger.Infof("Found managed frontend nsg %s", nsgId)
		return nsgId, etag, nil
	} else {
		logger.Infof("Found attached nsgs %s but not managed", pointer.StringDeref(nsg.Id, ""))
		return "", nil, nil
	}
}

func (cp *CloudProvider) getFrontendNsgByName(ctx context.Context, logger *zap.SugaredLogger, displayName, compartmentId, vcnId, uid string) (frontendNsgId string, etag *string, err error) {
	nsgs, err := cp.client.Networking().ListNetworkSecurityGroups(ctx, displayName, compartmentId, vcnId)
	for _, nsg := range nsgs {
		frontendNsgId, etag, err = cp.getFrontendNsg(ctx, logger, pointer.StringDeref(nsg.Id, ""), uid)
		if err != nil {
			return "", nil, err
		}
		if frontendNsgId != "" {
			logger.Infof("found frontend NSG %s", frontendNsgId)
			return frontendNsgId, etag, nil
		}
	}
	return "", nil, nil
}

// checkPendingLBWorkRequests checks if we have pending work requests before processing the LoadBalancer further
// Will error out if any in-progress work request are present for the LB
func (cp *CloudProvider) checkPendingLBWorkRequests(ctx context.Context, logger *zap.SugaredLogger, lbProvider CloudLoadBalancerProvider, lb *client.GenericLoadBalancer, service *v1.Service, startTime time.Time) (err error) {
	listWorkRequestTime := time.Now()
	loadBalancerType := getLoadBalancerType(service)
	lbName := GetLoadBalancerName(service)
	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = *lb.Id

	lbInProgressWorkRequests, err := lbProvider.lbClient.ListWorkRequests(ctx, *lb.CompartmentId, *lb.Id)
	logger.With("loadBalancerID", *lb.Id).Infof("time (in seconds) to list work-requests for LB %f", time.Since(listWorkRequestTime).Seconds())
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to list work-requests in-progress")
		errorType := util.GetError(err)
		lbMetricDimension := util.GetMetricDimensionForComponent(errorType, util.LoadBalancerType)
		dimensionsMap[metrics.ComponentDimension] = lbMetricDimension
		dimensionsMap[metrics.ResourceOCIDDimension] = lbName
		metrics.SendMetricData(cp.metricPusher, getMetric(loadBalancerType, List), time.Since(startTime).Seconds(), dimensionsMap)
		return err
	}
	for _, wr := range lbInProgressWorkRequests {
		switch loadBalancerType {
		case NLB:
			if wr.Status == string(networkloadbalancer.OperationStatusInProgress) || wr.Status == string(networkloadbalancer.OperationStatusAccepted) {
				logger.With("loadBalancerID", *lb.Id).Infof("current in-progress work requests for Network Load Balancer %s", *wr.Id)
				return errors.New("Network Load Balancer has work requests in progress, will wait and retry")
			}
		default:
			if *wr.LifecycleState == string(loadbalancer.WorkRequestLifecycleStateInProgress) || *wr.LifecycleState == string(loadbalancer.WorkRequestLifecycleStateAccepted) {
				logger.With("loadBalancerID", *lb.Id).Infof("current in-progress work requests for Load Balancer %s", *wr.Id)
				return errors.New("Load Balancer has work requests in progress, will wait and retry")
			}
		}
	}
	return
}

// checkForNetworkPartition return true if network partition is present (all nodes are not ready) else throws an error if any
func (cp *CloudProvider) checkForNetworkPartition(logger *zap.SugaredLogger, nodes []*v1.Node) (isNetworkPartition bool, err error) {
	// Service controller provided empty provisioned nodes list
	if len(nodes) == 0 {
		// List all nodes in the cluster
		nodeList, err := cp.NodeLister.List(labels.Everything())
		if err != nil {
			logger.With(zap.Error(err)).Error("Failed to check if all backend nodes are not ready, error listing nodes")
			return false, err
		}

		if len(nodeList) == 0 {
			logger.Info("Cluster has zero nodes, continue reconciling")
		} else if allNodesNotReady := cp.checkAllBackendNodesNotReady(nodeList); allNodesNotReady {
			logger.Info("Not removing backends since all nodes are Not Ready")
			return true, nil
		} else {
			err = errors.Errorf("backend node status is inconsistent, will retry")
			logger.With(zap.Error(err)).Error("Not removing backends since backend node status is inconsistent with what was observed by service controller")
			return false, err
		}
	}
	return
}
