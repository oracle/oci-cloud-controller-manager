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

func TestVolumeCRUD(t *testing.T) {
	client := helpers.GetClient("fixtures/volume")
	defer client.Stop()
	compartmentID, err := helpers.FindOrCreateCompartmentID(client)
	require.NoError(t, err, "Setup Compartment")
	ads, err := client.ListAvailabilityDomains(compartmentID)
	require.NoError(t, err, "Setup AvailabilityDomains")
	availabilityDomains := ads.AvailabilityDomains

	// Create
	v, err := client.CreateVolume(
		availabilityDomains[0].Name,
		compartmentID,
		nil,
	)
	assert.NoError(t, err, "Create")
	assert.NotEmpty(t, v.ID, "Create: ID")

	// Get
	for {
		vs, _ := client.GetVolume(v.ID)
		if vs.State != bm.ResourceProvisioning {
			break
		}
		helpers.Sleep(2 * time.Second) // wait until provisioning is complete
	}

	// TODO: Update
	// TODO: List

	// Delete
	err = client.DeleteVolume(v.ID, nil)
	assert.NoError(t, err, "Delete")
}
