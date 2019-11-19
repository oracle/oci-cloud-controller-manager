// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Database Service API
//
// The API for the Database Service.
//

package database

import (
	"github.com/oracle/oci-go-sdk/common"
)

// MoveDatabaseDetails Move Database Details
type MoveDatabaseDetails struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the destination DB Home to move the database.
	TargetDbHomeId *string `mandatory:"true" json:"targetDbHomeId"`
}

func (m MoveDatabaseDetails) String() string {
	return common.PointerString(m)
}
