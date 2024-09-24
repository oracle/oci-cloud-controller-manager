// Copyright 2020 Oracle and/or its affiliates. All rights reserved.
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

package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	kubeAPI "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/disk"
	"github.com/oracle/oci-go-sdk/v65/core"
)

const (
	// Prefix to apply to the name of a created volume. This should be the same as the option '--volume-name-prefix' of csi-provisioner.
	pvcPrefix                     = "csi"
	csiDriver                     = "csi"
	fsTypeKey                     = "csi.storage.k8s.io/fstype"
	fsTypeKeyDeprecated           = "fstype"
	timeout                       = time.Minute * 3
	kmsKey                        = "kms-key-id"
	attachmentType                = "attachment-type"
	attachmentTypeISCSI           = "iscsi"
	attachmentTypeParavirtualized = "paravirtualized"
	initialFreeformTagsOverride   = "oci.oraclecloud.com/initial-freeform-tags-override"
	initialDefinedTagsOverride    = "oci.oraclecloud.com/initial-defined-tags-override"
	backupType                    = "backupType"
	backupTypeFull                = "full"
	backupTypeIncremental         = "incremental"
	backupDefinedTags             = "oci.oraclecloud.com/defined-tags"
	backupFreeformTags            = "oci.oraclecloud.com/freeform-tags"
	newBackupAvailableTimeout     = 45 * time.Second
	needResize                    = "needResize"
	newSize                       = "newSize"
	multipathEnabled              = "multipathEnabled"
	multipathDevices              = "multipathDevices"
	//device is the consistent device path that would be used for paravirtualized attachment
	device                          = "device"
	resourceTrackingFeatureFlagName = "CPO_ENABLE_RESOURCE_ATTRIBUTION"
	OkeSystemTagNamesapce           = "orcl-containerengine"
	MaxDefinedTagPerVolume          = 64
)

var (
	// OCI currently only support a single node to be attached to a single node
	// in read/write mode. This corresponds to `accessModes.ReadWriteOnce` in a
	// PVC resource on Kubernetes
	supportedAccessModes = []*csi.VolumeCapability_AccessMode{
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		},
	}
)

var enableOkeSystemTags = csi_util.GetIsFeatureEnabledFromEnv(zap.S(), resourceTrackingFeatureFlagName, false)

// VolumeParameters holds configuration
type VolumeParameters struct {
	//kmsKey is the KMS key that would be used as CMEK key for BV attachment
	diskEncryptionKey   string
	attachmentParameter map[string]string
	// freeform tags to add for BVs
	freeformTags map[string]string
	// defined tags to add for BVs
	definedTags map[string]map[string]interface{}
	//volume performance units per gb describes the block volume performance level
	vpusPerGB int64
}

// VolumeAttachmentOption holds config for attachments
type VolumeAttachmentOption struct {
	//whether the attachment type is paravirtualized
	useParavirtualizedAttachment bool
	//whether to encrypt the compute to BV attachment as in-transit encryption.
	enableInTransitEncryption bool
}

type SnapshotParameters struct {
	//backupType is the parameter which is used to decide if the backup created will be FULL or INCREMENTAL
	backupType core.CreateVolumeBackupDetailsTypeEnum
	// freeform tags to add for backups
	freeformTags map[string]string
	// defined tags to add for backups
	definedTags map[string]map[string]interface{}
}

func extractVolumeParameters(log *zap.SugaredLogger, parameters map[string]string) (VolumeParameters, error) {
	p := VolumeParameters{
		diskEncryptionKey:   "",
		attachmentParameter: make(map[string]string),
		vpusPerGB:           10, // default value is 10 -> Balanced
	}
	for k, v := range parameters {
		switch k {
		case fsTypeKeyDeprecated:
			log.Warnf("%s is deprecated, please use %s instead", fsTypeKeyDeprecated, fsTypeKey)
		case kmsKey:
			if v != "" {
				p.diskEncryptionKey = v
			}
		case attachmentType:
			attachmentTypeLower := strings.ToLower(v)
			if attachmentTypeLower != attachmentTypeISCSI && attachmentTypeLower != attachmentTypeParavirtualized {
				return p, status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid attachment-type: %s provided "+
					"for storageclass. supported attachment-types are %s and %s", v, attachmentTypeISCSI, attachmentTypeParavirtualized))
			}
			p.attachmentParameter[attachmentType] = attachmentTypeLower

		case initialFreeformTagsOverride:
			if v == "" {
				continue
			}
			freeformTags := make(map[string]string)
			err := json.Unmarshal([]byte(v), &freeformTags)
			if err != nil {
				return p, status.Errorf(codes.InvalidArgument, "failed to parse freeform tags provided "+
					"for storageclass. please check the parameters block on the storage class")
			}

			p.freeformTags = freeformTags
		case initialDefinedTagsOverride:
			if v == "" {
				continue
			}
			definedTags := make(map[string]map[string]interface{})
			err := json.Unmarshal([]byte(v), &definedTags)
			if err != nil {
				return p, status.Errorf(codes.InvalidArgument, "failed to parse defined tags provided "+
					"for storageclass. please check the parameters block on the storage class")
			}
			p.definedTags = definedTags

		case csi_util.VpusPerGB:
			vpusPerGB, err := csi_util.ExtractBlockVolumePerformanceLevel(v)
			if err != nil {
				return p, status.Error(codes.InvalidArgument, err.Error())
			}
			p.vpusPerGB = vpusPerGB
		}

	}
	return p, nil
}

func extractSnapshotParameters(parameters map[string]string) (SnapshotParameters, error) {
	p := SnapshotParameters{
		backupType: core.CreateVolumeBackupDetailsTypeIncremental, //Default backupType is incremental
	}
	for k, v := range parameters {
		switch k {
		case backupType:
			backupTypeLower := strings.ToLower(v)
			if backupTypeLower == backupTypeIncremental {
				p.backupType = core.CreateVolumeBackupDetailsTypeIncremental
			} else if backupTypeLower == backupTypeFull {
				p.backupType = core.CreateVolumeBackupDetailsTypeFull
			} else {
				return p, status.Errorf(codes.InvalidArgument, fmt.Sprintf("invalid backupType: %s provided "+
					"in volumesnapshotclass. supported backupType are %s and %s", v, backupTypeIncremental, backupTypeFull))
			}
		case backupFreeformTags:
			if v == "" {
				continue
			}
			freeformTags := make(map[string]string)
			err := json.Unmarshal([]byte(v), &freeformTags)
			if err != nil {
				return p, status.Errorf(codes.InvalidArgument, "failed to parse freeform tags provided "+
					"in volumesnapshotclass. please check the parameters block on the volume snapshot class")
			}
			p.freeformTags = freeformTags
		case backupDefinedTags:
			if v == "" {
				continue
			}
			definedTags := make(map[string]map[string]interface{})
			err := json.Unmarshal([]byte(v), &definedTags)
			if err != nil {
				return p, status.Errorf(codes.InvalidArgument, "failed to parse defined tags provided "+
					"in volumesnapshotclass. please check the parameters block on the volume snapshot class")
			}
			p.definedTags = definedTags
		}
	}
	return p, nil
}

