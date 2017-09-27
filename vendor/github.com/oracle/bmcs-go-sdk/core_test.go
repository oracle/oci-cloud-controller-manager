// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CoreTestSuite struct {
	suite.Suite
	requestor *mockRequestor
	nilHeader http.Header
}

func (s *CoreTestSuite) SetupTest() {
	s.requestor = newMockRequestor(s)
}

func TestRunCoreTests(t *testing.T) {
	suite.Run(t, new(CoreTestSuite))
}

func (s *CoreTestSuite) testDeleteResource(name resourceName, id string, funcUnderTest func(string, *IfMatchOptions) error) {
	option := &IfMatchOptions{IfMatch: "abcd"}

	details := &requestDetails{
		ids:      urlParts{id},
		name:     name,
		optional: option,
	}
	s.requestor.On("deleteRequest", details).Return(nil)

	e := funcUnderTest(id, option)
	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
}
