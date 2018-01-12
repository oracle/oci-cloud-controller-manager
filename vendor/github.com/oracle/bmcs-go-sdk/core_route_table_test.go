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

	expectedResponse := &RouteTable{
		CompartmentID: "compartmentID",
		DisplayName:   "displayName",
		ID:            "id1",
		TimeModified:  t,
		RouteRules:    []RouteRule{rule},
		State:         ResourceProvisioning,
		TimeCreated:   t,
		VcnID:         "vcn_id",
	}

	opts := &CreateOptions{}
	opts.DisplayName = expectedResponse.DisplayName

	required := struct {
		ocidRequirement
		RouteRules []RouteRule `header:"-" json:"routeRules" url:"-"`
		VcnID      string      `header:"-" json:"vcnId" url:"-"`
	}{
		RouteRules: expectedResponse.RouteRules,
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
		body:   marshalObjectForTest(expectedResponse),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actualResponse, err := s.requestor.CreateRouteTable(
		expectedResponse.CompartmentID,
		"vcnID",
		expectedResponse.RouteRules,
		opts,
	)

	s.Nil(err)
	s.NotNil(actualResponse)
	s.Equal(expectedResponse.CompartmentID, actualResponse.CompartmentID)
	s.Equal(expectedResponse.VcnID, actualResponse.VcnID)
}

func (s *CoreTestSuite) TestGetRouteTable() {
	expectedResponse := &RouteTable{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
		VcnID:       "vcn_id",
	}

	details := &requestDetails{
		name: resourceRouteTables,
		ids:  urlParts{expectedResponse.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(expectedResponse),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actualResponse, e := s.requestor.GetRouteTable(expectedResponse.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actualResponse)
	s.Equal(expectedResponse.ID, actualResponse.ID)
	s.Equal(expectedResponse.VcnID, actualResponse.VcnID)
	s.Equal("ETAG", actualResponse.ETag)
}

func (s *CoreTestSuite) TestUpdateRouteTable() {
	t := Time{Time: time.Now()}
	rules := []RouteRule{
		{
			CidrBlock:       "cidr_block",
			NetworkEntityID: "network_entity_id",
		},
	}

	expectedResponse := &RouteTable{
		ID:          "id",
		RouteRules:  rules,
		TimeCreated: t,
		VcnID:       "vcn_id",
	}

	opts := &UpdateRouteTableOptions{RouteRules: rules}

	details := &requestDetails{
		ids:      urlParts{expectedResponse.ID},
		name:     resourceRouteTables,
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(expectedResponse),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actualResponse, e := s.requestor.UpdateRouteTable(expectedResponse.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actualResponse)
	s.Equal(1, len(actualResponse.RouteRules))
	s.Equal("cidr_block", actualResponse.RouteRules[0].CidrBlock)
	s.Equal("ETAG!", actualResponse.ETag)
	s.Equal(expectedResponse.VcnID, actualResponse.VcnID)
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

	expectedResponse := []RouteTable{
		{
			ID:            "id1",
			CompartmentID: compartmentID,
			DisplayName:   "res1",
			TimeCreated:   created,
			VcnID:         "vcnId",
		},
		{
			ID:            "id2",
			CompartmentID: compartmentID,
			DisplayName:   "res2",
			TimeCreated:   created,
			VcnID:         "vcnId",
		},
	}

	headers := http.Header{}
	headers.Set(headerOPCNextPage, "nextpage")
	headers.Set(headerOPCRequestID, "requestid")

	s.requestor.On("getRequest", details).Return(
		&response{
			header: headers,
			body:   marshalObjectForTest(expectedResponse),
		},
		nil,
	)

	actualResponse, e := s.requestor.ListRouteTables(compartmentID, "vcnId", opts)
	s.Nil(e)
	s.NotNil(actualResponse)
	s.Equal(len(expectedResponse), len(actualResponse.RouteTables))
	s.Equal(expectedResponse[0].VcnID, actualResponse.RouteTables[0].VcnID)
	s.Equal(expectedResponse[1].VcnID, actualResponse.RouteTables[1].VcnID)
}