// CreateVolume creates a new volume from the given request. The function is
// idempotent.
func (d *BlockVolumeControllerDriver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	startTime := time.Now()
	log := d.logger.With("volumeName", req.Name, "csiOperation", "create")
	var errorType string
	var metricDimension string
	volumeContext := map[string]string{
		needResize: "false",
		newSize:    "",
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume Name must be provided")
	}

	if req.VolumeCapabilities == nil || len(req.VolumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapabilities must be provided in CreateVolumeRequest")
	}

	if err := d.validateCapabilities(req.VolumeCapabilities); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	size, err := csi_util.ExtractStorage(req.CapacityRange)
	if err != nil {
		return nil, status.Errorf(codes.OutOfRange, "invalid capacity range: %v", err)
	}

	availableDomainShortName := ""
	volumeName := req.Name

	dimensionsMap := make(map[string]string)

	volumeParams, err := extractVolumeParameters(log, req.GetParameters())
	if err != nil {
		log.With(zap.Error(err)).Error("Failed to parse storageclass parameters.")
		metricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = metricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse storageclass parameters %v", err)
	}

	dimensionsMap[metrics.ResourceOCIDDimension] = volumeName
	dimensionsMap[metrics.VolumeVpusPerGBDimension] = strconv.Itoa(int(volumeParams.vpusPerGB))

	srcSnapshotId := ""
	srcVolumeId := ""
	volumeContentSource := req.GetVolumeContentSource()
	if volumeContentSource != nil {
		_, isVolumeContentSource_Snapshot := volumeContentSource.GetType().(*csi.VolumeContentSource_Snapshot)
		_, isVolumeContentSource_Volume := volumeContentSource.GetType().(*csi.VolumeContentSource_Volume)

		if !isVolumeContentSource_Snapshot && !isVolumeContentSource_Volume {
			log.Error("Unsupported volumeContentSource")
			return nil, status.Error(codes.InvalidArgument, "Unsupported volumeContentSource")
		}

		if isVolumeContentSource_Snapshot {
			srcSnapshot := volumeContentSource.GetSnapshot()
			if srcSnapshot == nil {
				log.With("volumeSourceType", "snapshot").Error("Error fetching snapshot from the volumeContentSource")
				return nil, status.Error(codes.InvalidArgument, "Error fetching snapshot from the volumeContentSource")
			}

			id := srcSnapshot.GetSnapshotId()
			volumeBackup, err := d.client.BlockStorage().GetVolumeBackup(ctx, id)
			if err != nil {
				if k8sapierrors.IsNotFound(err) {
					log.With("service", "blockstorage", "verb", "get", "resource", "volumeBackup", "statusCode", util.GetHttpStatusCode(err)).Errorf("Failed to get snapshot with ID %v", id)
					return nil, status.Errorf(codes.NotFound, "Failed to get snapshot with ID %v", id)
				}
				log.With("service", "blockstorage", "verb", "get", "resource", "volumeBackup", "statusCode", util.GetHttpStatusCode(err)).Errorf("Failed to fetch snapshot with ID %v with error %v", id, err)
				return nil, status.Errorf(codes.Internal, "Failed to fetch snapshot with ID %v with error %v", id, err)
			}

			volumeBackupSize := *volumeBackup.SizeInMBs * client.MiB
			if volumeBackupSize < size {
				volumeContext[needResize] = "true"
				volumeContext[newSize] = strconv.FormatInt(size, 10)
			}

			srcSnapshotId = id
		} else {
			srcVolume := volumeContentSource.GetVolume()
			if srcVolume == nil {
				log.With("volumeSourceType", "pvc").Error("Error fetching volume from the volumeContentSource")
				return nil, status.Error(codes.InvalidArgument, "Error fetching volume from the volumeContentSource")
			}

			id := srcVolume.GetVolumeId()
			srcBlockVolume, err := d.client.BlockStorage().GetVolume(ctx, id)
			if err != nil {
				if client.IsNotFound(err) {
					log.With("service", "blockstorage", "verb", "get", "resource", "blockVolume", "statusCode", util.GetHttpStatusCode(err)).Errorf("Failed to get volume with ID %v", id)
					return nil, status.Errorf(codes.NotFound, "Failed to get volume with ID %v", id)
				}
				log.With("service", "blockstorage", "verb", "get", "resource", "blockVolume", "statusCode", util.GetHttpStatusCode(err)).Errorf("Failed to fetch volume with ID %v with error %v", id, err)
				return nil, status.Errorf(codes.Internal, "Failed to fetch volume with ID %v with error %v", id, err)
			}

			availableDomainShortName = *srcBlockVolume.AvailabilityDomain
			log.With("AD", availableDomainShortName).Info("Using availability domain of source volume to provision clone volume.")

			srcBlockVolumeSize := *srcBlockVolume.SizeInMBs * client.MiB
			if srcBlockVolumeSize < size {
				volumeContext["needResize"] = "true"
				volumeContext["newSize"] = strconv.FormatInt(size, 10)
			}

			srcVolumeId = id
		}
	}

	if req.AccessibilityRequirements != nil && req.AccessibilityRequirements.Preferred != nil && availableDomainShortName == "" {
		for _, t := range req.AccessibilityRequirements.Preferred {
			var ok bool
			availableDomainShortName, ok = t.Segments[kubeAPI.LabelTopologyZone]
			if !ok {
				availableDomainShortName, _ = t.Segments[kubeAPI.LabelZoneFailureDomain]
			}
			log.With("AD", availableDomainShortName).Info("Using preferred topology for AD.")
			if len(availableDomainShortName) > 0 {
				break
			}
		}
	}

	if availableDomainShortName == "" {
		if req.AccessibilityRequirements != nil && req.AccessibilityRequirements.Requisite != nil {
			for _, t := range req.AccessibilityRequirements.Requisite {
				var ok bool
				availableDomainShortName, ok = t.Segments[kubeAPI.LabelTopologyZone]
				if !ok {
					availableDomainShortName, _ = t.Segments[kubeAPI.LabelZoneFailureDomain]
				}
				log.With("AD", availableDomainShortName).Info("Using requisite topology for AD.")
				if len(availableDomainShortName) > 0 {
					break
				}
			}
		}
	}

	metric := metrics.PVProvision
	metricType := util.CSIStorageType
	if srcSnapshotId != "" {
		metric = metrics.BlockSnapshotRestore
		metricType = util.CSIStorageType
	}
	if srcVolumeId != "" {
		metric = metrics.PVClone
		metricType = util.CSIStorageType
	}

	if availableDomainShortName == "" {
		metricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, metricType)
		dimensionsMap[metrics.ComponentDimension] = metricDimension
		metrics.SendMetricData(d.metricPusher, metric, time.Since(startTime).Seconds(), dimensionsMap)
		log.Error("Available domain short name is not found")
		return nil, status.Errorf(codes.InvalidArgument, "(%s) or (%s) is required in PreferredTopologies or allowedTopologies", kubeAPI.LabelTopologyZone, kubeAPI.LabelZoneFailureDomain)
	}

	//make sure this method is idempotent by checking existence of volume with same name.
	volumes, err := d.client.BlockStorage().GetVolumesByName(ctx, volumeName, d.config.CompartmentID)
	if err != nil {
		log.With("service", "blockstorage", "verb", "get", "resource", "volume", "statusCode", util.GetHttpStatusCode(err)).
			With(zap.Error(err)).Error("Failed to find existence of volume.")
		errorType = util.GetError(err)
		metricDimension = util.GetMetricDimensionForComponent(errorType, metricType)
		dimensionsMap[metrics.ComponentDimension] = metricDimension
		metrics.SendMetricData(d.metricPusher, metric, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Internal, "failed to check existence of volume %v", err)
	}

	if len(volumes) > 1 {
		log.Error("Duplicate volume exists")
		metricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, metricType)
		dimensionsMap[metrics.ComponentDimension] = metricDimension
		metrics.SendMetricData(d.metricPusher, metric, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, fmt.Errorf("duplicate volume %q exists", volumeName)
	}

	provisionedVolume := core.Volume{}

	if len(volumes) > 0 {
		//Volume already exists so checking state of the volume and returning the same.
		log.Info("Volume already created!")
		//Assigning existing volume
		provisionedVolume = volumes[0]

	} else {
		// Creating new volume
		ad, err := d.client.Identity().GetAvailabilityDomainByName(ctx, d.config.CompartmentID, availableDomainShortName)
		if err != nil {
			log.With("Compartment Id", d.config.CompartmentID, "service", "identity", "verb", "get", "resource", "AD", "statusCode", util.GetHttpStatusCode(err)).
				With(zap.Error(err)).Error("Failed to get available domain.")
			errorType = util.GetError(err)
			metricDimension = util.GetMetricDimensionForComponent(errorType, metricType)
			dimensionsMap[metrics.ComponentDimension] = metricDimension
			metrics.SendMetricData(d.metricPusher, metric, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.InvalidArgument, "invalid available domain: %s or compartment ID: %s", availableDomainShortName, d.config.CompartmentID)
		}

		bvTags := getBVTags(log, d.config.Tags, volumeParams)

		provisionedVolume, err = provision(ctx, log, d.client, volumeName, size, *ad.Name, d.config.CompartmentID, srcSnapshotId, srcVolumeId,
			volumeParams.diskEncryptionKey, volumeParams.vpusPerGB, bvTags)

		if err != nil && client.IsSystemTagNotFoundOrNotAuthorisedError(log, errors.Unwrap(err)) {
			log.With("Ad name", *ad.Name, "Compartment Id", d.config.CompartmentID).With(zap.Error(err)).Warn("New volume creation failed due to oke system tags error. sending metric & retrying without oke system tags")
			errorType = util.SystemTagErrTypePrefix + util.GetError(err)
			metricDimension = util.GetMetricDimensionForComponent(errorType, metricType)
			dimensionsMap[metrics.ComponentDimension] = metricDimension
			metrics.SendMetricData(d.metricPusher, metric, time.Since(startTime).Seconds(), dimensionsMap)

			// retry provision without oke system tags
			delete(bvTags.DefinedTags, OkeSystemTagNamesapce)
			provisionedVolume, err = provision(ctx, log, d.client, volumeName, size, *ad.Name, d.config.CompartmentID, srcSnapshotId, srcVolumeId,
				volumeParams.diskEncryptionKey, volumeParams.vpusPerGB, bvTags)
		}
		if err != nil {
			log.With("Ad name", *ad.Name, "Compartment Id", d.config.CompartmentID).With(zap.Error(err)).Error("New volume creation failed.")
			errorType = util.GetError(err)
			metricDimension = util.GetMetricDimensionForComponent(errorType, metricType)
			dimensionsMap[metrics.ComponentDimension] = metricDimension
			metrics.SendMetricData(d.metricPusher, metric, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.Internal, "New volume creation failed %v", err.Error())
		}
	}
	log.Info("Waiting for volume to become available.")

	if srcVolumeId != "" {
		_, err = d.client.BlockStorage().AwaitVolumeCloneAvailableOrTimeout(ctx, *provisionedVolume.Id)
	} else {
		_, err = d.client.BlockStorage().AwaitVolumeAvailableORTimeout(ctx, *provisionedVolume.Id)
	}
	if err != nil {
		log.With("service", "blockstorage", "verb", "get", "resource", "volume", "statusCode", util.GetHttpStatusCode(err)).
			With("volumeName", volumeName).Error("Create volume failed with time out")
		errorType = util.GetError(err)
		metricDimension = util.GetMetricDimensionForComponent(errorType, metricType)
		dimensionsMap[metrics.ComponentDimension] = metricDimension
		metrics.SendMetricData(d.metricPusher, metric, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.DeadlineExceeded, "Create volume failed with time out %v", err.Error())
	}

	volumeOCID := volumeName
	if provisionedVolume.Id != nil {
		volumeOCID = *provisionedVolume.Id
	}
	log.With("volumeID", volumeOCID).Info("Volume is created.")
	metricDimension = util.GetMetricDimensionForComponent(util.Success, metricType)
	dimensionsMap[metrics.ComponentDimension] = metricDimension
	dimensionsMap[metrics.ResourceOCIDDimension] = volumeOCID
	metrics.SendMetricData(d.metricPusher, metric, time.Since(startTime).Seconds(), dimensionsMap)

	volumeContext[attachmentType] = volumeParams.attachmentParameter[attachmentType]
	volumeContext[csi_util.VpusPerGB] = strconv.FormatInt(volumeParams.vpusPerGB, 10)

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      *provisionedVolume.Id,
			CapacityBytes: *provisionedVolume.SizeInMBs * client.MiB,
			AccessibleTopology: []*csi.Topology{
				{
					Segments: map[string]string{
						kubeAPI.LabelTopologyZone: d.util.GetAvailableDomainInNodeLabel(*provisionedVolume.AvailabilityDomain),
					},
				},
				{
					Segments: map[string]string{
						kubeAPI.LabelZoneFailureDomain: d.util.GetAvailableDomainInNodeLabel(*provisionedVolume.AvailabilityDomain),
					},
				},
			},

			VolumeContext: volumeContext,
			ContentSource: volumeContentSource,
		},
	}, nil
}

