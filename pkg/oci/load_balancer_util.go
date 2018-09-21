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
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/oracle/oci-go-sdk/loadbalancer"
	"go.uber.org/zap"

	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// SSLCAFileName is a key name for ca data in the secrets config.
	SSLCAFileName = "ca.crt"
	// SSLCertificateFileName is a key name for certificate data in the secrets config.
	SSLCertificateFileName = "tls.crt"
	// SSLPrivateKeyFileName is a key name for cartificate private key in the secrets config.
	SSLPrivateKeyFileName = "tls.key"
	// SSLPassphrase is a key name for certificate passphrase in the secrets config.
	SSLPassphrase = "passphrase"
)

const lbNamePrefixEnvVar = "LOAD_BALANCER_PREFIX"

// ActionType specifies what action should be taken on the resource.
type ActionType string

const (
	// Create the resource as it doesn't exist yet.
	Create = "create"
	// Update the resource.
	Update = "update"
	// Delete the resource.
	Delete = "delete"
)

// Action that should take place on the resource.
type Action interface {
	Type() ActionType
	Name() string
}

// BackendSetAction denotes the action that should be taken on the given
// BackendSet.
type BackendSetAction struct {
	Action

	actionType ActionType
	name       string

	BackendSet loadbalancer.BackendSetDetails

	Ports    portSpec
	OldPorts *portSpec
}

// Type of the Action.
func (b *BackendSetAction) Type() ActionType {
	return b.actionType
}

// Name of the action's object.
func (b *BackendSetAction) Name() string {
	return b.name
}

func (b *BackendSetAction) String() string {
	return fmt.Sprintf("BackendSetAction:{Name: %s, Type: %v, Ports: %+v}", b.Name(), b.actionType, b.Ports)
}

// ListenerAction denotes the action that should be taken on the given Listener.
type ListenerAction struct {
	Action

	actionType ActionType
	name       string

	Listener loadbalancer.ListenerDetails

	Ports    portSpec
	OldPorts *portSpec
}

// Type of the Action.
func (l *ListenerAction) Type() ActionType {
	return l.actionType
}

// Name of the action's object.
func (l *ListenerAction) Name() string {
	return l.name
}

func (l *ListenerAction) String() string {
	return fmt.Sprintf("ListenerAction:{Name: %s, Type: %v }", l.Name(), l.actionType)
}

func toBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func toString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func toInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func hasHealthCheckerChanged(actual *loadbalancer.HealthChecker, desired *loadbalancer.HealthCheckerDetails) bool {
	if actual == nil {
		return desired != nil
	}

	if toInt(actual.Port) != toInt(desired.Port) {
		return true
	}

	if toString(actual.ResponseBodyRegex) != toString(desired.ResponseBodyRegex) {
		return true
	}

	if toInt(actual.Retries) != toInt(desired.Retries) {
		return true
	}

	if toInt(actual.ReturnCode) != toInt(desired.ReturnCode) {
		return true
	}

	if toInt(actual.TimeoutInMillis) != toInt(desired.TimeoutInMillis) {
		return true
	}

	if toString(actual.UrlPath) != toString(desired.UrlPath) {
		return true
	}

	return false
}

// TODO(horwitz): this doesn't check weight which we may want in the future to
// evenly distribute Local traffic policy load.
func hasBackendSetChanged(actual loadbalancer.BackendSet, desired loadbalancer.BackendSetDetails) bool {
	if hasHealthCheckerChanged(actual.HealthChecker, desired.HealthChecker) {
		return true
	}

	if toString(actual.Policy) != toString(desired.Policy) {
		return true
	}

	if len(actual.Backends) != len(desired.Backends) {
		return true
	}

	nameFormat := "%s:%d"

	// Since the lengths are equal that means the membership must be the same
	// else there has been change.
	desiredSet := sets.NewString()
	for _, backend := range desired.Backends {
		name := fmt.Sprintf(nameFormat, *backend.IpAddress, *backend.Port)
		desiredSet.Insert(name)
	}

	for _, backend := range actual.Backends {
		name := fmt.Sprintf(nameFormat, *backend.IpAddress, *backend.Port)
		if !desiredSet.Has(name) {
			return true
		}
	}

	return false
}

func healthCheckerToDetails(hc *loadbalancer.HealthChecker) *loadbalancer.HealthCheckerDetails {
	if hc == nil {
		return nil
	}
	return &loadbalancer.HealthCheckerDetails{
		Protocol:          hc.Protocol,
		IntervalInMillis:  hc.IntervalInMillis,
		Port:              hc.Port,
		ResponseBodyRegex: hc.ResponseBodyRegex,
		Retries:           hc.Retries,
		ReturnCode:        hc.ReturnCode,
		TimeoutInMillis:   hc.TimeoutInMillis,
		UrlPath:           hc.UrlPath,
	}
}

