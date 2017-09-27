// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type IdentityTestSuite struct {
	suite.Suite
	requestor *mockRequestor
	nilHeader http.Header
	nilQuery  []interface{}
}

func (s *IdentityTestSuite) SetupTest() {
	s.requestor = newMockRequestor(s)
}

func (s *IdentityTestSuite) testDeleteResource(name resourceName, id string, funcUnderTest func(string, *IfMatchOptions) error) {
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

func TestRunIdentityTests(t *testing.T) {
	suite.Run(t, new(IdentityTestSuite))
}
