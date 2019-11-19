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

// Attribute Details of an Entity Attribute. An attribute of a data entity describing an item of data,
// with a name and data type. Synonymous with 'column' in a database.
type Attribute struct {

	// Unique Attribute key that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Detailed description of the Attribute.
	Description *string `mandatory:"false" json:"description"`

	// The unique key of the parent Entity.
	EntityKey *string `mandatory:"false" json:"entityKey"`

	// State of the Attribute.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the Attribute was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The last time that any change was made to the Attribute. An RFC3339 formatted datetime string.
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// Id (OCID) of the user who created this attribute in the catalog.
	CreatedById *string `mandatory:"false" json:"createdById"`

	// Id (OCID) of the user who modified this attribute in the catalog.
	UpdatedById *string `mandatory:"false" json:"updatedById"`

	// Data type of the attribute as defined in the external system. Type mapping across systems can be achieved
	// through term associations across domains in the ontology. The attribute can also be tagged to the datatype in
	// the domain ontology to resolve any ambiguity arising from type name similarity that can occur with user
	// defined types.
	ExternalDataType *string `mandatory:"false" json:"externalDataType"`

	// Unique external key of this attribute in the external source system
	ExternalKey *string `mandatory:"false" json:"externalKey"`

	// Property that identifies if this attribute can be used as a watermark to extract incremental data
	IsIncrementalData *bool `mandatory:"false" json:"isIncrementalData"`

	// Property that identifies if this attribute can be assigned null values
	IsNullable *bool `mandatory:"false" json:"isNullable"`

	// Max allowed length of the attribute value
	Length *int64 `mandatory:"false" json:"length"`

	// Position of the attribute in the record definition
	Position *int `mandatory:"false" json:"position"`

	// Precision of the attribute value usually applies to float data type
	Precision *int `mandatory:"false" json:"precision"`

	// Scale of the attribute value usually applies to float data type
	Scale *int `mandatory:"false" json:"scale"`

	// Last modified timestamp of this object in the external system
	TimeExternal *common.SDKTime `mandatory:"false" json:"timeExternal"`

	// URI to the Attribute instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// A map of maps which contains the properties which are specific to the Attribute type. Each Attribute type
	// definition defines it's set of required and optional properties. The map keys are category names and the
	// values are maps of property name to property value. Every property is contained inside of a category. Most
	// Attributes have required properties within the "default" category.
	// Example: `{"properties": { "default": { "key1": "value1"}}}`
	Properties map[string]map[string]string `mandatory:"false" json:"properties"`
}

func (m Attribute) String() string {
	return common.PointerString(m)
}
