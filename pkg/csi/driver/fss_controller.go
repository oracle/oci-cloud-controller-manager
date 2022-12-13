package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	csi_util "github.com/oracle/oci-cloud-controller-manager/pkg/csi-util"
	"github.com/oracle/oci-cloud-controller-manager/pkg/metrics"
	"github.com/oracle/oci-cloud-controller-manager/pkg/oci/client"
	"github.com/oracle/oci-cloud-controller-manager/pkg/util"
	fss "github.com/oracle/oci-go-sdk/v65/filestorage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var (
	fssSupportedVolumeCapabilities = []csi.VolumeCapability_AccessMode{
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER,
		},
		{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		},
	}
)

// StorageClassParameters holds configuration
type StorageClassParameters struct {
	// availabilityDomain where File System and Mount Target should exist
	availabilityDomain string
	// compartmentOcid where File System and Mount Target should exist
	compartmentOcid string
	//kmsKey is the KMS key that would be used as CMEK key for FSS
	kmsKey string
	// exportPath is the file system export path
	exportPath string
	// exportOptions are a json string to be passed for export creation
	exportOptions []fss.ClientOptions
	// mountTargetOcid is provided, a new mount target will not be created
	mountTargetOcid string
	// mountTargetSubnetOcid is provided, a new mount target will be created in this subnet
	mountTargetSubnetOcid string
	// encryptInTransit if enabled, it will be passed in the volume context
	encryptInTransit string
	// tags
	scTags *config.TagConfig
}

func (d *FSSControllerDriver) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	startTime := time.Now()
	var log = d.logger.With("volumeName", req.Name)

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume Name must be provided")
	}

	volumeCapabilities := req.GetVolumeCapabilities()

	if volumeCapabilities == nil || len(volumeCapabilities) == 0 {
		return nil, status.Error(codes.InvalidArgument, "VolumeCapabilities must be provided in CreateVolumeRequest")
	}

	volumeName := req.Name

	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = volumeName

	if err := checkForSupportedVolumeCapabilities(volumeCapabilities); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Requested Volume Capability not supported")
	}

	log, response, storageClassParameters, err, done := extractStorageClassParameters(d, log, dimensionsMap, volumeName, req.GetParameters(), startTime)
	if done {
		return response, err
	}

	log, mountTargetOCID, mountTargetIp, exportSetId, response, err, done := d.getOrCreateMountTarget(ctx, *storageClassParameters, volumeName, log, dimensionsMap, startTime)
	if done {
		return response, err
	}

	isDeleteMountTarget := "true"

	if storageClassParameters.mountTargetOcid != "" {
		isDeleteMountTarget = "false"
	}

	freeformTags := storageClassParameters.scTags.FreeformTags
	if freeformTags == nil {
		freeformTags = make(map[string]string)
		storageClassParameters.scTags.FreeformTags = freeformTags
	}
	freeformTags["isDeleteMountTarget"] = isDeleteMountTarget
	freeformTags["mountTargetOCID"] = mountTargetOCID
	freeformTags["exportSetId"] = exportSetId

	log, filesystemOCID, response, err, done := d.getOrCreateFileSystem(ctx, *storageClassParameters, volumeName, log, dimensionsMap, startTime)
	if done {
		return response, err
	}

	log, response, err, done = d.getOrCreateExport(ctx, err, *storageClassParameters, filesystemOCID, exportSetId, log, dimensionsMap, startTime)
	if done {
		return response, err
	}

	fssVolumeHandle := fmt.Sprintf("%s:%s:%s", filesystemOCID, mountTargetIp, storageClassParameters.exportPath)
	log.With("volumeID", fssVolumeHandle).Info("All FSS resource successfully created")
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      fssVolumeHandle,
			CapacityBytes: 0,

			VolumeContext: map[string]string{
				"encryptInTransit": storageClassParameters.encryptInTransit,
			},
		},
	}, nil
}

func checkForSupportedVolumeCapabilities(volumeCaps []*csi.VolumeCapability) error {
	hasSupport := func(cap *csi.VolumeCapability) error {
		if blk := cap.GetBlock(); blk != nil {
			return fmt.Errorf("driver does not support block volumes")
		}
		for _, c := range fssSupportedVolumeCapabilities {
			if c.GetMode() == cap.AccessMode.GetMode() {
				return nil
			}
		}
		return fmt.Errorf("driver does not support access mode %v", cap.AccessMode.GetMode())
	}

	for _, c := range volumeCaps {
		if err := hasSupport(c); err != nil {
			return err
		}
	}
	return nil
}

