// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *DatabaseTestSuite) TestGetDatabase() {
	res := &Database{
		DBHomeID:     "homeID1",
		DBName:       "dbName",
		DBUniqueName: "dbUniq1",
		ID:           "id1",
		State:        ResourceAvailable,
		TimeCreated:  time.Now(),
	}

	details := &requestDetails{
		name: resourceDatabases,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetDatabase(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *DatabaseTestSuite) TestListDatabases() {
	reqs := struct {
		listOCIDRequirement
		DBHomeID string `header:"-" json:"-" url:"dbHomeId"`
		Limit    uint64 `header:"-" json:"-" url:"limit"`
	}{
		DBHomeID: "homeID",
		Limit:    100,
	}
	reqs.CompartmentID = "compartmentID"
	opts := &PageListOptions{}

	details := &requestDetails{
		name:     resourceDatabases,
		optional: opts,
		required: reqs,
	}

	expected := ListDatabases{
		Databases: []Database{
			{
				DBHomeID:     "homeID1",
				DBName:       "dbName",
				DBUniqueName: "dbUniq1",
				ID:           "id1",
				State:        ResourceAvailable,
				TimeCreated:  time.Now(),
			},
			{
				DBHomeID:     "homeID2",
				DBName:       "dbName",
				DBUniqueName: "dbUniq2",
				ID:           "id2",
				State:        ResourceAvailable,
				TimeCreated:  time.Now(),
			},
		},
	}

	resp := &response{
		body: marshalObjectForTest(expected.Databases),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListDatabases(reqs.CompartmentID, reqs.DBHomeID, reqs.Limit, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.Databases), len(actual.Databases))
}
