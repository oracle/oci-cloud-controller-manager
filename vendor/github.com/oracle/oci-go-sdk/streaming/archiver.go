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

// Archiver Represents the current state of the stream archiver.
type Archiver struct {

	// The archiver user identifier.
	UserId *string `mandatory:"false" json:"userId"`

	// The archiver group identifier.
	GroupId *string `mandatory:"false" json:"groupId"`

	// Time when the resource was created.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The state of the stream archiver.
	LifecycleState ArchiverLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The namespace of the bucket.
	BucketNamespace *string `mandatory:"false" json:"bucketNamespace"`

	// The name of the bucket.
	BucketName *string `mandatory:"false" json:"bucketName"`

	// The flag to create a new bucket or use existing one.
	UseExistingBucket *bool `mandatory:"false" json:"useExistingBucket"`

	// The start message.
	StartPosition ArchiverStartPositionEnum `mandatory:"false" json:"startPosition,omitempty"`

	// The batch rollover size in bytes.
	BatchRolloverSize *int `mandatory:"false" json:"batchRolloverSize"`

	// The rollover time in milliseconds.
	BatchRolloverTime *int `mandatory:"false" json:"batchRolloverTime"`
}

func (m Archiver) String() string {
	return common.PointerString(m)
}

// ArchiverLifecycleStateEnum Enum with underlying type: string
type ArchiverLifecycleStateEnum string

// Set of constants representing the allowable values for ArchiverLifecycleStateEnum
const (
	ArchiverLifecycleStateCreating ArchiverLifecycleStateEnum = "CREATING"
	ArchiverLifecycleStateStopped  ArchiverLifecycleStateEnum = "STOPPED"
	ArchiverLifecycleStateStarting ArchiverLifecycleStateEnum = "STARTING"
	ArchiverLifecycleStateRunning  ArchiverLifecycleStateEnum = "RUNNING"
	ArchiverLifecycleStateStopping ArchiverLifecycleStateEnum = "STOPPING"
	ArchiverLifecycleStateUpdating ArchiverLifecycleStateEnum = "UPDATING"
)

var mappingArchiverLifecycleState = map[string]ArchiverLifecycleStateEnum{
	"CREATING": ArchiverLifecycleStateCreating,
	"STOPPED":  ArchiverLifecycleStateStopped,
	"STARTING": ArchiverLifecycleStateStarting,
	"RUNNING":  ArchiverLifecycleStateRunning,
	"STOPPING": ArchiverLifecycleStateStopping,
	"UPDATING": ArchiverLifecycleStateUpdating,
}

// GetArchiverLifecycleStateEnumValues Enumerates the set of values for ArchiverLifecycleStateEnum
func GetArchiverLifecycleStateEnumValues() []ArchiverLifecycleStateEnum {
	values := make([]ArchiverLifecycleStateEnum, 0)
	for _, v := range mappingArchiverLifecycleState {
		values = append(values, v)
	}
	return values
}

// ArchiverStartPositionEnum Enum with underlying type: string
type ArchiverStartPositionEnum string

// Set of constants representing the allowable values for ArchiverStartPositionEnum
const (
	ArchiverStartPositionLatest      ArchiverStartPositionEnum = "LATEST"
	ArchiverStartPositionTrimHorizon ArchiverStartPositionEnum = "TRIM_HORIZON"
)

var mappingArchiverStartPosition = map[string]ArchiverStartPositionEnum{
	"LATEST":       ArchiverStartPositionLatest,
	"TRIM_HORIZON": ArchiverStartPositionTrimHorizon,
}

// GetArchiverStartPositionEnumValues Enumerates the set of values for ArchiverStartPositionEnum
func GetArchiverStartPositionEnumValues() []ArchiverStartPositionEnum {
	values := make([]ArchiverStartPositionEnum, 0)
	for _, v := range mappingArchiverStartPosition {
		values = append(values, v)
	}
	return values
}
