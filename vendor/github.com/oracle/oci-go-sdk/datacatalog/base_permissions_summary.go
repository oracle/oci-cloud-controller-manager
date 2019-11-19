// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DataCatalog API
//
// A description of the DataCatalog API
//

package datacatalog

import (
	"github.com/oracle/oci-go-sdk/common"
)

// BasePermissionsSummary Permissions object sent as part of the response
type BasePermissionsSummary struct {

	// An array of permissions
	UserPermissions []string `mandatory:"false" json:"userPermissions"`
}

func (m BasePermissionsSummary) String() string {
	return common.PointerString(m)
}
