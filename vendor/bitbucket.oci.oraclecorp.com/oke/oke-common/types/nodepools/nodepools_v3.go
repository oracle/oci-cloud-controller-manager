package nodepools

import (
	"fmt"
	"sort"
	"strings"

	nodes "bitbucket.oci.oraclecorp.com/oke/oke-common/types/nodes"
)

// GetResponseV3 is the response type for the NodePoolList CLI API operation.
// TODO: This is not using the correct schema for the NodePoolList operation. It is calling /api/20180228/nodePools but should be using /api/20180222/nodePools
type GetResponseV3 struct {
	NodePools []NodePoolV3
}

// ToV3 converts a nodepools.GetResponse object to a NodePoolGetResponseV3 object understood by the higher layers
func (src *GetResponse) ToGetResponseV3(supportRegionalSubnet bool) GetResponseV3 {
	return toV3(src, supportRegionalSubnet)
}

func toV3(src *GetResponse, supportRegionalSubnet bool) GetResponseV3 {
	dst := GetResponseV3{
		NodePools: make([]NodePoolV3, 0),
	}
	if src == nil {
		return dst
	}
	for _, np := range src.NodePools {
		dst.NodePools = append(dst.NodePools, np.ToNodePoolV3(supportRegionalSubnet))
	}
	return dst
}

// NodePoolSummaryV3 is an array element in the response type for the ListNodePools API operation.
type NodePoolSummaryV3 struct {
	ID                string                     `json:"id"`
	Name              string                     `json:"name"`
	CompartmentID     string                     `json:"compartmentId"`
	ClusterID         string                     `json:"clusterId"`
	KubernetesVersion string                     `json:"kubernetesVersion"`
	NodeImageID       string                     `json:"nodeImageId"`
	NodeImageName     string                     `json:"nodeImageName"`
	NodeShape         string                     `json:"nodeShape"`
	InitialNodeLabels *[]KeyValueV3              `json:"initialNodeLabels,omitempty"`
	SSHPublicKey      string                     `json:"sshPublicKey"`
	QuantityPerSubnet uint32                     `json:"quantityPerSubnet"`
	SubnetIDs         []string                   `json:"subnetIds"`
	NodeMetadata      map[string]string          `json:"nodeMetadata,omitempty"`
	NodeConfigDetails *NodePoolNodeConfigDetails `json:"nodeConfigDetails,omitempty"`
}

// ToNodePoolSummaryV3s converts a nodepools.GetResponse to a NodePoolSummaryV3 object understood by the higher layers
func (src *GetResponse) ToNodePoolSummaryV3s(supportRegionalSubnet bool) []NodePoolSummaryV3 {
	return toSummaryV3(src, supportRegionalSubnet)
}

func toSummaryV3(src *GetResponse, supportRegionalSubnet bool) []NodePoolSummaryV3 {
	resp := []NodePoolSummaryV3{}
	if src == nil {
		return resp
	}

	for _, np := range src.NodePools {
		resp = append(resp, np.toSummaryV3(supportRegionalSubnet))
	}
	return resp
}

// ToSummaryV3 converts a nodepools.GetResponse to a NodePoolSummaryV3 object understood by the higher layers
func (src *NodePool) toSummaryV3(supportRegionalSubnet bool) NodePoolSummaryV3 {
	dst := NodePoolSummaryV3{}

	if src == nil {
		return dst
	}

	dst.ID = src.ID
	dst.Name = src.Name
	dst.CompartmentID = src.CompartmentID
	dst.ClusterID = src.ClusterID
	dst.KubernetesVersion = src.K8SVersion
	dst.NodeImageID = src.NodeImageID
	dst.NodeImageName = src.NodeImageName
	dst.NodeShape = src.NodeShape
	initialNodeLabels := KeyValuesFromString(src.InitialNodeLabels)
	dst.InitialNodeLabels = &initialNodeLabels
	dst.SSHPublicKey = src.SSHPublicKey

	// fill in Size/placementConfigs/SubnetIDs all the time
	// fill in QuantityPerSubnet if QuantityPerSubnet > 0
	dst.NodeConfigDetails = &NodePoolNodeConfigDetails{
		PlacementConfigs: convertSubnetInfoToPlacementConfigs(src.SubnetsInfo)}
	dst.SubnetIDs = convertSubnetInfoToUniqueSubnetIds(src.SubnetsInfo)
	if src.Size > 0 {
		dst.NodeConfigDetails.Size = src.Size
		dst.QuantityPerSubnet = 0
	} else { // QuantityPerSubnet > 0
		dst.NodeConfigDetails.Size = src.QuantityPerSubnet * uint32(len(src.SubnetsInfo))
		dst.QuantityPerSubnet = src.QuantityPerSubnet
	}

	// Do not show NodeConfigDetails if regional subnet is not whitelisted
	if !supportRegionalSubnet {
		dst.NodeConfigDetails = nil
	}

	dst.NodeMetadata = make(map[string]string)
	for k, v := range src.NodeMetadata {
		dst.NodeMetadata[k] = v
	}

	return dst
}

