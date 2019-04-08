// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Object Storage Service API
//
// Common set of Object Storage and Archive Storage APIs for managing buckets, objects, and related resources.
//

package objectstorage

import (
	"github.com/oracle/oci-go-sdk/common"
)

// PostObjectMetadataResponse Updated object information after user-metadata update.
type PostObjectMetadataResponse struct {

	// The new entity tag (ETag) for the object.
	ETag *string `mandatory:"true" json:"ETag"`

	// The time the object was modified, as described in RFC 2616 (https://tools.ietf.org/rfc/rfc2616), section 14.29.
	TimeModified *common.SDKTime `mandatory:"true" json:"timeModified"`
}

func (m PostObjectMetadataResponse) String() string {
	return common.PointerString(m)
}
