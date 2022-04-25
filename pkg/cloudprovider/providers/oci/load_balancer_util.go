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
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"

	"go.uber.org/zap"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v50/loadbalancer"
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

const (
	changeFmtStr        = "%v -> Actual:%v - Desired:%v"
	backendChangeFmtStr = "%v -> Backend:%v"
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

	BackendSet client.GenericBackendSetDetails

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

	Listener client.GenericListener

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

func toInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func getHealthCheckerChanges(actual *client.GenericHealthChecker, desired *client.GenericHealthChecker) []string {

	var healthCheckerChanges []string
	// We would let LBCS to set the default HealthChecker if desired is nil
	if desired == nil {
		return healthCheckerChanges
	}

	//desired is not nil and actual is nil. So we need to reconcile
	if actual == nil {
		healthCheckerChanges = append(healthCheckerChanges, fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker", "NOT_PRESENT", "PRESENT"))
		return healthCheckerChanges
	}

	if toInt(actual.Port) != toInt(desired.Port) {
		healthCheckerChanges = append(healthCheckerChanges, fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:Port", toInt(actual.Port), toInt(desired.Port)))
	}
	//If there is no value for ResponseBodyRegex and ReturnCode in the LBSpec,
	//We would let the LBCS to set the default value. There is no point of reconciling.
	if toString(desired.ResponseBodyRegex) != "" && toString(actual.ResponseBodyRegex) != toString(desired.ResponseBodyRegex) {
		healthCheckerChanges = append(healthCheckerChanges, fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:ResponseBodyRegex", toString(actual.ResponseBodyRegex), toString(desired.ResponseBodyRegex)))
	}

	if toInt(actual.Retries) != toInt(desired.Retries) {
		healthCheckerChanges = append(healthCheckerChanges, fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:Retries", toInt(actual.Retries), toInt(desired.Retries)))
	}

	if toInt(desired.ReturnCode) != 0 && toInt(actual.ReturnCode) != toInt(desired.ReturnCode) {
		healthCheckerChanges = append(healthCheckerChanges, fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:ReturnCode", toInt(actual.ReturnCode), toInt(desired.ReturnCode)))
	}

	if toInt(actual.TimeoutInMillis) != toInt(desired.TimeoutInMillis) {
		healthCheckerChanges = append(healthCheckerChanges, fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:TimeoutInMillis", toInt(actual.TimeoutInMillis), toInt(desired.TimeoutInMillis)))
	}

	if toInt(actual.IntervalInMillis) != toInt(desired.IntervalInMillis) {
		healthCheckerChanges = append(healthCheckerChanges, fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:IntervalInMillis", toInt(actual.IntervalInMillis), toInt(desired.IntervalInMillis)))
	}

	if toString(actual.UrlPath) != toString(desired.UrlPath) {
		healthCheckerChanges = append(healthCheckerChanges, fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:UrlPath", toString(actual.UrlPath), toString(desired.UrlPath)))
	}

	if toString(&actual.Protocol) != toString(&desired.Protocol) {
		healthCheckerChanges = append(healthCheckerChanges, fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:Protocol", toString(&actual.Protocol), toString(&desired.Protocol)))
	}

	return healthCheckerChanges
}

// TODO(horwitz): this doesn't check weight which we may want in the future to
// evenly distribute Local traffic policy load.
func hasBackendSetChanged(logger *zap.SugaredLogger, actual client.GenericBackendSetDetails, desired client.GenericBackendSetDetails) bool {
	logger = logger.With("BackEndSetName", toString(actual.Name))
	backendSetChanges := getHealthCheckerChanges(actual.HealthChecker, desired.HealthChecker)
	// Need to update the seclist if service nodeport has changed
	if len(actual.Backends) > 0 && len(desired.Backends) > 0 {
		if *actual.Backends[0].Port != *desired.Backends[0].Port {
			backendSetChanges = append(backendSetChanges,
				fmt.Sprintf(changeFmtStr, "BackEndSet:BackendPort",
					*actual.Backends[0].Port, *desired.Backends[0].Port))
		}
	}

	if toString(actual.Policy) != toString(desired.Policy) {
		backendSetChanges = append(backendSetChanges, fmt.Sprintf(changeFmtStr, "BackEndSet:Policy", toString(actual.Policy), toString(desired.Policy)))
	}

	nameFormat := "%s:%d"

	desiredSet := sets.NewString()
	for _, backend := range desired.Backends {
		name := fmt.Sprintf(nameFormat, *backend.IpAddress, *backend.Port)
		desiredSet.Insert(name)
	}

	actualSet := sets.NewString()
	var backendChanges []string
	for _, backend := range actual.Backends {
		name := fmt.Sprintf(nameFormat, *backend.IpAddress, *backend.Port)
		if !desiredSet.Has(name) {
			backendChanges = append(backendChanges, fmt.Sprintf(backendChangeFmtStr, "BackEndSet:Backend Remove", name))
		}
		actualSet.Insert(name)
	}

	for _, backend := range desired.Backends {
		name := fmt.Sprintf(nameFormat, *backend.IpAddress, *backend.Port)
		if !actualSet.Has(name) {
			backendChanges = append(backendChanges, fmt.Sprintf(backendChangeFmtStr, "BackEndSet:Backend Add", name))
		}
	}

	if len(backendChanges) != 0 {
		backendSetChanges = append(backendChanges)
	}

	if len(backendSetChanges) != 0 {
		logger.Infof("BackendSet needs to be updated for the change(s) - %s", strings.Join(backendSetChanges, ","))
		return true
	}
	return false
}

func healthCheckerToDetails(hc *client.GenericHealthChecker) *client.GenericHealthChecker {
	if hc == nil {
		return nil
	}
	return &client.GenericHealthChecker{
		Protocol:         hc.Protocol,
		IntervalInMillis: hc.IntervalInMillis,
		Port:             hc.Port,
		//ResponseBodyRegex: hc.ResponseBodyRegex,
		Retries:         hc.Retries,
		ReturnCode:      hc.ReturnCode,
		TimeoutInMillis: hc.TimeoutInMillis,
		UrlPath:         hc.UrlPath,
	}
}

func sslConfigurationToDetails(sc *client.GenericSslConfigurationDetails) *client.GenericSslConfigurationDetails {
	if sc == nil {
		return nil
	}
	return &client.GenericSslConfigurationDetails{
		CertificateName:       sc.CertificateName,
		VerifyDepth:           sc.VerifyDepth,
		VerifyPeerCertificate: sc.VerifyPeerCertificate,
	}
}

func backendsToBackendDetails(bs []client.GenericBackend) []client.GenericBackend {
	backends := make([]client.GenericBackend, len(bs))
	for i, backend := range bs {
		backends[i] = client.GenericBackend{
			IpAddress: backend.IpAddress,
			Port:      backend.Port,
			//Backup:    backend.Backup,
			//Drain:     backend.Drain,
			//Offline:   backend.Offline,
			Weight: backend.Weight,
		}
	}
	return backends
}

func portsFromBackendSetDetails(logger *zap.SugaredLogger, name string, bs *client.GenericBackendSetDetails) portSpec {
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

func portsFromBackendSet(logger *zap.SugaredLogger, name string, bs *client.GenericBackendSetDetails) portSpec {
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

func getBackendSetChanges(logger *zap.SugaredLogger, actual map[string]client.GenericBackendSetDetails, desired map[string]client.GenericBackendSetDetails) []Action {
	var backendSetActions []Action
	// First check to see if any backendsets need to be deleted or updated.
	for name, actualBackendSet := range actual {
		desiredBackendSet, ok := desired[name]
		if !ok {
			// No longer exists
			backendSetActions = append(backendSetActions, &BackendSetAction{
				name: *actualBackendSet.Name,
				BackendSet: client.GenericBackendSetDetails{
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

		if hasBackendSetChanged(logger, actualBackendSet, desiredBackendSet) {
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

func getSSLConfigurationChanges(actual *client.GenericSslConfigurationDetails, desired *client.GenericSslConfigurationDetails) []string {
	var sslConfigurationChanges []string
	if actual == nil && desired == nil {
		return sslConfigurationChanges
	}
	if actual == nil && desired != nil {
		sslConfigurationChanges = append(sslConfigurationChanges, fmt.Sprintf(changeFmtStr, "Listener:SSLConfiguration", "NOT_PRESENT", "PRESENT"))
		return sslConfigurationChanges
	}
	if actual != nil && desired == nil {
		sslConfigurationChanges = append(sslConfigurationChanges, fmt.Sprintf(changeFmtStr, "Listener:SSLConfiguration", "PRESENT", "NOT_PRESENT"))
		return sslConfigurationChanges
	}

	if toString(actual.CertificateName) != toString(desired.CertificateName) {
		sslConfigurationChanges = append(sslConfigurationChanges, fmt.Sprintf(changeFmtStr, "Listener:SSLConfiguration:CertificateName", toString(actual.CertificateName), toString(desired.CertificateName)))
	}
	if toInt(actual.VerifyDepth) != toInt(desired.VerifyDepth) {
		sslConfigurationChanges = append(sslConfigurationChanges, fmt.Sprintf(changeFmtStr, "Listener:SSLConfiguration:VerifyDepth", toInt(actual.VerifyDepth), toInt(desired.VerifyDepth)))
	}
	if toBool(actual.VerifyPeerCertificate) != toBool(desired.VerifyPeerCertificate) {
		sslConfigurationChanges = append(sslConfigurationChanges, fmt.Sprintf(changeFmtStr, "Listener:SSLConfiguration:VerifyPeerCertificate", toBool(actual.VerifyPeerCertificate), toBool(desired.VerifyPeerCertificate)))
	}
	return sslConfigurationChanges
}

func hasListenerChanged(logger *zap.SugaredLogger, actual client.GenericListener, desired client.GenericListener) bool {
	logger = logger.With("ListenerName", toString(actual.Name))
	var listenerChanges []string
	if toString(actual.DefaultBackendSetName) != toString(desired.DefaultBackendSetName) {
		listenerChanges = append(listenerChanges, fmt.Sprintf(changeFmtStr, "Listener:DefaultBackendSetName", toString(actual.DefaultBackendSetName), toString(desired.DefaultBackendSetName)))
	}
	if toInt(actual.Port) != toInt(desired.Port) {
		listenerChanges = append(listenerChanges, fmt.Sprintf(changeFmtStr, "Listener:Port", toInt(actual.Port), toInt(desired.Port)))
	}
	if toString(actual.Protocol) != toString(desired.Protocol) {
		listenerChanges = append(listenerChanges, fmt.Sprintf(changeFmtStr, "Listener:Protocol", toString(actual.Protocol), toString(desired.Protocol)))
	}
	listenerChanges = append(listenerChanges, getSSLConfigurationChanges(actual.SslConfiguration, desired.SslConfiguration)...)
	listenerChanges = append(listenerChanges, getConnectionConfigurationChanges(actual.ConnectionConfiguration, desired.ConnectionConfiguration)...)

	if len(listenerChanges) != 0 {
		logger.Infof("Listener needs to be updated for the change(s) - %s", strings.Join(listenerChanges, ","))
		return true
	}
	return false
}

func getConnectionConfigurationChanges(actual *client.GenericConnectionConfiguration, desired *client.GenericConnectionConfiguration) []string {
	var connectionConfigurationChanges []string
	// We would let LBCS to set the default IdleTimeout if desired is nil
	if desired == nil {
		return connectionConfigurationChanges
	}

	//desired is not nil and actual is nil. So we need to reconcile
	if actual == nil {
		connectionConfigurationChanges = append(connectionConfigurationChanges, fmt.Sprintf(changeFmtStr, "Listener:ConnectionConfiguration", "NOT_PRESENT", "PRESENT"))
		return connectionConfigurationChanges
	}

	if toInt64(actual.IdleTimeout) != toInt64(desired.IdleTimeout) {
		connectionConfigurationChanges = append(connectionConfigurationChanges, fmt.Sprintf(changeFmtStr, "Listener:ConnectionConfiguration:IdleTimeout", toInt64(actual.IdleTimeout), toInt64(desired.IdleTimeout)))
	}

	if toInt(actual.BackendTcpProxyProtocolVersion) != toInt(desired.BackendTcpProxyProtocolVersion) {
		connectionConfigurationChanges = append(connectionConfigurationChanges, fmt.Sprintf(changeFmtStr, "Listener:ConnectionConfiguration:BackendTcpProxyProtocolVersion", toInt(actual.BackendTcpProxyProtocolVersion), toInt(desired.BackendTcpProxyProtocolVersion)))
	}

	return connectionConfigurationChanges
}

func getListenerChanges(logger *zap.SugaredLogger, actual map[string]client.GenericListener, desired map[string]client.GenericListener) []Action {
	var listenerActions []Action

	// set to keep track of desired listeners that already exist and should not be created
	exists := sets.NewString()
	//sanitizedDesiredListeners convert the listener name HTTP-xxxx to TCP-xxx such that in sortAndCombineAction can
	//place BackendSet create before Listener Create and Listener delete before BackendSet delete. Also it would help
	//not to delete and create Listener if customer edit the service and add oci-load-balancer-backend-protocol: "HTTP"
	// and vice versa. It would help to only update the listener in case of protocol change. Refer OKE-10793 for details.
	sanitizedDesiredListeners := make(map[string]client.GenericListener)
	for name, desiredListener := range desired {
		sanitizedDesiredListeners[getSanitizedName(name)] = desiredListener
	}
	// First check to see if any listeners need to be deleted or updated.
	for name, actualListener := range actual {
		desiredListener, ok := sanitizedDesiredListeners[getSanitizedName(name)]
		if !ok {
			// no longer exists
			listenerActions = append(listenerActions, &ListenerAction{
				Listener: client.GenericListener{
					DefaultBackendSetName: actualListener.DefaultBackendSetName,
					Port:                  actualListener.Port,
					Protocol:              actualListener.Protocol,
					SslConfiguration:      sslConfigurationToDetails(actualListener.SslConfiguration),
				},
				name:       name,
				actionType: Delete,
			})
			continue
		}
		exists.Insert(getSanitizedName(name))
		if hasListenerChanged(logger, actualListener, desiredListener) {
			listenerActions = append(listenerActions, &ListenerAction{
				Listener:   desiredListener,
				name:       name,
				actionType: Update,
			})
		}
	}

	// Now check if any need to be created.
	for name, desiredListener := range desired {
		if !exists.Has(getSanitizedName(name)) {
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

func hasLoadbalancerShapeChanged(ctx context.Context, spec *LBSpec, lb *client.GenericLoadBalancer) bool {
	if *lb.ShapeName != spec.Shape {
		return true
	}

	// in case of fixed shape with no shape change (lb.ShapeDetails == nil for fixedshape),
	// or flexshape with missing property values return false
	if lb.ShapeDetails == nil || spec.FlexMin == nil || spec.FlexMax == nil {
		return false
	}

	// in case of flex shape with change of min/max bandwitch
	if *lb.ShapeDetails.MinimumBandwidthInMbps != *spec.FlexMin {
		return true
	}
	if *lb.ShapeDetails.MaximumBandwidthInMbps != *spec.FlexMax {
		return true
	}
	return false
}

// hasLoadBalancerNetworkSecurityGroupsChanged checks for the difference in actual NSGs
// associated to LoadBalancer with Desired NSGs provided in service annotation
func hasLoadBalancerNetworkSecurityGroupsChanged(ctx context.Context, actualNetworkSecurityGroup, desiredNetworkSecurityGroup []string) bool {
	if len(desiredNetworkSecurityGroup) != len(actualNetworkSecurityGroup) {
		return true
	}

	return !DeepEqualLists(actualNetworkSecurityGroup, desiredNetworkSecurityGroup)
}

func sslEnabled(sslConfigMap map[int]*loadbalancer.SslConfiguration) bool {
	return len(sslConfigMap) > 0
}

// getSanitizedName omits the suffix after protocol-port in the name.
// It also converts the listener name from HTTP-xxxx to TCP-xxx such that
// we can use oci-load-balancer-backend-protocol: "HTTP" annotation.
func getSanitizedName(name string) string {
	fields := strings.Split(name, "-")
	if strings.EqualFold(fields[0], "HTTP") {
		fields[0] = "TCP"
		name = fmt.Sprintf(strings.Join(fields, "-"))
	}
	if len(fields) > 2 {
		return fmt.Sprintf(strings.Join(fields[:2], "-"))
	}
	return name
}

func getListenerName(protocol string, port int) string {
	return fmt.Sprintf("%s-%d", protocol, port)
}

// GetLoadBalancerName gets the name of the load balancer based on the service
func GetLoadBalancerName(service *api.Service) string {
	lbType := getLoadBalancerType(service)
	var name string
	switch lbType {
	case NLB:
		{
			name = fmt.Sprintf("%s/%s/%s", service.Namespace, service.Name, service.UID)
		}
	default:
		{
			prefix := os.Getenv(lbNamePrefixEnvVar)
			if prefix != "" && !strings.HasSuffix(prefix, "-") {
				// Add the trailing hyphen if it's missing
				prefix += "-"
			}
			name = fmt.Sprintf("%s%s", prefix, service.UID)
		}
	}
	if len(name) > 1024 {
		// 1024 is the max length for display name
		// https://docs.oracle.com/en-us/iaas/api/#/en/networkloadbalancer/20200501/datatypes/UpdateNetworkLoadBalancerDetails
		// https://docs.us-phoenix-1.oraclecloud.com/api/#/en/loadbalancer/20170115/requests/UpdateLoadBalancerDetails
		name = name[:1024]
	}
	return name
}

// validateProtocols validates that OCI supports the protocol of all
// ServicePorts defined by a service.
func validateProtocols(servicePorts []api.ServicePort, lbType string, secListMgmtMode string) error {
	for _, servicePort := range servicePorts {
		if servicePort.Protocol == api.ProtocolUDP && lbType == LB {
			return fmt.Errorf("OCI load balancers do not support UDP")
		}
		if servicePort.Protocol == api.ProtocolUDP && lbType == NLB && secListMgmtMode != ManagementModeNone {
			return fmt.Errorf("Security list management mode can only be 'None' for UDP protocol")
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
	sort.SliceStable(actions, func(i, j int) bool {
		a1 := actions[i]
		a2 := actions[j]

		// Sort by the name until we get to the point a1 and a2 are Actions upon
		// an associated Listener and BackendSet (which share the same name).
		if getSanitizedName(a1.Name()) != getSanitizedName(a2.Name()) {
			return getSanitizedName(a1.Name()) < getSanitizedName(a2.Name())
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

func getMetric(lbtype string, metricType string) string {
	if lbtype == LB {
		switch metricType {
		case Create:
			return metrics.LBProvision
		case Update:
			return metrics.LBUpdate
		case Delete:
			return metrics.LBDelete
		}
	}
	if lbtype == NLB {
		switch metricType {
		case Create:
			return metrics.NLBProvision
		case Update:
			return metrics.NLBUpdate
		case Delete:
			return metrics.NLBDelete
		}
	}
	return ""
}
