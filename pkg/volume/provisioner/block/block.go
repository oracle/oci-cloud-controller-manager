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
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/plugin"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/core"
	"github.com/oracle/oci-go-sdk/identity"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/controller"
	"strconv"
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
)

// blockProvisioner is the internal provisioner for OCI block volumes
type blockProvisioner struct {
	client                client.Interface
	volumeRoundingEnabled bool
	minVolumeSize         resource.Quantity

	region        string
	compartmentID string

	logger *zap.SugaredLogger
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
	return &blockProvisioner{
		client:                client,
		region:                region,
		volumeRoundingEnabled: volumeRoundingEnabled,
		minVolumeSize:         minVolumeSize,
		compartmentID:         compartmentID,
		logger:                logger,
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
	ctx := context.Background()
	for _, accessMode := range options.PVC.Spec.AccessModes {
		if accessMode != v1.ReadWriteOnce {
			return nil, fmt.Errorf("invalid access mode %v specified. Only %v is supported", accessMode, v1.ReadWriteOnce)
		}
	}

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

	//make sure this method is idempotent by checking existence of volume with same name.
	volumes, err := block.client.BlockStorage().GetVolumesByName(context.Background(), string(options.PVC.UID), block.compartmentID)
	if err != nil {
		logger.Error("Failed to find existence of volume %s", err)
		return nil, fmt.Errorf("failed to check existence of volume %v", err)
	}

	if len(volumes) > 1 {
		logger.Error("Duplicate volume exists")
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
			return nil, err
		}
	}

	logger.With("volumeID", *volume.Id).Info("Waiting for volume to become available.")
	volume, err = block.client.BlockStorage().AwaitVolumeAvailable(ctx, *volume.Id)
	if err != nil {
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

	return pv, nil
}

func isVolumeBackupOcid(ocid string) bool {
	res, _ := regexp.MatchString(volumeBackupOCIDPrefixExp, ocid)
	return res
}

// Delete destroys a OCI volume created by Provision
func (block *blockProvisioner) Delete(volume *v1.PersistentVolume) error {
	ctx := context.Background()

	id, ok := volume.Annotations[OCIVolumeID]
	if !ok {
		return errors.New("volumeid annotation not found on PV")
	}

	logger := block.logger.With("volumeID", id)

	logger.Info("Deleting volume")
	err := block.client.BlockStorage().DeleteVolume(ctx, id)
	if client.IsNotFound(err) {
		logger.With(zap.Error(err)).Info("Volume not found. Presuming already deleted.")
		return nil
	}

	return errors.Wrap(err, "failed to delete volume from OCI")
}
