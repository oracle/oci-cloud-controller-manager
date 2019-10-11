// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

func (s *DatabaseTestSuite) TestListDBSupportOperations() {
	details := &requestDetails{
		name: resourceDBSupportedOperations,
	}

	expected := ListSupportedOperations{
		SupportedOperations: []SupportedOperation{
			{
				ID: "1",
			},
			{
				ID: "2",
			},
		},
	}

	resp := &response{
		body: marshalObjectForTest(expected.SupportedOperations),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListSupportedOperations()
	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.SupportedOperations), len(actual.SupportedOperations))
}
