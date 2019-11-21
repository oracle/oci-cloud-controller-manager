// Copyright (c) 2016, 2018, 2019, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// DataCatalog API
//
// A description of the DataCatalog API
//

package datacatalog

// ConnectionResultEnum Enum with underlying type: string
type ConnectionResultEnum string

// Set of constants representing the allowable values for ConnectionResultEnum
const (
	ConnectionResultSucceeded ConnectionResultEnum = "SUCCEEDED"
	ConnectionResultFailed    ConnectionResultEnum = "FAILED"
)

var mappingConnectionResult = map[string]ConnectionResultEnum{
	"SUCCEEDED": ConnectionResultSucceeded,
	"FAILED":    ConnectionResultFailed,
}

// GetConnectionResultEnumValues Enumerates the set of values for ConnectionResultEnum
func GetConnectionResultEnumValues() []ConnectionResultEnum {
	values := make([]ConnectionResultEnum, 0)
	for _, v := range mappingConnectionResult {
		values = append(values, v)
	}
	return values
}
