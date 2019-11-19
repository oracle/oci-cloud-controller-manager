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

// FolderSummary Summary of a Folder.
// A generic term used in the Catalog for an external organization concept used for a collection of data entities
// or processes within a Data Asset. This term is an internal term which models multiple external types of folder,
// such as file directories, database schemas etc. Some Data Assets, such as Object Store containers,
// may contain many levels of folders.
type FolderSummary struct {

	// Unique Folder key that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Detailed description of a Folder.
	Description *string `mandatory:"false" json:"description"`

	// The unique key of the parent Data Asset.
	DataAssetKey *string `mandatory:"false" json:"dataAssetKey"`

	// The key of the containing folder or null if there is no parent.
	ParentFolderKey *string `mandatory:"false" json:"parentFolderKey"`

	// Full path of the folder.
	Path *string `mandatory:"false" json:"path"`

	// Unique external key of this object from the source systems
	ExternalKey *string `mandatory:"false" json:"externalKey"`

	// Last modified timestamp of this object in the external system
	TimeExternal *common.SDKTime `mandatory:"false" json:"timeExternal"`

	// The date and time the Folder was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// URI of the Folder resource within the Data Catalog API.
	Uri *string `mandatory:"false" json:"uri"`

	// State of the Folder.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`
}

func (m FolderSummary) String() string {
	return common.PointerString(m)
}
