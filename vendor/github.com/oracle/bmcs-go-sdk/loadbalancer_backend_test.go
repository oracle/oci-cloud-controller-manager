// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"net/http"

	"github.com/stretchr/testify/assert"
)

func (s *LoadbalancerTestSuite) TestListBackends() {
	backend := Backend{
		Backup:    false,
		Drain:     false,
		IPAddress: "123.10.5.12",
		Port:      1234,
		Offline:   false,
		Weight:    1,
	}

	backends := &ListBackends{
		Backends: []Backend{
			backend,
		},
	}

	details := &requestDetails{
		name: resourceLoadBalancers,
		ids: urlParts{"lbID",
			resourceBackendSets, "lbBackendSet", resourceBackends},
	}

	resp := &response{
		header: http.Header{},
		body:   marshalObjectForTest(backends.Backends),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, err := s.requestor.ListBackends(
		"lbID",
		"lbBackendSet",
	)

	s.Nil(err)
	s.NotNil(actual)
	s.True(assert.ObjectsAreEqual(backends.Backends[0], actual.Backends[0]))
}
