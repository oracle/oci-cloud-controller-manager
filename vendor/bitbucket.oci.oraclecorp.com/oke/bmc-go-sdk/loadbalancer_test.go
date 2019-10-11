// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type LoadbalancerTestSuite struct {
	suite.Suite
	requestor *mockLoadbalancerRequestor
	nilHeader http.Header
}

type mockLoadbalancerRequestor struct {
	*Client
	mock.Mock
}

func (mr *mockLoadbalancerRequestor) request(method string, reqOpts request) (resp *response, e error) {
	args := mr.Called(method, reqOpts)
	return args.Get(0).(*response), args.Error(1)
}

func (mr *mockLoadbalancerRequestor) getRequest(reqOpts request) (resp *response, e error) {
	args := mr.Called(reqOpts)
	return args.Get(0).(*response), args.Error(1)

}

func (mr *mockLoadbalancerRequestor) deleteRequest(reqOpts request) (e error) {
	args := mr.Called(reqOpts)
	return args.Error(0)
}

func newMockLoadbalancerRequestor(cts *LoadbalancerTestSuite) (m *mockLoadbalancerRequestor) {
	m = new(mockLoadbalancerRequestor)
	m.Client = createClientForTest()
	m.loadBalancerApi = m
	return
}

func (s *LoadbalancerTestSuite) SetupTest() {
	s.requestor = newMockLoadbalancerRequestor(s)
}

func TestRunLoadbalancerTests(t *testing.T) {
	suite.Run(t, new(LoadbalancerTestSuite))
}
