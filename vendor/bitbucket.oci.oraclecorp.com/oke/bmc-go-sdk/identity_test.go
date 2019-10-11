// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type mockIdentityRequestor struct {
	*Client
	mock.Mock
}

func (mr *mockIdentityRequestor) request(method string, reqOpts request) (resp *response, e error) {
	args := mr.Called(method, reqOpts)
	return args.Get(0).(*response), args.Error(1)
}

func (mr *mockIdentityRequestor) getRequest(reqOpts request) (resp *response, e error) {
	args := mr.Called(reqOpts)
	return args.Get(0).(*response), args.Error(1)

}

func (mr *mockIdentityRequestor) deleteRequest(reqOpts request) (e error) {
	args := mr.Called(reqOpts)
	return args.Error(0)
}

func newMockIdentityRequestor(s *IdentityTestSuite) (m *mockIdentityRequestor) {
	m = new(mockIdentityRequestor)
	c := createClientForTest()

	m.Client = c
	m.identityApi = m

	return
}

type IdentityTestSuite struct {
	suite.Suite
	requestor *mockIdentityRequestor
	nilHeader http.Header
	nilQuery  []interface{}
}

func (s *IdentityTestSuite) SetupTest() {
	s.requestor = newMockIdentityRequestor(s)
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