// DeleteVolume deletes the given volume. The function is idempotent.
func (d *BlockVolumeControllerDriver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	startTime := time.Now()
	log := d.logger.With("volumeID", req.VolumeId, "csiOperation", "delete")
	var errorType string
	var csiMetricDimension string
	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.VolumeId

	if req.VolumeId == "" {
		log.Info("Unable to get Volume Id")
		csiMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDelete, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Error(codes.InvalidArgument, "DeleteVolume Volume ID must be provided")
	}

	log.Info("Deleting Volume")
	err := d.client.BlockStorage().DeleteVolume(ctx, req.VolumeId)
	if err != nil {
		if !client.IsNotFound(err) {
			log.With("service", "blockstorage", "verb", "delete", "resource", "volume", "statusCode", util.GetHttpStatusCode(err)).With(zap.Error(err)).Error("Failed to delete volume.")
			errorType = util.GetError(err)
			csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVDelete, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, fmt.Errorf("failed to delete volume, volumeId: %s, error: %v", req.VolumeId, err)
		}
		log.With("service", "blockstorage", "verb", "delete", "resource", "volume", "statusCode", util.GetHttpStatusCode(err)).With(zap.Error(err)).
			Error("Unable to find volume to delete. Volume is possibly already deleted. No Delete Operation required.")
	} else {
		log.Info("Volume is deleted.")
	}
	csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.PVDelete, time.Since(startTime).Seconds(), dimensionsMap)
	return &csi.DeleteVolumeResponse{}, nil
}

