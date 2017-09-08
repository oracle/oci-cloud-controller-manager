// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core_instance recording,all !recording

package acceptance

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

func TestInstanceList(t *testing.T) {
	// Arrange
	client := helpers.GetClient("fixtures/core/instance_list")
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
	id, err := helpers.CreateInstance(
		client,
		compartmentID,
		availabilityDomainName,
		helpers.FastestImageID,
		helpers.SmallestShapeName,
		subnetID,
	)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteInstance(client, id)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, id)

	// Act
	instances, err := client.ListInstances(compartmentID, nil)
	require.NoError(t, err)
	found := false
	for _, instance := range instances.Instances {
		if strings.Compare(id, instance.ID) == 0 {
			found = true
		}
	}

	// Assert
	require.True(t, found, "Created Instance not found in list")
}

func TestInstanceCreate(t *testing.T) {
	// Arrange
	client := helpers.GetClient("fixtures/core/instance")
	defer client.Stop()
	// get a compartment, any compartment
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

	// Act
	id, err := helpers.CreateInstance(
		client,
		compartmentID,
		availabilityDomainName,
		helpers.FastestImageID,
		helpers.SmallestShapeName,
		subnetID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteInstance(client, id)
		assert.NoError(t, err)
	}()

	// Assert
	require.NotEmpty(t, id)
}