func (d *FSSControllerDriver) getOrCreateFileSystem(ctx context.Context, storageClassParameters StorageClassParameters, volumeName string, log *zap.SugaredLogger, dimensionsMap map[string]string, startTime time.Time) (*zap.SugaredLogger, string, *csi.CreateVolumeResponse, error, bool) {
	//make sure this method is idempotent by checking existence of volume with same name.
	log.Info("searching for existing filesystem")
	foundConflictingFs, fileSystemSummaries, err := d.client.FSS().GetFileSystemSummaryByDisplayName(context.Background(), storageClassParameters.compartmentOcid, storageClassParameters.availabilityDomain, volumeName)
	if err != nil && !client.IsNotFound(err) {
		message := ""
		if foundConflictingFs {
			conflictMsg := ""
			for _, fileSystemSummary := range fileSystemSummaries {
				conflictMsg += fmt.Sprintf("%s: %s, ", *fileSystemSummary.Id, fileSystemSummary.LifecycleState)
			}
			message = fmt.Sprintf("conflicting File System(s) %v", conflictMsg)
		} else {
			message = "failed to check existence of File System"
		}
		log.With(zap.Error(err)).Error(message)
		csiMetricDimension := util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.FSSProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, "", nil, status.Errorf(codes.Internal, "%s, error: %s", message, err.Error()), true
	}

	if len(fileSystemSummaries) > 1 {
		log.Error("Duplicate File system exists")
		csiMetricDimension := util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.FSSProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, "", nil, fmt.Errorf("duplicate File system %q exists", volumeName), true
	}

	provisionedFileSystem := &fss.FileSystem{}

	if len(fileSystemSummaries) > 0 {
		//Assigning existing File system
		provisionedFileSystem = &fss.FileSystem{Id: fileSystemSummaries[0].Id}

	} else {
		// Creating new file system
		provisionedFileSystem, err = provisionFileSystem(log, d.client, volumeName, storageClassParameters)
		if err != nil {
			log.With(zap.Error(err)).Error("New File System creation failed")
			csiMetricDimension := util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.FSSProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, "", nil, status.Errorf(codes.Internal, "New File System creation failed, error: %s", err.Error()), true
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	filesystemOCID := volumeName
	if provisionedFileSystem.Id != nil {
		filesystemOCID = *provisionedFileSystem.Id
	}
	log = log.With("fssID", filesystemOCID)
	_, err = d.client.FSS().AwaitFileSystemActive(ctx, log, *provisionedFileSystem.Id)
	if err != nil {
		log.With(zap.Error(err)).Error("Await File System failed with time out")
		csiMetricDimension := util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.FSSProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, "", nil, status.Errorf(codes.DeadlineExceeded, "Await File System failed with time out, error: %s", err.Error()), true
	}

	log.Info("File system is Available.")
	csiMetricDimension := util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	dimensionsMap[metrics.ResourceOCIDDimension] = filesystemOCID
	metrics.SendMetricData(d.metricPusher, metrics.FSSProvision, time.Since(startTime).Seconds(), dimensionsMap)
	return log, filesystemOCID, nil, nil, false
}

func (d *FSSControllerDriver) getOrCreateMountTarget(ctx context.Context, storageClassParameters StorageClassParameters, volumeName string, log *zap.SugaredLogger, dimensionsMap map[string]string, startTime time.Time) (*zap.SugaredLogger, string, string, string, *csi.CreateVolumeResponse, error, bool) {
	// Mount Target creation
	provisionedMountTarget := &fss.MountTarget{}
	isExistingMountTargetUsed := false

	if storageClassParameters.mountTargetOcid != "" {
		isExistingMountTargetUsed = true
		provisionedMountTarget = &fss.MountTarget{Id: &storageClassParameters.mountTargetOcid}
	} else {
		log.Info("searching for existing Mount Target")
		//make sure this method is idempotent by checking existence of volume with same name.
		foundConflictingMt, mountTargets, err := d.client.FSS().GetMountTargetSummaryByDisplayName(context.Background(), storageClassParameters.compartmentOcid, storageClassParameters.availabilityDomain, volumeName)
		if err != nil && !client.IsNotFound(err) {
			message := ""
			if foundConflictingMt {
				conflictMsg := ""
				for _, mt := range mountTargets {
					conflictMsg += fmt.Sprintf("%s: %s, ", *mt.Id, mt.LifecycleState)
				}
				message = fmt.Sprintf("conflicting Mount Target(s) %v", conflictMsg)
			} else {
				message = "failed to check existence of Mount Target"
			}
			log.With(zap.Error(err)).Error(message)
			csiMetricDimension := util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.MTProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return log, "", "", "", nil, status.Errorf(codes.Internal, "%s, error: %s", message, err.Error()), true
		}

		if len(mountTargets) > 1 {
			log.Error("Duplicate Mount Target exists")
			csiMetricDimension := util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.MTProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return log, "", "", "", nil, status.Errorf(codes.Internal, "duplicate Mount Target %s exists", volumeName), true
		}

		if len(mountTargets) > 0 {
			//Mount Target already exists so checking state of the mount target and returning the same.
			log.Info("Mount Target is already created!")
			isExistingMountTargetUsed = true
			//Assigning existing mount target
			provisionedMountTarget = &fss.MountTarget{
				Id: mountTargets[0].Id,
			}

		} else {
			// Creating new mount target
			provisionedMountTarget, err = provisionMountTarget(log, d.client, volumeName, storageClassParameters)
			if err != nil {
				log.With(zap.Error(err)).Error("New Mount Target creation failed")
				csiMetricDimension := util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
				dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
				metrics.SendMetricData(d.metricPusher, metrics.MTProvision, time.Since(startTime).Seconds(), dimensionsMap)
				return log, "", "", "", nil, status.Errorf(codes.Internal, "New Mount Target creation failed, error: %s", err.Error()), true
			}
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	mountTargetOCID := volumeName
	if provisionedMountTarget.Id != nil {
		mountTargetOCID = *provisionedMountTarget.Id
	}
	log = log.With("mountTargetID", mountTargetOCID)
	activeMountTarget, err := d.client.FSS().AwaitMountTargetActive(ctx, log, *provisionedMountTarget.Id)
	if err != nil {
		log.With(zap.Error(err)).Error("await mount target to be available failed with time out")
		if !isExistingMountTargetUsed {
			csiMetricDimension := util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.MTProvision, time.Since(startTime).Seconds(), dimensionsMap)
		}
		return log, "", "", "", nil, status.Errorf(codes.DeadlineExceeded, "await mount target to be available failed with time out, error: %s", err.Error()), true
	}
	activeMountTargetName := *activeMountTarget.DisplayName
	log = log.With("mountTargetName", activeMountTargetName)
	if activeMountTarget.PrivateIpIds == nil || len(activeMountTarget.PrivateIpIds) == 0 {
		log.Error("IP not assigned to mount target")
		if !isExistingMountTargetUsed {
			csiMetricDimension := util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.MTProvision, time.Since(startTime).Seconds(), dimensionsMap)
		}
		return log, "", "", "", nil, status.Errorf(codes.Internal, "IP not assigned to mount target"), true
	}
	mountTargetIpId := activeMountTarget.PrivateIpIds[0]
	log.Infof("getting private IP of mount target from mountTargetIpId %s", mountTargetIpId)
	privateIpObject, err := d.client.Networking().GetPrivateIp(ctx, mountTargetIpId)
	if err != nil {
		log.With(zap.Error(err)).Error("Failed to fetch Mount Target Private IP from IP ID: %s", mountTargetIpId)
		if !isExistingMountTargetUsed {
			csiMetricDimension := util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.MTProvision, time.Since(startTime).Seconds(), dimensionsMap)
		}
		return log, "", "", "", nil, status.Errorf(codes.Internal, "Failed to fetch Mount Target Private IP from IP ID: %s, error: %s", mountTargetIpId, err.Error()), true
	}
	mountTargetIp := *privateIpObject.IpAddress
	log = log.With("mountTargetValidatedIp", mountTargetIp)
	if activeMountTarget.ExportSetId == nil || *activeMountTarget.ExportSetId == "" {
		log.Error("ExportSetId not assigned to mount target")
		if !isExistingMountTargetUsed {
			csiMetricDimension := util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.MTProvision, time.Since(startTime).Seconds(), dimensionsMap)
		}
		return log, "", "", "", nil, status.Errorf(codes.Internal, "ExportSetId not assigned to mount target"), true
	}
	exportSetId := *activeMountTarget.ExportSetId
	log.Infof("Mount Target is Active with exportSetId %s", exportSetId)

	if !isExistingMountTargetUsed {
		csiMetricDimension := util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		dimensionsMap[metrics.ResourceOCIDDimension] = mountTargetOCID
		metrics.SendMetricData(d.metricPusher, metrics.MTProvision, time.Since(startTime).Seconds(), dimensionsMap)
	}
	return log, mountTargetOCID, mountTargetIp, exportSetId, nil, nil, false
}

