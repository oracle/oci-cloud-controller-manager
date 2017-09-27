// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

func (s *IdentityTestSuite) TestListAPIKeys() {
	userID := "1ABC"
	headerVal := "abcde"
	details := &requestDetails{
		name: resourceUsers,
		ids:  urlParts{userID, apiKeys, "/"},
	}

	headers := http.Header{}
	headers.Set(headerOPCRequestID, headerVal)
	keys := []APIKey{
		{
			UserID:   userID,
			KeyID:    "2",
			KeyValue: "DEADBEEF",
		},
	}
	getResp := &response{
		header: headers,
		body:   marshalObjectForTest(keys),
	}

	s.requestor.On("getRequest", details).Return(getResp, nil)

	actual, e := s.requestor.ListAPIKeys(userID)
	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(keys), len(actual.Keys))
	s.Equal(headerVal, actual.RequestID)
}

func (s *IdentityTestSuite) TestDeleteAPIKey() {
	fingerprint := "b4:8a:7d:54:e6:81:04:b2:99:8e:b3:ed:10:e2:12:2b"
	userID := "123456"
	opts := &IfMatchOptions{IfMatch: "abcd"}

	details := &requestDetails{
		ids:      urlParts{userID, apiKeys, fingerprint},
		name:     resourceUsers,
		optional: opts,
	}

	s.requestor.On("deleteRequest", details).Return(nil)

	e := s.requestor.DeleteAPIKey(userID, fingerprint, opts)
	s.Nil(e)
}

func (s *IdentityTestSuite) TestUploadAPIKey() {
	userID := "123"
	key := "DEADBEEF"

	opts := &RetryTokenOptions{RetryToken: "abc"}

	required := struct {
		Key string `header:"-" json:"key" url:"-"`
	}{
		Key: key,
	}

	details := &requestDetails{
		ids:      urlParts{userID, apiKeys, "/"},
		name:     resourceUsers,
		optional: opts,
		required: required,
	}

	expected := APIKey{UserID: userID, KeyID: "2"}
	resp := &response{body: marshalObjectForTest(expected)}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, e := s.requestor.UploadAPIKey(userID, key, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(actual.UserID, expected.UserID)

}
