// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestCreateDrg() {
	res := &Drg{
		CompartmentID: "compartmentID",
		DisplayName:   "displayName",
		ID:            "id1",
		State:         ResourceProvisioning,
		TimeCreated:   Time{Time: time.Now()},
	}

	opts := &CreateOptions{}
	opts.DisplayName = res.DisplayName

	details := &requestDetails{
		name:     resourceDrgs,
		optional: opts,
		required: ocidRequirement{"compartmentID"},
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateDrg(res.CompartmentID, opts)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetDrg() {
	res := &Drg{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		name: resourceDrgs,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetDrg(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestDeleteDrg() {
	s.testDeleteResource(resourceDrgs, "id", s.requestor.DeleteDrg)
}

func (s *CoreTestSuite) TestListDrgs() {
	compartmentID := "compartmentid"
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "pageid"

	details := &requestDetails{
		name:     resourceDrgs,
		optional: opts,
		required: listOCIDRequirement{compartmentID},
	}

	created := Time{Time: time.Now()}
	expected := []Drg{
		{
			ID:            "id1",
			CompartmentID: compartmentID,
			DisplayName:   "drg1",
			TimeCreated:   created,
		},
		{
			ID:            "id2",
			CompartmentID: compartmentID,
			DisplayName:   "drg2",
			TimeCreated:   created,
		},
	}

	responseHeaders := http.Header{}
	responseHeaders.Set(headerOPCNextPage, "nextpage")
	responseHeaders.Set(headerOPCRequestID, "requestid")

	s.requestor.On("getRequest", details).Return(
		&response{
			header: responseHeaders,
			body:   marshalObjectForTest(expected),
		},
		nil,
	)

	actual, e := s.requestor.ListDrgs(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.Drgs))
}
