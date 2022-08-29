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
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/pkg/errors"
)

const (
	LB                        = "lb"
	NLB                       = "nlb"
	LBHealthCheckIntervalMin  = 1000
	LBHealthCheckIntervalMax  = 1800000
	NLBHealthCheckIntervalMin = 10000
	NLBHealthCheckIntervalMax = 1800000
)

const ProtocolTypeMixed = "TCP_AND_UDP"

const (
	// ServiceAnnotationLoadBalancerInternal is a service annotation for
	// specifying that a load balancer should be internal.
	ServiceAnnotationLoadBalancerInternal = "service.beta.kubernetes.io/oci-load-balancer-internal"

	// ServiceAnnotationLoadBalancerShape is a Service annotation for
	// specifying the Shape of a load balancer. The shape is a template that
	// determines the load balancer's total pre-provisioned maximum capacity
	// (bandwidth) for ingress plus egress traffic. Available shapes include
	// "100Mbps", "400Mbps", "8000Mbps", and "flexible". When using
	// "flexible" ,it is required to also supply
	// ServiceAnnotationLoadBalancerShapeFlexMin and
	// ServiceAnnotationLoadBalancerShapeFlexMax.
	ServiceAnnotationLoadBalancerShape = "service.beta.kubernetes.io/oci-load-balancer-shape"

	// ServiceAnnotationLoadBalancerShapeFlexMin is a Service annotation for
	// specifying the minimum bandwidth in Mbps if the LB shape is flex.
	ServiceAnnotationLoadBalancerShapeFlexMin = "service.beta.kubernetes.io/oci-load-balancer-shape-flex-min"

	// ServiceAnnotationLoadBalancerShapeFlexMax is a Service annotation for
	// specifying the maximum bandwidth in Mbps if the shape is flex.
	ServiceAnnotationLoadBalancerShapeFlexMax = "service.beta.kubernetes.io/oci-load-balancer-shape-flex-max"

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

	// ServiceAnnotationLoadBalancerConnectionProxyProtocolVersion is the annotation used
	// on the service to specify the proxy protocol version.
	ServiceAnnotationLoadBalancerConnectionProxyProtocolVersion = "service.beta.kubernetes.io/oci-load-balancer-connection-proxy-protocol-version"

	// ServiceAnnotationLoadBalancerSecurityListManagementMode is a Service annotation for
	// specifying the security list management mode ("All", "Frontend", "None") that configures how security lists are managed by the CCM
	ServiceAnnotationLoadBalancerSecurityListManagementMode = "service.beta.kubernetes.io/oci-load-balancer-security-list-management-mode"

	// ServiceAnnotationLoadBalancerHealthCheckRetries is the annotation used
	// on the service to specify the number of retries to attempt before a backend server is considered "unhealthy".
	ServiceAnnotationLoadBalancerHealthCheckRetries = "service.beta.kubernetes.io/oci-load-balancer-health-check-retries"

	// ServiceAnnotationLoadBalancerHealthCheckInterval is a Service annotation for
	// specifying the interval between health checks, in milliseconds.
	ServiceAnnotationLoadBalancerHealthCheckInterval = "service.beta.kubernetes.io/oci-load-balancer-health-check-interval"

	// ServiceAnnotationLoadBalancerHealthCheckTimeout is a Service annotation for
	// specifying the maximum time, in milliseconds, to wait for a reply to a health check. A health check is successful only if a reply
	// returns within this timeout period.
	ServiceAnnotationLoadBalancerHealthCheckTimeout = "service.beta.kubernetes.io/oci-load-balancer-health-check-timeout"

	// ServiceAnnotationLoadBalancerBEProtocol is a Service annotation for specifying the
	// load balancer listener backend protocol ("TCP", "HTTP").
	// See: https://docs.cloud.oracle.com/iaas/Content/Balance/Concepts/balanceoverview.htm#concepts
	ServiceAnnotationLoadBalancerBEProtocol = "service.beta.kubernetes.io/oci-load-balancer-backend-protocol"

	// ServiceAnnotationLoadBalancerNetworkSecurityGroups is a service annotation for
	// specifying Network security group Ids for the Loadbalancer
	ServiceAnnotationLoadBalancerNetworkSecurityGroups = "oci.oraclecloud.com/oci-network-security-groups"

	// ServiceAnnotationLoadBalancerPolicy is a service annotation for specifying
	// loadbalancer traffic policy("ROUND_ROBIN", "LEAST_CONNECTION", "IP_HASH")
	ServiceAnnotationLoadBalancerPolicy = "oci.oraclecloud.com/loadbalancer-policy"

	// ServiceAnnotationLoadBalancerInitialDefinedTagsOverride is a service annotation for specifying
	// defined tags on the LB
	ServiceAnnotationLoadBalancerInitialDefinedTagsOverride = "oci.oraclecloud.com/initial-defined-tags-override"

	// ServiceAnnotationLoadBalancerInitialFreeformTagsOverride is a service annotation for specifying
	// freeform tags on the LB
	ServiceAnnotationLoadBalancerInitialFreeformTagsOverride = "oci.oraclecloud.com/initial-freeform-tags-override"

	// ServiceAnnotationLoadBalancerType is a service annotation for specifying lb type
	ServiceAnnotationLoadBalancerType = "oci.oraclecloud.com/load-balancer-type"

	// ServiceAnnotationLoadBalancerNodeFilter is a service annotation to select specific nodes as your backend in the LB
	// based on label selector.
	ServiceAnnotationLoadBalancerNodeFilter = "oci.oraclecloud.com/node-label-selector"
)

