// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Storage Gateway API
//
// API for the Storage Gateway service. Use this API to manage storage gateways and related items. For more
// information, see Overview of Storage Gateway (https://docs.cloud.oracle.com/iaas/Content/StorageGateway/Concepts/storagegatewayoverview.htm).
//

package storagegateway

import (
	"github.com/oracle/oci-go-sdk/common"
)

// ChangeStorageGatewayCompartmentDetails The configuration details for the move operation.
type ChangeStorageGatewayCompartmentDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the compartment to move the storage gateway to.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`
}

func (m ChangeStorageGatewayCompartmentDetails) String() string {
	return common.PointerString(m)
}
