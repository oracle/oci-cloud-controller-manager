// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestCreateImage() {
	res := &Image{
		BaseImageID:            "base_image_id",
		CompartmentID:          "compartmentID",
		CreateImageAllowed:     true,
		DisplayName:            "displayName",
		ID:                     "id1",
		State:                  ResourceProvisioning,
		OperatingSystem:        "operatingSystem",
		OperatingSystemVersion: "operatingSystemVersion",
		TimeCreated:            Time{Time: time.Now()},
	}

	required := struct {
		ocidRequirement
		InstanceID string `header:"-" json:"instanceId" url:"-"`
	}{
		InstanceID: "instance_id",
	}
	required.CompartmentID = res.CompartmentID

	opts := &CreateOptions{}
	opts.DisplayName = res.DisplayName

	details := &requestDetails{
		name:     resourceImages,
		required: required,
		optional: opts,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateImage(res.CompartmentID, "instance_id", opts)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetImage() {
	res := &Image{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		name: resourceImages,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetImage(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestUpdateImage() {
	res := &Image{
		ID:          "id",
		DisplayName: "displayName",
		TimeCreated: Time{Time: time.Now()},
	}

	opts := &UpdateOptions{}
	opts.DisplayName = res.DisplayName

	details := &requestDetails{
		name:     resourceImages,
		ids:      urlParts{res.ID},
		optional: opts,
	}

	respHeaders := http.Header{}
	respHeaders.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: respHeaders,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateImage(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.DisplayName, actual.DisplayName)
	s.Equal("ETAG!", actual.ETag)
}

func (s *CoreTestSuite) TestDeleteImage() {
	s.testDeleteResource(resourceImages, "id", s.requestor.DeleteImage)
}

func (s *CoreTestSuite) TestListImages() {
	compartmentID := "compartmentid"
	opts := &ListImagesOptions{}
	opts.Limit = 100
	opts.Page = "pageid"
	opts.OperatingSystem = "operating_system"
	opts.OperatingSystemVersion = "operating_system_version"

	details := &requestDetails{
		name:     resourceImages,
		optional: opts,
		required: listOCIDRequirement{compartmentID},
	}

	created := Time{Time: time.Now()}
	expected := []Image{
		{
			ID:            "id1",
			CompartmentID: compartmentID,
			DisplayName:   "res1",
			TimeCreated:   created,
		},
		{
			ID:            "id2",
			CompartmentID: compartmentID,
			DisplayName:   "res2",
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

	actual, e := s.requestor.ListImages(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.Images))
}
