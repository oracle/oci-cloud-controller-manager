// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestGetVnic() {
	vnicID := "vnicid"
	res := &Vnic{
		AvailabilityDomain:  "availabilitydomain",
		CompartmentID:       "compartmentid",
		DisplayName:         "displayname",
		ID:                  vnicID,
		IsPrimary:           true,
		MacAddress:          "00:00:17:B6:4D:DD",
		State:               ResourceAvailable,
		PrivateIPAddress:    "10.10.10.10",
		PublicIPAddress:     "54.55.56.57",
		SkipSourceDestCheck: false,
		SubnetID:            "subnetid",
		TimeCreated:         Time{Time: time.Now()},
	}

	details := &requestDetails{
		name: resourceVnics,
		ids:  urlParts{res.ID},
	}

	resp := &response{body: marshalObjectForTest(res)}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetVnic(vnicID)
	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal(res.PublicIPAddress, actual.PublicIPAddress)
	s.Equal(res.PrivateIPAddress, actual.PrivateIPAddress)
}

func (s *CoreTestSuite) TestUpdateVnic() {
	vnicID := "vnicid"
	res := &Vnic{
		AvailabilityDomain:  "availabilitydomain",
		CompartmentID:       "compartmentid",
		DisplayName:         "displayname",
		ID:                  vnicID,
		IsPrimary:           true,
		MacAddress:          "00:00:17:B6:4D:DD",
		State:               ResourceAvailable,
		PrivateIPAddress:    "10.10.10.10",
		PublicIPAddress:     "54.55.56.57",
		SkipSourceDestCheck: true,
		SubnetID:            "subnetid",
		TimeCreated:         Time{Time: time.Now()},
	}

	opts := &UpdateVnicOptions{}
	opts.DisplayNameOptions = DisplayNameOptions{DisplayName: res.DisplayName}
	skipSourceDestCheck := res.SkipSourceDestCheck
	opts.SkipSourceDestCheck = &skipSourceDestCheck
	opts.IfMatchOptions = IfMatchOptions{IfMatch: "my_tag"}

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourceVnics,
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "my_tag_update")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateVnic(vnicID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal(res.PublicIPAddress, actual.PublicIPAddress)
	s.Equal(res.PrivateIPAddress, actual.PrivateIPAddress)
	s.Equal(res.SkipSourceDestCheck, actual.SkipSourceDestCheck)
}
