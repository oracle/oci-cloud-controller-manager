// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core recording,all !recording

package acceptance

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk"
	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

const (
	ipAddress     = "172.16.0.3"
	displayName1  = "privateIP"
	displayName2  = "privateIP2"
	hostnameLabel = "hostnamelabel"
)

func TestPrivateIP(t *testing.T) {
	client := helpers.GetClient("fixtures/core/private_ip")
	defer client.Stop()

	// Get compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)
	// Get Availability Domain
	ads, err := client.ListAvailabilityDomains(compartmentID)
	require.NoError(t, err)
	availabilityDomainName := ads.AvailabilityDomains[0].Name
	// Create VCN
	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, vcnID)
	// Create Subnet
	subnetID, err := helpers.CreateSubnet(client, compartmentID, availabilityDomainName, vcnID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteSubnet(client, subnetID)
		assert.NoError(t, err)
		helpers.Sleep(2 * time.Second)
	}()
	require.NotEmpty(t, subnetID)
	// Create Instance
	instanceID, err := helpers.CreateInstance(
		client,
		compartmentID,
		availabilityDomainName,
		helpers.FastestImageID,
		helpers.SmallestShapeName,
		subnetID,
	)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteInstance(client, instanceID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, instanceID)

	//get VNIC ID
	opts := &baremetal.ListVnicAttachmentsOptions{}
	opts.InstanceID = instanceID
	vnicAttachments, err := client.ListVnicAttachments(compartmentID, opts)
	var vnicID string
	if err == nil {
		vnicID = vnicAttachments.Attachments[0].VnicID
	}

	//Create private ip
	CreateOpts := &baremetal.CreatePrivateIPOptions{}
	CreateOpts.DisplayName = displayName1
	CreateOpts.IPAddress = ipAddress
	privateIP, err := client.CreatePrivateIP(vnicID, CreateOpts)
	assert.NoError(t, err)
	assert.NotNil(t, privateIP)
	privateIPID := privateIP.ID
	defer func() {
		err := client.DeletePrivateIP(privateIPID, nil)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, privateIPID)
	assert.Equal(t, vnicID, privateIP.VnicID)
	assert.Equal(t, displayName1, privateIP.DisplayName)
	assert.Equal(t, ipAddress, privateIP.IPAddress)

	//Get Private IP
	privateIPObj, err := client.GetPrivateIP(privateIPID)
	require.NoError(t, err)
	assert.Equal(t, privateIPID, privateIPObj.ID)
	assert.Equal(t, false, privateIPObj.IsPrimary)
	assert.Equal(t, subnetID, privateIPObj.SubnetID)
	assert.Equal(t, vnicID, privateIPObj.VnicID)
	assert.NotEmpty(t, privateIPObj.ETag)
	assert.NotEmpty(t, privateIPObj.RequestID)
	assert.NotEmpty(t, privateIPObj.TimeCreated)

	//Update Private IP
	updateOpts := &baremetal.UpdatePrivateIPOptions{}
	updateOpts.DisplayName = displayName2
	privateIP, err = client.UpdatePrivateIP(privateIPID, updateOpts)
	assert.NoError(t, err)
	assert.NotNil(t, privateIP)
	assert.Equal(t, privateIP.DisplayName, displayName2)
	updateOpts.HostnameLabel = hostnameLabel
	privateIP, err = client.UpdatePrivateIP(privateIPID, updateOpts)
	assert.NoError(t, err)
	assert.NotNil(t, privateIP)
	assert.Equal(t, privateIP.HostnameLabel, hostnameLabel)

	//test multiple ways to List (in accordance with spec)
	var privateIPs *baremetal.ListPrivateIPs
	var found bool

	//list using subnetID
	listOpts := &baremetal.ListPrivateIPsOptions{}
	listOpts.SubnetID = subnetID
	privateIPs, err = client.ListPrivateIPs(listOpts)
	require.NoError(t, err)
	found = false
	for _, privateIP := range privateIPs.PrivateIPs {
		if strings.Compare(privateIPID, privateIP.ID) == 0 {
			found = true
		}
	}
	require.True(t, found, "Created Private IP not found while listing by subnetID")

	//list using subnetID and IP, this is used to get PrivateIP by IPAddress
	listOpts.IPAddress = ipAddress
	privateIPs, err = client.ListPrivateIPs(listOpts)
	require.NoError(t, err)
	found = false
	for _, privateIP := range privateIPs.PrivateIPs {
		if strings.Compare(privateIPID, privateIP.ID) == 0 {
			found = true
		}
	}
	require.True(t, found, "Created Private IP not found while listing by subnetID and Private IP")

	//list using vnicID
	listOpts = &baremetal.ListPrivateIPsOptions{}
	listOpts.VnicID = vnicID
	privateIPs, err = client.ListPrivateIPs(listOpts)
	require.NoError(t, err)
	found = false
	for _, privateIP := range privateIPs.PrivateIPs {
		if strings.Compare(privateIPID, privateIP.ID) == 0 {
			found = true
		}
	}
	require.True(t, found, "Created Private IP not found while listing by VNIC ID")
}
