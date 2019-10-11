package types

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"bitbucket.oci.oraclecorp.com/oke/oke-common/protobuf"
)

const (
	// ServiceAuthFlag indicates dataplane access is via Service Principals
	ServiceAuthFlag = "SERVICE_AUTH"
	// K8InstanceMetadataCreatedAtKey is the key for the creation time of the K8Instance
	K8InstanceMetadataCreatedAtKey = "timeCreated"
	// K8InstanceMetadataCreatedByKey is the key for the ID of the creator of the K8Instance
	K8InstanceMetadataCreatedByKey = "createdByUserId"
	// K8InstanceMetadataCreatedByWorkItemKey is the key for the ID of the work request that is creating the K8Instance
	K8InstanceMetadataCreatedByWorkItemKey = "createdByWorkRequestId"

	// K8InstanceMetadataCreationFailedKey is the key for indicating cluster create failure
	K8InstanceMetadataCreationFailedKey = "creationFailed"
	// K8InstanceMetadataCreationFailedTrue is the value for indicating cluster create failure
	K8InstanceMetadataCreationFailedTrue = "true"

	// K8InstanceMetadataUpdatedPoolsAtKey is the key for the update time of the pools
	K8InstanceMetadataUpdatedPoolsAtKey = "timeUpdatedPools"
	// K8InstanceMetadataUpdatedPoolsByKey is the key for the ID of the updater of the pools
	K8InstanceMetadataUpdatedPoolsByKey = "updatedPoolsByUserId"
	// K8InstanceMetadataUpdatedPoolsByWorkItemKey is the key for the ID of the work request that is updating the pools
	K8InstanceMetadataUpdatedPoolsByWorkItemKey = "updatedPoolsByWorkRequestId"
	// K8InstanceMetadataUpdateMastersDirectiveKey is the key for the directive to update cluster masters
	K8InstanceMetadataUpdateMastersDirectiveKey = "updateMastersDirective"
	// K8InstanceMetadataUpdateMastersDirectiveUpdateForce is the value for the directive to force update the cluster masters
	K8InstanceMetadataUpdateMastersDirectiveUpdateForce = "update-force"
	// K8InstanceMetadataUpdatedMastersAtKey is the key for the update time of the masters
	K8InstanceMetadataUpdatedMastersAtKey = "timeUpdatedMasters"
	// K8InstanceMetadataUpdatedMastersByKey is the key for the ID of the updater of the masters
	K8InstanceMetadataUpdatedMastersByKey = "updatedMastersByUserId"
	// K8InstanceMetadataUpdatedMastersByWorkItemKey is the key for the ID of the work request that is updating the masters
	K8InstanceMetadataUpdatedMastersByWorkItemKey = "updatedMastersByWorkRequestId"
	// K8InstanceMetadataUpgradedMastersAtKey is the key for the upgrade time of the masters
	K8InstanceMetadataUpgradedMastersAtKey = "timeUpgradedMasters"
	// K8InstanceMetadataUpgradedMastersByKey is the key for the ID of the upgrader of the masters
	K8InstanceMetadataUpgradedMastersByKey = "upgradedMastersByUserId"
	// K8InstanceMetadataUpgradedMastersByWorkItemKey is the key for the ID of the work request that is upgrading the masters
	K8InstanceMetadataUpgradedMastersByWorkItemKey = "upgradedMastersByWorkRequestId"
	// K8InstanceMetadataDeletedAtKey is the key for the deletion time of the K8Instance
	K8InstanceMetadataDeletedAtKey = "timeDeleted"
	// K8InstanceMetadataDeletedByKey is the key for the ID of the deleter of the K8Instance
	K8InstanceMetadataDeletedByKey = "deletedByUserId"
	// K8InstanceMetadataDeletedByWorkItemKey is the key for the ID of the work request that is deleting the K8Instance
	K8InstanceMetadataDeletedByWorkItemKey = "deletedByWorkRequestId"
	// K8InstanceMetadataTenancyOCIDKey is the key for the tenancy OCID for the K8Instance
	K8InstanceMetadataTenancyOCIDKey = "tenancyOCID"
	// K8InstanceMetadataTenancyNameKey is the key for the tenancy name for the K8Instance
	K8InstanceMetadataTenancyNameKey = "tenancyName"
	// K8InstanceCreatedByTMVersion is the key for the version of tenant manager that created the K8Instance
	K8InstanceCreatedByTMVersion = "creationTMVersion"
	// K8InstanceTKCVersionOnCreation is the key for the version of tenant kubernetes cluster on creation
	K8InstanceCreationTKCVersion = "creationTKCVersion"

	// K8InstanceUpdateFieldName is the affected fields key to indicate that the name needs updating
	K8InstanceUpdateFieldName = "Name"
	// K8InstanceUpdateFieldK8SVersion is the affected fields key to indicate that the K8SVersion needs updating
	K8InstanceUpdateFieldK8SVersion = "K8SVersion"
	// K8InstanceUpdateFieldUpdateForce is the affected fields key to indicate forced cluster update
	K8InstanceUpdateFieldUpdateForce = "UpdateForce"
	// K8InstanceUpdateFieldTKWVersion is the affected fields key to indicate that the TKWVersion needs updating
	K8InstanceUpdateFieldTKWVersion = "TKWVersion"

	// TKMAdminTokenKey is the key name for the TKM admin token
	TKMAdminTokenKey = "admin-token"
)

