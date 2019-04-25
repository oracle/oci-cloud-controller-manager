// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Resource Search Service
//
// Search for resources across your cloud infrastructure
//

package resourcesearch

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// SearchDetails A base request type containing common criteria for searching for resources.
type SearchDetails interface {

	// Defines the type of matching context returned in response, default is NONE. If HIGHLIGHTS is set, then there will be highlighting
	// fragments returned from the service (see ResourceSummary.searchContext and SearchContext).  If NONE is set, then no search
	// context will be returned.
	GetMatchingContextType() SearchDetailsMatchingContextTypeEnum
}

type searchdetails struct {
	JsonData            []byte
	MatchingContextType SearchDetailsMatchingContextTypeEnum `mandatory:"false" json:"matchingContextType,omitempty"`
	Type                string                               `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *searchdetails) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalersearchdetails searchdetails
	s := struct {
		Model Unmarshalersearchdetails
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.MatchingContextType = s.Model.MatchingContextType
	m.Type = s.Model.Type

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *searchdetails) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "Structured":
		mm := StructuredSearchDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "FreeText":
		mm := FreeTextSearchDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

//GetMatchingContextType returns MatchingContextType
func (m searchdetails) GetMatchingContextType() SearchDetailsMatchingContextTypeEnum {
	return m.MatchingContextType
}

func (m searchdetails) String() string {
	return common.PointerString(m)
}

// SearchDetailsMatchingContextTypeEnum Enum with underlying type: string
type SearchDetailsMatchingContextTypeEnum string

// Set of constants representing the allowable values for SearchDetailsMatchingContextTypeEnum
const (
	SearchDetailsMatchingContextTypeNone       SearchDetailsMatchingContextTypeEnum = "NONE"
	SearchDetailsMatchingContextTypeHighlights SearchDetailsMatchingContextTypeEnum = "HIGHLIGHTS"
)

var mappingSearchDetailsMatchingContextType = map[string]SearchDetailsMatchingContextTypeEnum{
	"NONE":       SearchDetailsMatchingContextTypeNone,
	"HIGHLIGHTS": SearchDetailsMatchingContextTypeHighlights,
}

// GetSearchDetailsMatchingContextTypeEnumValues Enumerates the set of values for SearchDetailsMatchingContextTypeEnum
func GetSearchDetailsMatchingContextTypeEnumValues() []SearchDetailsMatchingContextTypeEnum {
	values := make([]SearchDetailsMatchingContextTypeEnum, 0)
	for _, v := range mappingSearchDetailsMatchingContextType {
		values = append(values, v)
	}
	return values
}
