// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Budgets API
//
// Use the Budgets API to manage budgets and budget alerts.
//

package budget

import (
	"github.com/oracle/oci-go-sdk/common"
)

// TargetType The type of target on which budget is applied. Valid values are COMPARTMENT or TAG.
type TargetType struct {
}

func (m TargetType) String() string {
	return common.PointerString(m)
}