func sslConfigurationToDetails(sc *loadbalancer.SslConfiguration) *loadbalancer.SslConfigurationDetails {
	if sc == nil {
		return nil
	}
	return &loadbalancer.SslConfigurationDetails{
		CertificateName:       sc.CertificateName,
		VerifyDepth:           sc.VerifyDepth,
		VerifyPeerCertificate: sc.VerifyPeerCertificate,
	}
}

func backendsToBackendDetails(bs []loadbalancer.Backend) []loadbalancer.BackendDetails {
	backends := make([]loadbalancer.BackendDetails, len(bs))
	for i, backend := range bs {
		backends[i] = loadbalancer.BackendDetails{
			IpAddress: backend.IpAddress,
			Port:      backend.Port,
			Backup:    backend.Backup,
			Drain:     backend.Drain,
			Offline:   backend.Offline,
			Weight:    backend.Weight,
		}
	}
	return backends
}

func portsFromBackendSetDetails(logger *zap.SugaredLogger, name string, bs *loadbalancer.BackendSetDetails) portSpec {
	spec := portSpec{}
	if len(bs.Backends) > 0 {
		spec.BackendPort = *bs.Backends[0].Port
	} else {
		logger.Warnf("BackendSet %q has no Backends", name)
	}
	if bs.HealthChecker != nil {
		spec.HealthCheckerPort = *bs.HealthChecker.Port
	} else {
		logger.Warnf("BackendSet %q has no health checker", name)
	}
	return spec
}

func portsFromBackendSet(logger *zap.SugaredLogger, name string, bs *loadbalancer.BackendSet) portSpec {
	spec := portSpec{}
	if len(bs.Backends) > 0 {
		spec.BackendPort = *bs.Backends[0].Port
	} else {
		logger.Warnf("BackendSet %q has no Backends", name)
	}
	if bs.HealthChecker != nil {
		spec.HealthCheckerPort = *bs.HealthChecker.Port
	} else {
		logger.Warnf("BackendSet %q has no health checker", name)
	}
	return spec
}

func getBackendSetChanges(logger *zap.SugaredLogger, actual map[string]loadbalancer.BackendSet, desired map[string]loadbalancer.BackendSetDetails) []Action {
	var backendSetActions []Action
	// First check to see if any backendsets need to be deleted or updated.
	for name, actualBackendSet := range actual {
		desiredBackendSet, ok := desired[name]
		if !ok {
			// No longer exists
			backendSetActions = append(backendSetActions, &BackendSetAction{
				name: *actualBackendSet.Name,
				BackendSet: loadbalancer.BackendSetDetails{
					HealthChecker:                   healthCheckerToDetails(actualBackendSet.HealthChecker),
					Policy:                          actualBackendSet.Policy,
					Backends:                        backendsToBackendDetails(actualBackendSet.Backends),
					SessionPersistenceConfiguration: actualBackendSet.SessionPersistenceConfiguration,
					SslConfiguration:                sslConfigurationToDetails(actualBackendSet.SslConfiguration),
				},
				Ports:      portsFromBackendSet(logger, *actualBackendSet.Name, &actualBackendSet),
				actionType: Delete,
			})
			continue
		}

		if hasBackendSetChanged(actualBackendSet, desiredBackendSet) {
			oldPorts := portsFromBackendSet(logger, name, &actualBackendSet)
			backendSetActions = append(backendSetActions, &BackendSetAction{
				name:       name,
				BackendSet: desiredBackendSet,
				Ports:      portsFromBackendSetDetails(logger, name, &desiredBackendSet),
				OldPorts:   &oldPorts,
				actionType: Update,
			})
		}
	}

	// Now check if any need to be created.
	for name, desiredBackendSet := range desired {
		if _, ok := actual[name]; !ok {
			// Doesn't exist so lets create it.
			backendSetActions = append(backendSetActions, &BackendSetAction{
				name:       name,
				BackendSet: desiredBackendSet,
				Ports:      portsFromBackendSetDetails(logger, name, &desiredBackendSet),
				actionType: Create,
			})
		}
	}

	return backendSetActions
}

func hasSSLConfigurationChanged(actual *loadbalancer.SslConfiguration, desired *loadbalancer.SslConfigurationDetails) bool {
	if actual == nil || desired == nil {
		if actual == nil && desired == nil {
			return false
		}
		return true
	}

	if toString(actual.CertificateName) != toString(desired.CertificateName) {
		return true
	}
	if toInt(actual.VerifyDepth) != toInt(desired.VerifyDepth) {
		return true
	}
	if toBool(actual.VerifyPeerCertificate) != toBool(desired.VerifyPeerCertificate) {
		return true
	}
	return false
}

func hasListenerChanged(actual loadbalancer.Listener, desired loadbalancer.ListenerDetails) bool {
	if toString(actual.DefaultBackendSetName) != toString(desired.DefaultBackendSetName) {
		return true
	}
	if toInt(actual.Port) != toInt(desired.Port) {
		return true
	}
	if toString(actual.Protocol) != toString(desired.Protocol) {
		return true
	}
	if hasSSLConfigurationChanged(actual.SslConfiguration, desired.SslConfiguration) {
		return true
	}
	return false
}

