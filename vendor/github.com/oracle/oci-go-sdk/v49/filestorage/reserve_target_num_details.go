// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// File Storage API
//
// API for the File Storage service. Use this API to manage file systems, mount targets, and snapshots. For more information, see Overview of File Storage (https://docs.cloud.oracle.com/iaas/Content/File/Concepts/filestorageoverview.htm).
//

package filestorage

import (
	"github.com/oracle/oci-go-sdk/v49/common"
)

// ReserveTargetNumDetails Details for creating a new replication target and resume snapshot num.
type ReserveTargetNumDetails struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the replication target id.
	ReplicationTargetId *string `mandatory:"false" json:"replicationTargetId"`

	// The resume snapshot num which represents point of replication continuinity.
	ResumeSnapshotNum *int64 `mandatory:"false" json:"resumeSnapshotNum"`
}

func (m ReserveTargetNumDetails) String() string {
	return common.PointerString(m)
}
