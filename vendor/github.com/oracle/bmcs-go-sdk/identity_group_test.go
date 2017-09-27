// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

func (s *IdentityTestSuite) TestCreateGroup() {
	compartmentID := s.requestor.authInfo.tenancyOCID
	res := Group{
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
		name:     resourceGroups,
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

	actual, e := s.requestor.CreateGroup(res.Name, res.Description, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestGetGroup() {
	res := Group{ID: "id"}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourceGroups,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetGroup(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestUpdateGroup() {
	res := Group{
		ID:          "id",
		Description: "desc",
	}

	opts := &UpdateIdentityOptions{}
	opts.Description = "desc"

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourceGroups,
		optional: opts,
	}

	respHeaders := http.Header{}
	respHeaders.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: respHeaders,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateGroup(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.Description, actual.Description)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestDeleteGroup() {
	opts := &IfMatchOptions{IfMatch: "abcd"}

	details := &requestDetails{
		ids:      urlParts{"id"},
		name:     resourceGroups,
		optional: opts,
	}

	s.requestor.On("deleteRequest", details).Return(nil)

	e := s.requestor.DeleteGroup("id", opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
}

func (s *IdentityTestSuite) TestListGroups() {
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "wxyz"

	details := &requestDetails{
		name:     resourceGroups,
		optional: opts,
		required: listOCIDRequirement{s.requestor.authInfo.tenancyOCID},
	}

	expected := ListGroups{
		Groups: []Group{
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
		body: marshalObjectForTest(expected.Groups),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListGroups(opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.Groups), len(actual.Groups))
}
