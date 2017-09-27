// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bm "github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestVolumeAttachmentCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/volume_attachment")
	defer client.Stop()
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	ads, err := client.ListAvailabilityDomains(compartmentID)
	availabilityDomains := ads.AvailabilityDomains
	require.NoError(t, err, "Setup AvailabilityDomains")
	// Create VCN
	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err, "Setup VCN")
	require.NotEmpty(t, vcnID, "Setup VCN: ID")
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err, "Teardown VCN")
	}()
	// Create Subnet
	subnetID, err := helpers.CreateSubnet(client, compartmentID, availabilityDomains[0].Name, vcnID)
	require.NoError(t, err, "Setup Subnet")
	defer func() {
		_, err := helpers.DeleteSubnet(client, subnetID)
		assert.NoError(t, err, "Teardown Subnet")
	}()
	// Create Instance
	instanceID, err := helpers.CreateInstance(client, compartmentID, availabilityDomains[0].Name, helpers.FastestImageID, helpers.SmallestShapeName, subnetID)
	require.NoError(t, err, "Setup Instance")
	defer func() {
		_, err := helpers.DeleteInstance(client, instanceID)
		assert.NoError(t, err, "Teardown Instance")
	}()
	// Create Volume
	v, err := client.CreateVolume(
		availabilityDomains[0].Name,
		compartmentID,
		nil,
	)
	require.NoError(t, err, "Setup Volume")
	// Attachment must wait until volume is available
	for {
		v, _ := client.GetVolume(v.ID)
		if v.State != bm.ResourceProvisioning {
			require.Equal(t, bm.ResourceAvailable, v.State, "Setup Volume: State")
			break
		}
		helpers.Sleep(2 * time.Second)
	}
	defer func() {
		err = client.DeleteVolume(v.ID, nil)
		assert.NoError(t, err, "Teardown Volume")
	}()

	// Attach
	opts := bm.CreateOptions{}
	opts.DisplayName = "Test Volume Attachment"
	va, err := client.AttachVolume("iscsi", instanceID, v.ID, &opts)
	assert.NoError(t, err, "Attach")

	// Get
	for {
		va, _ = client.GetVolumeAttachment(va.ID)
		if va.State != bm.ResourceAttaching {
			assert.Equal(t, bm.ResourceAttached, va.State, "Get: State")
			break
		}
		helpers.Sleep(2 * time.Second)
	}
	assert.Equal(t, compartmentID, va.CompartmentID, "Get: CompartmentID")
	assert.Equal(t, availabilityDomains[0].Name, va.AvailabilityDomain, "Get: AvailabilityDomain")
	assert.Equal(t, instanceID, va.InstanceID, "Get: InstanceID")
	assert.Equal(t, v.ID, va.VolumeID, "Get: VolumeID")
	assert.Equal(t, "iscsi", va.AttachmentType, "Get: AttachmentType")
	assert.Equal(t, bm.ResourceAttached, va.State, "Get: State")
	assert.NotEmpty(t, va.ID, "Get: ID")
	assert.NotEmpty(t, va.TimeCreated, "Get: TimeCreated")
	assert.NotEmpty(t, va.IPv4, "Get: IPv4")
	assert.NotEmpty(t, va.IQN, "Get: IQN")
	// TODO: create CHAP enabled test
	assert.Empty(t, va.CHAPSecret, "Get: CHAPSecret")
	assert.Empty(t, va.CHAPUsername, "Get: CHAPUsername")
	// FIXME: broken API behaviour
	// assert.Equal(t, "Test Volume Attachment", va.DisplayName, "Get: DisplayName")

	// TODO: List

	// Detatch
	err = client.DetachVolume(va.ID, nil)
	assert.NoError(t, err, "Detach")
	// DeleteVolume must wait until volume is detachched
	for {
		va, _ = client.GetVolumeAttachment(va.ID)
		if va.State != bm.ResourceDetaching {
			assert.Equal(t, bm.ResourceDetached, va.State, "Detatch")
			break
		}
		helpers.Sleep(2 * time.Second) // TODO: remove once state may be trusted
	}
}
