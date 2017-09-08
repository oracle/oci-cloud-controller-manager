// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core_volume recording,all !recording

package acceptance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

type VolumeTestSuite struct {
	compartmentID       string
	availabilityDomains []bm.AvailabilityDomain
	suite.Suite
}

func TestVolumeTestSuite(t *testing.T) {
	suite.Run(t, new(VolumeTestSuite))
}

func (s *VolumeTestSuite) SetupSuite() {
	client := helpers.GetClient("fixtures/core/setup")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	s.Require().NoError(err)
	s.compartmentID = compartmentID

	// Get Availability Domains
	ads, err := client.ListAvailabilityDomains(s.compartmentID)
	s.Require().NoError(err)
	s.availabilityDomains = ads.AvailabilityDomains
}

func (s *VolumeTestSuite) TestCreateVolume() {
	client := helpers.GetClient("fixtures/core/volume")
	defer client.Stop()

	v, err := client.CreateVolume(
		s.availabilityDomains[0].Name,
		s.compartmentID,
		nil,
	)

	s.Require().NoError(err)
	defer func() {
		for {
			vs, _ := client.GetVolume(v.ID)
			if vs.State != "PROVISIONING" {
				break
			}
			helpers.Sleep(2 * time.Second) // wait until provisioning is complete
		}
		err := client.DeleteVolume(v.ID, nil)
		s.NoError(err)
	}()
	s.Require().NotEmpty(v.ID)
}
