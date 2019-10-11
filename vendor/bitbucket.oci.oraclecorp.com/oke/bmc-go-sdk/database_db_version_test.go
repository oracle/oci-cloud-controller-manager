// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

func (s *DatabaseTestSuite) TestListDBVersions() {
	reqs := struct {
		listOCIDRequirement
	}{}
	reqs.CompartmentID = "compartmentID"
	opts := ListOptions{}

	details := &requestDetails{
		name:     resourceDBVersions,
		optional: &opts,
		required: reqs,
	}

	expected := ListDBVersions{
		DBVersions: []DBVersion{
			{
				Version: "1",
			},
			{
				Version: "2",
			},
		},
	}

	resp := &response{
		body: marshalObjectForTest(expected.DBVersions),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListDBVersions(reqs.CompartmentID, &opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.DBVersions), len(actual.DBVersions))
}
