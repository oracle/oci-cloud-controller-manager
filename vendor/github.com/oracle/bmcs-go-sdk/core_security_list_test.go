// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

var testSecurityList = SecurityList{
	CompartmentID: "compartmentID",
	DisplayName:   "displayName",
	EgressSecurityRules: []EgressSecurityRule{
		{
			Destination: "destination",
			ICMPOptions: &ICMPOptions{
				Code: 0,
				Type: 1,
			},
			Protocol: "protocol",
			TCPOptions: &TCPOptions{
				DestinationPortRange: PortRange{Max: 2, Min: 1},
			},
			UDPOptions: &UDPOptions{
				DestinationPortRange: PortRange{Max: 2, Min: 1},
			},
		},
	},
	ID: "id1",
	IngressSecurityRules: []IngressSecurityRule{
		{
			ICMPOptions: &ICMPOptions{
				Code: 0,
				Type: 1,
			},
			Protocol: "protocol",
			Source:   "source",
			TCPOptions: &TCPOptions{
				DestinationPortRange: PortRange{Max: 2, Min: 1},
			},
			UDPOptions: &UDPOptions{
				DestinationPortRange: PortRange{Max: 2, Min: 1},
			},
		},
	},
	State:       ResourceProvisioning,
	TimeCreated: Time{Time: time.Now()},
	VcnID:       "vcn_id",
}

func (s *CoreTestSuite) TestCreateSecurityList() {
	res := testSecurityList

	required := struct {
		ocidRequirement
		EgressRules  []EgressSecurityRule  `header:"-" json:"egressSecurityRules" url:"-"`
		IngressRules []IngressSecurityRule `header:"-" json:"ingressSecurityRules" url:"-"`
		VcnID        string                `header:"-" json:"vcnId" url:"-"`
	}{
		EgressRules:  res.EgressSecurityRules,
		IngressRules: res.IngressSecurityRules,
		VcnID:        res.VcnID,
	}
	required.CompartmentID = res.CompartmentID

	opts := &CreateOptions{}
	opts.DisplayName = res.DisplayName

	details := &requestDetails{
		name:     resourceSecurityLists,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateSecurityList(
		res.CompartmentID,
		res.VcnID,
		res.EgressSecurityRules,
		res.IngressSecurityRules,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetSecurityList() {
	expected := &SecurityList{ID: "id", TimeCreated: Time{Time: time.Now()}}

	details := &requestDetails{
		name: resourceSecurityLists,
		ids:  urlParts{expected.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(expected),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetSecurityList(expected.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(expected.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestUpdateSecurityList() {
	res := testSecurityList

	opts := &UpdateSecurityListOptions{}
	opts.DisplayName = res.DisplayName
	opts.EgressRules = res.EgressSecurityRules
	opts.IngressRules = res.IngressSecurityRules

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourceSecurityLists,
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateSecurityList(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.DisplayName, actual.DisplayName)
	s.Equal("ETAG!", actual.ETag)
}

func (s *CoreTestSuite) TestDeleteSecurityList() {
	s.testDeleteResource(resourceSecurityLists, "id", s.requestor.DeleteSecurityList)
}

func (s *CoreTestSuite) TestListSecurityLists() {
	compartmentID := "compartmentid"

	required := struct {
		listOCIDRequirement
		VcnID string `header:"-" json:"-" url:"vcnId"`
	}{
		VcnID: "vcn_id",
	}
	required.CompartmentID = compartmentID

	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "pageid"

	details := &requestDetails{
		name:     resourceSecurityLists,
		optional: opts,
		required: required,
	}

	res := testSecurityList
	res2 := res
	res2.ID = "id2"

	expected := []SecurityList{res, res2}

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

	actual, e := s.requestor.ListSecurityLists(compartmentID, "vcn_id", opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.SecurityLists))
}
