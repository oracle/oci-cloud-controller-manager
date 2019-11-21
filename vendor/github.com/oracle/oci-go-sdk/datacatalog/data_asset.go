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

// DataAsset Data Asset representation. A physical store, or stream, of data known to the Catalog and containing
// one or many Data Entities, possibly in an organized structure of Folders. A Data Asset is often synonymous
// with a 'System', such as a Database, or may be a file container, or a message stream.
type DataAsset struct {

	// Unique Data Asset key that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Detailed description of the Data Asset.
	Description *string `mandatory:"false" json:"description"`

	// The Catalog's Oracle ID (OCID).
	CatalogId *string `mandatory:"false" json:"catalogId"`

	// External uri which can be used to reference the object. Format will differ based on the type of object.
	ExternalKey *string `mandatory:"false" json:"externalKey"`

	// The key of the object type. Type key's can be found via the '/types' endpoint.
	TypeKey *string `mandatory:"false" json:"typeKey"`

	// The current state of the Data Asset.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the DataAsset was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The last time that any change was made to the Data Asset. An RFC3339 formatted datetime string.
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// Id (OCID) of the user who created the Data Asset.
	CreatedById *string `mandatory:"false" json:"createdById"`

	// Id (OCID) of the user who last modified the Data Asset.
	UpdatedById *string `mandatory:"false" json:"updatedById"`

	// URI to the Data Asset instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// A map of maps which contains the properties which are specific to the asset type. Each Data Asset type
	// definition defines it's set of required and optional properties. The map keys are category names and the
	// values are maps of property name to property value. Every property is contained inside of a category. Most
	// Data Assets have required properties within the "default" category.
	// Example: `{"properties": { "default": { "host": "host1", "port": "1521", "database": "orcl"}}}`
	Properties map[string]map[string]string `mandatory:"false" json:"properties"`
}

func (m DataAsset) String() string {
	return common.PointerString(m)
}
