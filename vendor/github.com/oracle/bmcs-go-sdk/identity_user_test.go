// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

func (s *IdentityTestSuite) TestCreateUser() {
	compartmentID := s.requestor.authInfo.tenancyOCID
	res := User{
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
		name:     resourceUsers,
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

	actual, e := s.requestor.CreateUser(res.Name, res.Description, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestGetUser() {
	res := User{ID: "id"}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourceUsers,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetUser(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestUpdateUser() {
	res := User{
		ID:          "id",
		Description: "desc",
	}

	opts := &UpdateIdentityOptions{}
	opts.Description = "desc"

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourceUsers,
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateUser(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.Description, actual.Description)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestUpdateUserState() {
	blocked := false
	res := User{
		ID:             "id",
		InactiveStatus: 0,
	}

	opts := &UpdateUserStateOptions{Blocked: &blocked}

	details := &requestDetails{
		ids:      urlParts{res.ID, "state"},
		name:     resourceUsers,
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateUserState(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.InactiveStatus, actual.InactiveStatus)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestDeleteUser() {
	opts := &IfMatchOptions{IfMatch: "abcd"}

	details := &requestDetails{
		ids:      urlParts{"id"},
		name:     resourceUsers,
		optional: opts,
	}

	s.requestor.On("deleteRequest", details).Return(nil)

	e := s.requestor.DeleteUser("id", opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
}

func (s *IdentityTestSuite) TestListUsers() {
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "wxyz"

	details := &requestDetails{
		name:     resourceUsers,
		optional: opts,
		required: listOCIDRequirement{s.requestor.authInfo.tenancyOCID},
	}

	expected := ListUsers{
		Users: []User{
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
		body: marshalObjectForTest(expected.Users),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListUsers(opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.Users), len(actual.Users))
}
