// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"os/exec"
	"time"
)

func getTestSubnet() *Subnet {
	buff, _ := exec.Command("uuidgen").Output()
	sn := &Subnet{
		AvailabilityDomain: "availabilityDomain",
		CompartmentID:      "compartmentID",
		CIDRBlock:          "10.10.10.0/24",
		DisplayName:        "displayName",
		ID:                 string(buff),
		RouteTableID:       "routetableid",
		SecurityListIDs:    []string{"id1", "id2"},
		State:              ResourceAvailable,
		TimeCreated:        Time{Time: time.Now()},
		VcnID:              "vcnid",
		VirtualRouterIP:    "routerid",
		VirtualRouterMac:   "routermac",
	}
	sn.ETag = "etag"
	sn.RequestID = "requestid"
	return sn
}

func (s *CoreTestSuite) TestCreateSubnet() {
	res := getTestSubnet()

	opts := &CreateSubnetOptions{}
	opts.DisplayName = res.DisplayName
	opts.RouteTableID = res.RouteTableID
	opts.SecurityListIDs = res.SecurityListIDs

	required := struct {
		ocidRequirement
		AvailabilityDomain string `header:"-" json:"availabilityDomain" url:"-"`
		CIDRBlock          string `header:"-" json:"cidrBlock" url:"-"`
		VcnID              string `header:"-" json:"vcnId" url:"-"`
	}{
		AvailabilityDomain: res.AvailabilityDomain,
		CIDRBlock:          res.CIDRBlock,
		VcnID:              res.VcnID,
	}
	required.CompartmentID = res.CompartmentID

	details := &requestDetails{
		name:     resourceSubnets,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateSubnet(
		res.AvailabilityDomain,
		res.CIDRBlock,
		res.CompartmentID,
		res.VcnID,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetSubnet() {
	res := getTestSubnet()

	details := &requestDetails{
		name: resourceSubnets,
		ids:  urlParts{res.ID},
	}

	resp := &response{body: marshalObjectForTest(res)}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetSubnet(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
}

func (s *CoreTestSuite) TestDeleteSubnet() {
	s.testDeleteResource(resourceSubnets, "id", s.requestor.DeleteSubnet)
}

func (s *CoreTestSuite) TestListSubnets() {
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "pageid"

	required := struct {
		listOCIDRequirement
		VcnID string `header:"-" json:"-" url:"vcn"`
	}{
		VcnID: "vcnID",
	}
	required.CompartmentID = "compartmentID"

	details := &requestDetails{
		name:     resourceSubnets,
		optional: opts,
		required: required,
	}

	expected := []Subnet{
		*getTestSubnet(),
		*getTestSubnet(),
		*getTestSubnet(),
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

	actual, e := s.requestor.ListSubnets("compartmentID", "vcnID", opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.Subnets))
}