// ControllerPublishVolume attaches the given volume to the node
func (d *BlockVolumeControllerDriver) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	startTime := time.Now()
	var errorType string
	var csiMetricDimension string

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Node ID must be provided")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability must be provided")
	}

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.VolumeId

	log := d.logger.With("volumeID", req.VolumeId, "nodeId", req.NodeId, "csiOperation", "attach")

	id, err := d.util.LookupNodeID(d.KubeClient, req.NodeId)
	if err != nil {
		log.With(zap.Error(err)).Error("Failed to lookup node")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.InvalidArgument, "failed to get ProviderID by nodeName. error : %s", err)
	}
	id = client.MapProviderIDToResourceID(id)

	//if the attachmentType is missing, default is iscsi
	attachType, ok := req.VolumeContext[attachmentType]
	if !ok {
		attachType = attachmentTypeISCSI
	}

	// Check if the access mode is ReadWriteMany and set isShareable to true
	isSharable := false
	if req.VolumeCapability.AccessMode != nil {
		mode := req.VolumeCapability.AccessMode.Mode
		if mode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY ||
			mode == csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER ||
			mode == csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER {
			isSharable = true
		}
	}

	volumeAttachmentOptions, err := getAttachmentOptions(ctx, d.client.Compute(), attachType, id)
	if err != nil {
		log.With("service", "compute", "verb", "get", "resource", "instance", "statusCode", util.GetHttpStatusCode(err)).
			With(zap.Error(err)).With("attachmentType", attachType, "instanceID", id).Error("failed to get the attachment options")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Unknown, "failed to get the attachment options. error : %s", err)
	}
	//in transit encryption is not supported for other attachment type than paravirtualized
	if volumeAttachmentOptions.enableInTransitEncryption && !volumeAttachmentOptions.useParavirtualizedAttachment {
		log.Errorf("node %s has in transit encryption enabled, but attachment type is not paravirtualized. invalid input", id)
		csiMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.InvalidArgument, "node %s has in transit encryption enabled, but attachment type is not paravirtualized. invalid input", id)
	}

	compartmentID, err := util.LookupNodeCompartment(d.KubeClient, req.NodeId)
	if err != nil {
		log.With(zap.Error(err)).With("instanceID", id).Errorf("failed to get compartmentID from node annotation: %s", util.CompartmentIDAnnotation)
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Unknown, "failed to get compartmentID from node annotation:. error : %s", err)
	}

	volumeAttached, err := d.client.Compute().FindActiveVolumeAttachment(ctx, compartmentID, req.VolumeId)

	if err != nil && !client.IsNotFound(err) {
		log.With("service", "compute", "verb", "get", "resource", "volumeAttachment", "statusCode", util.GetHttpStatusCode(err)).
			With(zap.Error(err)).Error("Got error in finding volume attachment.")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, err
	}

	vpusPerGB, ok := req.VolumeContext[csi_util.VpusPerGB]
	if !ok || vpusPerGB == "" {
		log.Warnf("No vpusPerGB found in Volume Context falling back to balanced performance")
		vpusPerGB = "10"
	}

	// volume already attached to an instance
	if err == nil && !isSharable {
		log = log.With("volumeAttachedId", *volumeAttached.GetId())
		if volumeAttached.GetLifecycleState() == core.VolumeAttachmentLifecycleStateDetaching {
			log.With("instanceID", *volumeAttached.GetInstanceId()).Info("Waiting for volume to get detached before attaching.")
			err = d.client.Compute().WaitForVolumeDetached(ctx, *volumeAttached.GetId())
			if err != nil {
				log.With("service", "compute", "verb", "get", "resource", "volumeAttachment", "statusCode", util.GetHttpStatusCode(err)).
					With("instanceID", *volumeAttached.GetInstanceId()).With(zap.Error(err)).Error("Error while waiting for volume to get detached before attaching.")
				errorType = util.GetError(err)
				csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
				dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
				metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
				return nil, status.Errorf(codes.Internal, "Error while waiting for volume to get detached before attaching: %s", err)
			}
		} else {
			if id != *volumeAttached.GetInstanceId() {
				log.Errorf("Volume is already attached to another node: %s", *volumeAttached.GetInstanceId())
				csiMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
				dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
				metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
				return nil, status.Errorf(codes.Internal, "Failed to attach volume to node. "+
					"The volume is already attached to another node.")
			}
			if volumeAttached.GetLifecycleState() == core.VolumeAttachmentLifecycleStateAttaching {
				log.With("instanceID", id).Info("Volume is in ATTACHING state. Waiting for Volume to attach to the Node.")
				volumeAttached, err = d.client.Compute().WaitForVolumeAttached(ctx, *volumeAttached.GetId())
				if err != nil {
					log.With("service", "compute", "verb", "get", "resource", "volumeAttachment", "statusCode", util.GetHttpStatusCode(err)).
						With("instanceID", id).With(zap.Error(err)).Error("Error while waiting: failed to attach volume to the node.")
					errorType = util.GetError(err)
					csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
					dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
					metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
					return nil, status.Errorf(codes.Internal, "Failed to attach volume to the node: %s", err)
				}
			}
			//Checking if Volume state is already Attached or Attachment (from above condition) is completed
			if volumeAttached.GetLifecycleState() == core.VolumeAttachmentLifecycleStateAttached {
				log.With("instanceID", id).Info("Volume is already ATTACHED to the Node.")
				resp, err := generatePublishContext(volumeAttachmentOptions, log, volumeAttached, vpusPerGB, req.VolumeContext[needResize], req.VolumeContext[newSize])
				if err != nil {
					log.With(zap.Error(err)).Error("Failed to generate publish context")
					errorType = util.GetError(err)
					csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
					dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
					metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
					return nil, status.Errorf(codes.Internal, "Failed to generate publish context: %s", err)
				}
				return resp, nil
			}
		}
	}

	log.Info("Attaching volume to instance")

	if volumeAttachmentOptions.useParavirtualizedAttachment {
		volumeAttached, err = d.client.Compute().AttachParavirtualizedVolume(ctx, id, req.VolumeId, volumeAttachmentOptions.enableInTransitEncryption, isSharable)
		if err != nil {
			log.With("service", "compute", "verb", "create", "resource", "volumeAttachment", "statusCode", util.GetHttpStatusCode(err)).
				With("instanceID", id).With(zap.Error(err)).Info("failed paravirtualized attachment instance to volume.")
			errorType = util.GetError(err)
			csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.Internal, "failed paravirtualized attachment instance to volume. error : %s", err)
		}
	} else {
		volumeAttached, err = d.client.Compute().AttachVolume(ctx, id, req.VolumeId, isSharable)
		if err != nil {
			log.With("service", "compute", "verb", "create", "resource", "volumeAttachment", "statusCode", util.GetHttpStatusCode(err)).
				With("instanceID", id).With(zap.Error(err)).Info("failed iscsi attachment instance to volume.")
			errorType = util.GetError(err)
			csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.Internal, "failed iscsi attachment instance to volume : %s", err)
		}
	}

	volumeAttached, err = d.client.Compute().WaitForVolumeAttached(ctx, *volumeAttached.GetId())
	if err != nil {
		log.With("service", "compute", "verb", "get", "resource", "volumeAttachment", "statusCode", util.GetHttpStatusCode(err)).
			With("instanceID", id).With(zap.Error(err)).Error("Failed to attach volume to the node.")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Internal, "Failed to attach volume to the node %s", err)
	}
	log.Info("Volume is ATTACHED to Node.")
	csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
	resp, err := generatePublishContext(volumeAttachmentOptions, log, volumeAttached, vpusPerGB, req.VolumeContext[needResize], req.VolumeContext[newSize])
	if err != nil {
		log.With(zap.Error(err)).Error("Failed to generate publish context")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Internal, "Failed to generate publish context: %s", err)
	}
	return resp, nil
}

