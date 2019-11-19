// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// KAM API
//
// description: |
//   Kubernetes Add-on Manager API for installing, uninstalling and upgrading
//   OKE add-ons or Marketplace applications on an OKE cluster
//

package kam

// ReleaseStateEnum Enum with underlying type: string
type ReleaseStateEnum string

// Set of constants representing the allowable values for ReleaseStateEnum
const (
	ReleaseStateCreating ReleaseStateEnum = "CREATING"
	ReleaseStateUpdating ReleaseStateEnum = "UPDATING"
	ReleaseStateActive   ReleaseStateEnum = "ACTIVE"
	ReleaseStateDeleting ReleaseStateEnum = "DELETING"
	ReleaseStateDeleted  ReleaseStateEnum = "DELETED"
	ReleaseStateFailed   ReleaseStateEnum = "FAILED"
)

var mappingReleaseState = map[string]ReleaseStateEnum{
	"CREATING": ReleaseStateCreating,
	"UPDATING": ReleaseStateUpdating,
	"ACTIVE":   ReleaseStateActive,
	"DELETING": ReleaseStateDeleting,
	"DELETED":  ReleaseStateDeleted,
	"FAILED":   ReleaseStateFailed,
}

// GetReleaseStateEnumValues Enumerates the set of values for ReleaseStateEnum
func GetReleaseStateEnumValues() []ReleaseStateEnum {
	values := make([]ReleaseStateEnum, 0)
	for _, v := range mappingReleaseState {
		values = append(values, v)
	}
	return values
}
