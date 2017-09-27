// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *CoreTestSuite) TestListCpes() {
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "pageid"

	compartmentID := "compartmentid"
	created := Time{Time: time.Now()}

	details := &requestDetails{
		name:     resourceCustomerPremiseEquipment,
		optional: opts,
		required: listOCIDRequirement{CompartmentID: compartmentID},
	}

	expected := []Cpe{
		{
			ID:            "id1",
			CompartmentID: compartmentID,
			DisplayName:   "cpe1",
			IPAddress:     "120.121.122.123",
			TimeCreated:   created,
		},
		{
			ID:            "id2",
			CompartmentID: compartmentID,
			DisplayName:   "cpe1",
			IPAddress:     "120.121.122.124",
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

	actual, e := s.requestor.ListCpes(compartmentID, opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.Cpes))

}

func (s *CoreTestSuite) TestCreateCpe() {
	created := Time{Time: time.Now()}

	compartmentID := "compartmentid"
	displayName := "cpe1"
	ip := "123.123.123.123"

	cpe := &Cpe{
		ID:            "id1",
		CompartmentID: compartmentID,
		DisplayName:   displayName,
		IPAddress:     ip,
		TimeCreated:   created,
	}

	opts := &CreateOptions{}
	opts.DisplayName = displayName

	required := struct {
		ocidRequirement
		IPAddress string `header:"-" json:"ipAddress" url:"-"`
	}{
		IPAddress: "123.123.123.123",
	}
	required.CompartmentID = compartmentID

	details := &requestDetails{
		name:     resourceCustomerPremiseEquipment,
		optional: opts,
		required: required,
	}

	s.requestor.On("postRequest", details).Return(
		&response{
			header: http.Header{},
			body:   marshalObjectForTest(cpe),
		},
		nil,
	)

	actual, err := s.requestor.CreateCpe(compartmentID, ip, opts)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(cpe.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestDeleteCpe() {
	s.testDeleteResource(resourceCustomerPremiseEquipment, "cpeid", s.requestor.DeleteCpe)
}

func (s *CoreTestSuite) TestGetCpe() {
	id := "cpeid"

	details := &requestDetails{
		name: resourceCustomerPremiseEquipment,
		ids:  urlParts{id},
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG")

	created := Time{Time: time.Now()}

	expected := &Cpe{
		ID:          id,
		TimeCreated: created,
	}
	resp := &response{
		body:   marshalObjectForTest(expected),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetCpe(id)
	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(expected.ID, actual.ID)
	s.Equal("ETAG", actual.ETag)
}
