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

// BackupLocationUri PreAuthenticated object storage URI to upload or download the backup
type BackupLocationUri struct {
	Uri *string `mandatory:"true" json:"uri"`
}

func (m BackupLocationUri) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m BackupLocationUri) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeBackupLocationUri BackupLocationUri
	s := struct {
		DiscriminatorParam string `json:"destination"`
		MarshalTypeBackupLocationUri
	}{
		"PRE_AUTHENTICATED_REQUEST_URI",
		(MarshalTypeBackupLocationUri)(m),
	}

	return json.Marshal(&s)
}
