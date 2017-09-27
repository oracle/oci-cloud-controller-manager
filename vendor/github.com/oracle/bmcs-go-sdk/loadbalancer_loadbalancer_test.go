// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"
	"time"
)

func (s *LoadbalancerTestSuite) TestCreateLoadbalancer() {
	res := &LoadBalancer{
		CompartmentID: "compartmentID",
		DisplayName:   "displayName",
		ID:            "id1",
		IPAddresses:   []IPAddress{IPAddress{"123"}},
		IsPrivate:     false,
		Shape:         "100Mbps",
		State:         ResourceProvisioning,
		SubnetIDs:     []string{"subnetId1"},
		TimeCreated:   Time{Time: time.Now()},
	}

	opts := &CreateLoadBalancerOptions{}
	opts.DisplayName = res.DisplayName
	opts.IsPrivate = false

	required := CreateLoadBalancerDetails{
		Shape:     res.Shape,
		SubnetIDs: res.SubnetIDs,
	}
	required.CompartmentID = res.CompartmentID

	details := &requestDetails{
		name:     resourceLoadBalancers,
		optional: opts,
		required: required,
	}

	workReqId := "ocid1.loadbalancerworkrequest.oc1.phx.aaaaaaaa"
	header := http.Header{}
	header.Set(headerOPCWorkRequestID, workReqId)
	resp := &response{
		header: header,
	}
	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, err := s.requestor.CreateLoadBalancer(
		nil,
		nil,
		res.CompartmentID,
		nil,
		res.Shape,
		res.SubnetIDs,
		opts,
	)

	s.Nil(err)
	s.NotNil(actual)
	s.Equal(workReqId, actual)
}
