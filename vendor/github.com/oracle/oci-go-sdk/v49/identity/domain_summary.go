// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Identity and Access Management Service API
//
// APIs for managing users, groups, compartments, and policies.
//

package identity

import (
	"github.com/oracle/oci-go-sdk/v49/common"
)

// DomainSummary As the name suggests, a `DomainSummary` object contains information about a `Domain`.
type DomainSummary struct {

	// The OCID of the domain
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the comparment containing the domain.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The immutable name of the domain that is added at the domain creation time and
	// needs to be used in policies to reference this domain.
	Name *string `mandatory:"true" json:"name"`

	// The mutable display name of the domain
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The domain descripition
	Description *string `mandatory:"true" json:"description"`

	// URL for the regional IDCS compatibility layer.
	Url *string `mandatory:"true" json:"url"`

	// The home region for the domain.
	HomeRegion *string `mandatory:"true" json:"homeRegion"`

	// The regions domain is replicated to.
	ReplicaRegions []ReplicatedRegionDetails `mandatory:"true" json:"replicaRegions"`

	// The type of the domain.
	Type DomainTypeEnum `mandatory:"true" json:"type"`

	// The License type of Domain
	LicenseType *string `mandatory:"true" json:"licenseType"`

	// Indicates whether domain is hidden on login screen or not.
	IsHiddenOnLogin *bool `mandatory:"true" json:"isHiddenOnLogin"`

	// Date and time the domain was created, in the format defined by RFC3339.
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The current state.
	LifecycleState DomainLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Any additional details about the current state of the Domain.
	LifecycleDetails DomainSummaryLifecycleDetailsEnum `mandatory:"false" json:"lifecycleDetails,omitempty"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m DomainSummary) String() string {
	return common.PointerString(m)
}

// DomainSummaryLifecycleDetailsEnum Enum with underlying type: string
type DomainSummaryLifecycleDetailsEnum string

// Set of constants representing the allowable values for DomainSummaryLifecycleDetailsEnum
const (
	DomainSummaryLifecycleDetailsDeactivating DomainSummaryLifecycleDetailsEnum = "DEACTIVATING"
	DomainSummaryLifecycleDetailsActivating   DomainSummaryLifecycleDetailsEnum = "ACTIVATING"
	DomainSummaryLifecycleDetailsUpdating     DomainSummaryLifecycleDetailsEnum = "UPDATING"
)

var mappingDomainSummaryLifecycleDetails = map[string]DomainSummaryLifecycleDetailsEnum{
	"DEACTIVATING": DomainSummaryLifecycleDetailsDeactivating,
	"ACTIVATING":   DomainSummaryLifecycleDetailsActivating,
	"UPDATING":     DomainSummaryLifecycleDetailsUpdating,
}

// GetDomainSummaryLifecycleDetailsEnumValues Enumerates the set of values for DomainSummaryLifecycleDetailsEnum
func GetDomainSummaryLifecycleDetailsEnumValues() []DomainSummaryLifecycleDetailsEnum {
	values := make([]DomainSummaryLifecycleDetailsEnum, 0)
	for _, v := range mappingDomainSummaryLifecycleDetails {
		values = append(values, v)
	}
	return values
}
