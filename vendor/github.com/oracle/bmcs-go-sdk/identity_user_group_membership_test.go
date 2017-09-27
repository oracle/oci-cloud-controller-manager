// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

func (s *IdentityTestSuite) TestAddUserToGroup() {
	res := UserGroupMembership{
		ID:            "1",
		CompartmentID: s.requestor.authInfo.tenancyOCID,
		GroupID:       "groupid",
		UserID:        "userid",
	}

	opts := &RetryTokenOptions{RetryToken: "aaaa"}

	required := struct {
		GroupID string `header:"-" json:"groupId" url:"-"`
		UserID  string `header:"-" json:"userId" url:"-"`
	}{
		GroupID: res.GroupID,
		UserID:  res.UserID,
	}

	details := &requestDetails{
		name:     resourceUserGroupMemberships,
		optional: opts,
		required: required,
	}

	header := http.Header{}
	header.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: header,
	}

	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, e := s.requestor.AddUserToGroup(res.UserID, res.GroupID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *IdentityTestSuite) TestGetUserGroupMembership() {
	res := UserGroupMembership{
		ID:      "1234",
		GroupID: "4567",
		UserID:  "bob123",
	}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourceUserGroupMemberships,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetUserGroupMembership("1234")

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal("1234", actual.ID)
	s.Equal(res.GroupID, actual.GroupID)
	s.Equal(res.UserID, actual.UserID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestDeleteUserGroupMembership() {
	s.testDeleteResource(resourceUserGroupMemberships, "666", s.requestor.DeleteUserGroupMembership)
}

func (s *IdentityTestSuite) TestListUserGroupMemberships() {
	opts := &ListMembershipsOptions{}
	opts.GroupID = "2"
	opts.Limit = 100
	opts.Page = "A"
	opts.UserID = "1"

	details := &requestDetails{
		name:     resourceUserGroupMemberships,
		optional: opts,
		required: ocidRequirement{s.requestor.authInfo.tenancyOCID},
	}

	resources := []UserGroupMembership{
		{
			ID:            "1",
			CompartmentID: s.requestor.authInfo.tenancyOCID,
			GroupID:       "1",
			UserID:        "1",
		},
		{
			ID:            "2",
			CompartmentID: s.requestor.authInfo.tenancyOCID,
			GroupID:       "1",
			UserID:        "1",
		},
	}

	resp := &response{body: marshalObjectForTest(resources)}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListUserGroupMemberships(opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(resources), len(actual.Memberships))
}
