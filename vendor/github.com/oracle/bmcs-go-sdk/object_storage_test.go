// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ObjectStorageTestSuite struct {
	suite.Suite
	requestor *mockRequestor
	nilHeader http.Header
}

func (s *ObjectStorageTestSuite) SetupTest() {
	s.requestor = newMockRequestor(s)
}

func TestRunObjectStorageTests(t *testing.T) {
	suite.Run(t, new(ObjectStorageTestSuite))
}

func (s *ObjectStorageTestSuite) testDeleteResource(
	resourceName resourceName,
	name string,
	namespaceName Namespace,
	funcUnderTest func(string, Namespace, *IfMatchOptions) error,
) {
	required := struct {
		ocidRequirement
		Name string `header:"-" json:"name" url:"-"`
	}{
		Name: name,
	}
	option := &IfMatchOptions{IfMatch: "abcd"}

	details := &requestDetails{
		ids:      urlParts{namespaceName, resourceBuckets, name},
		optional: option,
		required: required,
	}

	s.requestor.On("deleteRequest", details).Return(nil)

	e := funcUnderTest(name, namespaceName, option)
	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
}
