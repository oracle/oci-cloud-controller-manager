// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Database Service API
//
// The API for the Database Service.
//

package database

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// CreateDatabaseBase Details for creating a database.
// **Warning:** Oracle recommends that you avoid using any confidential information when you supply string values using the API.
type CreateDatabaseBase interface {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the Database Home.
	GetDbHomeId() *string

	// A valid Oracle Database version. To get a list of supported versions, use the ListDbVersions operation.
	GetDbVersion() *string

	// The OCID of the key container that is used as the master encryption key in database transparent data encryption (TDE) operations.
	GetKmsKeyId() *string

	// The OCID of the key container version that is used in database transparent data encryption (TDE) operations KMS Key can have multiple key versions. If none is specified, the current key version (latest) of the Key Id is used for the operation.
	GetKmsKeyVersionId() *string
}

type createdatabasebase struct {
	JsonData        []byte
	DbHomeId        *string `mandatory:"true" json:"dbHomeId"`
	DbVersion       *string `mandatory:"false" json:"dbVersion"`
	KmsKeyId        *string `mandatory:"false" json:"kmsKeyId"`
	KmsKeyVersionId *string `mandatory:"false" json:"kmsKeyVersionId"`
	Source          string  `json:"source"`
}

// UnmarshalJSON unmarshals json
func (m *createdatabasebase) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalercreatedatabasebase createdatabasebase
	s := struct {
		Model Unmarshalercreatedatabasebase
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.DbHomeId = s.Model.DbHomeId
	m.DbVersion = s.Model.DbVersion
	m.KmsKeyId = s.Model.KmsKeyId
	m.KmsKeyVersionId = s.Model.KmsKeyVersionId
	m.Source = s.Model.Source

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *createdatabasebase) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Source {
	case "NONE":
		mm := CreateNewDatabaseDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

//GetDbHomeId returns DbHomeId
func (m createdatabasebase) GetDbHomeId() *string {
	return m.DbHomeId
}

//GetDbVersion returns DbVersion
func (m createdatabasebase) GetDbVersion() *string {
	return m.DbVersion
}

//GetKmsKeyId returns KmsKeyId
func (m createdatabasebase) GetKmsKeyId() *string {
	return m.KmsKeyId
}

//GetKmsKeyVersionId returns KmsKeyVersionId
func (m createdatabasebase) GetKmsKeyVersionId() *string {
	return m.KmsKeyVersionId
}

func (m createdatabasebase) String() string {
	return common.PointerString(m)
}

// CreateDatabaseBaseSourceEnum Enum with underlying type: string
type CreateDatabaseBaseSourceEnum string

// Set of constants representing the allowable values for CreateDatabaseBaseSourceEnum
const (
	CreateDatabaseBaseSourceNone CreateDatabaseBaseSourceEnum = "NONE"
)

var mappingCreateDatabaseBaseSource = map[string]CreateDatabaseBaseSourceEnum{
	"NONE": CreateDatabaseBaseSourceNone,
}

// GetCreateDatabaseBaseSourceEnumValues Enumerates the set of values for CreateDatabaseBaseSourceEnum
func GetCreateDatabaseBaseSourceEnumValues() []CreateDatabaseBaseSourceEnum {
	values := make([]CreateDatabaseBaseSourceEnum, 0)
	for _, v := range mappingCreateDatabaseBaseSource {
		values = append(values, v)
	}
	return values
}
