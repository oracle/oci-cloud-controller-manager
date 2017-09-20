package bmcs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	baremetal "github.com/oracle/bmcs-go-sdk"

	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/cloudprovider"
)

const (
	sslCertificateFileName = "tls.crt"
	sslPrivateKeyFileName  = "tls.key"
)

// ActionType specifies what action should be taken on the resource
type ActionType string

const (
	// Create the resource as it doesn't exist yet
	Create = "create"
	// Update the resource
	Update = "update"
	// Delete the resource
	Delete = "delete"
)

// BackendSetAction denotes the action that should be taken on the given backend set.
type BackendSetAction struct {
	Type       ActionType
	BackendSet baremetal.BackendSet
}

// ListenerAction denotes the action that should be taken on the given listener.
type ListenerAction struct {
	Type     ActionType
	Listener baremetal.Listener
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

func getBackendSetChanges(actual, desired map[string]baremetal.BackendSet) []BackendSetAction {
	var backendSetActions []BackendSetAction
	// First check to see if any backendsets need to be deleted or updated.
	for name, actualBackendSet := range actual {
		desiredBackendSet, ok := desired[name]
		if !ok {
			// no longer exists
			backendSetActions = append(backendSetActions, BackendSetAction{
				BackendSet: actualBackendSet,
				Type:       Delete,
			})
			continue
		}

		if hasBackendSetChanged(actualBackendSet, desiredBackendSet) {
			backendSetActions = append(backendSetActions, BackendSetAction{
				BackendSet: desiredBackendSet,
				Type:       Update,
			})
		}
	}

	// Now check if any need to be created.
	for name, desiredBackendSet := range desired {
		_, ok := actual[name]
		if !ok {
			// doesn't exist so lets create it
			backendSetActions = append(backendSetActions, BackendSetAction{
				BackendSet: desiredBackendSet,
				Type:       Create,
			})
		}
	}

	return backendSetActions
}

func hasListenerChanged(actual, desired baremetal.Listener) bool {
	return !reflect.DeepEqual(actual, desired)
}

func getListenerChanges(actual, desired map[string]baremetal.Listener) []ListenerAction {
	var listenerActions []ListenerAction
	// First check to see if any listeners need to be deleted or updated.
	for name, actualListener := range actual {
		desiredListener, ok := desired[name]

		if !ok {
			// no longer exists
			listenerActions = append(listenerActions, ListenerAction{
				Listener: actualListener,
				Type:     Delete,
			})
			continue
		}

		if hasListenerChanged(actualListener, desiredListener) {
			listenerActions = append(listenerActions, ListenerAction{
				Listener: desiredListener,
				Type:     Update,
			})
		}
	}

	// Now check if any need to be created.
	for name, desiredListener := range desired {
		_, ok := actual[name]
		if !ok {
			// doesn't exist so lets create it
			listenerActions = append(listenerActions, ListenerAction{
				Listener: desiredListener,
				Type:     Create,
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
	return fmt.Sprintf("%s-%s", service.Name, cloudprovider.GetLoadBalancerName(service))
}

// Extract a list of all the external IP addresses for the available Kubernetes nodes
// Each node IP address must be added to the backend set
func extractNodeIPs(nodes []*api.Node) []string {
	nodeIPs := []string{}
	for _, node := range nodes {
		for _, nodeAddr := range node.Status.Addresses {
			if nodeAddr.Type == api.NodeInternalIP {
				nodeIPs = append(nodeIPs, nodeAddr.Address)
			}
		}
	}
	return nodeIPs
}

// validateProtocols validates that BMCS supports the protocol of all
// ServicePorts defined by a service.
func validateProtocols(servicePorts []api.ServicePort) error {
	for _, servicePort := range servicePorts {
		if servicePort.Protocol == api.ProtocolUDP {
			return fmt.Errorf("BMCS load balancers do not support UDP")
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