// List of supported master and worker kubernetes versions
type K8sVersionsV1 struct {
	MastersK8sVersions []string `json:"masters" yaml:"masters"`
	NodesK8sVersions   []string `json:"nodes" yaml:"nodes"`
}

// K8sVersionsRequest gets list of all supported kubernetes versions
type K8sVersionsRequestV1 struct {
}

// K8InstanceV1 stores information about an instance
type K8InstanceV1 struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	TenancyID            string            `json:"tenancyId"`
	Metadata             map[string]string `json:"metadata"`
	CloudType            string            `json:"cloudType"`
	MastersState         string            `json:"mastersState"`
	DesiredState         string            `json:"desiredState"`
	Endpoints            *K8EndpointsV1    `json:"endpoints"`
	CloudDetails         *CloudDetailsV1   `json:"cloudDetails"`
	ResourceOwnerID      string            `json:"resourceOwnerId"`
	K8Version            string            `json:"k8Version"`
	AvailableK8SUpgrades []string          `json:"availableK8sUpgrade"`
	DeletedAt            time.Time         `json:"deletedAt"`
}

type K8EndpointsV1 struct {
	Etcd string `json:"etcd"`
	K8   string `json:"k8"`
}

type CloudDetailsV1 struct {
	BMC *BMCCloudDetailsV1 `json:"bmc",omitempty"`
}

type BMCCloudDetailsV1 struct {
	LBType  string `json:"lbType"`
	LBShape string `json:"lbShape"`
}

// ToV1 converts a K8Instance object to a K8InstanceV1 object understood by the higher layers
func (src *K8Instance) ToV1() *K8InstanceV1 {
	var dst K8InstanceV1
	if src == nil {
		return &dst
	}

	dst.ID = src.ID
	dst.Name = src.Name
	dst.TenancyID = src.TenancyID
	dst.Metadata = src.Metadata
	dst.K8Version = src.K8Version
	dst.CloudDetails = &CloudDetailsV1{
		BMC: &BMCCloudDetailsV1{
			LBType: src.LBType,
			// LBShape: src.LBShape, // TODO: add to src.LBShape
		},
	}
	dst.Endpoints = &K8EndpointsV1{
		K8:   src.K8Addr,
		Etcd: src.ETCDAddress,
	}
	dst.CloudType = src.Cloud

	// determine state
	dst.MastersState = src.TKMState.Name()
	dst.AvailableK8SUpgrades = []string{}
	for _, k8Ver := range src.AvailableK8SUpgrades {
		dst.AvailableK8SUpgrades = append(dst.AvailableK8SUpgrades, k8Ver)
	}

	dst.DetermineState()

	dst.DeletedAt = protobuf.ToTime(src.DeletedAt).Truncate(time.Second)

	return &dst
}

// DetermineState generates overall cluster state
func (v1 *K8InstanceV1) DetermineState() {
}

// MastersStateExclusionsStringToSlice converts a comma separated list of
// TKMState strings to a slice of TKMState
func MastersStateExclusionsStringToSlice(param string) ([]TKMState, error) {
	states := make([]TKMState, 0)
	if len(strings.TrimSpace(param)) <= 0 {
		return states, nil
	}
	parts := strings.Split(param, ",")
	for _, stateStr := range parts {
		if state, ok := MastersState_valueV1[stateStr]; ok {
			states = append(states, state)
		} else {
			return nil, fmt.Errorf("invalid MastersStateExclusions '%s'", stateStr)
		}
	}
	return states, nil
}

// MastersState_valueV1 is a map of the V1 value for TKMState to the TKMState enum
var MastersState_valueV1 = map[string]TKMState{
	"INITIALIZING":      TKMState_Initializing,
	"RUNNING":           TKMState_Running,
	"SUCCEEDED":         TKMState_Succeeded,
	"FAILED":            TKMState_Failed,
	"UNKNOWN":           TKMState_Unknown,
	"TERMINATING":       TKMState_Terminating,
	"TERMINATED":        TKMState_Terminated,
	"UPDATING_POOLS":    TKMState_Updating_Pools,
	"UPDATING_MASTERS":  TKMState_Updating_Masters,
	"UPGRADING_MASTERS": TKMState_Upgrading_Masters,
}