func (d *FSSControllerDriver) getOrCreateExport(ctx context.Context, err error, storageClassParameters StorageClassParameters, filesystemOCID string, exportSetId string, log *zap.SugaredLogger, dimensionsMap map[string]string, startTime time.Time) (*zap.SugaredLogger, *csi.CreateVolumeResponse, error, bool) {
	log.Info("searching for existing export")
	exportSummary, err := d.client.FSS().FindExport(ctx, filesystemOCID, storageClassParameters.exportPath, exportSetId)

	if err != nil && !client.IsNotFound(err) {
		message := ""
		if exportSummary != nil {
			message = fmt.Sprintf("conflicting Export %s", *exportSummary.Path)
		} else {
			message = "failed to check existence of export"
		}
		log.With(zap.Error(err)).Error(message)
		csiMetricDimension := util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.ExportProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return log, nil, status.Errorf(codes.Internal, "%s, error: %s", message, err.Error()), true
	}

	provisionedExport := &fss.Export{}
	if exportSummary != nil {
		provisionedExport = &fss.Export{Id: exportSummary.ExportSetId}
	} else {
		// Creating new export
		provisionedExport, err = provisionExport(log, d.client, filesystemOCID, exportSetId, storageClassParameters)
		if err != nil {
			log.With(zap.Error(err)).Error("New Export creation failed")
			csiMetricDimension := util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
			dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
			metrics.SendMetricData(d.metricPusher, metrics.ExportProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return log, nil, status.Errorf(codes.Internal, "New Export creation failed, error: %s", err.Error()), true
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	exportId := storageClassParameters.exportPath
	if provisionedExport.Id != nil {
		exportId = *provisionedExport.Id
	}
	log = log.With("exportId", exportId)
	_, err = d.client.FSS().AwaitExportActive(ctx, log, exportId)
	if err != nil {
		log.With(zap.Error(err)).Error("await export failed with time out")
		csiMetricDimension := util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.ExportProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return log, nil, status.Errorf(codes.DeadlineExceeded, "await export failed with time out, error: %s", err.Error()), true
	}

	log.Info("Export is Active.")
	csiMetricDimension := util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
	dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
	dimensionsMap[metrics.ResourceOCIDDimension] = exportId
	metrics.SendMetricData(d.metricPusher, metrics.ExportProvision, time.Since(startTime).Seconds(), dimensionsMap)
	return log, nil, nil, false
}

func extractStorageClassParameters(d *FSSControllerDriver, log *zap.SugaredLogger, dimensionsMap map[string]string, volumeName string, parameters map[string]string, startTime time.Time) (*zap.SugaredLogger, *csi.CreateVolumeResponse, *StorageClassParameters, error, bool) {

	storageClassParameters := &StorageClassParameters{
		encryptInTransit: "false",
	}

	compartmentId, ok := parameters["compartmentOcid"]
	if !ok {
		compartmentId = d.config.CompartmentID
		log.Infof("compartmentOcid parameter not provided in storage class, defaulting to %s", compartmentId)
	}
	log = log.With("storageClassCompartmentOCID", compartmentId)
	storageClassParameters.compartmentOcid = compartmentId

	availabilityDomain, ok := parameters["availabilityDomain"]
	if !ok {
		log.Errorf("AvailabilityDomain not provided in storage class")
		dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.FSSProvision, time.Since(startTime).Seconds(), dimensionsMap)
		metrics.SendMetricData(d.metricPusher, metrics.MTProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return log, nil, nil, status.Errorf(codes.InvalidArgument, "AvailabilityDomain not provided in storage class"), true
	}

	ad, err := d.client.Identity().GetAvailabilityDomainByName(context.Background(), compartmentId, availabilityDomain)
	if err != nil {
		log.With(zap.Error(err)).Errorf("invalid available domain: %s or compartmentID: %s", availabilityDomain, compartmentId)
		dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.FSSProvision, time.Since(startTime).Seconds(), dimensionsMap)
		return log, nil, nil, status.Errorf(codes.InvalidArgument, "invalid available domain: %s or compartment ID: %s, error: %s", availabilityDomain, compartmentId, err.Error()), true
	}
	availabilityDomain = *ad.Name
	log = log.With("availabilityDomain", availabilityDomain)
	storageClassParameters.availabilityDomain = availabilityDomain

	mountTargetOcid, ok := parameters["mountTargetOcid"]
	if !ok {
		mountTargetSubnetOcid, ok := parameters["mountTargetSubnetOcid"]
		if !ok {
			log.Errorf("Neither Mount Target Ocid nor Mount Target Subnet Ocid provided in storage class")
			dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.MTProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "Neither Mount Target Ocid nor Mount Target Subnet Ocid provided in storage class"), true
		}
		log = log.With("mountTargetSubnetOcid", mountTargetSubnetOcid)
		log.Info("Mount Target Ocid not provided, to be created")
		storageClassParameters.mountTargetSubnetOcid = mountTargetSubnetOcid
	} else {
		storageClassParameters.mountTargetOcid = mountTargetOcid
		log = log.With("mountTargetOcid", mountTargetOcid)
		log.Info("Mount Target Ocid provided, new mount target will not be created")
	}

	exportPath, ok := parameters["exportPath"]
	if !ok {
		exportPath = "/" + volumeName
		log.Infof("exportPath not provided using %s as exportPath", exportPath)
	}
	log = log.With("exportPath", exportPath)
	storageClassParameters.exportPath = exportPath

	exportOptionsString, ok := parameters["exportOptions"]
	if ok && exportOptionsString != "" {
		log.Infof("exportOptions provided %s", exportOptionsString)
		var exportOptions []fss.ClientOptions
		err := json.Unmarshal([]byte(exportOptionsString), &exportOptions)
		if err != nil {
			log.With(zap.Error(err)).Errorf("failed to parse exportOptions provided " +
				"for storage class. please check the exportOptions in parameters section of storage class")
			dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.ExportProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "failed to parse exportOptions provided "+
				"for storage class. please check the exportOptions in parameters section of storage class"), true
		}
		storageClassParameters.exportOptions = exportOptions
	}

	encryptInTransit, ok := parameters["encryptInTransit"]
	if ok && encryptInTransit == "true" {
		storageClassParameters.encryptInTransit = "true"
	}

	kmsKey, ok := parameters["kmsKeyOcid"]
	if !ok {
		log.Info("kmsKeyOcid not provided, using oracle managed keys")
	} else {
		storageClassParameters.kmsKey = kmsKey
	}

	// use initial tags for all FSS resources
	fssTags := &config.TagConfig{}
	if d.config.Tags != nil && d.config.Tags.FSS != nil {
		fssTags = d.config.Tags.FSS
	}

	// use storage class level tags if provided
	scTags := &config.TagConfig{}

	initialFreeformTagsOverride, ok := parameters[initialFreeformTagsOverride]
	if ok && initialFreeformTagsOverride != "" {
		freeformTags := make(map[string]string)
		err := json.Unmarshal([]byte(initialFreeformTagsOverride), &freeformTags)
		if err != nil {
			log.With(zap.Error(err)).Errorf("failed to parse freeform tags provided " +
				"for storageclass. please check the parameters section on the storage class")
			dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.FSSProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "failed to parse freeform tags provided "+
				"for storageclass. please check the parameters section on the storage class"), true
		}
		scTags.FreeformTags = freeformTags
	}

	initialDefinedTagsOverride, ok := parameters[initialDefinedTagsOverride]
	if ok && initialDefinedTagsOverride != "" {
		definedTags := make(map[string]map[string]interface{})
		err := json.Unmarshal([]byte(initialDefinedTagsOverride), &definedTags)
		if err != nil {
			log.With(zap.Error(err)).Errorf("failed to parse defined tags provided " +
				"for storageclass. please check the parameters section on the storage class")
			dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.FSSProvision, time.Since(startTime).Seconds(), dimensionsMap)
			return log, nil, nil, status.Errorf(codes.InvalidArgument, "failed to parse defined tags provided "+
				"for storageclass. please check the parameters section on the storage class"), true
		}
		scTags.DefinedTags = definedTags
	}

	// storage class tags overwrite initial BV Tags
	if scTags.FreeformTags != nil || scTags.DefinedTags != nil {
		fssTags = scTags
	}
	storageClassParameters.scTags = fssTags

	log.Info("Successfully parsed storage class parameters")

	return log, nil, storageClassParameters, nil, false
}

