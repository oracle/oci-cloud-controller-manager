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

// JobMetricSummary JobMetric summary.
type JobMetricSummary struct {

	// Key of the Job Metric which is immutable
	Key *string `mandatory:"true" json:"key"`

	// Detailed description of the metric.
	Description *string `mandatory:"false" json:"description"`

	// The unique key of the parent Job Execution for which the Job Metric resource was created.
	JobExecutionKey *string `mandatory:"false" json:"jobExecutionKey"`

	// URI to the Job Metric instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// The date and time the JobMetric was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The time the metric was logged or captured in the system where the job executed.
	// An RFC3339 formatted datetime string.
	TimeInserted *common.SDKTime `mandatory:"false" json:"timeInserted"`

	// Category of this metric
	Category *string `mandatory:"false" json:"category"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Sub Category of this metric under the category. Used for aggregating values. May be null.
	SubCategory *string `mandatory:"false" json:"subCategory"`

	// Unit of this metric
	Unit *string `mandatory:"false" json:"unit"`

	// Value of this metric
	Value *string `mandatory:"false" json:"value"`

	// Batch key for grouping, may be null.
	BatchKey *string `mandatory:"false" json:"batchKey"`
}

func (m JobMetricSummary) String() string {
	return common.PointerString(m)
}
