// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestCreateDHCPOptions() {
	dhcpOptions := []DHCPDNSOption{{}}
	res := &DHCPOptions{
		CompartmentID: "compartmentID",
		DisplayName:   "displayName",
		ID:            "id1",
		Options:       dhcpOptions,
		State:         ResourceAvailable,
		TimeCreated:   Time{Time: time.Now()},
	}

	opts := &CreateOptions{}
	opts.DisplayName = res.DisplayName

	required := struct {
		ocidRequirement
		Options []DHCPDNSOption `header:"-" json:"options" url:"-"`
		VcnID   string          `header:"-" json:"vcnId" url:"-"`
	}{
		Options: res.Options,
		VcnID:   "vcn_id",
	}
	required.CompartmentID = "compartmentID"

	details := &requestDetails{
		name:     resourceDHCPOptions,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateDHCPOptions(res.CompartmentID, "vcn_id", res.Options, opts)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetDHCPOptions() {
	res := &DHCPOptions{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		name: resourceDHCPOptions,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetDHCPOptions(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestUpdateDHCPOptions() {
	dhcpOptions := []DHCPDNSOption{{}}
	res := &DHCPOptions{
		ID:          "id",
		Options:     dhcpOptions,
		TimeCreated: Time{Time: time.Now()},
	}

	opts := &UpdateDHCPDNSOptions{
		Options: res.Options,
	}

	details := &requestDetails{
		name:     resourceDHCPOptions,
		ids:      urlParts{res.ID},
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateDHCPOptions(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.Options, actual.Options)
	s.Equal("ETAG!", actual.ETag)
}

func (s *CoreTestSuite) TestDeleteDHCPOptions() {
	s.testDeleteResource(resourceDHCPOptions, "id", s.requestor.DeleteDHCPOptions)
}

func (s *CoreTestSuite) TestListDHCPOptions() {
	compartmentID := "compartmentid"
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "pageid"

	required := struct {
		listOCIDRequirement
		VcnID string `header:"-" json:"-" url:"vcnId"`
	}{
		VcnID: "vcn_id",
	}
	required.CompartmentID = compartmentID

	details := &requestDetails{
		name:     resourceDHCPOptions,
		optional: opts,
		required: required,
	}

	created := Time{Time: time.Now()}
	expected := []DHCPOptions{
		{
			ID:            "id1",
			CompartmentID: compartmentID,
			TimeCreated:   created,
		},
		{
			ID:            "id2",
			CompartmentID: compartmentID,
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

	actual, e := s.requestor.ListDHCPOptions(compartmentID, "vcn_id", opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.DHCPOptions))
}
