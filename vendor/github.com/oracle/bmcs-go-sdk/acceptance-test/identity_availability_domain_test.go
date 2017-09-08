// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,identity_availability_domain recording,all !recording

package acceptance

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

type IdentityAvailabilityDomainTestSuite struct {
	suite.Suite
	compartmentID string
}

func (s *IdentityAvailabilityDomainTestSuite) SetupSuite() {
	client := helpers.GetClient("fixtures/listad/setup")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	s.Require().NoError(err)
	s.compartmentID = compartmentID
}

func (s *IdentityAvailabilityDomainTestSuite) TestListAvailabilityDomain() {
	client := helpers.GetClient("fixtures/identity/listads")
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
