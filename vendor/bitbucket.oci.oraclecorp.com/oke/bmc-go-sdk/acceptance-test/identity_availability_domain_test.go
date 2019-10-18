// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,identity_availability_domain recording,all !recording

package acceptance

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
)

type IdentityAvailabilityDomainTestSuite struct {
	suite.Suite
	compartmentID string
}

func (s *IdentityAvailabilityDomainTestSuite) SetupSuite() {
	client := getClient("fixtures/listad/setup")
	defer client.Stop()
	// get a compartment, any compartment
	var listOpts bm.ListOptions
	listOpts.Limit = 1
	list, err := client.ListCompartments(&listOpts)
	s.Require().NoError(err)
	if len(list.Compartments) == 1 {
		s.compartmentID = list.Compartments[0].ID
	} else {
		id, err := resourceApply(createCompartment(client))
		s.Require().NoError(err)
		s.compartmentID = id
	}
}

func (s *IdentityAvailabilityDomainTestSuite) TestListAvailabilityDomain() {
	client := getClient("fixtures/identity/listads")
	defer client.Stop()
	list, err := client.ListAvailabilityDomains(s.compartmentID)
	s.Require().NoError(err)
	s.Require().NotNil(list)
	s.Equal(3, len(list.AvailabilityDomains))
	fmt.Printf("%v\n", list.AvailabilityDomains)
}

func TestIdentityAvailabilityDomainTestSuite(t *testing.T) {
	suite.Run(t, new(IdentityAvailabilityDomainTestSuite))
}
