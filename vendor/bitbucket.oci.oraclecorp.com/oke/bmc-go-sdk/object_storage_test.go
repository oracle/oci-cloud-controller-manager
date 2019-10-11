// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ObjectStorageTestSuite struct {
	suite.Suite
	requestor *mockObjectStorageRequestor
	nilHeader http.Header
}

type mockObjectStorageRequestor struct {
	*Client
	mock.Mock
}

func (mr *mockObjectStorageRequestor) request(method string, reqOpts request) (resp *response, e error) {
	args := mr.Called(method, reqOpts)
	return args.Get(0).(*response), args.Error(1)
}

func (mr *mockObjectStorageRequestor) getRequest(reqOpts request) (resp *response, e error) {
	args := mr.Called(reqOpts)
	return args.Get(0).(*response), args.Error(1)

}

func (mr *mockObjectStorageRequestor) deleteRequest(reqOpts request) (e error) {
	args := mr.Called(reqOpts)
	return args.Error(0)
}

func newMockObjectStorageRequestor(cts *ObjectStorageTestSuite) (m *mockObjectStorageRequestor) {
	m = new(mockObjectStorageRequestor)
	m.Client = createClientForTest()
	m.objectStorageApi = m
	return
}

func (s *ObjectStorageTestSuite) SetupTest() {
	s.requestor = newMockObjectStorageRequestor(s)
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
