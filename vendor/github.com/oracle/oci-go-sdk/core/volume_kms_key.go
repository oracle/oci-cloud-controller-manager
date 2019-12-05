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

// VolumeKmsKey The KMS key OCID associated with this volume.
type VolumeKmsKey struct {

	// The KMS key OCID associated with this volume. If the volume is not using KMS, then the `kmsKeyId` will be a null string.
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`
}

func (m VolumeKmsKey) String() string {
	return common.PointerString(m)
}
