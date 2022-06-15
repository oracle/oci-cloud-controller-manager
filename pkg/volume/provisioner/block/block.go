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

package block

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v8/controller"

	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/plugin"
	"github.com/oracle/oci-go-sdk/v50/common"
	"github.com/oracle/oci-go-sdk/v50/core"
	"github.com/oracle/oci-go-sdk/v50/identity"
)

const (
	// OCIVolumeID is the name of the oci volume id.
	OCIVolumeID = "ociVolumeID"
	// OCIVolumeBackupID is the name of the oci volume backup id annotation.
	OCIVolumeBackupID = "volume.beta.kubernetes.io/oci-volume-source"
	// FSType is the name of the file storage type parameter for storage classes.
	FSType                    = "fsType"
	volumeRoundingUpEnabled   = "volumeRoundingUpEnabled"
	volumeBackupOCIDPrefixExp = `^ocid[v]?[\d+]?[\.:]volumebackup[\.:]`
	timeout                   = time.Minute * 3
)

// blockProvisioner is the internal provisioner for OCI block volumes
type blockProvisioner struct {
	client                client.Interface
	volumeRoundingEnabled bool
	minVolumeSize         resource.Quantity

	region        string
	compartmentID string
	metricPusher  *metrics.MetricPusher
	logger        *zap.SugaredLogger
}

var _ plugin.ProvisionerPlugin = &blockProvisioner{}

// NewBlockProvisioner creates a new instance of the block storage provisioner
func NewBlockProvisioner(
	logger *zap.SugaredLogger,
	client client.Interface,
	region string,
	compartmentID string,
	volumeRoundingEnabled bool,
	minVolumeSize resource.Quantity,
) plugin.ProvisionerPlugin {
	var metricPusher *metrics.MetricPusher
	var err error

	metricPusher, err = metrics.NewMetricPusher(logger)
	if err != nil {
		logger.With("error", err).Error("Metrics collection could not be enabled")
		// disable metric collection
		metricPusher = nil
	}
	if metricPusher != nil {
		logger.Info("Metrics collection has been enabled")
	} else {
		logger.Info("Metrics collection has not been enabled")
	}

	return &blockProvisioner{
		client:                client,
		region:                region,
		volumeRoundingEnabled: volumeRoundingEnabled,
		minVolumeSize:         minVolumeSize,
		compartmentID:         compartmentID,
		logger:                logger,
		metricPusher:          metricPusher,
	}
}

func resolveFSType(options controller.ProvisionOptions) string {
	fsType, _ := options.StorageClass.Parameters[FSType]

	defaultFsType := "ext4"
	if fsType == "ext4" || fsType == "ext3" {
		return fsType
	} else if fsType != "" {
		//TODO: Remove this code when we support other than ext4 || ext3.
		return defaultFsType
	} else {
		//No fsType provided returning ext4
		return defaultFsType
	}
}

func roundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	return (volumeSizeBytes + allocationUnitBytes - 1) / allocationUnitBytes
}

func volumeRoundingEnabled(param map[string]string) bool {
	volumeRounding := true // default
	if volumeRoundingUpParam, ok := param[volumeRoundingUpEnabled]; ok {
		if enabled, err := strconv.ParseBool(volumeRoundingUpParam); err == nil && !enabled {
			volumeRounding = false
		}
	}
	return volumeRounding
}

