// Copyright 2018 Oracle and/or its affiliates. All rights reserved.
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

package fss

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner"
	"github.com/oracle/oci-cloud-controller-manager/pkg/volume/provisioner/plugin"
	fss "github.com/oracle/oci-go-sdk/v50/filestorage"
	"github.com/oracle/oci-go-sdk/v50/identity"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v8/controller"
)

const (
	ociVolumeID = "volume.beta.kubernetes.io/oci-volume-id"
	ociExportID = "volume.beta.kubernetes.io/oci-export-id"
	// AnnotationMountTargetID configures the mount target to use when
	// provisioning a FSS volume
	AnnotationMountTargetID = "volume.beta.kubernetes.io/oci-mount-target-id"

	// MntTargetID is the name of the parameter which hold the target mount target ocid.
	MntTargetID = "mntTargetId"
)

const (
	defaultTimeout  = 5 * time.Minute
	defaultInterval = 5 * time.Second
)

// filesystemProvisioner is the internal provisioner for OCI filesystem volumes
type filesystemProvisioner struct {
	client client.Interface

	// region is the oci region in which the kubernetes cluster is located.
	region string
	// compartmentID is the oci compartment in which the kubernetes cluster
	// is located.
	compartmentID string

	logger *zap.SugaredLogger
}

var _ plugin.ProvisionerPlugin = &filesystemProvisioner{}

var (
	errNoCandidateFound = errors.New("no candidate mount targets found")
	errNotFound         = errors.New("not found")
)

// NewFilesystemProvisioner creates a new file system provisioner that creates
// filesystems using OCI File System Service.
func NewFilesystemProvisioner(logger *zap.SugaredLogger, client client.Interface, region, compartmentID string) plugin.ProvisionerPlugin {
	return &filesystemProvisioner{
		client:        client,
		region:        region,
		compartmentID: compartmentID,
		logger:        logger,
	}
}

func (fsp *filesystemProvisioner) getOrCreateFileSystem(ctx context.Context, logger *zap.SugaredLogger, ad, displayName string) (*fss.FileSystem, error) {
	summary, err := fsp.client.FSS().GetFileSystemSummaryByDisplayName(ctx, fsp.compartmentID, ad, displayName)
	if err != nil && !client.IsNotFound(err) {
		return nil, err
	}
	if summary != nil {
		return fsp.client.FSS().AwaitFileSystemActive(ctx, logger, *summary.Id)
	}

	fs, err := fsp.client.FSS().CreateFileSystem(ctx, fss.CreateFileSystemDetails{
		CompartmentId:      &fsp.compartmentID,
		AvailabilityDomain: &ad,
		DisplayName:        &displayName,
	})
	if err != nil {
		return nil, err
	}

	logger.With("fileSystemID", *fs.Id).Info("Created FileSystem")

	return fsp.client.FSS().AwaitFileSystemActive(ctx, logger, *fs.Id)
}

func (fsp *filesystemProvisioner) getOrCreateExport(ctx context.Context, logger *zap.SugaredLogger, fsID, exportSetID string) (*fss.Export, error) {
	summary, err := fsp.client.FSS().FindExport(ctx, fsp.compartmentID, fsID, exportSetID)
	if err != nil && !client.IsNotFound(err) {
		return nil, err
	}
	if summary != nil {
		return fsp.client.FSS().AwaitExportActive(ctx, logger, *summary.Id)
	}

	path := "/" + fsID

	// If export doesn't already exist create it.
	export, err := fsp.client.FSS().CreateExport(ctx, fss.CreateExportDetails{
		ExportSetId:  &exportSetID,
		FileSystemId: &fsID,
		Path:         &path,
	})
	if err != nil {
		return nil, err
	}

	logger.With("exportID", *export.Id).Info("Created Export")
	return fsp.client.FSS().AwaitExportActive(ctx, logger, *export.Id)
}

// getMountTargetID retrieves MountTarget OCID if provided.
func getMountTargetID(opts controller.ProvisionOptions) string {
	if opts.PVC != nil {
		if mtID := opts.PVC.Annotations[AnnotationMountTargetID]; mtID != "" {
			return mtID
		}
	}
	return opts.StorageClass.Parameters[MntTargetID]
}

// isReadOnly determines if the given slice of PersistentVolumeAccessModes
// permits mounting as read only.
func isReadOnly(modes []v1.PersistentVolumeAccessMode) bool {
	for _, mode := range modes {
		if mode == v1.ReadWriteMany || mode == v1.ReadWriteOnce {
			return false
		}
	}
	return true
}

