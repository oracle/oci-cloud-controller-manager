// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestCreateDHCPOptions() {
	dhcpDnsOptions := []DHCPDNSOption{{}}
	expectedResponse := &DHCPOptions{
		CompartmentID: "compartmentID",
		DisplayName:   "displayName",
		ID:            "id1",
		Options:       dhcpDnsOptions,
		State:         ResourceAvailable,
		TimeCreated:   Time{Time: time.Now()},
		VcnID:         "vcn_id",
	}

	opts := &CreateOptions{}
	opts.DisplayName = expectedResponse.DisplayName

	required := struct {
		ocidRequirement
		Options []DHCPDNSOption `header:"-" json:"options" url:"-"`
		VcnID   string          `header:"-" json:"vcnId" url:"-"`
	}{
		Options: expectedResponse.Options,
		VcnID:   "vcn_id",
	}
	required.CompartmentID = "compartmentID"

	reqDetails := &requestDetails{
		name:     resourceDHCPOptions,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(expectedResponse),
	}
	s.requestor.On("postRequest", reqDetails).Return(resp, nil)

	actualResponse, err := s.requestor.CreateDHCPOptions(expectedResponse.CompartmentID, expectedResponse.VcnID, expectedResponse.Options, opts)

	s.Nil(err)
	s.NotNil(actualResponse)
	s.Equal(expectedResponse.CompartmentID, actualResponse.CompartmentID)
	s.Equal(expectedResponse.VcnID, actualResponse.VcnID)
	s.Equal(expectedResponse.ID, actualResponse.ID)
	s.Equal(expectedResponse.DisplayName, actualResponse.DisplayName)
}

func (s *CoreTestSuite) TestGetDHCPOptions() {
	expectedResponse := &DHCPOptions{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
		VcnID:       "vcn_id",
	}

	reqDetails := &requestDetails{
		name: resourceDHCPOptions,
		ids:  urlParts{expectedResponse.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(expectedResponse),
		header: headers,
	}

	s.requestor.On("getRequest", reqDetails).Return(resp, nil)

	actualResponse, e := s.requestor.GetDHCPOptions(expectedResponse.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actualResponse)
	s.Equal(expectedResponse.ID, actualResponse.ID)
	s.Equal(expectedResponse.VcnID, actualResponse.VcnID)
	s.Equal("ETAG", actualResponse.ETag)
}

func (s *CoreTestSuite) TestUpdateDHCPOptions() {
	dhcpDnsOptions := []DHCPDNSOption{{}}
	expectedResponse := &DHCPOptions{
		ID:          "id",
		Options:     dhcpDnsOptions,
		TimeCreated: Time{Time: time.Now()},
		VcnID:       "vcn_id",
	}

	opts := &UpdateDHCPDNSOptions{
		Options: expectedResponse.Options,
	}

	reqDetails := &requestDetails{
		name:     resourceDHCPOptions,
		ids:      urlParts{expectedResponse.ID},
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(expectedResponse),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, reqDetails).Return(resp, nil)

	actualResponse, e := s.requestor.UpdateDHCPOptions(expectedResponse.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actualResponse)
	s.Equal(expectedResponse.Options, actualResponse.Options)
	s.Equal("ETAG!", actualResponse.ETag)
	s.Equal(expectedResponse.VcnID, actualResponse.VcnID)
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
