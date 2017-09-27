// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"bytes"
	"encoding/json"

	"github.com/stretchr/testify/mock"
)

func marshalObjectForTest(obj interface{}) []byte {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.Encode(obj)

	return buffer.Bytes()
}

type mockRequestor struct {
	*Client
	mock.Mock
}

func (mr *mockRequestor) request(method string, reqOpts request) (resp *response, e error) {
	args := mr.Called(method, reqOpts)
	return args.Get(0).(*response), args.Error(1)
}

func (mr *mockRequestor) getRequest(reqOpts request) (resp *response, e error) {
	args := mr.Called(reqOpts)
	return args.Get(0).(*response), args.Error(1)

}

func (mr *mockRequestor) postRequest(reqOpts request) (resp *response, e error) {
	args := mr.Called(reqOpts)
	return args.Get(0).(*response), args.Error(1)

}

func (mr *mockRequestor) deleteRequest(reqOpts request) (e error) {
	args := mr.Called(reqOpts)
	return args.Error(0)
}

func newMockRequestor(s interface{}) (m *mockRequestor) {
	m = new(mockRequestor)

	m.Client = createClientForTest()
	m.coreApi = m
	m.objectStorageApi = m
	m.databaseApi = m
	m.identityApi = m
	m.loadBalancerApi = m

	return
}
