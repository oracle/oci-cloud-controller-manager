// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CopyBootVolumeBackupDetails The representation of CopyBootVolumeBackupDetails
type CopyBootVolumeBackupDetails struct {

	// The name of the destination region.
	// Example: `us-ashburn-1`
	DestinationRegion *string `mandatory:"true" json:"destinationRegion"`

	// A user-friendly name for the boot volume backup. Does not have to be unique and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The OCID of the KMS key in the destination region which will be the master encryption key
	// for the copied boot volume backup.
	// Required when copying a boot volume backup taken from a boot volume using a KMS key as its master encryption key.
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`
}

func (m CopyBootVolumeBackupDetails) String() string {
	return common.PointerString(m)
}