// NLB specific annotations
const (
	// ServiceAnnotationNetworkLoadBalancerInternal is a service annotation for
	// specifying that a network load balancer should be internal
	ServiceAnnotationNetworkLoadBalancerInternal = "oci-network-load-balancer.oraclecloud.com/internal"

	// ServiceAnnotationNetworkLoadBalancerSubnet is a Service annotation for
	// specifying the first subnet of a network load balancer
	ServiceAnnotationNetworkLoadBalancerSubnet = "oci-network-load-balancer.oraclecloud.com/subnet"

	// ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups is a Service annotation for
	// specifying network security group id's for the network load balancer
	ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups = "oci-network-load-balancer.oraclecloud.com/oci-network-security-groups"

	// ServiceAnnotationNetworkLoadBalancerHealthCheckRetries is the annotation used
	// The number of retries to attempt before a backend server is considered "unhealthy".
	ServiceAnnotationNetworkLoadBalancerHealthCheckRetries = "oci-network-load-balancer.oraclecloud.com/health-check-retries"

	// ServiceAnnotationNetworkLoadBalancerHealthCheckInterval is a Service annotation for
	// The interval between health checks requests, in milliseconds.
	ServiceAnnotationNetworkLoadBalancerHealthCheckInterval = "oci-network-load-balancer.oraclecloud.com/health-check-interval"

	// ServiceAnnotationNetworkLoadBalancerHealthCheckTimeout is a Service annotation for
	// The maximum time, in milliseconds, to wait for a reply to a health check. A health check is successful only if a reply returns within this timeout period.
	ServiceAnnotationNetworkLoadBalancerHealthCheckTimeout = "oci-network-load-balancer.oraclecloud.com/health-check-timeout"

	// ServiceAnnotationNetworkLoadBalancerBackendPolicy is a Service annotation for
	// The network load balancer policy for the backend set.
	ServiceAnnotationNetworkLoadBalancerBackendPolicy = "oci-network-load-balancer.oraclecloud.com/backend-policy"

	// ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode is a Service annotation for
	// specifying the security list management mode ("All", "Frontend", "None") that configures how security lists are managed by the CCM
	ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode = "oci-network-load-balancer.oraclecloud.com/security-list-management-mode"

	// ServiceAnnotationNetworkLoadBalancerDefinedTags is a service annotation for specifying
	// defined tags on the nlb
	// DEPRECATED
	ServiceAnnotationNetworkLoadBalancerDefinedTags = "oci-network-load-balancer.oraclecloud.com/defined-tags"

	// ServiceAnnotationNetworkLoadBalancerFreeformTags is a service annotation for specifying
	// freeform tags on the nlb
	// DEPRECATED
	ServiceAnnotationNetworkLoadBalancerFreeformTags = "oci-network-load-balancer.oraclecloud.com/freeform-tags"

	// ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride is a service annotation for specifying
	// defined tags on the nlb
	ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride = "oci-network-load-balancer.oraclecloud.com/initial-defined-tags-override"

	// ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride is a service annotation for specifying
	// freeform tags on the nlb
	ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride = "oci-network-load-balancer.oraclecloud.com/initial-freeform-tags-override"

	// ServiceAnnotationNetworkLoadBalancerNodeFilter is a service annotation to select specific nodes as your backend in the NLB
	// based on label selector.
	ServiceAnnotationNetworkLoadBalancerNodeFilter = "oci-network-load-balancer.oraclecloud.com/node-label-selector"

	// ServiceAnnotationNetworkLoadBalancerIsPreserveSource is a service annotation to enable/disable preserving source information
	// on the NLB traffic. Default value when no annotation is given is to enable this for NLBs with externalTrafficPolicy=Local.
	ServiceAnnotationNetworkLoadBalancerIsPreserveSource = "oci-network-load-balancer.oraclecloud.com/is-preserve-source"
)

// certificateData is a structure containing the data about a K8S secret required
// to store SSL information required for BackendSets and Listeners
type certificateData struct {
	Name       string
	CACert     []byte
	PublicCert []byte
	PrivateKey []byte
	Passphrase []byte
}

type sslSecretReader interface {
	readSSLSecret(ns, name string) (sslSecret *certificateData, err error)
}

type noopSSLSecretReader struct{}

func (ssr noopSSLSecretReader) readSSLSecret(ns, name string) (sslSecret *certificateData, err error) {
	return nil, nil
}

// SSLConfig is a description of a SSL certificate.
type SSLConfig struct {
	Ports sets.Int

	ListenerSSLSecretName      string
	ListenerSSLSecretNamespace string

	BackendSetSSLSecretName      string
	BackendSetSSLSecretNamespace string

	sslSecretReader
}

func requiresCertificate(svc *v1.Service) bool {
	if svc.Annotations[ServiceAnnotationLoadBalancerType] == NLB {
		return false
	}
	_, ok := svc.Annotations[ServiceAnnotationLoadBalancerSSLPorts]
	return ok
}

// NewSSLConfig constructs a new SSLConfig.
func NewSSLConfig(secretListenerString string, secretBackendSetString string, service *v1.Service, ports []int, ssr sslSecretReader) *SSLConfig {
	if ssr == nil {
		ssr = noopSSLSecretReader{}
	}

	listenerSecretName, listenerSecretNamespace := getSecretParts(secretListenerString, service)
	backendSecretName, backendSecretNamespace := getSecretParts(secretBackendSetString, service)

	return &SSLConfig{
		Ports:                        sets.NewInt(ports...),
		ListenerSSLSecretName:        listenerSecretName,
		ListenerSSLSecretNamespace:   listenerSecretNamespace,
		BackendSetSSLSecretName:      backendSecretName,
		BackendSetSSLSecretNamespace: backendSecretNamespace,
		sslSecretReader:              ssr,
	}
}

