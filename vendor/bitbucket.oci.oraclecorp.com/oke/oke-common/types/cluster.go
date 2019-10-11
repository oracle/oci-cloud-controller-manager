package types

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// ClusterStateForAPICreating is the creating lifecycle state used in API responses
	ClusterStateForAPICreating = "CREATING"
	// ClusterStateForAPIActive is the active lifecycle state used in API responses
	ClusterStateForAPIActive = "ACTIVE"
	// ClusterStateForAPIFailed is the failed lifecycle state used in API responses
	ClusterStateForAPIFailed = "FAILED"
	// ClusterStateForAPIDeleting is the deleting lifecycle state used in API responses
	ClusterStateForAPIDeleting = "DELETING"
	// ClusterStateForAPIDeleted is the deleted lifecycle state used in API responses
	ClusterStateForAPIDeleted = "DELETED"
	// ClusterStateForAPIUpdating is the updating lifecycle state used in API responses
	ClusterStateForAPIUpdating = "UPDATING"
)

// ClusterSummaryV3 is an array element in the response type for the ListClusters API operation.
type ClusterSummaryV3 struct {
	ID                          string                  `json:"id"`
	Name                        string                  `json:"name"`
	CompartmentID               string                  `json:"compartmentId"`
	VCNID                       string                  `json:"vcnId"`
	KubernetesVersion           string                  `json:"kubernetesVersion"`
	Options                     *ClusterCreateOptionsV3 `json:"options,omitempty"`
	Metadata                    map[string]string       `json:"metadata"`
	LifecycleState              string                  `json:"lifecycleState"`
	LifecycleDetails            string                  `json:"lifecycleDetails"`
	Endpoints                   *ClusterEndpointsV3     `json:"endpoints"`
	AvailableKubernetesUpgrades []string                `json:"availableKubernetesUpgrades"`
}

// ClusterV3 is the response type for the GetCluster API operation.
type ClusterV3 struct {
	ID                          string                  `json:"id"`
	Name                        string                  `json:"name"`
	CompartmentID               string                  `json:"compartmentId"`
	VCNID                       string                  `json:"vcnId"`
	KubernetesVersion           string                  `json:"kubernetesVersion"`
	KMSKeyID                    string                  `json:"kmsKeyId,omitempty"`
	Options                     *ClusterCreateOptionsV3 `json:"options,omitempty"`
	Metadata                    map[string]string       `json:"metadata"`
	LifecycleState              string                  `json:"lifecycleState"`
	LifecycleDetails            string                  `json:"lifecycleDetails"`
	Endpoints                   *ClusterEndpointsV3     `json:"endpoints"`
	AvailableKubernetesUpgrades []string                `json:"availableKubernetesUpgrades"`
}

// ToSummaryV3 a K8Instance object to a ClusterSummaryV3 object understood by the higher layers
func (src *K8Instance) ToSummaryV3() *ClusterSummaryV3 {
	var dst ClusterSummaryV3
	if src == nil {
		return &dst
	}

	dst.ID = src.ID
	dst.Name = src.Name
	dst.CompartmentID = src.CompartmentID
	dst.VCNID = src.NetworkConfig.VCNID
	dst.KubernetesVersion = src.K8Version
	dst.Options = &ClusterCreateOptionsV3{
		KubernetesNetworkConfig: &KubernetesNetworkConfigV3{
			PodsCIDR:     src.NetworkConfig.K8SPodsCIDR,
			ServicesCIDR: src.NetworkConfig.K8SServicesCIDR,
		},
	}
	if src.InstallOptions != nil {
		dst.Options.AddOns = &AddOnOptionsV3{
			Tiller:              src.InstallOptions.HasTiller,
			KubernetesDashboard: src.InstallOptions.HasKubernetesDashboard,
		}
	}
	dst.Options.ServiceLBSubnetIDs = make([]string, len(src.NetworkConfig.ServiceLBSubnets))
	for idx, sn := range src.NetworkConfig.ServiceLBSubnets {
		dst.Options.ServiceLBSubnetIDs[idx] = sn
	}
	dst.Metadata = src.Metadata
	dst.LifecycleState = src.TKMState.ToAPIV3()
	// dst.LifecycleDetails = "FIXME"

	dst.Endpoints = &ClusterEndpointsV3{
		Kubernetes: src.K8Addr,
	}

	dst.AvailableKubernetesUpgrades = []string{}
	for _, k8Ver := range src.AvailableK8SUpgrades {
		dst.AvailableKubernetesUpgrades = append(dst.AvailableKubernetesUpgrades, k8Ver)
	}

	return &dst
}

