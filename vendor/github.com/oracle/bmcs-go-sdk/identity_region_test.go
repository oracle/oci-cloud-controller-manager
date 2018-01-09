// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

func (s *IdentityTestSuite) TestListRegions() {
	details := &requestDetails{
		name:     resourceRegions,
		required: listOCIDRequirement{s.requestor.Client.authInfo.tenancyOCID},
	}

	expected := ListRegions{
		Regions: []IdentityRegion{
			{
				Name: "us-ashburn-1",
				Key:  "IAD",
			},
			{
				Name: "us-phoenix-1",
				Key:  "PHX",
			},
		},
	}

	resp := &response{
		body: marshalObjectForTest(expected.Regions),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListRegions()

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.Regions), len(actual.Regions))
}