func generatePublishContext(volumeAttachmentOptions VolumeAttachmentOption, log *zap.SugaredLogger, volumeAttached core.VolumeAttachment, vpusPerGB string, needsResize string, expectedSize string) (*csi.ControllerPublishVolumeResponse, error) {
	multipath := "false"

	if volumeAttached.GetIsMultipath() != nil {
		multipath = strconv.FormatBool(*volumeAttached.GetIsMultipath())
	}

	if volumeAttachmentOptions.useParavirtualizedAttachment {
		log.With("volumeAttachedId", *volumeAttached.GetId()).Info("Publishing paravirtualized Volume Completed.")
		return &csi.ControllerPublishVolumeResponse{
			PublishContext: map[string]string{
				attachmentType:     attachmentTypeParavirtualized,
				device:             *volumeAttached.GetDevice(),
				csi_util.VpusPerGB: vpusPerGB,
				needResize:         needsResize,
				newSize:            expectedSize,
				multipathEnabled:   multipath,
			},
		}, nil
	}
	iSCSIVolumeAttached := volumeAttached.(core.IScsiVolumeAttachment)
	multiPathDevicesJson := []byte{}
	if len(iSCSIVolumeAttached.MultipathDevices) > 0 {
		var err error
		multiPathDevicesJson, err = json.Marshal(iSCSIVolumeAttached.MultipathDevices)
		if err != nil {
			return nil, err
		}
	}

	log.With("volumeAttachedId", *volumeAttached.GetId()).Info("Publishing iSCSI Volume Completed.")

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			attachmentType:     attachmentTypeISCSI,
			device:             *volumeAttached.GetDevice(),
			disk.ISCSIIQN:      *iSCSIVolumeAttached.Iqn,
			disk.ISCSIIP:       *iSCSIVolumeAttached.Ipv4,
			disk.ISCSIPORT:     strconv.Itoa(*iSCSIVolumeAttached.Port),
			csi_util.VpusPerGB: vpusPerGB,
			needResize:         needsResize,
			newSize:            expectedSize,
			multipathEnabled:   multipath,
			multipathDevices:   string(multiPathDevicesJson),
		},
	}, nil
}

// ControllerUnpublishVolume detaches the given volume from the node
func (d *BlockVolumeControllerDriver) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	startTime := time.Now()
	log := d.logger.With("volumeID", req.VolumeId, "nodeId", req.NodeId, "csiOperation", "detach")
	var errorType string
	var csiMetricDimension string

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.VolumeId

	compartmentID, err := util.LookupNodeCompartment(d.KubeClient, req.NodeId)

	if err != nil {
		if k8sapierrors.IsNotFound(err) {
			log.Infof("Node with nodeID %s is not found, volume is likely already detached", req.NodeId)
			csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}
		log.With(zap.Error(err)).Errorf("failed to get compartmentID from node annotation: %s", util.CompartmentIDAnnotation)
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Unknown, "failed to get compartmentID from node annotation:: error : %s", err)
	}

	log = log.With("compartmentID", compartmentID)

	instanceID, err := d.util.LookupNodeID(d.KubeClient, req.NodeId)

	if err != nil {
		log.With(zap.Error(err)).Errorf("failed to get instanceID from node : %s", req.NodeId)
		return &csi.ControllerUnpublishVolumeResponse{}, nil
	}

	instanceID = client.MapProviderIDToResourceID(instanceID)

	log.Infof("Node with nodeID translates to instance ID : %s", instanceID)

	attachedVolume, err := d.client.Compute().FindVolumeAttachment(ctx, compartmentID, req.VolumeId, instanceID)

	if attachedVolume != nil && attachedVolume.GetId() != nil {
		log = log.With("volumeAttachedId", *attachedVolume.GetId())
	}
	if err != nil {
		if !client.IsNotFound(err) {
			log.With("service", "compute", "verb", "get", "resource", "volumeAttachment", "statusCode", util.GetHttpStatusCode(err)).
				With(zap.Error(err)).Error("Error while fetching the Volume details. Unable to detach Volume from the node.")
			errorType = util.GetError(err)
			csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, err
		}
		if attachedVolume == nil {
			log.With("service", "compute", "verb", "get", "resource", "volumeAttachment", "statusCode", util.GetHttpStatusCode(err)).With(zap.Error(err)).
				Error("Unable to find volume attachment for volume to detach. Volume is possibly already detached. Nothing to do in Un-publish Volume.")
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}
		log.Info("Attached Volume is still in Detaching state")
	}
	if attachedVolume.GetLifecycleState() != core.VolumeAttachmentLifecycleStateDetaching {
		log.With("instanceID", *attachedVolume.GetInstanceId()).Info("Detaching Volume")
		err = d.client.Compute().DetachVolume(ctx, *attachedVolume.GetId())
		if err != nil {
			log.With("service", "compute", "verb", "delete", "resource", "volumeAttachment", "statusCode", util.GetHttpStatusCode(err)).
				With("instanceID", *attachedVolume.GetInstanceId()).With(zap.Error(err)).Error("Volume can not be detached")
			errorType = util.GetError(err)
			csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.Unknown, "volume can not be detached %s", err)
		}
	}
	log.With("instanceID", *attachedVolume.GetInstanceId()).Info("Waiting for Volume to Detach")
	err = d.client.Compute().WaitForVolumeDetached(ctx, *attachedVolume.GetId())
	if err != nil {
		log.With("service", "compute", "verb", "get", "resource", "volumeAttachment", "statusCode", util.GetHttpStatusCode(err)).
			With("instanceID", *attachedVolume.GetInstanceId()).With(zap.Error(err)).Error("timed out waiting for volume to be detached")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Unknown, "timed out waiting for volume to be detached %s", err)
	}

	multipath := false

	if attachedVolume.GetIsMultipath() != nil {
		multipath = *attachedVolume.GetIsMultipath()
	}

	// sleeping to ensure block volume plugin logs out of iscsi connections on nodes before delete
	if multipath {
		log.Info("Waiting for 90 seconds to ensure block volume plugin logs out of iscsi connections on nodes")
		time.Sleep(90 * time.Second)
	}

	log.Info("Un-publishing Volume Completed")
	csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