func provisionFileSystem(log *zap.SugaredLogger, c client.Interface, volumeName string, storageClassParameters StorageClassParameters) (*fss.FileSystem, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	log.Info("Creating new File System")
	createFileSystemDetails := fss.CreateFileSystemDetails{
		AvailabilityDomain: &storageClassParameters.availabilityDomain,
		CompartmentId:      &storageClassParameters.compartmentOcid,
		DisplayName:        &volumeName,
		FreeformTags:       storageClassParameters.scTags.FreeformTags,
		DefinedTags:        storageClassParameters.scTags.DefinedTags,
	}
	if storageClassParameters.kmsKey != "" {
		createFileSystemDetails.KmsKeyId = &storageClassParameters.kmsKey
	}
	return c.FSS().CreateFileSystem(ctx, createFileSystemDetails)
}

func provisionMountTarget(log *zap.SugaredLogger, c client.Interface, volumeName string, storageClassParameters StorageClassParameters) (*fss.MountTarget, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	log.Info("Creating new Mount Target")
	createMountTargetDetails := fss.CreateMountTargetDetails{
		AvailabilityDomain: &storageClassParameters.availabilityDomain,
		CompartmentId:      &storageClassParameters.compartmentOcid,
		DisplayName:        &volumeName,
		SubnetId:           &storageClassParameters.mountTargetSubnetOcid,
		FreeformTags:       storageClassParameters.scTags.FreeformTags,
		DefinedTags:        storageClassParameters.scTags.DefinedTags,
	}
	return c.FSS().CreateMountTarget(ctx, createMountTargetDetails)
}

