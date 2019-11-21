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

// UpdateFileSystemDetails Configuration details for updating a storage gateway file system.
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type UpdateFileSystemDetails struct {

	// A description of the file system. It does not have to be unique, and it can be changed.
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

func (m UpdateFileSystemDetails) String() string {
	return common.PointerString(m)
}
