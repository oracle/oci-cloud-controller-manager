// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "time"

func (s *CoreTestSuite) TestGetVnic() {
	vnicID := "vnicid"
	res := &Vnic{
		AvailabilityDomain: "availabilitydomain",
		CompartmentID:      "compartmentid",
		DisplayName:        "displayname",
		ID:                 vnicID,
		State:              ResourceAvailable,
		PrivateIPAddress:   "10.10.10.10",
		PublicIPAddress:    "54.55.56.57",
		SubnetID:           "subnetid",
		TimeCreated:        Time{Time: time.Now()},
	}

	details := &requestDetails{
		name: resourceVnics,
		ids:  urlParts{res.ID},
	}

	resp := &response{body: marshalObjectForTest(res)}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetVnic(vnicID)
	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal(res.PublicIPAddress, actual.PublicIPAddress)
	s.Equal(res.PrivateIPAddress, actual.PrivateIPAddress)
}
