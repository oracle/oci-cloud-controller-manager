// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Autoscaling API
//
// APIs for dynamically scaling Compute resources to meet application requirements.
// For information about the Compute service, see Overview of the Compute Service (https://docs.cloud.oracle.com/Content/Compute/Concepts/computeoverview.htm).
//

package autoscaling

import (
	"github.com/oracle/oci-go-sdk/common"
)

// ChangeAutoScalingCompartmentDetails Contains details indicating which compartment the resource should move to
type ChangeAutoScalingCompartmentDetails struct {

	// The OCID of the new compartment
	CompartmentId *string `mandatory:"true" json:"compartmentId"`
}

func (m ChangeAutoScalingCompartmentDetails) String() string {
	return common.PointerString(m)
}
