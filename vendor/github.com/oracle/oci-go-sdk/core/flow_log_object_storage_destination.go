// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// FlowLogObjectStorageDestination Information to identify the Object Storage bucket where the flow logs will be stored.
type FlowLogObjectStorageDestination struct {

	// The Object Storage bucket name.
	BucketName *string `mandatory:"false" json:"bucketName"`

	// The Object Storage namespace.
	NamespaceName *string `mandatory:"false" json:"namespaceName"`
}

func (m FlowLogObjectStorageDestination) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m FlowLogObjectStorageDestination) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeFlowLogObjectStorageDestination FlowLogObjectStorageDestination
	s := struct {
		DiscriminatorParam string `json:"destinationType"`
		MarshalTypeFlowLogObjectStorageDestination
	}{
		"OBJECT_STORAGE",
		(MarshalTypeFlowLogObjectStorageDestination)(m),
	}

	return json.Marshal(&s)
}
