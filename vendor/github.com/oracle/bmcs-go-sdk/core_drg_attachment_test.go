// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestCreateDrgAttachment() {
	res := &DrgAttachment{
		CompartmentID: "compartmentID",
		DisplayName:   "displayName",
		DrgID:         "drgID",
		ID:            "id1",
		State:         ResourceAttached,
		TimeCreated:   Time{Time: time.Now()},
		VcnID:         "vcnID",
	}

	opts := &CreateOptions{}
	opts.DisplayName = res.DisplayName

	required := struct {
		DrgID string `header:"-" json:"drgId" url:"-"`
		VcnID string `header:"-" json:"vcnId" url:"-"`
	}{
		DrgID: res.DrgID,
		VcnID: res.VcnID,
	}

	details := &requestDetails{
		name:     resourceDrgAttachments,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateDrgAttachment(res.DrgID, res.VcnID, opts)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetDrgAttachment() {
	res := &DrgAttachment{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		name: resourceDrgAttachments,
		ids:  urlParts{res.ID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")

	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetDrgAttachment(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestDeleteDrgAttachment() {
	s.testDeleteResource(resourceDrgAttachments, "id", s.requestor.DeleteDrgAttachment)
}

func (s *CoreTestSuite) TestListDrgAttachments() {
	compartmentID := "compartment_id"
	opts := &ListDrgAttachmentsOptions{}
	opts.DrgID = "drg_id"
	opts.Limit = 1
	opts.Page = "page"
	opts.VcnID = "vcn_id"

	details := &requestDetails{
		name:     resourceDrgAttachments,
		optional: opts,
		required: listOCIDRequirement{compartmentID},
	}

	created := Time{Time: time.Now()}
	expected := []DrgAttachment{
		{
			ID:            "id1",
			CompartmentID: compartmentID,
			DisplayName:   "drg1",
			TimeCreated:   created,
		},
		{
			ID:            "id2",
			CompartmentID: compartmentID,
			DisplayName:   "drg2",
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

	actual, e := s.requestor.ListDrgAttachments(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.DrgAttachments))
}
