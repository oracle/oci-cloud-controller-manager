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
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	sets "k8s.io/apimachinery/pkg/util/sets"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-go-sdk/v31/common"
	"github.com/oracle/oci-go-sdk/v31/loadbalancer"
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
	Name           string
	Shape          string
	FlexMin        *int
	FlexMax        *int
	Subnets        []string
	Internal       bool
	Listeners      map[string]loadbalancer.ListenerDetails
	BackendSets    map[string]loadbalancer.BackendSetDetails
	LoadBalancerIP string

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

	backendSets, err := getBackendSets(logger, svc, nodes, sslConfig)
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

	return &LBSpec{
		Name:           GetLoadBalancerName(svc),
		Shape:          shape,
		FlexMin:        flexShapeMinMbps,
		FlexMax:        flexShapeMaxMbps,
		Internal:       internal,
		Subnets:        subnets,
		Listeners:      listeners,
		BackendSets:    backendSets,
		LoadBalancerIP: loadbalancerIP,

		Ports:                   ports,
		SSLConfig:               sslConfig,
		SourceCIDRs:             sourceCIDRs,
		NetworkSecurityGroupIds: networkSecurityGroupIds,

		service: svc,
		nodes:   nodes,
		securityListManager: secListFactory(
			svc.Annotations[ServiceAnnotaionLoadBalancerSecurityListManagementMode]),
		FreeformTags: lbTags.FreeformTags,
		DefinedTags:  lbTags.DefinedTags,
	}, nil
}

