// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// KAM API
//
// description: |
//   Kubernetes Add-on Manager API for installing, uninstalling and upgrading
//   OKE add-ons or Marketplace applications on an OKE cluster
//

package kam

import (
	"github.com/oracle/oci-go-sdk/common"
)

// WorkRequest KAM work request
type WorkRequest struct {

	// The OCID of the kam work request.
	Id *string `mandatory:"true" json:"id"`

	// The type of work request.
	OperationType WorkRequestTypeEnum `mandatory:"true" json:"operationType"`

	// The lifecycle state of a kam work request.
	Status WorkRequestStateEnum `mandatory:"true" json:"status"`

	// The resources affected by this work request.
	Resources []WorkRequestResource `mandatory:"true" json:"resources"`

	// Percentage of the request completed.
	PercentComplete *float32 `mandatory:"true" json:"percentComplete"`

	// The date and time the request was created, as described in
	// RFC 3339 (https://tools.ietf.org/rfc/rfc3339), section 14.29.
	TimeAccepted *common.SDKTime `mandatory:"true" json:"timeAccepted"`

	// The date and time the request was started, as described in RFC 3339 (https://tools.ietf.org/rfc/rfc3339),
	// section 14.29.
	TimeStarted *common.SDKTime `mandatory:"false" json:"timeStarted"`

	// The date and time the object was finished, as described in RFC 3339 (https://tools.ietf.org/rfc/rfc3339).
	TimeFinished *common.SDKTime `mandatory:"false" json:"timeFinished"`

	// The time this work request was updated
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// The OCID of the requesting user.
	RequestingUser *string `mandatory:"false" json:"requestingUser"`

	// The OCID of the requesting user tenancy.
	RequestingUserTenancy *string `mandatory:"false" json:"requestingUserTenancy"`
}

func (m WorkRequest) String() string {
	return common.PointerString(m)
}
