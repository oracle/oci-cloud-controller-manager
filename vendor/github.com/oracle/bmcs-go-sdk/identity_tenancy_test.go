// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

func (s *IdentityTestSuite) TestGetTenancy() {
	res := Tenancy{
		ID:            "id",
		Name:          "tenancyName",
		Description:   "This is a tenancy",
		HomeRegionKey: "IAD",
	}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourceTenancies,
	}

	resp := &response{
		body: marshalObjectForTest(res),
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetTenancy(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
}