// ToV3 converts a K8Instance object to a ClusterV3 object understood by the higher layers
func (src *K8Instance) ToV3() *ClusterV3 {
	var dst ClusterV3
	if src == nil {
		return &dst
	}

	dst.ID = src.ID
	dst.Name = src.Name
	dst.CompartmentID = src.CompartmentID
	dst.VCNID = src.NetworkConfig.VCNID
	dst.KubernetesVersion = src.K8Version
	dst.KMSKeyID = src.KMSKeyID
	dst.Options = &ClusterCreateOptionsV3{
		KubernetesNetworkConfig: &KubernetesNetworkConfigV3{
			PodsCIDR:     src.NetworkConfig.K8SPodsCIDR,
			ServicesCIDR: src.NetworkConfig.K8SServicesCIDR,
		},
		AddOns: &AddOnOptionsV3{
			Tiller:              src.InstallOptions.HasTiller,
			KubernetesDashboard: src.InstallOptions.HasKubernetesDashboard,
		},
	}
	dst.Options.ServiceLBSubnetIDs = make([]string, len(src.NetworkConfig.ServiceLBSubnets))
	for idx, sn := range src.NetworkConfig.ServiceLBSubnets {
		dst.Options.ServiceLBSubnetIDs[idx] = sn
	}
	dst.Metadata = src.Metadata
	dst.LifecycleState = src.TKMState.ToAPIV3()
	// dst.LifecycleDetails = "FIXME"

	dst.Endpoints = &ClusterEndpointsV3{
		Kubernetes: src.K8Addr,
	}

	dst.AvailableKubernetesUpgrades = []string{}
	for _, k8Ver := range src.AvailableK8SUpgrades {
		dst.AvailableKubernetesUpgrades = append(dst.AvailableKubernetesUpgrades, k8Ver)
	}

	return &dst
}

// ToAPIV3 converts the state to valid string used by the API
func (s TKMState) ToAPIV3() string {
	switch s {
	case TKMState_Initializing:
		return ClusterStateForAPICreating
	case TKMState_Running:
		return ClusterStateForAPIActive
	case TKMState_Failed:
		return ClusterStateForAPIFailed
	case TKMState_Terminating:
		return ClusterStateForAPIDeleting
	case TKMState_Terminated:
		return ClusterStateForAPIDeleted
	case TKMState_Updating_Masters:
		return ClusterStateForAPIUpdating
	default:
		return "UNKNOWN"
	}
}

// ClusterEndpointsV3 contains the endpoint URLs for the Kubernetes API server.
type ClusterEndpointsV3 struct {
	Kubernetes string `json:"kubernetes"`
}

// CreateClusterDetailsV3 is the request body type for the CreateCluster API operation.
type CreateClusterDetailsV3 struct {
	Name              string                  `json:"name" yaml:"name"`
	KubernetesVersion string                  `json:"kubernetesVersion" yaml:"kubernetesVersion"`
	CompartmentID     string                  `json:"compartmentId" yaml:"compartmentId"`
	VCNID             string                  `json:"vcnId" yaml:"vcnId"`
	KMSKeyID          string                  `json:"kmsKeyId,omitempty" yaml:"kmsKeyId,omitempty"`
	Options           *ClusterCreateOptionsV3 `json:"options,omitempty" yaml:"options,omitempty"`
}

// ClusterCreateOptionsV3 defines the options that can modify how a cluster is created.
type ClusterCreateOptionsV3 struct {
	ServiceLBSubnetIDs      []string                   `json:"serviceLbSubnetIds" yaml:"serviceLbSubnetIds"`
	KubernetesNetworkConfig *KubernetesNetworkConfigV3 `json:"kubernetesNetworkConfig,omitempty" yaml:"kubernetesNetworkConfig,omitempty"`
	AddOns                  *AddOnOptionsV3            `json:"addOns,omitempty" yaml:"addOns,omitempty"`
}

// KubernetesNetworkConfigV3 defines the networking to use in creating a cluster
type KubernetesNetworkConfigV3 struct {
	PodsCIDR     string `json:"podsCidr" yaml:"podsCidr"`
	ServicesCIDR string `json:"servicesCidr" yaml:"servicesCidr"`
}

// AddOnOptionsV3 defines the add-ons to use in creating a cluster
type AddOnOptionsV3 struct {
	Tiller              bool `json:"isTillerEnabled" yaml:"isTillerEnabled"`
	KubernetesDashboard bool `json:"isKubernetesDashboardEnabled" yaml:"isKubernetesDashboardEnabled"`
}

// ClusterCreateCLIResponseV3 is used by the CLI when creating a cluster
type ClusterCreateCLIResponseV3 struct {
	WorkRequestID string `json:"workRequestId"`
}

// UpdateClusterDetailsV3 is the request body type for the UpdateCluster API operation.
type UpdateClusterDetailsV3 struct {
	Name              string `json:"name"`
	KubernetesVersion string `json:"kubernetesVersion"`
}

func getK8SSemverNumbers(version string) ([]int64, error) {
	numbers := make([]int64, 3, 3)
	semvers := strings.Split(strings.TrimPrefix(version, "v"), ".")
	for i, v := range semvers {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Incompatible version supplied")
		}
		numbers[i] = n
	}
	return numbers, nil
}

// ClusterOptionsV3 contains the options available for specific fields that can be submitted
// to the CreateCluster operation.
type ClusterOptionsV3 struct {
	KubernetesVersions []string `json:"kubernetesVersions"`
}

// ClusterDeleteCLIResponseV3 is used by the CLI when deleting a cluster
type ClusterDeleteCLIResponseV3 struct {
	WorkRequestID string `json:"workRequestId"`
}

// APIOptionsV3 is a type to hold all of the node pools api options
type APIOptionsV3 struct {
	KubernetesVersions []string `json:"kubernetesVersions"`
}

// ClusterUpdateCLIResponseV3 is used by the CLI when updating a cluster
type ClusterUpdateCLIResponseV3 struct {
	WorkRequestID string `json:"workRequestId"`
}
