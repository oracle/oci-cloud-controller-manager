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

// UpdateKamReleaseDetails Details of an upgrade request
type UpdateKamReleaseDetails struct {

	// The OCID of the OKE Add-on or Marketplace app to deploy
	KamChartId *string `mandatory:"true" json:"kamChartId"`

	// List of overrides for default configuration
	Overrides []KamOverride `mandatory:"false" json:"overrides"`
}

func (m UpdateKamReleaseDetails) String() string {
	return common.PointerString(m)
}
