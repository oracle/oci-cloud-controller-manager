// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import "net/http"

func (s *IdentityTestSuite) TestCreatePolicy() {
	compartmentID := s.requestor.authInfo.tenancyOCID

	res := Policy{
		CompartmentID: compartmentID,
		Description:   "description",
		Name:          "name",
		Statements: []string{
			"statement1",
			"statement2",
		},
	}

	opts := &CreatePolicyOptions{}
	opts.RetryToken = "xxxxx"

	required := struct {
		identityCreationRequirement
		Statements []string `header:"-" json:"statements" url:"-"`
	}{
		Statements: res.Statements,
	}
	required.CompartmentID = res.CompartmentID
	required.Description = res.Description
	required.Name = res.Name

	details := &requestDetails{
		name:     resourcePolicies,
		optional: opts,
		required: required,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("postRequest", details).Return(resp, nil)

	actual, e := s.requestor.CreatePolicy(res.Name, res.Description, compartmentID, res.Statements, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestGetPolicy() {
	res := Policy{ID: "id"}

	details := &requestDetails{
		ids:  urlParts{res.ID},
		name: resourcePolicies,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.GetPolicy(res.ID)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.ID, actual.ID)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestUpdatePolicy() {
	res := Policy{
		ID:          "id",
		Description: "desc",
	}

	opts := &UpdatePolicyOptions{}
	opts.Description = "desc"

	details := &requestDetails{
		ids:      urlParts{res.ID},
		name:     resourcePolicies,
		optional: opts,
	}

	headers := http.Header{}
	headers.Set(headerETag, "ETAG!")
	resp := &response{
		body:   marshalObjectForTest(res),
		header: headers,
	}

	s.requestor.On("request", http.MethodPut, details).Return(resp, nil)

	actual, e := s.requestor.UpdatePolicy(res.ID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(res.Description, actual.Description)
	s.Equal("ETAG!", actual.ETag)
}

func (s *IdentityTestSuite) TestDeletePolicy() {
	s.testDeleteResource(resourcePolicies, "9999", s.requestor.DeletePolicy)
}

func (s *IdentityTestSuite) TestListPolicies() {
	opts := &ListOptions{}
	opts.Limit = 100
	opts.Page = "wxyz"

	details := &requestDetails{
		name:     resourcePolicies,
		optional: opts,
		required: listOCIDRequirement{s.requestor.authInfo.tenancyOCID},
	}

	expected := ListPolicies{
		Policies: []Policy{
			{
				ID:   "1",
				Name: "one",
			},
			{
				ID:   "2",
				Name: "two",
			},
		},
	}

	resp := &response{
		body: marshalObjectForTest(expected.Policies),
	}
	s.requestor.On("getRequest", details).Return(resp, nil)

	actual, e := s.requestor.ListPolicies(s.requestor.authInfo.tenancyOCID, opts)

	s.requestor.AssertExpectations(s.T())
	s.Nil(e)
	s.NotNil(actual)
	s.Equal(len(expected.Policies), len(actual.Policies))
}
