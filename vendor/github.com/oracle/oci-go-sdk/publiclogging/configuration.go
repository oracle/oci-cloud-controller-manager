// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// PublicLoggingControlplane API
//
// PublicLoggingControlplane API specification
//

package publiclogging

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// Configuration Log object configuration.
type Configuration struct {
	Source Source `mandatory:"true" json:"source"`

	Indexing *Indexing `mandatory:"false" json:"indexing"`
}

func (m Configuration) String() string {
	return common.PointerString(m)
}

// UnmarshalJSON unmarshals from json
func (m *Configuration) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		Indexing *Indexing `json:"indexing"`
		Source   source    `json:"source"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	m.Indexing = model.Indexing
	nn, e := model.Source.UnmarshalPolymorphicJSON(model.Source.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.Source = nn.(Source)
	} else {
		m.Source = nil
	}
	return
}
