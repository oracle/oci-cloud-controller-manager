// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// API Gateway API
//
// API for the API Gateway service. Use this API to manage gateways, deployments, and related items.
//

package apigateway

import (
	"github.com/oracle/oci-go-sdk/common"
)

// Deployment Deploys an API on a Gateway.
type Deployment struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the resource.
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the resource.
	GatewayId *string `mandatory:"true" json:"gatewayId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment in which
	// resource is created.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Path prefix
	PathPrefix *string `mandatory:"true" json:"pathPrefix"`

	// The endpoint to access this deployment on the gateway
	Endpoint *string `mandatory:"true" json:"endpoint"`

	Specification *ApiSpecification `mandatory:"true" json:"specification"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Example: `My new resource`
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The time this resource was created. An RFC3339 formatted datetime string
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The time this resource was last updated. An RFC3339 formatted datetime string
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// The current state of the Deployment.
	LifecycleState DeploymentLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// An message describing the current state in more detail.
	// For example, can be used to provide actionable information for a
	// resource in Failed state.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// Free-form tags for this resource. Each tag is a simple key-value pair
	// with no predefined name, type, or namespace. For more information, see
	// Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see
	// Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m Deployment) String() string {
	return common.PointerString(m)
}

// DeploymentLifecycleStateEnum Enum with underlying type: string
type DeploymentLifecycleStateEnum string

// Set of constants representing the allowable values for DeploymentLifecycleStateEnum
const (
	DeploymentLifecycleStateCreating DeploymentLifecycleStateEnum = "CREATING"
	DeploymentLifecycleStateActive   DeploymentLifecycleStateEnum = "ACTIVE"
	DeploymentLifecycleStateUpdating DeploymentLifecycleStateEnum = "UPDATING"
	DeploymentLifecycleStateDeleting DeploymentLifecycleStateEnum = "DELETING"
	DeploymentLifecycleStateDeleted  DeploymentLifecycleStateEnum = "DELETED"
	DeploymentLifecycleStateFailed   DeploymentLifecycleStateEnum = "FAILED"
)

var mappingDeploymentLifecycleState = map[string]DeploymentLifecycleStateEnum{
	"CREATING": DeploymentLifecycleStateCreating,
	"ACTIVE":   DeploymentLifecycleStateActive,
	"UPDATING": DeploymentLifecycleStateUpdating,
	"DELETING": DeploymentLifecycleStateDeleting,
	"DELETED":  DeploymentLifecycleStateDeleted,
	"FAILED":   DeploymentLifecycleStateFailed,
}

// GetDeploymentLifecycleStateEnumValues Enumerates the set of values for DeploymentLifecycleStateEnum
func GetDeploymentLifecycleStateEnumValues() []DeploymentLifecycleStateEnum {
	values := make([]DeploymentLifecycleStateEnum, 0)
	for _, v := range mappingDeploymentLifecycleState {
		values = append(values, v)
	}
	return values
}
