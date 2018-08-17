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
	"strconv"
	"strings"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/loadbalancer"
	"github.com/pkg/errors"

	"k8s.io/api/core/v1"
	sets "k8s.io/apimachinery/pkg/util/sets"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/util"
)

// certificateData is a structure containing the data about a K8S secret required
// to store SSL information required for BackendSets and Listeners
type certificateData struct {
	CACert     string
	PublicCert string
	PrivateKey string
	Passphrase string
}

type sslSecretReader interface {
	readSSLSecret(secretType string, svc *v1.Service) (sslSecret *certificateData, err error)
}

type noopSSLSecretReader struct{}

func (ssr noopSSLSecretReader) readSSLSecret(secretType string, svc *v1.Service) (sslSecret *certificateData, err error) {
	return &certificateData{}, nil
}

// SSLConfig is a description of a SSL certificate.
type SSLConfig struct {
	Name                string
	Type                string
	Ports               sets.Int
	ListenerSSLSecret   *certificateData
	BackendSetSSLSecret *certificateData

	sslSecretReader
}

func requiresCertificate(svc *v1.Service) bool {
	_, ok := svc.Annotations[ServiceAnnotationLoadBalancerSSLPorts]
	return ok
}

// NewSSLConfig constructs a new SSLConfig.
func NewSSLConfig(name string, sslType string, ports []int, ssr sslSecretReader) *SSLConfig {
	if ssr == nil {
		ssr = noopSSLSecretReader{}
	}
	return &SSLConfig{
		Name:            name,
		Type:            sslType,
		Ports:           sets.NewInt(ports...),
		sslSecretReader: ssr,
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
	ListenerSSLConfig   *SSLConfig
	BackendSetSSLConfig *SSLConfig
	securityListManager securityListManager

	service *v1.Service
	nodes   []*v1.Node
}

// NewLBSpec creates a LB Spec from a Kubernetes service and a slice of nodes.
func NewLBSpec(svc *v1.Service, nodes []*v1.Node, defaultSubnets []string, listenerSSLConfig *SSLConfig, backendSetSSLConfig *SSLConfig, secListFactory securityListManagerFactory) (*LBSpec, error) {
	if len(defaultSubnets) != 2 {
		return nil, errors.New("default subnets incorrectly configured")
	}

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

	// NOTE: These will be overridden for existing load balancers as load
	// balancer subnets cannot be modified.
	subnets := defaultSubnets
	if s, ok := svc.Annotations[ServiceAnnotationLoadBalancerSubnet1]; ok {
		subnets[0] = s
	}
	if s, ok := svc.Annotations[ServiceAnnotationLoadBalancerSubnet2]; ok {
		subnets[1] = s
	}

	if internal {
		// Only public load balancers need two subnets.  Internal load
		// balancers will always use the first subnet.
		if subnets[0] == "" {
			return nil, errors.Errorf("a configuration for subnet1 must be specified for an internal load balancer")
		}
		subnets = subnets[:1]
	} else {
		if subnets[0] == "" || subnets[1] == "" {
			return nil, errors.Errorf("a configuration for both subnets must be specified")
		}
	}

	listeners, err := getListeners(svc, listenerSSLConfig)
	if err != nil {
		return nil, err
	}

	return &LBSpec{
		Name:        GetLoadBalancerName(svc),
		Shape:       shape,
		Internal:    internal,
		Subnets:     subnets,
		Listeners:   listeners,
		BackendSets: getBackendSets(svc, nodes, backendSetSSLConfig),

		Ports:               getPorts(svc),
		ListenerSSLConfig:   listenerSSLConfig,
		BackendSetSSLConfig: backendSetSSLConfig,
		SourceCIDRs:         sourceCIDRs,

		service: svc,
		nodes:   nodes,
		securityListManager: secListFactory(
			svc.Annotations[ServiceAnnotaionLoadBalancerSecurityListManagementMode]),
	}, nil
}

// Certificates builds a map of required SSL certificates.
func buildCertificates(sslConfig *SSLConfig, svc *v1.Service, certs map[string]loadbalancer.CertificateDetails) error {
	if sslConfig == nil {
		return nil
	}
	sslSecret, err := sslConfig.readSSLSecret(sslConfig.Type, svc)
	if err != nil {
		return errors.Wrap(err, "reading SSL Secret")
	}
	if sslSecret.PublicCert == "" || sslSecret.PrivateKey == "" {
		return nil
	}

	certs[sslConfig.Name] = loadbalancer.CertificateDetails{
		CertificateName:   &sslConfig.Name,
		PublicCertificate: &sslSecret.PublicCert,
		CaCertificate:     &sslSecret.CACert,
		PrivateKey:        &sslSecret.PrivateKey,
		Passphrase:        &sslSecret.Passphrase,
	}
	return nil
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

func getPorts(svc *v1.Service) map[string]portSpec {
	ports := make(map[string]portSpec)
	for _, servicePort := range svc.Spec.Ports {
		name := getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
		ports[name] = portSpec{
			BackendPort:       int(servicePort.NodePort),
			ListenerPort:      int(servicePort.Port),
			HealthCheckerPort: *getHealthChecker(nil, int(servicePort.Port), svc).Port,
		}

	}
	return ports
}

func getBackends(nodes []*v1.Node, nodePort int32) []loadbalancer.BackendDetails {
	backends := make([]loadbalancer.BackendDetails, len(nodes))
	for i, node := range nodes {
		backends[i] = loadbalancer.BackendDetails{
			IpAddress: common.String(util.NodeInternalIP(node)),
			Port:      common.Int(int(nodePort)),
			Weight:    common.Int(1),
		}
	}
	return backends
}

func getBackendSets(svc *v1.Service, nodes []*v1.Node, sslCfg *SSLConfig) map[string]loadbalancer.BackendSetDetails {
	backendSets := make(map[string]loadbalancer.BackendSetDetails)
	for _, servicePort := range svc.Spec.Ports {
		name := getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
		port := int(servicePort.Port)
		sslConfig := getSSLConfiguration(sslCfg, port)
		if sslConfig != nil {
			backendSets[name] = loadbalancer.BackendSetDetails{
				Policy:           common.String(DefaultLoadBalancerPolicy),
				Backends:         getBackends(nodes, servicePort.NodePort),
				HealthChecker:    getHealthChecker(sslCfg, port, svc),
				SslConfiguration: sslConfig,
			}
		} else {
			backendSets[name] = loadbalancer.BackendSetDetails{
				Policy:        common.String(DefaultLoadBalancerPolicy),
				Backends:      getBackends(nodes, servicePort.NodePort),
				HealthChecker: getHealthChecker(sslCfg, port, svc),
			}
		}
	}
	return backendSets
}

func getHealthChecker(cfg *SSLConfig, port int, svc *v1.Service) *loadbalancer.HealthCheckerDetails {
	// If the health-check has SSL enabled use TCP rather than HTTP.
	protocol := lbNodesHealthCheckProtoHTTP
	if cfg != nil && cfg.Ports.Has(port) {
		protocol = lbNodesHealthCheckProtoTCP
	}
	checkPath, checkPort := apiservice.GetServiceHealthCheckPathPort(svc)
	if checkPath != "" {
		return &loadbalancer.HealthCheckerDetails{
			Protocol: &protocol,
			UrlPath:  &checkPath,
			Port:     common.Int(int(checkPort)),
		}
	}

	return &loadbalancer.HealthCheckerDetails{
		Protocol: &protocol,
		UrlPath:  common.String(lbNodesHealthCheckPath),
		Port:     common.Int(lbNodesHealthCheckPort),
	}
}

func getSSLConfiguration(cfg *SSLConfig, port int) *loadbalancer.SslConfigurationDetails {
	if cfg == nil || !cfg.Ports.Has(port) {
		return nil
	}
	return &loadbalancer.SslConfigurationDetails{
		CertificateName:       &cfg.Name,
		VerifyDepth:           common.Int(0),
		VerifyPeerCertificate: common.Bool(false),
	}
}

func getListeners(svc *v1.Service, sslCfg *SSLConfig) (map[string]loadbalancer.ListenerDetails, error) {
	// Determine if connection idle timeout has been specified
	var connectionIdleTimeout int
	connectionIdleTimeoutAnnotation := svc.Annotations[ServiceAnnotationLoadBalancerConnectionIdleTimeout]
	if connectionIdleTimeoutAnnotation != "" {
		timeout, err := strconv.ParseInt(connectionIdleTimeoutAnnotation, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing service annotation: %s=%s",
				ServiceAnnotationLoadBalancerConnectionIdleTimeout,
				connectionIdleTimeoutAnnotation,
			)
		}

		connectionIdleTimeout = int(timeout)
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
		sslConfiguration := getSSLConfiguration(sslCfg, port)
		name := getListenerName(protocol, port, sslConfiguration)

		listener := loadbalancer.ListenerDetails{
			DefaultBackendSetName: common.String(getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))),
			Protocol:              &protocol,
			Port:                  &port,
			SslConfiguration:      sslConfiguration,
		}

		if connectionIdleTimeout > 0 {
			listener.ConnectionConfiguration = &loadbalancer.ConnectionConfiguration{
				IdleTimeout: common.Int64(int64(connectionIdleTimeout)),
			}
		}

		listeners[name] = listener
	}

	return listeners, nil
}
