// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"os/exec"
	"time"
)

func getTestIPSecConnection() *IPSecConnection {
	buff, _ := exec.Command("uuidgen").Output()
	conn := &IPSecConnection{
		CompartmentID: "compartmentID",
		CpeID:         "cpeid",
		DisplayName:   "displayName",
		DrgID:         "drgid",
		ID:            string(buff),
		State:         ResourceUp,
		StaticRoutes:  []string{"route1", "route2", "route3"},
		TimeCreated:   Time{Time: time.Now()},
	}
	conn.ETag = "etag"
	conn.RequestID = "requestid"

	return conn
}

func (s *CoreTestSuite) TestCreateIPSecConnection() {
	res := getTestIPSecConnection()

	opts := &CreateOptions{}
	opts.DisplayName = res.DisplayName

	required := struct {
		ocidRequirement
		CpeID        string   `header:"-" json:"cpeId" url:"-"`
		DrgID        string   `header:"-" json:"drgId" url:"-"`
		StaticRoutes []string `header:"-" json:"staticRoutes" url:"-"`
	}{
		CpeID:        res.CpeID,
		DrgID:        res.DrgID,
		StaticRoutes: res.StaticRoutes,
	}
	required.CompartmentID = res.CompartmentID

	details := &requestDetails{
		name:     resourceIPSecConnections,
		optional: opts,
		required: required,
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(res),
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateIPSecConnection(
		res.CompartmentID,
		res.CpeID,
		res.DrgID,
		res.StaticRoutes,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(res.CompartmentID, actual.CompartmentID)
}

func (s *CoreTestSuite) TestGetIPSecConnection() {
	res := getTestIPSecConnection()

	details := &requestDetails{
		name: resourceIPSecConnections,
		ids:  urlParts{res.ID},
	}

	resp := &response{
		body: marshalObjectForTest(res),
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetIPSecConnection(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
}

func (s *CoreTestSuite) TestDeleteIPSecConnection() {
	s.testDeleteResource(resourceIPSecConnections, "id", s.requestor.DeleteIPSecConnection)
}

func (s *CoreTestSuite) TestListIPSecConnections() {
	opts := &ListIPSecConnsOptions{}
	opts.Limit = 100
	opts.Page = "pageid"

	details := &requestDetails{
		name:     resourceIPSecConnections,
		optional: opts,
		required: listOCIDRequirement{CompartmentID: "compartmentid"},
	}

	expected := []IPSecConnection{*getTestIPSecConnection()}

	headers := http.Header{}
	headers.Set(headerOPCNextPage, "nextpage")
	headers.Set(headerOPCRequestID, "requestid")

	s.requestor.On("getRequest", details).Return(
		&response{
			header: headers,
			body:   marshalObjectForTest(expected),
		},
		nil,
	)

	actual, e := s.requestor.ListIPSecConnections("compartmentid", opts)
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected), len(actual.Connections))
}

func (s *CoreTestSuite) TestGetIPConnectionDeviceConfig() {
	res := &IPSecConnectionDeviceConfig{
		OPCRequestIDUnmarshaller: OPCRequestIDUnmarshaller{
			RequestID: "requestid",
		},
		IPSecConnectionDevice: IPSecConnectionDevice{
			CompartmentID: "compartmentId",
			ID:            "id",
			TimeCreated:   Time{Time: time.Now()},
		},
		Tunnels: []TunnelConfig{
			{
				IPAddress:    "10.10.10.2",
				SharedSecret: "bobdobbs",
				TimeCreated:  Time{Time: time.Now()},
			},
			{
				IPAddress:    "10.10.10.3",
				SharedSecret: "fsm",
				TimeCreated:  Time{Time: time.Now()},
			},
		},
	}

	details := &requestDetails{
		name: resourceIPSecConnections,
		ids:  urlParts{"id", deviceConfig},
	}

	resp := &response{body: marshalObjectForTest(res)}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetIPSecConnectionDeviceConfig("id")
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal(len(res.Tunnels), len(actual.Tunnels))

}

func (s *CoreTestSuite) TestGetIPConnectionDeviceStatus() {
	res := &IPSecConnectionDeviceStatus{
		OPCRequestIDUnmarshaller: OPCRequestIDUnmarshaller{
			RequestID: "requestid",
		},
		IPSecConnectionDevice: IPSecConnectionDevice{
			CompartmentID: "compartmentId",
			ID:            "id",
			TimeCreated:   Time{Time: time.Now()},
		},
		Tunnels: []TunnelStatus{
			{
				IPAddress:         "10.10.10.2",
				State:             ResourceUp,
				TimeCreated:       Time{Time: time.Now()},
				TimeStateModified: Time{Time: time.Now()},
			},
			{
				IPAddress:         "10.10.10.3",
				State:             ResourceDown,
				TimeCreated:       Time{Time: time.Now()},
				TimeStateModified: Time{Time: time.Now()},
			},
		},
	}

	details := &requestDetails{
		name: resourceIPSecConnections,
		ids:  urlParts{"id", deviceStatus},
	}

	resp := &response{body: marshalObjectForTest(res)}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetIPSecConnectionDeviceStatus("id")
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal(len(res.Tunnels), len(actual.Tunnels))
}
