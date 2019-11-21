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

// ConnectionSummary Summary representation of a connection to a Data Asset.
type ConnectionSummary struct {

	// Unique connection key that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// A description of the connection.
	Description *string `mandatory:"false" json:"description"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The unique key of the parent Data Asset.
	DataAssetKey *string `mandatory:"false" json:"dataAssetKey"`

	// The key of the object type. Type key's can be found via the '/types' endpoint.
	TypeKey *string `mandatory:"false" json:"typeKey"`

	// URI to the Connection instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// Unique external key for this object as defined in the source systems
	ExternalKey *string `mandatory:"false" json:"externalKey"`

	// The current state of the connection.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Indicates whether this connection is the default connection.
	IsDefault *bool `mandatory:"false" json:"isDefault"`

	// The date and time the Connection was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m ConnectionSummary) String() string {
	return common.PointerString(m)
}
