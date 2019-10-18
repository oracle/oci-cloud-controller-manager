// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *DatabaseTestSuite) TestGetDBHome() {
	res := &DBHome{
		DBSystemID:  "id1",
		DBVersion:   "vers1",
		DisplayName: "blah-home",
		ID:          "id1",
		State:       ResourceAvailable,
		TimeCreated: time.Now(),
	}

	details := &requestDetails{
		name: resourceDBHomes,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetDBHome(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *DatabaseTestSuite) TestListDBHomes() {
	reqs := struct {
		listOCIDRequirement
		DBSystemID string `header:"-" json:"-" url:"dbSystemId"`
	}{
		DBSystemID: "blahID",
	}
	reqs.CompartmentID = "compartmentID"
	opts := ListOptions{}
	opts.Limit = 100

	details := &requestDetails{
		name:     resourceDBHomes,
		optional: &opts,
		required: reqs,
	}

	expected := ListDBHomes{
		DBHomes: []DBHome{
			{
				DBSystemID:  "id1",
				DBVersion:   "vers1",
				DisplayName: "blah-home",
				ID:          "id1",
				State:       ResourceAvailable,
				TimeCreated: time.Now(),
			},
			{
				DBSystemID:  "id1",
				DBVersion:   "vers2",
				DisplayName: "blah-home",
				ID:          "id2",
				State:       ResourceAvailable,
				TimeCreated: time.Now(),
			},
		},
	}

	resp := &response{
		body: marshalObjectForTest(expected.DBHomes),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListDBHomes(reqs.CompartmentID, reqs.DBSystemID, &opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.DBHomes), len(actual.DBHomes))
}
