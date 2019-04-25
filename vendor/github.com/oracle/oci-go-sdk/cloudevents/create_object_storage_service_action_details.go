// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// CloudEvents API
//
// API for the CloudEvents Service. Use this API to manage rules and actions that create automation
// in your tenancy. For more information, see Overview of Events (https://docs.cloud.oracle.com/iaas/Content/Events/Concepts/eventsoverview.htm).
//

package cloudevents

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// CreateObjectStorageServiceActionDetails Create an action that delivers to an Oracle Object Storage bucket.
type CreateObjectStorageServiceActionDetails struct {

	// Whether or not this action is currently enabled.
	// Example: `true`
	IsEnabled *bool `mandatory:"true" json:"isEnabled"`

	// The Object Storage namespace in which the bucket lives.
	NamespaceName *string `mandatory:"true" json:"namespaceName"`

	// The name of the bucket.
	// Example: `my-new-bucket1`
	BucketName *string `mandatory:"true" json:"bucketName"`

	// A string that describes the details of the action. It does not have to be unique, and you can change it. Avoid entering
	// confidential information.
	Description *string `mandatory:"false" json:"description"`
}

//GetIsEnabled returns IsEnabled
func (m CreateObjectStorageServiceActionDetails) GetIsEnabled() *bool {
	return m.IsEnabled
}

//GetDescription returns Description
func (m CreateObjectStorageServiceActionDetails) GetDescription() *string {
	return m.Description
}

func (m CreateObjectStorageServiceActionDetails) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m CreateObjectStorageServiceActionDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeCreateObjectStorageServiceActionDetails CreateObjectStorageServiceActionDetails
	s := struct {
		DiscriminatorParam string `json:"actionType"`
		MarshalTypeCreateObjectStorageServiceActionDetails
	}{
		"OBJECTSTORAGE",
		(MarshalTypeCreateObjectStorageServiceActionDetails)(m),
	}

	return json.Marshal(&s)
}
