// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

func (s *DatabaseTestSuite) TestListDBSystemShapes() {
	availabilityDomain := "availabilitydomain"
	compartmentID := "compartmentid"

	opts := &ListOptions{}
	opts.Page = "pageid"
	opts.Limit = 100

	required := struct {
		listOCIDRequirement
		AvailabilityDomain string `header:"-" json:"-" url:"availabilityDomain"`
	}{
		AvailabilityDomain: availabilityDomain,
	}
	required.CompartmentID = compartmentID

	details := &requestDetails{
		name:     resourceDBSystemShapes,
		optional: opts,
		required: required,
	}

	expected := []DBSystemShape{
		{Name: "shape1"},
		{Name: "shape2"},
	}

	headers := http.Header{}
	headers.Set(headerOPCNextPage, "nextpage")
	headers.Set(headerOPCRequestID, "requestid")
	s.requestor.On("getRequest", details).Return(
		&response{
			header: headers,
			body:   marshalObjectForTest(expected),
		},
		nil,
	)

	actual, e := s.requestor.ListDBSystemShapes(
		availabilityDomain,
		compartmentID,
		opts,
	)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.DBSystemShapes))
	s.Equal("nextpage", actual.NextPage)
	s.Equal("requestid", actual.RequestID)

}
