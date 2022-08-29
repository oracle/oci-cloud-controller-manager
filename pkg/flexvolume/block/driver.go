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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	ociprovider "github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci"
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"github.com/oracle/oci-cloud-controller-manager/pkg/flexvolume"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util/disk"
	"github.com/oracle/oci-go-sdk/v50/core"
)

const (
	// FIXME: Assume lun 1 for now?? Can we get the LUN via the API?
	diskIDByPathTemplate = "/dev/disk/by-path/ip-%s:%d-iscsi-%s-lun-1"
	volumeOCIDTemplate   = "ocid1.volume.oc1.%s.%s"
	ocidPrefix           = "ocid1."
	iscsiError           = "Only ISCSI volume attachments are currently supported"
)

// OCIFlexvolumeDriver implements the flexvolume.Driver interface for OCI.
type OCIFlexvolumeDriver struct {
	K            kubernetes.Interface
	master       bool
	metricPusher *metrics.MetricPusher
}

// NewOCIFlexvolumeDriver creates a new driver
func NewOCIFlexvolumeDriver(logger *zap.SugaredLogger) (fvd *OCIFlexvolumeDriver, err error) {
	defer func() {
		if e := recover(); e != nil {
			fvd = nil
			err = fmt.Errorf("%+v", e)
		}
	}()

	path := GetConfigPath()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		k, err := constructKubeClient()
		if err != nil {
			return nil, err
		}
		var metricPusher *metrics.MetricPusher

		metricPusher, err = metrics.NewMetricPusher(logger)
		if err != nil {
			logger.With("error", err).Error("Metrics collection could not be enabled")
			// disable metric collection
			metricPusher = nil
		}
		if metricPusher != nil {
			logger.Info("Metrics collection has been enabled")
		} else {
			logger.Info("Metric collection not is enabled")
		}

		return &OCIFlexvolumeDriver{K: k, master: true, metricPusher: metricPusher}, nil
	} else if os.IsNotExist(err) {
		logger.With(zap.Error(err), "path", path).Debug("Config file does not exist. Assuming worker node.")
		return &OCIFlexvolumeDriver{}, nil
	}
	return nil, err
}

// GetDriverDirectory gets the ath for the flexvolume driver either from the
// env or default.
func GetDriverDirectory() string {
	// TODO(apryde): Document this ENV var.
	path := os.Getenv("OCI_FLEXD_DRIVER_DIRECTORY")
	if path == "" {
		path = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/oracle~oci"
	}
	return path
}

// GetConfigDirectory gets the path to where config files are stored.
func GetConfigDirectory() string {
	path := os.Getenv("OCI_FLEXD_CONFIG_DIRECTORY")
	if path != "" {
		return path
	}

	return GetDriverDirectory()
}

// GetConfigPath gets the path to the OCI API credentials.
func GetConfigPath() string {
	path := GetConfigDirectory()
	return filepath.Join(path, "config.yaml")
}

// GetKubeconfigPath gets the override path of the 'kubeconfig'. This override
// can be uses to explicitly set the name and location of the kubeconfig file
// via the OCI_FLEXD_KUBECONFIG_PATH environment variable. If this value is not
// specified then the default GetConfigDirectory mechanism is used.
func GetKubeconfigPath() string {
	kcp := os.Getenv("OCI_FLEXD_KUBECONFIG_PATH")
	if kcp == "" {
		kcp = fmt.Sprintf("%s/%s", strings.TrimRight(GetConfigDirectory(), "/"), "kubeconfig")
	}
	return kcp
}

// Init checks that we have the appropriate credentials and metadata API access
// on driver initialisation.
func (d OCIFlexvolumeDriver) Init(logger *zap.SugaredLogger) flexvolume.DriverStatus {
	path := GetConfigPath()
	if d.master {
		cfg, err := config.FromFile(path)
		if err != nil {
			return flexvolume.Fail(logger, err)
		}
		_, err = client.GetClient(logger, cfg)
		if err != nil {
			return flexvolume.Fail(logger, err)
		}

		_, err = constructKubeClient()
		if err != nil {
			return flexvolume.Fail(logger, err)
		}
	} else {
		logger.Debug("Assuming worker node.")
	}

	return flexvolume.Succeed(zap.New(nil).Sugar())
}

