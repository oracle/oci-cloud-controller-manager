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
	"slices"
	"strconv"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"
	"k8s.io/utils/pointer"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/pkg/errors"
	helper "k8s.io/cloud-provider/service/helpers"
	net2 "k8s.io/utils/net"
)

const (
	LB                        = "lb"
	NLB                       = "nlb"
	NSG                       = "NSG"
	LBHealthCheckIntervalMin  = 1000
	LBHealthCheckIntervalMax  = 1800000
	NLBHealthCheckIntervalMin = 10000
	NLBHealthCheckIntervalMax = 1800000
	IPv4                      = string(client.GenericIPv4)
	IPv6                      = string(client.GenericIPv6)
	IPv4AndIPv6               = string("IPv4_AND_IPv6")
)

const (
	defaultLoadBalancerSourceRangesIPv4 = "0.0.0.0/0"
	defaultLoadBalancerSourceRangesIPv6 = "::/0"
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

	// ServiceAnnotationServiceAccountName is a service annotation to select Service Account to be used to
	// exchange for Workload Identity Token which can then be used for LB/NLB Client to communicate to OCI LB/NLB API.
	ServiceAnnotationServiceAccountName = "oci.oraclecloud.com/workload-service-account"

	// ServiceAnnotationLoadBalancerSecurityRuleManagementMode is a Service annotation for
	// specifying the security rule management mode ("SL-All", "SL-Frontend", "NSG", "None") that configures how security lists are managed by the CCM
	ServiceAnnotationLoadBalancerSecurityRuleManagementMode = "oci.oraclecloud.com/security-rule-management-mode"

	// ServiceAnnotationBackendSecurityRuleManagement is a service annotation to denote management of backend Network Security Group(s)
	// ingress / egress security rules for a given kubernetes service could be either LB or NLB
	ServiceAnnotationBackendSecurityRuleManagement = "oci.oraclecloud.com/oci-backend-network-security-group"

	// ServiceAnnotationLoadbalancerListenerSSLConfig is a service annotation allows you to set the cipher suite on the listener
	ServiceAnnotationLoadbalancerListenerSSLConfig = "oci.oraclecloud.com/oci-load-balancer-listener-ssl-config"

	// ServiceAnnotationLoadbalancerBackendSetSSLConfig is a service annotation allows you to set the cipher suite on the backendSet
	ServiceAnnotationLoadbalancerBackendSetSSLConfig = "oci.oraclecloud.com/oci-load-balancer-backendset-ssl-config"

	// ServiceAnnotationIngressIpMode is a service annotation allows you to set the ".status.loadBalancer.ingress.ipMode" for a Service
	// with type set to LoadBalancer.
	// https://kubernetes.io/docs/concepts/services-networking/service/#load-balancer-ip-mode:~:text=Specifying%20IPMode%20of%20load%20balancer%20status
	ServiceAnnotationIngressIpMode = "oci.oraclecloud.com/ingress-ip-mode"

	// ServiceAnnotationRuleSets allows the user to specify rule sets of actions applied to traffic at a load balancer listener
	// https://docs.oracle.com/en-us/iaas/Content/Balance/Tasks/managingrulesets.htm
	// Expected format is a JSON blob containing a JSON object literal with keys being rule names and values being a JSON
	// representation of a valid Rule object. https://docs.oracle.com/en-us/iaas/api/#/en/loadbalancer/20170115/datatypes/Rule
	ServiceAnnotationRuleSets = "oci.oraclecloud.com/oci-load-balancer-rule-sets"
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

	// ServiceAnnotationNetworkLoadBalancerIsPpv2Enabled is a service annotation to enable/disable PPv2 feature for the listeners of this NLB.
	ServiceAnnotationNetworkLoadBalancerIsPpv2Enabled = "oci-network-load-balancer.oraclecloud.com/is-ppv2-enabled"

	// ServiceAnnotationNetworkLoadBalancerExternalIpOnly is a service a boolean annotation to skip private ip when assigning to ingress resource for NLB service
	ServiceAnnotationNetworkLoadBalancerExternalIpOnly = "oci-network-load-balancer.oraclecloud.com/external-ip-only"

	// ServiceAnnotationNetworkLoadBalancerAssignedPrivateIpV4 is s service annotation to provision Network LoadBalancer with an assigned
	// IPv4 address from the subnet https://docs.oracle.com/en-us/iaas/api/#/en/networkloadbalancer/20200501/datatypes/CreateNetworkLoadBalancerDetails
	ServiceAnnotationNetworkLoadBalancerAssignedPrivateIpV4 = "oci-network-load-balancer.oraclecloud.com/assigned-private-ipv4"

	// ServiceAnnotationNetworkLoadBalancerAssignedIpV6 is s service annotation to provision Network LoadBalancer with an assigned
	// IPv6 address from the subnet https://docs.oracle.com/en-us/iaas/api/#/en/networkloadbalancer/20200501/datatypes/CreateNetworkLoadBalancerDetails
	ServiceAnnotationNetworkLoadBalancerAssignedIpV6 = "oci-network-load-balancer.oraclecloud.com/assigned-ipv6"
)

