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
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/glog"
	baremetal "github.com/oracle/bmcs-go-sdk"

	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	sslCertificateFileName = "tls.crt"
	sslPrivateKeyFileName  = "tls.key"
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

type Action interface {
	Type() ActionType
	Name() string
}

// BackendSetAction denotes the action that should be taken on the given backend set.
type BackendSetAction struct {
	Action

	actionType ActionType
	BackendSet baremetal.BackendSet
}

// Type of the Action
func (b *BackendSetAction) Type() ActionType {
	return b.actionType
}

// Name of the action's object
func (b *BackendSetAction) Name() string {
	return b.BackendSet.Name
}

func (b *BackendSetAction) String() string {
	return fmt.Sprintf("BackendSetAction:{Name: %s, Type: %v }", b.Name(), b.actionType)
}

// ListenerAction denotes the action that should be taken on the given listener.
type ListenerAction struct {
	Action

	actionType ActionType
	Listener   baremetal.Listener
}

// Type of the Action
func (l *ListenerAction) Type() ActionType {
	return l.actionType
}

// Name of the action's object
func (l *ListenerAction) Name() string {
	return l.Listener.Name
}

func (l *ListenerAction) String() string {
	return fmt.Sprintf("ListenerAction:{Name: %s, Type: %v }", l.Name(), l.actionType)
}

// TODO(horwitz): this doesn't check weight which we may want in the future to evenly distribute Local traffic policy load.
func hasBackendSetChanged(actual, desired baremetal.BackendSet) bool {
	if !reflect.DeepEqual(actual.HealthChecker, desired.HealthChecker) {
		return true
	}

	if actual.Policy != desired.Policy {
		return true
	}

	if len(actual.Backends) != len(desired.Backends) {
		return true
	}

	nameFormat := "%s:%d"

	// Since the lengths are equal that means the membership must be the same else
	// there has been change.
	desiredSet := sets.NewString()
	for _, backend := range desired.Backends {
		name := fmt.Sprintf(nameFormat, backend.IPAddress, backend.Port)
		desiredSet.Insert(name)
	}

	for _, backend := range actual.Backends {
		name := fmt.Sprintf(nameFormat, backend.IPAddress, backend.Port)
		if !desiredSet.Has(name) {
			return true
		}
	}

	return false
}

func getBackendSetChanges(actual, desired map[string]baremetal.BackendSet) []Action {
	var backendSetActions []Action
	// First check to see if any backendsets need to be deleted or updated.
	for name, actualBackendSet := range actual {
		desiredBackendSet, ok := desired[name]
		if !ok {
			// no longer exists
			backendSetActions = append(backendSetActions, &BackendSetAction{
				BackendSet: actualBackendSet,
				actionType: Delete,
			})
			continue
		}

		if hasBackendSetChanged(actualBackendSet, desiredBackendSet) {
			backendSetActions = append(backendSetActions, &BackendSetAction{
				BackendSet: desiredBackendSet,
				actionType: Update,
			})
		}
	}

	// Now check if any need to be created.
	for name, desiredBackendSet := range desired {
		if _, ok := actual[name]; !ok {
			// doesn't exist so lets create it
			backendSetActions = append(backendSetActions, &BackendSetAction{
				BackendSet: desiredBackendSet,
				actionType: Create,
			})
		}
	}

	return backendSetActions
}

func hasListenerChanged(actual, desired baremetal.Listener) bool {
	return !reflect.DeepEqual(actual, desired)
}

func getListenerChanges(actual, desired map[string]baremetal.Listener) []Action {
	var listenerActions []Action
	// First check to see if any listeners need to be deleted or updated.
	for name, actualListener := range actual {
		desiredListener, ok := desired[name]
		if !ok {
			// no longer exists
			listenerActions = append(listenerActions, &ListenerAction{
				Listener:   actualListener,
				actionType: Delete,
			})
			continue
		}

		if hasListenerChanged(actualListener, desiredListener) {
			listenerActions = append(listenerActions, &ListenerAction{
				Listener:   desiredListener,
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
				actionType: Create,
			})
		}
	}

	return listenerActions
}

func sslEnabled(sslConfigMap map[int]*baremetal.SSLConfiguration) bool {
	return len(sslConfigMap) > 0
}

func getListenerName(protocol string, port int, sslConfig *baremetal.SSLConfiguration) string {
	if sslConfig != nil {
		return fmt.Sprintf("%s-%d-%s", protocol, port, sslConfig.CertificateName)
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

// getSSLEnabledPorts returns a set (implemented as a map) of port numbers for
// which we need to enable SSL on the corresponding listener.
func getSSLEnabledPorts(annotations map[string]string) (map[int]bool, error) {
	sslPortsAnnotation, ok := annotations[ServiceAnnotationLoadBalancerSSLPorts]
	if !ok {
		return nil, nil
	}

	sslPorts := make(map[int]bool)
	for _, sslPort := range strings.Split(sslPortsAnnotation, ",") {
		i, err := strconv.Atoi(strings.TrimSpace(sslPort))
		if err != nil {
			return nil, fmt.Errorf("parse SSL port: %v", err)
		}
		sslPorts[i] = true
	}
	return sslPorts, nil
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

func sortAndCombineActions(backendSetActions []Action, listenerActions []Action) []Action {
	actions := append(backendSetActions, listenerActions...)
	sort.Slice(actions, func(i, j int) bool {
		// One action will be backendset and one will be the listener
		a1 := actions[i]
		a2 := actions[j]
		if a1.Name() != a2.Name() {
			// Since the actions aren't for the same listener/backendset then just
			// sort by the name until we get to the point we are
			return a1.Name() < a2.Name()
		}

		// For create and delete (which is what we really care about) the ActionType
		// will always be the same so we can get away with just checking the first action.
		switch a1.Type() {
		case Create:
			// Create the BackendSet then Listener
			_, ok := a1.(*BackendSetAction)
			return ok
		case Update:
			// Doesn't matter
			return true
		case Delete:
			// Delete the Listener then BackendSet
			_, ok := a2.(*BackendSetAction)
			return ok
		default:
			// Should never be reachable.
			glog.Errorf("Unknown action type received: %+v", a1)
			return true
		}
	})
	return actions
}

func getBackendPort(backends []baremetal.Backend) uint64 {
	// TODO: what happens if this is 0? e.g. we scale the pods to 0 for a deployment
	return uint64(backends[0].Port)
}