// deriveVolumeOCID will expand a partial OCID to a full OCID
// based on the region key and volume name.
func deriveVolumeOCID(regionKey string, volumeName string) string {
	if strings.HasPrefix(volumeName, ocidPrefix) {
		return volumeName
	}

	return fmt.Sprintf(volumeOCIDTemplate, regionKey, volumeName)
}

// constructKubeClient uses a kubeconfig layed down by a secret via deploy.sh to return
// a kube clientset.
func constructKubeClient() (*kubernetes.Clientset, error) {
	c, err := clientcmd.BuildConfigFromFlags("", GetKubeconfigPath())
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(c)
}

// lookupNodeID returns the OCID for the given nodeName.
func lookupNodeID(k kubernetes.Interface, nodeName string) (string, error) {
	n, err := k.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if n.Spec.ProviderID == "" {
		return "", errors.New("node is missing provider id")
	}
	return n.Spec.ProviderID, nil
}

func getISCSIAttachment(attachment core.VolumeAttachment) (*core.IScsiVolumeAttachment, error) {
	iscsiAttachment, ok := attachment.(core.IScsiVolumeAttachment)
	if !ok {
		return nil, errors.New(iscsiError)
	}
	return &iscsiAttachment, nil
}

// Attach initiates the attachment of the given OCI volume to the k8s worker
// node.
func (d OCIFlexvolumeDriver) Attach(logger *zap.SugaredLogger, opts flexvolume.Options, nodeName string) flexvolume.DriverStatus {
	startTime := time.Now()
	var errorType string
	var fvdMetricDimension string
	logger = logger.With("nodeName", nodeName)
	cfg, err := config.FromFile(GetConfigPath())
	if err != nil {
		return flexvolume.Fail(logger, err)
	}

	volumeOCID := deriveVolumeOCID(cfg.RegionKey, opts["kubernetes.io/pvOrVolumeName"])

	c, err := client.GetClient(logger, cfg)
	if err != nil {
		return flexvolume.Fail(logger, err)
	}

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = volumeOCID

	id, err := lookupNodeID(d.K, nodeName)
	if err != nil {
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, "Failed to look up node id: ", err)
	}

	// Handle possible oci:// prefix.
	id, err = ociprovider.MapProviderIDToInstanceID(id)
	if err != nil {
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, "Failed to map nodes provider id to instance id: ", err)
	}

	ctx := context.Background()

	instance, err := c.Compute().GetInstance(ctx, id)
	if err != nil {
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, "Failed to get instance: ", err)
	}

	compartmentID := *instance.CompartmentId

	//Checking if the volume is already attached
	attachment, err := c.Compute().FindVolumeAttachment(ctx, compartmentID, volumeOCID)
	if err != nil && !client.IsNotFound(err) {
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, "Got error in finding volume attachment", err)
	}
	// volume already attached to an instance
	if err == nil {
		if *attachment.GetInstanceId() != *instance.Id {
			fvdMetricDimension = util.GetMetricDimensionForComponent(util.ErrValidation, util.FVDStorageType)
			dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
			return flexvolume.Fail(logger, "Already attached to another instance: ", *attachment.GetInstanceId())
		}
		logger.With("volumeID", volumeOCID, "instanceID", *instance.Id).Info("Volume is already attached to instance")
		iscsiAttachment, err := getISCSIAttachment(attachment)
		if err != nil {
			errorType = util.GetError(err)
			fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
			dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
			return flexvolume.Fail(logger, iscsiError)
		}
		fvdMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.DriverStatus{
			Status: flexvolume.StatusSuccess,
			Device: fmt.Sprintf(diskIDByPathTemplate, *iscsiAttachment.Ipv4, *iscsiAttachment.Port, *iscsiAttachment.Iqn),
		}
	}
	// volume not attached to any instance, proceed with volume attachment
	logger.With("volumeID", volumeOCID, "instanceID", *instance.Id).Info("Attaching volume to instance")
	attachment, err = c.Compute().AttachVolume(ctx, *instance.Id, volumeOCID)
	if err != nil {
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, "Failed to attach volume: ", err)
	}
	attachment, err = c.Compute().WaitForVolumeAttached(ctx, *attachment.GetId())
	if err != nil {
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, err)
	}
	logger.With("attachmentID", *attachment.GetId()).Info("Volume attached")

	iscsiAttachment, err := getISCSIAttachment(attachment)
	if err != nil {
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, iscsiError)
	}

	fvdMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.FVDStorageType)
	dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.PVAttach, time.Since(startTime).Seconds(), dimensionsMap)

	return flexvolume.DriverStatus{
		Status: flexvolume.StatusSuccess,
		Device: fmt.Sprintf(diskIDByPathTemplate, *iscsiAttachment.Ipv4, *iscsiAttachment.Port, *iscsiAttachment.Iqn),
	}
}

