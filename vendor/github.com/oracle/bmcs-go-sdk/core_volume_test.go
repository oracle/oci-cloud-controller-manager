// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestCreateVolume() {
	res := &Volume{
		AvailabilityDomain: "availabilityDomain",
		CompartmentID:      "compartmentID",
		DisplayName:        "displayName",
		ID:                 "id1",
		SizeInGBs:          50,
		State:              ResourceProvisioning,
		TimeCreated:        Time{Time: time.Now()},
	}

	opts := &CreateVolumeOptions{}
	opts.DisplayName = res.DisplayName

	required := struct {
		ocidRequirement
		AvailabilityDomain string `header:"-" json:"availabilityDomain" url:"-"`
	}{
		AvailabilityDomain: res.AvailabilityDomain,
	}
	required.CompartmentID = res.CompartmentID

	details := &requestDetails{
		name:     resourceVolumes,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateVolume(
		res.AvailabilityDomain,
		res.CompartmentID,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetVolume() {
	res := &Volume{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		name: resourceVolumes,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetVolume(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestUpdateVolume() {
	res := &Volume{
		ID:          "id",
		DisplayName: "displayName",
		TimeCreated: Time{Time: time.Now()},
	}

	opts := &UpdateOptions{}
	opts.DisplayName = res.DisplayName

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourceVolumes,
		optional: opts,
	}

	respHeaders := http.Header{}
	respHeaders.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: respHeaders,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateVolume(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.DisplayName, actual.DisplayName)
	s.Equal("ETAG!", actual.ETag)
}

func (s *CoreTestSuite) TestDeleteVolume() {
	s.testDeleteResource(resourceVolumes, "id", s.requestor.DeleteVolume)
}

func (s *CoreTestSuite) TestListVolumes() {
	compartmentID := "compartmentid"
	opts := &ListVolumesOptions{}
	opts.Limit = 100
	opts.Page = "pageid"

	details := &requestDetails{
		name:     resourceVolumes,
		optional: opts,
		required: ocidRequirement{compartmentID},
	}

	created := Time{Time: time.Now()}
	expected := []Volume{
		{
			ID:            "id1",
			CompartmentID: compartmentID,
			DisplayName:   "vol1",
			TimeCreated:   created,
		},
		{
			ID:            "id2",
			CompartmentID: compartmentID,
			DisplayName:   "vol2",
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

	actual, e := s.requestor.ListVolumes(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.Volumes))
}
