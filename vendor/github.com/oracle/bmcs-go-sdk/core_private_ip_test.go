// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestCreatePrivateIP() {
	res := &PrivateIP{
		AvailabilityDomain: "availabilityDomain",
		CompartmentID:      "compartmentID",
		DisplayName:        "displayName",
		HostnameLabel:      "hostnameLabel",
		ID:                 "id",
		IPAddress:          "ipAddress",
		IsPrimary:          false,
		SubnetID:           "subnetID",
		TimeCreated:        Time{Time: time.Now()},
		VnicID:             "vnicID",
	}

	opts := &CreatePrivateIPOptions{}
	opts.DisplayName = res.DisplayName
	opts.HostnameLabel = res.HostnameLabel
	opts.IPAddress = res.IPAddress

	required := struct {
		VnicId string `header:"-" json:"vnicId" url:"-"`
	}{
		VnicId: res.VnicID,
	}

	details := &requestDetails{
		name:     resourcePrivateIPs,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreatePrivateIP(
		res.VnicID,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetPrivateIP() {
	res := &PrivateIP{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourcePrivateIPs,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetPrivateIP(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestUpdatePrivateIP() {
	res := &PrivateIP{
		ID:            "id",
		TimeCreated:   Time{Time: time.Now()},
		DisplayName:   "displayName2",
		HostnameLabel: "hostnameLabel2",
		VnicID:        "vnicID2",
	}

	opts := &UpdatePrivateIPOptions{}
	opts.DisplayNameOptions = DisplayNameOptions{DisplayName: res.DisplayName}
	opts.IfMatchOptions = IfMatchOptions{IfMatch: "ETAG"}
	opts.HostnameLabel = res.HostnameLabel
	opts.VnicID = res.VnicID

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourcePrivateIPs,
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdatePrivateIP(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal(res.DisplayName, actual.DisplayName)
	s.Equal(res.HostnameLabel, actual.HostnameLabel)
	s.Equal(res.VnicID, actual.VnicID)
}

func (s *CoreTestSuite) TestDeletePrivateIP() {
	s.testDeleteResource(resourcePrivateIPs, "id", s.requestor.DeletePrivateIP)
}

func (s *CoreTestSuite) TestListPrivateIPs() {
	opts := &ListPrivateIPsOptions{}
	opts.Limit = 100
	opts.Page = "pageid"
	opts.IPAddress = "ipAddress"
	opts.SubnetID = "subnetId"
	opts.VnicID = "vnicId"

	details := &requestDetails{
		name:     resourcePrivateIPs,
		optional: opts,
	}

	created := Time{Time: time.Now()}
	expected := []PrivateIP{
		{
			ID:          "id1",
			DisplayName: "privateIP1",
			TimeCreated: created,
		},
		{
			ID:          "id2",
			DisplayName: "PrivateIP2",
			TimeCreated: created,
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

	actual, e := s.requestor.ListPrivateIPs(opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.PrivateIPs))
}