// Detach detaches the volume from the worker node.
func (d OCIFlexvolumeDriver) Detach(logger *zap.SugaredLogger, pvOrVolumeName, nodeName string) flexvolume.DriverStatus {
	startTime := time.Now()
	logger = logger.With("node", nodeName, "volume", pvOrVolumeName)
	logger.Info("Looking for volume to detach.")
	var errorType string
	var fvdMetricDimension string
	cfg, err := config.FromFile(GetConfigPath())
	if err != nil {
		return flexvolume.Fail(logger, err)
	}
	c, err := client.GetClient(logger, cfg)
	if err != nil {
		return flexvolume.Fail(logger, err)
	}

	volumeOCID := deriveVolumeOCID(cfg.RegionKey, pvOrVolumeName)
	ctx := context.Background()

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = volumeOCID

	compartmentID, err := util.LookupNodeCompartment(d.K, nodeName)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Node is not found, volume is likely already detached.")
			fvdMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.FVDStorageType)
			dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
			return flexvolume.Succeed(logger, "Volume detachment completed.")
		}
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, "failed to get compartmentID from node annotation: ", err)
	}

	attachment, err := c.Compute().FindVolumeAttachment(ctx, compartmentID, volumeOCID)
	if err != nil {
		logger.Error("Error in finding volume attachment")
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, "Failed to find volume attachment: ", err)
	}
	logger.Info("Found volume to detach.")
	err = c.Compute().DetachVolume(ctx, *attachment.GetId())
	if err != nil {
		logger.Error("Error detaching the volume")
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, err)
	}

	err = c.Compute().WaitForVolumeDetached(ctx, *attachment.GetId())
	if err != nil {
		logger.Error("Error while waiting for volume to be detached")
		errorType = util.GetError(err)
		fvdMetricDimension = util.GetMetricDimensionForComponent(errorType, util.FVDStorageType)
		dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
		return flexvolume.Fail(logger, err)
	}

	fvdMetricDimension = util.GetMetricDimensionForComponent(util.Success, util.FVDStorageType)
	dimensionsMap[metrics.ComponentDimension] = fvdMetricDimension
	metrics.SendMetricData(d.metricPusher, metrics.PVDetach, time.Since(startTime).Seconds(), dimensionsMap)
	return flexvolume.Succeed(logger, "Volume detachment completed.")
}

// WaitForAttach searches for the the volume attachment created by Attach() and
// waits for its life cycle state to reach ATTACHED.
func (d OCIFlexvolumeDriver) WaitForAttach(mountDevice string, _ flexvolume.Options) flexvolume.DriverStatus {
	return flexvolume.DriverStatus{
		Status: flexvolume.StatusSuccess,
		Device: mountDevice,
	}
}

