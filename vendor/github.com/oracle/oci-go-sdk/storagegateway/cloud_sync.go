// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Storage Gateway API
//
// API for the Storage Gateway service. Use this API to manage storage gateways and related items. For more
// information, see Overview of Storage Gateway (https://docs.cloud.oracle.com/iaas/Content/StorageGateway/Concepts/storagegatewayoverview.htm).
//

package storagegateway

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CloudSync The configuration details of a cloud sync. For general information about cloud syncs, see
//   Using Storage Gateway Cloud Sync (https://docs.cloud.oracle.com/iaas/Content/StorageGateway/Reference/storagegatewaycloudsync.htm).
type CloudSync struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the associated
	// storage gateway.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the associated storage gateway.
	StorageGatewayId *string `mandatory:"true" json:"storageGatewayId"`

	// The unique name of the cloud sync.
	// Example: `cloud_sync_52019`
	Name *string `mandatory:"true" json:"name"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the cloud sync.
	Id *string `mandatory:"true" json:"id"`

	// The path to a source directory or file for the cloud sync.
	// The path configuration depends on the direction of the cloud sync. To upload from an on-premises system to
	// the cloud, the source path resembles the following:
	// /cloudsync/mounts/<var>&lt;user_mount&gt;</var>/<var>&lt;path_to_directory&gt;</var>
	// To download from the cloud to an on-premises system, the source path resembles the following:
	// <var>&lt;storage_gateway_file_system&gt;</var>/<var>&lt;path_to_directory&gt;</var>
	// **Note:** To configure a cloud sync, the file system on an on-premises storage gateway must be mounted under
	// `/cloudsync/mounts`.
	SourcePath *string `mandatory:"true" json:"sourcePath"`

	// The path to a target directory or file for the cloud sync.
	// The path configuration depends on the direction of the cloud sync. To upload from an on-premises system to
	// the cloud, the target path resembles the following:
	// <var>&lt;storage_gateway_file_system&gt;</var>/<var>&lt;path_to_directory&gt;</var>
	// To download from the cloud to an on-premises system, the target path resembles the following:
	// /cloudsync/mounts/<var>&lt;user_mount&gt;</var>/<var>&lt;path_to_directory&gt;</var>
	// **Note:** To configure a cloud sync, the file system on an on-premises storage gateway must be mounted under
	// `/cloudsync/mounts`.
	TargetPath *string `mandatory:"true" json:"targetPath"`

	// Whether the cloud sync uploads data to the cloud. If "true", the cloud sync uploads data to the cloud.
	// Example: `true`
	IsUpload *bool `mandatory:"true" json:"isUpload"`

	// Whether the cloud sync automatically deletes files from the target when source files are renamed or deleted.
	// If "true", the cloud sync automatically deletes files from the target.
	// Example: `true`
	IsAutoDeletionEnabled *bool `mandatory:"true" json:"isAutoDeletionEnabled"`

	// The path to a file that lists a set of files to sync to the target. If you do not specify a file list, the
	// service syncs all files. The list file should reside under the `/cloudsync/` directory on the machine running
	// the storage gateway instance.
	// Example: `/cloudsync/files.list`
	FilesFrom *string `mandatory:"true" json:"filesFrom"`

	// The date and time the cloud sync was created, in the format defined by RFC3339.
	// Example: `2019-05-16T21:52:40.793Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The date and time the cloud sync started, in the format defined by RFC3339.
	// Example: `2019-05-16T22:45:30.793Z`
	TimeStarted *common.SDKTime `mandatory:"true" json:"timeStarted"`

	// The current lifecycle state of the cloud sync.
	LifecycleState LifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// One of the following cloud sync-specific lifecycle substates:
	// *  NONE
	// *  CREATED
	// *  RUN
	// *  RUNNING
	// *  COMPLETED
	// *  FAILED
	// *  CANCELING
	// *  CANCELED
	// *  UPDATING
	LifecycleDetails *string `mandatory:"true" json:"lifecycleDetails"`

	// A description of the cloud sync. It does not have to be unique, and it can be changed.
	// Example: `my first cloud sync`
	Description *string `mandatory:"false" json:"description"`

	// The date and time the cloud sync was completed, canceled, or failed, in the format defined by RFC3339.
	// Example: `2019-05-16T23:30:30.793Z`
	TimeEnded *common.SDKTime `mandatory:"false" json:"timeEnded"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information,
	// see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m CloudSync) String() string {
	return common.PointerString(m)
}