// Provision creates an OCI block volume
func (block *blockProvisioner) Provision(options controller.ProvisionOptions, ad *identity.AvailabilityDomain) (*v1.PersistentVolume, error) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	for _, accessMode := range options.PVC.Spec.AccessModes {
		if accessMode != v1.ReadWriteOnce {
			return nil, fmt.Errorf("invalid access mode %v specified. Only %v is supported", accessMode, v1.ReadWriteOnce)
		}
	}

	var errorType string
	var fvdMetricDimension string

	// Calculate the volume size
	capacity, ok := options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	if !ok {
		return nil, fmt.Errorf("could not determine volume size for PVC")
	}

	volSizeMB := int(roundUpSize(capacity.Value(), 1024*1024))

	logger := block.logger.With(
		"availabilityDomain", *ad.Name,
		"volumeSize", volSizeMB,
	)
	logger.Info("Provisioning volume")

	if volumeRoundingEnabled(options.StorageClass.Parameters) {
		if block.volumeRoundingEnabled && block.minVolumeSize.Cmp(capacity) == 1 {
			volSizeMB = int(roundUpSize(block.minVolumeSize.Value(), 1024*1024))
			logger.With("roundedVolumeSize", volSizeMB).Warn("Attempted to provision volume with a capacity less than the minimum. Rounding up to ensure volume creation.")
			capacity = block.minVolumeSize
		}
	}

	volumeDetails := core.CreateVolumeDetails{
		AvailabilityDomain: ad.Name,
		CompartmentId:      common.String(block.compartmentID),
		DisplayName:        common.String(string(options.PVC.UID)),
		SizeInMBs:          common.Int64(int64(volSizeMB)),
	}

	if value, ok := options.PVC.Annotations[OCIVolumeBackupID]; ok {
		logger = logger.With("volumeBackupOCID", value)
		if isVolumeBackupOcid(value) {
			logger.Info("Creating volume from block volume backup.")
			volumeDetails.SourceDetails = &core.VolumeSourceFromVolumeBackupDetails{Id: &value}
		} else {
			logger.Info("Creating volume from block volume.")
			volumeDetails.SourceDetails = &core.VolumeSourceFromVolumeDetails{Id: &value}
		}
	}

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = string(options.PVC.UID)

	//make sure this method is idempotent by checking existence of volume with same name.
	volumes, err := block.client.BlockStorage().GetVolumesByName(ctx, string(options.PVC.UID), block.compartmentID)
	if err != nil {
		logger.Error("Failed to find existence of volume %s", err)
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(block.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, fmt.Errorf("failed to check existence of volume %v", err)
	}

	if len(volumes) > 1 {
		logger.Error("Duplicate volume exists")
		fvdMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(block.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, fmt.Errorf("duplicate volume %q exists", string(options.PVC.UID))
	}

	volume := &core.Volume{}

	if len(volumes) > 0 {
		//Volume already exists so checking state of the volume and returning the same.
		logger.Info("Volume already created!")
		//Assigning existing volume
		volume = &volumes[0]

	} else {
		// Create the volume.
		logger.Info("Creating new volume!")
		volume, err = block.client.BlockStorage().CreateVolume(ctx, volumeDetails)
		if err != nil {
			logger.With("Compartment Id", block.compartmentID).Error("Failed to create volume %s", err)
			errorType = util.GetError(err)
			fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
			dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
			metrics.SendMetricData(block.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, errors.Wrap(err, "Failed to create volume")
		}
	}

	logger.With("volumeID", *volume.Id).Info("Waiting for volume to become available.")
	volume, err = block.client.BlockStorage().AwaitVolumeAvailableORTimeout(ctx, *volume.Id)
	if err != nil {
		logger.With("volumeID", *volume.Id).Error("Timed out while waiting for the volume.")
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(block.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
		_ = block.client.BlockStorage().DeleteVolume(ctx, *volume.Id)
		return nil, errors.Wrap(err, "waiting for volume to become available")
	}

	filesystemType := resolveFSType(options)

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: *volume.Id,
			Annotations: map[string]string{
				OCIVolumeID: *volume.Id,
			},
			Labels: map[string]string{
				plugin.LabelZoneRegion:        block.region,
				plugin.LabelZoneFailureDomain: *ad.Name,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: *options.StorageClass.ReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): capacity,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				FlexVolume: &v1.FlexPersistentVolumeSource{
					Driver: plugin.OCIProvisionerName,
					FSType: filesystemType,
				},
			},
			MountOptions: options.StorageClass.MountOptions,
		},
	}
	fvdMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.FVDStorageType)
	dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
	dimensionsMap[metrics.ResourceOCIDDimension] = *volume.Id
	metrics.SendMetricData(block.metricPusher, metrics.PVProvision, time.Since(startTime).Seconds(), dimensionsMap)
	return pv, nil
}

func isVolumeBackupOcid(ocid string) bool {
	res, _ := regexp.MatchString(volumeBackupOCIDPrefixExp, ocid)
	return res
}

// Delete destroys a OCI volume created by Provision
func (block *blockProvisioner) Delete(volume *v1.PersistentVolume) error {
	startTime := time.Now()
	ctx := context.Background()

	var errorType string
	var fvdMetricDimension string

	id, ok := volume.Annotations[OCIVolumeID]
	if !ok {
		return errors.New("volumeid annotation not found on PV")
	}

	logger := block.logger.With("volumeID", id)

	logger.Info("Deleting volume")
	err := block.client.BlockStorage().DeleteVolume(ctx, id)

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = id

	if client.IsNotFound(err) {
		logger.With(zap.Error(err)).Info("Volume not found. Presuming already deleted.")
		fvdMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(block.metricPusher, metrics.PVDelete, time.Since(startTime).Seconds(), dimensionsMap)
		return nil
	}

	if err != nil {
		logger.Error("Couldn't delete the volume")
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(block.metricPusher, metrics.PVDelete, time.Since(startTime).Seconds(), dimensionsMap)
	} else {
		logger.Info("Successfully deleted the volume")
		fvdMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(block.metricPusher, metrics.PVDelete, time.Since(startTime).Seconds(), dimensionsMap)
	}

	return errors.Wrap(err, "failed to delete volume from OCI")
}