// ValidateVolumeCapabilities checks whether the volume capabilities requested
// are supported.
func (d *BlockVolumeControllerDriver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	log := d.logger.With("volumeID", req.VolumeId, "csiOperation", "validateVolumeCapabilities")

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}

	if req.VolumeCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities must be provided")
	}

	// check if volume exist before trying to validate it
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	volume, err := d.client.BlockStorage().GetVolume(ctx, req.VolumeId)
	if err != nil {
		log.With("service", "blockstorage", "verb", "get", "resource", "volume", "statusCode", util.GetHttpStatusCode(err)).
			With(zap.Error(err)).Error("Volume ID not found.")
		return nil, status.Errorf(codes.NotFound, "Volume ID not found.")
	}

	if *volume.Id == req.VolumeId {
		return &csi.ValidateVolumeCapabilitiesResponse{
			Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
				VolumeCapabilities: []*csi.VolumeCapability{
					{
						AccessMode: supportedAccessModes[0],
					},
					{
						AccessMode: supportedAccessModes[1],
					},
				},
			},
		}, nil
	}
	return nil, status.Errorf(codes.NotFound, "VolumeId mis-match.")
}

// ListVolumes returns a list of all requested volumes
func (d *BlockVolumeControllerDriver) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// GetCapacity returns the capacity of the storage pool
func (d *BlockVolumeControllerDriver) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

// ControllerGetCapabilities returns the capabilities of the controller service.
func (d *BlockVolumeControllerDriver) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	newCap := func(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
		return &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		}
	}

	var caps []*csi.ControllerServiceCapability
	for _, cap := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
		csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
	} {
		caps = append(caps, newCap(cap))
	}

	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}

	return resp, nil
}

// validateCapabilities validates the requested capabilities. It returns an error
// if it doesn't satisfy the currently supported modes of OCI Block Volume
func (d *BlockVolumeControllerDriver) validateCapabilities(caps []*csi.VolumeCapability) error {
	vcaps := supportedAccessModes

	hasSupport := func(mode csi.VolumeCapability_AccessMode_Mode) bool {
		for _, m := range vcaps {
			if mode == m.Mode {
				return true
			}
		}
		return false
	}

	for _, cap := range caps {
		if hasSupport(cap.AccessMode.Mode) {
			continue
		} else {
			// we need to make sure all capabilities are supported. Revert back
			// in case we have a cap that is supported, but is invalidated now
			d.logger.Errorf("The VolumeCapability isn't supported: %s", cap.AccessMode)
			return fmt.Errorf("invalid volume capabilities requested. Only SINGLE_NODE_WRITER is supported ('accessModes.ReadWriteOnce' on Kubernetes)")
		}
	}

	return nil
}