// LBSpec holds the data required to build a OCI load balancer from a
// kubernetes service.
type LBSpec struct {
	Type                    string
	Name                    string
	Shape                   string
	FlexMin                 *int
	FlexMax                 *int
	Subnets                 []string
	Internal                bool
	Listeners               map[string]client.GenericListener
	BackendSets             map[string]client.GenericBackendSetDetails
	LoadBalancerIP          string
	IsPreserveSource        *bool
	Ports                   map[string]portSpec
	SourceCIDRs             []string
	SSLConfig               *SSLConfig
	securityListManager     securityListManager
	NetworkSecurityGroupIds []string
	FreeformTags            map[string]string
	DefinedTags             map[string]map[string]interface{}

	service *v1.Service
	nodes   []*v1.Node
}

// NewLBSpec creates a LB Spec from a Kubernetes service and a slice of nodes.
func NewLBSpec(logger *zap.SugaredLogger, svc *v1.Service, nodes []*v1.Node, subnets []string, sslConfig *SSLConfig, secListFactory securityListManagerFactory, initialLBTags *config.InitialTags) (*LBSpec, error) {
	if err := validateService(svc); err != nil {
		return nil, errors.Wrap(err, "invalid service")
	}

	internal, err := isInternalLB(svc)
	if err != nil {
		return nil, err
	}

	shape, flexShapeMinMbps, flexShapeMaxMbps, err := getLBShape(svc)
	if err != nil {
		return nil, err
	}

	sourceCIDRs, err := getLoadBalancerSourceRanges(svc)
	if err != nil {
		return nil, err
	}

	listeners, err := getListeners(svc, sslConfig)
	if err != nil {
		return nil, err
	}

	isPreserveSource, err := getPreserveSource(logger, svc)
	if err != nil {
		return nil, err
	}

	backendSets, err := getBackendSets(logger, svc, nodes, sslConfig, isPreserveSource)
	if err != nil {
		return nil, err
	}

	ports, err := getPorts(svc)
	if err != nil {
		return nil, err
	}

	networkSecurityGroupIds, err := getNetworkSecurityGroupIds(svc)
	if err != nil {
		return nil, err
	}

	loadbalancerIP, err := getLoadBalancerIP(svc)
	if err != nil {
		return nil, err
	}

	lbTags, err := getLoadBalancerTags(svc, initialLBTags)
	if err != nil {
		return nil, err
	}

	secListManagerMode, err := getSecurityListManagementMode(svc)
	if err != nil {
		return nil, err
	}

	lbType := getLoadBalancerType(svc)

	return &LBSpec{
		Type:                    lbType,
		Name:                    GetLoadBalancerName(svc),
		Shape:                   shape,
		FlexMin:                 flexShapeMinMbps,
		FlexMax:                 flexShapeMaxMbps,
		Internal:                internal,
		Subnets:                 subnets,
		Listeners:               listeners,
		BackendSets:             backendSets,
		LoadBalancerIP:          loadbalancerIP,
		IsPreserveSource:        &isPreserveSource,
		Ports:                   ports,
		SSLConfig:               sslConfig,
		SourceCIDRs:             sourceCIDRs,
		NetworkSecurityGroupIds: networkSecurityGroupIds,
		service:                 svc,
		nodes:                   nodes,
		securityListManager:     secListFactory(secListManagerMode),
		FreeformTags:            lbTags.FreeformTags,
		DefinedTags:             lbTags.DefinedTags,
	}, nil
}

func getSecurityListManagementMode(svc *v1.Service) (string, error) {
	lbType := getLoadBalancerType(svc)
	knownSecListModes := map[string]struct{}{
		ManagementModeAll:      struct{}{},
		ManagementModeNone:     struct{}{},
		ManagementModeFrontend: struct{}{},
	}

	switch lbType {
	case NLB:
		{
			annotationExists := false
			var annotationValue string
			annotationValue, annotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode]
			if !annotationExists {
				return ManagementModeNone, nil
			}
			if _, ok := knownSecListModes[annotationValue]; !ok {
				return "", fmt.Errorf("invalid value: %s provided for annotation: %s", annotationValue, ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode)
			}
			return svc.Annotations[ServiceAnnotationNetworkLoadBalancerSecurityListManagementMode], nil
		}
	default:
		return svc.Annotations[ServiceAnnotationLoadBalancerSecurityListManagementMode], nil
	}
}

// Certificates builds a map of required SSL certificates.
func (s *LBSpec) Certificates() (map[string]client.GenericCertificate, error) {
	certs := make(map[string]client.GenericCertificate)

	if s.SSLConfig == nil {
		return certs, nil
	}

	if s.SSLConfig.ListenerSSLSecretName != "" {
		cert, err := s.SSLConfig.readSSLSecret(s.SSLConfig.ListenerSSLSecretNamespace, s.SSLConfig.ListenerSSLSecretName)
		if err != nil {
			return nil, errors.Wrap(err, "reading SSL Listener Secret")
		}
		certs[s.SSLConfig.ListenerSSLSecretName] = client.GenericCertificate{
			CertificateName:   &s.SSLConfig.ListenerSSLSecretName,
			CaCertificate:     common.String(string(cert.CACert)),
			PublicCertificate: common.String(string(cert.PublicCert)),
			PrivateKey:        common.String(string(cert.PrivateKey)),
			Passphrase:        common.String(string(cert.Passphrase)),
		}
	}

	if s.SSLConfig.BackendSetSSLSecretName != "" {
		cert, err := s.SSLConfig.readSSLSecret(s.SSLConfig.BackendSetSSLSecretNamespace, s.SSLConfig.BackendSetSSLSecretName)
		if err != nil {
			return nil, errors.Wrap(err, "reading SSL Backend Secret")
		}
		certs[s.SSLConfig.BackendSetSSLSecretName] = client.GenericCertificate{
			CertificateName:   &s.SSLConfig.BackendSetSSLSecretName,
			CaCertificate:     common.String(string(cert.CACert)),
			PublicCertificate: common.String(string(cert.PublicCert)),
			PrivateKey:        common.String(string(cert.PrivateKey)),
			Passphrase:        common.String(string(cert.Passphrase)),
		}
	}
	return certs, nil
}

