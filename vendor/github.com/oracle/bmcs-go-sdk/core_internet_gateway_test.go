// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func getTestInternetGateway(id string) *InternetGateway {
	conn := &InternetGateway{
		CompartmentID: "compartmentID",
		DisplayName:   "displayName",
		ID:            id,
		IsEnabled:     true,
		State:         ResourceAvailable,
		TimeCreated:   Time{Time: time.Now()},
	}
	conn.ETag = "etag"
	conn.RequestID = "requestid"

	return conn
}

func (s *CoreTestSuite) TestCreateInternetGateway() {
	res := getTestInternetGateway("id")

	opts := &CreateOptions{}
	opts.DisplayName = res.DisplayName

	required := struct {
		ocidRequirement
		IsEnabled bool   `header:"-" json:"isEnabled" url:"-"`
		VcnID     string `header:"-" json:"vcnId" url:"-"`
	}{
		IsEnabled: res.IsEnabled,
		VcnID:     "vcnID",
	}
	required.CompartmentID = res.CompartmentID

	details := &requestDetails{
		name:     resourceInternetGateways,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateInternetGateway(
		res.CompartmentID,
		"vcnID",
		res.IsEnabled,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetInternetGateway() {
	res := getTestInternetGateway("id")

	details := &requestDetails{
		name: resourceInternetGateways,
		ids:  urlParts{res.ID},
	}

	resp := &response{
		body: marshalObjectForTest(res),
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetInternetGateway(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
}

func (s *CoreTestSuite) TestDeleteInternetGateway() {
	s.testDeleteResource(resourceInternetGateways, "id", s.requestor.DeleteInternetGateway)
}

func (s *CoreTestSuite) TestUpdateInternetGateway() {
	res := getTestInternetGateway("id")

	opts := &UpdateGatewayOptions{}
	opts.IsEnabled = new(bool)
	*opts.IsEnabled = true
	opts.IfMatch = "etag"

	details := &requestDetails{
		ids:      urlParts{"id"},
		name:     resourceInternetGateways,
		optional: opts,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateInternetGateway("id", opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.IsEnabled, actual.IsEnabled)
}

func (s *CoreTestSuite) TestListInternetGateways() {
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "pageid"

	required := struct {
		listOCIDRequirement
		VcnID string `header:"-" json:"-" url:"vcnId"`
	}{
		VcnID: "vcnid",
	}
	required.CompartmentID = "compartmentid"

	details := &requestDetails{
		name:     resourceInternetGateways,
		optional: opts,
		required: required,
	}

	expected := []InternetGateway{*getTestInternetGateway("id1")}

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

	actual, e := s.requestor.ListInternetGateways("compartmentid", "vcnid", opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.Gateways))
}
