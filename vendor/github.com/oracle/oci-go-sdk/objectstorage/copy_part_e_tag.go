// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Object Storage Service API
//
// Common set of Object Storage and Archive Storage APIs for managing buckets, objects, and related resources.
//

package objectstorage

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CopyPartETag The representation of CopyPartETag
type CopyPartETag struct {

	// The entity tag (ETag) of the new part.
	ETag *string `mandatory:"true" json:"ETag"`
}

func (m CopyPartETag) String() string {
	return common.PointerString(m)
}
