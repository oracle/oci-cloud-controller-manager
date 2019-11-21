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

// DataAssetPermissionsSummary Permissions object for Data Assets
type DataAssetPermissionsSummary struct {

	// An array of permissions
	UserPermissions []string `mandatory:"false" json:"userPermissions"`

	// The unique key of the parent Data Asset.
	DataAssetKey *string `mandatory:"false" json:"dataAssetKey"`
}

func (m DataAssetPermissionsSummary) String() string {
	return common.PointerString(m)
}
