// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

func (s *IdentityTestSuite) TestListAvailabilityDomains() {
	compartmentID := "compartmentid"

	details := &requestDetails{
		name:     resourceAvailabilityDomains,
		required: listOCIDRequirement{CompartmentID: compartmentID},
	}

	resp := &response{
		body: marshalObjectForTest(
			[]AvailabilityDomain{
				{
					Name:          "one",
					CompartmentID: "1",
				},
				{
					Name:          "two",
					CompartmentID: "1",
				},
			},
		),
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListAvailabilityDomains(compartmentID)
	s.requestor.AssertExpectations(s.T())

	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(actual.AvailabilityDomains), 2)
}
