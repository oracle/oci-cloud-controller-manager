// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *IdentityTestSuite) TestCreateUIPassword() {
	res := &UIPassword{
		InactiveStatus: 0,
		State:          "state",
		Password:       "password",
		TimeCreated:    Time{Time: time.Now()},
		UserID:         "user_id",
	}

	opts := &RetryTokenOptions{RetryToken: "token"}

	details := &requestDetails{
		ids:      urlParts{res.UserID, resourceUiPassword},
		name:     resourceUsers,
		optional: opts,
	}

	header := http.Header{}
	header.Set(headerETag, "ETAG!")
	header.Set(headerOPCRequestID, "898989")
	resp := &response{
		header: header,
		body:   marshalObjectForTest(res),
	}

	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, e := s.requestor.CreateOrResetUIPassword(res.UserID, opts)

	s.Nil(e)
	s.NotNil(actual)
	s.Equal(actual.Password, res.Password)
	s.Equal(actual.UserID, res.UserID)
	s.Equal(actual.ETag, "ETAG!")
	s.Equal(actual.RequestID, "898989")
}
