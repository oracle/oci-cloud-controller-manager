// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

func (s *IdentityTestSuite) TestCreateSwiftPassword() {
	desc := "description"
	res := SwiftPassword{
		Password:    "password",
		ID:          "id",
		UserID:      "userID",
		Description: desc,
	}

	opts := &RetryTokenOptions{RetryToken: "xxxxx"}

	required := struct {
		Description string `header:"-" json:"description" url:"-"`
	}{
		Description: desc,
	}

	details := &requestDetails{
		ids:      urlParts{res.UserID, resourceSwiftPasswords},
		name:     resourceUsers,
		optional: opts,
		required: required,
	}

	respHeaders := http.Header{}
	respHeaders.Set(headerETag, "ETAG!")
	respHeaders.Set(headerOPCRequestID, "responseId")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: respHeaders,
	}

	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, e := s.requestor.CreateSwiftPassword(res.UserID, desc, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.UserID, actual.UserID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestUpdateSwiftPassword() {
	desc := "description"
	res := SwiftPassword{
		Password:    "password",
		ID:          "id",
		UserID:      "userID",
		Description: desc,
	}

	opts := &UpdateIdentityOptions{
		Description: desc,
	}

	details := &requestDetails{
		ids:      urlParts{res.UserID, resourceSwiftPasswords, res.ID},
		name:     resourceUsers,
		optional: opts,
	}

	respHeaders := http.Header{}
	respHeaders.Set(headerETag, "ETAG!")
	respHeaders.Set(headerOPCRequestID, "responseId")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: respHeaders,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateSwiftPassword(res.ID, res.UserID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.Description, actual.Description)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestDeleteSwiftPassword() {
	opts := &IfMatchOptions{IfMatch: "abcd"}
	id := "id"
	userID := "userId"

	details := &requestDetails{
		ids:      urlParts{userID, resourceSwiftPasswords, id},
		name:     resourceUsers,
		optional: opts,
	}

	s.requestor.On("deleteRequest", details).Return(nil)

	e := s.requestor.DeleteSwiftPassword(id, userID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
}

func (s *IdentityTestSuite) TestListSwiftPasswords() {
	userID := "userID"

	details := &requestDetails{
		name: resourceUsers,
		ids:  urlParts{userID, resourceSwiftPasswords},
	}

	expected := ListSwiftPasswords{
		SwiftPasswords: []SwiftPassword{
			{
				ID:     "1",
				UserID: userID,
			},
			{
				ID:     "2",
				UserID: userID,
			},
		},
	}

	resp := &response{
		body: marshalObjectForTest(expected.SwiftPasswords),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListSwiftPasswords(userID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.SwiftPasswords), len(actual.SwiftPasswords))
}
