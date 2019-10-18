// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

func (s *CoreTestSuite) TestListShapes() {
	compartmentID := "compartmentid"
	opts := &ListShapesOptions{}
	opts.AvailabilityDomain = "domainid"
	opts.ImageID = "imageid"
	opts.Limit = 100
	opts.Page = "pageid"

	details := &requestDetails{
		name:     resourceShapes,
		optional: opts,
		required: listOCIDRequirement{CompartmentID: compartmentID},
	}

	expected := []Shape{
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

	actual, e := s.requestor.ListShapes(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.Shapes))
	s.Equal("nextpage", actual.NextPage)
	s.Equal("requestid", actual.RequestID)

}