// Virtual Node Annotations
const (
	// PrivateIPOCIDAnnotation is the privateIP OCID of the Container Instance running a virtual pod
	// set by the virtual node
	PrivateIPOCIDAnnotation = "oci.oraclecloud.com/pod.info.private_ip_ocid"
)

const (
	ProtocolGrpc              = "GRPC"
	DefaultCipherSuiteForGRPC = "oci-default-http2-ssl-cipher-suite-v1"
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

type ManagedNetworkSecurityGroup struct {
	nsgRuleManagementMode string
	frontendNsgId         string
	backendNsgId          []string
}

func requiresCertificate(svc *v1.Service) bool {
	if getLoadBalancerType(svc) == NLB {
		return false
	}
	_, ok := svc.Annotations[ServiceAnnotationLoadBalancerSSLPorts]
	return ok
}

func requiresNsgManagement(svc *v1.Service) bool {
	manageNSG := strings.ToLower(svc.Annotations[ServiceAnnotationLoadBalancerSecurityRuleManagementMode])
	if manageNSG == "nsg" {
		return true
	}
	return false
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
	Type                        string
	Name                        string
	Shape                       string
	FlexMin                     *int
	FlexMax                     *int
	Subnets                     []string
	Internal                    bool
	Listeners                   map[string]client.GenericListener
	BackendSets                 map[string]client.GenericBackendSetDetails
	LoadBalancerIP              string
	IsPreserveSource            *bool
	Ports                       map[string]portSpec
	SourceCIDRs                 []string
	SSLConfig                   *SSLConfig
	securityListManager         securityListManager
	ManagedNetworkSecurityGroup *ManagedNetworkSecurityGroup
	NetworkSecurityGroupIds     []string
	IpVersions                  *IpVersions
	FreeformTags                map[string]string
	DefinedTags                 map[string]map[string]interface{}
	SystemTags                  map[string]map[string]interface{}
	ingressIpMode               *v1.LoadBalancerIPMode
	Compartment                 string
	RuleSets                    map[string]loadbalancer.RuleSetDetails
	AssignedPrivateIpv4         *string
	AssignedIpv6                *string

	service *v1.Service
	nodes   []*v1.Node
}

// NewLBSpec creates a LB Spec from a Kubernetes service and a slice of nodes.
func NewLBSpec(logger *zap.SugaredLogger, svc *v1.Service, provisionedNodes []*v1.Node, subnets []string,
	sslConfig *SSLConfig, secListFactory securityListManagerFactory, versions *IpVersions, initialLBTags *config.InitialTags,
	existingLB *client.GenericLoadBalancer, clusterCompartment string) (*LBSpec, error) {
	if err := validateService(svc); err != nil {
		return nil, errors.Wrap(err, "invalid service")
	}

	lbType := getLoadBalancerType(svc)
	ipVersions := &IpVersions{
		IpFamilyPolicy:           versions.IpFamilyPolicy,
		IpFamilies:               versions.IpFamilies,
		LbEndpointIpVersion:      versions.LbEndpointIpVersion,
		ListenerBackendIpVersion: versions.ListenerBackendIpVersion,
	}

	internal, err := isInternalLB(svc)
	if err != nil {
		return nil, err
	}

	shape, flexShapeMinMbps, flexShapeMaxMbps, err := getLBShape(svc, existingLB)
	if err != nil {
		return nil, err
	}

	sourceCIDRs, err := getLoadBalancerSourceRanges(svc)
	if err != nil {
		return nil, err
	}

	ruleSets, err := getRuleSets(svc)
	if err != nil {
		return nil, err
	}

	listeners, err := getListeners(svc, sslConfig, convertOciIpVersionsToOciIpFamilies(versions.ListenerBackendIpVersion))
	if err != nil {
		return nil, err
	}

	isPreserveSource, err := getPreserveSource(logger, svc)
	if err != nil {
		return nil, err
	}

	backendSets, err := getBackendSets(logger, svc, provisionedNodes, sslConfig, isPreserveSource, convertOciIpVersionsToOciIpFamilies(versions.ListenerBackendIpVersion))
	if err != nil {
		return nil, err
	}

	ports, err := getPorts(svc, convertOciIpVersionsToOciIpFamilies(versions.ListenerBackendIpVersion))
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
	// merge lbtags with common tags if present
	if enableOkeSystemTags && util.IsCommonTagPresent(initialLBTags) {
		lbTags = util.MergeTagConfig(lbTags, initialLBTags.Common)
	}

	ruleManagementMode, managedNsg, err := getRuleManagementMode(svc)
	if err != nil {
		return nil, err
	}

	backendNsgOcids, err := getManagedBackendNSG(svc)
	if err != nil {
		return nil, err
	}

	if managedNsg != nil && ruleManagementMode == RuleManagementModeNsg && len(backendNsgOcids) != 0 {
		managedNsg.backendNsgId = backendNsgOcids
	}

	ingressIpMode, err := getIngressIpMode(svc)
	if err != nil {
		return nil, err
	}

	compartment := getLoadBalancerCompartment(svc, clusterCompartment)

	assignedPrivateIpv4, assignedIpv6, err := getAssignedPrivateIP(logger, svc)
	if err != nil {
		return nil, err
	}

	return &LBSpec{
		Type:                        lbType,
		Name:                        GetLoadBalancerName(svc),
		Shape:                       shape,
		FlexMin:                     flexShapeMinMbps,
		FlexMax:                     flexShapeMaxMbps,
		Internal:                    internal,
		Subnets:                     subnets,
		Listeners:                   listeners,
		BackendSets:                 backendSets,
		LoadBalancerIP:              loadbalancerIP,
		IsPreserveSource:            &isPreserveSource,
		Ports:                       ports,
		SSLConfig:                   sslConfig,
		SourceCIDRs:                 sourceCIDRs,
		NetworkSecurityGroupIds:     networkSecurityGroupIds,
		ManagedNetworkSecurityGroup: managedNsg,
		service:                     svc,
		nodes:                       provisionedNodes,
		securityListManager:         secListFactory(ruleManagementMode),
		IpVersions:                  ipVersions,
		FreeformTags:                lbTags.FreeformTags,
		DefinedTags:                 lbTags.DefinedTags,
		SystemTags:                  getResourceTrackingSystemTagsFromConfig(logger, initialLBTags),
		ingressIpMode:               ingressIpMode,
		Compartment:                 compartment,
		RuleSets:                    ruleSets,
		AssignedPrivateIpv4:         assignedPrivateIpv4,
		AssignedIpv6:                assignedIpv6,
	}, nil
}

func getLoadBalancerCompartment(svc *v1.Service, clusterCompartment string) (compartment string) {
	compartment = clusterCompartment
	if value, exist := svc.Annotations[util.CompartmentIDAnnotation]; exist {
		compartment = value
	}
	return
}

func getSecurityListManagementMode(svc *v1.Service) (string, error) {
	lbType := getLoadBalancerType(svc)
	logger := *zap.L().Sugar()
	knownSecListModes := map[string]struct{}{
		ManagementModeAll:      struct{}{},
		ManagementModeNone:     struct{}{},
		ManagementModeFrontend: struct{}{},
	}
	annotationExists := false
	var annotationValue string
	switch lbType {
	case NLB:
		{
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
		annotationValue, annotationExists = svc.Annotations[ServiceAnnotationLoadBalancerSecurityListManagementMode]
		if !annotationExists {
			return ManagementModeAll, nil
		}
		if _, ok := knownSecListModes[annotationValue]; !ok {
			logger.Infof("invalid value: %s provided for annotation: %s; using default All", annotationValue, ServiceAnnotationLoadBalancerSecurityListManagementMode)
			return ManagementModeAll, nil
		}
		return svc.Annotations[ServiceAnnotationLoadBalancerSecurityListManagementMode], nil
	}
}

func getRuleManagementMode(svc *v1.Service) (string, *ManagedNetworkSecurityGroup, error) {

	knownRuleManagementModes := map[string]struct{}{
		RuleManagementModeSlAll:      struct{}{},
		RuleManagementModeSlFrontend: struct{}{},
		RuleManagementModeNsg:        struct{}{},
		ManagementModeNone:           struct{}{},
	}

	nsg := ManagedNetworkSecurityGroup{
		nsgRuleManagementMode: ManagementModeNone,
		frontendNsgId:         "",
		backendNsgId:          []string{},
	}

	annotationExists := false
	var annotationValue string
	annotationValue, annotationExists = svc.Annotations[ServiceAnnotationLoadBalancerSecurityRuleManagementMode]
	if !annotationExists {
		secListMode, err := getSecurityListManagementMode(svc)
		return secListMode, &nsg, err
	}

	if strings.ToLower(annotationValue) == strings.ToLower(RuleManagementModeSlAll) {
		return ManagementModeAll, &nsg, nil
	}
	if strings.ToLower(annotationValue) == strings.ToLower(RuleManagementModeSlFrontend) {
		return ManagementModeFrontend, &nsg, nil
	}

	if strings.ToLower(annotationValue) == strings.ToLower(RuleManagementModeNsg) {
		nsg = ManagedNetworkSecurityGroup{
			nsgRuleManagementMode: RuleManagementModeNsg,
			frontendNsgId:         "",
			backendNsgId:          []string{},
		}
		return RuleManagementModeNsg, &nsg, nil
	}

	if _, ok := knownRuleManagementModes[annotationValue]; !ok {
		return ManagementModeNone, &nsg, fmt.Errorf("invalid value: %s provided for annotation: %s", annotationValue, ServiceAnnotationLoadBalancerSecurityRuleManagementMode)
	}

	return ManagementModeNone, &nsg, nil
}

func getManagedBackendNSG(svc *v1.Service) ([]string, error) {
	backendNsgList := make([]string, 0)
	var networkSecurityGroupIds string
	var nsgAnnotationString string
	var ok bool
	networkSecurityGroupIds, ok = svc.Annotations[ServiceAnnotationBackendSecurityRuleManagement]
	nsgAnnotationString = ServiceAnnotationBackendSecurityRuleManagement
	if !ok || networkSecurityGroupIds == "" {
		return backendNsgList, nil
	}
	numOfNsgIds := 0
	for _, nsgOCID := range RemoveDuplicatesFromList(strings.Split(strings.ReplaceAll(networkSecurityGroupIds, " ", ""), ",")) {
		numOfNsgIds++
		if nsgOCID != "" {
			backendNsgList = append(backendNsgList, nsgOCID)
			continue
		}
		return nil, fmt.Errorf("invalid NetworkSecurityGroups OCID: [%s] provided for annotation: %s", networkSecurityGroupIds, nsgAnnotationString)
	}
	return backendNsgList, nil
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

	requireIPv6 := contains(getIpFamilies(service), IPv6)
	sourceCIDRs := make([]string, 0, len(sourceRanges))
	for _, sourceRange := range sourceRanges {
		sourceCIDRs = append(sourceCIDRs, sourceRange.String())
	}

	if len(sourceCIDRs) > 1 || (len(sourceCIDRs) == 1 && sourceCIDRs[0] != defaultLoadBalancerSourceRangesIPv4) {
		// User provided Loadbalancer source ranges, don't add any
		return sourceCIDRs, nil
	}

	if requireIPv6 {
		if !isServiceDualStack(service) {
			if len(sourceCIDRs) == 1 && sourceCIDRs[0] == defaultLoadBalancerSourceRangesIPv4 {
				sourceCIDRs = removeAtPosition(sourceCIDRs, 0)
			}
		}
		sourceCIDRs = append(sourceCIDRs, defaultLoadBalancerSourceRangesIPv6)
	}

	return sourceCIDRs, nil
}

func getBackendSetName(protocol string, port int) string {
	return fmt.Sprintf("%s-%d", protocol, port)
}

func getPorts(svc *v1.Service, listenerBackendIpVersion []string) (map[string]portSpec, error) {
	ports := make(map[string]portSpec)
	for backendSetName, servicePort := range getBackendSetNamePortMap(svc) {
		healthChecker, err := getHealthChecker(svc)
		if err != nil {
			return nil, err
		}
		if strings.Contains(backendSetName, IPv6) && contains(listenerBackendIpVersion, IPv6) {
			ports[backendSetName] = portSpec{
				BackendPort:       int(servicePort.NodePort),
				ListenerPort:      int(servicePort.Port),
				HealthCheckerPort: *healthChecker.Port,
			}
		} else if !strings.Contains(backendSetName, IPv6) && contains(listenerBackendIpVersion, IPv4) {
			ports[backendSetName] = portSpec{
				BackendPort:       int(servicePort.NodePort),
				ListenerPort:      int(servicePort.Port),
				HealthCheckerPort: *healthChecker.Port,
			}
		}
	}
	return ports, nil
}

func getBackends(logger *zap.SugaredLogger, provisionedNodes []*v1.Node, nodePort int32) ([]client.GenericBackend, []client.GenericBackend) {
	IPv4Backends := make([]client.GenericBackend, 0)
	IPv6Backends := make([]client.GenericBackend, 0)

	// Prepare provisioned nodes backends
	for _, node := range provisionedNodes {
		nodeAddressString := NodeInternalIP(node)
		nodeAddressStringV4 := common.String(nodeAddressString.V4)
		nodeAddressStringV6 := common.String(nodeAddressString.V6)

		if *nodeAddressStringV6 == "" {
			// Since Internal IP is optional for IPv6 populate external IP of node if present
			externalNodeAddressString := NodeExternalIp(node)
			nodeAddressStringV6 = common.String(externalNodeAddressString.V6)
		}

		if *nodeAddressStringV4 == "" && *nodeAddressStringV6 == "" {
			logger.Warnf("Node %q has an empty IP address'", node.Name)
			continue
		}
		instanceID, err := MapProviderIDToResourceID(node.Spec.ProviderID)
		if err != nil {
			logger.Warnf("Node %q has an empty ProviderID.", node.Name)
			continue
		}

		genericBackend := client.GenericBackend{
			Port:   common.Int(int(nodePort)),
			Weight: common.Int(1),
		}

		if net2.IsIPv6String(*nodeAddressStringV6) {
			// IPv6 IP
			genericBackend.IpAddress = nodeAddressStringV6
			genericBackend.TargetId = nil
			IPv6Backends = append(IPv6Backends, genericBackend)
		}
		if net2.IsIPv4String(*nodeAddressStringV4) {
			// IPv4 IP
			genericBackend.IpAddress = nodeAddressStringV4
			genericBackend.TargetId = &instanceID
			IPv4Backends = append(IPv4Backends, genericBackend)
		}
	}
	return IPv4Backends, IPv6Backends
}

func getBackendSets(logger *zap.SugaredLogger, svc *v1.Service, provisionedNodes []*v1.Node, sslCfg *SSLConfig, isPreserveSource bool, listenerBackendIpVersion []string) (map[string]client.GenericBackendSetDetails, error) {
	backendSets := make(map[string]client.GenericBackendSetDetails)
	loadbalancerPolicy, err := getLoadBalancerPolicy(svc)
	if err != nil {
		return nil, err
	}

	for backendSetName, servicePort := range getBackendSetNamePortMap(svc) {
		var secretName string
		var sslConfiguration *client.GenericSslConfigurationDetails
		if sslCfg != nil && len(sslCfg.BackendSetSSLSecretName) != 0 && getLoadBalancerType(svc) == LB {
			secretName = sslCfg.BackendSetSSLSecretName
			backendSetSSLConfig, _ := svc.Annotations[ServiceAnnotationLoadbalancerBackendSetSSLConfig]
			sslConfiguration, err = getSSLConfiguration(sslCfg, secretName, int(servicePort.Port), backendSetSSLConfig)
			if err != nil {
				return nil, err
			}
		}
		healthChecker, err := getHealthChecker(svc)
		if err != nil {
			return nil, err
		}
		backendsIPv4, backendsIPv6 := getBackends(logger, provisionedNodes, servicePort.NodePort)

		genericBackendSetDetails := client.GenericBackendSetDetails{
			Name:             common.String(backendSetName),
			Policy:           &loadbalancerPolicy,
			HealthChecker:    healthChecker,
			IsPreserveSource: &isPreserveSource,
			SslConfiguration: sslConfiguration,
		}

		if strings.Contains(backendSetName, IPv6) && contains(listenerBackendIpVersion, IPv6) {
			genericBackendSetDetails.IpVersion = GenericIpVersion(client.GenericIPv6)
			genericBackendSetDetails.Backends = backendsIPv6
			backendSets[backendSetName] = genericBackendSetDetails
		} else if !strings.Contains(backendSetName, IPv6) && contains(listenerBackendIpVersion, IPv4) {
			genericBackendSetDetails.IpVersion = GenericIpVersion(client.GenericIPv4)
			genericBackendSetDetails.Backends = backendsIPv4
			backendSets[backendSetName] = genericBackendSetDetails
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

	// Default healthcheck protocol is set to HTTP
	isForcePlainText := false

	// HTTP healthcheck for HTTPS backends
	_, ok := svc.Annotations[ServiceAnnotationLoadBalancerTLSBackendSetSecret]
	if ok {
		isForcePlainText = true
	}

	checkPath, checkPort := helper.GetServiceHealthCheckPathPort(svc)
	if checkPath != "" {
		return &client.GenericHealthChecker{
			Protocol:         lbNodesHealthCheckProto,
			IsForcePlainText: common.Bool(isForcePlainText),
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
		IsForcePlainText: common.Bool(isForcePlainText),
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

func GetSSLConfiguration(cfg *SSLConfig, name string, port int, sslConfigAnnotation string) (*client.GenericSslConfigurationDetails, error) {
	sslConfig, err := getSSLConfiguration(cfg, name, port, sslConfigAnnotation)
	if err != nil {
		return nil, err
	}
	return sslConfig, nil
}

func getSSLConfiguration(cfg *SSLConfig, name string, port int, lbSslConfigurationAnnotation string) (*client.GenericSslConfigurationDetails, error) {
	if cfg == nil || !cfg.Ports.Has(port) || len(name) == 0 {
		return nil, nil
	}
	// TODO: fast-follow to pass the sslconfiguration object directly to loadbalancer
	var extractCipherSuite *client.GenericSslConfigurationDetails

	if lbSslConfigurationAnnotation != "" {
		err := json.Unmarshal([]byte(lbSslConfigurationAnnotation), &extractCipherSuite)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse SSL Configuration annotation")
		}
	}
	genericSSLConfigurationDetails := &client.GenericSslConfigurationDetails{
		CertificateName:       &name,
		VerifyDepth:           common.Int(0),
		VerifyPeerCertificate: common.Bool(false),
	}
	if extractCipherSuite != nil {
		genericSSLConfigurationDetails.CipherSuiteName = extractCipherSuite.CipherSuiteName
		genericSSLConfigurationDetails.Protocols = extractCipherSuite.Protocols
	}

	return genericSSLConfigurationDetails, nil
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

	ruleSets, _ := getRuleSets(svc)
	var rs []string
	if ruleSets != nil {
		rs = maps.Keys(ruleSets)
		slices.Sort(rs)
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
			if strings.EqualFold(p, "HTTP") || strings.EqualFold(p, "TCP") || strings.EqualFold(p, "GRPC") {
				protocol = p
			} else {
				return nil, fmt.Errorf("invalid backend protocol %q requested for load balancer listener. Only 'HTTP', 'TCP' and 'GRPC' protocols supported", p)
			}
		}
		port := int(servicePort.Port)

		var secretName string
		var err error
		var sslConfiguration *client.GenericSslConfigurationDetails
		if sslCfg != nil && len(sslCfg.ListenerSSLSecretName) != 0 {
			secretName = sslCfg.ListenerSSLSecretName
			listenerCipherSuiteAnnotation, _ := svc.Annotations[ServiceAnnotationLoadbalancerListenerSSLConfig]
			sslConfiguration, err = getSSLConfiguration(sslCfg, secretName, port, listenerCipherSuiteAnnotation)
			if err != nil {
				return nil, err
			}
		}
		if strings.EqualFold(protocol, "GRPC") {
			protocol = ProtocolGrpc
			if sslConfiguration == nil {
				return nil, fmt.Errorf("SSL configuration cannot be empty for GRPC protocol")
			}
			if sslConfiguration.CipherSuiteName == nil {
				sslConfiguration.CipherSuiteName = common.String(DefaultCipherSuiteForGRPC)
			}
		}
		name := getListenerName(protocol, port)

		listener := client.GenericListener{
			Name:                  &name,
			DefaultBackendSetName: common.String(getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))),
			Protocol:              &protocol,
			Port:                  &port,
			RuleSetNames:          rs,
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

func getListenersNetworkLoadBalancer(svc *v1.Service, listenerBackendIpVersion []string) (map[string]client.GenericListener, error) {
	listeners := make(map[string]client.GenericListener)
	portsMap := make(map[int][]string)
	mixedProtocolsPortSet := make(map[int]bool)
	var enablePpv2 *bool

	requireIPv4, requireIPv6 := getRequireIpVersions(listenerBackendIpVersion)

	for _, servicePort := range svc.Spec.Ports {
		portsMap[int(servicePort.Port)] = append(portsMap[int(servicePort.Port)], string(servicePort.Protocol))
	}

	if ppv2EnabledValue, ppv2AnnotationSet := svc.Annotations[ServiceAnnotationNetworkLoadBalancerIsPpv2Enabled]; ppv2AnnotationSet {
		if strings.ToLower(ppv2EnabledValue) == "true" {
			enablePpv2 = pointer.Bool(true)
		} else if strings.ToLower(ppv2EnabledValue) == "false" {
			enablePpv2 = pointer.Bool(false)
		}
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

		genericListener := client.GenericListener{
			Protocol:      &protocol,
			Port:          &port,
			IsPpv2Enabled: enablePpv2,
		}
		if requireIPv4 {
			genericListener.Name = common.String(listenerName)
			genericListener.IpVersion = GenericIpVersion(client.GenericIPv4)
			genericListener.DefaultBackendSetName = common.String(backendSetName)
			listeners[listenerName] = genericListener
		}
		if requireIPv6 {
			listenerNameIPv6 := fmt.Sprintf("%s", listenerName+"-"+IPv6)
			backendSetNameIPv6 := fmt.Sprintf("%s", backendSetName+"-"+IPv6)
			genericListener.Name = common.String(listenerNameIPv6)
			genericListener.IpVersion = GenericIpVersion(client.GenericIPv6)
			genericListener.DefaultBackendSetName = common.String(backendSetNameIPv6)
			listeners[listenerNameIPv6] = genericListener
		}
	}

	return listeners, nil
}

func getListeners(svc *v1.Service, sslCfg *SSLConfig, listenerBackendIpVersion []string) (map[string]client.GenericListener, error) {

	lbType := getLoadBalancerType(svc)
	switch lbType {
	case NLB:
		{
			return getListenersNetworkLoadBalancer(svc, listenerBackendIpVersion)
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

func getLBShape(svc *v1.Service, existingLB *client.GenericLoadBalancer) (string, *int, *int, error) {
	if getLoadBalancerType(svc) == NLB {
		return flexible, nil, nil, nil
	}
	shape := lbDefaultShape
	var flexMinS, flexMaxS string
	// If existing LB is already flexible, retain the configuration done one the LB.
	// If min and max bandwith annotations are present, they take precedence over
	// what it configured on the LB.
	// This handles the situation where customers convert to flexible shape outside of K8s
	// but don't update the K8s service manifest
	if existingLB != nil && *existingLB.ShapeName == flexible {
		shape = flexible
		flexMinS = strconv.Itoa(*existingLB.ShapeDetails.MinimumBandwidthInMbps)
		flexMaxS = strconv.Itoa(*existingLB.ShapeDetails.MaximumBandwidthInMbps)
	}

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

	var flexShapeMinMbps, flexShapeMaxMbps int
	flexShapeMinMbps, err := parseFlexibleShapeBandwidth(flexMinS, ServiceAnnotationLoadBalancerShapeFlexMin)
	if err != nil {
		return "", nil, nil, err
	}

	flexShapeMaxMbps, err = parseFlexibleShapeBandwidth(flexMaxS, ServiceAnnotationLoadBalancerShapeFlexMax)
	if err != nil {
		return "", nil, nil, err
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

func getBackendSetNamePortMap(service *v1.Service) map[string]v1.ServicePort {
	backendSetPortMap := make(map[string]v1.ServicePort)

	portsMap := make(map[int][]string)
	for _, servicePort := range service.Spec.Ports {
		portsMap[int(servicePort.Port)] = append(portsMap[int(servicePort.Port)], string(servicePort.Protocol))
	}

	ipFamilies := getIpFamilies(service)
	requireIPv4, requireIPv6 := getRequireIpVersions(ipFamilies)

	mixedProtocolsPortSet := make(map[int]bool)
	for _, servicePort := range service.Spec.Ports {
		port := int(servicePort.Port)
		backendSetName := ""
		if requireIPv4 {
			if len(portsMap[port]) > 1 {
				if mixedProtocolsPortSet[port] {
					continue
				}
				backendSetName = getBackendSetName(ProtocolTypeMixed, port)
				mixedProtocolsPortSet[port] = true
			} else {
				backendSetName = getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
			}
			backendSetPortMap[backendSetName] = servicePort
		}
		if requireIPv6 {
			if len(portsMap[port]) > 1 {
				if mixedProtocolsPortSet[port] {
					continue
				}
				backendSetName = getBackendSetName(ProtocolTypeMixed, port)
				mixedProtocolsPortSet[port] = true
			} else {
				backendSetName = getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
			}
			backendSetNameIPv6 := fmt.Sprintf("%s", backendSetName+"-"+IPv6)
			backendSetPortMap[backendSetNameIPv6] = servicePort
		}

	}
	return backendSetPortMap
}

func addFrontendNsgToSpec(spec *LBSpec, frontendNsgId string) (*LBSpec, error) {
	if spec == nil {
		return nil, errors.New("service spec is empty")
	}
	if !contains(spec.NetworkSecurityGroupIds, frontendNsgId) {
		spec.NetworkSecurityGroupIds = append(spec.NetworkSecurityGroupIds, frontendNsgId)
	}
	return spec, nil
}

func updateSpecWithLbSubnets(spec *LBSpec, lbSubnetId []string) (*LBSpec, error) {
	if spec == nil {
		return nil, errors.New("service spec is empty")
	}
	spec.Subnets = lbSubnetId

	return spec, nil
}

// getIpFamilies gets ip families based on the field set in the spec
func getIpFamilies(svc *v1.Service) []string {
	ipFamilies := []string{}
	for _, ipFamily := range svc.Spec.IPFamilies {
		ipFamilies = append(ipFamilies, string(ipFamily))
	}
	return ipFamilies
}

// getIpFamilyPolicy from the service spec
func getIpFamilyPolicy(svc *v1.Service) string {
	if svc.Spec.IPFamilyPolicy == nil {
		return string(v1.IPFamilyPolicySingleStack)
	}
	return string(*svc.Spec.IPFamilyPolicy)
}

// getRequireIpVersions gets the required IP version for the service
func getRequireIpVersions(listenerBackendSetIpVersion []string) (requireIPv4, requireIPv6 bool) {
	if contains(listenerBackendSetIpVersion, IPv6) {
		requireIPv6 = true
	}
	if contains(listenerBackendSetIpVersion, IPv4) {
		requireIPv4 = true
	}
	return
}

// isServiceDualStack checks if a Service is dual-stack or not.
func isServiceDualStack(svc *v1.Service) bool {
	if svc.Spec.IPFamilyPolicy == nil {
		return false
	}
	if *svc.Spec.IPFamilyPolicy == v1.IPFamilyPolicyRequireDualStack || *svc.Spec.IPFamilyPolicy == v1.IPFamilyPolicyPreferDualStack {
		return true
	}
	return false
}

// getIngressIpMode reads ingress ipMode specified in the service annotation if exists
func getIngressIpMode(service *v1.Service) (*v1.LoadBalancerIPMode, error) {
	var ipMode, exists = "", false

	if ipMode, exists = service.Annotations[ServiceAnnotationIngressIpMode]; !exists {
		return nil, nil
	}

	switch strings.ToLower(ipMode) {
	case "proxy":
		ipModeProxy := v1.LoadBalancerIPModeProxy
		return &ipModeProxy, nil
	case "vip":
		ipModeProxy := v1.LoadBalancerIPModeVIP
		return &ipModeProxy, nil
	default:
		return nil, errors.New("IpMode can only be set as Proxy or VIP")
	}
}

// isSkipPrivateIP determines if skipPrivateIP annotation is set or not
func isSkipPrivateIP(svc *v1.Service) (bool, error) {
	lbType := getLoadBalancerType(svc)
	annotationValue := ""
	annotationExists := false
	annotationString := ""
	annotationValue, annotationExists = svc.Annotations[ServiceAnnotationNetworkLoadBalancerExternalIpOnly]
	if !annotationExists {
		return false, nil
	}

	if lbType != NLB {
		return false, nil
	}

	internal, err := isInternalLB(svc)
	if err != nil {
		return false, err
	}
	if internal {
		return false, nil
	}

	skipPrivateIp, err := strconv.ParseBool(annotationValue)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("invalid value: %s provided for annotation: %s", annotationValue, annotationString))
	}
	return skipPrivateIp, nil
}

func getRuleSets(svc *v1.Service) (rs map[string]loadbalancer.RuleSetDetails, err error) {
	annotation, exists := svc.Annotations[ServiceAnnotationRuleSets]
	if !exists {
		return nil, nil
	}

	if getLoadBalancerType(svc) == NLB {
		return rs, fmt.Errorf("invalid annotation %s. Rule Sets are not supported by Network Load Balancer", ServiceAnnotationRuleSets)
	}

	if annotation == "" {
		annotation = "{}"
	}
	err = json.NewDecoder(strings.NewReader(annotation)).Decode(&rs)
	return rs, err
}

func getAssignedPrivateIP(logger *zap.SugaredLogger, svc *v1.Service) (ipV4Adress, ipV6Adress *string, err error) {
	getIpAddress := func(key string) *string {
		address, exists := svc.Annotations[key]
		if !exists {
			return nil
		}
		if getLoadBalancerType(svc) != NLB {
			err = fmt.Errorf("Private IP assignment via annoations %s & %s is supported only in OCI Network Loadbalancer. Set %s to %s",
				ServiceAnnotationNetworkLoadBalancerAssignedPrivateIpV4,
				ServiceAnnotationNetworkLoadBalancerAssignedIpV6,
				ServiceAnnotationLoadBalancerType,
				NLB)
		}
		logger.Infof("Assigned IP address for NLB is %s", address)
		return &address
	}
	return getIpAddress(ServiceAnnotationNetworkLoadBalancerAssignedPrivateIpV4), getIpAddress(ServiceAnnotationNetworkLoadBalancerAssignedIpV6), err
}
