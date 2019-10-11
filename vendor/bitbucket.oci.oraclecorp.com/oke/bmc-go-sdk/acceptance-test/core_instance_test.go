// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core recording,core_instance recording,all !recording

package acceptance

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
	//"fmt"
	"time"
)

type InstanceTestSuite struct {
	compartmentID       string
	availabilityDomains []bm.AvailabilityDomain
	images              []bm.Image
	shapes              []bm.Shape
	suite.Suite
}

func TestInstanceTestSuite(t *testing.T) {
	suite.Run(t, new(InstanceTestSuite))
}

func (s *InstanceTestSuite) SetupSuite() {
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
	s.availabilityDomains = ads.AvailabilityDomains
	s.Require().NoError(err)

	// Get Instance Shapes

	shapes, err := client.ListShapes(s.compartmentID, nil)
	s.Require().NoError(err)
	s.shapes = shapes.Shapes

	// Get Image types

	images, err := client.ListImages(s.compartmentID, nil)
	s.Require().NoError(err)
	s.images = images.Images
}

func (s *InstanceTestSuite) TestInstanceCreate() {
	client := getClient("fixtures/core/instance")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	subnetID, err := resourceApply(createSubnet(client, s.compartmentID, s.availabilityDomains[0].Name, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteSubnet(client, subnetID))
		s.NoError(err)
		time.Sleep(2 * time.Second)
	}()
	s.Require().NotEmpty(subnetID)

	id, err := resourceApply(createInstance(client, s.compartmentID, s.availabilityDomains[0].Name, s.images[0].ID, s.shapes[0].Name, subnetID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteInstance(client, id))
		s.NoError(err)
	}()
	s.Require().NotEmpty(id)
}

func (s *InstanceTestSuite) TestInstanceList() {
	client := getClient("fixtures/core/instance_list")
	defer client.Stop()

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()
	s.Require().NotEmpty(vcnID)

	subnetID, err := resourceApply(createSubnet(client, s.compartmentID, s.availabilityDomains[0].Name, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteSubnet(client, subnetID))
		s.NoError(err)
		time.Sleep(2 * time.Second)
	}()
	s.Require().NotEmpty(subnetID)

	id, err := resourceApply(createInstance(client, s.compartmentID, s.availabilityDomains[0].Name, s.images[0].ID, s.shapes[0].Name, subnetID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteInstance(client, id))
		s.NoError(err)
	}()
	s.Require().NotEmpty(id)

	instances, err := client.ListInstances(s.compartmentID, nil)
	s.Require().NoError(err)
	found := false
	for _, instance := range instances.Instances {
		if strings.Compare(id, instance.ID) == 0 {
			found = true
		}
	}
	s.Require().True(found, "Created Instance not found in list")
}