func (fsp *filesystemProvisioner) Provision(options controller.ProvisionOptions, ad *identity.AvailabilityDomain) (*v1.PersistentVolume, error) {
	ctx := context.Background()
	fsDisplayName := fmt.Sprintf("%s%s", provisioner.GetPrefix(), options.PVC.UID)
	logger := fsp.logger.With(
		"availabilityDomain", ad,
		"fileSystemDisplayName", fsDisplayName,
	)

	// Require that a user provides a MountTarget ID.
	mtID := getMountTargetID(options)
	if mtID == "" {
		return nil, errors.New("no mount target ID provided (via PVC annotation nor StorageClass option)")
	}

	logger = logger.With("mountTargetID", mtID)

	// Wait for MountTarget to be ACTIVE.
	target, err := fsp.client.FSS().AwaitMountTargetActive(ctx, logger, mtID)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to retrieve mount target")
		return nil, err
	}

	// Ensure MountTarget required fields are set.
	if len(target.PrivateIpIds) == 0 {
		logger.Error("Failed to find private IPs associated with the Mount Target")
		return nil, errors.Errorf("mount target has no associated private IPs")
	}
	if target.ExportSetId == nil {
		logger.Error("Mount target has no export set associated with it")
		return nil, errors.Errorf("mount target has no export set associated with it")
	}

	// Randomly select a MountTarget IP address to attach to.
	var ip string
	{
		id := target.PrivateIpIds[rand.Int()%len(target.PrivateIpIds)]
		logger = logger.With("privateIPID", id)
		privateIP, err := fsp.client.Networking().GetPrivateIP(ctx, id)
		if err != nil {
			logger.With(zap.Error(err)).Error("Failed to retrieve IP address for mount target")
			return nil, err
		}
		if privateIP.IpAddress == nil {
			logger.Error("PrivateIp has no IpAddress")
			return nil, errors.Errorf("PrivateIp %q associated with MountTarget %q has no IpAddress", id, mtID)
		}
		ip = *privateIP.IpAddress
	}
	logger = logger.With("privateIP", ip)

	logger.Info("Creating FileSystem")
	fs, err := fsp.getOrCreateFileSystem(ctx, logger, *ad.Name, fsDisplayName)
	if err != nil {
		return nil, err
	}
	logger = logger.With("fileSystemID", *fs.Id)

	logger.Info("Creating Export")
	export, err := fsp.getOrCreateExport(ctx, logger, *fs.Id, *target.ExportSetId)
	if err != nil {
		logger.With(zap.Error(err)).Error("Failed to create export.")
		return nil, err
	}

	logger.With("exportID", *export.Id).Info("All OCI resources provisioned")

	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName,
			Annotations: map[string]string{
				ociVolumeID: *fs.Id,
				ociExportID: *export.Id,
			},
			Labels: map[string]string{plugin.LabelZoneRegion: fsp.region},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: *options.StorageClass.ReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			//NOTE: fs storage doesn't enforce quota, capacity is meaningless here.
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				NFS: &v1.NFSVolumeSource{
					Server:   ip,
					Path:     *export.Path,
					ReadOnly: isReadOnly(options.PVC.Spec.AccessModes),
				},
			},
			MountOptions: options.StorageClass.MountOptions,
		},
	}, nil
}

// Delete terminates the OCI resources associated with the given PVC.
func (fsp *filesystemProvisioner) Delete(volume *v1.PersistentVolume) error {
	ctx := context.Background()
	exportID := volume.Annotations[ociExportID]
	if exportID == "" {
		return errors.Errorf("%q annotation not found on PV", ociExportID)
	}

	filesystemID := volume.Annotations[ociVolumeID]
	if filesystemID == "" {
		return errors.Errorf("%q annotation not found on PV", ociVolumeID)
	}

	logger := fsp.logger.With(
		"fileSystemID", filesystemID,
		"exportID", exportID,
	)

	logger.Info("Deleting export")
	if err := fsp.client.FSS().DeleteExport(ctx, exportID); err != nil {
		if !client.IsNotFound(err) {
			logger.With(zap.Error(err)).Error("Failed to delete export")
			return err
		}
		logger.With(zap.Error(err)).Info("Export not found. Unable to delete it")
	}

	logger.Info("Deleting File System")
	if err := fsp.client.FSS().DeleteFileSystem(ctx, filesystemID); err != nil {
		if !client.IsNotFound(err) {
			return err
		}
		logger.Info("FileSystem not found. Unable to delete it")
	}
	return nil
}
