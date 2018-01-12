// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"time"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"

	bm "github.com/oracle/bmcs-go-sdk"
)

func TestVnicAttachmentCRUD(t *testing.T) {
	// Set up dependencies - actual VNIC tests are in helper methods.
	// Note that this tests both VNICs and VNIC Attachments.

	client := helpers.GetClient("fixtures/vnic_attachment")
	defer client.Stop()

	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	ads, err := client.ListAvailabilityDomains(compartmentID)
	availabilityDomains := ads.AvailabilityDomains
	require.NoError(t, err, "Setup AvailabilityDomains")
	// Create VCN
	vcnOptions := &bm.CreateVcnOptions{}
	vcnOptions.DnsLabel = "gosdktestvcn"
	vcnID, err := helpers.CreateVCNWithOptions(client, "172.16.0.0/16", compartmentID, vcnOptions)
	require.NoError(t, err, "Setup VCN")
	require.NotEmpty(t, vcnID, "Setup VCN: ID")
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err, "Teardown VCN")
	}()
	// Create Subnet
	subnetOptions := &bm.CreateSubnetOptions{}
	subnetOptions.DNSLabel = "gosdktestsubnet"
	subnetID, err := helpers.CreateSubnetWithOptions(client, compartmentID, availabilityDomains[0].Name, vcnID, "172.16.0.0/16", subnetOptions)
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

	VnicAttachmentCRUDTest(t, client, compartmentID, subnetID, instanceID)
	MinParamVnicAttachmentTest(t, client, compartmentID, subnetID, instanceID)
}

func VnicAttachmentCRUDTest(t *testing.T, client *helpers.TestClient, compartmentID string, subnetID string, instanceID string) {
	vnicName := "gosdk_test_vnic"
	vnicAttachmentName := "gosdk_test_vnic_attachment"
	vnicHostnameLabel := "vnica"
	assignPublicIp := false
	skipSourceDestCheck := true

	vaOpts := &bm.AttachVnicOptions{}
	vaOpts.DisplayName = vnicAttachmentName

	vnicOpts := &bm.CreateVnicOptions{}
	vnicOpts.DisplayName = vnicName
	vnicOpts.SubnetID = subnetID

	vnicOpts.AssignPublicIp = &assignPublicIp
	vnicOpts.PrivateIp = "172.16.0.8"
	vnicOpts.HostnameLabel = vnicHostnameLabel
	vnicOpts.SkipSourceDestCheck = &skipSourceDestCheck

	vnicAttachment, err := client.AttachVnic(instanceID, vnicOpts, vaOpts)
	require.NoError(t, err, "Attach VNIC")

	defer func() {
		DetachVnic(t, client, vnicAttachment)
	}()

	// Get VNIC Attachment (and wait until it's attached)
	for {
		vnicAttachment, _ = client.GetVnicAttachment(vnicAttachment.ID)
		if vnicAttachment.State != bm.ResourceAttaching {
			require.Equal(t, bm.ResourceAttached, vnicAttachment.State, "Attach VNIC: Wait for ATTACHED state")
			break
		}
		helpers.Sleep(1 * time.Second)
	}

	assert.NotEmpty(t, vnicAttachment.AvailabilityDomain, "Get: AvailabilityDomain")
	assert.Equal(t, compartmentID, vnicAttachment.CompartmentID, "Get: CompartmentID")
	assert.Equal(t, vnicAttachmentName, vnicAttachment.DisplayName, "Get: DisplayName")
	assert.NotEmpty(t, vnicAttachment.ID, "Get: ID")
	assert.NotEmpty(t, vnicAttachment.InstanceID, "Get: InstanceID")
	assert.Equal(t, bm.ResourceAttached, vnicAttachment.State, "Get: State")
	assert.Equal(t, subnetID, vnicAttachment.SubnetID, "Get: SubnetID")
	assert.NotEmpty(t, vnicAttachment.TimeCreated, "Get: TimeCreated")
	assert.NotEmpty(t, vnicAttachment.VlanTag, "Get: VlanTag")
	assert.NotEmpty(t, vnicAttachment.VnicID, "Get: VnicID")

	// Get VNIC
	vnic, err := client.GetVnic(vnicAttachment.VnicID)
	require.NoError(t, err, "Get VNIC")

	assert.NotEmpty(t, vnic.AvailabilityDomain, "Get: AvailabilityDomain")
	assert.Equal(t, compartmentID, vnic.CompartmentID, "Get: CompartmentID")
	assert.Equal(t, vnicName, vnic.DisplayName, "Get: DisplayName")
	assert.Equal(t, vnicHostnameLabel, vnic.HostnameLabel, "Get: HostnameLabel")
	assert.NotEmpty(t, vnic.ID, "Get: ID")
	assert.Equal(t, false, vnic.IsPrimary, "Get: IsPrimary")
	assert.NotEmpty(t, vnic.MacAddress, "Get: MacAddress")
	assert.Equal(t, bm.ResourceAvailable, vnic.State, "Get: State")
	assert.NotEmpty(t, vnic.PrivateIPAddress, "Get: PrivateIPAddress")
	assert.Equal(t, 0, len(vnic.PublicIPAddress), "Get: PublicIPAddress")
	assert.Equal(t, skipSourceDestCheck, vnic.SkipSourceDestCheck, "Get: SkipSourceDestCheck")
	assert.Equal(t, subnetID, vnic.SubnetID, "Get: SubnetID")
	assert.NotEmpty(t, vnic.TimeCreated, "Get: TimeCreated")

	// Update VNIC
	vnicName = "UPDATED_gosdk_test_vnic"
	vnicHostnameLabel = "vnicb"
	vnicUpdateOptions := &bm.UpdateVnicOptions{}
	vnicUpdateOptions.DisplayName = vnicName
	vnicUpdateOptions.HostnameLabel = vnicHostnameLabel

	vnic, err = client.UpdateVnic(vnicAttachment.VnicID, vnicUpdateOptions)
	require.NoError(t, err, "Update VNIC")
	assert.Equal(t, vnicName, vnic.DisplayName, "Get: DisplayName")
	assert.Equal(t, vnicHostnameLabel, vnic.HostnameLabel, "Get: HostnameLabel")
	// Make sure that SkipSourceDestCheck does not change, since it was not set.
	assert.Equal(t, skipSourceDestCheck, vnic.SkipSourceDestCheck, "Get: SkipSourceDestCheck")

	// Update SkipSourceDestCheck
	vnicName = "UPDATED_gosdk_test_vnic"
	vnicHostnameLabel = "vnicb"
	vnicUpdateOptions2 := &bm.UpdateVnicOptions{}
	skipSourceDestCheck = false
	vnicUpdateOptions2.SkipSourceDestCheck = &skipSourceDestCheck

	vnic, err = client.UpdateVnic(vnicAttachment.VnicID, vnicUpdateOptions2)
	require.NoError(t, err, "Update VNIC")
	assert.Equal(t, skipSourceDestCheck, vnic.SkipSourceDestCheck, "Get: SkipSourceDestCheck")

	// List VNIC Attachments
	listVnicAttachmentsOpts := &bm.ListVnicAttachmentsOptions{}
	listVnicAttachmentsOpts.InstanceID = instanceID

	vaList, err := client.ListVnicAttachments(compartmentID, listVnicAttachmentsOpts)
	assert.Equal(t, 2, len(vaList.Attachments), "List VNIC Attachments")
}

