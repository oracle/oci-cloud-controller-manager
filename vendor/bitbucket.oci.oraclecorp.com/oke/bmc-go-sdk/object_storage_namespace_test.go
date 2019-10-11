// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

func (s *ObjectStorageTestSuite) TestGetNamespace() {

	var opts interface{}
	var required interface{}

	details := &requestDetails{
		ids:      urlParts{},
		optional: opts,
		required: required,
	}

	namespace := "namespacename"
	resp := &response{
		body: []byte(namespace),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, err := s.requestor.GetNamespace()

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(namespace, string(*actual))
}
