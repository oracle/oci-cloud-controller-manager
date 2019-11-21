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

// JobDefinition Representation of a Job Definition Resource. Job Definitions define the harvest scope and includes the list
// of objects to be harvested along with a schedule. The list of objects is usually specified through a combination
// of object type, regular expressions or specific names of objects and a sample size for the data harvested.
type JobDefinition struct {

	// Unique key of the Job Definition resource that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The Catalog's Oracle ID (OCID).
	CatalogId *string `mandatory:"false" json:"catalogId"`

	// Type of the Job Definition.
	JobType JobTypeEnum `mandatory:"false" json:"jobType,omitempty"`

	// Specifies if the Job Definition is incremental or full.
	IsIncremental *bool `mandatory:"false" json:"isIncremental"`

	// The key of the Data Asset for which the job is defined
	DataAssetKey *string `mandatory:"false" json:"dataAssetKey"`

	// Detailed description of the Job Definition.
	Description *string `mandatory:"false" json:"description"`

	// The key of the default connection resource to be used for harvest , sampling , profiling  jobs.
	// This may be overridden in each Job instance.
	ConnectionKey *string `mandatory:"false" json:"connectionKey"`

	// Version of the job definition object. Used internally but can be visible to users
	InternalVersion *string `mandatory:"false" json:"internalVersion"`

	// Lifecycle state of the Job Definition.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the JobDefinition was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The last time that any change was made to the Data Asset. An RFC3339 formatted datetime string.
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// Id (OCID) of the user who created this Job Definition.
	CreatedById *string `mandatory:"false" json:"createdById"`

	// Id (OCID) of the user who updated this Job Definition.
	UpdatedById *string `mandatory:"false" json:"updatedById"`

	// URI to the JobDefinition instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// Specify if sample data to be extracted as part of this harvest
	IsSampleDataExtracted *bool `mandatory:"false" json:"isSampleDataExtracted"`

	// Specify the sample data size in MB, specified as number of rows, for this metadata harvest
	SampleDataSizeInMBs *int `mandatory:"false" json:"sampleDataSizeInMBs"`

	// A map of maps which contains the properties which are specific to the job type. Each job type
	// definition may define it's set of required and optional properties. The map keys are category names and the
	// values are maps of property name to property value. Every property is contained inside of a category. Most
	// job definitions have required properties within the "default" category.
	// Example: `{"properties": { "default": { "host": "host1", "port": "1521", "database": "orcl"}}}`
	Properties map[string]map[string]string `mandatory:"false" json:"properties"`
}

func (m JobDefinition) String() string {
	return common.PointerString(m)
}