// TODO(apryde): aggregate errors using an error list.
func validateService(svc *v1.Service) error {
	secListMgmtMode, err := getSecurityListManagementMode(svc)
	if err != nil {
		return err
	}

	lbType := getLoadBalancerType(svc)

	if err := validateProtocols(svc.Spec.Ports, lbType, secListMgmtMode); err != nil {
		return err
	}

	if svc.Spec.SessionAffinity != v1.ServiceAffinityNone {
		return errors.New("OCI only supports SessionAffinity \"None\" currently")
	}

	return nil
}

func getPreserveSource(logger *zap.SugaredLogger, svc *v1.Service) (bool, error) {
	if svc.Annotations[ServiceAnnotationLoadBalancerType] != NLB {
		return false, nil
	}
	// fail the request if externalTrafficPolicy is set to Cluster and is-preserve-source annotation is set
	if svc.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyTypeCluster {
		_, ok := svc.Annotations[ServiceAnnotationNetworkLoadBalancerIsPreserveSource]
		if ok {
			logger.Error("error : externalTrafficPolicy is set to Cluster and the %s annotation is set", ServiceAnnotationNetworkLoadBalancerIsPreserveSource)
			return false, fmt.Errorf("%s annotation cannot be set when externalTrafficPolicy is set to Cluster", ServiceAnnotationNetworkLoadBalancerIsPreserveSource)
		}
	}

	enablePreservation, err := getPreserveSourceAnnotation(logger, svc)
	if err != nil {
		return false, err
	}
	if svc.Spec.ExternalTrafficPolicy == v1.ServiceExternalTrafficPolicyTypeLocal && enablePreservation {
		return true, nil
	}
	return false, nil
}

func getPreserveSourceAnnotation(logger *zap.SugaredLogger, svc *v1.Service) (bool, error) {
	if annotationString, ok := svc.Annotations[ServiceAnnotationNetworkLoadBalancerIsPreserveSource]; ok {
		enable, err := strconv.ParseBool(annotationString)
		if err != nil {
			logger.Error("failed to to parse %s annotation value - %s", ServiceAnnotationNetworkLoadBalancerIsPreserveSource, annotationString)
			return false, fmt.Errorf("failed to to parse %s annotation value - %s", ServiceAnnotationNetworkLoadBalancerIsPreserveSource, annotationString)
		}
		return enable, nil
	}
	// default behavior is to enable source destination preservation
	return true, nil
}

