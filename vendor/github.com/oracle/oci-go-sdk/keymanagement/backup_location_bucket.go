// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Key Management Service API
//
// API for managing and performing operations with keys and vaults.
//

package keymanagement

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// BackupLocationBucket Object storage bucket details to upload or download the backup
type BackupLocationBucket struct {
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	Name *string `mandatory:"true" json:"name"`
}

func (m BackupLocationBucket) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m BackupLocationBucket) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeBackupLocationBucket BackupLocationBucket
	s := struct {
		DiscriminatorParam string `json:"destination"`
		MarshalTypeBackupLocationBucket
	}{
		"BUCKET",
		(MarshalTypeBackupLocationBucket)(m),
	}

	return json.Marshal(&s)
}
