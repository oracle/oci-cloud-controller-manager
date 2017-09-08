// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core_volume_attachment recording,all !recording

package acceptance

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bm "github.com/MustWin/baremetal-sdk-go"
	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

func TestAttachVolume(t *testing.T) {
	client := helpers.GetClient("fixtures/core/volume_attachment")
	defer client.Stop()

	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	// Get Availability Domains
	ads, err := client.ListAvailabilityDomains(compartmentID)
	availabilityDomains := ads.AvailabilityDomains
	require.NoError(t, err)
	//// Create dependant resources ////

	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err)
	}()

	subnetID, err := helpers.CreateSubnet(client, compartmentID, availabilityDomains[0].Name, vcnID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteSubnet(client, subnetID)
		assert.NoError(t, err)
	}()

	instanceID, err := helpers.CreateInstance(client, compartmentID, availabilityDomains[0].Name, helpers.FastestImageID, helpers.SmallestShapeName, subnetID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteInstance(client, instanceID)
		assert.NoError(t, err)
	}()

	v, err := client.CreateVolume(
		availabilityDomains[0].Name,
		compartmentID,
		nil,
	)
	require.NoError(t, err)
	// Attachment must wait until volume is available
	for {
		v, _ := client.GetVolume(v.ID)
		if v.State != bm.ResourceProvisioning {
			require.Equal(t, bm.ResourceAvailable, v.State, fmt.Sprintf("Volume in invalid state: %#v", v))
			break
		}
		helpers.Sleep(2 * time.Second)
	}
	// no defered DeleteVolume here, must happen after DetachVolume
	//// End dependant resource creation ////
	opts := bm.CreateOptions{}
	opts.DisplayName = "Test Volume Attachment"

	va, err := client.AttachVolume("iscsi", instanceID, v.ID, &opts)

	require.NoError(t, err)
	// wait until volume is attached for up-to-data values
	for {
		va, _ = client.GetVolumeAttachment(va.ID)
		if va.State != bm.ResourceAttaching {
			require.Equal(t, bm.ResourceAttached, va.State, fmt.Sprintf("Volume Attachment in invalid state: %#v", va))
			break
		}
		helpers.Sleep(2 * time.Second)
	}
	defer func() {
		err := client.DetachVolume(va.ID, nil)
		assert.NoError(t, err, fmt.Sprintf("VolumeAttachment: %#v", va))
		// DeleteVolume must wait until volume is detachched
		for {
			va, _ = client.GetVolumeAttachment(va.ID)
			if va.State != bm.ResourceDetaching {
				require.Equal(t, bm.ResourceDetached, va.State, fmt.Sprintf("Volume Attachment in invalid state: %#v", va))
				break
			}
			helpers.Sleep(2 * time.Second)
		}
		err = client.DeleteVolume(v.ID, nil)
		assert.NoError(t, err, fmt.Sprintf("Volume: %#v", v))
	}()

	// Precise checks
	assert.Equal(t, compartmentID, va.CompartmentID)

	assert.Equal(t, availabilityDomains[0].Name, va.AvailabilityDomain)
	assert.Equal(t, instanceID, va.InstanceID)
	assert.Equal(t, v.ID, va.VolumeID)
	assert.Equal(t, "iscsi", va.AttachmentType)
	assert.Equal(t, bm.ResourceAttached, va.State)
	// Non-empty checks
	assert.NotEmpty(t, va.ID)
	assert.NotEmpty(t, va.TimeCreated)
	assert.NotEmpty(t, va.IPv4)
	assert.NotEmpty(t, va.IQN)
	// Empty checks
	// TODO: create CHAP enabled test
	assert.Empty(t, va.CHAPSecret)
	assert.Empty(t, va.CHAPUsername)

	// Broken
	// assert.Equal(t, "Test Volume Attachment", va.DisplayName)
}
