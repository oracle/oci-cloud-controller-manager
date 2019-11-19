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

// CreateFileSystemDetails The configuration details for creating a storage gateway file system.
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type CreateFileSystemDetails struct {

	// A name for the file system. It must be unique within your tenancy, and it cannot be changed. If an
	// object storage bucket matching the file system name does not exist, it will be created.
	// Example: `file_system_52019`
	Name *string `mandatory:"true" json:"name"`

	// The type of object storage tier used for data storage. The Standard tier is the default for data that
	// requires frequent and fast access.
	StorageTier CreateFileSystemDetailsStorageTierEnum `mandatory:"true" json:"storageTier"`

	// A description of the storage gateway file system. It does not have to be unique, and it is changeable.
	// Example: `my first storage gateway file system`
	Description *string `mandatory:"false" json:"description"`

	// A list of hosts allowed to connect to the NFS export. The list is comma-separated and whitespace is optional.
	// Specify `*` to allow all hosts to connect.
	// Example: `2001:db8:9:e54::/64, 192.168.2.0/24`
	NfsAllowedHosts *string `mandatory:"false" json:"nfsAllowedHosts"`

	// The NFS export options.
	// Do not specify the `fsid` option.
	// Example: `rw, sync, insecure, no_subtree_check, no_root_squash`
	NfsExportOptions *string `mandatory:"false" json:"nfsExportOptions"`

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

func (m CreateFileSystemDetails) String() string {
	return common.PointerString(m)
}

// CreateFileSystemDetailsStorageTierEnum Enum with underlying type: string
type CreateFileSystemDetailsStorageTierEnum string

// Set of constants representing the allowable values for CreateFileSystemDetailsStorageTierEnum
const (
	CreateFileSystemDetailsStorageTierStandard CreateFileSystemDetailsStorageTierEnum = "STANDARD"
	CreateFileSystemDetailsStorageTierArchive  CreateFileSystemDetailsStorageTierEnum = "ARCHIVE"
)

var mappingCreateFileSystemDetailsStorageTier = map[string]CreateFileSystemDetailsStorageTierEnum{
	"STANDARD": CreateFileSystemDetailsStorageTierStandard,
	"ARCHIVE":  CreateFileSystemDetailsStorageTierArchive,
}

// GetCreateFileSystemDetailsStorageTierEnumValues Enumerates the set of values for CreateFileSystemDetailsStorageTierEnum
func GetCreateFileSystemDetailsStorageTierEnumValues() []CreateFileSystemDetailsStorageTierEnum {
	values := make([]CreateFileSystemDetailsStorageTierEnum, 0)
	for _, v := range mappingCreateFileSystemDetailsStorageTier {
		values = append(values, v)
	}
	return values
}
