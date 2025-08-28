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
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"go.uber.org/zap"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-go-sdk/v65/loadbalancer"
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
	// List the resource
	List = "list"
	// Get the resource
	Get = "get"
)

const nonAlphanumericRegexExpression = "[^a-zA-Z0-9]+"

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

type RuleSetAction struct {
	Action

	actionType ActionType
	name       string

	RuleSetDetails loadbalancer.RuleSetDetails
}

// Type of the Action.
func (b *RuleSetAction) Type() ActionType {
	return b.actionType
}

// Name of the action's object.
func (b *RuleSetAction) Name() string {
	return b.name
}

func (b *RuleSetAction) String() string {
	return fmt.Sprintf("RuleSetAction:{Name: %s, Type: %v, Rules: %+v}", b.Name(), b.actionType, b.RuleSetDetails)
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

// contains is a utility method to check if a string is part of a slice
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// removeAtPosition is a helper method to remove an element from remove and return it given the index
func removeAtPosition(slice []string, position int) []string {
	slice[position] = slice[len(slice)-1]
	return slice[:len(slice)-1]
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

	if toBool(actual.IsForcePlainText) != toBool(desired.IsForcePlainText) {
		healthCheckerChanges = append(healthCheckerChanges, fmt.Sprintf(changeFmtStr, "BackendSet:HealthChecker:IsForcePlainText", toBool(actual.IsForcePlainText), toBool(desired.IsForcePlainText)))
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

	if toBool(actual.IsPreserveSource) != toBool(desired.IsPreserveSource) {
		backendSetChanges = append(backendSetChanges, fmt.Sprintf(changeFmtStr, "BackEndSet:IsPreserveSource", toBool(actual.IsPreserveSource), toBool(desired.IsPreserveSource)))
	}

	backendSetChanges = append(backendSetChanges, getSSLConfigurationChanges(actual.SslConfiguration, desired.SslConfiguration)...)
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
		backendSetChanges = append(backendSetChanges, backendChanges...)
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
		IsForcePlainText: hc.IsForcePlainText,
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
		VerifyDepth:                    sc.VerifyDepth,
		VerifyPeerCertificate:          sc.VerifyPeerCertificate,
		HasSessionResumption:           sc.HasSessionResumption,
		TrustedCertificateAuthorityIds: sc.TrustedCertificateAuthorityIds,
		CertificateIds:                 sc.CertificateIds,
		CertificateName:                sc.CertificateName,
		Protocols:                      sc.Protocols,
		CipherSuiteName:                sc.CipherSuiteName,
		ServerOrderPreference:          sc.ServerOrderPreference,
	}
}

func backendsToBackendDetails(bs []client.GenericBackend) []client.GenericBackend {
	backends := make([]client.GenericBackend, len(bs))
	for i, backend := range bs {
		backends[i] = client.GenericBackend{
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
					IpVersion:                       actualBackendSet.IpVersion,
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

	if desired.CipherSuiteName != nil && len(*desired.CipherSuiteName) != 0 {
		if toString(actual.CipherSuiteName) != toString(desired.CipherSuiteName) {
			sslConfigurationChanges = append(sslConfigurationChanges, fmt.Sprintf(changeFmtStr, "Listener:SSLConfiguration:CipherSuiteName", toString(actual.CipherSuiteName), toString(desired.CipherSuiteName)))
		}
		if !reflect.DeepEqual(actual.Protocols, desired.Protocols) {
			sslConfigurationChanges = append(sslConfigurationChanges, fmt.Sprintf(changeFmtStr, "Listener:SSLConfiguration:Protocols", strings.Join(actual.Protocols, ","), strings.Join(desired.Protocols, ",")))
		}
	}

	return sslConfigurationChanges
}

func hasListenerChanged(logger *zap.SugaredLogger, actual client.GenericListener, desired client.GenericListener, ruleSets map[string]loadbalancer.RuleSetDetails) bool {
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
	if toBool(actual.IsPpv2Enabled) != toBool(desired.IsPpv2Enabled) {
		listenerChanges = append(listenerChanges, fmt.Sprintf(changeFmtStr, "Listener:IsPpv2Enabled", toBool(actual.IsPpv2Enabled), toBool(desired.IsPpv2Enabled)))
	}
	if ruleSets != nil && !sets.NewString(actual.RuleSetNames...).Equal(sets.NewString(desired.RuleSetNames...)) {
		listenerChanges = append(listenerChanges, fmt.Sprintf(changeFmtStr, "Listener:RuleSetNames", actual.RuleSetNames, desired.RuleSetNames))
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

func getListenerChanges(logger *zap.SugaredLogger, actual map[string]client.GenericListener, desired map[string]client.GenericListener, ruleSets map[string]loadbalancer.RuleSetDetails) []Action {
	var listenerActions []Action

	// set to keep track of desired listeners that already exist and should not be created
	exists := sets.NewString()
	//sanitizedDesiredListeners convert the listener name HTTP-xxxx to TCP-xxx such that in sortAndCombineAction can
	//place BackendSet create before Listener Create and Listener delete before BackendSet delete. Also it would help
	//not to delete and create Listener if customer edit the service and add oci-load-balancer-backend-protocol: "HTTP"
	// and vice versa. It would help to only update the listener in case of protocol change.
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
					RuleSetNames:          actualListener.RuleSetNames,
				},
				name:       name,
				actionType: Delete,
			})
			continue
		}
		exists.Insert(getSanitizedName(name))
		if hasListenerChanged(logger, actualListener, desiredListener, ruleSets) {
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

func getRuleSetChanges(actual map[string]loadbalancer.RuleSetDetails, desired map[string]loadbalancer.RuleSetDetails) []Action {
	var ruleSetActions []Action

	// First check to see if any rule sets need to be deleted.
	for name, a := range actual {
		_, ok := desired[name]
		if !ok {
			// No longer exists
			ruleSetActions = append(ruleSetActions, &RuleSetAction{
				name:           name,
				RuleSetDetails: a,
				actionType:     Delete,
			})
			continue
		}
	}

	// Now check if any need to be created or updated
	for name, desiredRuleSet := range desired {
		if _, ok := actual[name]; !ok {
			ruleSetActions = append(ruleSetActions, &RuleSetAction{
				name:           name,
				RuleSetDetails: desiredRuleSet,
				actionType:     Create,
			})
		} else {
			if !reflect.DeepEqual(actual[name], desired[name]) {
				ruleSetActions = append(ruleSetActions, &RuleSetAction{
					name:           name,
					RuleSetDetails: desiredRuleSet,
					actionType:     Update,
				})
			}
		}
	}

	return ruleSetActions
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
	return !DeepEqualLists(actualNetworkSecurityGroup, desiredNetworkSecurityGroup)
}

// hasIpVersionChanged checks if the IP version has changed
func hasIpVersionChanged(previousIpVersion, currentIpVersion string) bool {
	return !strings.EqualFold(previousIpVersion, currentIpVersion)
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
		name = fmt.Sprintf("%s", strings.Join(fields, "-"))
	}
	if len(fields) > 2 {
		if contains(fields, IPv6) {
			return fmt.Sprintf("%s", strings.Join(fields[:3], "-"))
		}
		return fmt.Sprintf("%s", strings.Join(fields[:2], "-"))
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

// generateNsgName gets the name of the NSG based on the service
func generateNsgName(service *api.Service) string {
	var name string
	name = fmt.Sprintf("%s/%s/%s/nsg", service.Namespace, service.Name, service.UID)
	if len(name) > 255 {
		// 255 is the max length for display name
		//https://docs.oracle.com/en-us/iaas/api/#/en/iaas/20160918/NetworkSecurityGroup/
		name = name[:255]
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

// GetSSLEnabledPorts returns a list of port numbers for which we need to enable
// SSL on the corresponding listener.
func GetSSLEnabledPorts(svc *api.Service) ([]int, error) {
	return getSSLEnabledPorts(svc)
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

// sortAndCombineActions combines three slices of Actions and then sorts them to
// ensure that BackendSets are created prior to their associated Listeners but
// deleted after their associated Listeners. Rule Sets are created/updated before any listener changes
// and deleted after listener changes.
func sortAndCombineActions(logger *zap.SugaredLogger, backendSetActions []Action, listenerActions []Action, ruleSetActions []Action) []Action {
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

	sort.SliceStable(actions, func(i, j int) bool {
		a1 := actions[i]
		a2 := actions[j]
		if a1.Type() != a2.Type() {
			return a1.Type() == Delete
		}
		return false
	})

	for _, a := range ruleSetActions {
		if a.Type() == Delete { // Rule Set can only be deleted if it's not attached to any Listener
			actions = append(actions, a)
		} else { // Rule Set needs to exist before it can be attached to a Listener. No requirements on updates.
			actions = append([]Action{a}, actions...)
		}
	}

	return actions
}

func getMetric(resourceType string, metricType string) string {
	if resourceType == LB {
		switch metricType {
		case Create:
			return metrics.LBProvision
		case Update:
			return metrics.LBUpdate
		case Delete:
			return metrics.LBDelete
		}
	}
	if resourceType == NLB {
		switch metricType {
		case Create:
			return metrics.NLBProvision
		case Update:
			return metrics.NLBUpdate
		case Delete:
			return metrics.NLBDelete
		}
	}

	if resourceType == NSG {
		switch metricType {
		case Create:
			return metrics.NSGProvision
		case Update:
			return metrics.NSGUpdate
		case Delete:
			return metrics.NSGDelete
		}
	}
	return ""
}

func parseFlexibleShapeBandwidth(shape, annotation string) (int, error) {
	reg, _ := regexp.Compile(nonAlphanumericRegexExpression)
	processedString := reg.ReplaceAllString(shape, "")
	if strings.HasSuffix(processedString, "Mbps") {
		processedString = strings.TrimSuffix(processedString, "Mbps")
	} else if strings.HasSuffix(processedString, "mbps") {
		processedString = strings.TrimSuffix(processedString, "mbps")
	}
	parsedIntFlexibleShape, err := strconv.Atoi(processedString)
	if err != nil {
		return 0, fmt.Errorf("invalid format for %s annotation : %v", annotation, shape)
	}
	return parsedIntFlexibleShape, nil
}

// GenericIpVersion returns the address of the value client.GenericIpVersion
func GenericIpVersion(value client.GenericIpVersion) *client.GenericIpVersion {
	return &value
}

// convertK8sIpFamiliesToOciIpVersion helper method to convert ipFamily string to GenericIpVersion
func convertK8sIpFamiliesToOciIpVersion(ipFamily string) client.GenericIpVersion {
	switch ipFamily {
	case IPv4:
		return client.GenericIPv4
	case IPv6:
		return client.GenericIPv6
	case IPv4AndIPv6:
		return client.GenericIPv4AndIPv6
	default:
		return client.GenericIPv4
	}
}

// convertOciIpVersionsToOciIpFamilies helper method to convert ociIpVersions slice to string slice
func convertOciIpVersionsToOciIpFamilies(ipVersions []client.GenericIpVersion) []string {
	if len(ipVersions) == 0 {
		return []string{IPv4}
	}
	k8sIpFamily := []string{}
	for _, ipVersion := range ipVersions {
		switch ipVersion {
		case client.GenericIPv4:
			k8sIpFamily = append(k8sIpFamily, IPv4)
		case client.GenericIPv6:
			k8sIpFamily = append(k8sIpFamily, IPv6)
		case client.GenericIPv4AndIPv6:
			k8sIpFamily = append(k8sIpFamily, IPv4AndIPv6)
		default:
			k8sIpFamily = append(k8sIpFamily, IPv4)
		}
	}
	return k8sIpFamily
}
