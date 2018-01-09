// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestCreateVolumeBackup() {
	res := &VolumeBackup{
		CompartmentID:       "compartmentID",
		DisplayName:         "displayName",
		ID:                  "id1",
		SizeInGBs:           50,
		State:               ResourceProvisioning,
		TimeCreated:         Time{Time: time.Now()},
		TimeRequestReceived: Time{Time: time.Now()},
		UniqueSizeInGBs:     49,
		VolumeID:            "volumd_id",
	}

	opts := &CreateOptions{}
	opts.DisplayName = res.DisplayName

	required := struct {
		VolumeID string `header:"-" json:"volumeId" url:"-"`
	}{
		VolumeID: res.VolumeID,
	}

	details := &requestDetails{
		name:     resourceVolumeBackups,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateVolumeBackup(res.VolumeID, opts)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetVolumeBackup() {
	res := &VolumeBackup{
		ID:          "id",
		TimeCreated: Time{Time: time.Now()},
	}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourceVolumeBackups,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetVolumeBackup(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestUpdateVolumeBackup() {
	res := &VolumeBackup{
		ID:          "id",
		DisplayName: "displayName",
		TimeCreated: Time{Time: time.Now()},
	}

	opts := &IfMatchDisplayNameOptions{}
	opts.DisplayName = res.DisplayName

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourceVolumeBackups,
		optional: opts,
	}

	respHeaders := http.Header{}
	respHeaders.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: respHeaders,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdateVolumeBackup(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.DisplayName, actual.DisplayName)
	s.Equal("ETAG!", actual.ETag)
}

func (s *CoreTestSuite) TestDeleteVolumeBackup() {
	s.testDeleteResource(resourceVolumeBackups, "id", s.requestor.DeleteVolumeBackup)
}

func (s *CoreTestSuite) TestListVolumeBackups() {
	compartmentID := "compartmentid"

	opts := &ListBackupsOptions{}
	opts.Limit = 100
	opts.Page = "pageid"
	opts.VolumeID = "volume_id"

	details := &requestDetails{
		name:     resourceVolumeBackups,
		optional: opts,
		required: ocidRequirement{compartmentID},
	}

	created := Time{Time: time.Now()}
	expected := []VolumeBackup{
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

	actual, e := s.requestor.ListVolumeBackups(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.VolumeBackups))
}
