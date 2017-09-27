// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

func (s *IdentityTestSuite) TestCreateCompartment() {
	compartmentID := s.requestor.authInfo.tenancyOCID
	res := Compartment{
		CompartmentID: compartmentID,
		Description:   "description",
		Name:          "name",
	}

	opts := &RetryTokenOptions{RetryToken: "xxxxx"}

	required := identityCreationRequirement{
		CompartmentID: res.CompartmentID,
		Description:   res.Description,
		Name:          res.Name,
	}

	details := &requestDetails{
		name:     resourceCompartments,
		optional: opts,
		required: required,
	}

	respHeaders := http.Header{}
	respHeaders.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: respHeaders,
	}

	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, e := s.requestor.CreateCompartment(res.Name, res.Description, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestGetCompartment() {
	res := Compartment{ID: "id"}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourceCompartments,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetCompartment(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestUpdateCompartment() {
	res := Compartment{
		ID:          "id",
		Description: "desc",
	}

	opts := &UpdateCompartmentOptions{}
	opts.Name = "name"
	opts.Description = "desc"

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourceCompartments,
		optional: opts,
	}

	respHeaders := http.Header{}
	respHeaders.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: respHeaders,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateCompartment(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.Name, actual.Name)
	s.Equal(res.Description, actual.Description)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestListCompartments() {
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "wxyz"

	details := &requestDetails{
		name:     resourceCompartments,
		optional: opts,
		required: listOCIDRequirement{s.requestor.authInfo.tenancyOCID},
	}

	expected := ListCompartments{
		Compartments: []Compartment{
			{
				ID:   "1",
				Name: "one",
			},
			{
				ID:   "2",
				Name: "two",
			},
		},
	}

	resp := &response{
		body: marshalObjectForTest(expected.Compartments),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListCompartments(opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.Compartments), len(actual.Compartments))
}
