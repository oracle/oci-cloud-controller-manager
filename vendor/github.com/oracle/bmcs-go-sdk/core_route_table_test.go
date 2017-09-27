// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestCreateRouteTable() {
	t := Time{Time: time.Now()}
	rule := RouteRule{
		CidrBlock:       "cidr_block",
		NetworkEntityID: "network_entity_id",
	}

	res := &RouteTable{
		CompartmentID: "compartmentID",
		DisplayName:   "displayName",
		ID:            "id1",
		TimeModified:  t,
		RouteRules:    []RouteRule{rule},
		State:         ResourceProvisioning,
		TimeCreated:   t,
	}

	opts := &CreateOptions{}
	opts.DisplayName = res.DisplayName

	required := struct {
		ocidRequirement
		RouteRules []RouteRule `header:"-" json:"routeRules" url:"-"`
		VcnID      string      `header:"-" json:"vcnId" url:"-"`
	}{
		RouteRules: res.RouteRules,
		VcnID:      "vcnID",
	}
	required.CompartmentID = "compartmentID"

	details := &requestDetails{
		name:     resourceRouteTables,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateRouteTable(
		res.CompartmentID,
		"vcnID",
		res.RouteRules,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetRouteTable() {
	res := &RouteTable{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		name: resourceRouteTables,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetRouteTable(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestUpdateRouteTable() {
	t := Time{Time: time.Now()}
	rules := []RouteRule{
		{
			CidrBlock:       "cidr_block",
			NetworkEntityID: "network_entity_id",
		},
	}

	res := &RouteTable{
		ID:          "id",
		RouteRules:  rules,
		TimeCreated: t,
	}

	opts := &UpdateRouteTableOptions{RouteRules: rules}

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourceRouteTables,
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateRouteTable(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(1, len(actual.RouteRules))
	s.Equal("cidr_block", actual.RouteRules[0].CidrBlock)
	s.Equal("ETAG!", actual.ETag)
}

func (s *CoreTestSuite) TestDeleteRouteTable() {
	s.testDeleteResource(resourceRouteTables, "id", s.requestor.DeleteRouteTable)
}

func (s *CoreTestSuite) TestListRouteTables() {
	created := Time{Time: time.Now()}
	compartmentID := "compartment_id"
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "page_id"

	required := struct {
		listOCIDRequirement
		VcnID string `header:"-" json:"-" url:"vcnId"`
	}{
		VcnID: "vcnId",
	}
	required.CompartmentID = compartmentID

	details := &requestDetails{
		name:     resourceRouteTables,
		optional: opts,
		required: required,
	}

	expected := []RouteTable{
		{
			ID:            "id1",
			CompartmentID: compartmentID,
			DisplayName:   "res1",
			TimeCreated:   created,
		},
		{
			ID:            "id2",
			CompartmentID: compartmentID,
			DisplayName:   "res2",
			TimeCreated:   created,
		},
	}

	headers := http.Header{}
	headers.Set(headerOPCNextPage, "nextpage")
	headers.Set(headerOPCRequestID, "requestid")

	s.requestor.On("getRequest", details).Return(
		&response{
			header: headers,
			body:   marshalObjectForTest(expected),
		},
		nil,
	)

	actual, e := s.requestor.ListRouteTables(compartmentID, "vcnId", opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.RouteTables))
}