func getLoadBalancerSourceRanges(service *v1.Service) ([]string, error) {
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

func getBackendSetName(protocol string, port int) string {
	return fmt.Sprintf("%s-%d", protocol, port)
}

func getPorts(svc *v1.Service) (map[string]portSpec, error) {
	ports := make(map[string]portSpec)
	portsMap := make(map[int][]string)
	mixedProtocolsPortSet := make(map[int]bool)
	for _, servicePort := range svc.Spec.Ports {
		portsMap[int(servicePort.Port)] = append(portsMap[int(servicePort.Port)], string(servicePort.Protocol))
	}
	for _, servicePort := range svc.Spec.Ports {
		port := int(servicePort.Port)
		backendSetName := ""
		if len(portsMap[port]) > 1 {
			if mixedProtocolsPortSet[port] {
				continue
			}
			backendSetName = getBackendSetName(ProtocolTypeMixed, port)
			mixedProtocolsPortSet[port] = true
		} else {
			backendSetName = getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
		}
		healthChecker, err := getHealthChecker(svc)
		if err != nil {
			return nil, err
		}
		ports[backendSetName] = portSpec{
			BackendPort:       int(servicePort.NodePort),
			ListenerPort:      int(servicePort.Port),
			HealthCheckerPort: *healthChecker.Port,
		}
	}
	return ports, nil
}

func getBackends(logger *zap.SugaredLogger, nodes []*v1.Node, nodePort int32) []client.GenericBackend {
	backends := make([]client.GenericBackend, 0)
	for _, node := range nodes {
		nodeAddressString := common.String(NodeInternalIP(node))
		if *nodeAddressString == "" {
			logger.Warnf("Node %q has an empty Internal IP address.", node.Name)
			continue
		}
		instanceID, err := MapProviderIDToInstanceID(node.Spec.ProviderID)
		if err != nil {
			logger.Warnf("Node %q has an empty ProviderID.", node.Name)
			continue
		}

		backends = append(backends, client.GenericBackend{
			IpAddress: nodeAddressString,
			Port:      common.Int(int(nodePort)),
			Weight:    common.Int(1),
			TargetId:  &instanceID,
		})
	}
	return backends
}

func getBackendSets(logger *zap.SugaredLogger, svc *v1.Service, nodes []*v1.Node, sslCfg *SSLConfig, isPreserveSource bool) (map[string]client.GenericBackendSetDetails, error) {
	backendSets := make(map[string]client.GenericBackendSetDetails)
	loadbalancerPolicy, err := getLoadBalancerPolicy(svc)
	if err != nil {
		return nil, err
	}
	portsMap := make(map[int][]string)
	mixedProtocolsPortSet := make(map[int]bool)
	for _, servicePort := range svc.Spec.Ports {
		portsMap[int(servicePort.Port)] = append(portsMap[int(servicePort.Port)], string(servicePort.Protocol))
	}
	for _, servicePort := range svc.Spec.Ports {
		port := int(servicePort.Port)
		backendSetName := ""
		if len(portsMap[port]) > 1 {
			if mixedProtocolsPortSet[port] {
				continue
			}
			backendSetName = getBackendSetName(ProtocolTypeMixed, port)
			mixedProtocolsPortSet[port] = true
		} else {
			backendSetName = getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
		}
		var secretName string
		if sslCfg != nil && len(sslCfg.BackendSetSSLSecretName) != 0 {
			secretName = sslCfg.BackendSetSSLSecretName
		}
		healthChecker, err := getHealthChecker(svc)
		if err != nil {
			return nil, err
		}
		backendSets[backendSetName] = client.GenericBackendSetDetails{
			Policy:           &loadbalancerPolicy,
			Backends:         getBackends(logger, nodes, servicePort.NodePort),
			HealthChecker:    healthChecker,
			IsPreserveSource: &isPreserveSource,
			SslConfiguration: getSSLConfiguration(sslCfg, secretName, port),
		}
	}
	return backendSets, nil
}

func getHealthChecker(svc *v1.Service) (*client.GenericHealthChecker, error) {

	retries, err := getHealthCheckRetries(svc)
	if err != nil {
		return nil, err
	}

	intervalInMillis, err := getHealthCheckInterval(svc)
	if err != nil {
		return nil, err
	}

	timeoutInMillis, err := getHealthCheckTimeout(svc)
	if err != nil {
		return nil, err
	}

	checkPath, checkPort := apiservice.GetServiceHealthCheckPathPort(svc)
	if checkPath != "" {
		return &client.GenericHealthChecker{
			Protocol:         lbNodesHealthCheckProto,
			UrlPath:          &checkPath,
			Port:             common.Int(int(checkPort)),
			Retries:          &retries,
			IntervalInMillis: &intervalInMillis,
			TimeoutInMillis:  &timeoutInMillis,
			ReturnCode:       common.Int(http.StatusOK),
		}, nil
	}

	return &client.GenericHealthChecker{
		Protocol:         lbNodesHealthCheckProto,
		UrlPath:          common.String(lbNodesHealthCheckPath),
		Port:             common.Int(lbNodesHealthCheckPort),
		Retries:          &retries,
		IntervalInMillis: &intervalInMillis,
		TimeoutInMillis:  &timeoutInMillis,
		ReturnCode:       common.Int(http.StatusOK),
	}, nil
}

func getHealthCheckRetries(svc *v1.Service) (int, error) {
	lbType := getLoadBalancerType(svc)
	var retries = 3
	annotationValue := ""
	annotationExists := false
	annotationString := ""
	switch lbType {
	case NLB:
		{
			annotationValue, annotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerHealthCheckRetries]
			annotationString = ServiceAnnotationNetworkLoadBalancerHealthCheckRetries
		}
	default:
		{
			annotationValue, annotationExists = svc.Annotations[ServiceAnnotationLoadBalancerHealthCheckRetries]
			annotationString = ServiceAnnotationLoadBalancerHealthCheckRetries
		}
	}
	if annotationExists {
		rInt, err := strconv.Atoi(annotationValue)
		if err != nil {
			return -1, fmt.Errorf("invalid value: %s provided for annotation: %s", annotationValue, annotationString)
		}
		retries = rInt
	}
	return retries, nil
}

func getHealthCheckInterval(svc *v1.Service) (int, error) {
	lbType := getLoadBalancerType(svc)
	var intervalInMillis = 10000
	annotationValue := ""
	annotationExists := false
	annotationString := ""
	HealthCheckIntervalMin := LBHealthCheckIntervalMin
	HealthCheckIntervalMax := LBHealthCheckIntervalMax
	switch lbType {
	case NLB:
		{
			annotationValue, annotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerHealthCheckInterval]
			annotationString = ServiceAnnotationNetworkLoadBalancerHealthCheckInterval
			HealthCheckIntervalMin = NLBHealthCheckIntervalMin
			HealthCheckIntervalMax = NLBHealthCheckIntervalMax
		}
	default:
		{
			annotationValue, annotationExists = svc.Annotations[ServiceAnnotationLoadBalancerHealthCheckInterval]
			annotationString = ServiceAnnotationLoadBalancerHealthCheckInterval
		}
	}
	if annotationExists {
		iInt, err := strconv.Atoi(annotationValue)
		if err != nil {
			return -1, fmt.Errorf("invalid value: %s provided for annotation: %s", annotationValue, annotationString)
		}
		intervalInMillis = iInt
		if intervalInMillis < HealthCheckIntervalMin || intervalInMillis > HealthCheckIntervalMax {
			return -1, fmt.Errorf("invalid value for health check interval, should be between %v and %v", HealthCheckIntervalMin, HealthCheckIntervalMax)
		}
	}
	return intervalInMillis, nil

}