// IsAttached checks whether the volume is attached to the host.
// TODO(apryde): The documentation states that this is called from the Kubelet
// and KCM. Implementation requries credentials which won't be present on nodes
// but I've only ever seen it called by the KCM.
func (d OCIFlexvolumeDriver) IsAttached(logger *zap.SugaredLogger, opts flexvolume.Options, nodeName string) flexvolume.DriverStatus {
	cfg, err := config.FromFile(GetConfigPath())
	if err != nil {
		return flexvolume.Fail(logger, err)
	}
	c, err := client.GetClient(logger, cfg)
	if err != nil {
		return flexvolume.Fail(logger, err)
	}

	ctx := context.Background()
	volumeOCID := deriveVolumeOCID(cfg.RegionKey, opts["kubernetes.io/pvOrVolumeName"])

	compartmentID, err := util.LookupNodeCompartment(d.K, nodeName)
	if err != nil {
		return flexvolume.Fail(logger, "Failed to look up node compartment id: ", err)
	}

	attachment, err := c.Compute().FindVolumeAttachment(ctx, compartmentID, volumeOCID)
	if err != nil {
		return flexvolume.DriverStatus{
			Status:   flexvolume.StatusSuccess,
			Message:  err.Error(),
			Attached: false,
		}
	}

	logger.With("attachmentID", *attachment.GetId()).Info("Found volume attachment")

	return flexvolume.DriverStatus{
		Status:   flexvolume.StatusSuccess,
		Attached: true,
	}
}

// MountDevice connects the iSCSI target on the k8s worker node before mounting
// and (if necessary) formatting the disk.
func (d OCIFlexvolumeDriver) MountDevice(logger *zap.SugaredLogger, mountDir, mountDevice string, opts flexvolume.Options) flexvolume.DriverStatus {
	iSCSIMounter, err := disk.NewFromDevicePath(logger, mountDevice)
	if err != nil {
		return flexvolume.Fail(logger, err)
	}

	if isMounted, oErr := iSCSIMounter.DeviceOpened(mountDevice); oErr != nil {
		return flexvolume.Fail(logger, oErr)
	} else if isMounted {
		return flexvolume.Succeed(logger, "Device already mounted. Nothing to do.")
	}

	if err = iSCSIMounter.AddToDB(); err != nil {
		return flexvolume.Fail(logger, err)
	}
	if err = iSCSIMounter.SetAutomaticLogin(); err != nil {
		return flexvolume.Fail(logger, err)
	}
	if err = iSCSIMounter.Login(); err != nil {
		return flexvolume.Fail(logger, err)
	}

	if !waitForPathToExist(mountDevice, 20) {
		return flexvolume.Fail(logger, "Failed waiting for device to exist: ", mountDevice)
	}

	options := []string{}
	if opts[flexvolume.OptionReadWrite] == "ro" {
		options = []string{"ro"}
	}
	err = iSCSIMounter.FormatAndMount(mountDevice, mountDir, opts[flexvolume.OptionFSType], options)
	if err != nil {
		return flexvolume.Fail(logger, err)
	}

	return flexvolume.Succeed(logger)
}

// UnmountDevice unmounts the disk, logs out the iscsi target, and deletes the
// iscsi node record.
func (d OCIFlexvolumeDriver) UnmountDevice(logger *zap.SugaredLogger, mountPath string) flexvolume.DriverStatus {
	iSCSIMounter, err := disk.NewFromMountPointPath(logger, mountPath)
	if err != nil {
		if err == disk.ErrMountPointNotFound {
			return flexvolume.Succeed(logger, "Mount point not found. Nothing to do.")
		}
		return flexvolume.Fail(logger, err)
	}

	if err = iSCSIMounter.UnmountPath(mountPath); err != nil {
		return flexvolume.Fail(logger, err)
	}
	if err = iSCSIMounter.Logout(); err != nil {
		return flexvolume.Fail(logger, err)
	}
	if err = iSCSIMounter.RemoveFromDB(); err != nil {
		return flexvolume.Fail(logger, err)
	}

	return flexvolume.Succeed(logger)
}

// Mount is unimplemented as we use the --enable-controller-attach-detach flow
// and as such mount the drive in MountDevice().
func (d OCIFlexvolumeDriver) Mount(logger *zap.SugaredLogger, mountDir string, opts flexvolume.Options) flexvolume.DriverStatus {
	return flexvolume.NotSupported(logger)
}

// Unmount is unimplemented as we use the --enable-controller-attach-detach flow
// and as such unmount the drive in UnmountDevice().
func (d OCIFlexvolumeDriver) Unmount(logger *zap.SugaredLogger, mountDir string) flexvolume.DriverStatus {
	return flexvolume.NotSupported(logger)
}
