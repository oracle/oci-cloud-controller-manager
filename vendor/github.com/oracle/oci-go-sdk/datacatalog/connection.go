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

// Connection Detailed representation of a connection to a Data Asset, minus any sensitive properties.
type Connection struct {

	// Unique connection key that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// A description of the connection.
	Description *string `mandatory:"false" json:"description"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The date and time the Connection was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The last time that any change was made to the Connection. An RFC3339 formatted datetime string.
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// Id (OCID) of the user who created the Connection.
	CreatedById *string `mandatory:"false" json:"createdById"`

	// Id (OCID) of the user who modified the Connection.
	UpdatedById *string `mandatory:"false" json:"updatedById"`

	// A map of maps which contains the properties which are specific to the connection type. Each connection type
	// definition defines it's set of required and optional properties. The map keys are category names and the
	// values are maps of property name to property value. Every property is contained inside of a category. Most
	// connections have required properties within the "default" category.
	// Example: `{"properties": { "default": { "username": "user1"}}}`
	Properties map[string]map[string]string `mandatory:"false" json:"properties"`

	// Unique external key of this object from the source system
	ExternalKey *string `mandatory:"false" json:"externalKey"`

	// Time that the connections status was last updated. An RFC3339 formatted datetime string.
	TimeStatusUpdated *common.SDKTime `mandatory:"false" json:"timeStatusUpdated"`

	// The current state of the connection.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Indicates whether this connection is the default connection.
	IsDefault *bool `mandatory:"false" json:"isDefault"`

	// Unique key of the parent Data Asset.
	DataAssetKey *string `mandatory:"false" json:"dataAssetKey"`

	// The key of the object type. Type key's can be found via the '/types' endpoint.
	TypeKey *string `mandatory:"false" json:"typeKey"`

	// URI to the Connection instance in the API.
	Uri *string `mandatory:"false" json:"uri"`
}

func (m Connection) String() string {
	return common.PointerString(m)
}