func getHealthCheckTimeout(svc *v1.Service) (int, error) {
	lbType := getLoadBalancerType(svc)
	var timeoutInMillis = 3000
	annotationValue := ""
	annotationExists := false
	annotationString := ""
	switch lbType {
	case NLB:
		{
			annotationValue, annotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerHealthCheckTimeout]
			annotationString = ServiceAnnotationNetworkLoadBalancerHealthCheckTimeout
		}
	default:
		{
			annotationValue, annotationExists = svc.Annotations[ServiceAnnotationLoadBalancerHealthCheckTimeout]
			annotationString = ServiceAnnotationLoadBalancerHealthCheckTimeout
		}
	}

	if annotationExists {
		tInt, err := strconv.Atoi(annotationValue)
		if err != nil {
			return -1, fmt.Errorf("invalid value: %s provided for annotation: %s", annotationValue, annotationString)
		}
		timeoutInMillis = tInt
	}
	return timeoutInMillis, nil
}

func getSSLConfiguration(cfg *SSLConfig, name string, port int) *client.GenericSslConfigurationDetails {
	if cfg == nil || !cfg.Ports.Has(port) || len(name) == 0 {
		return nil
	}
	return &client.GenericSslConfigurationDetails{
		CertificateName:       &name,
		VerifyDepth:           common.Int(0),
		VerifyPeerCertificate: common.Bool(false),
	}
}

func getListenersOciLoadBalancer(svc *v1.Service, sslCfg *SSLConfig) (map[string]client.GenericListener, error) {
	// Determine if connection idle timeout has been specified
	var connectionIdleTimeout *int64
	connectionIdleTimeoutAnnotation := svc.Annotations[ServiceAnnotationLoadBalancerConnectionIdleTimeout]
	if connectionIdleTimeoutAnnotation != "" {
		timeout, err := strconv.ParseInt(connectionIdleTimeoutAnnotation, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing service annotation: %s=%s",
				ServiceAnnotationLoadBalancerConnectionIdleTimeout,
				connectionIdleTimeoutAnnotation,
			)
		}

		connectionIdleTimeout = common.Int64(timeout)
	}

	// Determine if proxy protocol has been specified
	var proxyProtocolVersion *int
	proxyProtocolVersionAnnotation := svc.Annotations[ServiceAnnotationLoadBalancerConnectionProxyProtocolVersion]
	if proxyProtocolVersionAnnotation != "" {
		version, err := strconv.Atoi(proxyProtocolVersionAnnotation)
		if err != nil {
			return nil, fmt.Errorf("error parsing service annotation: %s=%s",
				ServiceAnnotationLoadBalancerConnectionProxyProtocolVersion,
				proxyProtocolVersionAnnotation,
			)
		}

		proxyProtocolVersion = common.Int(version)
	}

	listeners := make(map[string]client.GenericListener)
	for _, servicePort := range svc.Spec.Ports {
		protocol := string(servicePort.Protocol)
		// Annotation overrides the protocol.
		if p, ok := svc.Annotations[ServiceAnnotationLoadBalancerBEProtocol]; ok {
			// Default
			if p == "" {
				p = DefaultLoadBalancerBEProtocol
			}
			if strings.EqualFold(p, "HTTP") || strings.EqualFold(p, "TCP") {
				protocol = p
			} else {
				return nil, fmt.Errorf("invalid backend protocol %q requested for load balancer listener. Only 'HTTP' and 'TCP' protocols supported", p)
			}
		}
		port := int(servicePort.Port)
		var secretName string
		if sslCfg != nil && len(sslCfg.ListenerSSLSecretName) != 0 {
			secretName = sslCfg.ListenerSSLSecretName
		}
		sslConfiguration := getSSLConfiguration(sslCfg, secretName, port)
		name := getListenerName(protocol, port)

		listener := client.GenericListener{
			Name:                  &name,
			DefaultBackendSetName: common.String(getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))),
			Protocol:              &protocol,
			Port:                  &port,
			SslConfiguration:      sslConfiguration,
		}

		// If proxy protocol has been set, we also need to set connectionIdleTimeout
		// because it's a required parameter as per the LB API contract.
		// The default value is dependent on the protocol used for the listener.
		actualConnectionIdleTimeout := connectionIdleTimeout
		if proxyProtocolVersion != nil && connectionIdleTimeout == nil {
			// At that point LB only supports HTTP and TCP
			defaultIdleTimeoutPerProtocol := map[string]int64{
				"HTTP": lbConnectionIdleTimeoutHTTP,
				"TCP":  lbConnectionIdleTimeoutTCP,
			}
			actualConnectionIdleTimeout = common.Int64(defaultIdleTimeoutPerProtocol[strings.ToUpper(protocol)])
		}

		if actualConnectionIdleTimeout != nil {
			listener.ConnectionConfiguration = &client.GenericConnectionConfiguration{
				IdleTimeout:                    actualConnectionIdleTimeout,
				BackendTcpProxyProtocolVersion: proxyProtocolVersion,
			}
		}

		listeners[name] = listener
	}

	return listeners, nil
}

func getListenersNetworkLoadBalancer(svc *v1.Service) (map[string]client.GenericListener, error) {
	listeners := make(map[string]client.GenericListener)
	portsMap := make(map[int][]string)
	mixedProtocolsPortSet := make(map[int]bool)
	for _, servicePort := range svc.Spec.Ports {
		portsMap[int(servicePort.Port)] = append(portsMap[int(servicePort.Port)], string(servicePort.Protocol))
	}
	for _, servicePort := range svc.Spec.Ports {
		protocol := string(servicePort.Protocol)

		protocolMap := map[string]bool{
			"TCP": true,
			"UDP": true,
		}
		if !protocolMap[protocol] {
			return nil, fmt.Errorf("invalid backend protocol %q requested for network load balancer listener", protocol)
		}
		port := int(servicePort.Port)
		listenerName := ""
		backendSetName := ""
		if len(portsMap[port]) > 1 {
			if mixedProtocolsPortSet[port] {
				continue
			}
			listenerName = getListenerName(ProtocolTypeMixed, port)
			backendSetName = getBackendSetName(ProtocolTypeMixed, port)
			protocol = ProtocolTypeMixed
			mixedProtocolsPortSet[port] = true
		} else {
			listenerName = getListenerName(protocol, port)
			backendSetName = getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
		}

		listener := client.GenericListener{
			Name:                  &listenerName,
			DefaultBackendSetName: common.String(backendSetName),
			Protocol:              &protocol,
			Port:                  &port,
		}

		listeners[listenerName] = listener
	}

	return listeners, nil
}

