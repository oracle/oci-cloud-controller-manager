// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package acceptance

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oracle/bmcs-go-sdk/acceptance-test/helpers"
)

func TestSubnetCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/subnet")
	defer client.Stop()
	// get a compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	// Get Availability Domains
	ads, err := client.ListAvailabilityDomains(compartmentID)
	require.NoError(t, err, "Setup AvailabilityDomains")
	availabilityDomains := ads.AvailabilityDomains
	// Create VCN
	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err, "Setup VCN")
	require.NotEmpty(t, vcnID, "Setup VCN: ID")
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err, "Teardown VCN")
	}()

	// Create
	id, err := helpers.CreateSubnet(client, compartmentID, availabilityDomains[0].Name, vcnID)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, id, "Create: ID")

	// TODO: Get
	// TODO: Update

	// List
	subnets, err := client.ListSubnets(compartmentID, vcnID, nil)
	assert.NoError(t, err, "List")
	found := false
	for _, sub := range subnets.Subnets {
		if strings.Compare(sub.ID, id) == 0 {
			found = true
		}
	}
	assert.True(t, found, "List: Created Subnet not found")

	// Delete
	_, err = helpers.DeleteSubnet(client, id)
	assert.NoError(t, err, "Delete")
	helpers.Sleep(2 * time.Second)
}