func MinParamVnicAttachmentTest(t *testing.T, client *helpers.TestClient, compartmentID string, subnetID string, instanceID string) {
	vaOpts := &bm.AttachVnicOptions{}

	vnicOpts := &bm.CreateVnicOptions{}
	vnicOpts.SubnetID = subnetID

	vnicAttachment, err := client.AttachVnic(instanceID, vnicOpts, vaOpts)
	require.NoError(t, err, "Attach VNIC")

	defer func() {
		DetachVnic(t, client, vnicAttachment)
	}()

	// Get VNIC Attachment (and wait until it's attached)
	for {
		vnicAttachment, _ = client.GetVnicAttachment(vnicAttachment.ID)
		if vnicAttachment.State != bm.ResourceAttaching {
			require.Equal(t, bm.ResourceAttached, vnicAttachment.State, "Attach Vnic: Wait for ATTACHED state")
			break
		}
		helpers.Sleep(1 * time.Second)
	}

	assert.NotEmpty(t, vnicAttachment.AvailabilityDomain, "Get: AvailabilityDomain")
	assert.Equal(t, compartmentID, vnicAttachment.CompartmentID, "Get: CompartmentID")
	assert.Empty(t, vnicAttachment.DisplayName, "Get: DisplayName")
	assert.NotEmpty(t, vnicAttachment.ID, "Get: ID")
	assert.NotEmpty(t, vnicAttachment.InstanceID, "Get: InstanceID")
	assert.Equal(t, bm.ResourceAttached, vnicAttachment.State, "Get: State")
	assert.Equal(t, subnetID, vnicAttachment.SubnetID, "Get: SubnetID")
	assert.NotEmpty(t, vnicAttachment.TimeCreated, "Get: TimeCreated")
	assert.NotEmpty(t, vnicAttachment.VlanTag, "Get: VlanTag")
	assert.NotEmpty(t, vnicAttachment.VnicID, "Get: VnicID")

	// Get VNIC
	vnic, err := client.GetVnic(vnicAttachment.VnicID)
	require.NoError(t, err, "Get VNIC")

	assert.NotEmpty(t, vnic.AvailabilityDomain, "Get: AvailabilityDomain")
	assert.Equal(t, compartmentID, vnic.CompartmentID, "Get: CompartmentID")
	assert.NotEmpty(t, vnic.DisplayName, "Get: DisplayName")
	assert.Empty(t, vnic.HostnameLabel, "Get: HostnameLabel")
	assert.NotEmpty(t, vnic.ID, "Get: ID")
	assert.Equal(t, false, vnic.IsPrimary, "Get: IsPrimary")
	assert.NotEmpty(t, vnic.MacAddress, "Get: MacAddress")
	assert.Equal(t, bm.ResourceAvailable, vnic.State, "Get: State")
	assert.NotEmpty(t, vnic.PrivateIPAddress, "Get: PrivateIPAddress")
	assert.NotEmpty(t, vnic.PublicIPAddress, "Get: PublicIPAddress")
	assert.Equal(t, false, vnic.SkipSourceDestCheck, "Get: SkipSourceDestCheck")
	assert.Equal(t, subnetID, vnic.SubnetID, "Get: SubnetID")
	assert.NotEmpty(t, vnic.TimeCreated, "Get: TimeCreated")
}

func DetachVnic(t *testing.T, client *helpers.TestClient, vnicAttachment *bm.VnicAttachment) {
	err := client.DetachVnic(vnicAttachment.ID, nil)
	assert.NoError(t, err, "Detach VNIC")

	// Wait until VNIC Attachment is detached
	for {
		vnicAttachment, _ = client.GetVnicAttachment(vnicAttachment.ID)
		if vnicAttachment.State != bm.ResourceDetaching {
			require.Equal(t, bm.ResourceDetached, vnicAttachment.State, "Detach Vnic: Wait for DETACHED state")
			break
		}
		helpers.Sleep(1 * time.Second)
	}
}
