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

// Glossary Full Glossary details. A glossary of business terms, such as 'Customer', 'Account', 'Contact' , 'Address',
// 'Product' etc. with definitions, used to provide common meaning across disparate Data Assets. Business Glossaries
// may be hierarchical where some terms may contain child terms to allow them to be used as 'taxonomies'.
// By linking Data Assets, data entities and attributes to glossaries and glossary terms, the glossary can act as a
// way of organizing Catalog objects in a hierarchy to make a large number of objects more navigable and easier to
// consume. Objects in the Data Catalog, such as Data Assets or Data Entities, may be linked to any level in the
// Glossary, so that the Glossary can be used to browse the available data according to the business model of the
// organization.
type Glossary struct {

	// Unique glossary key that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Detailed description of the Glossary.
	Description *string `mandatory:"false" json:"description"`

	// The Catalog's Oracle ID (OCID).
	CatalogId *string `mandatory:"false" json:"catalogId"`

	// The current state of the Glossary.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the Glossary was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The last time that any change was made to the Glossary. An RFC3339 formatted datetime string.
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// Id (OCID) of the user who created this metadata element.
	CreatedById *string `mandatory:"false" json:"createdById"`

	// Id (OCID) of the user who updated this metadata element.
	UpdatedById *string `mandatory:"false" json:"updatedById"`

	// Id (OCID) of the user who is the owner of the glossary.
	Owner *string `mandatory:"false" json:"owner"`

	// URI to the Tag instance in the API.
	Uri *string `mandatory:"false" json:"uri"`
}

func (m Glossary) String() string {
	return common.PointerString(m)
}
