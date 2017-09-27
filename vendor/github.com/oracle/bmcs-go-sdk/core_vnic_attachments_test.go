// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestListVnicAttachments() {
	compartmentID := "compartmentid"
	opts := &ListVnicAttachmentsOptions{}
	opts.AvailabilityDomain = "domainid"
	opts.VnicID = "vnicid"
	opts.InstanceID = "instanceid"
	opts.Limit = 100
	opts.Page = "pageid"

	details := &requestDetails{
		name:     resourceVnicAttachments,
		optional: opts,
		required: ocidRequirement{compartmentID},
	}

	responseHeaders := http.Header{}
	responseHeaders.Set(headerOPCNextPage, "nextpage")
	responseHeaders.Set(headerOPCRequestID, "requestid")
	expected := []VnicAttachment{
		{
			AvailabilityDomain: opts.AvailabilityDomain,
			CompartmentID:      compartmentID,
			DisplayName:        "name1",
			ID:                 "id1",
			InstanceID:         "my_instance_id",
			State:              ResourceAttached,
			SubnetID:           "subnetid",
			TimeCreated:        time.Now(),
			VlanTag:            0,
			VnicID:             opts.VnicID,
		},
		{
			AvailabilityDomain: opts.AvailabilityDomain,
			CompartmentID:      compartmentID,
			DisplayName:        "name2",
			ID:                 "id2",
			InstanceID:         "my_instance_id",
			State:              ResourceAttached,
			SubnetID:           "subnetid",
			TimeCreated:        time.Now(),
			VlanTag:            1,
			VnicID:             opts.VnicID,
		},
	}

	s.requestor.On("getRequest", details).Return(
		&response{
			header: responseHeaders,
			body:   marshalObjectForTest(expected),
		},
		nil,
	)

	actual, e := s.requestor.ListVnicAttachments(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.Attachments))
	s.Equal("nextpage", actual.NextPage)
	s.Equal("requestid", actual.RequestID)
	s.Equal("name1", actual.Attachments[0].DisplayName)
	s.Equal("name2", actual.Attachments[1].DisplayName)
}

func (s *CoreTestSuite) TestAttachVnic() {
	instanceId := "my_instance_id"

	res := &VnicAttachment{
		AvailabilityDomain: "my_ad",
		CompartmentID:      "my_compartment",
		DisplayName:        "name1",
		ID:                 "id1",
		InstanceID:         instanceId,
		State:              ResourceAttached,
		SubnetID:           "subnetid",
		TimeCreated:        time.Now(),
		VlanTag:            0,
		VnicID:             "my_vnic",
	}

	vnicOpts := &CreateVnicOptions{}
	assignPublicIp := true
	skipSourceDestCheck := false
	vnicOpts.AssignPublicIp = &assignPublicIp
	vnicOpts.SkipSourceDestCheck = &skipSourceDestCheck
	attachmentOpts := &AttachVnicOptions{}

	required := struct {
		InstanceId        string             `header:"-" json:"instanceId" url:"-"`
		CreateVnicDetails *CreateVnicOptions `header:"-" json:"createVnicDetails" url:"-"`
	}{
		InstanceId:        instanceId,
		CreateVnicDetails: vnicOpts,
	}

	details := &requestDetails{
		name:     resourceVnicAttachments,
		optional: attachmentOpts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.AttachVnic(
		instanceId,
		vnicOpts,
		attachmentOpts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
	s.Equal(instanceId, actual.InstanceID)
	s.Equal(ResourceAttached, actual.State)
}

func (s *CoreTestSuite) TestGetVnicAttachment() {
	instanceId := "my_instance_id"

	res := &VnicAttachment{
		AvailabilityDomain: "my_ad",
		CompartmentID:      "my_compartment",
		DisplayName:        "name1",
		ID:                 "id1",
		InstanceID:         instanceId,
		State:              ResourceAttached,
		SubnetID:           "subnetid",
		TimeCreated:        time.Now(),
		VlanTag:            0,
		VnicID:             "my_vnic",
	}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourceVnicAttachments,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, err := s.requestor.GetVnicAttachment(res.ID)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
	s.Equal(instanceId, actual.InstanceID)
	s.Equal(ResourceAttached, actual.State)
}

func (s *CoreTestSuite) TestDetachVnic() {
	s.testDeleteResource(resourceVnicAttachments, "id", s.requestor.DetachVnic)
}
