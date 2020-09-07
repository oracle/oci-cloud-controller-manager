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
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"strings"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/loadbalancer"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	sets "k8s.io/apimachinery/pkg/util/sets"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"
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
	Name        string
	Shape       string
	Subnets     []string
	Internal    bool
	Listeners   map[string]loadbalancer.ListenerDetails
	BackendSets map[string]loadbalancer.BackendSetDetails

	Ports               map[string]portSpec
	SourceCIDRs         []string
	SSLConfig           *SSLConfig
	securityListManager securityListManager

	service *v1.Service
	nodes   []*v1.Node
}

// NewLBSpec creates a LB Spec from a Kubernetes service and a slice of nodes.
func NewLBSpec(logger *zap.SugaredLogger, svc *v1.Service, nodes []*v1.Node, subnets []string, sslConfig *SSLConfig, secListFactory securityListManagerFactory) (*LBSpec, error) {
	if err := validateService(svc); err != nil {
		return nil, errors.Wrap(err, "invalid service")
	}

	_, internal := svc.Annotations[ServiceAnnotationLoadBalancerInternal]

	// TODO (apryde): We should detect when this changes and WARN as we don't
	// support updating a load balancer's Shape.
	shape := lbDefaultShape
	if s, ok := svc.Annotations[ServiceAnnotationLoadBalancerShape]; ok {
		shape = s
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

	return &LBSpec{
		Name:        GetLoadBalancerName(svc),
		Shape:       shape,
		Internal:    internal,
		Subnets:     subnets,
		Listeners:   listeners,
		BackendSets: backendSets,

		Ports:       ports,
		SSLConfig:   sslConfig,
		SourceCIDRs: sourceCIDRs,

		service: svc,
		nodes:   nodes,
		securityListManager: secListFactory(
			svc.Annotations[ServiceAnnotaionLoadBalancerSecurityListManagementMode]),
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

	if svc.Spec.LoadBalancerIP != "" {
		// TODO(horwitz): We need to figure out in the WG if this should actually log or error.
		// The docs say: If the loadBalancerIP is specified, but the cloud provider does not support the feature, the field will be ignored.
		// But no one does that...
		// https://kubernetes.io/docs/concepts/services-networking/service/#type-loadbalancer
		return errors.New("OCI does not support setting LoadBalancerIP")
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
		healthChecker, err := getHealthChecker(nil, int(servicePort.Port), svc)
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
	for _, servicePort := range svc.Spec.Ports {
		name := getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
		port := int(servicePort.Port)
		var secretName string
		if sslCfg != nil && len(sslCfg.BackendSetSSLSecretName) != 0 {
			secretName = sslCfg.BackendSetSSLSecretName
		}
		healthChecker, err := getHealthChecker(sslCfg, port, svc)
		if err != nil {
			return nil, err
		}
		backendSets[name] = loadbalancer.BackendSetDetails{
			Policy:           common.String(DefaultLoadBalancerPolicy),
			Backends:         getBackends(logger, nodes, servicePort.NodePort),
			HealthChecker:    healthChecker,
			SslConfiguration: getSSLConfiguration(sslCfg, secretName, port),
		}
	}
	return backendSets, nil
}

func getHealthChecker(cfg *SSLConfig, port int, svc *v1.Service) (*loadbalancer.HealthCheckerDetails, error) {
	// If the health-check has SSL enabled use TCP rather than HTTP.
	protocol := lbNodesHealthCheckProtoHTTP
	if cfg != nil && cfg.Ports.Has(port) {
		protocol = lbNodesHealthCheckProtoTCP
	}
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
			Protocol: &protocol,
			UrlPath:  &checkPath,
			Port:     common.Int(int(checkPort)),
			Retries:  &retries,
			IntervalInMillis: &intervalInMillis,
			TimeoutInMillis: &timeoutInMillis,
		}, nil
	}

	return &loadbalancer.HealthCheckerDetails{
		Protocol: &protocol,
		UrlPath:  common.String(lbNodesHealthCheckPath),
		Port:     common.Int(lbNodesHealthCheckPort),
		Retries:  &retries,
		IntervalInMillis: &intervalInMillis,
		TimeoutInMillis: &timeoutInMillis,
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
