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

// DataAssetSummary Summary of a Data Asset. A physical store, or stream, of data known to the Catalog and containing one or
// many Data Entities, possibly in an organized structure of Folders. A Data Asset is often synonymous with
// a 'System', such as a Database, or may be a file container, or a message stream.
type DataAssetSummary struct {

	// Unique Data Asset key which is immutable.
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

	// URI to the Data Asset instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// The date and time the DataAsset was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The key of the object type. Type keys's can be found via the '/types' endpoint.
	TypeKey *string `mandatory:"false" json:"typeKey"`

	// State of the Data Asset.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`
}

func (m DataAssetSummary) String() string {
	return common.PointerString(m)
}
