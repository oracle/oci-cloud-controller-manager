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

func TestInstanceCRUD(t *testing.T) {
	// Arrange
	client := helpers.GetClient("fixtures/instance")
	defer client.Stop()
	// Get compartment, any compartment
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	// Get Availability Domain
	ads, err := client.ListAvailabilityDomains(compartmentID)
	require.NoError(t, err, "Setup AvailabilityDomains")
	availabilityDomainName := ads.AvailabilityDomains[0].Name
	// Create VCN
	vcnID, err := helpers.CreateVCN(client, "172.16.0.0/16", compartmentID)
	require.NoError(t, err, "Setup VCN")
	require.NotEmpty(t, vcnID, "Setup VCN: ID")
	defer func() {
		_, err = helpers.DeleteVCN(client, vcnID)
		assert.NoError(t, err, "Teardown VCN")
	}()
	// Create Subnet
	subnetID, err := helpers.CreateSubnet(client, compartmentID, availabilityDomainName, vcnID)
	require.NoError(t, err, "Setup Subnet")
	require.NotEmpty(t, subnetID, "Setup Subnet: ID")
	defer func() {
		_, err := helpers.DeleteSubnet(client, subnetID)
		assert.NoError(t, err, "Teardown Subnet")
		helpers.Sleep(2 * time.Second)
	}()

	// Create
	id, err := helpers.CreateInstance(
		client,
		compartmentID,
		availabilityDomainName,
		helpers.FastestImageID,
		helpers.SmallestShapeName,
		subnetID,
	)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, id, "Create: ID")

	// TODO: Get
	// TODO: Update
	// TODO: InstanceAction
	// TODO: GetWindowsInstanceInitialCredentials

	// List
	instances, err := client.ListInstances(compartmentID, nil)
	assert.NoError(t, err, "List")
	found := false
	for _, instance := range instances.Instances {
		if strings.Compare(id, instance.ID) == 0 {
			found = true
		}
	}
	assert.True(t, found, "List: Created Instance not found")

	// Delete
	_, err = helpers.DeleteInstance(client, id)
	assert.NoError(t, err, "Delete")
}
