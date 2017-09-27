// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestCreateVirtualNetwork() {
	res := &VirtualNetwork{
		CidrBlock:             "cidrBlock",
		CompartmentID:         "compartmentID",
		DefaultRouteTableID:   "defaultRouteTableId",
		DefaultSecurityListID: "defaultSecurityListId",
		DisplayName:           "displayName",
		DnsLabel:              "dnsLabel",
		ID:                    "id1",
		State:                 ResourceProvisioning,
		TimeCreated:           Time{Time: time.Now()},
		VcnDomainName:         "vcnDomainName",
	}

	opts := &CreateVcnOptions{}
	opts.DisplayName = res.DisplayName
	opts.DnsLabel = res.DnsLabel

	required := struct {
		ocidRequirement
		CidrBlock string `header:"-" json:"cidrBlock,omitempty" url:"-"`
	}{
		CidrBlock: res.CidrBlock,
	}
	required.CompartmentID = res.CompartmentID

	details := &requestDetails{
		name:     resourceVirtualNetworks,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateVirtualNetwork(
		res.CidrBlock,
		res.CompartmentID,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetVirtualNetwork() {
	res := &VirtualNetwork{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourceVirtualNetworks,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetVirtualNetwork(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestUpdateVirtualNetwork() {
	res := &VirtualNetwork{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
		DisplayName: "displayName",
	}

	opts := &IfMatchDisplayNameOptions{
		DisplayNameOptions: DisplayNameOptions{DisplayName: res.DisplayName},
		IfMatchOptions:     IfMatchOptions{IfMatch: "ETAG"},
	}

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourceVirtualNetworks,
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateVirtualNetwork(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal(res.DisplayName, actual.DisplayName)
}

func (s *CoreTestSuite) TestDeleteVirtualNetwork() {
	s.testDeleteResource(resourceVirtualNetworks, "id", s.requestor.DeleteVirtualNetwork)
}

func (s *CoreTestSuite) TestListVirtualNetworks() {
	compartmentID := "compartmentid"
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "pageid"

	details := &requestDetails{
		name:     resourceVirtualNetworks,
		optional: opts,
		required: listOCIDRequirement{CompartmentID: compartmentID},
	}

	created := Time{Time: time.Now()}
	expected := []VirtualNetwork{
		{
			ID:            "id1",
			CompartmentID: compartmentID,
			DisplayName:   "vcn1",
			TimeCreated:   created,
		},
		{
			ID:            "id2",
			CompartmentID: compartmentID,
			DisplayName:   "vcn2",
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

	actual, e := s.requestor.ListVirtualNetworks(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.VirtualNetworks))
}
