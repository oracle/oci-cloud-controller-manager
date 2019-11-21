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

// JobDefinitionSummary A List of Job Definition Resources. Job Definitions define the harvest scope and includes the list of objects
// to be harvested along with a schedule. The list of objects is usually specified through a combination of object
// type, regular expressions or specific names of objects and a sample size for the data harvested.
type JobDefinitionSummary struct {

	// Unique key of the Job Definition resource that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Detailed description of the Job Definition.
	Description *string `mandatory:"false" json:"description"`

	// The Catalog's Oracle ID (OCID).
	CatalogId *string `mandatory:"false" json:"catalogId"`

	// URI to the Job Definition instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// Type of the Job Definition.
	JobType JobTypeEnum `mandatory:"false" json:"jobType,omitempty"`

	// Lifecycle state of the Job Definition.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Specify if sample data to be extracted as part of this harvest
	IsSampleDataExtracted *bool `mandatory:"false" json:"isSampleDataExtracted"`

	// The date and time the JobDefinition was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m JobDefinitionSummary) String() string {
	return common.PointerString(m)
}