func provisionExport(log *zap.SugaredLogger, c client.Interface, filesystemOCID string, exportSetId string, storageClassParameters StorageClassParameters) (*fss.Export, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	log.Info("Creating new Export")
	createExportDetails := fss.CreateExportDetails{
		ExportSetId:  &exportSetId,
		FileSystemId: &filesystemOCID,
		Path:         &storageClassParameters.exportPath,
	}

	if storageClassParameters.exportOptions != nil {
		createExportDetails.ExportOptions = storageClassParameters.exportOptions
	}
	return c.FSS().CreateExport(ctx, createExportDetails)
}

func (d *FSSControllerDriver) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	startTime := time.Now()
	volumeId := req.GetVolumeId()
	log := d.logger.With("volumeID", volumeId)
	dimensionsMap := make(map[string]string)
	dimensionsMap[metrics.ResourceOCIDDimension] = req.VolumeId
	volumeHandler := csi_util.ValidateFssId(volumeId)
	filesystemOcid, mountTargetIP, exportPath := volumeHandler.FilesystemOcid, volumeHandler.MountTargetIPAddress, volumeHandler.FsExportPath
	if filesystemOcid == "" || mountTargetIP == "" || exportPath == "" {
		log.Error("Unable to parse Volume Id")
		csiMetricDimension := util.GetMetricDimensionForComponent(util.ErrValidation, util.CSIStorageType)
		dimensionsMap[metrics.ComponentDimension] = csiMetricDimension
		metrics.SendMetricData(d.metricPusher, metrics.FSSDelete, time.Since(startTime).Seconds(), dimensionsMap)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Volume ID provided %s", volumeId)
	}

	log = log.With("fssID", filesystemOcid).With("mountTargetIP", mountTargetIP).With("exportPath", exportPath)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Info("Getting file system to be deleted")
	fileSystem, err := d.client.FSS().GetFileSystem(ctx, filesystemOcid)
	if err != nil {
		if !client.IsNotFound(err) {
			log.With(zap.Error(err)).Error("Failed to delete filesystem.")
			dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.MTDelete, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.Internal, "failed to delete filesystem, volumeId: %s ERROR: %v", volumeId, err.Error())
		} else {
			log.Info("File system does not exist.")
		}
	}

	compartmentID := *fileSystem.CompartmentId
	log = log.With("storageClassCompartmentOCID", compartmentID, "volumeName", *fileSystem.DisplayName)

	freeformTags := fileSystem.FreeformTags

	mountTargetOCID := ""
	exportSetId := ""
	isDeleteMountTarget := false

	if freeformTags != nil {
		for k, v := range freeformTags {
			switch k {
			case "mountTargetOCID":
				mountTargetOCID = v
			case "isDeleteMountTarget":
				if v == "true" {
					isDeleteMountTarget = true
				}
			case "exportSetId":
				exportSetId = freeformTags["exportSetId"]
			}
		}
	}

	if isDeleteMountTarget {
		log = log.With("mountTargetOCID", mountTargetOCID)
		log.Info("filesystem tagged with mount target ocid, deleting mount target")
		// first delete Mount Target
		err = d.client.FSS().DeleteMountTarget(ctx, mountTargetOCID)
		if err != nil {
			if !client.IsNotFound(err) {
				log.With(zap.Error(err)).Error("Failed to delete mount target.")
				dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
				metrics.SendMetricData(d.metricPusher, metrics.MTDelete, time.Since(startTime).Seconds(), dimensionsMap)
				return nil, status.Errorf(codes.Internal, "failed to delete mount target, mountTargetOcid: %s, error: %s", mountTargetOCID, err.Error())
			} else {
				log.Info("Mount Target does not exist.")
			}
		} else {
			dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.MTDelete, time.Since(startTime).Seconds(), dimensionsMap)
			log.Info("Mount Target is deleted.")
		}
	} else {
		log.Infof("filesystem not tagged with isDeleteMountTarget as true, skip deleting Mount Target %s", mountTargetOCID)
	}

	if exportSetId != "" {
		log.Infof("searching export with tagged exportSetId %s", exportSetId)
		exportSummary, err := d.client.FSS().FindExport(ctx, filesystemOcid, exportPath, exportSetId)
		if err != nil {
			if !client.IsNotFound(err) {
				if exportSummary != nil {
					log.Infof("export %s is in state %s", *exportSummary.Id, exportSummary.LifecycleState)
				} else {
					log.With(zap.Error(err)).Error("Failed to find export.")
					dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
					metrics.SendMetricData(d.metricPusher, metrics.ExportDelete, time.Since(startTime).Seconds(), dimensionsMap)
					return nil, status.Errorf(codes.Internal, "failed to find export, exportPath: %s, error: %s", exportPath, err.Error())
				}
			} else {
				log.Info("Export does not exist")
			}
		} else {
			log.Infof("deleting export with exportId %s", *exportSummary.Id)
			err = d.client.FSS().DeleteExport(ctx, *exportSummary.Id)
			if err != nil {
				log.With(zap.Error(err)).Error("failed to delete export.")
				dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
				metrics.SendMetricData(d.metricPusher, metrics.ExportDelete, time.Since(startTime).Seconds(), dimensionsMap)
				return nil, status.Errorf(codes.Internal, "failed to delete export, exportId: %s, error: %s", *exportSummary.Id, err.Error())
			}
			dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.ExportDelete, time.Since(startTime).Seconds(), dimensionsMap)
			log.Info("Export is deleted.")
		}
	} else {
		log.Info("filesystem not tagged with exportSetId, skip deleting export")
	}
	log.Info("deleting file system")
	// last delete File System
	err = d.client.FSS().DeleteFileSystem(ctx, filesystemOcid)
	if err != nil {
		if !client.IsNotFound(err) {
			log.With(zap.Error(err)).Error("Failed to delete file system.")
			dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.GetError(err), util.CSIStorageType)
			metrics.SendMetricData(d.metricPusher, metrics.FSSDelete, time.Since(startTime).Seconds(), dimensionsMap)
			return nil, status.Errorf(codes.Internal, "failed to delete file system, volumeId: %s, error: %s", volumeId, err.Error())
		} else {
			log.Info("File system does not exist")
		}
	} else {
		log.Info("File system is deleted.")
		dimensionsMap[metrics.ComponentDimension] = util.GetMetricDimensionForComponent(util.Success, util.CSIStorageType)
		metrics.SendMetricData(d.metricPusher, metrics.FSSDelete, time.Since(startTime).Seconds(), dimensionsMap)
	}
	return &csi.DeleteVolumeResponse{}, nil
}

