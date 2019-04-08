// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Streaming Service API
//
// The API for the Streaming Service.
//

package streaming

import (
	"github.com/oracle/oci-go-sdk/common"
)

// UpdateArchiverDetails The update stream archiver parameters.
type UpdateArchiverDetails struct {

	// The namespace of the bucket.
	BucketNamespace *string `mandatory:"false" json:"bucketNamespace"`

	// The name of the bucket.
	BucketName *string `mandatory:"false" json:"bucketName"`

	// The flag to create a new bucket or use existing one.
	UseExistingBucket *bool `mandatory:"false" json:"useExistingBucket"`

	// The start message.
	StartPosition UpdateArchiverDetailsStartPositionEnum `mandatory:"false" json:"startPosition,omitempty"`

	// The batch rollover size in bytes.
	BatchRolloverSize *int `mandatory:"false" json:"batchRolloverSize"`

	// The rollover time in milliseconds.
	BatchRolloverTime *int `mandatory:"false" json:"batchRolloverTime"`
}

func (m UpdateArchiverDetails) String() string {
	return common.PointerString(m)
}

// UpdateArchiverDetailsStartPositionEnum Enum with underlying type: string
type UpdateArchiverDetailsStartPositionEnum string

// Set of constants representing the allowable values for UpdateArchiverDetailsStartPositionEnum
const (
	UpdateArchiverDetailsStartPositionLatest      UpdateArchiverDetailsStartPositionEnum = "LATEST"
	UpdateArchiverDetailsStartPositionTrimHorizon UpdateArchiverDetailsStartPositionEnum = "TRIM_HORIZON"
)

var mappingUpdateArchiverDetailsStartPosition = map[string]UpdateArchiverDetailsStartPositionEnum{
	"LATEST":       UpdateArchiverDetailsStartPositionLatest,
	"TRIM_HORIZON": UpdateArchiverDetailsStartPositionTrimHorizon,
}

// GetUpdateArchiverDetailsStartPositionEnumValues Enumerates the set of values for UpdateArchiverDetailsStartPositionEnum
func GetUpdateArchiverDetailsStartPositionEnumValues() []UpdateArchiverDetailsStartPositionEnum {
	values := make([]UpdateArchiverDetailsStartPositionEnum, 0)
	for _, v := range mappingUpdateArchiverDetailsStartPosition {
		values = append(values, v)
	}
	return values
}
