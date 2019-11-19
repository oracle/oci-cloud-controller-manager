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

// FileSystem The configuration details of a file system. For general information about storage gateway file systems, see the
// "How Storage Gateway Works" section of
// Overview of Storage Gateway (https://docs.cloud.oracle.com/iaas/Content/StorageGateway/Concepts/storagegatewayoverview.htm#howitworks).
type FileSystem struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the storage gateway.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment containing the file system.
	StorageGatewayId *string `mandatory:"true" json:"storageGatewayId"`

	// The name of the file system. It must be unique, and it cannot be changed.
	// Example: `file_system_52019`
	Name *string `mandatory:"true" json:"name"`

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the file system.
	Id *string `mandatory:"true" json:"id"`

	// The type of object storage tier used for data storage. The Standard tier is the default for data that
	// requires frequent and fast access.
	StorageTier FileSystemStorageTierEnum `mandatory:"true" json:"storageTier"`

	// Whether the object storage bucket is connected. If "true", the object storage bucket is connected.
	// Example: `true`
	IsConnected *bool `mandatory:"true" json:"isConnected"`

	// Whether the file system is in refresh mode. If "true", the file system is in refresh mode.
	// Example: `false`
	IsInRefreshMode *bool `mandatory:"true" json:"isInRefreshMode"`

	// A list of hosts allowed to connect to the NFS export. The list is comma-separated and whitespace is optional.
	// Specify `*` to allow all hosts to connect.
	// Example: `2001:db8:9:e54::/64, 192.168.2.0/24`
	NfsAllowedHosts *string `mandatory:"true" json:"nfsAllowedHosts"`

	// The NFS export options.
	// Do not specify the `fsid` option.
	// Example: `rw, sync, insecure, no_subtree_check, no_root_squash`
	NfsExportOptions *string `mandatory:"true" json:"nfsExportOptions"`

	// The date and time the file system was created, in the format defined by RFC3339.
	// Example: `2019-05-16T21:52:40.793Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The current lifecycle state of the file system.
	LifecycleState LifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// One of the following file system-specific lifecycle substates:
	// *  `NONE`
	// *  `CONNECTING`
	// *  `DISCONNECTING`
	// *  `RECLAIMING`
	// *  `REFRESHING`
	// *  `UPDATING`
	LifecycleDetails *string `mandatory:"true" json:"lifecycleDetails"`

	// A description of the file system. It does not have to be unique, and it can be changed.
	//  Example: `my first storage gateway file system`
	Description *string `mandatory:"false" json:"description"`

	// Whether to reclaim an object storage bucket owned by another file system. When set to "true", the file system
	// attempts to reclaim the bucket.
	// Example: `true`
	IsReclaimAttempt *bool `mandatory:"false" json:"isReclaimAttempt"`

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

func (m FileSystem) String() string {
	return common.PointerString(m)
}

// FileSystemStorageTierEnum Enum with underlying type: string
type FileSystemStorageTierEnum string

// Set of constants representing the allowable values for FileSystemStorageTierEnum
const (
	FileSystemStorageTierStandard FileSystemStorageTierEnum = "STANDARD"
	FileSystemStorageTierArchive  FileSystemStorageTierEnum = "ARCHIVE"
)

var mappingFileSystemStorageTier = map[string]FileSystemStorageTierEnum{
	"STANDARD": FileSystemStorageTierStandard,
	"ARCHIVE":  FileSystemStorageTierArchive,
}

// GetFileSystemStorageTierEnumValues Enumerates the set of values for FileSystemStorageTierEnum
func GetFileSystemStorageTierEnumValues() []FileSystemStorageTierEnum {
	values := make([]FileSystemStorageTierEnum, 0)
	for _, v := range mappingFileSystemStorageTier {
		values = append(values, v)
	}
	return values
}
