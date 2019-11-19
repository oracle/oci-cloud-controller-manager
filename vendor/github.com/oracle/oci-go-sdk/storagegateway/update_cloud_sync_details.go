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

// UpdateCloudSyncDetails Configuration details for updating a cloud sync.
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type UpdateCloudSyncDetails struct {

	// A description of the cloud sync. It does not have to be unique, and it is changeable.
	// Example: `my first cloud sync`
	Description *string `mandatory:"false" json:"description"`

	// Whether the cloud sync automatically deletes files from the target when source files are renamed or deleted.
	// If "true", the cloud sync automatically deletes files from the target.
	// Example: `true`
	IsAutoDeletionEnabled *bool `mandatory:"false" json:"isAutoDeletionEnabled"`

	// The path to a file that lists a set of files to sync to the target. If you do not specify a file list, the
	// service syncs all files. The list file should reside under the `/cloudsync/` directory on the machine running
	// the storage gateway instance.
	// Example: `/cloudsync/files.list`
	FilesFrom *string `mandatory:"false" json:"filesFrom"`

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

func (m UpdateCloudSyncDetails) String() string {
	return common.PointerString(m)
}
