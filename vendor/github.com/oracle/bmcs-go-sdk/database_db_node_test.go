// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *DatabaseTestSuite) TestGetDBNode() {
	res := &DBNode{
		DBSystemID:  "id1",
		Hostname:    "host1",
		ID:          "id1",
		State:       ResourceProvisioning,
		TimeCreated: time.Now(),
		VnicID:      "vnic1",
	}

	details := &requestDetails{
		name: resourceDBNodes,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetDBNode(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *DatabaseTestSuite) TestDBNodeAction() {
	res := &DBNode{
		DBSystemID:  "id1",
		Hostname:    "host1",
		ID:          "id1",
		State:       ResourceActive,
		TimeCreated: time.Now(),
		VnicID:      "vnic1",
	}

	required := struct {
		Action string `header:"-" json:"-" url:"action"`
	}{
		Action: string(DBNodeActionStart),
	}

	details := &requestDetails{
		name:     resourceDBNodes,
		ids:      urlParts{res.ID},
		required: required,
		optional: (*HeaderOptions)(nil),
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, e := s.requestor.DBNodeAction(res.ID, DBNodeActionStart, nil)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *DatabaseTestSuite) TestListDBNodes() {
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
		name:     resourceDBNodes,
		optional: &opts,
		required: reqs,
	}

	expected := ListDBNodes{
		DBNodes: []DBNode{
			{
				DBSystemID:  "id1",
				Hostname:    "host1",
				ID:          "id1",
				State:       ResourceProvisioning,
				TimeCreated: time.Now(),
				VnicID:      "vnic1",
			},
			{
				DBSystemID:  "id2",
				Hostname:    "host2",
				ID:          "id2",
				State:       ResourceActive,
				TimeCreated: time.Now(),
				VnicID:      "vnic2",
			},
		},
	}

	resp := &response{
		body: marshalObjectForTest(expected.DBNodes),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListDBNodes(reqs.CompartmentID, reqs.DBSystemID, &opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.DBNodes), len(actual.DBNodes))
}