// NodePoolV3 is the response type for the GetNodePool API operation.
type NodePoolV3 struct {
	ID                string                     `json:"id"`
	Name              string                     `json:"name"`
	CompartmentID     string                     `json:"compartmentId"`
	ClusterID         string                     `json:"clusterId"`
	KubernetesVersion string                     `json:"kubernetesVersion"`
	NodeImageID       string                     `json:"nodeImageId"`
	NodeImageName     string                     `json:"nodeImageName"`
	NodeShape         string                     `json:"nodeShape"`
	InitialNodeLabels *[]KeyValueV3              `json:"initialNodeLabels,omitempty"`
	SSHPublicKey      string                     `json:"sshPublicKey"`
	QuantityPerSubnet uint32                     `json:"quantityPerSubnet"`
	SubnetIDs         []string                   `json:"subnetIds"`
	Nodes             []nodes.NodeV3             `json:"nodes,omitempty"`
	NodeMetadata      map[string]string          `json:"nodeMetadata,omitempty"`
	NodeConfigDetails *NodePoolNodeConfigDetails `json:"nodeConfigDetails,omitempty"`
}

// ToV3 converts a NodePool object to a NodePoolV3 object understood by the higher layers
// supportRegionalSubnet is a whitelist flag
func (np *NodePool) ToNodePoolV3(supportRegionalSubnet bool) NodePoolV3 {
	return toNodePoolV3(np, supportRegionalSubnet)
}

func toNodePoolV3(src *NodePool, supportRegionalSubnet bool) NodePoolV3 {
	dst := NodePoolV3{}
	if src == nil {
		return dst
	}
	dst.ID = src.ID
	dst.Name = src.Name
	dst.CompartmentID = src.CompartmentID
	dst.ClusterID = src.ClusterID
	dst.KubernetesVersion = src.K8SVersion
	dst.NodeImageID = src.NodeImageID
	dst.NodeImageName = src.NodeImageName
	dst.NodeShape = src.NodeShape
	initialNodeLabels := KeyValuesFromString(src.InitialNodeLabels)
	dst.InitialNodeLabels = &initialNodeLabels
	dst.SSHPublicKey = src.SSHPublicKey

	// fill in SubnetIDs all the time
	// fill in QuantityPerSubnet if QuantityPerSubnet exists
	dst.NodeConfigDetails = &NodePoolNodeConfigDetails{
		PlacementConfigs: convertSubnetInfoToPlacementConfigs(src.SubnetsInfo)}
	dst.SubnetIDs = convertSubnetInfoToUniqueSubnetIds(src.SubnetsInfo)
	if src.Size > 0 {
		dst.NodeConfigDetails.Size = src.Size
		dst.QuantityPerSubnet = 0
	} else { // QuantityPerSubnet >= 0
		dst.NodeConfigDetails.Size = src.QuantityPerSubnet * uint32(len(src.SubnetsInfo))
		dst.QuantityPerSubnet = src.QuantityPerSubnet
	}

	// Do not show NodeConfigDetails if regional subnet is not whitelisted
	if !supportRegionalSubnet {
		dst.NodeConfigDetails = nil
	}

	for _, nd := range src.NodeStates {
		dst.Nodes = append(dst.Nodes, nd.ToV3())
	}

	dst.NodeMetadata = make(map[string]string)
	for k, v := range src.NodeMetadata {
		dst.NodeMetadata[k] = v
	}

	return dst
}

