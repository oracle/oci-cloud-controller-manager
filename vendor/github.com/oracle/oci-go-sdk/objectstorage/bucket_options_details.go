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

// BucketOptionsDetails Internal options associated with a bucket.
// Arbitrary JSON keys and values associated with the bucket (for internal-use-only). Keys must be in
// "opc-meta-*" format.
// To use any of the API operations, you must be authorized in an IAM policy. If you are not authorized,
// talk to an administrator. If you are an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/Content/Identity/Concepts/policygetstarted.htm).
type BucketOptionsDetails struct {
	FreeformOptions *interface{} `mandatory:"true" json:"freeformOptions"`
}

func (m BucketOptionsDetails) String() string {
	return common.PointerString(m)
}
