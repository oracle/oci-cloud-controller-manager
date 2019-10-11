// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

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
			State:              ResourceAttached,
			SubnetID:           "subnetid",
			VnicID:             opts.VnicID,
		},
		{
			AvailabilityDomain: opts.AvailabilityDomain,
			CompartmentID:      compartmentID,
			DisplayName:        "name2",
			ID:                 "id2",
			State:              ResourceAttached,
			SubnetID:           "subnetid",
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