// CreateSnapshot will be called by the CO to create a new snapshot from a
// source volume on behalf of a user.
func (d *BlockVolumeControllerDriver) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	startTime := time.Now()
	var snapshotMetricDimension string
	var errorType string

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.Name
	log := d.logger.With("snapshotName", req.Name, "sourceVolumeId", req.SourceVolumeId, "csiOperation", "createSnapshot")

	if req.Name == "" {
		log.Error("Volume Snapshot name must be provided.")
		snapshotMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Error(codes.InvalidArgument, "Volume snapshot name must be provided")
	}

	sourceVolumeId := req.SourceVolumeId
	if sourceVolumeId == "" {
		log.Error("Volume snapshot source ID must be provided")
		snapshotMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Error(codes.InvalidArgument, "Volume snapshot source ID must be provided")
	}

	snapshots, err := d.client.BlockStorage().GetVolumeBackupsByName(ctx, req.Name, d.config.CompartmentID)
	if err != nil {
		errorType = util.GetError(err)
		snapshotMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(startTime).Seconds(), dimensionsMap)
		log.With("service", "blockstorage", "verb", "get", "resource", "volumeBackup", "statusCode", util.GetHttpStatusCode(err)).
			Error("Failed to check the existence of the snapshot %s : %v", req.Name, err)
		return nil, status.Errorf(codes.Internal, "failed to check existence of snapshot %v", err)
	}

	if len(snapshots) > 1 {
		log.Error("Duplicate snapshot %q exists", req.Name)
		snapshotMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, fmt.Errorf("duplicate snapshot %q exists", req.Name)
	}

	if len(snapshots) > 0 {
		//Assigning existing snapshot

		snapshot := snapshots[0]
		log.Info("Snapshot already created, checking if lifecycleState is Available")

		log = log.With("volumeBackupId", *snapshot.Id)
		dimensionsMap[metrics.ResourceOCIDDimension] = *snapshot.Id

		if snapshot.VolumeId != nil && *snapshot.VolumeId != sourceVolumeId {
			log.Errorf("Snapshot %s exists for another volume with id %s", req.Name, *snapshot.VolumeId)
			snapshotMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.AlreadyExists, "Snapshot %s exists for another volume with id %s", req.Name, *snapshot.VolumeId)
		}

		ts := timestamppb.New(snapshot.TimeCreated.Time)

		log.Infof("Checking if backup %v has become available", *snapshot.Id)
		blockVolumeAvailable, err := isBlockVolumeAvailable(snapshot)
		if err != nil {
			log.Errorf("Error while waiting for backup to become available %q: %v", req.Name, err)
			errorType = util.GetError(err)
			snapshotMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(snapshot.TimeRequestReceived.Time).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.Internal, "Backup did not become available %q: %v", req.Name, err)
		}

		if blockVolumeAvailable {
			log.Info("Snapshot is created and available.")
			snapshotMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(snapshot.TimeRequestReceived.Time).Seconds(), dimensionsMap)
		} else {
			log.Infof("Backup has not become available yet, controller will retry")
			snapshotMetricDimension = util.GetMetricDimensionForComponent(util.BackupCreating, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(snapshot.TimeRequestReceived.Time).Seconds(), dimensionsMap)
		}

		return &csi.CreateSnapshotResponse{
			Snapshot: &csi.Snapshot{
				SnapshotId:     *snapshot.Id,
				SourceVolumeId: *snapshot.VolumeId,
				SizeBytes:      *snapshot.SizeInMBs * client.MiB,
				CreationTime:   ts,
				ReadyToUse:     blockVolumeAvailable,
			},
		}, nil
	}

	snapshotParams, err := extractSnapshotParameters(req.GetParameters())
	if err != nil {
		log.With(zap.Error(err)).Error("Failed to parse volumesnapshotclass parameters.")
		snapshotMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse volumesnapshotclass parameters %v", err)
	}

	backupTags := &config.TagConfig{
		FreeformTags: snapshotParams.freeformTags,
		DefinedTags:  snapshotParams.definedTags,
	}

	snapshot, err := d.client.BlockStorage().CreateVolumeBackup(ctx, core.CreateVolumeBackupDetails{
		VolumeId:     &sourceVolumeId,
		Type:         snapshotParams.backupType,
		DisplayName:  &req.Name,
		FreeformTags: backupTags.FreeformTags,
		DefinedTags:  backupTags.DefinedTags,
	})

	if err != nil {
		log.With("service", "blockstorage", "verb", "create", "resource", "volumeBackup", "statusCode", util.GetHttpStatusCode(err)).
			Errorf("Could not create snapshot %q: %v", req.Name, err)
		errorType = util.GetError(err)
		snapshotMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Internal, "Could not create snapshot %q: %v", req.Name, err)
	}

	log = log.With("volumeBackupId", *snapshot.Id)
	dimensionsMap[metrics.ResourceOCIDDimension] = *snapshot.Id

	ts := timestamppb.New(snapshot.TimeCreated.Time)

	_, err = d.client.BlockStorage().AwaitVolumeBackupAvailableOrTimeout(ctx, *snapshot.Id)
	if err != nil {
		if strings.Contains(err.Error(), "timed out") {
			log.Infof("Backup did not become available immediately after creation, controller will retry")
			snapshotMetricDimension = util.GetMetricDimensionForComponent(util.BackupCreating, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return &csi.CreateSnapshotResponse{
				Snapshot: &csi.Snapshot{
					SnapshotId:     *snapshot.Id,
					SourceVolumeId: *snapshot.VolumeId,
					SizeBytes:      *snapshot.SizeInMBs * client.MiB,
					CreationTime:   ts,
					ReadyToUse:     false,
				},
			}, nil
		} else {
			log.With("service", "blockstorage", "verb", "get", "resource", "volumeBackup", "statusCode", util.GetHttpStatusCode(err)).
				Errorf("Error while waiting for backup to become available %q: %v", req.Name, err)
			errorType = util.GetError(err)
			snapshotMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(startTime).Seconds(), dimensionsMap)
			log.Errorf("Backup did not become available %q: %v", req.Name, err)
			return nil, status.Errorf(codes.Internal, "Backup did not become available %q: %v", req.Name, err)
		}
	}

	log.Info("Snapshot is created and available.")
	snapshotMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotProvision, time.Since(startTime).Seconds(), dimensionsMap)

	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SnapshotId:     *snapshot.Id,
			SourceVolumeId: *snapshot.VolumeId,
			SizeBytes:      *snapshot.SizeInMBs * client.MiB,
			CreationTime:   ts,
			ReadyToUse:     true,
		},
	}, nil
}

// DeleteSnapshot will be called by the CO to delete a snapshot.
func (d *BlockVolumeControllerDriver) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	startTime := time.Now()
	var snapshotMetricDimension string
	var errorType string
	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.SnapshotId

	log := d.logger.With("SnapshotId", req.SnapshotId, "csiOperation", "deleteSnapshot")

	if req.SnapshotId == "" {
		snapshotMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotDelete, time.Since(startTime).Seconds(), dimensionsMap)
		log.Errorf("SnapshotID is empty")
		return nil, status.Error(codes.InvalidArgument, "SnapshotId must be provided")
	}

	err := d.client.BlockStorage().DeleteVolumeBackup(ctx, req.SnapshotId)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		errorType = util.GetError(err)
		snapshotMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotDelete, time.Since(startTime).Seconds(), dimensionsMap)
		log.With("service", "blockstorage", "verb", "delete", "resource", "volumeBackup", "statusCode", util.GetHttpStatusCode(err)).
			Errorf("Failed to delete snapshot, snapshotId: %s, error: %v", req.SnapshotId, err)
		return nil, fmt.Errorf("failed to delete snapshot, snapshotId: %s, error: %v", req.SnapshotId, err)
	}

	log.Info("Snapshot is deleted.")
	snapshotMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = snapshotMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.BlockSnapshotDelete, time.Since(startTime).Seconds(), dimensionsMap)
	return &csi.DeleteSnapshotResponse{}, nil
}

// ListSnapshots returns all the matched snapshots
func (d *BlockVolumeControllerDriver) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListSnapshots is not supported yet")
}

