// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build recording,core recording,all !recording

package acceptance

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MustWin/baremetal-sdk-go/acceptance-test/helpers"
)

func TestSubnetCreate(t *testing.T) {
	client := helpers.GetClient("fixtures/core/subnet")
	defer client.Stop()

	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	// Get Availability Domains
	ads, err := client.ListAvailabilityDomains(compartmentID)
	availabilityDomains := ads.AvailabilityDomains
	require.NoError(t, err)

	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, vcnID)

	id, err := helpers.CreateSubnet(client, compartmentID, availabilityDomains[0].Name, vcnID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteSubnet(client, id)
		assert.NoError(t, err)
		helpers.Sleep(2 * time.Second)
	}()
	require.NotEmpty(t, id)

	subnet, err := client.GetSubnet(id)
	require.NoError(t, err)
	assert.Equal(t, id, subnet.ID)
}

func TestSubnetList(t *testing.T) {
	client := helpers.GetClient("fixtures/core/subnet_list")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err)

	// Get Availability Domains
	ads, err := client.ListAvailabilityDomains(compartmentID)
	availabilityDomains := ads.AvailabilityDomains
	require.NoError(t, err)

	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err)
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err)
	}()
	require.NotEmpty(t, vcnID)

	id, err := helpers.CreateSubnet(client, compartmentID, availabilityDomains[0].Name, vcnID)
	require.NoError(t, err)
	defer func() {
		_, err := helpers.DeleteSubnet(client, id)
		assert.NoError(t, err)
		helpers.Sleep(2 * time.Second)
	}()
	require.NotEmpty(t, id)

	subnets, err := client.ListSubnets(compartmentID, vcnID, nil)
	require.NoError(t, err)

	found := false
	for _, sub := range subnets.Subnets {
		if strings.Compare(sub.ID, id) == 0 {
			found = true
		}
	}
	assert.True(t, found, "Created subnet not found in list")
}
