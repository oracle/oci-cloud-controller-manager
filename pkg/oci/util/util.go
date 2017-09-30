package util

import (
	"strings"

	api "k8s.io/api/core/v1"
)

const (
	// ProviderName uniquely identifies the Oracle Bare Metal Cloud Services (OCI)
	// cloud-provider.
	ProviderName   = "oci"
	providerPrefix = ProviderName + "://"
)

// MapProviderIDToInstanceID parses the provider id and returns the instance ocid.
func MapProviderIDToInstanceID(providerID string) string {
	if strings.HasPrefix(providerID, providerPrefix) {
		return strings.TrimPrefix(providerID, providerPrefix)
	}
	return providerID
}

// NodeInternalIP returns the nodes internal ip
// A node managed by the CCM will always have an internal ip
// since it's not possible to deploy an instance without a private ip.
func NodeInternalIP(node *api.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == api.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}
