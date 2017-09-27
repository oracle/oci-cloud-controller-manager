// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestLaunchInstance() {
	metadata := map[string]string{"foo": "bar"}
	extendedMetadata := make(map[string]interface{})
	extendedMetadata["key"] = map[string]string{"key1": "value1"}
	created := Time{Time: time.Now()}
	ipxeScript := "#!ipxe\nchain http://boot.ipxe.org/demo/boot.php"

	res := &Instance{
		AvailabilityDomain: "availabilityDomain",
		CompartmentID:      "compartmentID",
		DisplayName:        "displayName",
		ID:                 "id1",
		ImageID:            "image1",
		Metadata:           metadata,
		ExtendedMetadata:   extendedMetadata,
		IpxeScript:         ipxeScript,
		Region:             "Perth",
		Shape:              "x5-2.36.512.nvme-6.4",
		State:              ResourceAvailable,
		TimeCreated:        created,
	}

	opts := &LaunchInstanceOptions{}
	opts.DisplayName = res.DisplayName
	opts.Metadata = res.Metadata
	opts.ExtendedMetadata = res.ExtendedMetadata
	opts.IpxeScript = res.IpxeScript

	required := struct {
		ocidRequirement
		AvailabilityDomain string `header:"-" json:"availabilityDomain" url:"-"`
		ImageID            string `header:"-" json:"imageId" url:"-"`
		Shape              string `header:"-" json:"shape" url:"-"`
		SubnetID           string `header:"-" json:"subnetId,omitempty" url:"-"`
	}{
		AvailabilityDomain: res.AvailabilityDomain,
		ImageID:            res.ImageID,
		Shape:              res.Shape,
		SubnetID:           "subnetid",
	}
	required.CompartmentID = res.CompartmentID

	details := &requestDetails{
		name:     resourceInstances,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.LaunchInstance(
		res.AvailabilityDomain,
		res.CompartmentID,
		res.ImageID,
		res.Shape,
		"subnetid",
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
	s.Equal(res.IpxeScript, actual.IpxeScript)
}

func (s *CoreTestSuite) TestGetInstance() {
	res := &Instance{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		name: resourceInstances,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetInstance(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestGetWindowsInstanceInitialCredentials() {
	res := &InstanceCredentials{
		Username: "username",
		Password: "password",
	}
	details := &requestDetails{
		name: resourceInstances,
		ids:  urlParts{"instanceId", "initialCredentials"},
	}
	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}
	s.requestor.On("getRequest", details).Return(resp, nil)
	actual, e := s.requestor.GetWindowsInstanceInitialCredentials("instanceId")
	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal("username", actual.Username)
}

func (s *CoreTestSuite) TestInstanceAction() {
	res := &Instance{
		ID:          "id",
		DisplayName: "displayName",
		TimeCreated: Time{Time: time.Now()},
	}

	required := struct {
		Action string `header:"-" json:"-" url:"action"`
	}{
		Action: string(actionStart),
	}

	details := &requestDetails{
		name:     resourceInstances,
		ids:      urlParts{res.ID},
		required: required,
		optional: (*HeaderOptions)(nil),
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, e := s.requestor.InstanceAction(res.ID, actionStart, nil)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *CoreTestSuite) TestUpdateInstance() {
	res := &Instance{
		ID:          "id",
		DisplayName: "displayName",
		TimeCreated: Time{Time: time.Now()},
	}

	opts := &UpdateOptions{}
	opts.DisplayName = res.DisplayName

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourceInstances,
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateInstance(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.DisplayName, actual.DisplayName)
	s.Equal("ETAG!", actual.ETag)
}

func (s *CoreTestSuite) TestTerminateInstance() {
	s.testDeleteResource(resourceInstances, "id", s.requestor.TerminateInstance)
}

func (s *CoreTestSuite) TestListInstances() {
	compartmentID := "compartmentid"
	opts := &ListInstancesOptions{}
	opts.Limit = 100
	opts.Page = "pageid"

	details := &requestDetails{
		name:     resourceInstances,
		optional: opts,
		required: listOCIDRequirement{CompartmentID: compartmentID},
	}

	created := Time{Time: time.Now()}
	expected := []Instance{
		{
			ID:            "id1",
			CompartmentID: compartmentID,
			DisplayName:   "one",
			TimeCreated:   created,
		},
		{
			ID:            "id2",
			CompartmentID: compartmentID,
			DisplayName:   "two",
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

	actual, e := s.requestor.ListInstances(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.Instances))
}