// CreateNodePoolDetailsV3 is the request body for a CreateNodePool operation.
type CreateNodePoolDetailsV3 struct {
	Name              string                    `json:"name"`
	CompartmentID     string                    `json:"compartmentId"`
	ClusterID         string                    `json:"clusterId"`
	KubernetesVersion string                    `json:"kubernetesVersion"`
	NodeImageName     string                    `json:"nodeImageName"`
	NodeShape         string                    `json:"nodeShape"`
	NodeMetadata      map[string]string         `json:"nodeMetadata"`
	InitialNodeLabels []KeyValueV3              `json:"initialNodeLabels,omitempty"`
	SSHPublicKey      string                    `json:"sshPublicKey"`
	QuantityPerSubnet uint32                    `json:"quantityPerSubnet"`
	SubnetIDs         []string                  `json:"subnetIds"`
	NodeConfigDetails NodePoolNodeConfigDetails `json:"nodeConfigDetails,omitempty"`
}

// UpdateNodePoolDetailsV3 is the request body for a UpdateNodePool operation.
type UpdateNodePoolDetailsV3 struct {
	Name              string                    `json:"name"`
	KubernetesVersion string                    `json:"kubernetesVersion"`
	QuantityPerSubnet uint32                    `json:"quantityPerSubnet"`
	InitialNodeLabels []KeyValueV3              `json:"initialNodeLabels,omitempty"`
	SubnetIDs         []string                  `json:"subnetIds"`
	NodeConfigDetails NodePoolNodeConfigDetails `json:"nodeConfigDetails,omitempty"`
}

// NodePoolOptionsV3 contains the options available for specific fields that can be submitted
// to the CreateNodePool operation. Used by the
type NodePoolOptionsV3 struct {
	KubernetesVersions []string `json:"kubernetesVersions"`
	Images             []string `json:"images"`
	Shapes             []string `json:"shapes"`
}

// NodePoolNodeConfigDetails Contains the size and placement configuration of
// the node pool.
type NodePoolNodeConfigDetails struct {
	Size             uint32                           `json:"size"`
	PlacementConfigs []NodePoolPlacementConfigDetails `json:"placementConfigs"`
}

// NodePoolPlacementConfigDetails contains the AD info of the subnet. A regional subnet
// spans all ADs in a region.
type NodePoolPlacementConfigDetails struct {
	AvailabilityDomain string `json:"availabilityDomain"`
	SubnetID           string `json:"subnetId"`
}

// ToProto converts a CreateNodePoolDetailsV3 to a NodePoolNewRequest object understood by grpc
func (v3 *CreateNodePoolDetailsV3) ToProto() *NewRequest {
	var dst NewRequest
	if v3 != nil {
		dst.Name = v3.Name
		dst.CompartmentID = v3.CompartmentID
		dst.ClusterID = v3.ClusterID
		dst.K8SVersion = v3.KubernetesVersion
		dst.NodeImageName = v3.NodeImageName
		dst.NodeShape = v3.NodeShape

		initialNodeLabels := KeyValuesToString(v3.InitialNodeLabels)
		dst.InitialNodeLabels = initialNodeLabels

		dst.SSHPublicKey = v3.SSHPublicKey
		dst.SubnetsInfo = make(map[string]*SubnetInfo)

		// nodeConfigDetails model.
		if len(v3.NodeConfigDetails.PlacementConfigs) > 0 {
			dst.Size = v3.NodeConfigDetails.Size
			for _, placementConfig := range v3.NodeConfigDetails.PlacementConfigs {
				dst.SubnetsInfo[CreateNodePoolSubnetInfoKey(placementConfig)] = &SubnetInfo{
					ID: placementConfig.SubnetID,
					AD: placementConfig.AvailabilityDomain,
				}
			}
		} else { // legacy model, use QuantityPerSubnet/subnetIds
			dst.QuantityPerSubnet = v3.QuantityPerSubnet
			for _, id := range v3.SubnetIDs {
				dst.SubnetsInfo[id] = &SubnetInfo{ID: id}
			}
		}

		dst.NodeMetadata = make(map[string]string)
		for k, v := range v3.NodeMetadata {
			dst.NodeMetadata[k] = v
		}
	}

	return &dst
}

