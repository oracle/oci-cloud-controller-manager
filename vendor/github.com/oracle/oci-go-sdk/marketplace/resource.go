// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Marketplace Service API
//
// Manage applications in Oracle Cloud Infrastructure Marketplace.
//

package marketplace

import (
	"github.com/oracle/oci-go-sdk/common"
)

// Resource The model for a package's primary resource.
type Resource struct {

	// The type of the service.
	ServiceType *string `mandatory:"false" json:"serviceType"`

	// The type of the resource.
	ResourceType *string `mandatory:"false" json:"resourceType"`
}

func (m Resource) String() string {
	return common.PointerString(m)
}