// Name returns the uppercase of TKMState.String() that matches TKMState
func (x TKMState) Name() string {
	return strings.ToUpper(x.String())
}

// K8InstanceListResponseV1 are a collection that stores information about all instances
type K8InstanceListResponseV1 struct {
	Clusters []*K8InstanceV1 `json:"clusters"`
}

// ToProto converts a K8InstanceListResponseV1 to a K8InstanceListResponse object understood by grpc
func (v1 *K8InstanceListResponseV1) ToProto() *K8InstanceListResponse {
	//TODO: Write some code here when necessary
	return nil
}

func (v1 *K8InstanceListResponseV1) SetResourceOwnerID(id string) {
	for _, cluster := range v1.Clusters {
		cluster.ResourceOwnerID = id
	}
}

// ToV1 converts a K8InstanceListResponse object to a K8InstanceListResponseV1 object understood by the higher layers
func (src *K8InstanceListResponse) ToV1() K8InstanceListResponseV1 {
	v1 := K8InstanceListResponseV1{
		Clusters: make([]*K8InstanceV1, 0),
	}

	if src != nil {
		for _, item := range src.K8Instances {
			v1.Clusters = append(v1.Clusters, item.ToV1())
		}
	}
	return v1
}

// Filter filters a K8InstanceListResponse by tenantID
func (src *K8InstanceListResponse) Filter(tenancyID string) {
	filteredInstances := make([]*K8Instance, 0, len(src.K8Instances))
	for _, instance := range src.K8Instances {
		if instance.TenancyID == tenancyID {
			filteredInstances = append(filteredInstances, instance)
		}
	}
	src.K8Instances = filteredInstances
}

// K8InstanceNewRequest is the request to create a new instance
type K8InstanceNewRequestV1 struct {
	Name        string `json:"name"`
	TenancyID   string `json:"tenancyId"`
	CloudType   string `json:"cloudType"`
	LBType      string `json:"lbType"`
	K8Version   string `json:"k8Version"`
	ServiceAuth bool   `json:"serviceAuth"`
}

// ToProto converts a K8InstanceV1 to a K8Instance object understood by grpc
func (v1 *K8InstanceNewRequestV1) ToProto() *K8InstanceNewRequest {
	var dst K8InstanceNewRequest
	if v1 != nil {
		dst.Name = v1.Name
		dst.TenancyID = v1.TenancyID
		dst.Cloud = v1.CloudType
		dst.LBType = v1.LBType
		dst.K8Version = v1.K8Version
	}

	return &dst
}

// K8InstanceNewResponse is the response from creating a new instance
type K8InstanceNewResponseV1 struct {
	JobID     string `json:"workItemId"`
	ClusterID string `json:"clusterId"`
}

// ToV1 converts a K8InstanceNewResponse object to a K8InstanceNewResponse object understood by the higher layers
func (src *K8InstanceNewResponse) ToV1() K8InstanceNewResponseV1 {
	var dst K8InstanceNewResponseV1
	if src == nil {
		return dst
	}
	if src != nil {
		dst.JobID = src.JobID
		dst.ClusterID = src.ClusterID
	}

	return dst
}

type K8InstanceUpdateTKMRequestV1 struct {
	ClusterID  string `json:"clusterId"`
	K8SVersion string `json:"k8sVersion"`
}

