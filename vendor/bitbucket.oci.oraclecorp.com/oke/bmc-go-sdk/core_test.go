// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CoreTestSuite struct {
	suite.Suite
	requestor *mockCoreRequestor
	nilHeader http.Header
}

type mockCoreRequestor struct {
	*Client
	mock.Mock
}

func (mr *mockCoreRequestor) request(method string, reqOpts request) (resp *response, e error) {
	args := mr.Called(method, reqOpts)
	return args.Get(0).(*response), args.Error(1)
}

func (mr *mockCoreRequestor) getRequest(reqOpts request) (resp *response, e error) {
	args := mr.Called(reqOpts)
	return args.Get(0).(*response), args.Error(1)

}

func (mr *mockCoreRequestor) deleteRequest(reqOpts request) (e error) {
	args := mr.Called(reqOpts)
	return args.Error(0)
}

func newMockCoreRequestor(cts *CoreTestSuite) (m *mockCoreRequestor) {
	m = new(mockCoreRequestor)
	m.Client = createClientForTest()
	m.coreApi = m
	return
}

func (s *CoreTestSuite) SetupTest() {
	s.requestor = newMockCoreRequestor(s)
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
