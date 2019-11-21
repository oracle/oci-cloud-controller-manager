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

// JobDefinitionPermissionsSummary Permissions object for Job Definitions
type JobDefinitionPermissionsSummary struct {

	// An array of permissions
	UserPermissions []string `mandatory:"false" json:"userPermissions"`

	// The unique key of the parent Job Definition.
	JobDefinitionKey *string `mandatory:"false" json:"jobDefinitionKey"`
}

func (m JobDefinitionPermissionsSummary) String() string {
	return common.PointerString(m)
}
