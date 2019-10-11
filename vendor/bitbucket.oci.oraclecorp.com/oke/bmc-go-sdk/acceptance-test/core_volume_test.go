// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core recording,core_volume recording,all !recording

package acceptance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
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
	client := getClient("fixtures/core/setup")
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

	// Get Availability Domains
	ads, err := client.ListAvailabilityDomains(s.compartmentID)
	s.Require().NoError(err)
	s.availabilityDomains = ads.AvailabilityDomains
}

func (s *VolumeTestSuite) TestCreateVolume() {
	client := getClient("fixtures/core/volume")
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
			time.Sleep(2 * time.Second) // wait until provisioning is complete
		}
		err := client.DeleteVolume(v.ID, nil)
		s.NoError(err)
	}()
	s.Require().NotEmpty(v.ID)
}
