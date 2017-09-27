// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestAttachVolume() {
	res := &VolumeAttachment{
		AttachmentType:     "attachmentType",
		AvailabilityDomain: "availabilityDomain",
		CompartmentID:      "compartmentID",
		DisplayName:        "displayName",
		ID:                 "id1",
		InstanceID:         "instance",
		State:              ResourceProvisioning,
		TimeCreated:        Time{Time: time.Now()},
		VolumeID:           "volumeID",
	}

	opts := &CreateOptions{}
	opts.DisplayName = res.DisplayName

	required := struct {
		AttachmentType string `header:"-" json:"type" url:"-"`
		InstanceID     string `header:"-" json:"instanceId" url:"-"`
		VolumeID       string `header:"-" json:"volumeId" url:"-"`
	}{
		AttachmentType: res.AttachmentType,
		InstanceID:     res.InstanceID,
		VolumeID:       res.VolumeID,
	}

	details := &requestDetails{
		name:     resourceVolumeAttachments,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.AttachVolume(
		res.AttachmentType,
		res.InstanceID,
		res.VolumeID,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetVolumeAttachment() {
	res := &VolumeAttachment{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourceVolumeAttachments,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetVolumeAttachment(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestDetachVolumet() {
	s.testDeleteResource(resourceVolumeAttachments, "id", s.requestor.DetachVolume)
}

func (s *CoreTestSuite) TestListVolumeAttachments() {
	compartmentID := "compartmentid"

	opts := &ListVolumeAttachmentsOptions{}
	opts.AvailabilityDomain = "availability_domain"
	opts.Limit = 100
	opts.Page = "page"
	opts.InstanceID = "instance_id"
	opts.VolumeID = "volume_id"

	details := &requestDetails{
		name:     resourceVolumeAttachments,
		optional: opts,
		required: ocidRequirement{compartmentID},
	}

	created := Time{Time: time.Now()}
	expected := []VolumeAttachment{
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

	actual, e := s.requestor.ListVolumeAttachments(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.VolumeAttachments))
}
