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

// Term Full Term definition. A defined business term in a business glossary. As well as a term definition, simple format
// rules for attributes mapping to the term (for example, the expected data type and length restrictions) may be
// stated at the term level. Nesting of terms to support a hierarchy is supported by default.
type Term struct {

	// Unique Term key that is immutable.
	Key *string `mandatory:"true" json:"key"`

	// The display name of a user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Detailed description of the Term.
	Description *string `mandatory:"false" json:"description"`

	// Unique id of the parent Glossary.
	GlossaryKey *string `mandatory:"false" json:"glossaryKey"`

	// This terms parent term key. Will be null if the term has no parent term.
	ParentTermKey *string `mandatory:"false" json:"parentTermKey"`

	// Indicates whether a term may contain child terms.
	IsAllowedToHaveChildTerms *bool `mandatory:"false" json:"isAllowedToHaveChildTerms"`

	// Absolute path of the term.
	Path *string `mandatory:"false" json:"path"`

	// The current state of the Term.
	LifecycleState LifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the Term was created, in the format defined by RFC3339.
	// Example: `2019-03-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The last time that any change was made to the Term. An RFC3339 formatted datetime string.
	TimeUpdated *common.SDKTime `mandatory:"false" json:"timeUpdated"`

	// Id (OCID) of the user who created the Term.
	CreatedById *string `mandatory:"false" json:"createdById"`

	// Id (OCID) of the user who modified the Term.
	UpdatedById *string `mandatory:"false" json:"updatedById"`

	// Id (OCID) of the user who is the owner of this business terminology.
	Owner *string `mandatory:"false" json:"owner"`

	// Status of the approval process workflow for this business term in the glossary
	WorkflowStatus TermWorkflowStatusEnum `mandatory:"false" json:"workflowStatus,omitempty"`

	// URI to the Term instance in the API.
	Uri *string `mandatory:"false" json:"uri"`

	// The number of objects tagged with this term
	AssociatedObjectCount *int `mandatory:"false" json:"associatedObjectCount"`

	// Array of objects associated to a term.
	AssociatedObjects []TermAssociatedObject `mandatory:"false" json:"associatedObjects"`
}

func (m Term) String() string {
	return common.PointerString(m)
}
