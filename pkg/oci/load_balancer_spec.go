// Copyright 2017 The OCI Cloud Controller Manager Authors
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
	"errors"
	"fmt"

	"github.com/golang/glog"

	baremetal "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/util"
	api "k8s.io/api/core/v1"
	apiservice "k8s.io/kubernetes/pkg/api/v1/service"
)

// LBSpec holds the data required to build a OCI load balancer from a
// kubernetes service.
type LBSpec struct {
	Name    string
	Shape   string
	Service *api.Service
	Nodes   []*api.Node
	Subnets []string
}

// NewLBSpec creates a LB Spec from a kubernetes service and a slice of nodes.
func NewLBSpec(cp *CloudProvider, service *api.Service, nodes []*api.Node) (LBSpec, error) {
	if err := validateProtocols(service.Spec.Ports); err != nil {
		return LBSpec{}, err
	}

	if service.Spec.SessionAffinity != api.ServiceAffinityNone {
		return LBSpec{}, errors.New("OCI only supports SessionAffinity `None` currently")
	}

	if service.Spec.LoadBalancerIP != "" {
		return LBSpec{}, errors.New("OCI does not support setting the LoadBalancerIP")
	}

	internalLB := false
	internalAnnotation := service.Annotations[ServiceAnnotationLoadBalancerInternal]
	if internalAnnotation != "" {
		internalLB = true
	}

	if internalLB {
		return LBSpec{}, errors.New("OCI does not currently support internal load balancers")
	}

	// TODO (apryde): We should detect when this changes and WARN as we don't
	// support updating a load balancer's Shape.
	lbShape := service.Annotations[ServiceAnnotationLoadBalancerShape]
	if lbShape == "" {
		lbShape = lbDefaultShape
	}

	// NOTE: These will be overridden for existing load balancers as load
	// balancer subnets cannot be modified.
	subnet1, ok := service.Annotations[ServiceAnnotationLoadBalancerSubnet1]
	if !ok {
		subnet1 = cp.config.LoadBalancer.Subnet1
	}
	subnet2, ok := service.Annotations[ServiceAnnotationLoadBalancerSubnet2]
	if !ok {
		subnet2 = cp.config.LoadBalancer.Subnet2
	}

	return LBSpec{
		Name:    GetLoadBalancerName(service),
		Shape:   lbShape,
		Service: service,
		Nodes:   nodes,
		Subnets: []string{subnet1, subnet2},
	}, nil
}

func getBackendSetName(protocol string, port int) string {
	return fmt.Sprintf("%s-%d", protocol, port)
}

// GetBackendSets builds a map of BackendSets based on the LBSpec.
func (s *LBSpec) GetBackendSets() map[string]baremetal.BackendSet {
	backendSets := make(map[string]baremetal.BackendSet)
	for _, servicePort := range s.Service.Spec.Ports {
		name := getBackendSetName(string(servicePort.Protocol), int(servicePort.Port))
		backendSet := baremetal.BackendSet{
			Name:          name,
			Policy:        client.DefaultLoadBalancerPolicy,
			Backends:      []baremetal.Backend{},
			HealthChecker: s.getHealthChecker(),
		}
		for _, node := range s.Nodes {
			backendSet.Backends = append(backendSet.Backends, baremetal.Backend{
				IPAddress: util.NodeInternalIP(node),
				Port:      int(servicePort.NodePort),
				Weight:    1,
			})
		}
		backendSets[name] = backendSet
	}
	return backendSets
}

func (s *LBSpec) getHealthChecker() *baremetal.HealthChecker {
	path, port := apiservice.GetServiceHealthCheckPathPort(s.Service)
	if path != "" {
		return &baremetal.HealthChecker{
			Protocol: lbNodesHealthCheckProto,
			URLPath:  path,
			Port:     int(port),
		}
	}

	return &baremetal.HealthChecker{
		Protocol: lbNodesHealthCheckProto,
		URLPath:  lbNodesHealthCheckPath,
		Port:     lbNodesHealthCheckPort,
	}
}

// GetSSLConfig builds a map of SSL configuration per listener port based on
// the LBSpec.
func (s *LBSpec) GetSSLConfig(certificateName string) (map[int]*baremetal.SSLConfiguration, error) {
	sslConfigMap := make(map[int]*baremetal.SSLConfiguration)

	sslEnabledPorts, err := getSSLEnabledPorts(s.Service.ObjectMeta.Annotations)
	if err != nil {
		return nil, err
	}

	if len(sslEnabledPorts) == 0 {
		glog.V(4).Infof("No SSL enabled ports found for service %q", s.Service.Name)
		return sslConfigMap, nil
	}

	for _, servicePort := range s.Service.Spec.Ports {
		port := int(servicePort.Port)
		if _, ok := sslEnabledPorts[port]; ok {
			sslConfigMap[port] = &baremetal.SSLConfiguration{
				CertificateName:       certificateName,
				VerifyDepth:           0,
				VerifyPeerCertificate: false,
			}
		}
	}
	return sslConfigMap, nil
}

// GetListeners builds a map of listeners based on the LBSpec.
func (s *LBSpec) GetListeners(sslConfigMap map[int]*baremetal.SSLConfiguration) map[string]baremetal.Listener {
	listeners := make(map[string]baremetal.Listener)
	for _, servicePort := range s.Service.Spec.Ports {
		protocol := string(servicePort.Protocol)
		port := int(servicePort.Port)
		sslConfig := sslConfigMap[port]
		name := getListenerName(protocol, port, sslConfig)
		listener := baremetal.Listener{
			Name: name,
			DefaultBackendSetName: getBackendSetName(string(servicePort.Protocol), int(servicePort.Port)),
			Protocol:              protocol,
			Port:                  port,
			SSLConfig:             sslConfig,
		}
		listeners[name] = listener
	}
	return listeners
}
