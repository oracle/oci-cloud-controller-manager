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

// Catalog Details of Catalog resource
type Catalog struct {

	// OCID of the Catalog Instance
	Id *string `mandatory:"true" json:"id"`

	// Compartment Identifier
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Catalog Identifier, can be renamed
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The time the the Catalog was created. An RFC3339 formatted datetime string
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The time the Catalog was updated. An RFC3339 formatted datetime string
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// The REST front endpoint url to the catalog instance
	ServiceApiUrl *string `mandatory:"false" json:"serviceApiUrl"`

	// The console front endpoint url to the catalog instance
	ServiceConsoleUrl *string `mandatory:"false" json:"serviceConsoleUrl"`

	// The number of data objects added to the Catalog.
	// Please see the Data Catalog documentation for further information on how this is calculated.
	NumberOfObjects *int `mandatory:"false" json:"numberOfObjects"`

	// The current state of the catalog resource.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// An message describing the current state in more detail.
	// For example, can be used to provide actionable information for a resource in Failed state.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// Simple key-value pair that is applied without any predefined name, type or scope. Exists for cross-compatibility only.
	// Example: `{"bar-key": "value"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Usage of predefined tag keys. These predefined keys are scoped to namespaces.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m Catalog) String() string {
	return common.PointerString(m)
}
