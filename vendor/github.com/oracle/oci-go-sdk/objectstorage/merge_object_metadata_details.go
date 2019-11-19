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

// MergeObjectMetadataDetails To merge Objects User metadata with existing metadata we specify the new metadata in the body.
type MergeObjectMetadataDetails struct {

	// Arbitrary string keys-values pair for the user-defined metadata for the object to be merged with its current
	// metadata. Keys must be in "opc-meta-*" format. Avoid entering confidential information.
	// The size of user-defined metadata is measured by taking the sum of the number of bytes in the UTF-8 encoding
	// of each key and value. The maximum metadata size after merge is 2975 bytes.
	Metadata map[string]string `mandatory:"true" json:"metadata"`
}

func (m MergeObjectMetadataDetails) String() string {
	return common.PointerString(m)
}
