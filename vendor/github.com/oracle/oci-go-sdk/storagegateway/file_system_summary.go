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

// FileSystemSummary Summary view of the specified file system.
type FileSystemSummary struct {

	// The file system name, which is unique within your tenancy.
	// Example: `file_system_52019`
	Name *string `mandatory:"true" json:"name"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the file system.
	Id *string `mandatory:"true" json:"id"`

	// The type of object storage tier used for data storage. The Standard tier is the default for data that
	// requires frequent and fast access.
	StorageTier FileSystemSummaryStorageTierEnum `mandatory:"true" json:"storageTier"`

	// Whether the object storage bucket is connected. If "true", the object storage bucket is connected.
	// Example: `true`
	IsConnected *bool `mandatory:"true" json:"isConnected"`

	// Whether the file system is in refresh mode. If "true", the file system is in refresh mode.
	// Example: `false`
	IsInRefreshMode *bool `mandatory:"true" json:"isInRefreshMode"`

	// The number of errors returned by the file system.
	// Example: `1`
	ErrorCount *float32 `mandatory:"true" json:"errorCount"`

	// The number of warnings returned by the file system.
	// Example: `3`
	WarnCount *float32 `mandatory:"true" json:"warnCount"`

	// The date and time the file system was created, in the format defined by RFC3339.
	// Example: `2019-05-16T21:52:40.793Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The current lifecycle state of the file system. You cannot use the file system before the state is ACTIVE.
	// When you disconnect a file system, its lifecycle state changes to INACTIVE.
	LifecycleState LifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// One of the following file system-specific lifecycle substates:
	// *  `NONE`
	// *  `CONNECTING`
	// *  `DISCONNECTING`
	// *  `RECLAIMING`
	// *  `REFRESHING`
	// *  `UPDATING`
	LifecycleDetails *string `mandatory:"true" json:"lifecycleDetails"`

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

func (m FileSystemSummary) String() string {
	return common.PointerString(m)
}

// FileSystemSummaryStorageTierEnum Enum with underlying type: string
type FileSystemSummaryStorageTierEnum string

// Set of constants representing the allowable values for FileSystemSummaryStorageTierEnum
const (
	FileSystemSummaryStorageTierStandard FileSystemSummaryStorageTierEnum = "STANDARD"
	FileSystemSummaryStorageTierArchive  FileSystemSummaryStorageTierEnum = "ARCHIVE"
)

var mappingFileSystemSummaryStorageTier = map[string]FileSystemSummaryStorageTierEnum{
	"STANDARD": FileSystemSummaryStorageTierStandard,
	"ARCHIVE":  FileSystemSummaryStorageTierArchive,
}

// GetFileSystemSummaryStorageTierEnumValues Enumerates the set of values for FileSystemSummaryStorageTierEnum
func GetFileSystemSummaryStorageTierEnumValues() []FileSystemSummaryStorageTierEnum {
	values := make([]FileSystemSummaryStorageTierEnum, 0)
	for _, v := range mappingFileSystemSummaryStorageTier {
		values = append(values, v)
	}
	return values
}