// KeyValueV3 is used for holding a key/value pair whose value is a string
type KeyValueV3 struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// String converts a key value pair to key=value pair
func (kv *KeyValueV3) String() string {
	return fmt.Sprintf("%s=%s", strings.TrimSpace(kv.Key), strings.TrimSpace(kv.Value))
}

// KeyValuesToString converts a slice of KeyValueV3 to a comma-separated string of key=value pairs
func KeyValuesToString(keyValues []KeyValueV3) string {
	kvs := []string{}
	for _, kv := range keyValues {
		if len(kv.Key) > 0 {
			kvs = append(kvs, kv.String())
		}
	}
	return strings.Join(kvs, ",")
}

// KeyValuesFromString creates a slice of KeyValueV3 from a comma-separated string of key=value pairs
func KeyValuesFromString(s string) []KeyValueV3 {
	keyValues := []KeyValueV3{}
	if len(s) > 0 {
		pairs := strings.Split(s, ",")
		for _, p := range pairs {
			parts := strings.Split(p, "=")
			kv := KeyValueV3{
				Key: parts[0],
			}
			if len(parts) > 1 {
				kv.Value = parts[1]
			}
			keyValues = append(keyValues, kv)
		}
	}
	return keyValues
}

// NodePoolCreateCLIResponseV3 is used by the CLI when creating a node pool
type NodePoolCreateCLIResponseV3 struct {
	WorkRequestID string `json:"workRequestId"`
}

// NodePoolDeleteCLIResponseV3 is used by the CLI when deleting a node pool
type NodePoolDeleteCLIResponseV3 struct {
	WorkRequestID string `json:"workRequestId"`
}

// NodePoolUpdateCLIResponseV3 is used by the CLI when updating a node pool
type NodePoolUpdateCLIResponseV3 struct {
	WorkRequestID string `json:"workRequestId"`
}

// helper functions
func convertSubnetInfoToUniqueSubnetIds(subnetsInfo map[string]*SubnetInfo) []string {
	subnetIds := make([]string, 0, len(subnetsInfo))
	seenSubnetIds := make(map[string]bool, len(subnetsInfo))
	for _, subnetInfo := range subnetsInfo {
		if _, ok := seenSubnetIds[subnetInfo.ID]; ok {
			continue
		}

		subnetIds = append(subnetIds, subnetInfo.ID)
		seenSubnetIds[subnetInfo.ID] = true
	}
	return subnetIds
}

func convertSubnetInfoToPlacementConfigs(subnetsInfo map[string]*SubnetInfo) []NodePoolPlacementConfigDetails {
	placementConfigs := make([]NodePoolPlacementConfigDetails, 0, len(subnetsInfo))
	for _, subnetInfo := range subnetsInfo {
		placementConfig := NodePoolPlacementConfigDetails{
			AvailabilityDomain: subnetInfo.GetAD(),
			SubnetID:           subnetInfo.GetID(),
		}
		placementConfigs = append(placementConfigs, placementConfig)
	}

	sort.SliceStable(placementConfigs, func(i, j int) bool {
		return placementConfigs[i].AvailabilityDomain < placementConfigs[j].AvailabilityDomain ||
			(placementConfigs[i].AvailabilityDomain == placementConfigs[j].AvailabilityDomain &&
				placementConfigs[i].SubnetID < placementConfigs[j].SubnetID)
	})

	return placementConfigs
}

func CreateNodePoolSubnetInfoKey(placementcfg NodePoolPlacementConfigDetails) string {
	keyPattern := "%s:%s"
	return fmt.Sprintf(keyPattern, placementcfg.AvailabilityDomain, placementcfg.SubnetID)
}
