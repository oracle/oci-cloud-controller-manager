// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,identity_compartment recording,all !recording

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
)

type IdentityCompartmentTestSuite struct {
	suite.Suite
}

func (s *IdentityCompartmentTestSuite) TestCompartment() {
	client := getClient("fixtures/identity/compartment")
	defer client.Stop()
	id, err := resourceApply(createCompartment(client))
	s.Require().NoError(err)

	compartment, err := client.GetCompartment(id)
	s.Require().NoError(err)
	s.Require().NotNil(compartment)
	updateCompartment := bm.UpdateIdentityOptions{
		Description: "new desc",
	}
	compartment, err = client.UpdateCompartment(id, &updateCompartment)
	s.NoError(err)

	s.Equal("new desc", compartment.Description)
}

func (s *IdentityCompartmentTestSuite) TestListCompartment() {
	client := getClient("fixtures/identity/listcompartments")
	defer client.Stop()
	var options bm.ListOptions
	calls := 0
	compartments := 0
	options.Limit = 4
	for {
		list, err := client.ListCompartments(&options)
		s.Require().NoError(err)
		s.Require().NotNil(list)
		calls++
		compartments += len(list.Compartments)
		if list.NextPage == "" {
			break
		}
		options.Page = list.NextPage
	}
	s.Equal(7, calls)
	s.Equal(23, compartments)
}

func TestIdentityCompartmentTestSuite(t *testing.T) {
	suite.Run(t, new(IdentityCompartmentTestSuite))
}
