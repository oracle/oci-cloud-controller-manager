// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

func (s *DatabaseTestSuite) TestGetDBSystem() {
	res := &DBSystem{ID: "id1"}

	details := &requestDetails{
		name: resourceDBSystems,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetDBSystem(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *DatabaseTestSuite) TestTerminateDBSystem() {
	s.testDeleteResource(resourceDBSystems, "id", s.requestor.TerminateDBSystem)
}

func (s *DatabaseTestSuite) TestListDBSystems() {
	required := struct {
		listOCIDRequirement
	}{}
	required.CompartmentID = "compartmentID"

	opts := &ListOptions{}
	opts.Limit = 100

	details := &requestDetails{
		name:     resourceDBSystems,
		optional: opts,
		required: required,
	}

	expected := ListDBSystems{
		DBSystems: []DBSystem{{ID: "id1"}, {ID: "id2"}},
	}

	resp := &response{
		body: marshalObjectForTest(expected.DBSystems),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListDBSystems(required.CompartmentID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.DBSystems), len(actual.DBSystems))
}