func getListenerChanges(actual map[string]loadbalancer.Listener, desired map[string]loadbalancer.ListenerDetails) []Action {
	var listenerActions []Action
	// First check to see if any listeners need to be deleted or updated.
	for name, actualListener := range actual {
		desiredListener, ok := desired[name]
		if !ok {
			// no longer exists
			listenerActions = append(listenerActions, &ListenerAction{
				Listener: loadbalancer.ListenerDetails{
					DefaultBackendSetName: actualListener.DefaultBackendSetName,
					Port:             actualListener.Port,
					Protocol:         actualListener.Protocol,
					SslConfiguration: sslConfigurationToDetails(actualListener.SslConfiguration),
				},
				name:       name,
				actionType: Delete,
			})
			continue
		}

		if hasListenerChanged(actualListener, desiredListener) {
			listenerActions = append(listenerActions, &ListenerAction{
				Listener:   desiredListener,
				name:       name,
				actionType: Update,
			})
		}
	}

	// Now check if any need to be created.
	for name, desiredListener := range desired {
		if _, ok := actual[name]; !ok {
			// doesn't exist so lets create it
			listenerActions = append(listenerActions, &ListenerAction{
				Listener:   desiredListener,
				name:       name,
				actionType: Create,
			})
		}
	}

	return listenerActions
}

func sslEnabled(sslConfigMap map[int]*loadbalancer.SslConfiguration) bool {
	return len(sslConfigMap) > 0
}

func getListenerName(protocol string, port int, sslConfig *loadbalancer.SslConfigurationDetails) string {
	if sslConfig != nil {
		return fmt.Sprintf("%s-%d-%s", protocol, port, *sslConfig.CertificateName)
	}
	return fmt.Sprintf("%s-%d", protocol, port)
}

// GetLoadBalancerName gets the name of the load balancer based on the service
func GetLoadBalancerName(service *api.Service) string {
	prefix := os.Getenv(lbNamePrefixEnvVar)
	if prefix != "" && !strings.HasSuffix(prefix, "-") {
		// Add the trailing hyphen if it's missing
		prefix += "-"
	}

	name := fmt.Sprintf("%s%s", prefix, service.UID)
	if len(name) > 1024 {
		// 1024 is the max length for display name
		// https://docs.us-phoenix-1.oraclecloud.com/api/#/en/loadbalancer/20170115/requests/UpdateLoadBalancerDetails
		name = name[:1024]
	}

	return name
}

// validateProtocols validates that OCI supports the protocol of all
// ServicePorts defined by a service.
func validateProtocols(servicePorts []api.ServicePort) error {
	for _, servicePort := range servicePorts {
		if servicePort.Protocol == api.ProtocolUDP {
			return fmt.Errorf("OCI load balancers do not support UDP")
		}
	}
	return nil
}

// getSSLEnabledPorts returns a list of port numbers for which we need to enable
// SSL on the corresponding listener.
func getSSLEnabledPorts(svc *api.Service) ([]int, error) {
	ports := []int{}
	annotation, ok := svc.Annotations[ServiceAnnotationLoadBalancerSSLPorts]
	if !ok || annotation == "" {
		return ports, nil
	}

	for _, s := range strings.Split(annotation, ",") {
		port, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return nil, fmt.Errorf("parse SSL port: %v", err)
		}
		ports = append(ports, port)
	}
	return ports, nil
}

// parseSecretString returns the secret name and secret namespace from the
// given secret string (taken from the ssl annotation value).
func parseSecretString(secretString string) (string, string) {
	fields := strings.Split(secretString, "/")
	if len(fields) >= 2 {
		return fields[0], fields[1]
	}
	return "", secretString
}

// sortAndCombineActions combines two slices of Actions and then sorts them to
// ensure that BackendSets are created prior to their associated Listeners but
// deleted after their associated Listeners.
func sortAndCombineActions(logger *zap.SugaredLogger, backendSetActions []Action, listenerActions []Action) []Action {
	actions := append(backendSetActions, listenerActions...)
	sort.Slice(actions, func(i, j int) bool {
		a1 := actions[i]
		a2 := actions[j]

		// Sort by the name until we get to the point a1 and a2 are Actions upon
		// an associated Listener and BackendSet (which share the same name).
		if a1.Name() != a2.Name() {
			return a1.Name() < a2.Name()
		}

		// For Create and Delete (which is what we really care about) the
		// ActionType will always be the same so we can get away with just
		// checking the type of the first action.
		switch a1.Type() {
		case Create:
			// Create the BackendSet then Listener.
			_, ok := a1.(*BackendSetAction)
			return ok
		case Update:
			// Doesn't matter.
			return true
		case Delete:
			// Delete the Listener then BackendSet.
			_, ok := a2.(*BackendSetAction)
			return ok
		default:
			// Should never be reachable.
			logger.Errorf("Unknown action type received: %+v", a1)
			return true
		}
	})
	return actions
}