// Certificates builds a map of required SSL certificates.
func (s *LBSpec) Certificates() (map[string]loadbalancer.CertificateDetails, error) {
	certs := make(map[string]loadbalancer.CertificateDetails)

	if s.SSLConfig == nil {
		return certs, nil
	}

	if s.SSLConfig.ListenerSSLSecretName != "" {
		cert, err := s.SSLConfig.readSSLSecret(s.SSLConfig.ListenerSSLSecretNamespace, s.SSLConfig.ListenerSSLSecretName)
		if err != nil {
			return nil, errors.Wrap(err, "reading SSL Listener Secret")
		}
		certs[s.SSLConfig.ListenerSSLSecretName] = loadbalancer.CertificateDetails{
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
		certs[s.SSLConfig.BackendSetSSLSecretName] = loadbalancer.CertificateDetails{
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
	if err := validateProtocols(svc.Spec.Ports); err != nil {
		return err
	}

	if svc.Spec.SessionAffinity != v1.ServiceAffinityNone {
		return errors.New("OCI only supports SessionAffinity \"None\" currently")
	}

	return nil
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
	for _, servicePort := range svc.Spec.Ports {
		name := getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
		healthChecker, err := getHealthChecker(svc)
		if err != nil {
			return nil, err
		}
		ports[name] = portSpec{
			BackendPort:       int(servicePort.NodePort),
			ListenerPort:      int(servicePort.Port),
			HealthCheckerPort: *healthChecker.Port,
		}
	}
	return ports, nil
}

func getBackends(logger *zap.SugaredLogger, nodes []*v1.Node, nodePort int32) []loadbalancer.BackendDetails {
	backends := make([]loadbalancer.BackendDetails, 0)
	for _, node := range nodes {
		nodeAddressString := common.String(NodeInternalIP(node))
		if *nodeAddressString == "" {
			logger.Warnf("Node %q has an empty Internal IP address.", node.Name)
			continue
		}
		backends = append(backends, loadbalancer.BackendDetails{
			IpAddress: nodeAddressString,
			Port:      common.Int(int(nodePort)),
			Weight:    common.Int(1),
		})
	}
	return backends
}

func getBackendSets(logger *zap.SugaredLogger, svc *v1.Service, nodes []*v1.Node, sslCfg *SSLConfig) (map[string]loadbalancer.BackendSetDetails, error) {
	backendSets := make(map[string]loadbalancer.BackendSetDetails)
	loadbalancerPolicy, err := getLoadBalancerPolicy(svc)
	if err != nil {
		return nil, err
	}
	for _, servicePort := range svc.Spec.Ports {
		name := getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
		port := int(servicePort.Port)
		var secretName string
		if sslCfg != nil && len(sslCfg.BackendSetSSLSecretName) != 0 {
			secretName = sslCfg.BackendSetSSLSecretName
		}
		healthChecker, err := getHealthChecker(svc)
		if err != nil {
			return nil, err
		}
		backendSets[name] = loadbalancer.BackendSetDetails{
			Policy:           common.String(loadbalancerPolicy),
			Backends:         getBackends(logger, nodes, servicePort.NodePort),
			HealthChecker:    healthChecker,
			SslConfiguration: getSSLConfiguration(sslCfg, secretName, port),
		}
	}
	return backendSets, nil
}

func getHealthChecker(svc *v1.Service) (*loadbalancer.HealthCheckerDetails, error) {
	// Setting default values as per defined in the doc (https://docs.cloud.oracle.com/en-us/iaas/Content/Balance/Tasks/editinghealthcheck.htm#console)
	var retries = 3
	if r, ok := svc.Annotations[ServiceAnnotationLoadBalancerHealthCheckRetries]; ok {
		rInt, err := strconv.Atoi(r)
		if err != nil {
			return nil, fmt.Errorf("invalid value: %s provided for annotation: %s", r, ServiceAnnotationLoadBalancerHealthCheckRetries)
		}
		retries = rInt
	}
	// Setting default values as per defined in the doc (https://docs.cloud.oracle.com/en-us/iaas/Content/Balance/Tasks/editinghealthcheck.htm#console)
	var intervalInMillis = 10000
	if i, ok := svc.Annotations[ServiceAnnotationLoadBalancerHealthCheckInterval]; ok {
		iInt, err := strconv.Atoi(i)
		if err != nil {
			return nil, fmt.Errorf("invalid value: %s provided for annotation: %s", i, ServiceAnnotationLoadBalancerHealthCheckInterval)
		}
		intervalInMillis = iInt
	}
	// Setting default values as per defined in the doc (https://docs.cloud.oracle.com/en-us/iaas/Content/Balance/Tasks/editinghealthcheck.htm#console)
	var timeoutInMillis = 3000
	if t, ok := svc.Annotations[ServiceAnnotationLoadBalancerHealthCheckTimeout]; ok {
		tInt, err := strconv.Atoi(t)
		if err != nil {
			return nil, fmt.Errorf("invalid value: %s provided for annotation: %s", t, ServiceAnnotationLoadBalancerHealthCheckTimeout)
		}
		timeoutInMillis = tInt
	}
	checkPath, checkPort := apiservice.GetServiceHealthCheckPathPort(svc)
	if checkPath != "" {
		return &loadbalancer.HealthCheckerDetails{
			Protocol:         common.String(lbNodesHealthCheckProto),
			UrlPath:          &checkPath,
			Port:             common.Int(int(checkPort)),
			Retries:          &retries,
			IntervalInMillis: &intervalInMillis,
			TimeoutInMillis:  &timeoutInMillis,
		}, nil
	}

	return &loadbalancer.HealthCheckerDetails{
		Protocol:         common.String(lbNodesHealthCheckProto),
		UrlPath:          common.String(lbNodesHealthCheckPath),
		Port:             common.Int(lbNodesHealthCheckPort),
		Retries:          &retries,
		IntervalInMillis: &intervalInMillis,
		TimeoutInMillis:  &timeoutInMillis,
	}, nil
}

func getSSLConfiguration(cfg *SSLConfig, name string, port int) *loadbalancer.SslConfigurationDetails {
	if cfg == nil || !cfg.Ports.Has(port) || len(name) == 0 {
		return nil
	}
	return &loadbalancer.SslConfigurationDetails{
		CertificateName:       &name,
		VerifyDepth:           common.Int(0),
		VerifyPeerCertificate: common.Bool(false),
	}
}

func getListeners(svc *v1.Service, sslCfg *SSLConfig) (map[string]loadbalancer.ListenerDetails, error) {
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

	listeners := make(map[string]loadbalancer.ListenerDetails)
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

		listener := loadbalancer.ListenerDetails{
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
			listener.ConnectionConfiguration = &loadbalancer.ConnectionConfiguration{
				IdleTimeout:                    actualConnectionIdleTimeout,
				BackendTcpProxyProtocolVersion: proxyProtocolVersion,
			}
		}

		listeners[name] = listener
	}

	return listeners, nil
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
	var nsgList []string
	networkSecurityGroupIds, ok := svc.Annotations[ServiceAnnotationLoadBalancerNetworkSecurityGroups]
	if !ok || networkSecurityGroupIds == "" {
		return nsgList, nil
	}

	numOfNsgIds := 0
	for _, nsgOCID := range RemoveDuplicatesFromList(strings.Split(strings.ReplaceAll(networkSecurityGroupIds, " ", ""), ",")) {
		numOfNsgIds++
		if numOfNsgIds > lbMaximumNetworkSecurityGroupIds {
			return nil, fmt.Errorf("invalid number of Network Security Groups (Max: 5) provided for annotation: %s", ServiceAnnotationLoadBalancerNetworkSecurityGroups)
		}
		if nsgOCID != "" {
			nsgList = append(nsgList, nsgOCID)
			continue
		}
		return nil, fmt.Errorf("invalid NetworkSecurityGroups OCID: [%s] provided for annotation: %s", networkSecurityGroupIds, ServiceAnnotationLoadBalancerNetworkSecurityGroups)
	}

	return nsgList, nil
}

func isInternalLB(svc *v1.Service) (bool, error) {
	if private, ok := svc.Annotations[ServiceAnnotationLoadBalancerInternal]; ok {
		internal, err := strconv.ParseBool(private)
		if err != nil {
			return false, errors.Wrap(err, fmt.Sprintf("invalid value: %s provided for annotation: %s", private, ServiceAnnotationLoadBalancerInternal))
		}
		return internal, nil
	}
	return false, nil
}

func getLBShape(svc *v1.Service) (string, *int, *int, error) {
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
	lbPolicy, ok := svc.Annotations[ServiceAnnotationLoadBalancerPolicy]
	if !ok {
		return DefaultLoadBalancerPolicy, nil
	}
	knownLBPolicies := map[string]struct{}{
		IPHashLoadBalancerPolicy:           struct{}{},
		LeastConnectionsLoadBalancerPolicy: struct{}{},
		RoundRobinLoadBalancerPolicy:       struct{}{},
	}

	if _, ok := knownLBPolicies[lbPolicy]; ok {
		return lbPolicy, nil
	}

	return "", fmt.Errorf("loadbalancer policy \"%s\" is not valid", svc.Annotations[ServiceAnnotationLoadBalancerPolicy])
}

func getLoadBalancerIP(svc *v1.Service) (string, error) {
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
	freeformTags := make(map[string]string)
	freeformTagsAnnotation, ok := svc.Annotations[ServiceAnnotationLoadBalancerInitialFreeformTagsOverride]
	if ok && freeformTagsAnnotation != "" {
		err := json.Unmarshal([]byte(freeformTagsAnnotation), &freeformTags)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse free form tags annotation")
		}
	}

	definedTags := make(map[string]map[string]interface{})
	definedTagsAnnotation, ok := svc.Annotations[ServiceAnnotationLoadBalancerInitialDefinedTagsOverride]
	if ok && definedTagsAnnotation != "" {
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