func (r K8InstanceUpdateTKMRequestV1) Validate(currentVersion string, f func() *K8sVersionsV1) error {
	if currentVersion == r.K8SVersion {
		return fmt.Errorf("current k8s version is already the same as the requested version")
	}

	var found bool
	versions := f().MastersK8sVersions
	for _, v := range versions {
		if r.K8SVersion == v {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("invalid k8sversion use one of: %q", versions)
	}

	current, err := getK8SSemverNumbers(currentVersion)
	if err != nil {
		return err
	}
	newer, err := getK8SSemverNumbers(r.K8SVersion)
	if err != nil {
		return err
	}

	for i, j := 0, 0; i < len(current) && j < len(newer); i, j = i+1, j+1 {
		if current[i] > newer[j] {
			return fmt.Errorf("current k8s version %s is already newer than requested version %s", currentVersion, r.K8SVersion)
		} else if current[i] < newer[j] {
			return nil
		}
	}

	return nil
}

// K8InstanceUpdateTKMResponseV1 contains the job response when
// issuing a tkm update or tkm upgrade
type K8InstanceUpdateTKMResponseV1 struct {
	ClusterID string `json:"clusterId"`
	JobID     string `json:"workItemId"`
	Err       error  `json:"message,omitempty"`
}

// ToV1 converts proto type to v1 K8InstanceUpdateTKMResponseV1
func (src *K8InstanceUpdateResponse) ToV1() K8InstanceUpdateTKMResponseV1 {
	var dst K8InstanceUpdateTKMResponseV1
	if src == nil {
		return dst
	}
	dst.ClusterID = src.ClusterID
	dst.JobID = src.JobID
	return dst
}

// K8InstanceResponseV1 are a collection that stores information about all instances
type K8InstanceResponseV1 struct {
	K8Instance      *K8InstanceV1 `json:"k8"`
	Update          string        `json:"update,omitempty"`
	UpdateAvailable bool          `json:"updateAvailable"`
}

// ToProto converts a K8InstanceResponseV1 to a K8InstanceResponse object understood by grpc
func (v1 *K8InstanceResponseV1) ToProto() *K8InstanceResponse {
	//TODO: Write some code here when necessary
	return nil
}

func (v1 *K8InstanceResponseV1) SetResourceOwnerID(id string) {
	v1.K8Instance.ResourceOwnerID = id
}

// ToV1 converts a K8InstanceResponse object to a K8InstanceResponseV1 object understood by the higher layers
func (src *K8InstanceResponse) ToV1() K8InstanceResponseV1 {
	dst := K8InstanceResponseV1{}

	if src != nil {
		dst.K8Instance = src.K8Instance.ToV1()
	}
	return dst
}

// K8InstanceDeleteResponse is the response from creating a new instance
type K8InstanceDeleteResponseV1 struct {
	JobID     string `json:"workItemId"`
	ClusterID string `json:"clusterId"`
}

// ToV1 converts a K8InstanceDeleteResponse object to a K8InstanceDeleteResponse object understood by the higher layers
func (src *K8InstanceDeleteResponse) ToV1() K8InstanceDeleteResponseV1 {
	var dst K8InstanceDeleteResponseV1
	if src == nil {
		return dst
	}
	if src != nil {
		dst.JobID = src.JobID
		dst.ClusterID = src.ClusterID
	}

	return dst
}

// K8InstanceDeleteJobInfo is the payload passed to the job to delete a K8instance
type K8InstanceDeleteJobInfo struct {
	Request    K8InstanceDeleteRequest
	K8Instance K8Instance
}

type KubernetesVersionsGetter interface {
	GetMasterKubernetesVersions(ctx context.Context, tenancyID string) ([]string, error)
	GetWorkerKubernetesVersions(ctx context.Context, tenancyID string) ([]string, error)
}

// GetSupportedK8sVersions returns lists of supported master and worker kubernetes versions
func GetSupportedK8sVersions(ctx context.Context, tenancyID string, versionsGetter KubernetesVersionsGetter) (*K8sVersionsV1, error) {
	masterVersions, err := versionsGetter.GetMasterKubernetesVersions(ctx, tenancyID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting master versions")
	}
	workerVersions, err := versionsGetter.GetWorkerKubernetesVersions(ctx, tenancyID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting worker versions")
	}

	return &K8sVersionsV1{
		MastersK8sVersions: masterVersions,
		NodesK8sVersions:   workerVersions,
	}, nil
}

func validateMasterK8Version(ctx context.Context, tenancyID string, versionsGetter KubernetesVersionsGetter, v string) error {
	versions, err := GetSupportedK8sVersions(ctx, tenancyID, versionsGetter)
	if err != nil {
		return errors.Wrap(err, "unable to validate master kubernetes version")
	}
	if err := validateK8Version(v, versions.MastersK8sVersions); err != nil {
		return errors.Wrap(err, "unsupported master kubernetes version")
	}
	return nil
}

func validateWorkerK8Version(ctx context.Context, tenancyID string, versionsGetter KubernetesVersionsGetter, pool, v string) error {
	versions, err := GetSupportedK8sVersions(ctx, tenancyID, versionsGetter)
	if err != nil {
		return errors.Wrap(err, "unable to validate worker kubernetes version")
	}
	if err := validateK8Version(v, versions.NodesK8sVersions); err != nil {
		return errors.Wrapf(err, "unsupported kubernetes version on pool '%s'", pool)
	}
	return nil
}

func validateK8Version(v string, versions []string) error {
	for _, version := range versions {
		if v == version {
			return nil
		}
	}
	return fmt.Errorf("Unknown version, %s not one of %s", v, strings.Join(versions, ", "))
}
