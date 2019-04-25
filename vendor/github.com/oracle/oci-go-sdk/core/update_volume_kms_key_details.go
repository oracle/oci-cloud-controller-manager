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

// UpdateVolumeKmsKeyDetails The representation of UpdateVolumeKmsKeyDetails
type UpdateVolumeKmsKeyDetails struct {

	// The OCID of the new KMS key which will be used to protect the specified volume.
	// This key has to be a valid KMS key OCID, and the user must have key delegation policy to allow them to access this key.
	// Even if the new KMS key is the same as the previous KMS key ID, the Block Volume service will use it to regenerate a new volume encryption key.
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`
}

func (m UpdateVolumeKmsKeyDetails) String() string {
	return common.PointerString(m)
}
