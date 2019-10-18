// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core recording,core_volume_attachment recording,all !recording

package acceptance

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	bm "github.com/MustWin/baremetal-sdk-go"
)

type VolumeAttachmentTestSuite struct {
	compartmentID       string
	availabilityDomains []bm.AvailabilityDomain
	images              []bm.Image
	shapes              []bm.Shape
	suite.Suite
}

func TestVolumeAttachmentTestSuite(t *testing.T) {
	suite.Run(t, new(VolumeAttachmentTestSuite))
}

func (s *VolumeAttachmentTestSuite) SetupSuite() {
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

func (s *VolumeAttachmentTestSuite) TestAttachVolume() {
	client := getClient("fixtures/core/volume_attachment")
	defer client.Stop()

	//// Create dependant resources ////

	vcnID, err := resourceApply(createVCN(client, "172.16.0.0/16", s.compartmentID))
	s.Require().NoError(err)
	defer func() {
		_, err = resourceApply(deleteVCN(client, vcnID))
		s.NoError(err)
	}()

	subnetID, err := resourceApply(createSubnet(client, s.compartmentID, s.availabilityDomains[0].Name, vcnID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteSubnet(client, subnetID))
		s.NoError(err)
	}()

	instanceID, err := resourceApply(createInstance(client, s.compartmentID, s.availabilityDomains[0].Name, s.images[0].ID, s.shapes[0].Name, subnetID))
	s.Require().NoError(err)
	defer func() {
		_, err := resourceApply(deleteInstance(client, instanceID))
		s.NoError(err)
	}()

	v, err := client.CreateVolume(
		s.availabilityDomains[0].Name,
		s.compartmentID,
		nil,
	)
	s.Require().NoError(err)
	// Attachment must wait until volume is available
	for {
		v, _ := client.GetVolume(v.ID)
		if v.State != bm.ResourceProvisioning {
			s.Require().Equal(bm.ResourceAvailable, v.State, fmt.Sprintf("Volume in invalid state: %#v", v))
			break
		}
		time.Sleep(2 * time.Second)
	}
	// no defered DeleteVolume here, must happen after DetachVolume
	//// End dependant resource creation ////
	opts := bm.CreateOptions{}
	opts.DisplayName = "Test Volume Attachment"

	va, err := client.AttachVolume("iscsi", instanceID, v.ID, &opts)

	s.Require().NoError(err)
	// wait until volume is attached for up-to-data values
	for {
		va, _ = client.GetVolumeAttachment(va.ID)
		if va.State != bm.ResourceAttaching {
			s.Require().Equal(bm.ResourceAttached, va.State, fmt.Sprintf("Volume Attachment in invalid state: %#v", va))
			break
		}
		time.Sleep(2 * time.Second)
	}
	defer func() {
		err := client.DetachVolume(va.ID, nil)
		s.NoError(err, fmt.Sprintf("VolumeAttachment: %#v", va))
		// DeleteVolume must wait until volume is detachched
		for {
			va, _ = client.GetVolumeAttachment(va.ID)
			if va.State != bm.ResourceDetaching {
				s.Require().Equal(bm.ResourceDetached, va.State, fmt.Sprintf("Volume Attachment in invalid state: %#v", va))
				break
			}
			time.Sleep(2 * time.Second)
		}
		err = client.DeleteVolume(v.ID, nil)
		s.NoError(err, fmt.Sprintf("Volume: %#v", v))
	}()

	// Precise checks
	s.Equal(va.CompartmentID, s.compartmentID)
	s.Equal(va.AvailabilityDomain, s.availabilityDomains[0].Name)
	s.Equal(va.InstanceID, instanceID)
	s.Equal(va.VolumeID, v.ID)
	s.Equal(va.AttachmentType, "iscsi")
	s.Equal(va.State, bm.ResourceAttached)
	// Non-empty checks
	s.NotEmpty(va.ID)
	s.NotEmpty(va.TimeCreated)
	s.NotEmpty(va.IPv4)
	s.NotEmpty(va.IQN)
	// Empty checks
	// TODO: create CHAP enabled test
	s.Empty(va.CHAPSecret)
	s.Empty(va.CHAPUsername)

	// Broken
	s.Equal(va.DisplayName, "Test Volume Attachment")
}
