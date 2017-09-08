// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core recording,all !recording

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

const vcnAddress = "172.16.0.0/16"

type CoreTestSuite struct {
	compartmentID       string
	availabilityDomains []bm.AvailabilityDomain
	suite.Suite
}

func TestCoreTestSuite(t *testing.T) {
	suite.Run(t, new(CoreTestSuite))
}

func (s *CoreTestSuite) SetupSuite() {
	client := helpers.GetClient("fixtures/core/setup")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	s.Require().NoError(err)
	s.compartmentID = compartmentID

	// Get Availability Domains

	ads, err := client.ListAvailabilityDomains(s.compartmentID)
	s.availabilityDomains = ads.AvailabilityDomains
	s.Require().NoError(err)
}
