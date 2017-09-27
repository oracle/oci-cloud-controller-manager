// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"errors"
	"net/http"
	"time"

	"github.com/stretchr/testify/mock"
)

func (s *ObjectStorageTestSuite) TestDeletePreauthenticatedRequest() {
	namesp := Namespace("namesp")
	opts := &ClientRequestOptions{}
	details := &requestDetails{
		ids:      buildPARUrlParts(namesp, "bucket1", "par1"),
		optional: opts,
	}
	s.requestor.On("deleteRequest", details).Return(nil)

	e := s.requestor.DeletePreauthenticatedRequest(namesp, "bucket1", "par1", opts)
	s.Nil(e)
}

func (s *ObjectStorageTestSuite) TestDeletePreauthenticatedRequestError() {
	namesp := Namespace("namesp")
	opts := &ClientRequestOptions{}
	details := &requestDetails{
		ids:      buildPARUrlParts(namesp, "bucket1", "par1"),
		optional: opts,
	}
	fe := errors.New("fake error")
	s.requestor.On("deleteRequest", details).Return(fe)

	e := s.requestor.DeletePreauthenticatedRequest(namesp, "bucket1", "par1", opts)
	s.NotNil(e)
}

func (s *ObjectStorageTestSuite) TestCreatePreauthenticatedFailure() {

	namesp := Namespace("namesp")
	opts := &CreatePreauthenticatedRequestDetails{}
	details := &requestDetails{
		ids:      buildPARUrlParts(namesp, "bucket1"),
		name:     resourcePAR,
		required: opts,
	}

	e := errors.New("fake error")
	s.requestor.On("postRequest", details).Return(&response{}, e)

	actual, err := s.requestor.CreatePreauthenticatedRequest(
		namesp,
		"bucket1",
		opts,
	)

	s.NotNil(err)
	s.Nil(actual)
}

func (s *ObjectStorageTestSuite) TestCreatePreauthenticatedRequest() {
	var timecreated time.Time = time.Now()
	timeCreated := Time{Time: timecreated}
	timeExpires := Time{Time: timecreated.AddDate(0, 0, 1)}
	namesp := Namespace("namesp")

	par := &PreauthenticatedRequest{
		ID:          "string",
		Name:        "name",
		AccessURI:   "/some/uri",
		ObjectName:  "objectname",
		TimeCreated: timeCreated,
		AccessType:  PARObjectRead,
		TimeExpires: timeExpires,
	}

	opts := &CreatePreauthenticatedRequestDetails{
		Name:        "name",
		ObjectName:  "objectname",
		AccessType:  PARObjectRead,
		TimeExpires: timeExpires,
	}

	details := &requestDetails{
		ids:      buildPARUrlParts(namesp, "bucket1"),
		name:     resourcePAR,
		required: opts,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(par),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreatePreauthenticatedRequest(
		namesp,
		"bucket1",
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(par.Name, actual.Name)
	s.Equal(par.AccessURI, actual.AccessURI)
	s.Equal(par.AccessType, actual.AccessType)
	//s.Equal(par.TimeExpires.Unix(), actual.TimeExpires.Unix())
}
func (s *ObjectStorageTestSuite) TestGetPreauthenticatedRequest() {
	var timecreated time.Time = time.Now()
	timeCreated := Time{Time: timecreated}
	timeExpires := Time{Time: timecreated.AddDate(0, 0, 1)}
	namesp := Namespace("namesp")

	par := &PreauthenticatedRequestSummary{
		ID:          "string",
		Name:        "name",
		ObjectName:  "objectname",
		AccessType:  PARObjectRead,
		TimeExpires: timeExpires,
		TimeCreated: timeCreated,
	}

	opts := &ClientRequestOptions{}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(par),
	}
	s.requestor.On("getRequest",
		mock.MatchedBy(func(req *requestDetails) bool { return true })).Return(resp, nil)

	actual, err := s.requestor.GetPreauthenticatedRequest(
		namesp,
		"bucket1",
		"par1",
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(par.Name, actual.Name)
	s.Equal(par.AccessType, actual.AccessType)
	//s.Equal(par.TimeExpires, actual.TimeExpires)
}

func (s *ObjectStorageTestSuite) TestGetPreauthenticatedRequestError() {
	var timecreated time.Time = time.Now()
	timeCreated := Time{Time: timecreated}
	timeExpires := Time{Time: timecreated.AddDate(0, 0, 1)}
	namesp := Namespace("namesp")

	par := &PreauthenticatedRequestSummary{
		ID:          "string",
		Name:        "name",
		ObjectName:  "objectname",
		AccessType:  PARObjectRead,
		TimeExpires: timeExpires,
		TimeCreated: timeCreated,
	}

	opts := &ClientRequestOptions{}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(par),
	}
	e := errors.New("fake error")

	s.requestor.On("getRequest",
		mock.MatchedBy(func(req *requestDetails) bool { return true })).
		Return(resp, e)

	actual, err := s.requestor.GetPreauthenticatedRequest(
		namesp,
		"bucket1",
		"par1",
		opts,
	)

	s.NotNil(err)
	s.Nil(actual)
}
func (s *ObjectStorageTestSuite) TestListPreauthenticatedRequests() {
	var timecreated time.Time = time.Now()
	timeCreated := Time{Time: timecreated}
	timeExpires := Time{Time: timecreated.AddDate(0, 0, 1)}
	namesp := Namespace("namesp")

	parSum := &ListPreauthenticatedRequests{
		PreauthenticatedRequests: []PreauthenticatedRequestSummary{
			PreauthenticatedRequestSummary{
				ID:          "par1",
				Name:        "name",
				ObjectName:  "objectname2",
				AccessType:  PARObjectRead,
				TimeExpires: timeExpires,
				TimeCreated: timeCreated,
			},
			PreauthenticatedRequestSummary{
				ID:          "par2",
				Name:        "name2",
				ObjectName:  "objectname2",
				AccessType:  PARObjectRead,
				TimeExpires: timeExpires,
				TimeCreated: timeCreated,
			},
		},
	}

	opts := &ListPreauthenticatedRequestOptions{}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(parSum),
	}

	s.requestor.On("getRequest",
		mock.MatchedBy(func(req *requestDetails) bool { return true })).
		Return(resp, nil)

	actual, err := s.requestor.ListPreauthenticatedRequests(namesp, "bucket1", opts)
	s.NotNil(actual)
	s.Equal(2, len(actual.GetList()))
	s.Nil(err)
}

func (s *ObjectStorageTestSuite) TestListPreauthenticatedRequestsFailure() {
	namesp := Namespace("namesp")

	opts := &ListPreauthenticatedRequestOptions{}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(opts),
	}

	err := errors.New("fake error")
	s.requestor.On("getRequest",
		mock.MatchedBy(func(req *requestDetails) bool { return true })).
		Return(resp, err)

	actual, actual_err := s.requestor.ListPreauthenticatedRequests(namesp, "bucket1", opts)
	s.Nil(actual)
	s.NotNil(err)
	s.Equal(err, actual_err)
}