// ControllerExpandVolume returns ControllerExpandVolume request
func (d *BlockVolumeControllerDriver) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	startTime := time.Now()
	volumeId := req.GetVolumeId()
	if volumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "UpdateVolume volumeId must be provided")
	}
	log := d.logger.With("volumeID", volumeId, "csiOperation", "expandVolume")
	var errorType string
	var csiMetricDimension string

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.VolumeId

	newSize, err := csi_util.ExtractStorage(req.CapacityRange)
	if err != nil {
		log.With(zap.Error(err)).Error("invalid capacity range")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVExpand, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.OutOfRange, "invalid capacity range: %v", err)
	}

	//make sure this method is idempotent by checking existence of volume with same name.
	volume, err := d.client.BlockStorage().GetVolume(ctx, volumeId)
	if err != nil {
		log.With("service", "blockstorage", "verb", "get", "resource", "volume", "statusCode", util.GetHttpStatusCode(err)).
			With(zap.Error(err)).Error("Failed to find existence of volume")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVExpand, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.Internal, "failed to check existence of volume %v", err)
	}

	log = log.With("volumeName", volume.DisplayName)
	newSizeInGB := csi_util.RoundUpSize(newSize, 1*client.GiB)
	oldSize := *volume.SizeInGBs

	if newSizeInGB <= oldSize {
		log.Infof("Existing volume size: %v Requested volume size: %v No action needed.", *volume.SizeInGBs, newSizeInGB)
		return &csi.ControllerExpandVolumeResponse{
			CapacityBytes:         oldSize * client.GiB,
			NodeExpansionRequired: true,
		}, nil
	}

	updateVolumeDetails := core.UpdateVolumeDetails{
		DisplayName: volume.DisplayName,
		SizeInGBs:   &newSizeInGB,
	}

	volume, err = d.client.BlockStorage().UpdateVolume(ctx, volumeId, updateVolumeDetails)

	if err != nil {
		message := fmt.Sprintf("Update volume failed %v", err)
		log.With("service", "blockstorage", "verb", "update", "resource", "volume", "statusCode", util.GetHttpStatusCode(err)).
			Error(message)
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVExpand, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Error(codes.Internal, message)
	}
	_, err = d.client.BlockStorage().AwaitVolumeAvailableORTimeout(ctx, volumeId)
	if err != nil {
		log.With("service", "blockstorage", "verb", "get", "resource", "volume", "statusCode", util.GetHttpStatusCode(err)).
			Error("Volume Expansion failed with time out")
		errorType = util.GetError(err)
		csiMetricDimension = util.GetMetricDimensionForComponent(errorType, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVExpand, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.DeadlineExceeded, "ControllerExpand failed with time out %v", err.Error())
	}

	log.Info("Volume is expanded.")
	csiMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.PVExpand, time.Since(startTime).Seconds(), dimensionsMap)

	return &csi.ControllerExpandVolumeResponse{
		CapacityBytes:         newSize,
		NodeExpansionRequired: true,
	}, nil
}

// ControllerGetVolume returns ControllerGetVolumeResponse response
func (d *BlockVolumeControllerDriver) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerGetVolume is not supported yet")
}

func (d *BlockVolumeControllerDriver) ControllerModifyVolume(ctx context.Context, request *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerModifyVolume is not supported yet")
}

func provision(ctx context.Context, log *zap.SugaredLogger, c client.Interface, volName string, volSize int64, availDomainName, compartmentID,
	backupID, srcVolumeID, kmsKeyID string, vpusPerGB int64, bvTags *config.TagConfig) (core.Volume, error) {

	volSizeGB, minSizeGB := csi_util.RoundUpSize(volSize, 1*client.GiB), csi_util.RoundUpMinSize()

	if minSizeGB > volSizeGB {
		volSizeGB = minSizeGB
	}

	volumeDetails := core.CreateVolumeDetails{
		AvailabilityDomain: &availDomainName,
		CompartmentId:      &compartmentID,
		DisplayName:        &volName,
		SizeInGBs:          &volSizeGB,
		VpusPerGB:          &vpusPerGB,
	}

	if backupID != "" {
		volumeDetails.SourceDetails = &core.VolumeSourceFromVolumeBackupDetails{Id: &backupID}
	} else if srcVolumeID != "" {
		volumeDetails.SourceDetails = &core.VolumeSourceFromVolumeDetails{Id: &srcVolumeID}
	}

	if kmsKeyID != "" {
		volumeDetails.KmsKeyId = &kmsKeyID
	}
	if bvTags != nil && bvTags.FreeformTags != nil {
		volumeDetails.FreeformTags = bvTags.FreeformTags
	}
	if bvTags != nil && bvTags.DefinedTags != nil {
		volumeDetails.DefinedTags = bvTags.DefinedTags
		if len(volumeDetails.DefinedTags) > MaxDefinedTagPerVolume {
			log.With("service", "blockstorage", "verb", "create", "resource", "volume", "volumeName", volName).
				Warn("the number of defined tags in the volume create request is beyond the limit. removing system tags from the details")
			delete(volumeDetails.DefinedTags, OkeSystemTagNamesapce)
		}
	}

	newVolume, err := c.BlockStorage().CreateVolume(ctx, volumeDetails)

	if err != nil {
		log.With("service", "blockstorage", "verb", "create", "resource", "volume", "statusCode", util.GetHttpStatusCode(err)).
			With(zap.Error(err)).With("volumeName", volName).Error("Create volume failed.")
		status.Errorf(codes.Unknown, "Create volume failed")
		return core.Volume{}, err
	}
	log.With("volumeName", volName).Info("Volume is provisioned.")
	return *newVolume, nil
}

// We would derive whether the customer wants in-transit encryption or not based on if the node is launched using
// in-transit encryption enabled or not.
func getAttachmentOptions(ctx context.Context, client client.ComputeInterface, attachmentType, instanceID string) (VolumeAttachmentOption, error) {
	volumeAttachmentOption := VolumeAttachmentOption{}
	if attachmentType == attachmentTypeParavirtualized {
		volumeAttachmentOption.useParavirtualizedAttachment = true
	}
	instance, err := client.GetInstance(ctx, instanceID)
	if err != nil {
		return volumeAttachmentOption, err
	}
	if *instance.LaunchOptions.IsPvEncryptionInTransitEnabled {
		volumeAttachmentOption.enableInTransitEncryption = true
	}
	return volumeAttachmentOption, nil
}

func isBlockVolumeAvailable(backup core.VolumeBackup) (bool, error) {
	switch state := backup.LifecycleState; state {
	case core.VolumeBackupLifecycleStateAvailable:
		return true, nil
	case core.VolumeBackupLifecycleStateFaulty,
		core.VolumeBackupLifecycleStateTerminated,
		core.VolumeBackupLifecycleStateTerminating:
		return false, errors.Errorf("snapshot did not become available (lifecycleState=%q)", state)
	}
	return false, nil
}

func getBVTags(logger *zap.SugaredLogger, tags *config.InitialTags, vp VolumeParameters) *config.TagConfig {

	bvTags := &config.TagConfig{}
	if tags != nil && tags.BlockVolume != nil {
		bvTags = tags.BlockVolume
	}

	// use storage class level tags if provided
	scTags := &config.TagConfig{
		FreeformTags: vp.freeformTags,
		DefinedTags:  vp.definedTags,
	}
	if scTags.FreeformTags != nil || scTags.DefinedTags != nil {
		bvTags = scTags
	}
	// merge final tags with common tags
	if enableOkeSystemTags && util.IsCommonTagPresent(tags) {
		bvTags = util.MergeTagConfig(bvTags, tags.Common)
	}
	return bvTags
}
