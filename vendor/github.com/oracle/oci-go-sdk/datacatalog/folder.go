// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DataCatalog API
//
// A description of the DataCatalog API
//

package datacatalog

import (
	"github.com/oracle/oci-go-sdk/common"
)

// Folder A generic term used in the Catalog for an external organization concept used for a collection of data entities
// or processes within a Data Asset. This term is an internal term which models multiple external types of folder,
// such as file directories, database schemas etc. Some Data Assets, such as Object Store containers, may contain
// many levels of folders.
type Folder struct {

	// Unique Folder key that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Detailed description of a Folder.
	Description *string `mandatory:"false" json:"description"`

	// The unique key of the containing folder or null if there is no parent folder.
	ParentFolderKey *string `mandatory:"false" json:"parentFolderKey"`

	// Full path of the folder.
	Path *string `mandatory:"false" json:"path"`

	// The key of the associated Data Asset.
	DataAssetKey *string `mandatory:"false" json:"dataAssetKey"`

	// A map of maps which contains the properties which are specific to the folder type. Each folder type
	// definition defines it's set of required and optional properties. The map keys are category names and the
	// values are maps of property name to property value. Every property is contained inside of a category. Most
	// folders have required properties within the "default" category.
	// Example: `{"properties": { "default": { "key1": "value1"}}}`
	Properties map[string]map[string]string `mandatory:"false" json:"properties"`

	// Unique External key of this object in the source system
	ExternalKey *string `mandatory:"false" json:"externalKey"`

	// The date and time the Folder was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The last time that any change was made to the Folder. An RFC3339 formatted datetime string.
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// Id (OCID) of the user who created the Folder
	CreatedById *string `mandatory:"false" json:"createdById"`

	// Id (OCID) of the user who modified the Folder
	UpdatedById *string `mandatory:"false" json:"updatedById"`

	// Last modified timestamp of this object in the external system
	TimeExternal *common.SDKTime `mandatory:"false" json:"timeExternal"`

	// The current state of the Folder.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Status of the object as updated by the harvest process
	HarvestStatus HarvestStatusEnum `mandatory:"false" json:"harvestStatus,omitempty"`

	// The key of the last harvest process to update the metadata of this object
	LastJobKey *string `mandatory:"false" json:"lastJobKey"`

	// URI to the Folder instance in the API.
	Uri *string `mandatory:"false" json:"uri"`
}

func (m Folder) String() string {
	return common.PointerString(m)
}
