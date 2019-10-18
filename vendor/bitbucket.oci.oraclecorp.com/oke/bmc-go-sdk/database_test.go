// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type mockDatabaseRequestor struct {
	*Client
	mock.Mock
}

func (mr *mockDatabaseRequestor) request(method string, reqOpts request) (resp *response, e error) {
	args := mr.Called(method, reqOpts)
	return args.Get(0).(*response), args.Error(1)
}

func (mr *mockDatabaseRequestor) getRequest(reqOpts request) (resp *response, e error) {
	args := mr.Called(reqOpts)
	return args.Get(0).(*response), args.Error(1)

}

func (mr *mockDatabaseRequestor) deleteRequest(reqOpts request) (e error) {
	args := mr.Called(reqOpts)
	return args.Error(0)
}

func newMockDatabaseRequestor(s *DatabaseTestSuite) (m *mockDatabaseRequestor) {
	m = new(mockDatabaseRequestor)
	c := createClientForTest()

	m.Client = c
	m.databaseApi = m

	return
}

type DatabaseTestSuite struct {
	suite.Suite
	requestor *mockDatabaseRequestor
	nilHeader http.Header
	nilQuery  []interface{}
}

func (s *DatabaseTestSuite) SetupTest() {
	s.requestor = newMockDatabaseRequestor(s)
}

func (s *DatabaseTestSuite) testDeleteResource(name resourceName, id string, funcUnderTest func(string, *IfMatchOptions) error) {
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

func TestRunDatabaseTests(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}