func (d *FSSControllerDriver) ControllerGetCapabilities(ctx context.Context, request *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
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
	for _, capability := range []csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	} {
		caps = append(caps, newCap(capability))
	}
	resp := &csi.ControllerGetCapabilitiesResponse{
		Capabilities: caps,
	}
	return resp, nil
}

func (d *FSSControllerDriver) ControllerPublishVolume(ctx context.Context, request *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *FSSControllerDriver) ControllerUnpublishVolume(ctx context.Context, request *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *FSSControllerDriver) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "Volume ID must be provided")
	}
	volumeId := req.GetVolumeId()

	log := d.logger.With("volumeID", volumeId)

	if req.VolumeCapabilities == nil {
		log.Error("Volume Capabilities must be provided")
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities must be provided")
	}

	volumeHandler := csi_util.ValidateFssId(volumeId)
	filesystemOcid, mountTargetIP, exportPath := volumeHandler.FilesystemOcid, volumeHandler.MountTargetIPAddress, volumeHandler.FsExportPath

	if filesystemOcid == "" || mountTargetIP == "" || exportPath == "" {
		log.Info("Unable to parse Volume Id")
		return nil, status.Error(codes.InvalidArgument, "Invalid Volume ID provided")
	}

	log = log.With("fssID", filesystemOcid).With("mountTargetIP", mountTargetIP).With("exportPath", exportPath)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log.Info("Fetching filesystem")
	fileSystem, err := d.client.FSS().GetFileSystem(ctx, filesystemOcid)
	if err != nil {
		log.With(zap.Error(err)).Error("File system not found.")
		return nil, status.Errorf(codes.NotFound, "File system not found. error: %s", err.Error())
	}

	freeformTags := fileSystem.FreeformTags

	mountTargetOCID := ""
	exportSetId := ""

	if freeformTags != nil {
		for k, v := range freeformTags {
			switch k {
			case "mountTargetOCID":
				mountTargetOCID = v
			case "exportSetId":
				exportSetId = freeformTags["exportSetId"]
			}
		}
	}

	mountTarget := &fss.MountTarget{}

	if mountTargetOCID != "" {
		log = log.With("mountTargetOCID", mountTargetOCID)
		log.Info("filesystem tagged with mount target ocid, getting mount target")
		mountTarget, err = d.client.FSS().GetMountTarget(ctx, mountTargetOCID)
		if err != nil && !client.IsNotFound(err) {
			log.With(zap.Error(err)).Error("failed to get mount target")
			return nil, status.Errorf(codes.NotFound, "failed to get mount target, error: %s", err.Error())
		}
	}

	if mountTarget != nil && mountTarget.PrivateIpIds != nil {
		mountTargetIpId := mountTarget.PrivateIpIds[0]
		log = log.With("mountTargetIpId", mountTargetIpId)
		privateIpObject, err := d.client.Networking().GetPrivateIp(ctx, mountTargetIpId)
		if err != nil {
			log.With(zap.Error(err)).Errorf("Failed to fetch Mount Target Private IP from IP ID: %s", mountTargetIpId)
			return nil, status.Errorf(codes.NotFound, "Failed to fetch Mount Target Private IP from IP ID: %s, error: %s", mountTargetIpId, err.Error())
		}
		mountTargetIp := *privateIpObject.IpAddress
		log = log.With("mountTargetValidatedIp", mountTargetIp)
		if mountTargetIp != mountTargetIP {
			log.Errorf("Failed to fetch Mount Target Private IP from IP ID: %s", mountTargetIpId)
			return nil, status.Errorf(codes.NotFound, "Mount Target IP mis-match.")
		}
	} else {
		log.Error("Mount Target not found")
		return nil, status.Errorf(codes.NotFound, "Mount Target not found")
	}

	exportSummary := &fss.ExportSummary{}
	if exportSetId != "" {
		log.Infof("searching export with tagged exportSetId %s", exportSetId)
		exportSummary, err = d.client.FSS().FindExport(ctx, filesystemOcid, exportPath, exportSetId)
		if err != nil {
			log.With(zap.Error(err)).Error("export not found.")
			return nil, status.Errorf(codes.NotFound, "export not found. error: %s", err.Error())
		}
	}

	if exportSummary == nil || exportSummary.Path == nil || *exportSummary.Path != exportPath {
		return nil, status.Errorf(codes.NotFound, "ExportPath mis-match.")
	}

	volumeCapabilities := req.GetVolumeCapabilities()

	for _, capability := range volumeCapabilities {
		// Not supporting experimental volume capabilities
		if capability.GetAccessMode().Mode == csi.VolumeCapability_AccessMode_SINGLE_NODE_SINGLE_WRITER || capability.GetAccessMode().Mode == csi.VolumeCapability_AccessMode_SINGLE_NODE_MULTI_WRITER {
			return &csi.ValidateVolumeCapabilitiesResponse{}, nil
		}
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: volumeCapabilities,
		},
	}, nil

}

func (d *FSSControllerDriver) ListVolumes(ctx context.Context, request *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *FSSControllerDriver) GetCapacity(ctx context.Context, request *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *FSSControllerDriver) CreateSnapshot(ctx context.Context, request *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *FSSControllerDriver) DeleteSnapshot(ctx context.Context, request *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *FSSControllerDriver) ListSnapshots(ctx context.Context, request *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *FSSControllerDriver) ControllerExpandVolume(ctx context.Context, request *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (d *FSSControllerDriver) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
