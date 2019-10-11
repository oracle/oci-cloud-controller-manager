package nodes

import (
	"bitbucket.oci.oraclecorp.com/oke/oke-common/types/apierrors"
)

// NodeV3 is a single node in a nodepool.
type NodeV3 struct {
	ID                 string             `json:"id"`
	Name               string             `json:"name"`
	KubernetesVersion  string             `json:"kubernetesVersion"`
	AvailabilityDomain string             `json:"availabilityDomain"`
	SubnetID           string             `json:"subnetId"`
	NodePoolID         string             `json:"nodePoolId"`
	PublicIP           string             `json:"publicIp"`
	Error              *apierrors.ErrorV3 `json:"nodeError,omitempty"`
	LifecycleState     string             `json:"lifecycleState"`
	LifecycleDetails   string             `json:"lifecycleDetails"`
}

// ToV3 converts a Node object to a NodeV3 object understood by the higher layers
func (src *NodeState) ToV3() NodeV3 {
	dst := NodeV3{}
	if src == nil {
		return dst
	}
	dst.ID = src.ID
	dst.Name = src.Name
	dst.KubernetesVersion = src.K8SVersion
	dst.AvailabilityDomain = src.AvailabilityDomain
	dst.SubnetID = src.SubnetID
	dst.NodePoolID = src.NodePoolID
	dst.PublicIP = src.PublicIP

	// if node contains ErrorOCI return that to user else return generic apierrors.ErrorV3
	if src.ErrorOCI != nil {
		dst.Error = src.ErrorOCI.ToV3()
	} else if len(src.Error) > 0 {
		dst.Error = &apierrors.ErrorV3{
			Code:    apierrors.UnknownNodeErrorCode,
			Message: src.Error,
		}
	}
	dst.LifecycleState = src.State.String()
	dst.LifecycleDetails = src.StateDetails
	return dst
}