func getListeners(svc *v1.Service, sslCfg *SSLConfig) (map[string]client.GenericListener, error) {

	lbType := getLoadBalancerType(svc)
	switch lbType {
	case NLB:
		{
			return getListenersNetworkLoadBalancer(svc)
		}
	default:
		{
			return getListenersOciLoadBalancer(svc, sslCfg)
		}
	}
}

func getSecretParts(secretString string, service *v1.Service) (name string, namespace string) {
	if secretString == "" {
		return "", ""
	}
	if !strings.Contains(secretString, "/") {
		return secretString, service.Namespace
	}
	parts := strings.Split(secretString, "/")
	return parts[1], parts[0]
}

func getNetworkSecurityGroupIds(svc *v1.Service) ([]string, error) {
	lbType := getLoadBalancerType(svc)
	nsgList := make([]string, 0)
	var networkSecurityGroupIds string
	var nsgAnnotationString string
	var ok bool
	switch lbType {
	case NLB:
		{
			networkSecurityGroupIds, ok = svc.Annotations[ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups]
			nsgAnnotationString = ServiceAnnotationNetworkLoadBalancerNetworkSecurityGroups
		}
	default:
		{
			networkSecurityGroupIds, ok = svc.Annotations[ServiceAnnotationLoadBalancerNetworkSecurityGroups]
			nsgAnnotationString = ServiceAnnotationLoadBalancerNetworkSecurityGroups
		}
	}
	if !ok || networkSecurityGroupIds == "" {
		return nsgList, nil
	}
	numOfNsgIds := 0
	for _, nsgOCID := range RemoveDuplicatesFromList(strings.Split(strings.ReplaceAll(networkSecurityGroupIds, " ", ""), ",")) {
		numOfNsgIds++
		if numOfNsgIds > lbMaximumNetworkSecurityGroupIds {
			return nil, fmt.Errorf("invalid number of Network Security Groups (Max: 5) provided for annotation: %s", nsgAnnotationString)
		}
		if nsgOCID != "" {
			nsgList = append(nsgList, nsgOCID)
			continue
		}
		return nil, fmt.Errorf("invalid NetworkSecurityGroups OCID: [%s] provided for annotation: %s", networkSecurityGroupIds, nsgAnnotationString)
	}

	return nsgList, nil
}

func isInternalLB(svc *v1.Service) (bool, error) {
	lbType := getLoadBalancerType(svc)
	annotationValue := ""
	annotationExists := false
	annotationString := ""
	switch lbType {
	case NLB:
		annotationValue, annotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerInternal]
		annotationString = ServiceAnnotationNetworkLoadBalancerInternal
	default:
		annotationValue, annotationExists = svc.Annotations[ServiceAnnotationLoadBalancerInternal]
		annotationString = ServiceAnnotationLoadBalancerInternal
	}
	if annotationExists {
		internal, err := strconv.ParseBool(annotationValue)
		if err != nil {
			return false, errors.Wrap(err, fmt.Sprintf("invalid value: %s provided for annotation: %s", annotationValue, annotationString))
		}
		return internal, nil
	}
	return false, nil
}

func getLBShape(svc *v1.Service) (string, *int, *int, error) {
	if getLoadBalancerType(svc) == NLB {
		return flexible, nil, nil, nil
	}
	shape := lbDefaultShape
	if s, ok := svc.Annotations[ServiceAnnotationLoadBalancerShape]; ok {
		shape = s
	}

	// For flexible shape, LBaaS requires the ShapeName to be in lower case `flexible`
	// but they have a public documentation bug where it is mentioned as `Flexible`
	// We are converting to lowercase to check the shape name and send to LBaaS
	shapeLower := strings.ToLower(shape)

	if shapeLower == flexible {
		shape = flexible
	}

	// if it's not a flexshape LB return the ShapeName as the shape
	if shape != flexible {
		return shape, nil, nil, nil
	}

	var flexMinS, flexMaxS string
	var flexShapeMinMbps, flexShapeMaxMbps int

	if fmin, ok := svc.Annotations[ServiceAnnotationLoadBalancerShapeFlexMin]; ok {
		flexMinS = fmin
	}

	if fmax, ok := svc.Annotations[ServiceAnnotationLoadBalancerShapeFlexMax]; ok {
		flexMaxS = fmax
	}

	if flexMinS == "" || flexMaxS == "" {
		return "", nil, nil, fmt.Errorf("error parsing service annotation: %s=flexible requires %s and %s to be set",
			ServiceAnnotationLoadBalancerShape,
			ServiceAnnotationLoadBalancerShapeFlexMin,
			ServiceAnnotationLoadBalancerShapeFlexMax,
		)
	}

	flexShapeMinMbps, err := strconv.Atoi(flexMinS)
	if err != nil {
		return "", nil, nil, errors.Wrap(err,
			fmt.Sprintf("The annotation %s should contain only integer value", ServiceAnnotationLoadBalancerShapeFlexMin))
	}
	flexShapeMaxMbps, err = strconv.Atoi(flexMaxS)
	if err != nil {
		return "", nil, nil, errors.Wrap(err,
			fmt.Sprintf("The annotation %s should contain only integer value", ServiceAnnotationLoadBalancerShapeFlexMax))
	}
	if flexShapeMinMbps < 10 {
		flexShapeMinMbps = 10
	}
	if flexShapeMaxMbps < 10 {
		flexShapeMaxMbps = 10
	}
	if flexShapeMinMbps > 8192 {
		flexShapeMinMbps = 8192
	}
	if flexShapeMaxMbps > 8192 {
		flexShapeMaxMbps = 8192
	}
	if flexShapeMaxMbps < flexShapeMinMbps {
		flexShapeMaxMbps = flexShapeMinMbps
	}

	return shape, &flexShapeMinMbps, &flexShapeMaxMbps, nil
}

