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

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/loadbalancer"
	"github.com/pkg/errors"

	api "k8s.io/api/core/v1"
	sets "k8s.io/apimachinery/pkg/util/sets"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/util"
)

type sslSecretReader interface {
	readSSLSecret(svc *api.Service) (cert, key string, err error)
}

type noopSSLSecretReader struct{}

func (ssr noopSSLSecretReader) readSSLSecret(svc *api.Service) (cert, key string, err error) {
	return "", "", nil
}

// SSLConfig is a description of a SSL certificate.
type SSLConfig struct {
	Name  string
	Ports sets.Int

	sslSecretReader
}

func needsCerts(svc *api.Service) bool {
	_, ok := svc.Annotations[ServiceAnnotationLoadBalancerSSLPorts]
	return ok
}

// NewSSLConfig constructs a new SSLConfig.
func NewSSLConfig(name string, ports []int, ssr sslSecretReader) *SSLConfig {
	if ssr == nil {
		ssr = noopSSLSecretReader{}
	}
	return &SSLConfig{
		Name:            name,
		Ports:           sets.NewInt(ports...),
		sslSecretReader: ssr,
	}
}

// LBSpec holds the data required to build a OCI load balancer from a
// kubernetes service.
type LBSpec struct {
	Name     string
	Shape    string
	Nodes    []*api.Node
	Subnets  []string
	Internal bool

	SSLConfig *SSLConfig
	service   *api.Service
}

// NewLBSpec creates a LB Spec from a Kubernetes service and a slice of nodes.
func NewLBSpec(service *api.Service, nodes []*api.Node, defaultSubnets []string, sslCfg *SSLConfig) (LBSpec, error) {
	if len(defaultSubnets) != 2 {
		return LBSpec{}, errors.New("defualt subnets incorrectly configured")
	}

	if err := validateService(service); err != nil {
		return LBSpec{}, err
	}

	_, internal := service.Annotations[ServiceAnnotationLoadBalancerInternal]

	// TODO (apryde): We should detect when this changes and WARN as we don't
	// support updating a load balancer's Shape.
	shape := lbDefaultShape
	if s, ok := service.Annotations[ServiceAnnotationLoadBalancerShape]; ok {
		shape = s
	}

	// NOTE: These will be overridden for existing load balancers as load
	// balancer subnets cannot be modified.
	subnets := defaultSubnets
	if s, ok := service.Annotations[ServiceAnnotationLoadBalancerSubnet1]; ok {
		subnets[0] = s
	}
	if s, ok := service.Annotations[ServiceAnnotationLoadBalancerSubnet2]; ok {
		subnets[1] = s
	}
	if internal {
		// Only public load balancers need two subnets.  Internal load
		// balancers will always use the first subnet.
		subnets = subnets[:1]
	}

	return LBSpec{
		Name:     GetLoadBalancerName(service),
		Shape:    shape,
		Nodes:    nodes,
		Internal: internal,
		Subnets:  subnets,

		service:   service,
		SSLConfig: sslCfg,
	}, nil
}

// TODO(apryde): aggragate errors using an error list.
func validateService(svc *api.Service) error {
	if err := validateProtocols(svc.Spec.Ports); err != nil {
		return err
	}

	if svc.Spec.SessionAffinity != api.ServiceAffinityNone {
		return errors.New("OCI only supports SessionAffinity `None` currently")
	}

	if svc.Spec.LoadBalancerIP != "" {
		// TODO(horwitz): We need to figure out in the WG if this should actually log or error.
		// The docs say: If the loadBalancerIP is specified, but the cloud provider does not support the feature, the field will be ignored.
		// But no one does that...
		// https://kubernetes.io/docs/concepts/services-networking/service/#type-loadbalancer
		return errors.New("OCI does not support setting the LoadBalancerIP")
	}

	return nil
}

func getBackendSetName(protocol string, port int) string {
	return fmt.Sprintf("%s-%d", protocol, port)
}

func (s *LBSpec) getBackends(nodePort int32) []loadbalancer.BackendDetails {
	backends := make([]loadbalancer.BackendDetails, len(s.Nodes))
	for i, node := range s.Nodes {
		backends[i] = loadbalancer.BackendDetails{
			IpAddress: common.String(util.NodeInternalIP(node)),
			Port:      common.Int(int(nodePort)),
			Weight:    common.Int(1),
		}
	}
	return backends
}

// GetBackendSets builds a map of BackendSets based on the LBSpec.
func (s *LBSpec) GetBackendSets() map[string]loadbalancer.BackendSetDetails {
	backendSets := make(map[string]loadbalancer.BackendSetDetails)
	for _, servicePort := range s.service.Spec.Ports {
		name := getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
		backendSets[name] = loadbalancer.BackendSetDetails{
			Policy:        common.String(DefaultLoadBalancerPolicy),
			Backends:      s.getBackends(servicePort.NodePort),
			HealthChecker: s.getHealthChecker(),
		}

	}
	return backendSets
}

func (s *LBSpec) getHealthChecker() *loadbalancer.HealthCheckerDetails {
	path, port := apiservice.GetServiceHealthCheckPathPort(s.service)
	if path != "" {
		return &loadbalancer.HealthCheckerDetails{
			Protocol: common.String(lbNodesHealthCheckProto),
			UrlPath:  &path,
			Port:     common.Int(int(port)),
		}
	}

	return &loadbalancer.HealthCheckerDetails{
		Protocol: common.String(lbNodesHealthCheckProto),
		UrlPath:  common.String(lbNodesHealthCheckPath),
		Port:     common.Int(lbNodesHealthCheckPort),
	}
}

func (s *LBSpec) getSSLConfiguration(port int) *loadbalancer.SslConfigurationDetails {
	if s.SSLConfig == nil || !s.SSLConfig.Ports.Has(port) {
		return nil
	}
	return &loadbalancer.SslConfigurationDetails{
		CertificateName:       &s.SSLConfig.Name,
		VerifyDepth:           common.Int(0),
		VerifyPeerCertificate: common.Bool(false),
	}
}

// GetCertificates builds a map of required SSL certificates.
func (s *LBSpec) GetCertificates() (map[string]loadbalancer.CertificateDetails, error) {
	certs := make(map[string]loadbalancer.CertificateDetails)
	if s.SSLConfig == nil {
		return certs, nil
	}

	cert, key, err := s.SSLConfig.readSSLSecret(s.service)
	if err != nil {
		return nil, errors.Wrap(err, "reading SSL Secret")
	}
	if cert == "" || key == "" {
		return certs, nil
	}

	certs[s.SSLConfig.Name] = loadbalancer.CertificateDetails{
		CertificateName:   &s.SSLConfig.Name,
		PublicCertificate: &cert,
		PrivateKey:        &key,
	}

	return certs, nil
}

// GetListeners builds a map of listeners based on the LBSpec.
func (s *LBSpec) GetListeners() map[string]loadbalancer.ListenerDetails {
	listeners := make(map[string]loadbalancer.ListenerDetails)
	for _, servicePort := range s.service.Spec.Ports {
		protocol := string(servicePort.Protocol)
		port := int(servicePort.Port)
		sslConfiguration := s.getSSLConfiguration(port)
		name := getListenerName(protocol, port, sslConfiguration)
		listeners[name] = loadbalancer.ListenerDetails{
			DefaultBackendSetName: common.String(getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))),
			Protocol:              &protocol,
			Port:                  &port,
			SslConfiguration:      sslConfiguration,
		}
	}
	return listeners
}
