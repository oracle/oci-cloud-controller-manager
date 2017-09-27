// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
)

func (s *CoreTestSuite) TestListInstanceConsoleHistories() {
	instanceID := "instanceid"

	opts := &ListConsoleHistoriesOptions{}
	opts.InstanceID = instanceID
	opts.AvailabilityDomain = "domainid"
	opts.Limit = 100
	opts.Page = "pageid"

	compartmentID := "compartmentid"

	reqOpts := &requestDetails{
		name:     resourceInstanceConsoleHistories,
		optional: opts,
		required: listOCIDRequirement{CompartmentID: compartmentID},
	}

	expected := []ConsoleHistoryMetadata{
		{
			AvailabilityDomain: "availabilityDomain",
			CompartmentID:      compartmentID,
			DisplayName:        "cpe1",
			ID:                 "id1",
			InstanceID:         instanceID,
			State:              ResourceRequested,
		},
		{
			AvailabilityDomain: "availabilityDomain",
			CompartmentID:      compartmentID,
			DisplayName:        "cpe1",
			ID:                 "id2",
			InstanceID:         instanceID,
			State:              ResourceRequested,
		},
	}

	responseHeaders := http.Header{}
	s.requestor.On("getRequest", reqOpts).Return(
		&response{
			header: responseHeaders,
			body:   marshalObjectForTest(expected),
		},
		nil,
	)

	actual, e := s.requestor.ListConsoleHistories(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.ConsoleHistories))
}

func (s *CoreTestSuite) TestCaptureConsoleHistory() {
	instanceID := "instanceid"

	consoleHistoryMetadata := ConsoleHistoryMetadata{
		AvailabilityDomain: "availabilityDomain",
		CompartmentID:      "compartmentid",
		DisplayName:        "cpe1",
		ID:                 "id1",
		InstanceID:         instanceID,
		State:              "REQUESTED",
	}

	required := struct {
		InstanceID string `header:"-" json:"instanceId" url:"-"`
	}{
		InstanceID: instanceID,
	}

	details := &requestDetails{
		name:     resourceInstanceConsoleHistories,
		optional: (*RetryTokenOptions)(nil),
		required: required,
	}

	s.requestor.On("postRequest", details).Return(
		&response{
			header: http.Header{},
			body:   marshalObjectForTest(consoleHistoryMetadata),
		},
		nil,
	)

	actual, err := s.requestor.CaptureConsoleHistory(instanceID, nil)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(consoleHistoryMetadata.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetConsoleHistory() {
	instanceID := "instanceid"

	details := &requestDetails{
		name: resourceInstanceConsoleHistories,
		ids:  urlParts{instanceID},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")

	expected := &ConsoleHistoryMetadata{
		AvailabilityDomain: "availabilityDomain",
		CompartmentID:      "compartmentid",
		DisplayName:        "cpe1",
		ID:                 "id1",
		InstanceID:         instanceID,
		State:              "REQUESTED",
	}

	resp := &response{
		body:   marshalObjectForTest(expected),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetConsoleHistory(instanceID)
	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(expected.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}

func (s *CoreTestSuite) TestShowConsoleHistoryData() {
	opts := &ConsoleHistoryDataOptions{
		Length: 1000,
		Offset: 1,
	}

	details := &requestDetails{
		name:     resourceInstanceConsoleHistories,
		ids:      urlParts{"id", dataURLPart},
		optional: opts,
	}

	h := http.Header{}
	h.Set(headerBytesRemaining, "1000")

	resp := &response{
		body:   marshalObjectForTest("wubalubadubdub"),
		header: h,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ShowConsoleHistoryData("id", opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal("\"wubalubadubdub\"\n", actual.Data)
	s.Equal(1000, actual.BytesRemaining)
}