func getLoadBalancerPolicy(svc *v1.Service) (string, error) {
	lbType := getLoadBalancerType(svc)
	annotationValue := ""
	annotationExists := false

	knownLBPolicies := map[string]struct{}{
		IPHashLoadBalancerPolicy:           struct{}{},
		LeastConnectionsLoadBalancerPolicy: struct{}{},
		RoundRobinLoadBalancerPolicy:       struct{}{},
	}

	knownNLBPolicies := map[string]struct{}{
		NetworkLoadBalancingPolicyTwoTuple:   struct{}{},
		NetworkLoadBalancingPolicyThreeTuple: struct{}{},
		NetworkLoadBalancingPolicyFiveTuple:  struct{}{},
	}

	switch lbType {
	case NLB:
		{
			annotationValue, annotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerBackendPolicy]
			if !annotationExists {
				return DefaultNetworkLoadBalancerPolicy, nil
			}
			if _, ok := knownNLBPolicies[annotationValue]; ok {
				return annotationValue, nil
			}
		}
	default:
		{
			annotationValue, annotationExists = svc.Annotations[ServiceAnnotationLoadBalancerPolicy]
			if !annotationExists {
				return DefaultLoadBalancerPolicy, nil
			}
			if _, ok := knownLBPolicies[annotationValue]; ok {
				return annotationValue, nil
			}
		}
	}

	return "", fmt.Errorf("loadbalancer policy \"%s\" is not valid", annotationValue)
}

func getLoadBalancerIP(svc *v1.Service) (string, error) {
	// There are no changes here wrt NLB since NLB doesn't support private ip reservation

	ipAddress := svc.Spec.LoadBalancerIP
	if ipAddress == "" {
		return "", nil
	}

	//checks the validity of loadbalancerIP format
	if net.ParseIP(ipAddress) == nil {
		return "", fmt.Errorf("invalid value %q provided for LoadBalancerIP", ipAddress)
	}

	//checks if private loadbalancer is trying to use reservedIP
	isInternal, err := isInternalLB(svc)
	if isInternal {
		return "", fmt.Errorf("invalid service: cannot create a private load balancer with Reserved IP")
	}
	return ipAddress, err
}

func getLoadBalancerTags(svc *v1.Service, initialTags *config.InitialTags) (*config.TagConfig, error) {
	lbType := getLoadBalancerType(svc)
	var freeformTagsAnnotation string
	var definedTagsAnnotation string
	var freeformTagsAnnotationExists bool
	var definedTagsAnnotationExists bool

	freeformTags := make(map[string]string)
	definedTags := make(map[string]map[string]interface{})

	switch lbType {
	case NLB:
		{
			freeformTagsAnnotation, freeformTagsAnnotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerInitialFreeformTagsOverride]
			definedTagsAnnotation, definedTagsAnnotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerInitialDefinedTagsOverride]

			if !freeformTagsAnnotationExists {
				freeformTagsAnnotation, freeformTagsAnnotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerFreeformTags]
			}

			if !definedTagsAnnotationExists {
				definedTagsAnnotation, definedTagsAnnotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerDefinedTags]
			}
		}
	default:
		{
			freeformTagsAnnotation, freeformTagsAnnotationExists = svc.Annotations[ServiceAnnotationLoadBalancerInitialFreeformTagsOverride]
			definedTagsAnnotation, definedTagsAnnotationExists = svc.Annotations[ServiceAnnotationLoadBalancerInitialDefinedTagsOverride]
		}
	}

	if freeformTagsAnnotationExists && freeformTagsAnnotation != "" {
		err := json.Unmarshal([]byte(freeformTagsAnnotation), &freeformTags)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse free form tags annotation")
		}
	}

	if definedTagsAnnotationExists && definedTagsAnnotation != "" {
		err := json.Unmarshal([]byte(definedTagsAnnotation), &definedTags)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse defined tags annotation")
		}
	}

	var resourceTags config.TagConfig
	// if no tags are provided freeformTags will be nil
	if len(freeformTags) > 0 {
		resourceTags.FreeformTags = freeformTags
	}

	if len(definedTags) > 0 {
		resourceTags.DefinedTags = definedTags
	}

	// if resource level tags are present return resource level tags
	if len(freeformTags) > 0 || len(definedTags) > 0 {
		return &resourceTags, nil
	}

	// if initialTags are not set return nil tags
	if initialTags == nil || initialTags.LoadBalancer == nil {
		return &config.TagConfig{}, nil
	}
	// return initial tags
	return initialTags.LoadBalancer, nil
}

func getLoadBalancerType(svc *v1.Service) string {
	lbType := strings.ToLower(svc.Annotations[ServiceAnnotationLoadBalancerType])
	switch lbType {
	case NLB, LB:
		return lbType
	default:
		return LB
	}
}
