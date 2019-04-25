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

// CreateNoOpActionDetails Create an action that writes to a log.
type CreateNoOpActionDetails struct {

	// Whether or not this action is currently enabled.
	// Example: `true`
	IsEnabled *bool `mandatory:"true" json:"isEnabled"`

	// A string that describes the details of the action. It does not have to be unique, and you can change it. Avoid entering
	// confidential information.
	Description *string `mandatory:"false" json:"description"`

	// This string is meant for internal testing.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

//GetIsEnabled returns IsEnabled
func (m CreateNoOpActionDetails) GetIsEnabled() *bool {
	return m.IsEnabled
}

//GetDescription returns Description
func (m CreateNoOpActionDetails) GetDescription() *string {
	return m.Description
}

func (m CreateNoOpActionDetails) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m CreateNoOpActionDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeCreateNoOpActionDetails CreateNoOpActionDetails
	s := struct {
		DiscriminatorParam string `json:"actionType"`
		MarshalTypeCreateNoOpActionDetails
	}{
		"NOOP",
		(MarshalTypeCreateNoOpActionDetails)(m),
	}

	return json.Marshal(&s)
}
