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

// CreateArchiverDetails Represents the parameters of the stream archiver.
type CreateArchiverDetails struct {

	// The namespace of the bucket.
	BucketNamespace *string `mandatory:"true" json:"bucketNamespace"`

	// The name of the bucket.
	BucketName *string `mandatory:"true" json:"bucketName"`

	// The flag to create a new bucket or use existing one.
	UseExistingBucket *bool `mandatory:"true" json:"useExistingBucket"`

	// The start message.
	StartPosition CreateArchiverDetailsStartPositionEnum `mandatory:"true" json:"startPosition"`

	// The batch rollover size in bytes.
	BatchRolloverSize *int `mandatory:"true" json:"batchRolloverSize"`

	// The rollover time in milliseconds.
	BatchRolloverTime *int `mandatory:"true" json:"batchRolloverTime"`
}

func (m CreateArchiverDetails) String() string {
	return common.PointerString(m)
}

// CreateArchiverDetailsStartPositionEnum Enum with underlying type: string
type CreateArchiverDetailsStartPositionEnum string

// Set of constants representing the allowable values for CreateArchiverDetailsStartPositionEnum
const (
	CreateArchiverDetailsStartPositionLatest      CreateArchiverDetailsStartPositionEnum = "LATEST"
	CreateArchiverDetailsStartPositionTrimHorizon CreateArchiverDetailsStartPositionEnum = "TRIM_HORIZON"
)

var mappingCreateArchiverDetailsStartPosition = map[string]CreateArchiverDetailsStartPositionEnum{
	"LATEST":       CreateArchiverDetailsStartPositionLatest,
	"TRIM_HORIZON": CreateArchiverDetailsStartPositionTrimHorizon,
}

// GetCreateArchiverDetailsStartPositionEnumValues Enumerates the set of values for CreateArchiverDetailsStartPositionEnum
func GetCreateArchiverDetailsStartPositionEnumValues() []CreateArchiverDetailsStartPositionEnum {
	values := make([]CreateArchiverDetailsStartPositionEnum, 0)
	for _, v := range mappingCreateArchiverDetailsStartPosition {
		values = append(values, v)
	}
	return values
}
