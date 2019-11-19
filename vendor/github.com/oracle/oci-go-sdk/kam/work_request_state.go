// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// KAM API
//
// description: |
//   Kubernetes Add-on Manager API for installing, uninstalling and upgrading
//   OKE add-ons or Marketplace applications on an OKE cluster
//

package kam

// WorkRequestStateEnum Enum with underlying type: string
type WorkRequestStateEnum string

// Set of constants representing the allowable values for WorkRequestStateEnum
const (
	WorkRequestStateAccepted   WorkRequestStateEnum = "ACCEPTED"
	WorkRequestStateInProgress WorkRequestStateEnum = "IN_PROGRESS"
	WorkRequestStateFailed     WorkRequestStateEnum = "FAILED"
	WorkRequestStateSucceeded  WorkRequestStateEnum = "SUCCEEDED"
	WorkRequestStateCanceling  WorkRequestStateEnum = "CANCELING"
	WorkRequestStateCanceled   WorkRequestStateEnum = "CANCELED"
)

var mappingWorkRequestState = map[string]WorkRequestStateEnum{
	"ACCEPTED":    WorkRequestStateAccepted,
	"IN_PROGRESS": WorkRequestStateInProgress,
	"FAILED":      WorkRequestStateFailed,
	"SUCCEEDED":   WorkRequestStateSucceeded,
	"CANCELING":   WorkRequestStateCanceling,
	"CANCELED":    WorkRequestStateCanceled,
}

// GetWorkRequestStateEnumValues Enumerates the set of values for WorkRequestStateEnum
func GetWorkRequestStateEnumValues() []WorkRequestStateEnum {
	values := make([]WorkRequestStateEnum, 0)
	for _, v := range mappingWorkRequestState {
		values = append(values, v)
	}
	return values
}
